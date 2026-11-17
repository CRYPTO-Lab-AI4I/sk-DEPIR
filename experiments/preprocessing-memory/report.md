# Lower-memory server preprocessing prototypes

Status: **bounded spectrum streaming validation76704 passed; export integration remains next**.
This starts the [s18 design sequence](../../docs/encoding-scaling.md#s18-preprocessing-memory-design-and-comparison-scaling-096).
It does not extend full preprocessing beyond s17 or change any default mode.

## Reproduction and evidence

[Runner](run.py), [Slurm launcher](run.sbatch), [completion analyzer](analyze.py),
[source-derived memory model/self-check](memory.py), [model results](memory-model.json)
and [independent arithmetic/failure tests](../../encoding/liftrs-shared/preprocessing_stream_test.go).
The runner reuses the [existing guarded controller](../gpu-preprocessing-workflow/run.py)
and its complete small-workflow/regression/sanitizer checks. The copied source is
hashed against the job's recorded revision; commit the producer before submission.
Runtime results and logs: `runs/stream-01/`; retained binaries, reference caches
and profiles: `.cache/liftrs/preprocessing-memory/stream-01/`. The [completed analysis](runs/stream-01/analysis.json), [review summary](runs/stream-01/review.json)
and [raw producer results](runs/stream-01/results.json) now provide reconciled
bounded runtime evidence. Full-size preprocessing remains unmeasured.

```sh
python3 -B experiments/preprocessing-memory/memory.py
python3 -B experiments/preprocessing-memory/run.py --check
module load slurm/slurm/25.05
sbatch --output=experiments/preprocessing-memory/slurm-%j.log \
  experiments/preprocessing-memory/run.sbatch stream-01
# After user completion ping only:
python3 -B experiments/preprocessing-memory/analyze.py stream-01
```

Budget: existing bounded **4 H200 / 4 GiB aggregate owned** admission, one HPC
node, 4 CPU cores, 8 GiB requested host RAM / 6 GiB process RSS guard / 4 GiB Go
soft target, 2 GiB new retained cache, 4 GiB scratch; 10-minute worker commands,
15-minute controller, 20-minute Slurm cutoff. The new stream tests cap each card
at128 MiB (512 MiB aggregate); inherited complete-workflow checks retain1 GiB/card.
Actual node, CPU, GPU, topology, compiler/CUDA/Go versions and source revision are
recorded in results; no node is pinned, and gnode03 is excluded. Native/GPU phase
logs and stream element counts are flushed; 30-second resource heartbeats explicitly
leave remaining work unknown. CPU profiles cover whole new test phases; timings
include profiling, logging, hash/writer and synchronization overhead.

## First host-lifetime audit

| Final host data at s18 | Principal payload | Lifetime / action |
|---|---:|---|
| Support | 121.446 GiB | Caller-owned immutable support; remains live in the current workflow. A future consuming/export workflow may drop its reference after root work. Do not mutate a caller's support. |
| Normalized P / root checkpoint | 60.723 GiB | Checkpoint aliases P; it is not a second full root allocation. Retain P until canonical publication. |
| Generator G | 195.277 GiB | Needed for spectrum upload and reproducible generator output; retain until both complete. |
| Host spectrum | 512 GiB | New primitive can write this directly from resident device shards using a bounded tile. |
| GPU product operands | Released before final spectrum | Existing `ReleaseProducts` already removes dead second operands; retain reusable transform data/scratch. No new saving claimed. |

The [shared preparation flow](../../encoding/liftrs-shared/preprocessing.go) and
[polynomial verification](../../encoding/liftrs-shared/preprocessing_core.go) already
let identity-result temporaries die before the final spectrum. Go reachability
and actual RSS/GC behavior differ. Loaded discovery/root/reciprocal recovery arrays
can also remain referenced by caller options; dropping those needs a separately
validated ownership change. Do not sum mutually exclusive stage arrays or count
root/P aliases twice.

With a1 MiB stream tile, final host-array payload projects **889.446 → 377.447 GiB**
at s18 (saving approximately512 GiB). Releasing the caller's support later could
reduce that final stage further to approximately256.001 GiB, but is **not
implemented**. Identity-stage principal arrays still project633.446 GiB; discovery
still projects770.893 GiB, and Newton/product intermediates, runtime, caches and
headroom are additional. Therefore this primitive alone does not admit fulls18.

## Streamed spectrum primitive

`Preparation.SpectrumTo(g, n, writer, tileBytes)` is an explicit low-level GPU
server operation. Both resident and streamed output call the same zero-padding,
upload and forward-transform implementation. Optional plugin symbols preserve
older plugin loading; requesting streaming from an older plugin returns an error.
The Go engine lock covers transform and all chunk reads/writes. No Go pointer or
writer is retained; errors leave partial bytes for the caller to reject, and the
engine can be reused. Canonical output is little-endian4-byte slots throughs16,
8-byte slots ats17, with bounded whole-slot tiles. Full-domain n is required for
an actual code spectrum; smaller admitted powers of two are component tests.

This adds **no transforms, no device arrays and no extra canonical file bytes**.
It replaces one whole download followed by serialization with synchronous tiled
download/write calls. Transfer and serialization latency can increase; penalty
is **unknown**. Filesystem cache can still charge written bytes to host memory.
Disk publication, P/G identity checks and canonical header/digest/readback must be
integrated and validated before advertising a full streamed preprocessing mode.
The current public workflow/default and s17/full-size admission are unchanged.

Correctness compares every output point with scalar Horner at s8 and true s17
width on1/2/4GPUs, peer/local, including tiles crossing shard boundaries and tails.
Short/error writers, bad shapes/field/tile bounds, closed engines and reuse after
errors are checked. CUDA memcheck repeats small cases on one GPU; inherited
workflow memcheck covers four GPUs. The32 MiB spectrum component uses dense G,
two balanced samples per strategy, complete matching hashes and a1 MiB stream tile.
Its boundary is transform+download+hash/write, excluding engine/input setup;
**it is not disk bandwidth, full preprocessing latency or a peak-memory comparison**
(the test deliberately retains the resident comparison array in both strategies).
The resident comparison uses the existing generic transform adapter, while the
stream path uses base-coefficient upload; report these preparation differences
before interpreting the timing ratio.

## Remaining measurements and choice

1. After correctness passes, integrate explicit streamed canonical publication
   with scoped host ownership, atomic manifests, P/G identity and full byte/digest
   equality against the resident path on small complete codes. Measure separate
   stage RSS/cgroup/cache peaks and setup/transform/write/readback latency.
2. Measure a host-backed tiled transform (bounded slabs, required transposes,
   transfers and optional scratch files) against resident execution with the same
   field/order and input/output hashes. Report per-pass bytes and peak host/device
   usage. Existing resident-vector tile calibration is not a host-backed full FFT.
3. Measure a split polynomial product with exact coefficient assembly against
   the current product and independent CPU multiplication. Include all subproducts,
   overlap accumulation, host output, transfers and temporary storage. A proposed
   P=2-block/G=7-block final identity requires14 subproducts; this is a scenario,
   not a measured runtime ratio or adopted design.
4. Compare a four-H200 hybrid against sixteen H200s across four HPC nodes using
   complete stage lifetimes, measured kernel/transfer/I/O costs, persistence and
   explicit unmeasured communication costs. Present uncertainty and ask the user
   to select any material tradeoff before full implementation/admission. Neither
   cross-node execution nor fulls18 storage is admitted by this bounded run.

## First validation submission

Job **76704**, prepared executable sources at commit `46d16d2`, uses the budget
above. [Receipt](submission.json); results `runs/stream-01/`, scheduler log
`slurm-76704.log`, retained binaries/cache `.cache/liftrs/preprocessing-memory/stream-01/`.
The controller records its execution-start Git revision and exact copied source
hashes; later message-board/receipt updates do not change executable sources.
No scheduler polling or log inspection follows submission; await completion ping.

## Completed bounded validation (SCALING-099)

**76704 passed**, HPC nodegnode04,4H200 NVL,4 CPU cores,8 GiB requested host RAM.
Slurm **COMPLETED0:0**, **3 min 59 s**; controller **236.909 s**. Go1.27.3,
GCC14.2.0,CUDA12.9.86,driver580.126.20; producer revision
`46d16d22d247364dd05cbf0dbcac50a7d8cd6d89`. [Raw results](runs/stream-01/results.json),
[analysis](runs/stream-01/analysis.json), [review](runs/stream-01/review.json),
[accounting](runs/stream-01/scheduler-completion.txt), [scheduler log](slurm-76704.log).
Reconciliation verifies **561 source /43 evidence /74 retained-cache files**,
without source overrides. Sampled maximum command process-group RSS is
**0.565 GiB**, including build/validation work; it is not isolated stream memory.
Retained binaries/cache are31,275,710 bytes, outside Git at the path above.
Profiles and complete small preprocessing fixtures remain reusable.

Twelve scalar-Horner cases pass at s8 and true s17 extension width with1/2/4GPUs,
peer/local, all points checked, including tiny/tail/cross-shard stream chunks.
Writer short/error, field/shape/tile bounds, closed state and engine reuse after
failure pass. Stream CUDA memcheck reports **0 errors /0 leaked bytes**; inherited
four-GPU complete-workflow memcheck also passes. Existing small canonical-byte,
fresh/cache/root-restart/CUDA-free readback, native regression/race/ASan/vet checks
pass. This validates the primitive, not full streamed artifact publication.

| Component on4H200, s17 arithmetic,32 MiB spectrum | Sample1 | Sample2 | Mean |
|---|---:|---:|---:|
| Resident transform + output hash/write |242.058 ms |221.534 ms |231.796 ms |
| Stream transform + output hash/write,1 MiB tile |217.848 ms |215.225 ms |216.536 ms |

All four full hashes match. Stream mean is **6.58% lower** in these two balanced
samples; no slowdown is observed here. Do not infer a full-size speedup or a
cost-free512 GiB export: both timings include profiling/logging, use an in-memory
hash sink rather than disk, and have different input preparation boundaries.
There is no confidence bound, end-to-end preprocessing comparison or measured
RSS saving (the resident array is intentionally retained during both strategies).

The final-stage s18 array projection stays889.446→377.447 GiB with1 MiB streaming,
not a measured peak. Earlier discovery/product pressure and full publication
remain unsolved. **Next:** integrate explicit canonical streaming with scoped
host ownership and atomic verified publication on small complete codes, then
measure representative host-backed transforms and split products before selecting
the four-versus-sixteen-H200 route. Resident defaults and fulls18 admission stay
unchanged. GPU encoding retry76706 was not inspected during this review.
