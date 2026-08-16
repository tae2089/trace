# Trace - Modern Go Error Handling

[![Go Reference](https://pkg.go.dev/badge/github.com/tae2089/trace/v2.svg)](https://pkg.go.dev/github.com/tae2089/trace/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/tae2089/trace/v2)](https://goreportcard.com/report/github.com/tae2089/trace/v2)

A modern error handling package for Go, inspired by [gravitational/trace](https://github.com/gravitational/trace) but upgraded for Go 1.25+ with:

- 🔍 **Stack traces** - Know exactly where errors originate
- 🏷️ **Typed errors** - NotFound, BadParameter, AccessDenied, etc.
- 📊 **slog integration** - Structured logging out of the box
- 🔗 **Full errors.Is/As support** - Compatible with Go 1.13+ error handling
- 🧬 **Generics** - Result types, pipelines, and type-safe operations
- 🌐 **HTTP utilities** - Middleware, error responses, status code mapping
- 📦 **Context integration** - Trace IDs, fields, cancel cause, detached contexts
- 🔄 **Error chain iterator** - `for e := range trace.Errors(err)` (Go 1.23+)
- 🧰 **System error conversion** - `os`, `io/fs`, and `syscall` failures become typed errors

## Installation

```bash
go get github.com/tae2089/trace/v2
```

## Packages

| Package | Import path | Contents |
| --- | --- | --- |
| `trace` | `github.com/tae2089/trace/v2` | Error values, stack frames, typed categories, context helpers, slog integration, `Result`/`Pipeline` |
| `tracehttp` | `github.com/tae2089/trace/v2/tracehttp` | Everything that touches `net/http`: status mapping, JSON error responses, middleware, client |

The split exists so that programs which only need error values never pay for
`net/http`. Measured on Go 1.25:

| Dependency graph | stdlib packages |
| --- | ---: |
| `trace` | 72 |
| `trace` + `tracehttp` | 185 |

Importing `net/http` also drags in the whole `crypto/tls` and `crypto/x509`
tree. CI fails the build if `net/http` reappears in the root package's graph.

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/tae2089/trace/v2"
)

func main() {
    err := fetchUser("user-123")
    if err != nil {
        if trace.IsNotFound(err) {
            fmt.Println("User not found")
        }
        // Full debug output with stack trace
        fmt.Printf("%+v\n", err)
    }
}

func fetchUser(id string) error {
    _, err := queryDatabase(id)
    if err != nil {
        return trace.Wrapf(err, "failed to fetch user %s", id)
    }
    return nil
}

type User struct{}

func queryDatabase(id string) (*User, error) {
    // Simulate not found
    return nil, trace.NotFound("user %s does not exist", id)
}
```

Output:

```
User not found
[main.go:25 <- main.go:18 <- main.go:12] failed to fetch user user-123
→ user user-123 does not exist
Stack trace:
  main.go:25 main.queryDatabase
  main.go:18 main.fetchUser
  main.go:12 main.main
```

## Core Features

### Basic Wrapping

```go
// Simple wrap - adds stack frame
err := trace.Wrap(originalErr)

// Wrap with message
err := trace.Wrap(originalErr, "operation failed")

// Wrap with formatted message
err := trace.Wrapf(originalErr, "failed to process user %s", userID)

// Create new error with stack trace
err := trace.New("something went wrong")
err := trace.Errorf("failed to process %d items", count)
```

### Typed Errors

```go
// Create typed errors
err := trace.NotFound("user %s not found", userID)
err := trace.AlreadyExists("email already registered")
err := trace.BadParameter("invalid email format")
err := trace.Unauthenticated("invalid bearer token")
err := trace.AccessDenied("insufficient permissions")
err := trace.Conflict("version mismatch")
err := trace.LimitExceeded("rate limit exceeded")
err := trace.Timeout(originalErr, "request timed out")
err := trace.ConnectionProblem(originalErr, "database unreachable")
err := trace.NotImplemented("feature coming soon")

// Wrap existing error as typed
err := trace.WrapNotFound(sql.ErrNoRows, "user not found")
err := trace.WrapAccessDenied(err, "permission check failed")

// Check error types (works through wrapped errors)
if trace.IsNotFound(err) { /* handle 404 */ }
if trace.IsUnauthenticated(err) { /* handle 401 */ }
if trace.IsAccessDenied(err) { /* handle 403 */ }
if trace.IsRetryable(err) { /* retry the operation */ }

// Get HTTP status code
statusCode := tracehttp.GetHTTPStatusCode(err) // e.g., 404, 403, 500
```

### Structured Fields

```go
// Add fields for structured logging
err := trace.NotFound("user not found")
err = trace.WithField(err, "user_id", userID)
err = trace.WithFields(err, map[string]any{
    "request_id": reqID,
    "tenant":     tenant,
})

// Or create with fields directly
err := trace.WrapWithFields(originalErr, map[string]any{
    "user_id": userID,
    "action":  "delete",
}, "operation failed")

// Extract fields
fields := trace.GetFields(err)
```

### slog Integration

```go
import "log/slog"

logger := slog.Default()

// Log with full trace information
err := trace.Wrap(dbErr, "query failed")
logger.Error("operation failed", trace.SlogError(err))

// Output (JSON):
// {
//   "level": "ERROR",
//   "msg": "operation failed",
//   "error": {
//     "message": "query failed",
//     "cause": "connection refused",
//     "trace": [
//       {"file": "repo.go", "line": 42, "func": "repo.Query"},
//       {"file": "service.go", "line": 28, "func": "service.GetUser"}
//     ]
//   }
// }

// Use the error handler for automatic extraction
handler := trace.NewErrorHandler(slog.NewJSONHandler(os.Stdout, nil))
logger := slog.New(handler)

// Passing nil uses slog.DiscardHandler — safe for tests or optional logging
discardHandler := trace.NewErrorHandler(nil)
```

### HTTP Utilities

```go
// The application owns request logging; trace only classifies and renders.
func Handle(fn tracehttp.ErrorHandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := fn(w, r); err != nil {
            httpErr := tracehttp.ToHTTPError(err)
            logger.Error("request failed",
                trace.SlogError(err),
                slog.Int("status_code", httpErr.Status),
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
            )

            requestID := r.Header.Get("X-Request-ID")
            if writeErr := tracehttp.WriteError(w, err, requestID); writeErr != nil {
                logger.Error("failed to write error response", "error", writeErr)
            }
        }
    }
}

http.Handle("/users/{id}", Handle(getUserHandler))
```

The response contains only stable client-safe fields:

```json
{
  "error": {
    "code": "not_found",
    "message": "user not found",
    "request_id": "01KREQUEST"
  }
}
```

Framework adapters such as Gin can use `ErrorResponseFor` without writing through
`net/http`:

```go
status, response := tracehttp.ErrorResponseFor(err, requestID)
```

`request_id` is supplied explicitly and is separate from internal `trace_id`
fields. Error fields, details, causes, stack frames, and outer `trace.Wrap`
messages are never copied into the response. All 5xx messages are normalized to
`internal server error`.

`WriteErrorWithLogger`, `ErrorMiddlewareWithLogger`, and `RecoverMiddleware`
were removed in v2. Applications should decide logging level, duration, route,
response-size, and panic-recovery policy themselves.

### Which error decides the response

`ToHTTPError` walks the chain from the outside in and stops at the first link
it can classify, so an outer wrap always beats an inner one:

```go
err := trace.NotFound("secret project not found")
err = trace.WrapAccessDenied(err, "not a member")

tracehttp.ToHTTPError(err) // 403 access_denied — not 404
```

That ordering matters for more than tidiness. If the inner `NotFound` won, the
response would tell an unauthorized caller that the resource exists.

An application type that implements `HTTPErrorProvider` joins the same walk and
wins at whatever depth it sits:

```go
type QuotaError struct{ Plan string }

func (e *QuotaError) Error() string { return "plan quota reached" }

func (e *QuotaError) HTTPError() tracehttp.HTTPError {
    return tracehttp.HTTPError{
        Status:  http.StatusPaymentRequired,
        Code:    "quota_reached",
        Message: "upgrade required",
    }
}
```

Statuses outside 400–599, or a blank `Code`, are rejected and become a generic
internal error, and every 5xx message is replaced with `internal server error`.

Clients using this response contract can safely restore built-in typed errors:

```go
resp, err := http.Get("https://api.example.com/users/123")
if err != nil {
    return err
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
    return err
}
if err := tracehttp.ReadErrorResponse(resp.StatusCode, body); err != nil {
    if trace.IsNotFound(err) {
        // error.code was "not_found"
    }
    return err
}
```

`ReadErrorResponse` reads only `error.code`, `error.message`, and `request_id`
from the public envelope. It creates a new local trace and never deserializes
remote causes, fields, details, frames, or an internal `TraceError`. The HTTP
status must agree with the code; malformed envelopes, unknown/custom codes, and
status/code mismatches become a generic internal error without retaining the raw
response body. Authentication, authorization, cancellation, and all 5xx
messages are normalized again while decoding.

| `error.code` | HTTP status | Restored predicate |
| --- | ---: | --- |
| `bad_request` | 400 | `IsBadParameter` |
| `unauthenticated` | 401 | `IsUnauthenticated` |
| `access_denied` | 403 | `IsAccessDenied` |
| `not_found` | 404 | `IsNotFound` |
| `already_exists` | 409 | `IsAlreadyExists` |
| `conflict` | 409 | `IsConflict` |
| `limit_exceeded` | 429 | `IsLimitExceeded` |
| `canceled` | 499 | `IsCanceled` |
| `not_implemented` | 501 | `IsNotImplemented` |
| `unavailable` | 503 | `IsConnectionProblem` |
| `timeout` | 504 | `IsTimeout` |
| `internal` | 5xx | Generic internal error |

`FromHTTPResponse` remains available for legacy status/plain-text responses. It
classifies by status and retains the response body in the developer-facing
error, so prefer `ReadErrorResponse` for the safe JSON contract above.

### Context Integration

```go
// Add trace ID and fields to context
ctx := trace.ContextWithTraceID(ctx, "req-abc-123")
ctx = trace.ContextWithField(ctx, "user_id", userID)

// Wrap errors with context information
err := trace.WrapContext(ctx, dbErr, "query failed")
// err now contains trace_id and user_id in fields

// Check context errors — FromContext uses context.Cause to preserve the
// original cancellation reason (e.g., the error passed to CancelCauseFunc)
if err := trace.FromContext(ctx); err != nil {
    if trace.IsCanceled(err) {
        // Handle cancellation
    }
    if trace.IsTimeout(err) {
        // Handle timeout
    }
}

// Cancel with cause — the cause is preserved through FromContext
ctx, cancel := trace.WithCancelCause(parentCtx)
cancel(errors.New("shutdown requested"))
err := trace.FromContext(ctx) // inner error is "shutdown requested", not generic "context canceled"

// Timeout with cause
ctx, cancel := trace.WithTimeoutCause(parentCtx, 5*time.Second, errors.New("slow query"))
defer cancel()
// if timeout fires, FromContext(ctx) wraps "slow query" as a TimeoutError

// Detached context — carries values but not cancellation
// Useful for background goroutines that should outlive the request
detached := trace.DetachedContext(ctx)
go cleanup(detached) // won't be canceled when parent ctx is canceled

// Cancel callback — run a function when context is canceled
c := trace.NewContextualizer(ctx)
stop := c.OnCancel(func() {
    releaseResource()
})
// call stop() to prevent the callback if no longer needed
```

### Generic Result Type

```go
// Create results
result := trace.Ok(user)
result := trace.Err[*User](errors.New("not found"))

// Use Try for (value, error) functions
result := trace.Try(db.QueryUser(id))

// Check and unwrap
if result.IsOk() {
    user := result.Unwrap()
}

// Safe unwrap with default
user := result.UnwrapOr(defaultUser)

// Transform
nameResult := trace.Map(userResult, func(u *User) string {
    return u.Name
})

// Chain operations
result := trace.FlatMap(userResult, func(u *User) trace.Result[*Profile] {
    return trace.Try(db.GetProfile(u.ID))
})

// Collect multiple results
results := trace.Collect(result1, result2, result3)
if results.IsErr() {
    // Handle aggregated errors
}
```

### Pipeline Pattern

```go
// Chain operations with automatic error propagation
result, err := trace.NewPipeline(userInput).
    Then(validate).
    Then(normalize).
    Then(save).
    Result()

// With recovery
result, err := trace.NewPipeline(data).
    Then(process).
    Recover(func(err error) (Data, error) {
        if trace.IsNotFound(err) {
            return defaultData, nil
        }
        return Data{}, err
    }).
    Result()

// Transform between types
pipeline := trace.NewPipeline(userID)
result := trace.TransformPipeline(pipeline, func(id string) (*User, error) {
    return db.FindUser(id)
})
```

### Aggregate Errors (Go 1.20+)

```go
// Combine multiple errors
errs := []error{err1, err2, err3}
combined := trace.Aggregate(errs...)

// Works with errors.Is/As
if trace.IsNotFound(combined) { /* at least one is NotFound */ }

// Get most severe HTTP status
statusCode := tracehttp.GetHTTPStatusCode(combined)
```

### Error Chain Iterator (Go 1.23+)

```go
// Iterate over the entire error chain using range-over-func
err := trace.Wrap(trace.Wrap(dbErr, "repo"), "service")

for e := range trace.Errors(err) {
    fmt.Println(e)
}

// Works with AggregateError — traverses all branches
agg := trace.Aggregate(err1, err2, err3)
for e := range trace.Errors(agg) {
    if trace.IsNotFound(e) {
        // found a NotFound somewhere in the tree
    }
}
```

### System Error Conversion

`ConvertSystemError` turns `os`, `io/fs`, and `syscall` failures into typed
trace errors so the rest of your code can use `trace.IsNotFound` and friends
instead of matching sentinel values by hand.

```go
f, err := os.Open(path)
if err != nil {
    return trace.ConvertSystemError(err) // fs.ErrNotExist becomes a NotFoundError
}
```

| Source error | Result |
| --- | --- |
| `fs.ErrNotExist` | `NotFoundError` |
| `fs.ErrExist` | `AlreadyExistsError` |
| `fs.ErrPermission` | `AccessDeniedError` |
| `context.Canceled` | `CanceledError` |
| `context.DeadlineExceeded`, `os.ErrDeadlineExceeded`, `Timeout() bool` reporting true | `TimeoutError` |
| `ECONNREFUSED`, `ECONNRESET`, `ECONNABORTED`, `EHOSTUNREACH`, `ENETUNREACH`, `ENETDOWN`, `EPIPE` | `ConnectionProblemError` |
| `ETIMEDOUT` | `TimeoutError` |
| `EMFILE`, `ENFILE` | `LimitExceededError` |
| `Temporary() bool` reporting true | `ConnectionProblemError` |

Rules:

- `nil` returns `nil`, and an error that already carries a trace category is
  returned unchanged — an outer `WrapAccessDenied` is never downgraded by an
  inner `fs.ErrNotExist`.
- Anything unrecognized is returned as-is, not forced into a category.
- The conversion attaches **no message**. Operating system text often contains a
  file path, so it stays in the cause, where `%+v` and logs can see it, and out
  of `tracehttp`'s client response.
- Errno matching is compiled only on `unix` and `windows`; other platforms fall
  back to the portable rules above and still build.

### Debug Output

```go
// Simple error string
fmt.Println(err)
// [repo.go:42 <- service.go:28] failed to fetch user
// → connection refused

// Verbose with full stack trace
fmt.Printf("%+v\n", err)
// [repo.go:42 <- service.go:28] failed to fetch user
// → connection refused
// Stack trace:
//   repo.go:42 repo.Query
//   service.go:28 service.GetUser
//   handler.go:15 handler.HandleRequest
// Fields:
//   user_id: abc123
//   request_id: req-456

// Full debug report
fmt.Println(trace.DebugReport(err))

// User-friendly message (without stack traces)
msg := trace.UserMessage(err) // "failed to fetch user"
```

## Performance

Construction is the hot path — errors are created far more often than they are
printed — so v2 moves cost from creation to rendering:

- A `Frame` records only the program counter. The function name, file, and line
  resolve when the error is rendered (`Error()`, `%+v`, slog, JSON).
- The structured fields map is allocated only when a field is attached.
- Hot paths walk the error chain with plain type assertions instead of
  `errors.As`, with the same semantics (including the `As(any) bool` hook and
  aggregate branches).

Measured on Apple M1 Pro, Go 1.25 (`go test -bench . -benchmem`):

| Benchmark | v1 layout | v2 |
| --- | --- | --- |
| `Wrap` (1 level) | 435 ns, 416 B, 6 allocs | 162 ns, 72 B, 2 allocs |
| `NotFound` | 389 ns, 416 B, 6 allocs | 165 ns, 80 B, 3 allocs |
| `Wrap` ×10 deep | 4.9 µs, 6.4 KB, 69 allocs | 1.9 µs, 1.2 KB, 29 allocs |
| `CaptureFrame` | 285 ns, 248 B, 2 allocs | 95 ns, 0 B, 0 allocs |
| `IsNotFound` (10 deep) | 638 ns | 102 ns |
| `err.Error()` (5 deep) | 1.3 µs | 4.4 µs |

The last row is the deliberate trade: rendering pays for the deferred symbol
resolution. `fmt.Errorf("%w")` is still ~2× faster than `Wrap` — that is the
price of carrying a stack trace at all.

## Best Practices

### 1. Wrap at Every Layer

```go
// Repository
func (r *UserRepo) FindByID(id string) (*User, error) {
    user, err := r.db.Query(...)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, trace.WrapNotFound(err, "user %s not found", id)
        }
        return nil, trace.Wrap(err, "database query failed")
    }
    return user, nil
}

// Service
func (s *UserService) GetUser(id string) (*User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, trace.Wrap(err, "service: get user")
    }
    return user, nil
}

// Handler
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) error {
    user, err := h.service.GetUser(r.PathValue("id"))
    if err != nil {
        return trace.Wrap(err) // Final wrap before response
    }
    return json.NewEncoder(w).Encode(user)
}
```

### 2. Use Typed Errors for Business Logic

```go
func (s *OrderService) PlaceOrder(ctx context.Context, order Order) error {
    // Check inventory
    if !s.inventory.HasStock(order.ItemID) {
        return trace.Conflict("item %s out of stock", order.ItemID)
    }

    // Check user permissions
    if !s.auth.CanPurchase(ctx, order.UserID) {
        return trace.AccessDenied("user cannot place orders")
    }

    // Validate
    if order.Quantity <= 0 {
        return trace.BadParameter("quantity must be positive")
    }

    return s.repo.SaveOrder(order)
}
```

### 3. Add Context for Debugging

```go
func ProcessBatch(ctx context.Context, items []Item) error {
    for i, item := range items {
        if err := processItem(ctx, item); err != nil {
            return trace.WrapWithFields(err, map[string]any{
                "batch_index": i,
                "item_id":     item.ID,
            }, "batch processing failed")
        }
    }
    return nil
}
```

### 4. Use Context for Request Tracing

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        ctx = trace.ContextWithTraceID(ctx, generateTraceID())
        ctx = trace.ContextWithFields(ctx, map[string]any{
            "method": r.Method,
            "path":   r.URL.Path,
        })
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Migration from gravitational/trace

This package is inspired by gravitational/trace, but it is not a drop-in
replacement. Important differences:

| gravitational/trace | This package |
| --- | --- |
| `trace.Traces` embed | Use `*TraceError` directly |
| `trace.OrigError()` | Use `errors.Unwrap()` or `errors.Is/As` |
| Manual `SetTrace` | Automatic via `Wrap()` |
| `WriteError` serializes trace internals | `WriteError` emits only the public `ErrorResponse` envelope |
| `ReadError` deserializes remote trace internals | `ReadErrorResponse` creates a new local typed error from the public code |
| Status-driven HTTP reconstruction | Code + status validation distinguishes categories sharing a status |
| Concrete type assertions | Behavior interfaces (`ErrorNotFound`, `HTTPErrorProvider`) that your own types can implement |
| `gravitational_trace.nocrypto` build tag drops `net/http` | Separate `tracehttp` package; the root package never imports `net/http` |
| `ConvertSystemError` | `ConvertSystemError` |
| `CompareFailed`, `OAuth2`, `Trust`, `Retry` | No equivalent |
| No equivalent | `slog`, generics, pipelines, and context integration |

## Migration from v1

v2 moves every `net/http`-dependent symbol into `tracehttp` and removes the
deprecated logger-coupled helpers. There are no compatibility shims in the root
package, because re-exporting the HTTP helpers would pull `net/http` back in
and undo the reason for the split.

```go
import (
    "github.com/tae2089/trace/v2"
    "github.com/tae2089/trace/v2/tracehttp"
)
```

| v1 | v2 |
| --- | --- |
| `trace.ToHTTPError` | `tracehttp.ToHTTPError` |
| `trace.ErrorResponseFor` | `tracehttp.ErrorResponseFor` |
| `trace.WriteError` | `tracehttp.WriteError` |
| `trace.ErrorMiddleware` | `tracehttp.ErrorMiddleware` |
| `trace.ReadErrorResponse` | `tracehttp.ReadErrorResponse` |
| `trace.FromHTTPResponse` | `tracehttp.FromHTTPResponse` |
| `trace.GetHTTPStatusCode` | `tracehttp.GetHTTPStatusCode` |
| `trace.IsHTTPError`, `trace.WrapHTTPError` | `tracehttp.IsHTTPError`, `tracehttp.WrapHTTPError` |
| `trace.NewClient` | `tracehttp.NewClient` |
| `trace.HTTPError`, `trace.ErrorCode`, `trace.Code*` | `tracehttp.HTTPError`, `tracehttp.ErrorCode`, `tracehttp.Code*` |
| `trace.HTTPStatusCode` interface | Removed — implement `tracehttp.HTTPErrorProvider` instead |
| `err.HTTPStatusCode()` / `err.HTTPError()` methods | Removed — call `tracehttp.ToHTTPError(err)` |
| `trace.WriteErrorWithLogger` | Removed |
| `trace.ErrorMiddlewareWithLogger` | Removed |
| `trace.RecoverMiddleware` | Removed |

`trace.WithField` and `trace.WithFields` rebuild an error chain when they
replace the inner `*TraceError`. Built-in wrapper types and
`tracehttp.WrapHTTPError`'s status override survive this. A third-party wrapper
type is dropped unless it implements `trace.TraceErrorReplacer` — one method,
`ReplaceTraceError(original, replacement *TraceError) (error, bool)`, that
rebuilds the wrapper around the replacement.

## Requirements

- Go 1.25+ (module requirement in `go.mod`)
- Key feature minimum versions:
  - `iter.Seq` / `for range N`: Go 1.23+
  - `context.WithoutCancel`, `context.AfterFunc`, `context.WithTimeoutCause`: Go 1.21+
  - `slog.DiscardHandler`: Go 1.24+
  - `log/slog`: Go 1.21+
  - `errors.Join` (`Aggregate`): Go 1.20+
  - Generics: Go 1.18+

## Changelog

### v2.0.0

- **Breaking**: Module path is now `github.com/tae2089/trace/v2`
- **Breaking**: Every `net/http`-dependent symbol moved to the `tracehttp` package
- **Breaking**: `HTTPStatusCode()` and `HTTPError()` methods removed from the typed errors;
  `tracehttp` classifies through the behavior interfaces instead
- **Breaking**: `HTTPStatusCode` interface removed; `HTTPErrorProvider` is the single extension point
- **Breaking**: `WriteErrorWithLogger`, `ErrorMiddlewareWithLogger`, and `RecoverMiddleware` removed
- **Added**: `ConvertSystemError(err)` for `os`, `io/fs`, and `syscall` failures
- **Added**: `Canceled(err, msg)` and `WrapLimitExceeded(err, msg)` constructors
- **Added**: `CaptureFrame(skip)` is now exported so other packages can build trace errors
- **Added**: `TraceErrorReplacer` hook so wrapper types outside the package survive
  `WithField`/`WithFields`; `tracehttp.WrapHTTPError`'s status override now survives them
  (it was silently dropped in v1)
- **Breaking**: `Frame` stores only a program counter; `Function`, `File`, and `Line` are
  methods now, and symbol resolution happens at render time. `MarshalJSON` keeps the
  `{"function","file","line"}` wire shape
- **Performance**: `Wrap` 435→162 ns and 6→2 allocs; a 10-deep wrap chain 4.9µs→1.9µs;
  `IsNotFound` on a 10-deep chain 638→102 ns; the structured fields map is allocated
  lazily and `errors.As` was replaced with a reflection-free chain walk on hot paths
- **Added**: Apache-2.0 `LICENSE` and a CI workflow that rejects `net/http` in the root package
- **Changed**: `ToHTTPError` documents and tests outermost-wins classification ordering

### v1.2.0

- **Added**: Safe `HTTPError`, `ErrorResponseFor`, and `WriteError` response contract
- **Added**: `ReadErrorResponse(statusCode, body)` for safe code-based typed error restoration
- **Changed**: Authentication and authorization now use distinct 401/403 categories and stable codes

### v1.1.0

- **Added**: `Errors(err) iter.Seq[error]` — range-based error chain iterator (Go 1.23+)
- **Added**: `DetachedContext(ctx)` — `context.WithoutCancel` wrapper for cancel-free child contexts
- **Added**: `Contextualizer.OnCancel(fn)` — `context.AfterFunc` wrapper for cancel callbacks
- **Added**: `WithCancelCause(ctx)` / `WithTimeoutCause(ctx, d, cause)` — context cause wrappers
- **Changed**: `FromContext` now uses `context.Cause` to preserve the original cancellation reason
- **Changed**: `NewErrorHandler(nil)` uses `slog.DiscardHandler` instead of panicking
- **Changed**: Benchmarks use `for range N` (Go 1.22+)

### v1.0.2 (2026-01-25)

- **Changed**: Error message formatting now uses `\n→ ` separator for better readability of error chains.

### v1.0.0

- Initial release with core error handling features

## License

Apache-2.0. See [LICENSE](LICENSE).
