---
name: trace-docs-workflow
description: "Worker procedure for Trace README, examples, CCG generated docs, and usage-guide maintenance when routed by the Trace orchestrator. Use for docs updates, stale example audits, command verification, generated docs refresh, follow-up refinements, reviews, and partial reruns. For code implementation, route through trace-orchestrator or trace-api-workflow first."
---

# Trace Docs Workflow

## Purpose

Keep user-facing Trace documentation accurate against source behavior. Documentation should teach the package API without inventing behavior or drifting from examples and tests.

## Inputs

- Requested documentation area or example.
- Source files, generated docs, and tests that define the behavior.
- Prior artifacts when this is a follow-up or partial rerun.
- Runtime fields only when the orchestrator assigns durable Agent Team work: `RUN_ID`, `TASK_ID`, `AGENT`, `ARTIFACT_ROOT`, and optional `AGENT_TEAM_STATE_DIR`.

## Workflow

1. Verify source behavior before changing prose.
2. Check package paths and commands. Use `github.com/tae2089/trace` in examples unless the task explicitly asks for another module path.
3. For generated docs, prefer:
   - `ccg build --namespace trace .`
   - `ccg docs --namespace trace`
   - `ccg lint --namespace trace`
4. Do not hand-edit generated CCG docs when regeneration is the correct source of truth.
5. Preserve unaffected sections during follow-up edits and record what changed.
6. Run `go test ./...` when examples or docs changes imply code behavior changes.

## Output Format

Write `_workspace/trace-harness/{run_or_task}/trace-docs-workflow-result.md` or the artifact path provided by the orchestrator:

- Summary
- Source references checked
- Documentation changed
- Commands run
- Generated-doc impact
- Risks or follow-up

## Validation

- Examples compile conceptually and use current exported names.
- README, examples, and generated docs agree on package path and behavior.
- CCG docs are refreshed or the reason for skipping refresh is recorded.
- Runtime task completion, when active, includes evidence and an artifact path.
