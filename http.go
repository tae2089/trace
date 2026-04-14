package trace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// ErrorResponse represents a structured JSON error response
type ErrorResponse struct {
	Error   string         `json:"error"`
	Code    int            `json:"code"`
	Message string         `json:"message,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

// WriteError writes an error response to the HTTP response writer
func WriteError(w http.ResponseWriter, err error) {
	WriteErrorWithLogger(w, err, nil)
}

// WriteErrorWithLogger writes an error response and logs it
func WriteErrorWithLogger(w http.ResponseWriter, err error, logger *slog.Logger) {
	if err == nil {
		return
	}

	statusCode := GetHTTPStatusCode(err)

	resp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    statusCode,
		Message: UserMessage(err),
	}

	// Add trace ID if available in fields
	if fields := GetFields(err); fields != nil {
		if traceID, ok := fields["trace_id"].(string); ok {
			resp.TraceID = traceID
		}
		// Only include safe details
		if details, ok := fields["details"].(map[string]any); ok {
			resp.Details = details
		}
	}

	// Log the full error with stack trace
	if logger != nil {
		logger.Error("http error",
			SlogError(err),
			slog.Int("status_code", statusCode),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil && logger != nil {
		logger.Error("failed to encode error response", "encode_error", encErr)
	}
}

// ErrorHandlerFunc is a function that handles HTTP requests and may return an error
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorMiddleware converts an ErrorHandlerFunc to a standard http.HandlerFunc
func ErrorMiddleware(h ErrorHandlerFunc) http.HandlerFunc {
	return ErrorMiddlewareWithLogger(h, nil)
}

// ErrorMiddlewareWithLogger converts an ErrorHandlerFunc to http.HandlerFunc with logging
func ErrorMiddlewareWithLogger(h ErrorHandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			// Add request context to error
			err = WithFields(err, map[string]any{
				"method": r.Method,
				"path":   r.URL.Path,
			})
			WriteErrorWithLogger(w, err, logger)
		}
	}
}

// RecoverMiddleware recovers from panics and converts them to errors
func RecoverMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				var err error
				switch v := rec.(type) {
				case error:
					err = Wrap(v, "panic recovered")
				default:
					err = Errorf("panic recovered: %v", v)
				}

				err = WithFields(err, map[string]any{
					"method": r.Method,
					"path":   r.URL.Path,
					"panic":  true,
				})

				if logger != nil {
					logger.Error("panic recovered",
						SlogError(err),
					)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				if encErr := json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "Internal Server Error",
					Code:    http.StatusInternalServerError,
					Message: "An unexpected error occurred",
				}); encErr != nil && logger != nil {
					logger.Error("failed to encode panic response", "encode_error", encErr)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// FromHTTPResponse creates an appropriate error from an HTTP response
func FromHTTPResponse(resp *http.Response, body []byte) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	msg := string(body)
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	frame := captureFrame(2)
	te := &TraceError{
		Message: msg,
		Frames:  Frames{frame},
		Fields: map[string]any{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		},
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return &NotFoundError{TraceError: te}
	case http.StatusConflict:
		return &ConflictError{TraceError: te}
	case http.StatusBadRequest:
		return &BadParameterError{TraceError: te}
	case http.StatusUnauthorized, http.StatusForbidden:
		return &AccessDeniedError{TraceError: te}
	case http.StatusTooManyRequests:
		return &LimitExceededError{TraceError: te}
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return &TimeoutError{TraceError: te}
	case http.StatusServiceUnavailable, http.StatusBadGateway:
		return &ConnectionProblemError{TraceError: te}
	default:
		return &httpStatusError{
			TraceError: te,
			statusCode: resp.StatusCode,
		}
	}
}

// IsHTTPError checks if an error corresponds to a specific HTTP status code
func IsHTTPError(err error, statusCode int) bool {
	return GetHTTPStatusCode(err) == statusCode
}

// WrapHTTPError wraps an error with HTTP status code information
func WrapHTTPError(err error, statusCode int, msg ...string) error {
	if err == nil {
		return nil
	}

	frame := captureFrame(2)
	var message string
	if len(msg) > 0 {
		message = msg[0]
	}
	te := wrapTypedInternal(err, message, frame)
	if te.Fields == nil {
		te.Fields = make(map[string]any)
	}
	te.Fields["http_status"] = statusCode

	return &httpStatusError{
		TraceError: te,
		statusCode: statusCode,
	}
}

type httpStatusError struct {
	*TraceError
	statusCode int
}

func (e *httpStatusError) HTTPStatusCode() int { return e.statusCode }
func (e *httpStatusError) Error() string       { return e.TraceError.Error() }
func (e *httpStatusError) Unwrap() error       { return e.TraceError }

// Client is an HTTP client that wraps errors with trace information
type Client struct {
	*http.Client
}

// NewClient creates a new trace-aware HTTP client
func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{Client: client}
}

// Do executes the request and wraps any errors with trace information
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, Timeout(err, fmt.Sprintf("request to %s timed out", req.URL.Host))
		}
		return nil, ConnectionProblem(err, fmt.Sprintf("request to %s failed", req.URL.Host))
	}
	return resp, nil
}
