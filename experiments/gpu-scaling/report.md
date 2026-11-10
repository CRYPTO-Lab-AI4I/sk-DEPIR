# GPU domain metadata and bounded scaling

The [runner](run.py), [Slurm launcher](run.sbatch) and [analyzer](analyze.py)
validate full-domain metadata without full-domain arrays, then compare progressively
larger transforms/products. Parent checkpoint: `a997c4e`; each run records revision,
exact dirty source hashes/snapshots, compiler/library/test hashes and commands.
[scaling-01 results](runs/scaling-01/results.json) and
[verified analysis](runs/scaling-01/analysis.json), job72844, pass all checks.
[Measurements](runs/scaling-01/measurements.json),
[metadata checks](runs/scaling-01/metadata.json),
[CUDA memcheck](runs/scaling-01/memcheck.log) and retained per-size profiles
provide the raw evidence.

Run from the repository root:

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-scaling/slurm-%j.log experiments/gpu-scaling/run.sbatch scaling-01
python3 -B experiments/gpu-scaling/analyze.py scaling-01
```

Use a fresh run name. Requires the pinned Go1.27.3 toolchain/dependencies,
GCC14.2.0, CUDA12.9, one available NVIDIA GPU and compute-sanitizer. Ordinary
library builds retain the CPU implementation and have no CUDA dependency.
The runner records the actual node/CPU/GPU, OS, driver, compiler and toolkit.
Its shape cap is not a measured hardware limit.

## Method and bounds

[Tests](../../encoding/liftrs-shared/gpu_scaling_test.go) use s=16,m=2 field
coordinates, polynomial0x1002d and deterministic seeds1900+dimension. Private
version2 descriptions store count low/high words; fields/coefficient arrays remain
uint32. A 2^32-domain description has17 row levels, including the extra polynomial
needed by Taylor conversion. CheckPlan parses it without device plan/workspace
allocation. AddPlan must reject it under the job cap. Truncated/trailing/oversized
counts and an underfunded engine are rejected. Legacy descriptions and loading an
old transform-only binary remain covered by the100 resident product regression
checks. Small metadata checks run under CUDA memcheck.

Actual execution covers2^22,2^23,2^24,2^25 values, in that order on one engine.
Each size has exact full-array CPU forward/inverse comparisons, two transfer-
inclusive GPU forward/inverse samples and two unequal-operand resident products
compared with a complete CPU reference. CPU uses16-thread native transforms and
canonical Go field multiplication for the product's pointwise loop; that complete
product is not a matched native-only throughput comparison. Setup clocks are
separate. CPU has one timing sample per operation; GPU has two in fixed order.
Checks/hashing are outside operation clocks. CPU profiles cover each whole size,
including checks, allocation and both backends. GPU event stage clocks plus wall
transfer/total clocks are retained; native instruction profiling is unavailable on
this host, so Go profiles do not identify native/CUDA instruction hotspots.

Budget per job: one GPU/16 CPUs,16 GiB host request,12 GiB sampled process-group
RSS guard,512 MiB explicit device allocations,30-minute allocation/25-minute
controller deadline. Go soft target4 GiB. CUDA driver/context memory is excluded
from OwnedBytes; the inventory records total device capacity, not its overhead. No full s16 run.
The guard samples process RSS and is not a hard cgroup memory limit.

## Memory schedule and alternatives

Three resident arrays require12*n bytes:384 MiB at2^25, plus field tables and
small plans. Synchronous single-caller ownership means their contents are dead
between calls. Growth frees all old arrays before allocating replacements;
tables/plans remain resident. Admission checks payload plus retained tables/plans
before growth, and checks the next plan too. Allocation failure cannot leave a
false workspace capacity; each later operation rechecks/grows workspace.

At2^32 the three arrays alone require48 GiB, before tables/plans/context and host
inputs/outputs/CPU references. This is an arithmetic model, not measured RSS or
GPU feasibility. It exceeds this job's device/host bounds even though the locally
available H200 has more device memory. A larger job needs explicit device, host,
cache and time accounting; metadata support does not authorize that allocation.

A bounded-block alternative retains smaller domains for polynomial products and
combines coefficient blocks with overlap-add, increasing products and staging.
It does not by itself implement a full-domain spectrum transform: that still
needs a validated global row/column schedule with its transfers/I/O. Compare this
extra work with contiguous allocation on available hardware before choosing; no
blocked full-domain GPU implementation or performance is claimed here.

## Retention and limitations

The runner retains profiles/results and source snapshots in Git, and binaries,
verified CUDA library and reusable small plan descriptions under
`.cache/liftrs/gpu-scaling/RUN_NAME`. Hashes are in results.json. Large generated
random arrays and complete CPU expected arrays are not cached: regeneration from
recorded seeds is cheap and avoids redundant GiB-scale fixtures; complete output
hashes are retained. Existing support/constants/profiles are untouched. No code
discovery or s16 constants are regenerated. New cache size and process RSS are
recorded by analysis.

Execution through2^25 would still leave seven doublings to2^32. Do not apply the
early root-interval ratio or small-code ratio to all preparation. A whole-run
estimate also needs later-stage scaling, CPU roots, allocation, reusable setup,
cache loading/save and I/O. TRL5 acceptance, cryptographic/application integration
and full s16 preprocessing remain open.


## Measured checkpoint

Job72844 ran on gnode03: Xeon6710E, H200 NVL (143,771 MiB reported device
capacity), driver580.126.20, CUDA12.9.86, GCC14.2.0 and Go1.27.3. This is a
data-center GPU measurement; no consumer-GPU performance is inferred.
GPU times below are medians of two profiled transfer-inclusive samples; CPU
forward time is a single16-thread native sample. Seconds:

| Domain | CPU forward | GPU forward | GPU inverse | GPU resident product |
|---|---:|---:|---:|---:|
| 2^22 | 0.769 | 0.140 | 0.131 | 0.384 |
| 2^23 | 1.595 | 0.237 | 0.224 | 0.644 |
| 2^24 | 3.382 | 0.370 | 0.345 | 1.010 |
| 2^25 | 7.090 | 0.630 | 0.590 | 1.693 |

At2^25, measured forward speedup is11.25x over this CPU sample. Product CPU
reference takes24.372 s, but includes a Go pointwise loop and is identified
separately above. All full arrays match; this is not a whole-preprocessing ratio.
The GPU forward time grows4.52x for8x more values across this range. That measured
trend is not extrapolated across the remaining seven doublings.

Engine setup takes0.524 s separately. The2^32 count and17-row-level description
passes without changing device ownership (786,424 bytes of tables before/after).
Its allocation is rejected by the execution budget. Malformed descriptions and
an underfunded engine are rejected. CUDA memcheck reports zero errors for these
metadata checks; it does not cover the large execution arrays. The100 small
resident product regressions, including legacy-library loading, also pass.

Peak explicit device ownership is403,487,272 bytes (384.8 MiB), including four
plans, safely below512 MiB after growth. Sampled command process RSS peaks at
818.7 MiB; retained private cache is8.15 MiB. The analyzer verifies279 exact
source files plus all sealed logs, profiles, cached plans and binaries.

Next gate: use this scaling evidence to review a larger contiguous-workspace
budget versus a bounded-block design, including host validation arrays. The full
2^32 three-array payload remains48 GiB, not a measured allocation. No full s16
attempt or whole-run runtime estimate is supported by this checkpoint.


[CPU validation-06](../hpc-s16/runs/validation-06/results.json) and its
[analysis](../hpc-s16/runs/validation-06/analysis.json), job72845, pass full
CGO-disabled Go and native repository suites, focused race/vet, legacy build and
fresh-cache s6 equivalence. Reused [validator](../hpc-s16/run.py) and
[launcher](../hpc-s16/run.sbatch) with `validation-06 --validate`, overriding the
allocation to4 CPUs,16 GiB and15 minutes. The validation controller uses an8 GiB
RSS guard and800-second deadline. Its exact revision, source hashes, environment,
commands, profiles and cache checksums are retained separately; these small checks
do not validate full s16 feasibility. Python/shell syntax, Go formatting,
documentation links and `git diff --check` pass. No native instruction profile is
available because host perf permissions reject it.

A concrete next increment would test2^26,2^27,2^28 values with a4 GiB explicit
device cap, retaining16 GiB host request/12 GiB RSS,one GPU/16 CPUs and the
30-minute allocation/25-minute controller limit. Largest resident payload would
be3 GiB, with about1 GiB for tables/plans and margin; driver/context is separate.
The measured818.7 MiB process peak scaled by8 is6.40 GiB, a planning scenario,
not a host-RSS guarantee. Keep the guard and stop rather than exceeding it.
This was the proposed increment at the scaling-01 checkpoint; the explicit
user approval and scaling-02 submission are recorded below. The
[current plan](../../docs/encoding-scaling.md) requires agreeing numerical
budgets before increasing them. Full-domain metadata support is already complete;
full-domain execution remains a later, larger decision.


## Approved larger increment

User approved the proposed4 GiB cap for dimensions26..28. Run with
`sbatch --output=experiments/gpu-scaling/slurm-%j.log experiments/gpu-scaling/run.sbatch scaling-02 large`.
The optional `large` stage chooses only this fixed sweep/cap; unknown stages are
rejected. All other bounds/methodology stay unchanged. Earlier scaling-01 evidence
remains immutable. The completed measurements are recorded below; no full-domain
execution is included.


## Completed larger increment

[scaling-02 results](runs/scaling-02/results.json),
[verified analysis](runs/scaling-02/analysis.json),
[raw measurements](runs/scaling-02/measurements.json) and
[worker log](runs/scaling-02/scaling.log), job72847, pass. Same recorded
Xeon6710E/H200 NVL/toolchain as the smaller checkpoint; see the run's GPU,
CPU, compiler, toolkit and allocation logs for exact inventory. Measurements
remain profiled, fixed-order operation comparisons: two GPU samples, one CPU
sample. Times include host/device transfers but exclude separately recorded
setup and correctness comparisons. Seconds:

| Domain | CPU forward | GPU forward | GPU inverse | GPU resident product |
|---|---:|---:|---:|---:|
| 2^26 | 14.962 | 0.959 | 0.903 | 2.566 |
| 2^27 | 31.001 | 1.665 | 1.579 | 4.370 |
| 2^28 | 65.114 | 2.651 | 2.533 | 6.669 |

At2^28 (268,435,456 values), forward is24.56x faster than the16-core native
CPU sample; inverse is26.08x. The complete CPU product reference takes219.307 s,
but includes the canonical Go pointwise multiplication loop, so its32.88x ratio
is not a native-only CPU/GPU comparison. Exact full-array forward/inverse/product
comparisons pass at every size. The100 small resident regression checks,
legacy-library compatibility, malformed/over-budget metadata checks and small
CUDA memcheck pass; memcheck does not cover the large arrays. Device ownership
returns to zero on close.

Peak explicit device ownership is3,222,047,644 bytes (3.001 GiB), below the
approved4 GiB cap. Sampled command process RSS peaks at4,445,843,456 bytes
(4.141 GiB), below12 GiB. Retained private cache totals8,540,751 bytes
(8.145 MiB). Setup takes0.527 s separately. All279 exact source hashes and
sealed logs/profiles/cache artifacts verify. Existing caches remain untouched.
No full suite is repeated for the fixed sweep selection alone; validation-06
covers the preceding implementation and current-source focused tests plus the
hardware regression cover the stage-selector change.

This extends measured execution by another8x, leaving a16x domain-size gap
(four doublings) to2^32. The GPU remains faster as sizes grow in the measured
range. This does not establish full s16 preparation speedup: root generation,
small products, later full-domain operations and persistence have different
costs. The next gate is a reviewed budget and validation schedule for the last
four sizes or a validated blocked alternative. Three full-domain arrays alone
need48 GiB of device memory, versus the3 GiB payload measured here; host reference
arrays also need separate accounting. No larger job has been submitted.
