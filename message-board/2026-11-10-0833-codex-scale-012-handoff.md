# SCALE-012 — Full run stopped; discovery preserved

Timestamp: 2026-11-10T08:33:56.650639+00:00
Agent: Codex
Status: handoff

User authorized stopping job 72768. Slurm confirms terminal FAILED/CANCELLED after 8h58m32s. Discovery completed in 2h56m37.545s, K=918,672,421, N=2^32. The 3,674,689,684-byte support file remains in `.cache/liftrs/hpc-s16/s16-m2-t5/`; streaming SHA-256 matches its saved metadata. Root-product progress reached 124,780,544 positions. Recent rate implied roughly 36 further hours for that loop, an extrapolation exceeding the budget. Peak sampled process-group RSS was 80.06 GiB.

Evidence: [stop record](../experiments/hpc-s16/runs/first-01/interruption.json), [recording script](../experiments/hpc-s16/record-stop.py), [report](../experiments/hpc-s16/report.md). Original results.json remains unchanged and says running because batch-shell termination prevented controller finalization. The active preparation CPU profile is empty; prior phase/heap profiles remain. The launcher now execs Python so batch TERM reaches its existing guarded cleanup; the analyzer recognizes the separate stop record. No cache was removed. README and handoff checked/updated. No full preparation, encoding, TRL or catalogue claim. No full retry; changed-source cache provenance must be reviewed explicitly before support reuse.
