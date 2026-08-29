# CCG Lint & Annotation Guide for Trace v3

Trace v3 is a protocol-independent error-tracking core. CCG annotations should
help future maintainers find and preserve that boundary, not document
application policy that no longer belongs in this package.

## What To Annotate

- The four public functions: `New`, `Errorf`, `Wrap`, and `Wrapf`.
- Internal helpers whose behavior protects the core contract, such as caller
  program-counter capture, standard error tree traversal, cycle protection, and
  detailed debug formatting.
- Package-level intent when it clarifies that Trace records origin and
  propagation without interpreting error meaning.

Good annotations explain why the symbol exists and which contract it preserves:

```go
// @intent add one propagation context and call site without changing the wrapped error's meaning.
func Wrap(err error, message string) error
```

## What To Skip

- Mechanical methods such as `Error`, `Unwrap`, and `Format` when the nearby
  type or helper already explains the contract.
- Test-only helpers and fixtures.
- Generated docs under `docs/`; regenerate them with CCG instead of hand
  editing.
- Any typed error, HTTP, slog, context, generics, retry, metric, response, or
  middleware examples as Trace-owned APIs. Those are application policy.

## v3 Boundary Checks

When reviewing CCG lint findings, keep these checks in mind:

- Root production code should export only `New`, `Errorf`, `Wrap`, and `Wrapf`.
- Root production code should not import `net/http`, `log/slog`, `context`, or
  other policy-heavy packages for application handling.
- `errors.Is`, `errors.As`, and `errors.Unwrap` are the inspection contract.
- `%+v` is a human-readable diagnostic format, not a machine-readable schema.
- `docs/migration-v3.md` may show short application-owned migration examples,
  but README and the runnable example should teach protocol-free error
  tracking.

## Commands

Run these after source or documentation changes:

```sh
ccg build --namespace trace .
ccg docs --namespace trace
ccg lint --namespace trace
```

`ccg lint` should finish with zero stale, orphan, missing, dead-ref,
contradiction, and drifted issues before release work continues.
