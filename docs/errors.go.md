<!-- generated-by: code-context-graph docs -->
# errors.go

> Typed error categories, retryability rules, and HTTP status mapping for trace errors.

## Functions

### IsNotFound
- **Lines:** 102–102
- **Intent:** advertise missing-resource semantics for behavior-based error checks.

### HTTPStatusCode
- **Lines:** 105–105
- **Intent:** map not-found errors to HTTP 404 responses.

### HTTPError
- **Lines:** 108–114
- **Intent:** expose the typed public message without using outer trace wrapper context.

### Error
- **Lines:** 117–117
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 120–120
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### LogValue
- **Lines:** 123–123
- **Intent:** reuse TraceError structured logging output for not-found errors.
- **Calls:** LogValue

### IsAlreadyExists
- **Lines:** 132–132
- **Intent:** advertise duplicate-resource semantics for behavior-based error checks.

### HTTPStatusCode
- **Lines:** 135–135
- **Intent:** map duplicate-resource errors to HTTP 409 responses.

### HTTPError
- **Lines:** 138–140
- **Intent:** expose the typed public message and a stable duplicate-resource code.

### Error
- **Lines:** 143–143
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 146–146
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsBadParameter
- **Lines:** 155–155
- **Intent:** advertise invalid-input semantics for behavior-based error checks.

### HTTPStatusCode
- **Lines:** 158–158
- **Intent:** map invalid-input errors to HTTP 400 responses.

### HTTPError
- **Lines:** 161–163
- **Intent:** expose the typed public message and a stable invalid-input code.

### Error
- **Lines:** 166–166
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 169–169
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsNotImplemented
- **Lines:** 178–178
- **Intent:** advertise unsupported-operation semantics for behavior-based error checks.

### HTTPStatusCode
- **Lines:** 181–181
- **Intent:** map unsupported-operation errors to HTTP 501 responses.

### HTTPError
- **Lines:** 184–186
- **Intent:** classify unsupported functionality while leaving 5xx message sanitization to ToHTTPError.

### Error
- **Lines:** 189–189
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 192–192
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsUnauthenticated
- **Lines:** 201–201
- **Intent:** advertise authentication-failure semantics for behavior-based checks.

### HTTPStatusCode
- **Lines:** 204–204
- **Intent:** map authentication failures to HTTP 401 responses.

### HTTPError
- **Lines:** 207–213
- **Intent:** expose only a fixed authentication message to clients.

### Error
- **Lines:** 216–216
- **Intent:** delegate developer-facing rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 219–219
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsAccessDenied
- **Lines:** 228–228
- **Intent:** advertise authorization-failure semantics for behavior-based error checks.

### HTTPStatusCode
- **Lines:** 231–231
- **Intent:** map access-denied errors to HTTP 403 responses.

### HTTPError
- **Lines:** 234–236
- **Intent:** expose only a fixed authorization message to clients.

### Error
- **Lines:** 239–239
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 242–242
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsConflict
- **Lines:** 251–251
- **Intent:** advertise state-conflict semantics for behavior-based error checks.

### HTTPStatusCode
- **Lines:** 254–254
- **Intent:** map conflict errors to HTTP 409 responses.

### HTTPError
- **Lines:** 257–259
- **Intent:** expose the typed public message and a stable state-conflict code.

### Error
- **Lines:** 262–262
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 265–265
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsConnectionProblem
- **Lines:** 274–274
- **Intent:** advertise transport-failure semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 277–277
- **Intent:** advertise retry-safe semantics for transient connection failures.

### HTTPStatusCode
- **Lines:** 280–280
- **Intent:** map connection problems to HTTP 503 responses.

### HTTPError
- **Lines:** 283–285
- **Intent:** classify transient unavailability while leaving 5xx message sanitization to ToHTTPError.

### Error
- **Lines:** 288–288
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 291–291
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsLimitExceeded
- **Lines:** 300–300
- **Intent:** advertise throttling semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 303–303
- **Intent:** advertise retry-safe semantics for throttled operations.

### HTTPStatusCode
- **Lines:** 306–306
- **Intent:** map throttling and quota failures to HTTP 429 responses.

### HTTPError
- **Lines:** 309–311
- **Intent:** expose the typed public message and a stable throttling code.

### Error
- **Lines:** 314–314
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 317–317
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### IsTimeout
- **Lines:** 326–326
- **Intent:** advertise timeout semantics for behavior-based error checks.

### IsRetryable
- **Lines:** 329–329
- **Intent:** advertise retry-safe semantics for timeout failures.

### HTTPStatusCode
- **Lines:** 332–332
- **Intent:** map timeout failures to HTTP 504 responses.

### HTTPError
- **Lines:** 335–337
- **Intent:** classify timeouts while leaving 5xx message sanitization to ToHTTPError.

### Error
- **Lines:** 340–340
- **Intent:** delegate user-facing string rendering to the embedded TraceError.
- **Calls:** Error

### Unwrap
- **Lines:** 343–343
- **Intent:** expose the embedded TraceError to standard Go error traversal.

### wrapTypedInternal
- **Lines:** 347–363
- **Intent:** centralize typed-error wrapping so constructors preserve prior trace frames and structured fields.
- **Ensures:**
  - returns a TraceError that prepends the supplied frame and carries forward existing fields.
- **Calls:** As

### NotFound
- **Lines:** 370–379
- **Intent:** classify a missing resource so callers can branch on lookup failure semantics.
- **Domain Rules:**
  - not found errors map to HTTP 404 through HTTPStatusCode.
- **Ensures:**
  - records the current call site as the first trace frame.
  - returns a NotFoundError with initialized structured fields.
NotFound creates a new NotFoundError
- **Calls:** captureFrame, formatMessage

### WrapNotFound
- **Lines:** 385–393
- **Intent:** preserve an existing cause while reclassifying it as a missing-resource failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapNotFound wraps an error as NotFoundError
- **Calls:** wrapTypedInternal, captureFrame, formatMessage

### AlreadyExists
- **Lines:** 399–408
- **Intent:** classify duplicate-resource failures so callers can branch on uniqueness semantics.
- **Domain Rules:**
  - already exists errors map to HTTP 409 through HTTPStatusCode.
- **Ensures:**
  - records the current call site as the first trace frame.
AlreadyExists creates a new AlreadyExistsError
- **Calls:** captureFrame, formatMessage

### WrapAlreadyExists
- **Lines:** 414–422
- **Intent:** preserve an existing cause while reclassifying it as a duplicate-resource failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapAlreadyExists wraps an error as AlreadyExistsError
- **Calls:** wrapTypedInternal, captureFrame, formatMessage

### BadParameter
- **Lines:** 428–437
- **Intent:** classify invalid input so callers can branch on client-side request errors.
- **Domain Rules:**
  - bad parameter errors map to HTTP 400 through HTTPStatusCode.
- **Ensures:**
  - records the current call site as the first trace frame.
BadParameter creates a new BadParameterError
- **Calls:** captureFrame, formatMessage

### WrapBadParameter
- **Lines:** 443–451
- **Intent:** preserve an existing cause while surfacing it as a client input failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapBadParameter wraps an error as BadParameterError
- **Calls:** wrapTypedInternal, captureFrame, formatMessage

### NotImplemented
- **Lines:** 457–466
- **Intent:** classify unsupported behavior so callers can surface capability gaps consistently.
- **Domain Rules:**
  - not implemented errors map to HTTP 501 through HTTPStatusCode.
- **Ensures:**
  - records the current call site as the first trace frame.
NotImplemented creates a new NotImplementedError
- **Calls:** captureFrame, formatMessage

### Unauthenticated
- **Lines:** 472–481
- **Intent:** classify authentication failures separately from authorization failures.
- **Domain Rules:**
  - unauthenticated errors map to HTTP 401 and expose only a fixed client message.
- **Ensures:**
  - records the current call site and keeps the supplied message for diagnostics only.
Unauthenticated creates a new UnauthenticatedError.
- **Calls:** captureFrame, formatMessage

### WrapUnauthenticated
- **Lines:** 486–494
- **Intent:** preserve a lower-level cause while classifying an authentication failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
WrapUnauthenticated wraps an error as UnauthenticatedError.
- **Calls:** wrapTypedInternal, captureFrame, formatMessage

### AccessDenied
- **Lines:** 501–510
- **Intent:** classify authorization failures so callers can deny access consistently.
- **Domain Rules:**
  - access denied errors map to HTTP 403 through HTTPStatusCode.
- **Ensures:**
  - records the current call site as the first trace frame.
  - returns an AccessDeniedError with initialized structured fields.
AccessDenied creates a new AccessDeniedError
- **Calls:** captureFrame, formatMessage

### WrapAccessDenied
- **Lines:** 516–524
- **Intent:** preserve a lower-level cause while surfacing it as an authorization failure.
- **Domain Rules:**
  - returns nil unchanged when the source error is nil.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
WrapAccessDenied wraps an error as AccessDeniedError
- **Calls:** wrapTypedInternal, captureFrame, formatMessage

### Conflict
- **Lines:** 530–539
- **Intent:** classify state mismatches that prevent the requested operation from succeeding.
- **Domain Rules:**
  - conflict errors map to HTTP 409 through HTTPStatusCode.
- **Ensures:**
  - records the current call site as the first trace frame.
Conflict creates a new ConflictError
- **Calls:** captureFrame, formatMessage

### ConnectionProblem
- **Lines:** 547–555
- **Intent:** mark infrastructure or network failures as transient connection problems.
- **Domain Rules:**
  - connection problems are retryable and map to HTTP 503.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
  - returns nil unchanged when the source error is nil.
ConnectionProblem creates a new ConnectionProblemError.
If err is nil, returns nil.
- **Calls:** wrapTypedInternal, captureFrame, formatMessage

### LimitExceeded
- **Lines:** 561–570
- **Intent:** classify quota or rate-limit failures so callers can branch on throttling semantics.
- **Domain Rules:**
  - limit exceeded errors are retryable and map to HTTP 429.
- **Ensures:**
  - records the current call site as the first trace frame.
LimitExceeded creates a new LimitExceededError
- **Calls:** captureFrame, formatMessage

### Timeout
- **Lines:** 578–586
- **Intent:** classify operations that exceeded their allowed completion window.
- **Domain Rules:**
  - timeout errors are retryable and map to HTTP 504.
- **Ensures:**
  - prepends the current call site to any existing trace frames on the returned error.
  - returns nil unchanged when the source error is nil.
Timeout creates a new TimeoutError.
If err is nil, returns nil.
- **Calls:** wrapTypedInternal, captureFrame, formatMessage

### IsNotFound
- **Lines:** 591–597
- **Intent:** detect missing-resource failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsNotFound checks if error is a not found error
- **Calls:** IsNotFound, As

### IsAlreadyExists
- **Lines:** 602–608
- **Intent:** detect duplicate-resource failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsAlreadyExists checks if error is an already exists error
- **Calls:** IsAlreadyExists, As

### IsBadParameter
- **Lines:** 613–619
- **Intent:** detect caller-input failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsBadParameter checks if error is a bad parameter error
- **Calls:** IsBadParameter, As

### IsNotImplemented
- **Lines:** 624–630
- **Intent:** detect unsupported-operation failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsNotImplemented checks if error is a not implemented error
- **Calls:** IsNotImplemented, As

### IsUnauthenticated
- **Lines:** 635–641
- **Intent:** detect authentication failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsUnauthenticated checks if error is an unauthenticated error.
- **Calls:** IsUnauthenticated, As

### IsAccessDenied
- **Lines:** 646–652
- **Intent:** detect authorization failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsAccessDenied checks if error is an access denied error
- **Calls:** IsAccessDenied, As

### IsConflict
- **Lines:** 657–663
- **Intent:** detect state-conflict failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsConflict checks if error is a conflict error
- **Calls:** IsConflict, As

### IsConnectionProblem
- **Lines:** 668–674
- **Intent:** detect transient transport or infrastructure failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsConnectionProblem checks if error is a connection problem
- **Calls:** IsConnectionProblem, As

### IsLimitExceeded
- **Lines:** 679–685
- **Intent:** detect throttling or quota failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsLimitExceeded checks if error is a limit exceeded error
- **Calls:** IsLimitExceeded, As

### IsTimeout
- **Lines:** 690–696
- **Intent:** detect timeout failures anywhere in an error chain.
- **Ensures:**
  - returns false for nil errors.
IsTimeout checks if error is a timeout error
- **Calls:** IsTimeout, As

### IsRetryable
- **Lines:** 701–707
- **Intent:** detect failures that explicitly advertise retry-safe semantics.
- **Ensures:**
  - returns false for nil errors.
IsRetryable checks if error is retryable
- **Calls:** IsRetryable, As

### GetHTTPStatusCode
- **Lines:** 713–722
- **Intent:** translate trace error categories into HTTP response codes at service boundaries.
- **Domain Rules:**
  - typed errors decide the status code; unknown errors default to HTTP 500.
- **Ensures:**
  - returns HTTP 200 for nil errors and HTTP 500 for unknown non-typed errors.
GetHTTPStatusCode returns the HTTP status code for an error
- **Calls:** As

### Aggregate
- **Lines:** 728–748
- **Intent:** preserve multiple concurrent failures as one error value for later inspection.
- **Domain Rules:**
  - nil errors are discarded so only real failures participate in aggregation.
- **Ensures:**
  - returns nil for an all-nil input set and the sole error unchanged for a single failure.
Aggregate combines multiple errors into a single error using errors.Join (Go 1.20+)

### Error
- **Lines:** 758–765
- **Intent:** render a readable summary that includes every child error in the aggregate.
- **Ensures:**
  - includes the number of collected errors in the output.
- **Calls:** Error

### Unwrap
- **Lines:** 769–771
- **Intent:** expose all child errors so callers can traverse the aggregate tree.
Unwrap returns the list of errors (Go 1.20+ multiple error unwrapping)

### HTTPError
- **Lines:** 775–787
- **Intent:** choose the safe semantic HTTP representation with the highest child status.
- **Domain Rules:**
  - classification priority matches AggregateError.HTTPStatusCode rather than traversal order.
- **Calls:** ToHTTPError, internalHTTPError

### HTTPStatusCode
- **Lines:** 792–803
- **Intent:** choose one HTTP status code that best represents the aggregate failure set.
- **Domain Rules:**
  - the highest child status code wins, with 500 as the fallback for all-OK children.
HTTPStatusCode returns the most severe HTTP status code
- **Calls:** GetHTTPStatusCode

## Classes

### NotFoundError
- **Lines:** 97–99
- **Intent:** mark failures caused by missing resources so callers can branch on lookup semantics.
NotFoundError represents a "not found" error

### AlreadyExistsError
- **Lines:** 127–129
- **Intent:** mark failures caused by uniqueness or duplicate-resource conflicts.
AlreadyExistsError represents an "already exists" error

### BadParameterError
- **Lines:** 150–152
- **Intent:** mark failures caused by invalid caller input or malformed parameters.
BadParameterError represents an invalid parameter error

### NotImplementedError
- **Lines:** 173–175
- **Intent:** mark code paths that are recognized but intentionally unsupported.
NotImplementedError represents a "not implemented" error

### UnauthenticatedError
- **Lines:** 196–198
- **Intent:** mark authentication failures independently from authorization failures.
UnauthenticatedError represents a missing or invalid authentication identity.

### AccessDeniedError
- **Lines:** 223–225
- **Intent:** mark authorization failures so callers can deny access consistently.
AccessDeniedError represents an access denied error

### ConflictError
- **Lines:** 246–248
- **Intent:** mark state conflicts that block the requested operation until data changes.
ConflictError represents a conflict error

### ConnectionProblemError
- **Lines:** 269–271
- **Intent:** mark infrastructure or transport failures that may succeed on retry.
ConnectionProblemError represents a connection error

### LimitExceededError
- **Lines:** 295–297
- **Intent:** mark throttling or quota failures so callers can apply backoff behavior.
LimitExceededError represents a rate limit or quota exceeded error

### TimeoutError
- **Lines:** 321–323
- **Intent:** mark operations that exceeded their allowed completion window.
TimeoutError represents a timeout error

### AggregateError
- **Lines:** 752–754
- **Intent:** preserve multiple related failures as one traversable error tree.
AggregateError holds multiple errors

## Types

### ErrorNotFound
- **Lines:** 14–17
- **Intent:** let callers recognize missing-resource semantics through behavior instead of concrete error types.
ErrorNotFound indicates a resource was not found

### ErrorAlreadyExists
- **Lines:** 21–24
- **Intent:** let callers recognize duplicate-resource semantics through behavior instead of concrete error types.
ErrorAlreadyExists indicates a resource already exists

### ErrorBadParameter
- **Lines:** 28–31
- **Intent:** let callers recognize invalid-input semantics through behavior instead of concrete error types.
ErrorBadParameter indicates invalid input parameters

### ErrorNotImplemented
- **Lines:** 35–38
- **Intent:** let callers recognize unsupported-operation semantics through behavior instead of concrete error types.
ErrorNotImplemented indicates functionality is not implemented

### ErrorUnauthenticated
- **Lines:** 42–45
- **Intent:** let callers distinguish authentication failures from authorization failures.
ErrorUnauthenticated indicates authentication is missing or invalid.

### ErrorAccessDenied
- **Lines:** 49–52
- **Intent:** let callers recognize authorization failure semantics through behavior instead of concrete error types.
ErrorAccessDenied indicates access was denied

### ErrorConflict
- **Lines:** 56–59
- **Intent:** let callers recognize state-conflict semantics through behavior instead of concrete error types.
ErrorConflict indicates a conflict occurred

### ErrorConnectionProblem
- **Lines:** 63–66
- **Intent:** let callers recognize transport-failure semantics through behavior instead of concrete error types.
ErrorConnectionProblem indicates a connection issue

### ErrorLimitExceeded
- **Lines:** 70–73
- **Intent:** let callers recognize throttling or quota semantics through behavior instead of concrete error types.
ErrorLimitExceeded indicates a rate limit or quota was exceeded

### ErrorTimeout
- **Lines:** 77–80
- **Intent:** let callers recognize timeout semantics through behavior instead of concrete error types.
ErrorTimeout indicates an operation timed out

### ErrorRetryable
- **Lines:** 84–87
- **Intent:** let callers recognize retry-safe failures through behavior instead of concrete error types.
ErrorRetryable indicates an error that can be retried

### HTTPStatusCode
- **Lines:** 91–93
- **Intent:** let error types declare the HTTP status code they should map to at service boundaries.
HTTPStatusCode is an interface for errors that can provide HTTP status codes
