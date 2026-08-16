<!-- generated-by: code-context-graph docs -->
# syserrors.go

> Translation of operating system, filesystem, and network errors into typed trace categories.

## Functions

### ConvertSystemError
- **Lines:** 34–71
- **Intent:** give callers one entry point that turns stdlib failures into trace's typed categories.
- **Domain Rules:**
  - the first matching category wins, checked from most specific to least.
  - an error that already carries a trace category is returned unchanged so
existing classification is never downgraded.
  - the converted error carries no trace message, so the operating system string
stays in the cause for debugging instead of becoming a client-facing HTTP message.
- **Ensures:**
  - returns nil for nil input and the original error when no category applies.
ConvertSystemError converts an operating system, filesystem, or network error
into the matching typed trace error.
- **Calls:** Canceled, WrapNotFound, WrapAlreadyExists, WrapAccessDenied, ConnectionProblem, Timeout, Timeout, Timeout, As, As, alreadyTyped

### alreadyTyped
- **Lines:** 74–86
- **Intent:** avoid re-classifying an error that already carries a trace category.
- **Calls:** IsCanceled, IsNotFound, IsAlreadyExists, IsBadParameter, IsNotImplemented, IsUnauthenticated, IsAccessDenied, IsConflict, IsConnectionProblem, IsLimitExceeded, IsTimeout

## Types

### timeoutReporter
- **Lines:** 12–14
- **Intent:** recognize timeouts reported by net.Error and similar types without importing net.

### temporaryReporter
- **Lines:** 17–19
- **Intent:** recognize temporary transport failures reported by net.Error without importing net.
