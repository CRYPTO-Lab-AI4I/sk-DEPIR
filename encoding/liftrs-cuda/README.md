# Optional CUDA encoding backend

This is the maintained optional GPU path alongside liftrs-go/liftrs-cpp. Client
hardware determines backend choice; public code preprocessing may be obtained
elsewhere. CPU-only clients/builds do not require this package, CUDA headers,
toolkit or a GPU. CUDA-only Go builds do not require liftrs_cpp.

The shared API is `ScalingOnlinePlan.WithCUDA(CUDAOptions)`, built with
`liftrs_cuda` and cgo. Prefer that API to these lower-level trusted-metadata
primitives. It accepts the same verified P/spectrum/code specification as CPU
encoding, supports packed arithmetic s2..16, and returns the same canonical
uint32 codeword. This is standalone mathematical encoding, not masking,
application/cryptographic integration or production-security validation.

For the width-selected `ScalingCodeData.WithCUDA(CUDACodeOptions)` client API,
full resident s17 encoding passes76706 on4H200, with CPU-identical canonical and
packed output. See [API and measured boundaries](../../docs/gpu-encoding.md#full-s17-resident-gpu-validation).
The older API's s≤16 limits/memory modes remain unchanged.

## Build and deployment

On a CUDA machine, build a trusted shared library for its target architecture:

```sh
nvcc -std=c++17 -O3 -lineinfo -arch=native -shared -Xcompiler=-fPIC encoding/liftrs-cuda/matrix.cu -o /path/to/libliftrs-cuda.so
# CPU and CUDA are independently selected build features:
go build -tags liftrs_cuda ./...
go build -tags liftrs_cpp,liftrs_cuda ./...
```

Supply the verified library path explicitly. Dynamic loading keeps ordinary Go
and native CPU builds independent of CUDA. The CUDA driver/runtime and device
must support that compiled binary; an architecture-specific build is a deployment
cost, not a prerequisite for CPU clients. This first GPU implementation uses
NVIDIA CUDA; the shared code data/CPU path is not tied to CUDA. Keep existing
preprocessing/cache hashes when transferring data between machines/backends.

## State and ownership

Initialization copies public base-field tables, two matrix descriptions and
P/spectrum to the device. The high-level encoder does not retain host P/spectrum
slices. It owns three reusable full-domain arrays, uses a padded prefix domain
covering the full P*message product, and retains all five transforms, pointwise
products and canonical assembly on device until final download. Only the input
message and output word transfer per call; code uploads are separate reusable
setup. Device-owned scratch is repurposed for assembly parity/error flags after
transforms. No Go pointer is retained after a synchronous call.

Calls serialize; Close waits for active work and releases explicit state.
Canonical assembly validates a primitive generator before use, so scatter
positions are disjoint. One chunk computes a disjoint source interval and local
parity; a joined reduction writes position zero. Reject out-of-field inputs
before kernel access, detect escaped base-field results before host output copy,
clear every reused padding region and reset the error flag each call.

Explicit caps include arrays/tables/plans, exclude driver/context memory. The
array payload for s16 is about67.42 GiB:48 GiB workspace plus19.42 GiB code.
This is a model, not measured peak or minimum total device capacity; small plans,
tables and driver overhead require headroom. The completed s16 experiment
records explicit admission: the [full cached runner](../../experiments/cuda-s16-encoding/report.md)
uses a96 GiB cap and requires at least98 GiB free device memory. Smaller s uses
smaller arrays; neither s nor a particular database capacity is hardcoded.
Low-level legacy transform/product ABIs remain available. Old binaries missing
online symbols fail clearly for online construction while remaining usable for
historical primitives. Experiment import/build shims preserve old commands.

Field arithmetic and table indexing are **not constant-time**. Inputs/code
metadata are trusted within the academic prototype contract; checksums establish
identity rather than mathematical authenticity. No claim of production side-channel
security is made. CPU and GPU correctness/performance validation is recorded in
[the online experiment](../../experiments/cuda-encoding/report.md): builds, small
full outputs, s10 workflow and CUDA memcheck/leak checks pass. The
[full s16 run](../../experiments/cuda-s16-encoding/report.md) also matches CPU:
85.95/60.22 s encoding,7.19 s reusable setup,67.423 GiB owned device allocations
and39.024 GiB peak sampled host RSS. Transfers are included; load/packing/checks
are separate. Timing comparisons with earlier CPU runs are historical.

## Opt-in buffer-reuse experiment

`NewBoundedReuse`/`gm_create_reuse` provide an experimental online-only engine
with two work arrays; default constructors and the public shared `WithCUDA` keep
the validated three-array layout. Reuse requires2L<=N; Apply/Product are rejected.
Prefix operands share an arena, and final assembly uses dead transpose scratch
with separate parity storage. Small shapes outside this prefix bound fail clearly,
not through silent fallback. Old libraries without the reuse ABI remain usable
for their existing operations. See the [validation runner/report](../../experiments/cuda-buffer-reuse/report.md)
for test-only selection, resource limits and completed measurements: s16 uses
51.427 GiB owned device memory,23.7% less, with57.418/56.975 s encoding and matching
CPU hashes. Matched s10 shows no observed encoding slowdown; s16 timing comparisons
are historical across nodes. Public shared selection remains unchanged.

## Transform-only scratch study

`NewTransformBounded(..., directRows)` explicitly selects a component experiment:
false owns data+full transpose scratch; true operates rows in original layout
and owns data only, potentially trading coalescing/runtime for memory. Product
and SetOnline are rejected. The [runner](../../experiments/gpu-transform-memory/report.md)
defaults to `--layout fast`; testing the lower-memory mode requires `--layout direct`
or `--layout compare`. Validation now passes through2^25 values; buffers256→128 MiB
at that size, with2.1% forward/2.8% inverse slowdown (larger tested cases roughly
2–8%). Per policy, fast stays default when it fits; direct remains capacity-driven
or explicitly selected. Public online defaults are unchanged; safe canonical
output assembly remains a separate memory gate, so this is not fulls16 encoding
memory/performance evidence.

## Experimental compact online output

`NewBoundedCompact` explicitly selects direct-row transforms and a single full
arena. After the final inverse it compacts base values through a bounded2 MiB
tile into one uint16 half, then writes canonical output into the disjoint other
half. It downloads half as many bytes and expands backwards into the existing
owning host uint32 result. Range checking precedes narrowing. Requires s<=16 and
2L<=N; it may run slower from direct access, compaction and host conversion.
`CompactStats` exposes those costs for the most recent call, interpreted only
after successful EncodeOnline and with other encodes excluded. Older libraries
missing optional symbols fail clearly. Public WithCUDA/default constructors are
unchanged. The [explicit CLI experiment](../../experiments/gpu-output-memory/report.md)
defaults resident; `--memory-mode compact` is required when trying this mode.
Validation passes two full s16 calls at35.429 GiB owned device memory and
101.5/93.3 s encoding, versus prior resident-reuse51.427 GiB and57.4/57.0 s
(historical timing comparison). Matched s10 compact is9.6–10.7% slower. Host RAM
remains39.03 GiB. It stays explicit and optional; s17 is unsupported.


## Experimental fast-layout streamed code

`NewBoundedStreamed` retains the two-work-array fast-row layout but no device
P/spectrum arrays. Add prefix/full plans and call `SetOnline` as usual; this
constructor validates code and installs metadata without retaining host pointers.
For each call, use `EncodeOnlineStreamed(message, word, p, spectrum, tileElements)`
with the same verified immutable host code and a tile of1..min(N,2^26) elements.
The method borrows code until synchronous return and rejects output/code overlap.
Ordinary `EncodeOnline` rejects this engine because code must be supplied per call.
Missing optional library symbols fail explicitly. Close and owning host outputs
retain their ordinary lifetime rules; no automatic mode switch is added.

P uploads directly to prefix data; spectrum tiles borrow dead transpose scratch
between full forward/inverse. Copy/kernel joins precede tile overwrite and scratch
reuse. `StreamStats()` reports code-copy seconds, bytes, spectrum chunks and tile
bytes after a successful encode, excluding other encodes while reading it.
Code-copy time is already included in H2D/native totals. The tile needs no extra
GPU allocation. The [explicit experiment](../../experiments/gpu-streamed-constants/report.md)
uses `--memory-mode streamed` (default resident) and now measures32.005 GiB at s16,
with19.422 GiB code upload per encoding and unchanged resident host backing.
Validation passes:63.49/58.85 s full s16 encodings with retained CPU hashes,
about2.24 s code upload per call, and39.03 GiB host RSS. Matched s14 is5.6–6.7%
slower than resident; full s16 timing comparisons are historical. Memchecks report
0 errors/leaks. Public WithCUDA/CPU defaults remain unchanged; s17 is unsupported.


## Explicit packed output sink

`EncodePackedTo(message, writer, bufferBytes)` handles resident code;
`EncodePackedStreamedTo(message, p, spectrum, writer, bufferBytes, codeTileElements)`
borrows immutable streamed code. Both keep the engine lock through writes, so the
writer must not reenter it. Buffer1 byte..64 MiB bounds encoder staging only. Optional
ABI symbols fail clearly on older libraries. Fast resident/reuse layouts pack
s-bit canonical output directly into dead scratch; compact/direct-row mode rejects
this sink. Neighboring bit fields use atomic OR after zero fill, with range
checking before writes and a joined parity reduction. No full canonical array is
assembled on host or device; coefficient/transform arrays remain counted.

The packed read state is sequential and cleared after completion/error. Native
code retains no host pointer. Writer failures may leave a partial destination;
caller handles publication/Sync. PackedStats SHA256 is valid as the full emitted
result only on success. Write/hash timing is separate, download is included in
native transfer/total timing. GPU workspace is unchanged. Validated s16 packed
pipelines take65.5–66.2 s including writing/hash (23.01 GiB encoding RSS; separate
verifier31.01 GiB). Matched s14 file pipelines are27–29% shorter in two fixed-order
samples, including CPU packing in the owning reference. See the
[measured boundaries and cache costs](../../experiments/gpu-packed-output/report.md).
Owning/default constructors and s<=16 arithmetic bounds stay unchanged.

## Explicit host-backed transform component

`NewHostBacked(..., maxPoints, deviceBytes, tileElements)` and
`ApplyBacked(hostData, inverse)` keep the mutable uint32 domain in caller RAM,
with two device tiles and at most one host staging tile. AddPlan requires a tile
large enough for a full row and column. All arithmetic runs on GPU; CPU gathers
columns. Transfers repeat at each Taylor correction and row/column batch.
Resident transforms remain the default. No host-backed online encoder exists yet.

```go
engine, err := liftrscuda.NewHostBacked(library, s, polynomial, delta,
    logs, powers, maxPoints, deviceCap, tileElements)
if err != nil { return err }
defer engine.Close()
if err := engine.AddPlan(matrixWords); err != nil { return err }
stages, err := engine.ApplyBacked(hostData, false)
if err != nil { return err } // hostData may be partially modified on GPU failure.
_ = stages // Timing/transfer indices are documented in gpu_api.h.
```

Only s2..16 arithmetic; maxPoints<=2^32 is a format bound. Tile4..2^26 elements,
explicit allocation cap includes plans/tables but excludes driver context. Host
array4N bytes remains required, plus at most4T staging; GPU arrays8min(N,T).
No host pointer survives a call. Calls/Close serialize. Product and ordinary
Apply reject this engine. New ABI symbols are optional for older libraries.
The initial [component experiment](../../experiments/gpu-host-backed/report.md)
passed correctness, but N2^22/T65536 is60.5x slower than resident Apply despite
32.76→1.26MiB device savings. Per-stage kernels now parallelize work within
tiles at unchanged workspace/transfer bounds;
[their GPU comparison](../../experiments/gpu-backed-parallel/report.md) is pending.
Repeated transfers and extra launches can still make it much slower. Do not select this path
when the faster validated resident path fits without an explicit CLI choice.


`EncodeOnlineBacked(p,spectrum,message,arena,word,omega)` integrates five
host-backed transforms with bounded GPU products and orbit point generation.
Two disjoint host N buffers are required: one mutable workspace and one owning
canonical output. Native CPU scatters values and XORs coordinate-zero parity;
no N-sized orbit map is allocated. All borrowed pointers expire on return;
Go checks all mutable/immutable overlaps, native validates shape/fields and
primitive omega. Errors may leave workspace/output partial. The new online ABI
is optional: older transform libraries still load and reject this operation.
Stats[16] extends the ordinary10 online counters with bytes/staging/routing,
transform count and arena size (see gpu_api.h). The total excludes admission.
This is explicit and may be much slower: [validation and memory costs](../../experiments/gpu-host-encoding/report.md)
passes complete s10 correctness:1.61 s, device560400 bytes, versus native CPU
0.31–0.32 s. No full s16 runtime, host-backed packed sink or s17 support is claimed.

## Resident GPU server preprocessing

The [GPU preprocessing prototype](../../experiments/gpu-s17-preprocessing/report.md)
implements width-selected kernels and1/2/4/8 same-node peer shards through s17,
reusing the shared discovery/polynomial flow. Select `GPUDevices` explicitly in
`PreprocessingOptions`; `GPUBytes` caps each card. Build the server library from
both `matrix.cu` and `preparation.cu`. Common arithmetic/FFT/Taylor and sparse-plan
headers avoid separate s17 algorithms. Ordinary clients remain CUDA-free.
CUDA/peer validation is pending; source support is not a full s17 run or admission.
The old empty-device CUDA32 adapter remains a reference through s16. GPU server
execution is the active target; CPU preprocessing supplies correctness oracles.

## Width-selected resident client

The new experimental `ScalingCodeData.WithCUDA(CUDACodeOptions)` selects explicit
same-node devices and resident constants through s17. It reuses the preprocessing
executor's width-selected transforms/peer storage and the CPU cyclic schedule;
client views, transfers and canonical assembly are separate adapters. Build the
plugin with both `matrix.cu` and `preparation.cu`, and `-Xcompiler=-fPIC,-pthread`.
Read the [client API/ownership guide](../../docs/gpu-encoding.md) and
[bounded integration experiment](../../experiments/gpu-code-encoding/report.md).
Bounded runtime validation passes76426; no fulls17 encoding result is implied. Existing legacy
constructors/modes and CPU-only dependencies remain unchanged.

## Bounded server spectrum export prototype

`Preparation.SpectrumTo(g, n, writer, tileBytes)` shares the resident generator
spectrum's zero-padding/upload/forward transform and streams canonical extension
slots through one bounded host tile. The explicit tile cap is8 bytes..64 MiB,
rounded down to an8-byte boundary. Use n=2^(2s) for a full generator spectrum;
smaller admitted powers of two are diagnostic transforms. The writer runs under
the engine lock and must not reenter it. Errors leave partial output, which the
caller must not publish; no caller pointer survives return. Old preprocessing
plugins still load, but return an error if streaming symbols are unavailable.

This is a low-level server prototype, not an integrated preprocessing mode. It
removes the full host spectrum allocation but keeps resident device workspace;
bounded scalar/memcheck validation passes76704. The32 MiB in-memory hash-sink
component is measured; full-size transfer/disk latency remains unmeasured. The existing resident path/default
and s<=17 engine limit remain. See [tests, budget and lifetime model](../../experiments/preprocessing-memory/report.md).

## Experimental preprocessing components

`tiled_preparation.cuh` and `split_preparation.cuh` are diagnostic adapters
compiled with `preparation.cu` by the [bounded component driver](../../experiments/preprocessing-tiles/report.md).
They reuse field validation, allocation caps, plans and arithmetic recurrences.
The tiled adapter borrows a fresh resident executor exclusively; the caller owns
the full host array, with bounded per-card staging and device tiles. Split
products borrow immutable inputs and assemble characteristic-two overlaps by XOR.
These prototypes introduce no public ABI, automatic fallback or default change.
Bounded GPU validation76756 passes complete independent CPU/hash comparisons
and CUDA memcheck. The report gives measured allocation savings and substantial
latency penalties (about5.53 times forward-transform latency and5.60 times product
latency on four cards at2^22 points). Full-size costs/feasibility remain unknown.

## Experimental cross-node staging gate

The optional [CUDA/MPI probe](../../experiments/gpu-staged-transport/report.md)
reuses `transform.cuh` multiplication, canonical Go field metadata and independent
native CPU arithmetic to validate a GPU-producer/D2H/MPI/H2D/GPU-consumer pipeline.
Each rank uses one GPU and bounded pinned buffers; MPI sees host pointers only.
It introduces no backend/CLI default, public API, distributed transform/root
support or CUDA-aware-MPI requirement. Normal CPU-client builds stay independent.
Retry77281 passed on two HPC nodes with one H200 NVL and2CPU/4GiB requested
host RAM each. At32MiB input/rank, median complete pipeline is11.827ms versus
2.269ms matched host-only exchange. Full CPU/result/roundtrip checks and both
CUDA memchecks pass. Peak normal rank RSS421.016MiB; largest owned device
allocation33.500MiB/card. Actual root-polynomial transfer/merge is next;
GPU-direct, distributed preprocessing and fulls18 costs remain unmeasured.

The next [distributed-root component](../../experiments/distributed-root/report.md)
reuses this executor for partial roots and receiving-GPU polynomial merges.
Optional MPI transport is experiment-only (`liftrs_mpi` plus the existing CUDA/
native tags); bounded validation77285 passed on2H200/2HPC nodes with full
cross-rank/CPU coefficient equality and both CUDA memchecks. Largest synthetic
s17 subset262,144 exponents takes1.180s median pipeline; peak normal rank RSS
410.523MiB and ownedGPU15.241MiB/card. No ordinary client dependency or full
scalable distributed-preprocessing capability is added.

The bounded [rank-owned GPU preprocessing tree](../../experiments/ucx-cache-gpu-tree/report.md#review-of-78261)
**passed its full four-round crossover in job 78261**. All CPU/state/ownership/
transfer checks passed, with 16 clean CUDA memory-check processes. The validated
configuration explicitly selects UCX with `UCX_RCACHE_ENABLE=n`; earlier direct
failures under the default cache remain recorded, and no runtime default changes.

HPC: **2 H200s across 2 nodes, 4 logical ranks/CPU cores, 8 GiB requested host RAM**.
The largest partial synthetic s=17 fixture (262,144 exponents) took **1.153–1.247 s**
per-round median; normal process peak RSS was **410.836–412.703 MiB/rank**, owned
GPU arrays **15.241 MiB/rank** (contexts additional). Mode advantage reversed
between pairs, so no consistent staging/direct speed winner is established.
These are profiled, setup-excluded diagnostic timings, not full preprocessing.

Next: collective large-product/distributed FFT integration, followed by distributed
discovery and generator/checkpoint/publication composition. The coordinator still
holds a full source/reference and the final product still uses one executor.
Full s=18 preprocessing remains unrun; current full s=17 results and CPU-client
support are unchanged. [Original failed gate](../../experiments/root-merge-tree/report.md)
and [partial logging-failed retry](../../experiments/ucx-cache-gpu-tree/report.md#review-of-78249)
remain preserved in the evidence.

The internal [VectorStage adapter](vector_stage.cuh) borrows a fresh single-GPU
preparation executor exclusively, with global metadata/origin independent of
local capacity. Separate upload/apply/download operations retain a bounded
interleaved device tile across calls; no full host grid or public ABI is added.
[Bounded validation](../../experiments/gpu-vector-stage/report.md) independently passes48 normal /42 split-sanitized cases in78293, zero CUDA
errors/leaks. Raw controller failed a stderr-only summary check, now fixed;
component evidence verified without GPU rerun. It reuses the existing vector kernels and s≤17 admission.
Vector/transpose composition passed78302; Taylor-complete validation and full
s18 support remain pending.

### Experimental distributed vector owner

`NewVectorPreparation(options, localWords)` creates a dedicated single-GPU
server stage owner; `options.MaxElements` fixes the global domain while
`localWords` caps one interleaved tile. It requires the optional `pg_vector_stage`
plugin symbol; older plugins remain usable for existing preparation APIs.
`Apply32`/`Apply64` transform complete columns or affine rows at global vector
origins, with synchronous transfers and no transpose/Taylor conversion. Close
the owner after use. CPU-only client builds retain no CUDA/MPI dependency.

Given trusted s17 `PreparationOptions` for a4096-word domain and device0:

```go
stages, err := liftrscuda.NewVectorPreparation(options, 192)
if err != nil { return err }
defer stages.Close()
// Three length64 rows, interleaved as coefficient*3+vector.
if err := stages.Apply64(tile, false, false, 10, 3); err != nil { return err }
```

The [bounded composition experiment](../../experiments/distributed-gpu-vectors/report.md)
joins these stages to the shared rank transpose. Validation78302 passed all small CPU/output/proof and CUDA checks. Taylor-complete composition is now implemented and awaits the bounded GPU gate
below. The host-shard control layout/extra transfers are explicit and
do not change resident defaults or establish fulls18 feasibility.

`TaylorDegrees(level)` returns sparse source displacements in canonical order.
`Taylor32`/`Taylor64` update a `TaylorTile` snapshot in the same capped storage;
the caller supplies target words followed by one equally sized source window per
term. Unused source prefixes and beyond-block suffixes may be zero-filled.
Global scheduling/barriers remain the caller's responsibility, not implicit API
behavior. Given a populated snapshot and validated target extent:

```go
terms, err := stages.TaylorDegrees(level)
if err != nil { return err }
// snapshot has (len(terms)+1)*count canonical slots in returned term order.
_ = terms
window := liftrscuda.TaylorTile{Level: level, Quotient: true, Block: block, First: first, Count: count}
if err := stages.Taylor64(snapshot, window); err != nil { return err }
```

The [complete bounded FFT gate](../../experiments/distributed-gpu-fft/report.md)
reuses the extracted host-tiled source accessor/kernel. New Taylor symbols are
optional; missing symbols return errors without a CPU fallback. Local interface/
CPU checks pass; composed MPI/GPU/Taylor validation78307 is submitted. Existing s≤17
field admission, client independence and resident defaults remain unchanged.
