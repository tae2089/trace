# Trace Domain Rules

## Package Contract

Trace is a Go 1.25 error-tracking package. It records where an error originated
and how it propagated without interpreting or changing what the error means.
Its durable contracts are:

- error wrapping preserves the original cause and standard Go traversal with
  `errors.Is`, `errors.As`, and `errors.Unwrap`
- diagnostic context may be added without changing application-owned semantics
- stack frames point to useful origin and propagation locations
- formatting exposes the collected trace for debugging

Domain error categories, public error codes, protocol mappings, client response
shapes, retry decisions, logging policy, metrics, middleware, and panic recovery
belong to applications. Do not add `StatusCode`, `CodedError`, HTTP response
mapping, or generic policy-extension interfaces to Trace. An API belongs in
Trace only when it directly improves protocol-agnostic error tracking and has
behavior that should not vary between applications.

## Source Areas

| Area | Primary Files | Checks |
| --- | --- | --- |
| Core wrapping | `trace.go` | nil behavior, frame accumulation, fields, formatting, traversal |
| Typed errors | `errors.go` | compatibility review and migration away from application-owned semantics |
| Integrations | integration packages | compatibility review and migration away from application-owned policy |
| Examples/docs | `README.md`, `example/main.go`, `docs/` | package path, runnable examples, generated CCG freshness |

## Validation Defaults

- For code changes: run `go test ./...`.
- For CCG-affecting changes: run `ccg build --namespace trace .`, `ccg docs --namespace trace`, and `ccg lint --namespace trace`.
- For docs-only changes: verify source references and run CCG docs/lint when generated docs are affected.
- For harness-only edits: static validation and dry-run validation are sufficient unless executable code changes.

## Artifact Expectations

Substantial Trace harness work should write:

- producer result artifacts for code or docs phases
- QA artifact with `PASS`, `FIX`, or `BLOCKED`
- final orchestrator summary that names changed files, commands, artifacts, and skipped validation
