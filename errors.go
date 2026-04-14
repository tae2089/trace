package trace

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// Error type interfaces for behavior-based error checking
type (
	// ErrorNotFound indicates a resource was not found
	ErrorNotFound interface {
		error
		IsNotFound() bool
	}

	// ErrorAlreadyExists indicates a resource already exists
	ErrorAlreadyExists interface {
		error
		IsAlreadyExists() bool
	}

	// ErrorBadParameter indicates invalid input parameters
	ErrorBadParameter interface {
		error
		IsBadParameter() bool
	}

	// ErrorNotImplemented indicates functionality is not implemented
	ErrorNotImplemented interface {
		error
		IsNotImplemented() bool
	}

	// ErrorAccessDenied indicates access was denied
	ErrorAccessDenied interface {
		error
		IsAccessDenied() bool
	}

	// ErrorConflict indicates a conflict occurred
	ErrorConflict interface {
		error
		IsConflict() bool
	}

	// ErrorConnectionProblem indicates a connection issue
	ErrorConnectionProblem interface {
		error
		IsConnectionProblem() bool
	}

	// ErrorLimitExceeded indicates a rate limit or quota was exceeded
	ErrorLimitExceeded interface {
		error
		IsLimitExceeded() bool
	}

	// ErrorTimeout indicates an operation timed out
	ErrorTimeout interface {
		error
		IsTimeout() bool
	}

	// ErrorRetryable indicates an error that can be retried
	ErrorRetryable interface {
		error
		IsRetryable() bool
	}
)

// HTTPStatusCode is an interface for errors that can provide HTTP status codes
type HTTPStatusCode interface {
	HTTPStatusCode() int
}

// NotFoundError represents a "not found" error
type NotFoundError struct {
	*TraceError
}

func (e *NotFoundError) IsNotFound() bool     { return true }
func (e *NotFoundError) HTTPStatusCode() int  { return http.StatusNotFound }
func (e *NotFoundError) Error() string        { return e.TraceError.Error() }
func (e *NotFoundError) Unwrap() error        { return e.TraceError }
func (e *NotFoundError) LogValue() slog.Value { return e.TraceError.LogValue() }

// AlreadyExistsError represents an "already exists" error
type AlreadyExistsError struct {
	*TraceError
}

func (e *AlreadyExistsError) IsAlreadyExists() bool { return true }
func (e *AlreadyExistsError) HTTPStatusCode() int   { return http.StatusConflict }
func (e *AlreadyExistsError) Error() string         { return e.TraceError.Error() }
func (e *AlreadyExistsError) Unwrap() error         { return e.TraceError }

// BadParameterError represents an invalid parameter error
type BadParameterError struct {
	*TraceError
}

func (e *BadParameterError) IsBadParameter() bool { return true }
func (e *BadParameterError) HTTPStatusCode() int  { return http.StatusBadRequest }
func (e *BadParameterError) Error() string        { return e.TraceError.Error() }
func (e *BadParameterError) Unwrap() error        { return e.TraceError }

// NotImplementedError represents a "not implemented" error
type NotImplementedError struct {
	*TraceError
}

func (e *NotImplementedError) IsNotImplemented() bool { return true }
func (e *NotImplementedError) HTTPStatusCode() int    { return http.StatusNotImplemented }
func (e *NotImplementedError) Error() string          { return e.TraceError.Error() }
func (e *NotImplementedError) Unwrap() error          { return e.TraceError }

// AccessDeniedError represents an access denied error
type AccessDeniedError struct {
	*TraceError
}

func (e *AccessDeniedError) IsAccessDenied() bool { return true }
func (e *AccessDeniedError) HTTPStatusCode() int  { return http.StatusForbidden }
func (e *AccessDeniedError) Error() string        { return e.TraceError.Error() }
func (e *AccessDeniedError) Unwrap() error        { return e.TraceError }

// ConflictError represents a conflict error
type ConflictError struct {
	*TraceError
}

func (e *ConflictError) IsConflict() bool    { return true }
func (e *ConflictError) HTTPStatusCode() int { return http.StatusConflict }
func (e *ConflictError) Error() string       { return e.TraceError.Error() }
func (e *ConflictError) Unwrap() error       { return e.TraceError }

// ConnectionProblemError represents a connection error
type ConnectionProblemError struct {
	*TraceError
}

func (e *ConnectionProblemError) IsConnectionProblem() bool { return true }
func (e *ConnectionProblemError) IsRetryable() bool         { return true }
func (e *ConnectionProblemError) HTTPStatusCode() int       { return http.StatusServiceUnavailable }
func (e *ConnectionProblemError) Error() string             { return e.TraceError.Error() }
func (e *ConnectionProblemError) Unwrap() error             { return e.TraceError }

// LimitExceededError represents a rate limit or quota exceeded error
type LimitExceededError struct {
	*TraceError
}

func (e *LimitExceededError) IsLimitExceeded() bool { return true }
func (e *LimitExceededError) IsRetryable() bool     { return true }
func (e *LimitExceededError) HTTPStatusCode() int   { return http.StatusTooManyRequests }
func (e *LimitExceededError) Error() string         { return e.TraceError.Error() }
func (e *LimitExceededError) Unwrap() error         { return e.TraceError }

// TimeoutError represents a timeout error
type TimeoutError struct {
	*TraceError
}

func (e *TimeoutError) IsTimeout() bool     { return true }
func (e *TimeoutError) IsRetryable() bool   { return true }
func (e *TimeoutError) HTTPStatusCode() int { return http.StatusGatewayTimeout }
func (e *TimeoutError) Error() string       { return e.TraceError.Error() }
func (e *TimeoutError) Unwrap() error       { return e.TraceError }

// Import slog for LogValue - already imported at top

func wrapTypedInternal(err error, msg string, frame Frame) *TraceError {
	var existingFrames Frames
	existingFields := make(map[string]any)
	var te *TraceError
	if err != nil && errors.As(err, &te) {
		existingFrames = te.Frames
		for k, v := range te.Fields {
			existingFields[k] = v
		}
	}
	return &TraceError{
		Err:     err,
		Message: msg,
		Frames:  append(Frames{frame}, existingFrames...),
		Fields:  existingFields,
	}
}

// Constructor functions for typed errors

// NotFound creates a new NotFoundError
func NotFound(msgAndArgs ...any) error {
	frame := captureFrame(2)
	return &NotFoundError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
			Fields:  make(map[string]any),
		},
	}
}

// WrapNotFound wraps an error as NotFoundError
func WrapNotFound(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := captureFrame(2)
	return &NotFoundError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// AlreadyExists creates a new AlreadyExistsError
func AlreadyExists(msgAndArgs ...any) error {
	frame := captureFrame(2)
	return &AlreadyExistsError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
			Fields:  make(map[string]any),
		},
	}
}

// WrapAlreadyExists wraps an error as AlreadyExistsError
func WrapAlreadyExists(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := captureFrame(2)
	return &AlreadyExistsError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// BadParameter creates a new BadParameterError
func BadParameter(msgAndArgs ...any) error {
	frame := captureFrame(2)
	return &BadParameterError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
			Fields:  make(map[string]any),
		},
	}
}

// WrapBadParameter wraps an error as BadParameterError
func WrapBadParameter(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := captureFrame(2)
	return &BadParameterError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// NotImplemented creates a new NotImplementedError
func NotImplemented(msgAndArgs ...any) error {
	frame := captureFrame(2)
	return &NotImplementedError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
			Fields:  make(map[string]any),
		},
	}
}

// AccessDenied creates a new AccessDeniedError
func AccessDenied(msgAndArgs ...any) error {
	frame := captureFrame(2)
	return &AccessDeniedError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
			Fields:  make(map[string]any),
		},
	}
}

// WrapAccessDenied wraps an error as AccessDeniedError
func WrapAccessDenied(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := captureFrame(2)
	return &AccessDeniedError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// Conflict creates a new ConflictError
func Conflict(msgAndArgs ...any) error {
	frame := captureFrame(2)
	return &ConflictError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
			Fields:  make(map[string]any),
		},
	}
}

// ConnectionProblem creates a new ConnectionProblemError.
// If err is nil, returns nil.
func ConnectionProblem(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := captureFrame(2)
	return &ConnectionProblemError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// LimitExceeded creates a new LimitExceededError
func LimitExceeded(msgAndArgs ...any) error {
	frame := captureFrame(2)
	return &LimitExceededError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
			Fields:  make(map[string]any),
		},
	}
}

// Timeout creates a new TimeoutError.
// If err is nil, returns nil.
func Timeout(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := captureFrame(2)
	return &TimeoutError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// Error type checking functions using behavior-based approach

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorNotFound
	return errors.As(err, &e) && e.IsNotFound()
}

// IsAlreadyExists checks if error is an already exists error
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorAlreadyExists
	return errors.As(err, &e) && e.IsAlreadyExists()
}

// IsBadParameter checks if error is a bad parameter error
func IsBadParameter(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorBadParameter
	return errors.As(err, &e) && e.IsBadParameter()
}

// IsNotImplemented checks if error is a not implemented error
func IsNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorNotImplemented
	return errors.As(err, &e) && e.IsNotImplemented()
}

// IsAccessDenied checks if error is an access denied error
func IsAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorAccessDenied
	return errors.As(err, &e) && e.IsAccessDenied()
}

// IsConflict checks if error is a conflict error
func IsConflict(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorConflict
	return errors.As(err, &e) && e.IsConflict()
}

// IsConnectionProblem checks if error is a connection problem
func IsConnectionProblem(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorConnectionProblem
	return errors.As(err, &e) && e.IsConnectionProblem()
}

// IsLimitExceeded checks if error is a limit exceeded error
func IsLimitExceeded(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorLimitExceeded
	return errors.As(err, &e) && e.IsLimitExceeded()
}

// IsTimeout checks if error is a timeout error
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorTimeout
	return errors.As(err, &e) && e.IsTimeout()
}

// IsRetryable checks if error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorRetryable
	return errors.As(err, &e) && e.IsRetryable()
}

// GetHTTPStatusCode returns the HTTP status code for an error
func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var e HTTPStatusCode
	if errors.As(err, &e) {
		return e.HTTPStatusCode()
	}
	return http.StatusInternalServerError
}

// Aggregate combines multiple errors into a single error using errors.Join (Go 1.20+)
func Aggregate(errs ...error) error {
	// Filter out nil errors
	var nonNil []error
	for _, err := range errs {
		if err != nil {
			nonNil = append(nonNil, err)
		}
	}

	if len(nonNil) == 0 {
		return nil
	}

	if len(nonNil) == 1 {
		return nonNil[0]
	}

	return &AggregateError{
		Errs: nonNil,
	}
}

// AggregateError holds multiple errors
type AggregateError struct {
	Errs []error
}

func (e *AggregateError) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("multiple errors (%d):\n", len(e.Errs)))
	for i, err := range e.Errs {
		fmt.Fprintf(&b, "  [%d] %s\n", i, err.Error())
	}
	return b.String()
}

// Unwrap returns the list of errors (Go 1.20+ multiple error unwrapping)
func (e *AggregateError) Unwrap() []error {
	return e.Errs
}

// HTTPStatusCode returns the most severe HTTP status code
func (e *AggregateError) HTTPStatusCode() int {
	code := http.StatusOK
	for _, err := range e.Errs {
		if c := GetHTTPStatusCode(err); c > code {
			code = c
		}
	}
	if code == http.StatusOK {
		return http.StatusInternalServerError
	}
	return code
}
