# Host-backed transform and split-product comparison

Status: prototypes prepared; GPU correctness and timing validation pending.
Work: SCALING-103. No full s18 preprocessing run or default change.

## Reproduction and evidence

[Controller](run.py), [Slurm entry point](run.sbatch), [reviewer](analyze.py),
[review tests](test_review.py), [integer model](model.py), [payload projections](model.json).
The controller records the exact Git revision and per-file source hashes in
`runs/<name>/results.json`, including uncommitted source identity if present.
Results, command logs, CPU profiles and resource observations are retained there;
fixtures, actual comparison outputs, binaries and native profiles are retained
outside Git at `.cache/liftrs/preprocessing-tiles/<name>/` and sealed by checksum.
No results exist until the submitted job completes and is reviewed.

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
transforms/transfers. Their measured latency penalties remain unknown. Retain the
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
