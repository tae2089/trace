// @index HTTP adapters that translate trace errors into API responses, middleware behavior, and client-side classifications.
package trace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// @intent give clients a stable machine-readable classification independent of HTTP status text.
// ErrorCode is a stable machine-readable HTTP error classification.
type ErrorCode string

const (
	// CodeBadRequest identifies invalid client input.
	CodeBadRequest ErrorCode = "bad_request"
	// CodeUnauthenticated identifies a missing or invalid authentication identity.
	CodeUnauthenticated ErrorCode = "unauthenticated"
	// CodeAccessDenied identifies an authenticated caller without permission.
	CodeAccessDenied ErrorCode = "access_denied"
	// CodeNotFound identifies a missing resource.
	CodeNotFound ErrorCode = "not_found"
	// CodeAlreadyExists identifies a duplicate resource.
	CodeAlreadyExists ErrorCode = "already_exists"
	// CodeConflict identifies a state conflict.
	CodeConflict ErrorCode = "conflict"
	// CodeLimitExceeded identifies a rate or quota limit.
	CodeLimitExceeded ErrorCode = "limit_exceeded"
	// CodeCanceled identifies a canceled request.
	CodeCanceled ErrorCode = "canceled"
	// CodeNotImplemented identifies unavailable functionality.
	CodeNotImplemented ErrorCode = "not_implemented"
	// CodeUnavailable identifies a transiently unavailable dependency.
	CodeUnavailable ErrorCode = "unavailable"
	// CodeTimeout identifies a timed-out operation.
	CodeTimeout ErrorCode = "timeout"
	// CodeInternal identifies a failure that is intentionally hidden from clients.
	CodeInternal ErrorCode = "internal"
)

// @intent carry only validated status, semantic code, and client-safe message across an HTTP boundary.
// HTTPError is the safe client-facing representation of an error.
type HTTPError struct {
	Status  int
	Code    ErrorCode
	Message string
}

// @intent let application-owned error types opt into explicit safe HTTP classification.
// HTTPErrorProvider lets custom errors define a safe HTTP representation.
type HTTPErrorProvider interface {
	error
	HTTPError() HTTPError
}

// @intent classify an error chain into a validated client-safe HTTP representation.
// @domainRule invalid classifications become internal errors and every 5xx message is sanitized.
// @ensures returns the zero value for nil and never derives a message from an outer trace wrapper.
// ToHTTPError returns the safe HTTP representation for err.
func ToHTTPError(err error) HTTPError {
	if err == nil {
		return HTTPError{}
	}

	var provider HTTPErrorProvider
	if !errors.As(err, &provider) {
		return internalHTTPError()
	}

	result := provider.HTTPError()
	if result.Status < http.StatusBadRequest || result.Status > 599 || result.Code == "" {
		return internalHTTPError()
	}
	if result.Status >= http.StatusInternalServerError {
		result.Message = "internal server error"
	}
	return result
}

// @intent centralize the generic fallback used whenever a failure cannot be exposed safely.
func internalHTTPError() HTTPError {
	return HTTPError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternal,
		Message: "internal server error",
	}
}

// @intent group safe client fields under one stable JSON error envelope.
// ErrorBody is the stable machine-readable body nested under the error key.
type ErrorBody struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id,omitempty"`
}

// @intent define the client-facing error payload returned by HTTP helpers in this package.
// ErrorResponse represents a structured JSON error response.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// @intent build a framework-neutral response from the safe HTTP classification.
// @domainRule request IDs are caller-supplied and trace fields or details are never copied from err.
// @ensures returns zero values for nil errors.
// ErrorResponseFor returns the status and safe response body for err.
func ErrorResponseFor(err error, requestID string) (status int, response ErrorResponse) {
	httpError := ToHTTPError(err)
	if httpError.Status == 0 {
		return 0, ErrorResponse{}
	}

	return httpError.Status, ErrorResponse{
		Error: ErrorBody{
			Code:      httpError.Code,
			Message:   httpError.Message,
			RequestID: requestID,
		},
	}
}

// @intent expose a simple entry point for converting domain errors into HTTP responses.
// @sideEffect writes status code, headers, and a JSON body to the response writer.
// @ensures returns without writing anything when the input error is nil.
// WriteError writes a safe error response without logging.
func WriteError(w http.ResponseWriter, err error, requestID string) error {
	statusCode, resp := ErrorResponseFor(err, requestID)
	if statusCode == 0 {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(resp)
}

// @intent retain the legacy opt-in logging adapter while directing new callers to application-owned logging.
// WriteErrorWithLogger writes an error response and logs it.
//
// Deprecated: applications should log request completion and call WriteError separately.
func WriteErrorWithLogger(w http.ResponseWriter, err error, logger *slog.Logger) {
	if err == nil {
		return
	}

	statusCode := ToHTTPError(err).Status

	// Log the full error with stack trace
	if logger != nil {
		logger.Error("http error",
			SlogError(err),
			slog.Int("status_code", statusCode),
		)
	}

	if encErr := WriteError(w, err, ""); encErr != nil && logger != nil {
		logger.Error("failed to encode error response", "encode_error", encErr)
	}
}

// @intent let HTTP handlers return errors directly so middleware can centralize response rendering.
// ErrorHandlerFunc is a function that handles HTTP requests and may return an error
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// @intent adapt error-returning handlers into standard net/http handlers.
// @ensures renders handler errors through the logger-free safe response writer.
// ErrorMiddleware converts an ErrorHandlerFunc to a standard http.HandlerFunc
func ErrorMiddleware(h ErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		err = WithFields(err, map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
		})
		_ = WriteError(w, err, "")
	}
}

// @intent normalize handler failures into traced HTTP responses enriched with request metadata.
// @sideEffect attaches method and path fields before writing the HTTP error response.
// @ensures successful handlers pass through without writing an additional error response.
// ErrorMiddlewareWithLogger converts an ErrorHandlerFunc to http.HandlerFunc with logging.
//
// Deprecated: applications should own request logging and use ErrorMiddleware or ErrorResponseFor.
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

// @intent prevent panics from escaping the HTTP boundary and turn them into traceable server errors.
// @domainRule recovered panics always return HTTP 500 with a generic client-facing message.
// @sideEffect recovers panics, logs structured diagnostics, and writes a fallback JSON response.
// RecoverMiddleware recovers from panics and converts them to errors.
//
// Deprecated: applications should own panic recovery and request-completion logging.
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
					Error: ErrorBody{
						Code:    CodeInternal,
						Message: "internal server error",
					},
				}); encErr != nil && logger != nil {
					logger.Error("failed to encode panic response", "encode_error", encErr)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// @intent restore typed trace semantics from the public ErrorResponse envelope without deserializing internal trace data.
// @domainRule HTTP 2xx and 3xx statuses are not errors.
// @domainRule malformed, unknown, or status-mismatched responses fail closed without retaining the raw body.
// ReadErrorResponse converts a public HTTP error response into a trace error.
func ReadErrorResponse(statusCode int, responseBody []byte) error {
	if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
		return nil
	}

	frame := captureFrame(2)
	var response ErrorResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return newReadErrorResponseTrace(frame, statusCode, "", "invalid error response")
	}
	if response.Error.Code == "" ||
		response.Error.Message == "" ||
		!errorCodeMatchesStatus(response.Error.Code, statusCode) {
		return newReadErrorResponseTrace(
			frame,
			statusCode,
			response.Error.RequestID,
			"invalid error response",
		)
	}

	message := response.Error.Message
	switch {
	case response.Error.Code == CodeUnauthenticated:
		message = "authentication required"
	case response.Error.Code == CodeAccessDenied:
		message = "access denied"
	case response.Error.Code == CodeCanceled:
		message = "request canceled"
	case statusCode >= http.StatusInternalServerError:
		message = "internal server error"
	}

	traceError := newReadErrorResponseTrace(
		frame,
		statusCode,
		response.Error.RequestID,
		message,
	)
	switch response.Error.Code {
	case CodeBadRequest:
		return &BadParameterError{TraceError: traceError}
	case CodeUnauthenticated:
		return &UnauthenticatedError{TraceError: traceError}
	case CodeAccessDenied:
		return &AccessDeniedError{TraceError: traceError}
	case CodeNotFound:
		return &NotFoundError{TraceError: traceError}
	case CodeAlreadyExists:
		return &AlreadyExistsError{TraceError: traceError}
	case CodeConflict:
		return &ConflictError{TraceError: traceError}
	case CodeLimitExceeded:
		return &LimitExceededError{TraceError: traceError}
	case CodeCanceled:
		return &CanceledError{TraceError: traceError}
	case CodeNotImplemented:
		return &NotImplementedError{TraceError: traceError}
	case CodeUnavailable:
		return &ConnectionProblemError{TraceError: traceError}
	case CodeTimeout:
		return &TimeoutError{TraceError: traceError}
	case CodeInternal:
		return traceError
	default:
		return newReadErrorResponseTrace(
			frame,
			statusCode,
			response.Error.RequestID,
			"invalid error response",
		)
	}
}

// @intent validate the public code and HTTP status as one coherent wire contract.
func errorCodeMatchesStatus(code ErrorCode, statusCode int) bool {
	switch code {
	case CodeBadRequest:
		return statusCode == http.StatusBadRequest
	case CodeUnauthenticated:
		return statusCode == http.StatusUnauthorized
	case CodeAccessDenied:
		return statusCode == http.StatusForbidden
	case CodeNotFound:
		return statusCode == http.StatusNotFound
	case CodeAlreadyExists, CodeConflict:
		return statusCode == http.StatusConflict
	case CodeLimitExceeded:
		return statusCode == http.StatusTooManyRequests
	case CodeCanceled:
		return statusCode == 499
	case CodeNotImplemented:
		return statusCode == http.StatusNotImplemented
	case CodeUnavailable:
		return statusCode == http.StatusServiceUnavailable
	case CodeTimeout:
		return statusCode == http.StatusGatewayTimeout
	case CodeInternal:
		return statusCode >= http.StatusInternalServerError && statusCode <= 599
	default:
		return false
	}
}

// @intent construct a fresh local trace from only the safe response metadata allowed across the HTTP boundary.
func newReadErrorResponseTrace(
	frame Frame,
	statusCode int,
	requestID string,
	message string,
) *TraceError {
	fields := map[string]any{"status_code": statusCode}
	if requestID != "" {
		fields["request_id"] = requestID
	}
	return &TraceError{
		Message: message,
		Frames:  Frames{frame},
		Fields:  fields,
	}
}

// @intent retain legacy status-based classification for plain-text upstream responses.
// @domainRule 2xx responses are not errors, while known status codes map to typed trace errors.
// @ensures records the current call site as the first trace frame and stores upstream status metadata in error fields.
// FromHTTPResponse creates an appropriate error from an HTTP response.
// It retains body in the developer-facing error; prefer ReadErrorResponse for the safe JSON envelope.
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
	case http.StatusUnauthorized:
		return &UnauthenticatedError{TraceError: te}
	case http.StatusForbidden:
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

// @intent test whether an error chain resolves to a specific HTTP status mapping.
// @ensures delegates status resolution to GetHTTPStatusCode.
// IsHTTPError checks if an error corresponds to a specific HTTP status code
func IsHTTPError(err error, statusCode int) bool {
	return GetHTTPStatusCode(err) == statusCode
}

// @intent override or attach explicit HTTP status semantics to an existing error chain.
// @domainRule returns nil unchanged when the source error is nil.
// @mutates adds http_status metadata to the returned traced wrapper.
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

// @intent carry an explicit HTTP status override for errors that do not map to one of the standard typed categories.
type httpStatusError struct {
	*TraceError
	statusCode int
}

// @intent expose the explicit status override carried by this internal HTTP error wrapper.
func (e *httpStatusError) HTTPStatusCode() int { return e.statusCode }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *httpStatusError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *httpStatusError) Unwrap() error { return e.TraceError }

// @intent wrap http.Client so transport failures come back as trace-classified errors.
// Client is an HTTP client that wraps errors with trace information
type Client struct {
	*http.Client
}

// @intent provide an HTTP client wrapper that returns trace-classified transport failures.
// @ensures falls back to http.DefaultClient when no custom client is supplied.
// NewClient creates a new trace-aware HTTP client
func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{Client: client}
}

// @intent classify outbound HTTP transport failures into timeout or connection problem errors.
// @domainRule deadline exceeded maps to Timeout and other transport failures map to ConnectionProblem.
// @sideEffect executes the underlying HTTP request through the wrapped client.
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
