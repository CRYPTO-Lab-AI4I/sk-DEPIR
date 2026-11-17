# Host-backed transform and split-product comparison

Status: **bounded validation76756 passed** on HPC gnode04 (4 H200 GPUs,
4 CPU cores, 8 GiB requested host RAM). Complete job: **11 min 44 s**;
controller: **702.609 s**. Producer revision `10b5f9a`.
[Submission receipt](submission.json), [raw results](runs/components-01/results.json),
[reconciled analysis](runs/components-01/analysis.json),
[component records](runs/components-01/components.json),
[accounting](runs/components-01/accounting.txt), [Slurm log](slurm-76756.log).
Work: SCALING-103/104. No full s18 preprocessing run or default change.

## Reproduction and evidence

[Controller](run.py), [Slurm entry point](run.sbatch), [reviewer](analyze.py),
[review tests](test_review.py), [integer model](model.py), [payload projections](model.json).
The controller records the exact Git revision and per-file source hashes in
`runs/<name>/results.json`, including uncommitted source identity if present.
Results, command logs, CPU profiles and resource observations are retained there;
fixtures, actual comparison outputs, binaries and native profiles are retained
outside Git at `.cache/liftrs/preprocessing-tiles/<name>/` and sealed by checksum.
All **578 source / 45 evidence / 377 cache identities** reconcile without overrides.
The completed run also passes the inherited complete small-workflow regressions.

```sh
python3 -B experiments/preprocessing-tiles/test_review.py
python3 -B experiments/preprocessing-tiles/model.py
python3 -B experiments/preprocessing-tiles/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/preprocessing-tiles/slurm-%j.log \
  experiments/preprocessing-tiles/run.sbatch components-01
# After the completion ping:
python3 -B experiments/preprocessing-tiles/analyze.py components-01
```

HPC budget reuses the approved bounded comparison: four H200 GPUs on one node,
four CPU cores, 8 GiB requested host RAM, 6 GiB process RSS guard, 4 GiB Go target,
2 GiB new retained storage and 4 GiB scratch. Native component engines cap owned
allocations at 128 MiB/card; inherited small CLI regressions retain their
1 GiB/card cap (4 GiB aggregate). Context/runtime allocations are outside owned
payloads. The shared controller enforces ten minutes per command, fifteen minutes
overall and a twenty-minute Slurm cutoff. gnode03 is excluded; other eligible HPC
nodes are allowed. This budget does not authorize a full-size run.

## What changes, and what stays shared

The [host adapter](../../encoding/liftrs-cuda/tiled_preparation.cuh) borrows a fresh
resident executor's validated field metadata and capped allocations. It keeps a
full extension-field array in host RAM, plus one bounded host staging tile and
one device tile per GPU. Taylor source windows and row/column vectors are gathered,
processed and scattered through the existing arithmetic recurrences. Global row
coordinates remain 64-bit. Full-device transpose arrays are avoided. Transfers
and launches are synchronous waves; CPU packing is serial, with no promised
copy/compute overlap or pinned-memory benefit. There is no disk scratch array.
One thread processes each vector, as in the resident local-vector kernel;
small-domain throughput cannot establish performance for full-size long vectors.

The [split orchestrator](../../encoding/liftrs-cuda/split_preparation.cuh) calls
the same resident product for every block pair, then XORs overlapping coefficients
into one host output. The scaled actual s18 K/N ratio produces two P blocks and
seven G blocks, hence fourteen products. At full s18, the largest padded domain
is N/4; one tail product uses N/8. No polynomial arithmetic is duplicated.
Neither prototype is a public Go/C ABI, production preprocessing route or
implicit memory fallback. The existing resident default is unchanged.

## Explicit memory–time trade-offs

These are payload projections from the linked integer model, not measured peaks:

| Final s18 identity product | Resident | Split prototype geometry |
|---|---|---|
| Three uint64 GPU arrays, aggregate | 1536 GiB | 384 GiB |
| Host uint32 output | 256 GiB | 256 GiB |
| Extra host subproduct temporary | none | up to 64 GiB minus 4 bytes |
| Product calls | 1 | 14 |
| Coefficient H2D/D2H traffic | about 256 GiB each | about 815.616 GiB each |

Split arithmetic work is about 3.184 times the resident N log2(N) model, **not a
runtime estimate**. Metadata, context, allocator, filesystem cache and headroom
are additional. This does not model discovery, roots, reciprocal work, input
lifetimes, full workflow peak or multi-node communications. No extra canonical
storage is introduced; host tiling adds transfers/packing and splitting repeats
transforms/transfers. Their bounded measured latency penalties are reported below; full-size costs
remain unknown. Retain the
faster resident path whenever it fits; any later slower mode needs explicit
selection and a reviewed memory constraint. Four GPUs being sufficient for this
one projected product does not establish complete s18 feasibility.

## Measurement and correctness boundaries

The [driver](check.cu) first measures transforms, then products. Independent native
CPU fixtures come from the opt-in [fixture test](../../encoding/liftrs-shared/preprocessing_tiles_test.go).
Transform cases cover true s4 and s17 fields, odd/even dimensions, both storage
widths, short/tail tile batches and forward/inverse recovery. The largest case
has 2^22 values (32 MiB in uint64 storage), with 1 MiB and 8 MiB device tiles on
one/four GPUs. Smaller cases also use two GPUs and 1536/8192-byte tiles. Separate
s17 affine-row checks use coordinates above 2^32 and an independent Go Horner
oracle; they do not constitute a full 2^34-point transform.

Every output is compared elementwise with a complete independent CPU result;
small or sampled scalar Horner checks and small schoolbook products supplement
that reference. Persisted actual outputs are fully hashed against fixture
manifests. The product suite uses 64, 512 and 2^22-point domains, compares full
and split GPU products with CPU, and checks fourteen subproducts. CUDA memcheck
repeats small cases including the wide-coordinate witness and checks leaks,
input rejection and recovery. Existing CUDA-free/native/race/ASan/CLI regressions
run through the shared controller before components.

Each case has constructor setup, a separately reported first call, then two
profiled timing samples (one for memcheck). Resident first calls include workspace
allocation; tiled constructors allocate their tiles and plans. Transform call
latency includes field validation, host packing/scattering, transfers, kernels
and synchronization. Product latency includes output initialization, repeated
GPU calls and overlap assembly. Input cloning, CPU comparison and output-file
persistence are outside call clocks and remain in whole-command timings. Copy
counters count requested coefficient payload, not physical PCIe/NVLink traffic;
`kernel_wall_seconds` includes launch/synchronization, not pure GPU duration.
Owned device/staging payloads are not NVML or process peak measurements. Shared
resource sampling captures process RSS and job context separately.

Go fixture CPU profiles and native gprof profiles are retained; native timings
include gprof instrumentation and progress logging. There is no GPU instruction
profile or unprofiled baseline. Fixed sample order and small sample count preclude
strong speedup claims. GPU stages and block progress flush timestamped events;
shared resource samples run throughout. Logging/profile overhead is included,
not independently isolated. A stalled phase is not counted as forward progress.

## Completed component validation (SCALING-104)

Eight transform fixtures, three product fixtures and one wide-coordinate witness
pass complete independent CPU comparisons. All retained actual outputs match
canonical fixture SHA256 digests. The reviewer checks 171 normal and 76 sanitized
records, exact one/two/four-card coverage and both timed samples where applicable.
CUDA memcheck and the inherited CLI memcheck both report **0 errors and 0 leaked
bytes**. Invalid input/tile/host-cap rejection and subsequent reuse pass, as do
native/reference/race/ASan/vet, root/cache restarts and CUDA-free code readback.
No arithmetic or execution code changed during this results review.

The following are **two warm profiled samples**, with units explicit. Setup and
first-call timings are in the linked component records; comparison/persistence
remain outside call clocks. GPU memory means aggregate owned allocations, not
driver/context/physical peak. These cases use the true s17 extension field with
**2^22 points**, not a complete s17 or s18 preprocessing workload.

### Transform: 32 MiB host array

| GPUs | Transform storage | Forward runtime | Inverse runtime | Owned GPU allocations | Extra host staging |
|---|---|---|---|---|---|
| 1 H200 | Resident | 186.772–186.988 ms | 162.470–163.213 ms | 65.569 MiB | none in this adapter |
| 1 H200 | Host-backed, 1 MiB/card tile | 5.943–5.945 s | 5.828–5.832 s | 2.569 MiB | 1 MiB |
| 1 H200 | Host-backed, 8 MiB/card tile | 1.861–1.868 s | 1.837–1.838 s | 9.569 MiB | 8 MiB |
| 4 H200 | Resident | 235.820–236.090 ms | 208.893–210.278 ms | 70.276 MiB | none in this adapter |
| 4 H200 | Host-backed, 1 MiB/card tile | 2.180–2.183 s | 2.145–2.153 s | 10.276 MiB | 4 MiB |
| 4 H200 | Host-backed, 8 MiB/card tile | 1.303–1.305 s | 1.289–1.295 s | 38.276 MiB | 32 MiB |

The fastest measured host-backed option (four cards, 8 MiB/card) takes about
**5.53 times the forward / 6.16 times the inverse latency** of four-card resident,
using ratios of two-sample means. One card is faster than four for this resident
component; do not infer card scaling for the full domain. Each tiled direction
uploads **3230.969 MiB** and downloads **327.914 MiB**, versus 32 MiB each for
resident. Identical transfer counts across tile sizes expose repeated Taylor
source gathering, not just vector packing. Four-card 8 MiB forward time includes
about **1.055 s Taylor**, **0.101 s columns**, **0.144 s rows** and **4.7 ms input
validation**. Host packing/scattering accounts for about0.349 s; H2D/D2H walls
about0.317/0.129 s. These breakdowns overlap the phase totals and must not be
added to them. The four-card 1 MiB schedule has4463 waves; 8 MiB has4114.

### Product: 16 MiB host output

| GPUs | Product schedule | Runtime | Owned GPU allocations | Extra host temporary | Product calls |
|---|---|---|---|---|---|
| 1 H200 | Resident | 396.535–399.061 ms | 97.569 MiB | none | 1 |
| 1 H200 | Split | 2.1875–2.1882 s | 25.603 MiB | 4 MiB minus 4 bytes | 14 |
| 4 H200 | Resident | 539.435–539.450 ms | 102.276 MiB | none | 1 |
| 4 H200 | Split | 3.019–3.020 s | 30.414 MiB | 4 MiB minus 4 bytes | 14 |

The split product takes **5.50 times** the one-card resident latency and
**5.60 times** the four-card resident latency (ratios of sample means). Its array
payload is one quarter of resident, while metadata keeps total owned allocations
above that ratio. Output remains16 MiB; input coefficients remain borrowed and
immutable. Coefficient uploads/downloads are about **50.976/50.976 MiB**, versus
**16 MiB plus 4 bytes / 16 MiB** resident. XOR overlap assembly takes about8.3 ms
on one card and8.6 ms on four, so it is a small part of the observed penalty.

### Profiles, memory and limitations

The normal component command takes110.255 s; component memcheck347.750 s;
CPU fixture generation11.550 s. Maximum sampled process RSS across commands is
**676.516 MiB** (normal component command), including fixtures/references/runtime,
not isolated algorithm memory. Scheduler batch MaxRSS is **686108 KiB**;
metrics use different samplers. Reusable fixtures, binary outputs, binaries and
profiles occupy **592.426 MiB** outside Git; no cache is discarded.
Native gprof is retained in the cache with [its report](runs/components-01/native-profile-0.log),
as is the [Go fixture CPU profile](runs/components-01/cpu-fixtures.pprof).
The gprof host sample attributes63.23% self CPU time to host vector packing work;
this is host CPU sampling, **not a share of GPU or wall time**. Profile overhead
and fixed ordering remain included, with no unprofiled baseline.

The sanitizer wide-row witness took316.264 s in a single vector kernel and its
inverse, with start/end counters but no internal advancing progress; controller
heartbeats correctly report unknown remaining work. That logging gap is retained
as a limitation, not counted as progress. Split or instrument such a long vector
before future long-vector jobs; the normal bounded case was short. Sanitizer
latency is excluded from normal component comparisons.

These results complete the requested bounded transform/product measurements.
They do **not** justify extrapolating the5.5–6.2-times penalties to full s18 or
all preprocessing: vector lengths, utilization and stage proportions differ,
and root construction dominates the existing full s17 workflow. The next plan
step compares complete four-H200 hybrid and sixteen-H200 distributed lifetimes,
communication and stage costs. Current tiling needs less repeated Taylor traffic
and fewer small launches before it is a credible fast full-size choice. Keep the
resident default; no new run, mode adoption or larger budget follows from this
review.
