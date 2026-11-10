# Bounded s16 resident root-product intervals

Status: job **72842 passes**, including full saved-support census and exact
partial-product equality across all three variants. Follows the validated
[resident integration](../gpu-products/report.md) toward an s16 estimate without
running full preprocessing or repeating discovery. Starting revision `715bac2`
plus exact recorded current sources. [run.py](run.py), [run.sbatch](run.sbatch),
[test harness](../../encoding/liftrs-shared/gpu_s16_intervals_test.go),
[analyzer](analyze.py) and [operation inventory](inventory.py) define the work.

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/s16-resident-intervals/slurm-%j.log \
 experiments/s16-resident-intervals/run.sbatch intervals-01
python3 -B experiments/s16-resident-intervals/analyze.py intervals-01
python3 -B experiments/s16-resident-intervals/inventory.py intervals-01
```

One GPU/16 allocated CPU cores, four native workers, 16 GiB requested host RAM,
12 GiB sampled process-group RSS guard, 512 MiB explicit device bound, 30-minute
allocation/25-minute controller deadline. Maximum transform2^22, prefix2^20
saved support entries in two consecutive2^19 intervals. Driver/context memory
is separate. Larger transforms and the full s16 executor are not allocated.

Read the original3.42 GiB support file read-only, verifying its frozen checksum,
specification and sorted order. Count every accepted Frobenius factor and each
32768-factor batch's degree, retaining fixed-factor ordinals for exact scheduling.
Their total degree must equal K=918,672,421. This is a census of verified saved
discovery, not a new independent discovery/closure proof; no whole support list
or membership bitset is allocated. Only the first1,048,576 entries are retained.
Saved code/source provenance is reviewed for support-only reuse: coordinates,
field polynomial, curve degree and root formula are unchanged; no full cache
fingerprint is overwritten or bypassed.

CPU cutoff32768 and resident GPU cutoffs1024/4096 replay the same root formula
and factor stream. Each backend keeps one product builder across both intervals,
so actual early binary carries are included. Finalizing the partial product is
timed separately and is not a full code polynomial. Arrays/files must match
exactly across variants. Reusable native/GPU plans/workspace are set up separately;
Go profiles, interval clocks and explicit device ownership are retained. Intervals
include root generation, small products, carries, allocation and diagnostic
profiling; no unprofiled or randomized latency claim. Backend order is fixed.

The shared library is checksum-verified from products-02. Matching Go binaries,
partial product fixtures and hashes are outside Git at
`.cache/liftrs/s16-resident-intervals/<run>/`; retain them for later checks.
No full P/G/online cache is created or modified. Raw data/analysis and exact
source snapshots are retained with hardware, compiler, driver and allocation logs.

These early intervals cannot establish late carries, Newton reciprocal, identity
or spectrum costs. The census allows an operation-size inventory to expose that
gap. Current CUDA shape counts/metadata/workspace admission stop below the full
2^32 domain: extrapolation is conditional on extending and validating those
boundaries. Do not report a whole-run speedup or deadline prediction by scaling
these early intervals or the s10 ratio. Larger-kernel measurements need a separate
recorded workload policy before execution; no full s16 job is launched here.

## Results and estimate boundary

[Raw record](runs/intervals-01/results.json), [analysis](runs/intervals-01/analysis.json),
[interval clocks](runs/intervals-01/stages.json), [census](runs/intervals-01/census.json)
and [logical operation inventory](runs/intervals-01/inventory.json) are retained.
The analyzer verifies273 sources and all library/binary/cache/evidence hashes.
Job ran on gnode03, Xeon6710E, H200 NVL, driver580.126.20/CUDA12.9.86,
GCC14.2.0/Go1.27.3. The support file checksum matches the frozen discovery record.
Census found459,356,469 accepted factors,40,517 fixed factors, and14,019 batches;
their degrees sum to K=918,672,421. Scanning/hash/census took38.231 s, separately
from root-product work. Only the first2^20 support entries were kept in memory.

| Backend / product cutoff | First2^19 positions, s | Next2^19 positions, s | Sum, s |
|---|---:|---:|---:|
| Current CPU /32768 | 16.699 | 19.992 | 36.691 |
| Resident GPU /1024 | 9.679 | 10.526 | 20.205 |
| Resident GPU /4096 | 8.252 | 9.015 | 17.267 |

Cutoff4096 gives2.12× early-interval speedup against the current CPU path using
four native workers. This differs from the s10 cutoff result: use measured
field/shape-specific choices, not one global threshold. Backend order is fixed;
one prefix/two intervals are observations, not randomized whole-code samples.
Each variant's final partial product has2,096,712 coefficients and identical hash.
Finalization outside interval clocks costs5.980/0.849/0.816 s respectively. This
partial finalization is artificial and must not be charged as complete-code
finalization or included silently in the interval ratio.

Resident product calls are1,994 at cutoff1024 and503 at4096; explicit owned device
state is51,240,040 bytes before Close, excluding driver/context and temporary
growth. Command RSS peak248.58 MiB, retained partial artifacts/binary30.99 MiB.
The native transpose and GPU buffers are bounded independently of full N.

The inventory derives exact carry/final-merge lengths from batch degrees and
Newton/identity lengths from current code, without allocating those arrays.
Root carries reach2^29 points and final root merging2^30; reciprocal/identity
need convolutions at2^31 and2^32, followed by a2^32-point spectrum transform.
Many of these are outside the2^22 measurement/admission cap. Internal leaf trees,
CPU root/arithmetic, memory/GC and I/O are explicitly excluded from this operation
inventory; logical counts alone are not a timing model.

Therefore there is **no defensible whole-s16 time/speedup estimate yet**. Needed
next: extend/validate the private GPU domain metadata (a2^32 count and17 row-basis
levels), or choose a validated bounded-block design; record a suitable workspace/
kernel budget, then measure representative
later sizes and validate extrapolation at more than one scale. The current bridge
already uses size_t for call lengths; the remaining width issue is its private
plan description/admission, not a need to widen every field value or public file.
These are software/admission boundaries, not measured proof that the available
GPU cannot fit the work. No full s16 job is admitted by these early results.
