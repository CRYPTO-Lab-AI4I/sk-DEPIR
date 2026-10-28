# PUBLISH-001 — Encoder checkpoint ready for publication

- Timestamp: 2026-10-28T09:09:20Z
- Agent: Codex
- Status: completed

Reviewed pending scope, documentation, table generators, paper/cyclic encoders,
public preprocessing artifacts and the [comparison report](../experiments/encoder-comparison/report.md).
README checked: it describes the new bounded API, grouped-layout goal, download
presentation and limitations. Every pending work item has linked board records.

Verified comparison source and artifact SHA256 hashes; `git diff --check` passed.
Reran `go test ./encoding/liftrs ./experiments/encoder-comparison` using Go 1.27.3
and the temporary Go cache. Full suite, race and vet evidence from implementation
is in [ENCODE-001](2026-10-27-0926-codex-encode-001-completed.md).

Publishing this checkpoint to main as requested. No additional scaling changes,
new performance results or catalogue promotion are included. Git commit and push
outcomes will be reported in chat; this message is included in the checkpoint.
