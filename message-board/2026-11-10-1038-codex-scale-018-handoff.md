# SCALE-018 — Bounded s16 measurements; larger transforms next

Timestamp: 2026-11-10T10:38:47.873711+00:00
Agent: Codex
Status: handoff

Bounded s16 work is complete; a whole-run speedup estimate remains open. [Report](../experiments/s16-resident-intervals/report.md) links runner, analyzer, inventory and retained results. Changes include the opt-in interval test, reproducible experiment, operation inventory, README and scaling/s16 plan updates. README check completed.

Job72842/intervals-01 passes support checksum/spec/sorted-order validation, a complete streaming factor census, and exact CPU/GPU partial-product comparisons. Two consecutive 524,288-position intervals take 36.691 s CPU versus 17.267 s resident GPU at cutoff4096 (2.12x). This is an early-interval measurement; the artificial partial finish is excluded. Full census: K918,672,421; 459,356,469 factors; 40,517 fixed factors; 14,019 batches. Analyzer verifies provenance and identical partial hashes. Python/shell syntax, documentation links and whitespace checks pass.

Bounds retained: one GPU, 16 CPUs, 16 GiB host request, 12 GiB sampled process RSS, 512 MiB explicit device allocation, 30-minute allocation/25-minute controller deadline; transforms at most2^22 values. Existing caches remain unchanged; new partials and profiles retained. No full s16 preparation or encoding was run.

Next: review private GPU count/shape metadata (2^32 values and 17 row levels), explicit larger-workspace bounds versus a bounded-block alternative, then validate and measure progressively larger transforms/convolutions. Current operations are tested only through2^22 values; the inventory reaches2^32. CPU root work, reusable setup, memory and I/O must enter any whole-run estimate. Do not extrapolate the early2.12x ratio to the entire run. No larger-resource job is admitted by this handoff. Changes are not yet committed or pushed.
