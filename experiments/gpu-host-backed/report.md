# GPU host-backed transforms

GPU-039 is gate 2 of the [remaining memory work](../../docs/encoding-scaling.md#remaining-outputbacked-transform-gates-gpu-then-cpu).
Gate 1, [GPU packed output](../gpu-packed-output/report.md), passed job74094.
This experiment implements and validates the transform component before online
encoding integration. The complete mutable uint32 domain stays in host RAM;
GPU arithmetic uses two bounded tiles. This is an explicit alternative to
resident transforms, which remain the default. CPU output/backed work follows
the completed GPU route. No full host-backed encoder or s17 feasibility claim.

## Reproduction and evidence

Run [controller](run.py) using [Slurm script](run.sbatch) from repository root:

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-host-backed/slurm-%j.log experiments/gpu-host-backed/run.sbatch backed-01
# After the completion ping:
python3 -B experiments/gpu-host-backed/analyze.py backed-01
```

[Analyzer](analyze.py) verifies source snapshots, commands, profiles, fixture
hashes and assertions; [self-check](check.py) covers admission/traffic records
and the quotient's source dependency. Run `python3 -B experiments/gpu-host-backed/check.py`.
[Login-node compile check](compile-check.json) passed native sm90 compilation,
CUDA-only and combined Go API checks (GPU test skipped), controller self-check
and shell syntax. Job **74121**, run `backed-01`, passed on gnode17/H200 in3m59s.
[Analysis](runs/backed-01/analysis.json), [results](runs/backed-01/results.json)
and [Slurm log](slurm-74121.log) retain exact evidence;370 source identities
and all cache/evidence hashes reconciled after the completion ping. Each run records exact HEAD and changed-source
SHA256 in `runs/NAME/results.json`, with durable content in
`experiments/source-snapshots/`; caches remain outside Git under
`.cache/liftrs/gpu-host-backed/NAME/` (library, test binary, input/reference arrays).
The first run is based on ce29557 plus the recorded working-source snapshots.

One H200, 16 CPUs, GCC14.2/CUDA12.9/Go1.27.3; Slurm16 GiB RAM,
12 GiB sampled process RSS, Go8 GiB soft limit, GPU512 MiB allocation cap,
16 GiB new cache and scratch guards each. Controller20 min, worker15 min,
Slurm30 min deadlines. Initial host availability32 GiB, runtime headroom16 GiB.
These limits are below the previously admitted full s16 experiment budgets.
Worker guards run inside the submitted job; the agent does not monitor it.
GPU ownership excludes runtime/context allocations; one GiB free device memory
is required at startup. No full-domain device array is allocated by this route.

## Method

[CUDA host-backed implementation](../../encoding/liftrs-cuda/backed.hpp) reuses
Taylor coefficients and row/column kernels from the resident matrix transform.
Forward: Taylor corrections, columns, rows; inverse reverses that sequence.
Quotient updates read its unchanged high half; remainder updates read the joined
quotient in the disjoint upper half. Transfers and kernels complete before
reusing tiles. Row batches retain their global affine origin. Column staging
copies values only: all field arithmetic remains on GPU. No host pointer is
retained across calls. The mutable host argument may be partially transformed
on CUDA failure and is then unsuitable as an input checkpoint.

[Test](../../encoding/liftrs-shared/gpu_backed_test.go) compares every value
against Go additive-subspace transforms and resident CUDA, then checks inverse,
error recovery, closed-engine behavior, domain/range admission, stable exact
allocation, and byte counters. Eight cases: `(s,dimension,tileElements)`
`(4,7,48),(4,8,48),(6,11,320),(6,12,320),(8,15,1536),(8,16,1536),`
`(16,20,65536),(16,22,65536)`. Non-power-of-two tiles force batch tails;
odd dimensions exercise rectangular matrices. Six small cases additionally run
under compute-sanitizer memcheck/leak checking. Existing 80 matrix regressions,
CUDA-only build, API checks, race checks and vet guard the resident path.
Full s16 dimension32 is plan/allocation admission only (no full data/run);
too-short tile32768 must reject its65536-element vectors.

Timings are seconds, byte counters bytes. Native stats indices: H2D, Taylor,
columns, rows, GPU transposes, D2H, CPU gather/scatter, native total,
H2D bytes, D2H bytes, host staging bytes, row+column batch count.
Go reference, resident and backed forward run in that fixed order; no warmup
or robust speedup claim. CPU profiles are retained for full/sanitized checks;
profiled component results do not estimate unprofiled full database encoding.

## Memory and runtime trade-off

For transform length N and tile T, device work arrays shrink from8N bytes to
8min(N,T), plus unchanged field tables and plan. Column staging adds at most4T
host bytes; the mutable host array still requires4N. Reference/test arrays are
extra and excluded from production workspace accounting, but included in RSS.
The test asserts this exact device allocation difference and stable ownership.
Each transform now transfers at least8N bytes in each direction (columns and
rows); Taylor corrections add transfers and launches per coefficient/stage.
There is no asynchronous transfer overlap or pinned staging. The runtime
penalty is unknown until measurements; this is not a free memory saving.
Fixtures require two4N files per case, not a production disk workspace. Library
setup, code caches and full encoding remain separate unmeasured integration
costs. No production disk I/O occurs inside this transform component.

For full s16 N=2^32 with T65536, principal device work is512 KiB instead of
32 GiB, but the mutable host array is16 GiB. These are allocation formulas,
not full-run measurements, and exclude tables/runtime/reference arrays. Smaller
VRAM alone does not solve host memory or the s17 field representation changes.
Full online integration must specify message, code mappings/cache, convolution
and output lifetimes and measure their costs before admission.

## Measured component result: GPU-039

Eight forward/Go/resident/inverse cases passed, six cases under memcheck with
zero errors/leaks;80 prior matrix cases, API/build/race/vet checks passed. Full
s16 metadata fits its bound; no full s16 transform executed. Peak controller-
sampled process-group RSS291241984bytes includes compilation. Slurm MaxRSS
238824KiB is a different sample/boundary. Cache and profiles retained.

| s16 component | Resident | Host-backed, T=65536 |
|---|---:|---:|
| N=2^20, forward seconds |0.052801|1.242238|
| N=2^22, forward seconds |0.127805|7.734124|
| N=2^22, owned device bytes |34352772|1322628|

At N=2^22 this is60.5x slower for96.2% less device memory (exact work-array
saving31.5MiB; field tables/plan unchanged). Mutable host array16MiB and staging
256KiB remain, plus test/reference arrays. Forward/inverse transfer counts are
1186942976 H2D and171921408 D2H bytes per transform. Resident transfer is16MiB
each direction. These are component measurements at fixed small T, not a full
encoding runtime projection. Backed forward time includes3.147s columns,
3.348s rows,0.400s Taylor GPU,0.290s H2D,0.066s D2H,0.010s host routing;
remaining time includes synchronous launch/event/driver overhead. No production
transform disk workspace; retained fixtures are input/reference evidence only.

The initial row/column kernel assigns a complete vector to one thread. At
N=2^22/T65536 there are only32 vectors per batch, severely limiting parallel
work. GPU-040 replaces these bounded-batch kernels with per-stage arithmetic
parallel across vector elements, preserving two device tiles and host staging.
This is a runtime optimization within the same memory bound, before online
integration; Taylor transfer costs remain. Resident defaults/kernels stay intact.
