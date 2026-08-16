<!-- generated-by: code-context-graph docs -->
# errors.go

> Typed error categories and retryability rules for trace errors.

## Functions

### IsNotFound
- **Lines:** 94–94
- **Intent:** advertise missing-resource semantics for behavior-based error checks.

### Error
- **Lines:** 97–97
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 100–100
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### LogValue
- **Lines:** 103–103
- **Intent:** reuse TraceError structured logging output for not-found errors.
- **Calls:** LogValue

### IsAlreadyExists
- **Lines:** 112–112
- **Intent:** advertise duplicate-resource semantics for behavior-based error checks.

### Error
- **Lines:** 115–115
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 118–118
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsBadParameter
- **Lines:** 127–127
- **Intent:** advertise invalid-input semantics for behavior-based error checks.

### Error
- **Lines:** 130–130
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 133–133
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsNotImplemented
- **Lines:** 142–142
- **Intent:** advertise unsupported-operation semantics for behavior-based error checks.

### Error
- **Lines:** 145–145
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 148–148
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsUnauthenticated
- **Lines:** 157–157
- **Intent:** advertise authentication-failure semantics for behavior-based checks.

### Error
- **Lines:** 160–160
- **Intent:** delegate developer-facing rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 163–163
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsAccessDenied
- **Lines:** 172–172
- **Intent:** advertise authorization-failure semantics for behavior-based error checks.

### Error
- **Lines:** 175–175
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 178–178
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsConflict
- **Lines:** 187–187
- **Intent:** advertise state-conflict semantics for behavior-based error checks.

### Error
- **Lines:** 190–190
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 193–193
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsConnectionProblem
- **Lines:** 202–202
- **Intent:** advertise transport-failure semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 205–205
- **Intent:** advertise retry-safe semantics for transient connection failures.

### Error
- **Lines:** 208–208
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 211–211
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsLimitExceeded
- **Lines:** 220–220
- **Intent:** advertise throttling semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 223–223
- **Intent:** advertise retry-safe semantics for throttled operations.

### Error
- **Lines:** 226–226
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 229–229
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsTimeout
- **Lines:** 238–238
- **Intent:** advertise timeout semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 241–241
- **Intent:** advertise retry-safe semantics for timeout failures.

### Error
- **Lines:** 244–244
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 247–247
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### wrapTypedInternal
- **Lines:** 251–266
- **Intent:** centralize typed-error wrapping so constructors preserve prior trace frames and structured fields.
- **Ensures:**
  - returns a TraceError that prepends the supplied frame and carries forward existing fields.
- **Calls:** findTraceError, copyFields

### NotFound
- **Lines:** 272–280
- **Intent:** classify a missing resource so callers can branch on lookup failure semantics.
- **Domain Rules:**
  - not found errors are classified as HTTP 404 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
NotFound creates a new NotFoundError
- **Calls:** CaptureFrame, formatMessage

### WrapNotFound
- **Lines:** 286–294
- **Intent:** preserve an existing cause while reclassifying it as a missing-resource failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapNotFound wraps an error as NotFoundError
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### AlreadyExists
- **Lines:** 300–308
- **Intent:** classify duplicate-resource failures so callers can branch on uniqueness semantics.
- **Domain Rules:**
  - already exists errors are classified as HTTP 409 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
AlreadyExists creates a new AlreadyExistsError
- **Calls:** CaptureFrame, formatMessage

### WrapAlreadyExists
- **Lines:** 314–322
- **Intent:** preserve an existing cause while reclassifying it as a duplicate-resource failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapAlreadyExists wraps an error as AlreadyExistsError
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### BadParameter
- **Lines:** 328–336
- **Intent:** classify invalid input so callers can branch on client-side request errors.
- **Domain Rules:**
  - bad parameter errors are classified as HTTP 400 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
BadParameter creates a new BadParameterError
- **Calls:** CaptureFrame, formatMessage

### WrapBadParameter
- **Lines:** 342–350
- **Intent:** preserve an existing cause while surfacing it as a client input failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapBadParameter wraps an error as BadParameterError
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### NotImplemented
- **Lines:** 356–364
- **Intent:** classify unsupported behavior so callers can surface capability gaps consistently.
- **Domain Rules:**
  - not implemented errors are classified as HTTP 501 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
NotImplemented creates a new NotImplementedError
- **Calls:** CaptureFrame, formatMessage

### Unauthenticated
- **Lines:** 370–378
- **Intent:** classify authentication failures separately from authorization failures.
- **Domain Rules:**
  - unauthenticated errors map to HTTP 401 and expose only a fixed client message.
- **Ensures:**
  - records the current call site and keeps the supplied message for diagnostics only.
Unauthenticated creates a new UnauthenticatedError.
- **Calls:** CaptureFrame, formatMessage

### WrapUnauthenticated
- **Lines:** 383–391
- **Intent:** preserve a lower-level cause while classifying an authentication failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
WrapUnauthenticated wraps an error as UnauthenticatedError.
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### AccessDenied
- **Lines:** 397–405
- **Intent:** classify authorization failures so callers can deny access consistently.
- **Domain Rules:**
  - access denied errors are classified as HTTP 403 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
AccessDenied creates a new AccessDeniedError
- **Calls:** CaptureFrame, formatMessage

### WrapAccessDenied
- **Lines:** 411–419
- **Intent:** preserve a lower-level cause while surfacing it as an authorization failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapAccessDenied wraps an error as AccessDeniedError
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### Conflict
- **Lines:** 425–433
- **Intent:** classify state mismatches that prevent the requested operation from succeeding.
- **Domain Rules:**
  - conflict errors are classified as HTTP 409 by the tracehttp package.
- **Ensures:**
  - records the current call site as the first trace frame.
Conflict creates a new ConflictError
- **Calls:** CaptureFrame, formatMessage

### ConnectionProblem
- **Lines:** 441–449
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
- **Lines:** 455–463
- **Intent:** classify quota or rate-limit failures so callers can branch on throttling semantics.
- **Domain Rules:**
  - limit exceeded errors are retryable and map to HTTP 429.
- **Ensures:**
  - records the current call site as the first trace frame.
LimitExceeded creates a new LimitExceededError
- **Calls:** CaptureFrame, formatMessage

### WrapLimitExceeded
- **Lines:** 469–477
- **Intent:** reclassify an existing failure as a quota or rate-limit failure while keeping its cause.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapLimitExceeded wraps an existing error as a LimitExceededError.
- **Calls:** wrapTypedInternal, CaptureFrame, formatMessage

### Timeout
- **Lines:** 485–493
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
- **Lines:** 498–504
- **Intent:** detect missing-resource failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsNotFound checks if error is a not found error
- **Calls:** IsNotFound, chainAs

### IsAlreadyExists
- **Lines:** 509–515
- **Intent:** detect duplicate-resource failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsAlreadyExists checks if error is an already exists error
- **Calls:** IsAlreadyExists, chainAs

### IsBadParameter
- **Lines:** 520–526
- **Intent:** detect caller-input failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsBadParameter checks if error is a bad parameter error
- **Calls:** IsBadParameter, chainAs

### IsNotImplemented
- **Lines:** 531–537
- **Intent:** detect unsupported-operation failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsNotImplemented checks if error is a not implemented error
- **Calls:** IsNotImplemented, chainAs

### IsUnauthenticated
- **Lines:** 542–548
- **Intent:** detect authentication failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsUnauthenticated checks if error is an unauthenticated error.
- **Calls:** IsUnauthenticated, chainAs

### IsAccessDenied
- **Lines:** 553–559
- **Intent:** detect authorization failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsAccessDenied checks if error is an access denied error
- **Calls:** IsAccessDenied, chainAs

### IsConflict
- **Lines:** 564–570
- **Intent:** detect state-conflict failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsConflict checks if error is a conflict error
- **Calls:** IsConflict, chainAs

### IsConnectionProblem
- **Lines:** 575–581
- **Intent:** detect transient transport or infrastructure failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsConnectionProblem checks if error is a connection problem
- **Calls:** IsConnectionProblem, chainAs

### IsLimitExceeded
- **Lines:** 586–592
- **Intent:** detect throttling or quota failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsLimitExceeded checks if error is a limit exceeded error
- **Calls:** IsLimitExceeded, chainAs

### IsTimeout
- **Lines:** 597–603
- **Intent:** detect timeout failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsTimeout checks if error is a timeout error
- **Calls:** IsTimeout, chainAs

### IsRetryable
- **Lines:** 608–614
- **Intent:** detect failures that explicitly advertise retry-safe semantics.
- **Ensures:**
  - returns false for nil errors.
IsRetryable checks if error is retryable
- **Calls:** IsRetryable, chainAs

### Aggregate
- **Lines:** 620–640
- **Intent:** preserve multiple concurrent failures as one error value for later inspection.
- **Domain Rules:**
  - nil errors are discarded so only real failures participate in aggregation.
- **Ensures:**
  - returns nil for an all-nil input set and the sole error unchanged for a single failure.
Aggregate combines multiple errors into a single error using errors.Join (Go 1.20+)

### Error
- **Lines:** 650–657
- **Intent:** render a readable summary that includes every child error in the aggregate.
- **Ensures:**
  - includes the number of collected errors in the output.
- **Calls:** Error

### Unwrap
- **Lines:** 661–663
- **Intent:** expose all child errors so callers can traverse the aggregate tree.
Unwrap returns the list of errors (Go 1.20+ multiple error unwrapping)

## Classes

### NotFoundError
- **Lines:** 89–91
- **Intent:** mark failures caused by missing resources so callers can branch on lookup semantics.
NotFoundError represents a "not found" error

### AlreadyExistsError
- **Lines:** 107–109
- **Intent:** mark failures caused by uniqueness or duplicate-resource conflicts.
AlreadyExistsError represents an "already exists" error

### BadParameterError
- **Lines:** 122–124
- **Intent:** mark failures caused by invalid caller input or malformed parameters.
BadParameterError represents an invalid parameter error

### NotImplementedError
- **Lines:** 137–139
- **Intent:** mark code paths that are recognized but intentionally unsupported.
NotImplementedError represents a "not implemented" error

### UnauthenticatedError
- **Lines:** 152–154
- **Intent:** mark authentication failures independently from authorization failures.
UnauthenticatedError represents a missing or invalid authentication identity.

### AccessDeniedError
- **Lines:** 167–169
- **Intent:** mark authorization failures so callers can deny access consistently.
AccessDeniedError represents an access denied error

### ConflictError
- **Lines:** 182–184
- **Intent:** mark state conflicts that block the requested operation until data changes.
ConflictError represents a conflict error

### ConnectionProblemError
- **Lines:** 197–199
- **Intent:** mark infrastructure or transport failures that may succeed on retry.
ConnectionProblemError represents a connection error

### LimitExceededError
- **Lines:** 215–217
- **Intent:** mark throttling or quota failures so callers can apply backoff behavior.
LimitExceededError represents a rate limit or quota exceeded error

### TimeoutError
- **Lines:** 233–235
- **Intent:** mark operations that exceeded their allowed completion window.
TimeoutError represents a timeout error

### AggregateError
- **Lines:** 644–646
- **Intent:** preserve multiple related failures as one traversable error tree.
AggregateError holds multiple errors

## Types

### ErrorNotFound
- **Lines:** 12–15
- **Intent:** let callers recognize missing-resource semantics through behavior instead of concrete error types.
ErrorNotFound indicates a resource was not found

### ErrorAlreadyExists
- **Lines:** 19–22
- **Intent:** let callers recognize duplicate-resource semantics through behavior instead of concrete error types.
ErrorAlreadyExists indicates a resource already exists

### ErrorBadParameter
- **Lines:** 26–29
- **Intent:** let callers recognize invalid-input semantics through behavior instead of concrete error types.
ErrorBadParameter indicates invalid input parameters

### ErrorNotImplemented
- **Lines:** 33–36
- **Intent:** let callers recognize unsupported-operation semantics through behavior instead of concrete error types.
ErrorNotImplemented indicates functionality is not implemented

### ErrorUnauthenticated
- **Lines:** 40–43
- **Intent:** let callers distinguish authentication failures from authorization failures.
ErrorUnauthenticated indicates authentication is missing or invalid.

### ErrorAccessDenied
- **Lines:** 47–50
- **Intent:** let callers recognize authorization failure semantics through behavior instead of concrete error types.
ErrorAccessDenied indicates access was denied

### ErrorConflict
- **Lines:** 54–57
- **Intent:** let callers recognize state-conflict semantics through behavior instead of concrete error types.
ErrorConflict indicates a conflict occurred

### ErrorConnectionProblem
- **Lines:** 61–64
- **Intent:** let callers recognize transport-failure semantics through behavior instead of concrete error types.
ErrorConnectionProblem indicates a connection issue

### ErrorLimitExceeded
- **Lines:** 68–71
- **Intent:** let callers recognize throttling or quota semantics through behavior instead of concrete error types.
ErrorLimitExceeded indicates a rate limit or quota was exceeded

### ErrorTimeout
- **Lines:** 75–78
- **Intent:** let callers recognize timeout semantics through behavior instead of concrete error types.
ErrorTimeout indicates an operation timed out

### ErrorRetryable
- **Lines:** 82–85
- **Intent:** let callers recognize retry-safe failures through behavior instead of concrete error types.
ErrorRetryable indicates an error that can be retried
