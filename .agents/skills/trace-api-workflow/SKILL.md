---
name: trace-api-workflow
description: "Worker procedure for implementing and reviewing Trace Go API changes when routed by the Trace orchestrator. Use for error wrapping, typed error predicates, Result/Pipeline generics, compatibility fixes, test updates, follow-up refinements, and partial reruns. For full Trace workflows, route through trace-orchestrator; do not use for docs-only requests unless API behavior must be verified."
---

# Trace API Workflow

## Purpose

Maintain Trace core behavior without breaking public API contracts. The package is a Go 1.25 error-handling library, so changes must preserve standard error traversal, stack trace usefulness, and typed semantic checks.

## Inputs

- User request and acceptance criteria.
- Target files or symbols.
- Prior artifacts when this is a follow-up or partial rerun.
- Runtime fields only when the orchestrator assigns durable Agent Team work: `RUN_ID`, `TASK_ID`, `AGENT`, `ARTIFACT_ROOT`, and optional `AGENT_TEAM_STATE_DIR`.

## Workflow

1. Inspect source with `rg` and CCG namespace `trace` before editing.
2. Classify the behavior surface:
   - core wrapping and frame capture: `trace.go`
   - typed categories and HTTP status mapping: `errors.go`
   - generic Result/Pipeline helpers: `generics.go`
   - cross-boundary usage: `context.go`, `http.go`, `slog.go`
3. Preserve compatibility unless the task explicitly asks for a breaking change.
4. Prefer behavior tests over implementation tests. Cover `errors.Is`, `errors.As`, exported predicates, frame accumulation, fields, and nil behavior when affected.
5. If public docs or generated CCG docs become stale, route a docs task or update docs in the same run when the orchestrator asks for it.
6. Before completion, run the smallest relevant tests; use `go test ./...` for broad behavior changes.

## Output Format

Write `_workspace/trace-harness/{run_or_task}/trace-api-workflow-result.md` or the artifact path provided by the orchestrator:

- Summary
- Changed files
- Behavior contract
- Tests and commands
- CCG/docs impact
- Risks or follow-up

## Validation

- Exported API compatibility is preserved or explicitly documented.
- Typed errors remain discoverable through standard Go traversal.
- Tests cover the changed behavior.
- CCG annotations are added only where they clarify durable code intent.
- Runtime task completion, when active, includes evidence and an artifact path.
