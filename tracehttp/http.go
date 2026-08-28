// @index HTTP adapters that translate trace errors into API responses, middleware behavior, and client-side classifications.
//
// Package tracehttp holds every part of trace that depends on net/http. It is a
// separate package so that programs which only need error values, stack frames,
// and slog integration never pull net/http, encoding/json, or the crypto/tls
// tree into their binaries.
package tracehttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tae2089/trace/v2"
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

// StatusClientClosedRequest is the non-standard status used for a request the
// client canceled before the server finished. It matches nginx's 499.
const StatusClientClosedRequest = 499

// @intent carry only validated status, semantic code, and client-safe message across an HTTP boundary.
// HTTPError is the safe client-facing representation of an error.
type HTTPError struct {
	Status  int
	Code    ErrorCode
	Message string
}

// @intent let application-owned error types opt into explicit safe HTTP classification.
// HTTPErrorProvider lets custom errors define a safe HTTP representation.
//
// It is the single extension point for application-defined mapping. An error
// that implements it wins over this package's built-in classification, and the
// outermost implementer in a chain wins over inner ones.
type HTTPErrorProvider interface {
	error
	HTTPError() HTTPError
}

// @intent classify an error chain into a validated client-safe HTTP representation.
// @domainRule the outermost classifiable link in the chain decides the response, so an
// outer AccessDenied wrap is never overridden by an inner NotFound.
// @domainRule invalid classifications become internal errors and every 5xx message is sanitized.
// @ensures returns the zero value for nil and never derives a message from an outer trace wrapper.
// ToHTTPError returns the safe HTTP representation for err.
func ToHTTPError(err error) HTTPError {
	if err == nil {
		return HTTPError{}
	}

	for link := range trace.Errors(err) {
		if provider, ok := link.(HTTPErrorProvider); ok {
			return validate(provider.HTTPError())
		}
		if aggregate, ok := link.(*trace.AggregateError); ok {
			return aggregateHTTPError(aggregate)
		}
		if result, ok := classify(link); ok {
			return validate(result)
		}
	}

	return internalHTTPError()
}

// @intent classify exactly one error value without following its Unwrap chain.
// @domainRule categories are tested in a fixed order so an error implementing several
// behavior interfaces always produces the same status.
func classify(link error) (HTTPError, bool) {
	message := linkMessage(link)

	switch e := link.(type) {
	case trace.ErrorNotFound:
		if e.IsNotFound() {
			return HTTPError{http.StatusNotFound, CodeNotFound, message}, true
		}
	case trace.ErrorAlreadyExists:
		if e.IsAlreadyExists() {
			return HTTPError{http.StatusConflict, CodeAlreadyExists, message}, true
		}
	case trace.ErrorBadParameter:
		if e.IsBadParameter() {
			return HTTPError{http.StatusBadRequest, CodeBadRequest, message}, true
		}
	case trace.ErrorUnauthenticated:
		if e.IsUnauthenticated() {
			return HTTPError{http.StatusUnauthorized, CodeUnauthenticated, "authentication required"}, true
		}
	case trace.ErrorAccessDenied:
		if e.IsAccessDenied() {
			return HTTPError{http.StatusForbidden, CodeAccessDenied, "access denied"}, true
		}
	case trace.ErrorConflict:
		if e.IsConflict() {
			return HTTPError{http.StatusConflict, CodeConflict, message}, true
		}
	case trace.ErrorLimitExceeded:
		if e.IsLimitExceeded() {
			return HTTPError{http.StatusTooManyRequests, CodeLimitExceeded, message}, true
		}
	case trace.ErrorCanceled:
		if e.IsCanceled() {
			return HTTPError{StatusClientClosedRequest, CodeCanceled, "request canceled"}, true
		}
	case trace.ErrorNotImplemented:
		if e.IsNotImplemented() {
			return HTTPError{http.StatusNotImplemented, CodeNotImplemented, message}, true
		}
	case trace.ErrorConnectionProblem:
		if e.IsConnectionProblem() {
			return HTTPError{http.StatusServiceUnavailable, CodeUnavailable, message}, true
		}
	case trace.ErrorTimeout:
		if e.IsTimeout() {
			return HTTPError{http.StatusGatewayTimeout, CodeTimeout, message}, true
		}
	}

	return HTTPError{}, false
}

// @intent read the message this specific error carries, never the message of an outer wrapper.
func linkMessage(link error) string {
	var te *trace.TraceError
	if errors.As(link, &te) {
		return te.Message
	}
	return ""
}

// @intent choose the safe representation with the highest child status for an aggregate failure.
// @domainRule the most severe child wins rather than the first one traversed.
func aggregateHTTPError(aggregate *trace.AggregateError) HTTPError {
	var selected HTTPError
	for _, child := range aggregate.Unwrap() {
		if candidate := ToHTTPError(child); candidate.Status > selected.Status {
			selected = candidate
		}
	}
	if selected.Status == 0 {
		return internalHTTPError()
	}
	return selected
}

// @intent reject malformed classifications and keep server-side detail out of 5xx responses.
func validate(result HTTPError) HTTPError {
	if result.Status < http.StatusBadRequest || result.Status > 599 || result.Code == "" {
		return internalHTTPError()
	}
	if result.Status >= http.StatusInternalServerError {
		result.Message = "internal server error"
	}
	if result.Message == "" {
		result.Message = string(result.Code)
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

// @intent translate trace error categories into HTTP response codes at service boundaries.
// @ensures returns HTTP 200 for nil errors and HTTP 500 for unclassifiable errors.
// GetHTTPStatusCode returns the HTTP status code for an error.
func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	return ToHTTPError(err).Status
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

// @intent let HTTP handlers return errors directly so middleware can centralize response rendering.
// ErrorHandlerFunc is a function that handles HTTP requests and may return an error.
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// @intent retain a best-effort v2 compatibility adapter for standard net/http handlers.
// @domainRule application-owned adapters handle logging, request IDs, and response-write failures.
// @sideEffect attaches method and path fields before writing the HTTP error response.
// @ensures attempts to render handler errors through the logger-free safe response writer.
// ErrorMiddleware converts an ErrorHandlerFunc to a standard http.HandlerFunc.
//
// ErrorMiddleware is a best-effort convenience adapter retained for v2
// compatibility. Because http.HandlerFunc cannot return an error, this adapter
// cannot report failures from WriteError. Applications that need to observe
// response-write failures should own the adapter and call WriteError directly.
// ErrorMiddleware is a candidate for removal in v3.
func ErrorMiddleware(h ErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		err = trace.WithFields(err, map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
		})
		_ = WriteError(w, err, "")
	}
}

// @intent restore typed trace semantics from the public ErrorResponse envelope without deserializing internal trace data.
// @domainRule HTTP 2xx and 3xx statuses are not errors.
// @domainRule malformed, unknown, or status-mismatched responses fail closed without retaining the raw body.
// ReadErrorResponse converts a public HTTP error response into a trace error.
func ReadErrorResponse(statusCode int, responseBody []byte) error {
	if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
		return nil
	}

	frame := trace.CaptureFrame(2)
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
		return &trace.BadParameterError{TraceError: traceError}
	case CodeUnauthenticated:
		return &trace.UnauthenticatedError{TraceError: traceError}
	case CodeAccessDenied:
		return &trace.AccessDeniedError{TraceError: traceError}
	case CodeNotFound:
		return &trace.NotFoundError{TraceError: traceError}
	case CodeAlreadyExists:
		return &trace.AlreadyExistsError{TraceError: traceError}
	case CodeConflict:
		return &trace.ConflictError{TraceError: traceError}
	case CodeLimitExceeded:
		return &trace.LimitExceededError{TraceError: traceError}
	case CodeCanceled:
		return &trace.CanceledError{TraceError: traceError}
	case CodeNotImplemented:
		return &trace.NotImplementedError{TraceError: traceError}
	case CodeUnavailable:
		return &trace.ConnectionProblemError{TraceError: traceError}
	case CodeTimeout:
		return &trace.TimeoutError{TraceError: traceError}
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

// @intent build the trace error that carries the safe message and upstream response metadata.
func newReadErrorResponseTrace(frame trace.Frame, statusCode int, requestID, message string) *trace.TraceError {
	fields := map[string]any{"status_code": statusCode}
	if requestID != "" {
		fields["request_id"] = requestID
	}
	return &trace.TraceError{
		Message: message,
		Frames:  trace.Frames{frame},
		Fields:  fields,
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
		return statusCode == StatusClientClosedRequest
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

// @intent classify an upstream response for developer-facing debugging rather than client rendering.
// @domainRule 2xx responses are not errors, while known status codes map to typed trace errors.
// @ensures records the current call site as the first trace frame and stores upstream status metadata in error fields.
// FromHTTPResponse creates an appropriate error from an HTTP response.
//
// It retains body in the developer-facing error, so the resulting error must not
// be handed to WriteError on a public boundary. Prefer ReadErrorResponse, which
// reads the safe JSON envelope and keeps no raw body.
func FromHTTPResponse(resp *http.Response, body []byte) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	msg := string(body)
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	te := &trace.TraceError{
		Message: msg,
		Frames:  trace.Frames{trace.CaptureFrame(2)},
		Fields: map[string]any{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		},
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return &trace.NotFoundError{TraceError: te}
	case http.StatusConflict:
		return &trace.ConflictError{TraceError: te}
	case http.StatusBadRequest:
		return &trace.BadParameterError{TraceError: te}
	case http.StatusUnauthorized:
		return &trace.UnauthenticatedError{TraceError: te}
	case http.StatusForbidden:
		return &trace.AccessDeniedError{TraceError: te}
	case http.StatusTooManyRequests:
		return &trace.LimitExceededError{TraceError: te}
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return &trace.TimeoutError{TraceError: te}
	case http.StatusServiceUnavailable, http.StatusBadGateway:
		return &trace.ConnectionProblemError{TraceError: te}
	default:
		return &statusError{err: te, status: resp.StatusCode}
	}
}

// @intent test whether an error chain resolves to a specific HTTP status mapping.
// @ensures delegates status resolution to GetHTTPStatusCode.
// IsHTTPError checks if an error corresponds to a specific HTTP status code.
func IsHTTPError(err error, statusCode int) bool {
	return GetHTTPStatusCode(err) == statusCode
}

// @intent override or attach explicit HTTP status semantics to an existing error chain.
// @domainRule returns nil unchanged when the source error is nil.
// @mutates adds http_status metadata to the returned traced wrapper.
// @ensures the returned error accumulates the caller frame on top of the source frames and
// carries the source fields forward, so GetFrames and GetFields stay useful after wrapping.
// WrapHTTPError wraps an error with HTTP status code information.
func WrapHTTPError(err error, statusCode int, msg ...string) error {
	if err == nil {
		return nil
	}

	var message string
	if len(msg) > 0 {
		message = msg[0]
	}

	fields := trace.GetFields(err)
	if fields == nil {
		fields = make(map[string]any)
	}
	fields["http_status"] = statusCode

	return &statusError{
		err: &trace.TraceError{
			Err:     err,
			Message: message,
			Frames:  append(trace.Frames{trace.CaptureFrame(2)}, trace.GetFrames(err)...),
			Fields:  fields,
		},
		status: statusCode,
	}
}

// @intent carry an explicit HTTP status override for errors that do not map to one of the standard typed categories.
type statusError struct {
	err    error
	status int
}

// @intent expose the explicit status override through the single HTTP extension point.
func (e *statusError) HTTPError() HTTPError {
	code := CodeInternal
	if e.status < http.StatusInternalServerError {
		code = CodeBadRequest
	}
	return HTTPError{Status: e.status, Code: code, Message: trace.UserMessage(e.err)}
}

// @intent delegate user-facing string rendering to the wrapped error.
func (e *statusError) Error() string { return e.err.Error() }

// @intent expose the wrapped error to standard Go error traversal.
func (e *statusError) Unwrap() error { return e.err }

// @intent keep the explicit status override alive when trace.WithField or trace.WithFields rebuild the chain.
// @ensures returns ok=false when original is not this wrapper's direct inner TraceError.
// ReplaceTraceError implements trace.TraceErrorReplacer so the status override
// survives field updates.
func (e *statusError) ReplaceTraceError(original, replacement *trace.TraceError) (error, bool) {
	if te, ok := e.err.(*trace.TraceError); ok && te == original {
		return &statusError{err: replacement, status: e.status}, true
	}
	return nil, false
}

// @intent wrap http.Client so transport failures come back as trace-classified errors.
// Client is an HTTP client that wraps errors with trace information.
type Client struct {
	*http.Client
}

// @intent provide an HTTP client wrapper that returns trace-classified transport failures.
// @ensures falls back to http.DefaultClient when no custom client is supplied.
// NewClient creates a new trace-aware HTTP client.
func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{Client: client}
}

// @intent classify outbound HTTP transport failures into timeout or connection problem errors.
// @domainRule deadline exceeded maps to Timeout and other transport failures map to ConnectionProblem.
// @sideEffect executes the underlying HTTP request through the wrapped client.
// Do executes the request and wraps any errors with trace information.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, trace.Timeout(err, fmt.Sprintf("request to %s timed out", req.URL.Host))
		}
		return nil, trace.ConnectionProblem(err, fmt.Sprintf("request to %s failed", req.URL.Host))
	}
	return resp, nil
}
