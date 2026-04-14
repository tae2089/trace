# Trace - Modern Go Error Handling

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/trace.svg)](https://pkg.go.dev/github.com/yourusername/trace)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/trace)](https://goreportcard.com/report/github.com/yourusername/trace)

A modern error handling package for Go, inspired by [gravitational/trace](https://github.com/gravitational/trace) but upgraded for Go 1.25+ with:

- 🔍 **Stack traces** - Know exactly where errors originate
- 🏷️ **Typed errors** - NotFound, BadParameter, AccessDenied, etc.
- 📊 **slog integration** - Structured logging out of the box
- 🔗 **Full errors.Is/As support** - Compatible with Go 1.13+ error handling
- 🧬 **Generics** - Result types, pipelines, and type-safe operations
- 🌐 **HTTP utilities** - Middleware, error responses, status code mapping
- 📦 **Context integration** - Trace IDs, fields, cancel cause, detached contexts
- 🔄 **Error chain iterator** - `for e := range trace.Errors(err)` (Go 1.23+)

## Installation

```bash
go get github.com/tae2089/trace
```

## Quick Start

```go
package main

import (
    "errors"
    "fmt"
    "github.com/yourusername/trace"
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
    user, err := queryDatabase(id)
    if err != nil {
        return trace.Wrap(err, "failed to fetch user %s", id)
    }
    return nil
}

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
err := trace.Wrap(originalErr, "failed to process user %s", userID)

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
if trace.IsAccessDenied(err) { /* handle 403 */ }
if trace.IsRetryable(err) { /* retry the operation */ }

// Get HTTP status code
statusCode := trace.GetHTTPStatusCode(err) // e.g., 404, 403, 500
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
// Error middleware
http.HandleFunc("/users/{id}", trace.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
    user, err := service.GetUser(r.PathValue("id"))
    if err != nil {
        return err // Automatically converts to proper HTTP response
    }
    json.NewEncoder(w).Encode(user)
    return nil
}))

// With logging
logger := slog.Default()
http.HandleFunc("/users/{id}", trace.ErrorMiddlewareWithLogger(handler, logger))

// Panic recovery
http.Handle("/", trace.RecoverMiddleware(mux, logger))

// Create error from HTTP response
resp, _ := http.Get("https://api.example.com/users/123")
body, _ := io.ReadAll(resp.Body)
if err := trace.FromHTTPResponse(resp, body); err != nil {
    // err is typed (NotFound, BadParameter, etc.) based on status code
}
```

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
statusCode := trace.GetHTTPStatusCode(combined)
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

This package is mostly API-compatible with gravitational/trace. Main differences:

| gravitational/trace  | This package                            |
| -------------------- | --------------------------------------- |
| `trace.Traces` embed | Not needed - use `*TraceError` directly |
| `trace.OrigError()`  | Use `errors.Unwrap()` or `errors.Is/As` |
| Manual `SetTrace`    | Automatic via `Wrap()`                  |
| -                    | `slog` integration                      |
| -                    | Generic `Result` type                   |
| -                    | `Pipeline` pattern                      |
| -                    | Context integration                     |

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

Apache 2.0
