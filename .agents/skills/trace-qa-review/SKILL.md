---
name: trace-qa-review
description: "Worker procedure for QA review of Trace code, docs, examples, CCG freshness, artifacts, and Agent Team runtime evidence when routed by the Trace orchestrator. Use for review, audit, regression checks, integration coherence, follow-up verification, and partial reruns. For implementation or docs production, route through trace-orchestrator first."
---

# Trace QA Review

## Purpose

Act as an integration quality gate for Trace changes. QA verifies contracts and evidence; it does not simply confirm that files exist.

## Inputs

- Producer artifact paths and claimed changes.
- Changed files, expected behavior, and commands already run.
- Prior QA artifacts for follow-up or partial reruns.
- Runtime fields only when the orchestrator assigns durable Agent Team work: `RUN_ID`, `TASK_ID`, `AGENT`, `ARTIFACT_ROOT`, and optional `AGENT_TEAM_STATE_DIR`.

## Workflow

1. Read producer artifacts and inspect the referenced source or docs.
2. Check cross-boundary contracts:
   - typed error category to predicate and HTTP status
   - TraceError wrapping to `errors.Is` and `errors.As`
   - context cancellation cause to `FromContext`
   - slog value to handler output
   - README/examples to exported API
3. Verify command evidence. Prefer `go test ./...` for code changes and `ccg build/docs/lint --namespace trace` for CCG-affecting changes.
4. Return one verdict:
   - `PASS`: evidence is sufficient.
   - `FIX`: producer should retry with specific instructions.
   - `BLOCKED`: missing decision, permission, owner, or unverifiable dependency.
5. Do not pass missing artifacts, skipped tests without reason, or broken source/docs links.

## Output Format

Write `_workspace/trace-harness/{run_or_task}/trace-qa-review.md` or the artifact path provided by the orchestrator:

- Verdict: PASS | FIX | BLOCKED
- Evidence
- Findings
- Contract Checks
- Recommendation

## Validation

- Every verdict cites inspected files, artifacts, or commands.
- FIX includes exact retry instructions.
- BLOCKED includes a concrete blocked reason.
- Runtime task completion, when active, includes evidence and an artifact path.
