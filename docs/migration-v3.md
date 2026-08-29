# Migrating to Trace v3

Trace v3 keeps only protocol-independent error tracking. The package records
where an error was created and where context was added, while the application
keeps ownership of error meaning and handling policy.

## Import Path

```go
import "github.com/tae2089/trace/v3"
```

## Core Replacements

Use only the four core functions:

```go
err := trace.New("operation failed")
err := trace.Errorf("query user %s: %w", userID, cause)
err = trace.Wrap(err, "load user")
err = trace.Wrapf(err, "load user %s", userID)
```

`Wrap` and `Wrapf` return `nil` when the original error is `nil`. If there is
no meaningful context to add, return the original error instead of adding a
location-only wrapper.

## Moved Responsibilities

| v2 feature area | v3 direction |
| --- | --- |
| Typed errors and predicates | Define application sentinels or custom error types. Use `errors.Is` and `errors.As`. |
| HTTP status and response helpers | Keep status mapping, response serialization, request logging, and recovery in the application. |
| slog helpers and fields | Decide log attributes and levels in the application. Use `%+v` only where detailed diagnostics are useful. |
| context helpers | Use the standard `context` package directly and wrap returned errors where they cross a boundary. |
| generics helpers | Use ordinary `(T, error)` flow or application-owned helpers. |
| system error conversion | Match standard library errors with `errors.Is` and map them in the application. |
| inspectors and iterators | Use `errors.Unwrap`, `errors.Is`, and `errors.As`. |

## Application-Owned HTTP Adapter

The HTTP contract is an application decision. One service may expose stable
public codes, another may hide every detail behind a generic message, and a
third may use a framework-specific response object. Trace should not choose
between those policies.

```go
type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func getUser(id string) error {
	return trace.Errorf("get user %s: %w", id, &APIError{
		Status:  http.StatusNotFound,
		Code:    "user_not_found",
		Message: "user not found",
	})
}

func writeHTTPError(w http.ResponseWriter, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		apiErr = &APIError{
			Status:  http.StatusInternalServerError,
			Code:    "internal",
			Message: "internal server error",
		}
	}

	w.WriteHeader(apiErr.Status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    apiErr.Code,
			"message": apiErr.Message,
		},
	})
}
```

The application can still log the internal trace separately:

```go
logger.Error("request failed", "trace", fmt.Sprintf("%+v", err))
```

## Formatting

Plain formatting stays clean:

```go
fmt.Println(err)
fmt.Printf("%v\n", err)
fmt.Printf("%s\n", err)
```

Detailed diagnostics are opt-in:

```go
fmt.Printf("%+v\n", err)
```

The `%+v` output is deterministic and testable, but it is a human-readable
debug representation, not a stable machine-readable schema.
