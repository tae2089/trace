# Trace Harness Trigger Tests

## Should Trigger `trace-orchestrator`

- "Add a new typed error and update all docs."
- "Review the HTTP middleware behavior for trace errors."
- "Audit Trace for stale examples and generated CCG docs."
- "Rerun only the QA phase from the previous Trace artifact."
- "Refine the previous result and preserve the unchanged API sections."
- "Use agents to split API, docs, and QA for this Trace change."
- "Check whether context cancellation causes are preserved end to end."
- "Update README and run CCG docs/lint afterward."
- "Resume RUN_ID abc for the Trace harness and report blockers."
- "Do a release readiness review for this Go error package."

## Should Not Trigger `trace-orchestrator`

- "What is the capital of Korea?"
- "Generate a logo for Trace."
- "Explain Go errors.Is in general without looking at this repo."
- "Open localhost in the browser."
- "Install a new Codex skill from GitHub."
- "Show the current date."
- "Refactor an unrelated project outside this repository."
- "Run a shell command and paste its raw output."
- "Create a new Python web app."
- "Summarize this pasted article."

## Follow-Up Prompts

- "Only retry the docs lane."
- "Audit the previous QA result and tell me what is still blocked."
- "Update the same harness to add a release reviewer later."
