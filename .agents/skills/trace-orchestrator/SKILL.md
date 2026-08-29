---
name: trace-orchestrator
description: "Trace Codex harness orchestrator. Routes Trace Go error-handling library work: implementation, API compatibility, typed errors, HTTP/slog/context integration, README/examples, CCG docs, QA review, audit, update, modify, refine, retry, rerun, partial rerun, resume, and previous-artifact follow-up. Simple questions may be answered directly; full workflows use this orchestrator before worker skills."
---

# Trace Codex Orchestrator

## Route

| Request | Action |
| --- | --- |
| Simple question, file lookup, or one-line explanation | Answer directly after inspecting relevant files when needed. |
| Core API implementation or behavior fix | Use `trace-api-workflow`; route to `trace-api-maintainer` when delegation is allowed or useful. |
| HTTP, slog, context, or example integration work | Use `trace-api-workflow` plus `trace-integration-specialist` for boundary checks. |
| README, examples, generated CCG docs, or usage guidance | Use `trace-docs-workflow`; route to `trace-docs-curator` when delegation is allowed or useful. |
| Review, audit, regression check, or release confidence | Use `trace-qa-review`; route to `trace-qa-reviewer` as the quality gate. |
| Orchestrated durable workflow | Load Agent Team runtime skills, resolve or create internal runtime context, create tasks, collect evidence, and integrate artifacts. |
| Harness setup/edit/audit | Inspect local harness files only; do not call `agent-team` runtime state. |

## Specialist Roster

| Agent | Role | Model | Reasoning | Sandbox | Output |
| --- | --- | --- | --- | --- | --- |
| `trace-api-maintainer` | Core Go API implementation and compatibility | `gpt-5.5` | `high` | `workspace-write` | `${ARTIFACT_ROOT}/${TASK_ID}_trace-api-maintainer-result.md` |
| `trace-integration-specialist` | HTTP, slog, context, and example boundary behavior | `gpt-5.5` | `high` | `workspace-write` | `${ARTIFACT_ROOT}/${TASK_ID}_trace-integration-specialist-result.md` |
| `trace-docs-curator` | README, examples, and CCG docs alignment | `gpt-5.5` | `medium` | `workspace-write` | `${ARTIFACT_ROOT}/${TASK_ID}_trace-docs-curator-result.md` |
| `trace-qa-reviewer` | Integration QA, evidence checks, and regression review | `gpt-5.5` | `high` | `workspace-write` | `${ARTIFACT_ROOT}/${TASK_ID}_trace-qa-reviewer-result.md` |

## Execution Modes

- `direct`: default for tightly coupled fixes, small questions, and local harness maintenance. The orchestrator performs the work using relevant skills and writes artifacts when useful.
- `delegated`: use only when the user asks for agents/delegation/parallel work or when independent implementation, docs, and QA lanes materially improve quality and the active Codex environment permits delegation.
- `hybrid`: use for code changes followed by independent QA, docs refresh followed by QA, or parallel analysis followed by direct synthesis.

Ordinary workers do not spawn subagents. Only this orchestrator may delegate, and only within active Codex tool policy.

## Runtime Contract

- Runtime coordination is skill-first: load runtime recipe/service/helper skills, then use daemonless `agent-team` commands only through those helper contracts.
- `agent-team` stores run/task/message/inbox/sync state only. Codex orchestrator and workers perform the actual work.
- Load `agent-team-shared` first for global runtime behavior.
- Use recipes for workflow shape: `recipe-agent-team-run-lifecycle` for full runs, `recipe-agent-team-worker-checkpoint` for worker checkpoints, and `recipe-agent-team-operational-audit` for audit/status/cleanup.
- For exact command behavior, load helper skills such as `agent-team-run-create`, `agent-team-task-create`, `agent-team-task-complete`, `agent-team-sync-check`, `agent-team-message-send`, or `agent-team-event-log`.
- Use service skills only for navigation: `agent-team-run`, `agent-team-task`, `agent-team-inbox`, `agent-team-sync`, and `agent-team-ops`.
- Setup, edit, and audit-only harness requests do not probe runtime state.
- `RUN_ID` and `TASK_ID` are orchestrator-owned internal context, not required user input.
- Resolve context in this order: active in-session context, advanced/debug user-provided IDs, recent open run plus previous artifacts, user choice among ambiguous recent runs, then a new generated-ID run.
- If no runtime context is available for an orchestrated run, load `agent-team-run-create` and `agent-team-task-create`, then create one run and task records without explicit IDs and capture returned JSON IDs.
- Workers receive orchestrator-supplied `RUN_ID`, `TASK_ID`, `AGENT`, `ARTIFACT_ROOT`, and optional `AGENT_TEAM_STATE_DIR` only for assigned durable tasks.
- Assign `ARTIFACT_ROOT` as `_workspace/trace-harness/{RUN_ID}` unless the run explicitly needs a different artifact root.
- Workers update only their assigned task and write their own completion evidence through `agent-team task complete`.
- The orchestrator logical runtime recipient is `trace-orchestrator`; workers send progress or coordination messages there.
- The orchestrator creates tasks, assigns worker context, verifies worker-written evidence, checks inbox/sync status, decides retry/reassign/block paths, and integrates artifacts.
- `_workspace/trace-harness/` stores artifacts, reports, inputs, and generated outputs only.

## Workflow

1. Classify the request by target surface: core API, integration, docs, QA, harness maintenance, or runtime execution.
2. Check context:
   - active runtime context: resume it
   - user-provided `RUN_ID` or `TASK_ID`: inspect and resume/check state before creating new work
   - `_workspace/trace-harness/` exists and user asks for follow-up: preserve unaffected artifacts and rerun only requested scope
   - new unrelated input: create a new artifact subdirectory
3. Load only relevant worker skills and references:
   - `trace-api-workflow` for code behavior
   - `trace-docs-workflow` for docs and generated CCG docs
   - `trace-qa-review` for review and audit
   - `references/domain-rules.md` for Trace-specific contracts
4. Select pattern and mode:
   - `pipeline`: implementation or docs change then QA
   - `producer_reviewer`: any meaningful code/docs change
   - `fan_out_fan_in`: independent API, integration, and docs research before synthesis
   - `expert_pool`: route a small task to one specialist
5. Execute direct work or delegate bounded worker tasks when allowed. In delegated runtime mode, do not perform worker-owned implementation directly except to integrate or reconcile returned artifacts.
6. Require artifacts for substantial work:
   - `_workspace/trace-harness/{run}/00_input/`
   - `_workspace/trace-harness/{run}/{phase}_{agent}_{artifact}.md`
   - `_workspace/trace-harness/{run}/{task_id}_result.md`
7. Validate before final response:
   - `go test ./...` for code or example changes
   - `ccg build --namespace trace .`, `ccg docs --namespace trace`, and `ccg lint --namespace trace` for CCG-affecting changes
   - static harness validation for harness-only edits
8. Retry failed worker phases at most two times after the initial failure. Change prompt or scope on retry. If unresolved, block with a concrete reason.
9. Report final artifacts, commands run, changed files, skipped checks, and unresolved risks.

## Data Flow

| Phase | Inputs | Mode | Specialists | Output | Evidence | Next Consumer | Failure Path |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Classify | User request, repo files, CCG docs | direct | orchestrator | routing note | inspected files/searches | active workflow | ask only if acceptance or safety boundary is unknowable |
| Produce | Target source/docs, prior artifacts | direct or delegated | selected producer | producer artifact | worker-written task evidence, changed files, tests or source references | QA | retry with narrower scope |
| Review | Producer artifact and changed files | direct or delegated | `trace-qa-reviewer` | QA artifact | PASS/FIX/BLOCKED with worker-written evidence | orchestrator | FIX retry or BLOCKED reason |
| Integrate | Artifacts, task records, user output needs | direct | orchestrator | final summary or requested file | verified evidence, commands, artifacts, risk notes | user | report missing evidence |

## Follow-Up Behavior

- For update, modify, refine, retry, rerun, partial rerun, review, audit, or previous-result requests, inspect prior artifacts first.
- Preserve unaffected sections and cite the changed scope.
- If multiple plausible open/recent runs exist, ask the user to choose by title, status, and artifact summary rather than raw ID.
- If the user provides a new input that does not depend on previous artifacts, create a new artifact subdirectory.

## Completion Rules

- Do not complete a substantial workflow without worker-written evidence and artifact paths.
- Do not silently advance past missing producer, docs, QA, CCG, or test evidence.
- Preserve conflicting findings with source attribution.
- Report skipped validation explicitly with the reason.

## Test Scenarios

### Normal Flow

Request: "Add a typed validation error and update docs."

Expected route: `hybrid` pipeline. `trace-api-maintainer` updates behavior and tests, `trace-docs-curator` updates README/generated docs if needed, `trace-qa-reviewer` verifies compatibility, examples, `go test ./...`, and CCG freshness.

### Failure Flow

Request: "Change HTTP error mapping."

Expected route: producer-reviewer. If tests fail, retry once with the failing assertion as scope. If status semantics are ambiguous after two retries, block with the missing decision and preserve the QA artifact.

### Follow-Up Flow

Request: "Only rerun the docs part based on the previous result."

Expected route: inspect `_workspace/trace-harness/`, preserve code artifacts, rerun only docs checks, refresh CCG docs if source docs changed, then run QA on docs/source alignment.

### Near-Miss Flow

Request: "Explain what `trace.Wrap` does."

Expected route: direct answer with source references; do not create runtime state or delegate unless the user asks for a durable workflow.

## References

- Trace contracts: `references/domain-rules.md`
- Trigger validation prompts: `references/trigger-tests.md`
