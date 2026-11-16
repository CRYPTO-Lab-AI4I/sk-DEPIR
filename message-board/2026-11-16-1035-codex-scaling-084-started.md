# SCALING-084 — longer full run with incremental recovery

UTC: 2026-11-16T10:35:49.820716+00:00; agent: Codex; status: started.

User authorizes a longer replacement for job75819 and low-overhead progress persistence. Prepare 100h Slurm /98h controller /97h worker limits, keeping the approved compute, memory and storage caps. Keep the old job intact until a validated replacement is ready.

Add atomic rolling checkpoints at completed discovery samples, root-product batches and reciprocal iterations; preserve canonical support/root/code stages and independent correctness checks. Avoid saving transform workspace or introducing extra transforms. Document write overhead, restart granularity, bounded retention and unavailable in-flight recovery. Validate interruption/resume and corruption rejection on small instances before any replacement. Relevant paths: `encoding/liftrs-shared`, `cmd/pir-preprocess`, `experiments/gpu-s17-full-preprocessing`, `docs/preprocessing.md`, `docs/encoding-status.md`. README reviewed; update its persistence description. Do not monitor long jobs.
