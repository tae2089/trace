<!-- generated-by: code-context-graph docs -->
# errors.go

> Typed error categories and retryability rules for trace errors.

## Functions

### IsNotFound
- **Lines:** 95–95
- **Intent:** advertise missing-resource semantics for behavior-based error checks.

### Error
- **Lines:** 98–98
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 101–101
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### LogValue
- **Lines:** 104–104
- **Intent:** reuse TraceError structured logging output for not-found errors.
- **Calls:** LogValue

### IsAlreadyExists
- **Lines:** 113–113
- **Intent:** advertise duplicate-resource semantics for behavior-based error checks.

### Error
- **Lines:** 116–116
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 119–119
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsBadParameter
- **Lines:** 128–128
- **Intent:** advertise invalid-input semantics for behavior-based error checks.

### Error
- **Lines:** 131–131
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 134–134
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsNotImplemented
- **Lines:** 143–143
- **Intent:** advertise unsupported-operation semantics for behavior-based error checks.

### Error
- **Lines:** 146–146
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 149–149
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsUnauthenticated
- **Lines:** 158–158
- **Intent:** advertise authentication-failure semantics for behavior-based checks.

### Error
- **Lines:** 161–161
- **Intent:** delegate developer-facing rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 164–164
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsAccessDenied
- **Lines:** 173–173
- **Intent:** advertise authorization-failure semantics for behavior-based error checks.

### Error
- **Lines:** 176–176
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 179–179
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsConflict
- **Lines:** 188–188
- **Intent:** advertise state-conflict semantics for behavior-based error checks.

### Error
- **Lines:** 191–191
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 194–194
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsConnectionProblem
- **Lines:** 203–203
- **Intent:** advertise transport-failure semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 206–206
- **Intent:** advertise retry-safe semantics for transient connection failures.

### Error
- **Lines:** 209–209
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 212–212
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsLimitExceeded
- **Lines:** 221–221
- **Intent:** advertise throttling semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 224–224
- **Intent:** advertise retry-safe semantics for throttled operations.

### Error
- **Lines:** 227–227
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 230–230
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsTimeout
- **Lines:** 239–239
- **Intent:** advertise timeout semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 242–242
- **Intent:** advertise retry-safe semantics for timeout failures.

### Error
- **Lines:** 245–245
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 248–248
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### wrapTypedInternal
- **Lines:** 252–268
- **Intent:** centralize typed-error wrapping so constructors preserve prior trace frames and structured fields.
- **Ensures:**
  - returns a TraceError that prepends the supplied frame and carries forward existing fields.
- **Calls:** As

### NotFound
- **Lines:** 275–284
- **Intent:** classify a missing resource so callers can branch on lookup failure semantics.
- **Domain Rules:**
  - not found errors are classified as HTTP 404 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
  - returns a NotFoundError with initialized structured fields.
NotFound creates a new NotFoundError
- **Calls:** CaptureFrame, formatMessage

### WrapNotFound
- **Lines:** 290–298
- **Intent:** preserve an existing cause while reclassifying it as a missing-resource failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapNotFound wraps an error as NotFoundError
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### AlreadyExists
- **Lines:** 304–313
- **Intent:** classify duplicate-resource failures so callers can branch on uniqueness semantics.
- **Domain Rules:**
  - already exists errors are classified as HTTP 409 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
AlreadyExists creates a new AlreadyExistsError
- **Calls:** CaptureFrame, formatMessage

### WrapAlreadyExists
- **Lines:** 319–327
- **Intent:** preserve an existing cause while reclassifying it as a duplicate-resource failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapAlreadyExists wraps an error as AlreadyExistsError
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### BadParameter
- **Lines:** 333–342
- **Intent:** classify invalid input so callers can branch on client-side request errors.
- **Domain Rules:**
  - bad parameter errors are classified as HTTP 400 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
BadParameter creates a new BadParameterError
- **Calls:** CaptureFrame, formatMessage

### WrapBadParameter
- **Lines:** 348–356
- **Intent:** preserve an existing cause while surfacing it as a client input failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapBadParameter wraps an error as BadParameterError
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### NotImplemented
- **Lines:** 362–371
- **Intent:** classify unsupported behavior so callers can surface capability gaps consistently.
- **Domain Rules:**
  - not implemented errors are classified as HTTP 501 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
NotImplemented creates a new NotImplementedError
- **Calls:** CaptureFrame, formatMessage

### Unauthenticated
- **Lines:** 377–386
- **Intent:** classify authentication failures separately from authorization failures.
- **Domain Rules:**
  - unauthenticated errors map to HTTP 401 and expose only a fixed client message.
- **Ensures:**
  - records the current call site and keeps the supplied message for diagnostics only.
Unauthenticated creates a new UnauthenticatedError.
- **Calls:** CaptureFrame, formatMessage

### WrapUnauthenticated
- **Lines:** 391–399
- **Intent:** preserve a lower-level cause while classifying an authentication failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
WrapUnauthenticated wraps an error as UnauthenticatedError.
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### AccessDenied
- **Lines:** 406–415
- **Intent:** classify authorization failures so callers can deny access consistently.
- **Domain Rules:**
  - access denied errors are classified as HTTP 403 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
  - returns an AccessDeniedError with initialized structured fields.
AccessDenied creates a new AccessDeniedError
- **Calls:** CaptureFrame, formatMessage

### WrapAccessDenied
- **Lines:** 421–429
- **Intent:** preserve a lower-level cause while surfacing it as an authorization failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapAccessDenied wraps an error as AccessDeniedError
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### Conflict
- **Lines:** 435–444
- **Intent:** classify state mismatches that prevent the requested operation from succeeding.
- **Domain Rules:**
  - conflict errors are classified as HTTP 409 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
Conflict creates a new ConflictError
- **Calls:** CaptureFrame, formatMessage

### ConnectionProblem
- **Lines:** 452–460
- **Intent:** mark infrastructure or network failures as transient connection problems.
- **Domain Rules:**
  - connection problems are retryable and map to HTTP 503.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
  - returns nil unchanged when the source error is nil.
ConnectionProblem creates a new ConnectionProblemError.
If err is nil, returns nil.
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### LimitExceeded
- **Lines:** 466–475
- **Intent:** classify quota or rate-limit failures so callers can branch on throttling semantics.
- **Domain Rules:**
  - limit exceeded errors are retryable and map to HTTP 429.
- **Ensures:**
  - records the current call site as the first trace frame.
LimitExceeded creates a new LimitExceededError
- **Calls:** CaptureFrame, formatMessage

### WrapLimitExceeded
- **Lines:** 481–489
- **Intent:** reclassify an existing failure as a quota or rate-limit failure while keeping its cause.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapLimitExceeded wraps an existing error as a LimitExceededError.
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### Timeout
- **Lines:** 497–505
- **Intent:** classify operations that exceeded their allowed completion window.
- **Domain Rules:**
  - timeout errors are retryable and map to HTTP 504.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
  - returns nil unchanged when the source error is nil.
Timeout creates a new TimeoutError.
If err is nil, returns nil.
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### IsNotFound
- **Lines:** 510–516
- **Intent:** detect missing-resource failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsNotFound checks if error is a not found error
- **Calls:** IsNotFound, As

### IsAlreadyExists
- **Lines:** 521–527
- **Intent:** detect duplicate-resource failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsAlreadyExists checks if error is an already exists error
- **Calls:** IsAlreadyExists, As

### IsBadParameter
- **Lines:** 532–538
- **Intent:** detect caller-input failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsBadParameter checks if error is a bad parameter error
- **Calls:** IsBadParameter, As

### IsNotImplemented
- **Lines:** 543–549
- **Intent:** detect unsupported-operation failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsNotImplemented checks if error is a not implemented error
- **Calls:** IsNotImplemented, As

### IsUnauthenticated
- **Lines:** 554–560
- **Intent:** detect authentication failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsUnauthenticated checks if error is an unauthenticated error.
- **Calls:** IsUnauthenticated, As

### IsAccessDenied
- **Lines:** 565–571
- **Intent:** detect authorization failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsAccessDenied checks if error is an access denied error
- **Calls:** IsAccessDenied, As

### IsConflict
- **Lines:** 576–582
- **Intent:** detect state-conflict failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsConflict checks if error is a conflict error
- **Calls:** IsConflict, As

### IsConnectionProblem
- **Lines:** 587–593
- **Intent:** detect transient transport or infrastructure failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsConnectionProblem checks if error is a connection problem
- **Calls:** IsConnectionProblem, As

### IsLimitExceeded
- **Lines:** 598–604
- **Intent:** detect throttling or quota failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsLimitExceeded checks if error is a limit exceeded error
- **Calls:** IsLimitExceeded, As

### IsTimeout
- **Lines:** 609–615
- **Intent:** detect timeout failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsTimeout checks if error is a timeout error
- **Calls:** IsTimeout, As

### IsRetryable
- **Lines:** 620–626
- **Intent:** detect failures that explicitly advertise retry-safe semantics.
- **Ensures:**
  - returns false for nil errors.
IsRetryable checks if error is retryable
- **Calls:** IsRetryable, As

### Aggregate
- **Lines:** 632–652
- **Intent:** preserve multiple concurrent failures as one error value for later inspection.
- **Domain Rules:**
  - nil errors are discarded so only real failures participate in aggregation.
- **Ensures:**
  - returns nil for an all-nil input set and the sole error unchanged for a single failure.
Aggregate combines multiple errors into a single error using errors.Join (Go 1.20+)

### Error
- **Lines:** 662–669
- **Intent:** render a readable summary that includes every child error in the aggregate.
- **Ensures:**
  - includes the number of collected errors in the output.
- **Calls:** Error

### Unwrap
- **Lines:** 673–675
- **Intent:** expose all child errors so callers can traverse the aggregate tree.
Unwrap returns the list of errors (Go 1.20+ multiple error unwrapping)

## Classes

### NotFoundError
- **Lines:** 90–92
- **Intent:** mark failures caused by missing resources so callers can branch on lookup semantics.
NotFoundError represents a "not found" error

### AlreadyExistsError
- **Lines:** 108–110
- **Intent:** mark failures caused by uniqueness or duplicate-resource conflicts.
AlreadyExistsError represents an "already exists" error

### BadParameterError
- **Lines:** 123–125
- **Intent:** mark failures caused by invalid caller input or malformed parameters.
BadParameterError represents an invalid parameter error

### NotImplementedError
- **Lines:** 138–140
- **Intent:** mark code paths that are recognized but intentionally unsupported.
NotImplementedError represents a "not implemented" error

### UnauthenticatedError
- **Lines:** 153–155
- **Intent:** mark authentication failures independently from authorization failures.
UnauthenticatedError represents a missing or invalid authentication identity.

### AccessDeniedError
- **Lines:** 168–170
- **Intent:** mark authorization failures so callers can deny access consistently.
AccessDeniedError represents an access denied error

### ConflictError
- **Lines:** 183–185
- **Intent:** mark state conflicts that block the requested operation until data changes.
ConflictError represents a conflict error

### ConnectionProblemError
- **Lines:** 198–200
- **Intent:** mark infrastructure or transport failures that may succeed on retry.
ConnectionProblemError represents a connection error

### LimitExceededError
- **Lines:** 216–218
- **Intent:** mark throttling or quota failures so callers can apply backoff behavior.
LimitExceededError represents a rate limit or quota exceeded error

### TimeoutError
- **Lines:** 234–236
- **Intent:** mark operations that exceeded their allowed completion window.
TimeoutError represents a timeout error

### AggregateError
- **Lines:** 656–658
- **Intent:** preserve multiple related failures as one traversable error tree.
AggregateError holds multiple errors

## Types

### ErrorNotFound
- **Lines:** 13–16
- **Intent:** let callers recognize missing-resource semantics through behavior instead of concrete error types.
ErrorNotFound indicates a resource was not found

### ErrorAlreadyExists
- **Lines:** 20–23
- **Intent:** let callers recognize duplicate-resource semantics through behavior instead of concrete error types.
ErrorAlreadyExists indicates a resource already exists

### ErrorBadParameter
- **Lines:** 27–30
- **Intent:** let callers recognize invalid-input semantics through behavior instead of concrete error types.
ErrorBadParameter indicates invalid input parameters

### ErrorNotImplemented
- **Lines:** 34–37
- **Intent:** let callers recognize unsupported-operation semantics through behavior instead of concrete error types.
ErrorNotImplemented indicates functionality is not implemented

### ErrorUnauthenticated
- **Lines:** 41–44
- **Intent:** let callers distinguish authentication failures from authorization failures.
ErrorUnauthenticated indicates authentication is missing or invalid.

### ErrorAccessDenied
- **Lines:** 48–51
- **Intent:** let callers recognize authorization failure semantics through behavior instead of concrete error types.
ErrorAccessDenied indicates access was denied

### ErrorConflict
- **Lines:** 55–58
- **Intent:** let callers recognize state-conflict semantics through behavior instead of concrete error types.
ErrorConflict indicates a conflict occurred

### ErrorConnectionProblem
- **Lines:** 62–65
- **Intent:** let callers recognize transport-failure semantics through behavior instead of concrete error types.
ErrorConnectionProblem indicates a connection issue

### ErrorLimitExceeded
- **Lines:** 69–72
- **Intent:** let callers recognize throttling or quota semantics through behavior instead of concrete error types.
ErrorLimitExceeded indicates a rate limit or quota was exceeded

### ErrorTimeout
- **Lines:** 76–79
- **Intent:** let callers recognize timeout semantics through behavior instead of concrete error types.
ErrorTimeout indicates an operation timed out

### ErrorRetryable
- **Lines:** 83–86
- **Intent:** let callers recognize retry-safe failures through behavior instead of concrete error types.
ErrorRetryable indicates an error that can be retried
