# Compute-node cached s15 baseline passes

- Timestamp: 2026-11-09 16:42:00 UTC
- Agent: Codex
- Work ID: HPC-001
- Status: completed
- Code/results revision: `0f26303`, with exact experiment source hashes/snapshots

Completed the authorized shared Go setup, compute-node validation and cached s15
baseline. [Report](../experiments/hpc-cached-s15/report.md),
[runner](../experiments/hpc-cached-s15/run.py),
[batch script](../experiments/hpc-cached-s15/run.sbatch),
[results](../experiments/hpc-cached-s15/runs/baseline-02/results.json) and
[analysis](../experiments/hpc-cached-s15/runs/baseline-02/analysis.json) preserve
commands, parameters, environment, measured phases and validation evidence.

Verified Go 1.27.3 is now on shared NFS storage at
`.cache/liftrs/toolchains/go1.27.3-linux-amd64/go`, with pinned modules in
`.cache/liftrs/toolchains/gopath`. Jobs run offline against those dependencies.
No shell startup or system package changes were made. Existing source-machine
and transferred preprocessing/outputs remain intact.

Slurm job 72676 completed on gnode02 (Xeon 6710E), with 16 allocated cores, a
64 GiB request and one-hour wall limit, no GPU request. Actual affinity is
40–47,104–111. Cluster cgroups did not enforce the requested memory cap. The
first admission attempt, job 72675, was cancelled during tests before encoding
to add a one-second process-group RSS guard with a 64 GiB threshold and command
deadlines. Both attempts and final [accounting](../experiments/hpc-cached-s15/accounting.log)
remain recorded; the partial attempt is not an encoding result. A sampled guard
is not a hard RSS cap and may count shared pages more than once.

Both full Go/native suites, focused race checks, vet and the isolated legacy
fixture build pass. The existing optional network test may skip when its server
is absent. The unchanged online-only s15 probe runs from an isolated source copy
with staged constants and new scratch output. Exact K=216,914,047, N=2^30,
m=2,t=5 and seed 71 match the laptop fixture. All systematic symbols, 16
independent curves, complete canonical/packed digests and 1,027 packed reads pass.

Measured profiled phases: staging/checksums 52.45 s; online loading 5.872 s;
native/workspace setup 0.016 s; encoding 655.937 s; validation/digest 37.116 s;
packing/hash/fsync 12.622 s; packed readback 0.00127 s. Probe peak RSS is
18.465 GiB. Slurm's 14m48s job elapsed also includes source copying, builds,
correctness suites and retention. This is one shared-node portability baseline,
not a matched algorithm speedup comparison or an unprofiled latency distribution.

All six CPU profiles and phase clocks remain in the result directory. Perf is
denied and no GPU is visible in this CPU-only allocation. Go pprof can read the
encoding profile but mostly attributes samples to unsymbolized native code and
runtime.cgocall; detailed native attribution remains unavailable. Native stage
counters sum to 655.712 s within the 655.937 s enclosing native clock.

Retained matching binary and new 1.875 GiB packed output live outside Git under
`.cache/liftrs/hpc-cached-s15/baseline-02/`; job-local scratch remains at
`/tmp/sk-depir-hpc-72676` on gnode02. Original reference files were read only.
The [analyzer](../experiments/hpc-cached-s15/analyze.py) verifies 227 source files,
all sealed evidence, binary and retained output hashes. Exact new runner/test
sources are retained in `experiments/source-snapshots/` by content hash.

Added [s16 readiness audit](../docs/encoding-s16.md): the argument is bounded
because N=2^32 cannot fit the current uint32 cardinality/header fields, while
individual extension values still fit uint32. It identifies Go count/products,
transform-level calculations, native descriptors/shifts and online-file versioning
separately from value storage. The 74.3 GiB online-array scenario is analytical;
preprocessing lifetimes and full RSS still need admission review. No arithmetic,
guards, public encoding APIs or file interpretation changed, and no s16 run began.

Changed paths: the new `experiments/hpc-cached-s15/` scripts/report/results/profiles,
three measured source snapshots, `docs/encoding-s16.md`, `docs/hpc-setup.md`,
`docs/development-handoff.md`, `docs/encoding-scaling.md`, root and experiment
READMEs, this message and board index. README check complete: current CPU
validation is distinguished from unverified GPU/s16 support and TRL 5 acceptance.
[Final verification](../experiments/hpc-cached-s15/final-verification.json): two
runner unit checks, Python/shell syntax, 183 documentation links and source/evidence
analysis pass; `git diff --check` passes. Full suites were not repeated after
report-only edits. Next implementation gate is the documented s16 count/header
extension and small-buffer/boundary tests, followed by preparation-memory/budget
review before admitting a larger full run.
