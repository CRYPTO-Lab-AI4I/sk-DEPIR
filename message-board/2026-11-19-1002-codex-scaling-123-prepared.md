# Targeted transfer capture ready within existing envelope

- Timestamp: 2026-11-19T10:02:46.453938+00:00
- Agent: Codex
- Work: SCALING-123
- Status: Prepared

[Report](../experiments/root-transfer-diagnostics/report.md),
[runner](../experiments/root-transfer-diagnostics/run.py),
[readiness](../experiments/root-transfer-diagnostics/readiness/checks.json).
Explicit diagnostic test flag exchanges64-byte pre/post SHA256 proofs before GPU
consumption, records buffer addresses and saves both mismatching payloads/proofs.
Post-save acknowledgement ensures both files durable before MPI abort. Uses
bounded canonical codec hash buffers, no full-array staging copy; extra CPU hashes/
messages alter timing/allocation and cannot justify a production default.

Original77704 evidence preserved and main gate remains failed/unresolved. Dedicated
cache/run name,12timed repetitions of largest case only, existing smaller checks,
full CPU equality and memcheck if reached. Larger warm-up shard/local snapshots
retained for failure binding. Captured bytes report first differing byte/word but
not cause; unreproduced instrumentation pass does not establish a repair.

Tagged binary compiles; pre/post proof symmetry and131073-word tail codec tests
pass. Snapshot difference/tail/length test passes, plus3tree/4distributed/5staged/
3shared reviewer regressions. Python/shell syntax, doc links and edited-text
whitespace pass. Local noGPU/MPI or root-cause claim. README/reference/plan and
experiment index checked/updated with pending diagnostics, original failure intact.

Same2HPCnodes/2H200/4logical ranks/4CPU/8GiB RAM,64MiB owned/rank≤128MiB/card,
1GiB RSS/rank,128MiB retained/2GiB scratch,15min scheduler; no budget increase.
Next commit exact source, verify frozen inputs and submit capture-01, then hand
off without polling/log-following/result inspection until user completion ping.
