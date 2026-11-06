# SCALE-006 started

Timestamp: 2026-11-05T21:41:57+00:00. Agent: Codex.

Audit s=15 code-preparation memory before extending support or launching a large
run. Use retained s=14 profiles/results, actual source allocation lifetimes and
catalogue capacity uncertainty. Separate live coefficient payload, transient
allocation, GC-retained memory and process RSS. Create a reproducible analytical
report/script under `experiments/preparation-memory/`, then update the scaling
plan and README with a measured-versus-estimated admission decision. This work
item does not authorize a full s=15 benchmark or change its current guards.
