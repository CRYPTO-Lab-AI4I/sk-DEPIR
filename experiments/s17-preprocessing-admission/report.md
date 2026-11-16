# Exact dimension and full s17 preprocessing admission

The persistent GPU workflow passes small complete cases (job75342). This step
obtains exact K for full resource sizing, without sampling support or constructing
P/G/spectrum. It is real GPU monomial enumeration, not another transform benchmark
or full s17 preprocessing run. Count-only data does not certify sampled support.

## Reproduction and provenance

[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py) reuse the
[complete-workflow controller](../gpu-preprocessing-workflow/run.py) and its
[verification](../gpu-preprocessing-workflow/analyze.py), including short peer
preflight and bounded shutdown. Their optional experiment-directory/mode arguments keep
admission runs separate without duplicating orchestration; the admission entry
selects/requires exact dimension mode. Original workflow
scripts/defaults and result interpretation remain compatible. All copied source
hashes, binary/plugin identities, revision, environment and exact commands are
sealed in `runs/<name>/results.json`.

```sh
module load slurm/slurm/25.05 gcc/14.2.0 cuda12.9/toolkit/12.9.1
sbatch --exclude=gnode03 --output=experiments/s17-preprocessing-admission/slurm-%j.log \
  experiments/s17-preprocessing-admission/run.sbatch counts-01 --dimension-s17
# Only after user completion ping:
python3 -B experiments/s17-preprocessing-admission/analyze.py counts-01
python3 -B experiments/s17-preprocessing-admission/sizes.py \
  .cache/liftrs/s17-preprocessing-admission/counts-01/dimension-s17/dimension/manifest.json
```

Count manifests/profiles/binaries persist outside Git in
`.cache/liftrs/s17-preprocessing-admission/<name>/`; command logs, progress and
small results remain in the experiment. [Local checks](local-checks.json),
[native test log](local-native-tests.log) and [size calculator](sizes.py) make
pre-submission checks and modeled sizes reproducible. Runtime results are pending;
no source-only capacity or API capability is described as a successful full run.

## One count implementation, separate host admission

The native and CUDA `GoodMonomialCount` methods reuse their existing C count ABIs
and exact ternary-class/bitmap algorithms. `GoodMonomials` delegates to the same
locked helper before filling indices; no duplicated arithmetic, ABI or default
changes. Executor membership buffers are released on reuse/Close. New
`CountCodeMonomials` constructs a private count-only session; only unused N-element
host-grid admission is bypassed. The private session never escapes to callers;
normal `NewCodePreprocessor` retains its full grid/array checks. Count still
requires explicit extent, metadata/native and per-GPU byte caps. GPU count
requires resident devices and never falls back to CPU. CPU count remains an
independent reference path. Normal client builds remain CUDA-free.

`pir-preprocess --dimension-only` publishes only a trusted `dimension/manifest.json`
with Spec,K,N, producer/source/executable/plugin identity. Unique profiles and
phase progress remain active. It creates no support/root/code payloads or full
host index list. `--resume` can reuse that local dimension record, but full
preprocessing still enumerates the monomials and certifies sampled support;
count metadata does not skip mathematical certification. The temporary device
bitmap is not saved: copying/persisting2 GiB would add I/O and storage to a
bounded sizing step. Its regeneration is accounted as discovery work in a later
full run. Count-only API is not a low-memory full preprocessing mode.

## Bounded checks and resources

Keep the approved4H200,4 GiB aggregate owned cap (1 GiB/card),4CPU,8 GiB requested
host/6 GiB sampled process-group RSS/4 GiB Go soft limit,2 GiB new cache/4 GiB
scratch,600/900/1200-second worker/controller/Slurm deadlines. Use eligible nodes
without pinning; gnode03 remains excluded for the unresolved reset fault.
Preflight, controller tests and the full small executable workflow suite run
before large counts. Local native tests pass independent-Go s4 count, known
s10 K146641, and unchanged rejection of ordinary s10 preprocessing under1 MiB
host-array admission. Publication/restart/corruption tests remain active.

The scheduled suite includes CUDA-free/native/race/ASan/vet, small full1/4GPU
peer/local workflows, independent-Go canonical bytes and fresh workflow CUDA
memcheck. A separate fresh count-only s4 four-GPU memcheck tests destruction
without host fill. Then count-only CLI processes enumerate s8/s16/s17. s8 must
match the verified Go-derived support dimension; s16 must match918672421. All
three must contain no support/root/code directories. These comparisons and full
s17 enumeration are pending the submitted run, not local measurements.

At s17, the sharded device bitmap is2 GiB aggregate; four replicated class tables
are about493 MiB total, plus tiny ternary/field/plan buffers. Conservative admission
checks keep each card below1 GiB; expected aggregate principal count payload is
about2.5 GiB, not a measured peak. No~29 GiB host index list or128 GiB sampling
grid is allocated. All count kernels/math are unchanged; existing native phase
start/completion and controller resource heartbeats are retained. Whole command
latency includes setup, metadata publication and profiling/logging; any GPU count
latency must not be used as full discovery/preparation time. No full support/cache
or new database encoding is produced.

## Full-resource sizing draft, not approval

[sizes.py](sizes.py) computes representation sizes from the exact count record:
N=2^34, support8K bytes, root4(K+1), canonical code53+4(K+1)+8N, generator4(N−K).
Total retained successful payload is12K+12N+61 bytes, strictly ≤288 GiB+61 bytes
for the current admitted shape0<K≤N/2. These are format-derived sizes, not measured
storage/peak RAM. Failed partial stages, metadata/logs/profiles and headroom are
additional. Fresh persistence writes/hashes and reads back each payload once.

Discovery principal live-array model is16K+8N+N/8 bytes (two index lists, grid,
bitmap); final retained arrays are8K+12N+4 bytes (support,P,G,spectrum, with the
saved root sharing P). Root/Newton products, old GC-retained buffers, native/
runtime/staging and filesystem cache are additional. These formulas do not bound
full process RSS. Existing~600 GiB host scenarios remain unmeasured.

[Full budget draft](full-budget-draft.json) proposes one exclusive128-core node
with32 active workers,864 GiB requested host RAM/800 GiB RSS guard/720 GiB Go soft
limit,4H200 with112 GiB owned cap each,512 GiB retained storage/16 GiB scratch and
48-hour scheduler cutoff. This is **not agreed, submitted, or measured**. The
exclusive node costs all128 allocated cores to avoid shared unaccounted RAM;
worker count is not its billing count. GC cost if the soft limit is reached is
unknown. Explicit local layout is the measured four-card candidate, with its
extra transposes/transfers retained in the admission decision. Eight/sixteen GPUs
remain authorized options to consider, but require unimplemented cross-node work.

Exact K, remaining heap-lifetime audit, guard configuration, free-space/quota
checks and explicit user agreement to numeric host/storage/time/CPU budgets are
still needed before full execution. Filesystem free space alone is not quota or
job admission. No full budget increase is authorized by the bounded count run.
