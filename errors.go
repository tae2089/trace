// @index Typed error categories and retryability rules for trace errors.
package trace

import (
	"fmt"
	"log/slog"
	"strings"
)

// @intent let callers recognize missing-resource semantics through behavior instead of concrete error types.
// ErrorNotFound indicates a resource was not found
type ErrorNotFound interface {
	error
	IsNotFound() bool
}

// @intent let callers recognize duplicate-resource semantics through behavior instead of concrete error types.
// ErrorAlreadyExists indicates a resource already exists
type ErrorAlreadyExists interface {
	error
	IsAlreadyExists() bool
}

// @intent let callers recognize invalid-input semantics through behavior instead of concrete error types.
// ErrorBadParameter indicates invalid input parameters
type ErrorBadParameter interface {
	error
	IsBadParameter() bool
}

// @intent let callers recognize unsupported-operation semantics through behavior instead of concrete error types.
// ErrorNotImplemented indicates functionality is not implemented
type ErrorNotImplemented interface {
	error
	IsNotImplemented() bool
}

// @intent let callers distinguish authentication failures from authorization failures.
// ErrorUnauthenticated indicates authentication is missing or invalid.
type ErrorUnauthenticated interface {
	error
	IsUnauthenticated() bool
}

// @intent let callers recognize authorization failure semantics through behavior instead of concrete error types.
// ErrorAccessDenied indicates access was denied
type ErrorAccessDenied interface {
	error
	IsAccessDenied() bool
}

// @intent let callers recognize state-conflict semantics through behavior instead of concrete error types.
// ErrorConflict indicates a conflict occurred
type ErrorConflict interface {
	error
	IsConflict() bool
}

// @intent let callers recognize transport-failure semantics through behavior instead of concrete error types.
// ErrorConnectionProblem indicates a connection issue
type ErrorConnectionProblem interface {
	error
	IsConnectionProblem() bool
}

// @intent let callers recognize throttling or quota semantics through behavior instead of concrete error types.
// ErrorLimitExceeded indicates a rate limit or quota was exceeded
type ErrorLimitExceeded interface {
	error
	IsLimitExceeded() bool
}

// @intent let callers recognize timeout semantics through behavior instead of concrete error types.
// ErrorTimeout indicates an operation timed out
type ErrorTimeout interface {
	error
	IsTimeout() bool
}

// @intent let callers recognize retry-safe failures through behavior instead of concrete error types.
// ErrorRetryable indicates an error that can be retried
type ErrorRetryable interface {
	error
	IsRetryable() bool
}

// @intent mark failures caused by missing resources so callers can branch on lookup semantics.
// NotFoundError represents a "not found" error
type NotFoundError struct {
	*TraceError
}

// @intent advertise missing-resource semantics for behavior-based error checks.
func (e *NotFoundError) IsNotFound() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *NotFoundError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *NotFoundError) Unwrap() error { return e.TraceError }

// @intent reuse TraceError structured logging output for not-found errors.
func (e *NotFoundError) LogValue() slog.Value { return e.TraceError.LogValue() }

// @intent mark failures caused by uniqueness or duplicate-resource conflicts.
// AlreadyExistsError represents an "already exists" error
type AlreadyExistsError struct {
	*TraceError
}

// @intent advertise duplicate-resource semantics for behavior-based error checks.
func (e *AlreadyExistsError) IsAlreadyExists() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *AlreadyExistsError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *AlreadyExistsError) Unwrap() error { return e.TraceError }

// @intent mark failures caused by invalid caller input or malformed parameters.
// BadParameterError represents an invalid parameter error
type BadParameterError struct {
	*TraceError
}

// @intent advertise invalid-input semantics for behavior-based error checks.
func (e *BadParameterError) IsBadParameter() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *BadParameterError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *BadParameterError) Unwrap() error { return e.TraceError }

// @intent mark code paths that are recognized but intentionally unsupported.
// NotImplementedError represents a "not implemented" error
type NotImplementedError struct {
	*TraceError
}

// @intent advertise unsupported-operation semantics for behavior-based error checks.
func (e *NotImplementedError) IsNotImplemented() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *NotImplementedError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *NotImplementedError) Unwrap() error { return e.TraceError }

// @intent mark authentication failures independently from authorization failures.
// UnauthenticatedError represents a missing or invalid authentication identity.
type UnauthenticatedError struct {
	*TraceError
}

// @intent advertise authentication-failure semantics for behavior-based checks.
func (e *UnauthenticatedError) IsUnauthenticated() bool { return true }

// @intent delegate developer-facing rendering to the embedded TraceError.
func (e *UnauthenticatedError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *UnauthenticatedError) Unwrap() error { return e.TraceError }

// @intent mark authorization failures so callers can deny access consistently.
// AccessDeniedError represents an access denied error
type AccessDeniedError struct {
	*TraceError
}

// @intent advertise authorization-failure semantics for behavior-based error checks.
func (e *AccessDeniedError) IsAccessDenied() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *AccessDeniedError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *AccessDeniedError) Unwrap() error { return e.TraceError }

// @intent mark state conflicts that block the requested operation until data changes.
// ConflictError represents a conflict error
type ConflictError struct {
	*TraceError
}

// @intent advertise state-conflict semantics for behavior-based error checks.
func (e *ConflictError) IsConflict() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *ConflictError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *ConflictError) Unwrap() error { return e.TraceError }

// @intent mark infrastructure or transport failures that may succeed on retry.
// ConnectionProblemError represents a connection error
type ConnectionProblemError struct {
	*TraceError
}

// @intent advertise transport-failure semantics for behavior-based error checks.
func (e *ConnectionProblemError) IsConnectionProblem() bool { return true }

// @intent advertise retry-safe semantics for transient connection failures.
func (e *ConnectionProblemError) IsRetryable() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *ConnectionProblemError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *ConnectionProblemError) Unwrap() error { return e.TraceError }

// @intent mark throttling or quota failures so callers can apply backoff behavior.
// LimitExceededError represents a rate limit or quota exceeded error
type LimitExceededError struct {
	*TraceError
}

// @intent advertise throttling semantics for behavior-based error checks.
func (e *LimitExceededError) IsLimitExceeded() bool { return true }

// @intent advertise retry-safe semantics for throttled operations.
func (e *LimitExceededError) IsRetryable() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *LimitExceededError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *LimitExceededError) Unwrap() error { return e.TraceError }

// @intent mark operations that exceeded their allowed completion window.
// TimeoutError represents a timeout error
type TimeoutError struct {
	*TraceError
}

// @intent advertise timeout semantics for behavior-based error checks.
func (e *TimeoutError) IsTimeout() bool { return true }

// @intent advertise retry-safe semantics for timeout failures.
func (e *TimeoutError) IsRetryable() bool { return true }

// @intent delegate user-facing string rendering to the embedded TraceError.
func (e *TimeoutError) Error() string { return e.TraceError.Error() }

// @intent expose the embedded TraceError to standard Go error traversal.
func (e *TimeoutError) Unwrap() error { return e.TraceError }

// @intent centralize typed-error wrapping so constructors preserve prior trace frames and structured fields.
// @ensures returns a TraceError that prepends the supplied frame and carries forward existing fields.
func wrapTypedInternal(err error, msg string, frame Frame) *TraceError {
	var existingFrames Frames
	var existingFields map[string]any
	if te := findTraceError(err); te != nil {
		existingFrames = te.Frames
		if len(te.Fields) > 0 {
			existingFields = copyFields(te.Fields)
		}
	}
	return &TraceError{
		Err:     err,
		Message: msg,
		Frames:  append(Frames{frame}, existingFrames...),
		Fields:  existingFields,
	}
}

// @intent classify a missing resource so callers can branch on lookup failure semantics.
// @domainRule not found errors are classified as HTTP 404 by the tracehttp package.
// @ensures records the current call site as the first trace frame.
// NotFound creates a new NotFoundError
func NotFound(msgAndArgs ...any) error {
	frame := CaptureFrame(2)
	return &NotFoundError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
		},
	}
}

// @intent preserve an existing cause while reclassifying it as a missing-resource failure.
// @domainRule returns nil unchanged when the source error is nil.
// @ensures prepends the current call site to any existing trace frames on the returned error.
// WrapNotFound wraps an error as NotFoundError
func WrapNotFound(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := CaptureFrame(2)
	return &NotFoundError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// @intent classify duplicate-resource failures so callers can branch on uniqueness semantics.
// @domainRule already exists errors are classified as HTTP 409 by the tracehttp package.
// @ensures records the current call site as the first trace frame.
// AlreadyExists creates a new AlreadyExistsError
func AlreadyExists(msgAndArgs ...any) error {
	frame := CaptureFrame(2)
	return &AlreadyExistsError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
		},
	}
}

// @intent preserve an existing cause while reclassifying it as a duplicate-resource failure.
// @domainRule returns nil unchanged when the source error is nil.
// @ensures prepends the current call site to any existing trace frames on the returned error.
// WrapAlreadyExists wraps an error as AlreadyExistsError
func WrapAlreadyExists(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := CaptureFrame(2)
	return &AlreadyExistsError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// @intent classify invalid input so callers can branch on client-side request errors.
// @domainRule bad parameter errors are classified as HTTP 400 by the tracehttp package.
// @ensures records the current call site as the first trace frame.
// BadParameter creates a new BadParameterError
func BadParameter(msgAndArgs ...any) error {
	frame := CaptureFrame(2)
	return &BadParameterError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
		},
	}
}

// @intent preserve an existing cause while surfacing it as a client input failure.
// @domainRule returns nil unchanged when the source error is nil.
// @ensures prepends the current call site to any existing trace frames on the returned error.
// WrapBadParameter wraps an error as BadParameterError
func WrapBadParameter(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := CaptureFrame(2)
	return &BadParameterError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// @intent classify unsupported behavior so callers can surface capability gaps consistently.
// @domainRule not implemented errors are classified as HTTP 501 by the tracehttp package.
// @ensures records the current call site as the first trace frame.
// NotImplemented creates a new NotImplementedError
func NotImplemented(msgAndArgs ...any) error {
	frame := CaptureFrame(2)
	return &NotImplementedError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
		},
	}
}

// @intent classify authentication failures separately from authorization failures.
// @domainRule unauthenticated errors map to HTTP 401 and expose only a fixed client message.
// @ensures records the current call site and keeps the supplied message for diagnostics only.
// Unauthenticated creates a new UnauthenticatedError.
func Unauthenticated(msgAndArgs ...any) error {
	frame := CaptureFrame(2)
	return &UnauthenticatedError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
		},
	}
}

// @intent preserve a lower-level cause while classifying an authentication failure.
// @domainRule returns nil unchanged when the source error is nil.
// WrapUnauthenticated wraps an error as UnauthenticatedError.
func WrapUnauthenticated(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := CaptureFrame(2)
	return &UnauthenticatedError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// @intent classify authorization failures so callers can deny access consistently.
// @domainRule access denied errors are classified as HTTP 403 by the tracehttp package.
// @ensures records the current call site as the first trace frame.
// AccessDenied creates a new AccessDeniedError
func AccessDenied(msgAndArgs ...any) error {
	frame := CaptureFrame(2)
	return &AccessDeniedError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
		},
	}
}

// @intent preserve a lower-level cause while surfacing it as an authorization failure.
// @domainRule returns nil unchanged when the source error is nil.
// @ensures prepends the current call site to any existing trace frames on the returned error.
// WrapAccessDenied wraps an error as AccessDeniedError
func WrapAccessDenied(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := CaptureFrame(2)
	return &AccessDeniedError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// @intent classify state mismatches that prevent the requested operation from succeeding.
// @domainRule conflict errors are classified as HTTP 409 by the tracehttp package.
// @ensures records the current call site as the first trace frame.
// Conflict creates a new ConflictError
func Conflict(msgAndArgs ...any) error {
	frame := CaptureFrame(2)
	return &ConflictError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
		},
	}
}

// @intent mark infrastructure or network failures as transient connection problems.
// @domainRule connection problems are retryable and map to HTTP 503.
// @ensures prepends the current call site to any existing trace frames on the returned error.
// @ensures returns nil unchanged when the source error is nil.
// ConnectionProblem creates a new ConnectionProblemError.
// If err is nil, returns nil.
func ConnectionProblem(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := CaptureFrame(2)
	return &ConnectionProblemError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// @intent classify quota or rate-limit failures so callers can branch on throttling semantics.
// @domainRule limit exceeded errors are retryable and map to HTTP 429.
// @ensures records the current call site as the first trace frame.
// LimitExceeded creates a new LimitExceededError
func LimitExceeded(msgAndArgs ...any) error {
	frame := CaptureFrame(2)
	return &LimitExceededError{
		TraceError: &TraceError{
			Message: formatMessage(msgAndArgs...),
			Frames:  Frames{frame},
		},
	}
}

// @intent reclassify an existing failure as a quota or rate-limit failure while keeping its cause.
// @domainRule returns nil unchanged when the source error is nil.
// @ensures prepends the current call site to any existing trace frames on the returned error.
// WrapLimitExceeded wraps an existing error as a LimitExceededError.
func WrapLimitExceeded(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := CaptureFrame(2)
	return &LimitExceededError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// @intent classify operations that exceeded their allowed completion window.
// @domainRule timeout errors are retryable and map to HTTP 504.
// @ensures prepends the current call site to any existing trace frames on the returned error.
// @ensures returns nil unchanged when the source error is nil.
// Timeout creates a new TimeoutError.
// If err is nil, returns nil.
func Timeout(err error, msgAndArgs ...any) error {
	if err == nil {
		return nil
	}
	frame := CaptureFrame(2)
	return &TimeoutError{
		TraceError: wrapTypedInternal(err, formatMessage(msgAndArgs...), frame),
	}
}

// @intent detect missing-resource failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorNotFound
	return chainAs(err, &e) && e.IsNotFound()
}

// @intent detect duplicate-resource failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsAlreadyExists checks if error is an already exists error
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorAlreadyExists
	return chainAs(err, &e) && e.IsAlreadyExists()
}

// @intent detect caller-input failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsBadParameter checks if error is a bad parameter error
func IsBadParameter(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorBadParameter
	return chainAs(err, &e) && e.IsBadParameter()
}

// @intent detect unsupported-operation failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsNotImplemented checks if error is a not implemented error
func IsNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorNotImplemented
	return chainAs(err, &e) && e.IsNotImplemented()
}

// @intent detect authentication failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsUnauthenticated checks if error is an unauthenticated error.
func IsUnauthenticated(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorUnauthenticated
	return chainAs(err, &e) && e.IsUnauthenticated()
}

// @intent detect authorization failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsAccessDenied checks if error is an access denied error
func IsAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorAccessDenied
	return chainAs(err, &e) && e.IsAccessDenied()
}

// @intent detect state-conflict failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsConflict checks if error is a conflict error
func IsConflict(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorConflict
	return chainAs(err, &e) && e.IsConflict()
}

// @intent detect transient transport or infrastructure failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsConnectionProblem checks if error is a connection problem
func IsConnectionProblem(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorConnectionProblem
	return chainAs(err, &e) && e.IsConnectionProblem()
}

// @intent detect throttling or quota failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsLimitExceeded checks if error is a limit exceeded error
func IsLimitExceeded(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorLimitExceeded
	return chainAs(err, &e) && e.IsLimitExceeded()
}

// @intent detect timeout failures anywhere in an error chain.
// @ensures returns false for nil errors.
// IsTimeout checks if error is a timeout error
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorTimeout
	return chainAs(err, &e) && e.IsTimeout()
}

// @intent detect failures that explicitly advertise retry-safe semantics.
// @ensures returns false for nil errors.
// IsRetryable checks if error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var e ErrorRetryable
	return chainAs(err, &e) && e.IsRetryable()
}

// @intent preserve multiple concurrent failures as one error value for later inspection.
// @domainRule nil errors are discarded so only real failures participate in aggregation.
// @ensures returns nil for an all-nil input set and the sole error unchanged for a single failure.
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

// @intent preserve multiple related failures as one traversable error tree.
// AggregateError holds multiple errors
type AggregateError struct {
	Errs []error
}

// @intent render a readable summary that includes every child error in the aggregate.
// @ensures includes the number of collected errors in the output.
func (e *AggregateError) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("multiple errors (%d):\n", len(e.Errs)))
	for i, err := range e.Errs {
		fmt.Fprintf(&b, "  [%d] %s\n", i, err.Error())
	}
	return b.String()
}

// @intent expose all child errors so callers can traverse the aggregate tree.
// Unwrap returns the list of errors (Go 1.20+ multiple error unwrapping)
func (e *AggregateError) Unwrap() []error {
	return e.Errs
}
