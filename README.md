# Trace

[![Go Reference](https://pkg.go.dev/badge/github.com/tae2089/trace/v3.svg)](https://pkg.go.dev/github.com/tae2089/trace/v3)
[![Go Report Card](https://goreportcard.com/badge/github.com/tae2089/trace/v3)](https://goreportcard.com/report/github.com/tae2089/trace/v3)

Trace is a small Go error-tracking package. It records where an error was
created and which application layers added context while keeping the original
cause available through the standard Go error chain.

Trace does not decide what an error means. Domain error types, public error
codes, HTTP status codes, retry policy, logging fields, metrics, and response
bodies belong to the application.

## Installation

```bash
go get github.com/tae2089/trace/v3
```

## Quick Start

```go
package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/tae2089/trace/v3"
)

var ErrUserNotFound = errors.New("user not found")

func main() {
	err := loadUser("user-123")
	if err == nil {
		return
	}

	if errors.Is(err, ErrUserNotFound) {
		fmt.Println("not found")
	}

	fmt.Println(err)
	fmt.Printf("%+v\n", err)
}

func loadUser(id string) error {
	if err := queryUser(id); err != nil {
		return trace.Wrapf(err, "load user %s", id)
	}
	return nil
}

func queryUser(id string) error {
	err := sql.ErrNoRows
	if errors.Is(err, sql.ErrNoRows) {
		return trace.Errorf("query user %s: %w", id, ErrUserNotFound)
	}
	return nil
}
```

Plain output contains only the context and cause message:

```text
load user user-123: query user user-123: user not found
```

Use `%+v` when you want the human-readable debug trace with source locations.
The exact layout is for people, not for machine parsing.

## API

v3 intentionally exposes only four functions:

```go
err := trace.New("not found")
err := trace.Errorf("query user %s: %w", id, cause)
err = trace.Wrap(err, "load user")
err = trace.Wrapf(err, "load user %s", id)
```

- `New` and `Errorf` create a new traced error and record the call site.
- `Errorf` supports `%w`, including multiple wrapped causes, like
  `fmt.Errorf`.
- `Wrap` and `Wrapf` add one context message and one propagation call site.
- Use meaningful non-empty context; if there is no context to add, return the
  original error.
- `Wrap(nil, ...)` and `Wrapf(nil, ...)` return `nil`.
- `Error()`, `%s`, and `%v` print only ordinary error text.
- `%+v` prints the collected trace frames for debugging.

All values are used as ordinary `error` values. Trace provides no public error
type, frame type, iterator, inspector, or policy interface.

## Standard Error Handling

Trace preserves standard Go traversal. Use application-owned sentinels or typed
errors and inspect them with `errors.Is` and `errors.As`.

```go
type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

func validateEmail(email string) error {
	if email == "" {
		return trace.Errorf("validate email: %w", &ValidationError{Field: "email"})
	}
	return nil
}

func handle() {
	err := trace.Wrap(validateEmail(""), "create account")

	var validation *ValidationError
	if errors.As(err, &validation) {
		fmt.Println(validation.Field)
	}
}
```

For multiple independent causes, use the standard library:

```go
err := trace.Wrap(errors.Join(cacheErr, databaseErr), "load profile")

if errors.Is(err, databaseErr) {
	// found through the standard error tree
}
```

`%+v` follows both `Unwrap() error` and `Unwrap() []error` branches so traced
errors inside joined trees remain visible.

## Debug Formatting

Use ordinary formatting for user-facing messages or low-detail logs:

```go
fmt.Println(err)
fmt.Printf("%v\n", err)
fmt.Printf("%s\n", err)
```

Use detailed formatting only at debugging or diagnostic log boundaries:

```go
fmt.Printf("%+v\n", err)
```

Trace captures only the program counter when errors are created or wrapped.
Function names, files, and line numbers are resolved only when detailed output
is rendered.

## Application Policy

Trace is deliberately protocol agnostic. Applications can still wrap and inspect
their own domain errors because the error chain remains intact.

```go
type DomainError struct {
	Kind     string
	Resource string
}

func (e *DomainError) Error() string {
	return e.Kind
}

func findUser(id string) error {
	return trace.Errorf("find user %s: %w", id, &DomainError{
		Kind:     "not found",
		Resource: "user",
	})
}

func handle(err error) {
	var domain *DomainError
	if errors.As(err, &domain) {
		// The application decides what "not found" means at each boundary.
	}
}
```

The same boundary applies to HTTP, gRPC, CLI exit codes, structured logging,
retry decisions, metrics, request IDs, and panic recovery.

## Migration From v2

v3 is a smaller package, not a feature-equivalent v2 replacement. General
development now focuses on v3. v2 receives only serious correctness and
security fixes.

Change the import path:

```go
import "github.com/tae2089/trace/v3"
```

Then move application policy out of Trace:

| v2 area | v3 direction |
| --- | --- |
| Typed errors and predicates | Define application sentinels or custom error types; use `errors.Is` and `errors.As`. |
| HTTP helpers and middleware | Put status mapping, response envelopes, request logging, and recovery in your HTTP adapter. |
| slog helpers and structured fields | Log `err` or `fmt.Sprintf("%+v", err)` according to your logging policy. |
| context helpers | Use the standard `context` package directly and wrap returned errors at the application boundary. |
| `Result`, `Pipeline`, and other generics | Use ordinary `(T, error)` flow or an application-owned helper. |
| system error conversion | Use `errors.Is` against standard library errors such as `fs.ErrNotExist`, then map in your application. |
| error iterators and inspectors | Use standard `errors.Unwrap`, `errors.Is`, and `errors.As`. |

See [Migrating to Trace v3](docs/migration-v3.md) for a short example of moving
HTTP response mapping into an application adapter.

## Migration From gravitational/trace

Trace v3 has a narrower goal than `gravitational/trace`.

| gravitational/trace pattern | Trace v3 direction |
| --- | --- |
| `trace.Wrap(err)` only to attach a location | Add meaningful context with `trace.Wrap(err, "context")`, or return the original error. |
| `trace.OrigError` / trace-specific inspection | Use `errors.Unwrap`, `errors.Is`, and `errors.As`. |
| Remote HTTP error serialization | Keep protocol response contracts in the application. |
| Trace-defined categories | Keep categories in application-owned error values. |

## Requirements

- Go 1.25+
- Standard library only

## Release Plan

The first v3 release should be `v3.0.0-rc.1`. Promote it to `v3.0.0` only after
Trace's tests pass and a real application migration verifies error creation,
multi-layer wrapping, `errors.Is`/`errors.As`, application-owned HTTP mapping,
and `%+v` diagnostic logging.

## License

Apache-2.0. See [LICENSE](LICENSE).
