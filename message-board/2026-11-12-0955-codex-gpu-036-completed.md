# GPU-036 — Streamed fast-layout validation complete

Timestamp: 2026-11-12T09:55:41.026831+00:00
Agent: Codex
Status: completed

Reviewed job73940 after user completion ping. Slurm COMPLETED/0:0 on gnode03,
H200 NVL,16 CPUs;16m22s versus conservative30–45-minute forecast, below2h budget,
no guard tripped. Final accounting saved with run, outside earlier sealed map.
No new job, default switch or resource increase.

[Report](../experiments/gpu-streamed-constants/report.md),
[analysis](../experiments/gpu-streamed-constants/runs/streamed-01/analysis.json),
[manifest](../experiments/gpu-streamed-constants/runs/streamed-01/results.json),
[accounting](../experiments/gpu-streamed-constants/runs/streamed-01/accounting.log).
Recorded-cache analyzer passes350 exact source identities plus sealed evidence/
cache hashes, full retained CPU/GPU canonical/packed agreement, every systematic
position,16 sampled curves and1027 packed readbacks. Uses the completed controller
payload hash with metadata/physical size checks, not another8 GiB bulk rehash.

Small Go oracle/independent K1 tile1/64/255/256 boundaries, tail/alias/field/product
range checks, source immutability, retained ownership/error recovery/Close/old ABI
pass. Both memchecks0 errors and0 leaked bytes. All80 matrix regressions,
ordinary CPU/CUDA-only/combined/CGO-disabled suites, focused Go race/vet pass.

Measured device34,364,742,932 bytes (32.005 GiB), exact predicted savings
20,854,558,872 bytes /19.422 GiB /37.77% versus prior reuse51.427 GiB. Host RSS
39.029 GiB unchanged. Code upload20,854,558,868 bytes /19.422 GiB per encode,
64 scratch-borrowed tiles of256 MiB,2.234/2.239 s. No new GPU tile allocation,
extra host result, code disk reads or format change. Packed output8 GiB retained.

Two full encodes63.487/58.851 s; native56.939/56.991 s, rows23.65–23.71 s,
columns20.57–20.58 s. Outer allocation/telemetry differences are not attributed to
one cause. Historical prior reuse57.418/56.975 s on gnode10 is not a matched
s16 comparison. Matched s14 resident5.1790/5.1759 s versus streamed5.5273/5.4666 s
shows6.72%/5.62% slowdown (two fixed-order samples). Cold/warm load bias and other
load/setup/verification/packing/I/O costs remain separate. Faster resident remains
default when memory fits; streaming remains explicit optional CLI mode.

Updated root README, CUDA/shared API README and scaling plan, replacing pending
claims with measured evidence and limitations; README check complete. Existing
preprocessing/profiles/binaries/output retained. No kernel changes on review.
Self-checks, local report links and whitespace pass; full tests are completed-job
results, no redundant long checks. Source and this evidence staged, uncommitted.

Next capacity gates: one-arena streamed layout,34-bit extension/coordinates and
compatible artifacts/preprocessing plus host memory/file backing. Current fast
s17 two-array model256 GiB cannot fit this GPU;128.25 GiB one-arena projection is
not implemented or admitted. Keep further material time/RAM/disk trade-offs for
explicit user choice. No s17/application/cryptographic/TRL5 gate closes.
