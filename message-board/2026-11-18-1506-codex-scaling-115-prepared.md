# Failed snapshot preserved; build-only retry validated

- Timestamp: 2026-11-18T15:06:30.707123+00:00
- Agent: Codex
- Work: SCALING-115
- Status: Prepared

77241 FAILED1:0,5s,HPCgnode04/05. Original result says File too large,zero commands,
no GPU/MPI execution. Source copy selected unrelated37479757-byte historical JSON,
exceeding32MiB per-file guard. Source freeze incomplete, so no complete source seal
or compute metrics claimed. [Failure review/accounting/raw data](../experiments/gpu-staged-transport/report.md#first-run-source-snapshot-failure-and-corrected-retry-scaling-115)
preserve original hashes/log/results/scratch; no cache deletion or cap increase.

Corrected selector admits code/header/assembly/tool sources,Go module files and
core JSON assets,not experiment datasets/artifacts. Git object metadata preflights
32MiB/file and128MiB total source before reading/writing contents.4 unit tests
pass: historical-sized artifact excluded while core config retained,oversized
source/total rejected before content reads,rank/cap/timing/payload checks.
Actual corrected selection against26bb006 Git inputs has613files/2908958bytes;
shutdown helper import,Go fixture compile/generation and profiled NVCC/MPI-cudart
link **from that snapshot** pass. [Readiness record](../experiments/gpu-staged-transport/snapshot-check.json)
binds selector SHA256,original input revision and retained local source path.
AST,bash syntax,runner --check and git diff --check pass. Kernels/measurement
boundaries/resources unchanged. GPU runtime still unrun,not a hardware diagnosis.

README checked/updated, as are actual main reference,active plan,CUDA/experiment
index and report:failed-before-compute distinct from pending staged-02 retry.
Budget unchanged,explicitly approved2nodes/2H200/4CPU/8GiBhost/15min,128MiBowned/
card,64MiBpinned/2GiBRSS/rank,128MiBretained/2GiBscratch. Commit fix/evidence,then
submit fresh staged-02 and hand off without monitoring; no new approval needed.
