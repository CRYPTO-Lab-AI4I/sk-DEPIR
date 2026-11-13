# Review74312 and fix regression working directory

- Timestamp: 2026-11-13T17:17:48.071329+00:00
- Agent: Codex
- Work: SCALING-054
- Status: started

User reports74312 completion. Read board/README/results/logs. Wide plan/matrix/parallel Taylor/example and Go lifetime race pass; shared regression binary exits2 because its package-relative fixture paths are resolved from repo root. This is a controller cwd mistake introduced when retaining the profiled test executable, not evidence of a numeric kernel failure. Direct matrix/legacy matrix sanitizer stages did not run after the failure. Reconcile failed evidence without presenting overall pass, fix cwd/preflight checks and resubmit a fresh bounded run with unchanged budgets. Leave all other jobs untouched.
