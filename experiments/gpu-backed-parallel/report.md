# Parallel arithmetic within host-backed GPU tiles

GPU-040 addresses the measured small-tile computation bottleneck from
[GPU-039](../gpu-host-backed/report.md). At s16/N2^22/T65536, its one-vector-per-
thread implementation reduced device ownership32.76→1.26MiB but slowed one
forward transform0.128→7.73s. Rows/columns accounted for6.50s. This experiment
parallelizes arithmetic across vector elements within the **same two device
tiles**. Resident kernels and defaults stay unchanged; full online GPU integration,
then CPU output/disk-backed encoding, remain subsequent gates. No s17 claim.

## Reproduction and provenance

From repository root:

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-backed-parallel/slurm-%j.log experiments/gpu-backed-parallel/run.sbatch parallel-01
# Only after the user reports completion:
python3 -B experiments/gpu-backed-parallel/analyze.py parallel-01
```

[Controller](run.py), [Slurm script](run.sbatch), [analyzer](analyze.py),
[stage/layout self-check](check.py), and [unchanged GPU reference test](../../encoding/liftrs-shared/gpu_backed_test.go).
The controller is the GPU-039 method plus a sealed component checkpoint and
fixed-input comparison. It records exact HEAD and source hashes/snapshots,
commands, hardware, caches, profiles and timings. HEAD is ce29557 plus recorded
working sources. Only production `backed.hpp` may differ from job74121.
[Compile/self-check evidence](compile-check.json) is local compilation only;
runtime validation **passed job74125**, submitted2026-11-12;374 source identities
and all cache/evidence digests reconciled. See [analysis](runs/parallel-01/analysis.json).
[Results](runs/parallel-01/) and [Slurm log](slurm-74125.log) are populated by
the job; no monitoring until the user completion ping. Results will be under `runs/parallel-01/`,
retained libraries/binaries/fixtures under
`.cache/liftrs/gpu-backed-parallel/parallel-01/`; no prior cache is overwritten.

Same H200/GCC14.2/CUDA12.9/Go1.27.3 environment; one GPU,16CPUs,16GiB host
request,12GiB sampled process RSS guard,8GiB Go soft limit,512MiB device cap,
16GiB new cache/scratch guards each. Controller20min, worker15min, Slurm30min.
Initial host availability32GiB/runtime headroom16GiB and1GiB free GPU check.
These budgets are unchanged. Device ownership excludes CUDA context/driver.
Job-internal guards remain active; the agent does not monitor the submitted job.

## Change and correctness

[backed.hpp](../../encoding/liftrs-cuda/backed.hpp) assigns a butterfly pair to
each thread instead of assigning a whole vector. For each forward stage, a joined
kernel first corrects quotient entries using its unchanged high half. A second
kernel reads that joined quotient and writes remainder/butterfly results to the
other bounded tile. Forward stages alternate these tiles. Inverse stages write
the joined quotient into scratch, then restore coefficients into the original
tile. All inverse correction reads use the immutable joined quotient, preventing
races with its correction writes. Linearized lower polynomial degrees<=half/2
are required and already validated in the matrix description. No floating-point
field arithmetic or cryptographic behavior changes.

Columns retain their interleaved staging layout; rows use their contiguous
host order directly and retain their full-domain row origins. There is no
explicit row transpose in this path; layout timing becomes0. The controller
requires identical input/output fixture hashes, dimensions, T, exact device
ownership, host staging, H2D/D2H bytes and batch counts relative to74121.
All eight Go/resident/forward/inverse cases repeat, six under memcheck, plus80
existing resident regressions, API/build/race/vet. The independent CPU stage
check compares the parallel sparse formulas to sequential division/butterfly
arithmetic over GF(256) and checks pair indexing covers every element exactly.
It is a complementary check; GPU/reference comparison validates actual fields.

## Measurement boundaries and trade-offs

Timings seconds; allocation and traffic bytes. Existing12-element native stats
retain their meanings (see [GPU-039 method](../gpu-host-backed/report.md)).
Compare fixed-input forward latency and sum of row/column GPU time to retained
74121 measurements; no new repeat or warmup. These are profiled, fixed-order
component samples, not a full encoding speedup or robust latency baseline.

Workspace remains8min(N,T) device bytes plus tables/plans; mutable host4N,
staging<=4T, and test/reference arrays extra. No memory/time exchange is intended:
this uses the same two tiles with more parallel work, adding kernel boundaries
per stage and removing row transposes. The test requires unchanged tracked
memory/transfer counts; any runtime improvement remains unmeasured until the
job completes. Repeated Taylor transfers and CPU column routing remain, so a
host-backed route can still be much slower than resident transforms. No
production disk workspace or larger CPU code cache is introduced. Fixture disk
cost remains two4N files per case. Full encoder setup, message/code mappings,
convolution and output lifetimes/latency require later integration and admission.


## Measured result

Job74125 completed0:0 in2m32s on gnode17/H200. All8 reference/resident/inverse
cases,6 memchecks (zero errors/leaks),80 regressions, API/build/race/vet pass.
Exact input/output hashes, device ownership, staging and transfer counts match
74121. Peak controller-sampled RSS292618240bytes includes compilation; Slurm
MaxRSS233568KiB is a different sampling boundary. Profiles/fixtures retained.

| s16 component | Old backed seconds | Parallel backed seconds | Resident seconds |
|---|---:|---:|---:|
| N=2^20,T65536 |1.242238|0.556460|0.052801 (prior job)|
| N=2^22,T65536 |7.734124|1.338257|0.127805 (prior job)|

At2^22, forward improves5.78x (82.7% shorter), row/column GPU time6.496→0.03943s.
Device ownership1322628bytes (1.261MiB), mutable host array16MiB/staging256KiB,
H2D1186942976/D2H171921408bytes unchanged. Taylor0.417s, H2D0.299s,D2H0.0685s,
host routing0.0086s, plus launch/event overhead dominate. The full transform
still costs10.47x the prior resident sample, so this remains a slower memory
route. Tiny s4 cases regress20–43% because more kernel boundaries outweigh
parallelism there. These are fixed-order/profiled component measurements.
Full online integration and its host/code/output lifetimes remain unmeasured;
no automatic default change or s17 feasibility follows.
