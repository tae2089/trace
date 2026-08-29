# AGENTS.md

## CCG Usage

- Use `trace` as the namespace for CCG.
- During development, use both the CCG MCP tools and grep-style search tools such as `rg`.
- After code changes, run the appropriate CCG `build`, `docs`, and `lint` workflows.
- Add suitable CCG annotations where they help clarify code intent and improve code search.

## Package Purpose and Responsibility Boundary

Trace exists to make Go errors easier to track. It records where an error
originated and how it propagated without interpreting or changing what the
error means.

Trace owns only error-tracking mechanics:

- preserve the original cause and standard Go traversal through `errors.Is`,
  `errors.As`, and `errors.Unwrap`
- attach diagnostic context to an error
- capture useful origin and propagation stack frames
- format the collected trace for debugging

Applications own error meaning and handling policy:

- domain error types, categories, and public error codes
- HTTP status codes, gRPC codes, CLI exit codes, and other protocol mappings
- client-safe response shapes and serialization or deserialization
- user-facing messages, retry decisions, logging policy, metrics, request IDs,
  middleware, and panic recovery

Do not add `StatusCode`, `CodedError`, HTTP response mapping, middleware, or
similar application-policy interfaces to Trace. Do not add generic extension
points whose primary purpose is to host those policies. Applications may use
Trace-wrapped errors in their own adapters because the standard Go error chain
must remain intact.

Use this decision rule for every proposed API: if it does not directly improve
protocol-agnostic error tracking, or its behavior can reasonably vary between
applications, it does not belong in Trace.

## Harness: Trace

**Goal:** Coordinate Trace Go error-handling library work across API correctness, integration behavior, docs/examples, CCG freshness, and QA.

**Trigger:** Use `.agents/skills/trace-orchestrator/SKILL.md` for Trace implementation, docs, review, audit, rerun, refinement, and orchestrated multi-specialist work. Simple questions may be answered directly.

**Model:** Follows the `revfactory/harness` orchestrator/specialist structure adapted to Codex-native skills, agents, artifacts, and Agent Team runtime.

**Orchestrator:** `.agents/skills/trace-orchestrator/SKILL.md`
**Agents:** `.codex/agents/`
**Artifacts:** `_workspace/trace-harness/`

**Runtime State:**

- Load `agent-team-shared` first for global runtime rules.
- Use recipe skills for workflow shape: `recipe-agent-team-run-lifecycle` for full runs, `recipe-agent-team-worker-checkpoint` for worker checkpoints, and `recipe-agent-team-operational-audit` for audit/status/cleanup.
- Use service skills for navigation: `agent-team-run`, `agent-team-task`, `agent-team-inbox`, `agent-team-sync`, and `agent-team-ops`.
- Use exact command helper skills for command syntax and flags, for example `agent-team-task-complete`, `agent-team-sync-check`, `agent-team-message-send`, or `agent-team-event-log`.
- `RUN_ID` and `TASK_ID` are orchestrator-owned internal context, not required user input.
- If the user provides an advanced/debug `RUN_ID` or `TASK_ID`, inspect and resume that run/task before creating new state.
- If no runtime context is available, the orchestrator loads `agent-team-run-create` and `agent-team-task-create`, then creates one generated-ID run and generated-ID task records through those helper contracts.
- Do not use runtime state during harness setup, editing, audit-only work, simple one-shot answers, or explicitly local-only runs.
- `agent-team` stores state only; Codex orchestrator and workers perform the actual work.
- Orchestrator owns run creation, task creation, worker context assignment, evidence verification/aggregation, inbox/sync checks, retry/reassign/block decisions, and artifact integration.
- Orchestrator assigns `ARTIFACT_ROOT` as `_workspace/trace-harness/{RUN_ID}` unless a run explicitly needs a different artifact root.
- Workers update only their assigned task and write their own completion evidence through `agent-team task complete`.
- Workers send runtime progress or coordination messages to `trace-orchestrator`.
- Completed tasks require evidence and an artifact path.
- Blocked tasks require a concrete blocked reason.
- `_workspace/` is for artifacts and reports only.

**Change History:**
| Date | Change | Target | Reason |
| --- | --- | --- | --- |
| 2026-05-10 | Initial Trace harness | `.codex/agents/`, `.agents/skills/trace-orchestrator/`, `.agents/skills/trace-api-workflow/`, `.agents/skills/trace-docs-workflow/`, `.agents/skills/trace-qa-review/` | Add Codex-native Agent Team harness for Trace development and QA |
| 2026-05-10 | Harden runtime contract | `AGENTS.md`, `.agents/skills/trace-orchestrator/`, `.codex/agents/` | Clarify `agent-team` as state-only, worker-owned evidence, artifact root, and orchestrator message recipient |
| 2026-08-29 | Define HTTP boundary ownership | `AGENTS.md` | Keep application-specific HTTP policy out of Trace while retaining safe, opt-in translation primitives |
| 2026-08-29 | Narrow package responsibility | `AGENTS.md`, Trace domain rules | Limit Trace to protocol-agnostic error tracking and move all error meaning and handling policy to applications |
