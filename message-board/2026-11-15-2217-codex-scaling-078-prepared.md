# SCALING-078 — Persistent CLI and complete-path validation prepared

- UTC: 2026-11-15T22:17:24.112986+00:00
- Agent: Codex
- Status: Handoff (bounded submission follows)

Implemented [pir-preprocess CLI](../cmd/pir-preprocess/README.md) around existing
shared preparation and artifact APIs, without changing mathematical kernels,
representation, CPU client paths or GPU layout defaults. Explicit resource/
backend/spec flags; source/executable/plugin SHA256; trusted support/root/code
manifests; advisory exclusive lock; file/directory sync and atomic directory
publication; retained incomplete stages and unique CPU profiles. No completed
artifact overwritten. Streams final data and readback without a resident copy.

Fresh builds certify discovery and full P·G identities. Complete-cache restart
verifies all four payloads without compute setup; root restart loads bound support/
root, skips discovery/root construction, then rechecks full identities and spectrum.
Sampling/root-intermediate restart granularity is documented honestly. SIGTERM
stops profiling and retains partial stages. CLI/library/native phase/byte progress
and controller resource heartbeats remain distinct, with unknown remaining work.

Changes: cmd/pir-preprocess main/artifact I/O and unit/reference tests/README;
new [complete workflow experiment](../experiments/gpu-preprocessing-workflow/report.md)
with guarded run.py/run.sbatch/analyze.py and local evidence; main README,
preprocessing guide, scaling plan/reference. Earlier staged calibrated results
remain included for reviewable provenance. README checked and updated for the
new public command and its experimental/source-only s17 limits.

Verification: short CUDA-free and native command tests pass, including independent
Go canonical bytes, actual fresh/root/final restarts, corrupt cache and mismatched
root parameters, interrupted publication retention, no overwrite and device bounds.
Python AST/runner --check, bash -n, gofmt and git diff --check pass. Exactcommands,
environment/source hashes and logs in experiment local-checks.json. Complete CUDA
CLI/race/ASan/vet/sanitizer validation remains pending the bounded job.

Prepared fixed s8/s4 executable validation with1/4GPU peer/local, retained profiles/
cache and independent Go fixture identity from74635. Same agreed4H200/4GiB,
4CPU/8GiBhost/6GiBRSS/2GiBcache/4GiBscratch,600/900/1200s caps; not fulls17.
No useful cache/profile deletion; test-only temporary profiles explicitly excluded.

Material cost: streaming persistence writes and reads each fresh payload once,
plus syncs. Projected s17 retained support/root/code/G~235GiB before headroom,
failed stages and profiles; full host peak and I/O latency unknown. No compression,
slower memory default or full-resource admission. User permits8/16GPUs if needed,
but current same-node engine on observed4GPU nodes does not implement cross-node
communication; capacity models96/48/24GiB percard are projections, not support.

Next commit exact source, submit bounded workflow validation, recordID and yield
without monitoring until user completion ping. Fulls17 RAM/storage/time agreement
and guarded full launch remain separate after validation; no further tile benchmark.
