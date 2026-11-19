# Reusable code preprocessing

Preprocessing constructs public encoding data from a code specification, without
reading a plaintext database or client key. Delegated preprocessing targets the
fastest measured server execution. CPU-only client encoding remains a separate
requirement; the native preprocessing executor is an explicit backend, not an
automatic fallback from GPU execution.

The experimental `NewCodePreprocessor` session implements the m=2 procedure
through s17. Corrected bounded native/reference, regression, race and sanitizer
validation passes job74447. **GPU server execution is now the active target**:
a width-selected resident peer engine implements the heavy preprocessing kernels
on1/2/4/8 same-node GPUs in source. One-GPU checks pass job74453; corrected two-GPU validation
passes job74466, including peer-link lifetime sanitizer checks. Full resident s17 client encoding passes CPU76645 and four-H200 GPU76706
with identical canonical/packed hashes; preprocessing and encoding remain
separate measured phases. Full GPU preprocessing now passes75832:
genuine discovered support, P/G/spectrum, full identity and complete cache
readback; see [measured full-run results](../experiments/gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088). The legacy CUDA32
adapter remains available through s16; the resident prototype is selected
explicitly with nonempty `GPUDevices`. There is no CPU/host-backed fallback.
See the [scaling reference](encoding-status.md),
[plan](encoding-scaling.md#current-execution-plan-for-s17) and
[GPU validation report](../experiments/gpu-s17-preprocessing/report.md).

GPU count is chosen for the fastest measured complete server workflow, without
a two-card design ceiling. The current two-GPU test is only initial validation.
Compare more cards and, where useful, multiple nodes; include setup, communication,
host memory and persistence. The current peer engine handles same-node devices;
a cross-node implementation is still required before larger distributed runs.

## One procedure, interchangeable execution

| Responsibility | Implementation | Invariant |
|---|---|---|
| Exact monomial membership | Go `GoodMonomial`; native `preparation.hpp` | Paper Lemma 4.8; positive reachable multiples of q−1 make a monomial bad |
| Sampling and support union | `preprocessing_core.go: discoverSupport` | Accept only after the observed support has exact independently counted K |
| Root product, normalization, reciprocal and identity | `preprocessing_core.go: prepareCyclicPolynomials` | Verify the entire P·G=1+X^(N−1) before publication, including restart |
| Small products and Newton reciprocal | Existing `orbit_product.go`, `preparation_products.go` | Base-field coefficients stay uint32; full convolution padding prevents prefix aliasing |
| Transforms, large products, batched roots and enumeration | Native `preparation.hpp`, Go binding `preparation.go` | One C++ template selects uint32 through s16, uint64 at s17 |
| GPU discovery, roots, transforms/products | `liftrs-cuda/preparation.cu` and shared kernel headers | Same sampling/polynomial flow; width-selected peer shards and explicit per-device caps |
| Legacy CUDA32 adapter | `preprocessing_cuda.go` | Retains existing explicit CUDA32 execution for reference/comparison |
| Session, progress and publication boundary | `preprocessing.go` | Single caller; no retained Go pointers; returned Go data survives Close |
| Support/root persistence | `preprocessing_artifacts.go` | Trusted manifests, checksums, exact EOF and support-bound restart |

These files live in `encoding/liftrs-shared/` or `encoding/liftrs-cpp/` as their
names indicate. Native sparse plan construction in `prepare_plan.hpp` also serves
the existing wide transform ABI; it is not copied into the new executor. The
older GPU-assisted scaling harness uses the shared root/reciprocal procedure.
The scalar Go root arithmetic and ordinary `PrepareCyclicFFT` remain correctness
references. Neither old caches nor ordinary client backend selection changes.

Exact enumeration evaluates each ternary bit-weight vector once, then expands
it into the monomial indices. For m=2, each bit's weight is 0, 1 or 2, so there
are 3^s classes rather than q² separate carry problems. This is the paper's same
membership criterion, not approximate dimension estimation. At s17 the class
cache is about123 MiB, the membership bitmap2 GiB and the small index table0.5
MiB, before the final 8K-byte index array. Runtime improvement is unmeasured.

Root batches are sorted. Their first root is computed by exponentiation; later
roots advance by the exponent gap. Short-gap multiplication maps are reused.
For root a+b·u, the quadratic conjugate is (a+b)+b·u, so trace=b and
norm=a²+a·b+delta·b². This replaces separate inversion/conjugation per root
without changing factors. Independent scalar arithmetic checks this identity.

## Using and retaining the result

Build with `-tags liftrs_cpp` and cgo enabled. Ordinary builds remain CUDA-free;
CUDA preparation additionally needs `liftrs_cuda` and an explicitly supplied
compatible library. Backend and budgets are mandatory, independent of Spec:

```go
options := liftrs.PreprocessingOptions{
    Backend: "cpp", Workers: 4,
    Limits: liftrs.PreprocessingLimits{
        MaxElements: 65536, MaxHostArrayBytes: 64 << 20,
        NativeBytes: 64 << 20,
    },
}
p, err := liftrs.NewCodePreprocessor(liftrs.Spec{
    FieldDegree: 8, Dimension: 2, CurveDegree: 5,
}, options)
// Check err, defer p.Close(), then p.Discover(liftrs.OrbitOptions{}).
```

`Discover` defaults to cryptographic randomness and errors on an incomplete
bounded support union. Deterministic test readers are not a production
randomness recommendation. `CodeSupport.Write` streams its indices and returns
`SupportRecord`; publish payload and JSON manifest atomically, with source
revision/provenance. `ReadCodeSupport` accepts a trusted manifest and explicit
limits. A checksum protects byte identity; it does not prove exact K or code
membership. Do not accept an attacker-supplied manifest as certified discovery.

`Prepare(support, PreparationReuse{SaveRoot: callback})` can checkpoint normalized
P before reciprocal work. Persist support first so that its digest binds the
root checkpoint. The callback streams `RootCheckpoint.Write` and its `RootRecord`
to atomically published files; return an error if persistence fails. Restart with
`ReadPreparationRoot` and `PreparationReuse{Root: saved}`. Restart saves root
construction work, but still computes G and verifies the full identity. Discovery
sample count is provenance, not part of mathematical code identity.

`PreparedCode.CodeData()` yields canonical `ScalingCodeData` for existing client
backends. `WriteScalingCode` serializes it; `WriteGenerator` separately retains G.
P/G coefficients use four-byte base-field slots even at s17. Support coordinates
use four bytes through s16 and eight at s17. Spectrum values use the existing
canonical four/eight-byte width selected from s. No plaintext packing or masking
changes. See [code data](scaling-code-data.md) and
[representation conventions](data-representation.md).

## Resident GPU selection

Build the explicit server plugin with CUDA, leaving ordinary client builds alone:

```sh
nvcc -std=c++17 -O3 -lineinfo -arch=native -shared -Xcompiler=-fPIC,-pthread \
  encoding/liftrs-cuda/matrix.cu encoding/liftrs-cuda/preparation.cu \
  -o /path/to/libliftrs-cuda.so
```

Set `Backend: "cuda"`, `GPUPath` to the verified plugin, `GPUDevices` to the
visible device ordinals and `GPUBytes` to the cap **per selected GPU**. Keep
`Workers`/host limits explicit as before. An empty device list retains the legacy
CUDA32 adapter; it cannot execute s17. The resident engine selects uint32/uint64
from s, uses all-pairs CUDA peer access and rejects unavailable topology/caps.

For the bounded s8 example above, the GPU selection is:

```go
options.Backend = "cuda"
options.GPUPath = "/path/to/libliftrs-cuda.so"
options.GPUDevices = []int{0}
options.GPUBytes = 512 << 20 // per-GPU cap; this example uses one card
```

Contexts are process-scoped. Peer links are leased across live engines; Close
releases owned payloads and the last owned link lease without disabling borrowed
external mappings,
and the calling thread's original CUDA device is restored. Errors drain queued
work before borrowed Go pointers can be released.

Large products use three sharded resident arrays, never host backing. At s17
four cards model96 GiB each (384 GiB aggregate), with additional metadata and
headroom. This power-of-two implementation does not accept three GPUs, despite
three being an arithmetic capacity floor. GPU monomial classification and root
warp scans are also implemented; host orchestration and small base-polynomial
kernels reuse the existing shared code. See the report for communication and
host staging costs; speedup and full-size peak remain unmeasured.

## Budgets, ownership and measurement

Construction builds bounded metadata only. `MaxElements` must admit the complete
q² domain for discovery. `MaxHostArrayBytes` caps each principal Go array, **not
their sum or RSS**. `NativeBytes` caps explicitly accounted native payload and
metadata; allocator/thread/runtime overhead remains additional. Caller arrays,
Go garbage awaiting collection and filesystem cache also need aggregate headroom.
These API limits do not authorize scheduler allocations or full-size runs.

The CPU reference’s fast product keeps two operand arrays plus transpose scratch. At s17
that is3·8·2^34 bytes =384 GiB, excluding Go arrays and metadata. A single native product returns its owned base-field result directly, avoiding
a duplicate Go result/copy; overlap-add still accumulates when needed. Once products
are dead, operands are released before allocating the final Go spectrum; scratch
is reused without extra transforms or transfers. With the still-estimated s17 K,
support, P and G add tens of GiB each, and a full Go identity result adds64 GiB.
A host peak around600 GiB or higher is plausible, **not measured or bounded**.
The resident GPU prototype models384 GiB device arrays, partitioned over four
cards, plus host data and communication costs. Its full-size peak remains
unmeasured; source implementation and capacity alone do not admit a full run.

No slower memory-saving mode is selected implicitly. Any future streaming,
recomputation or disk trade-off must report memory, runtime, transfers and storage
and remain explicit when the faster path fits. The CPU reference route's speed is
unknown; it is not asserted faster than measured GPU-assisted s16 preparation.

Use `Progress` for synchronous stage callbacks (do not reenter the session), or
`LIFTRS_CPU_PROGRESS=1` for flushed JSON stage/native progress logs. Counters name
work within a phase; a heartbeat alone is not an ETA. Discovery, reusable
preparation, persistence and database encoding must be timed separately. Retain
CPU profiles and matching binaries during optimization. The bounded runner keeps
small reusable caches. The separately admitted full s17 trial has its own
[resource manifest](../experiments/gpu-s17-full-preprocessing/budget.json); its full mathematical preprocessing artifacts now pass75832.

## Experimental local-vector scheduling

The resident CUDA plugin snapshots `LIFTRS_GPU_PREPARATION_VECTOR_LAYOUT` at
engine creation: `peer` (default) or `local`; other values are rejected. The
bounded experiment exposes this as `run.py --vector-layout local`. This temporary
experiment selector does not change the code specification, public data format
or ordinary CPU/client builds. Do not mutate this process-wide variable while
constructing engines concurrently. Tiny shapes with fewer vectors than GPUs
retain the peer schedule because complete vectors cannot be partitioned evenly.

Local mode gathers complete vectors onto their owner GPU, transposes each slab
for coalesced local access, runs the existing recurrence with its original global
vector offset, and restores canonical slabs. Both operands reuse the same scratch;
no additional full-domain device/host array or cache is introduced. Taylor passes
remain peer-aware and retain their cross-device synchronization. Whole operations
return canonical results; no external layout or arithmetic changes.

The initial schedule adds eight transposes per FFT: four local and four peer
gather/restore passes,16N word reads/writes (~2 TiB at fulls17, excluding existing
passes). Actual link traffic is a subset of that array traffic. Host/retained
storage requirements are unchanged; allocation headroom and full-size setup/latency cost
remain unmeasured. Bounded local-mode correctness and CUDA memcheck pass74468;
a32 MiB tile shows18.9%/9.7% lower two-card forward/inverse latency, with single
samples and little two-card versus one-card gain. This is an explicit performance candidate, not a lower-memory
default. Compare independent oracles, memcheck, dense tile hashes and same-device
latencies before selecting it. [Reproduction/evidence](../experiments/gpu-s17-preprocessing/report.md#local-vector-comparison-scaling-074).

Repeated256 MiB calibration now passes on1/2/4H200 (job74635), including four-card
small complete preparation, exact tile hashes/inverses and zero CUDA errors/leaks.
Four-card local medians0.778 s/0.754 s compare with one-card peer0.866 s/0.755 s;
this modest component gain does not establish full-preparation latency. Owned
payload is equal across layouts. Source launch geometry yields few blocks at
this tile size, a possible scaling limit rather than measured occupancy evidence.
Keep peer default; finish persistent full-workflow/resource admission separately.
See the [completed report](../experiments/gpu-s17-preprocessing/report.md#completed-four-gpu-calibration-scaling-077).

## Persistent server command

[`cmd/pir-preprocess`](../cmd/pir-preprocess/README.md) wraps the existing API with
explicit backend/specification/resource flags, atomic stage directories and trusted
source/digest manifests. It retains CPU profiles and advancing byte/native phase
logs. Completed-cache restart verifies all files without GPU setup or a resident
code copy; root restart skips discovery/root construction but rechecks the full
identity. Fresh stages are streamed, synced and read back once; storage/readback
cost is included rather than described as free. Failed partial stages are retained.

[Bounded complete-workflow validation](../experiments/gpu-preprocessing-workflow/report.md)
checks publication failure/corruption, independent Go bytes, fresh and root/final
restart across processes,1/4GPU peer/local and CUDA-free readback. Local native
checks and complete GPU/race/sanitizer validation pass75342. The CLI accepts
s2..17 with explicit array/code caps; those flags alone do not approve allocations.
Fulls17 cache now passes75832,235.324GiB canonical files with348.393GiB sampled
peak hostRSS. The fixed small validation controller keeps its approved2GiB cache
guard; the separately admitted full controller produced that cache.

User authorization includes considering8/16GPUs. Current same-node peer pointers
do not implement cross-node execution; measured nodes have four GPUs. Larger
groups need explicit distributed communication, host layout/headroom and end-to-end
measurements. Retain this distinction in admission, rather than treating a GPU
count option or arithmetic memory model as implementation support.

Persistent workflow job75194 passes one-GPU fresh/root/final restart and CUDA-free
readback, but fails during the first four-GPU discovery transpose on gnode03;
Slurm reports GPUs requiring reset. Cause remains unproven; later local cases and
CUDA memcheck are unrun. The retry adds a small runtime-copy/transpose preflight
and bounds shutdown waits, retaining all failure evidence and resource caps.
It targets the previously validated gnode02; a pass would not establish node03
health. No fulls17 admission. [Failure review/reproduction](../experiments/gpu-preprocessing-workflow/report.md#job75194-failure-and-bounded-retry-scaling-079).

Persistent workflow validation now passes job75342 ongnode04: runtime peer-copy/
transpose preflight, native/race/ASan/vet, all peer/local1/4GPU fresh/root/final
restart cases, independent Go canonical bytes and CUDA-free readback, plus zero
CUDA errors/leaks. Source/evidence/cache checks reconcile509/37/73 identities.
That bounded complete-runner gate preceded count/admission and the successful
fulls17 cache75832. Gnode03's earlier reset cause stays unresolved.
[Exact results](../experiments/gpu-preprocessing-workflow/report.md#completed-persistent-workflow-validation-scaling-080).

## Count-only admission

`CountCodeMonomials(spec, options)` returns exact K from the requested native or
resident CUDA executor without an N-element host grid or K-element index list.
It privately constructs/closes metadata and enumeration state, reusing the same
count routine as normal discovery; it never exposes a session with relaxed array
admission. `NewCodePreprocessor` retains its existing full-grid limits. Count
still obeys MaxElements, metadata/native and GPU byte caps. GPU counting requires
explicit resident devices and never silently uses CPU.

For example, a count-only GPU call at s17 may use MaxElements=1<<34,
MaxHostArrayBytes=8<<20, NativeBytes=64<<20 and GPUBytes=1<<30 on four devices;
ordinary s17 preprocessing rejects that host-array budget. This has~2.5 GiB
modeled count-device state, distinct from384 GiB product arrays or actual RSS.
`pir-preprocess --dimension-only` publishes a separate trusted dimension record
and profile; no support/root/code payloads or discovery sampling occur. The
record cannot certify support or skip full discovery/preparation. The device
bitmap is discarded on Close instead of adding a2 GiB host/file copy; full
discovery later re-enumerates it. [Validation, provenance and exact size models](../experiments/s17-preprocessing-admission/report.md).

Count admission75810 passes exhaustive s17 K=3876566203 with s8 Go/s16 cache
references, complete small workflows, race/ASan/vet and zero CUDA errors/leaks.
This is exact dimension, not sampled support or full preparation. Successful
s17 files calculate to235.324 GiB. Discovery's conservative principal model is
189.765 GiB (including union and validation bitmaps); final retained host arrays
220.883 GiB, plus unmeasured temporaries/GC/runtime/cache. The [guarded full plan](../experiments/gpu-s17-full-preprocessing/report.md)
is now approved and completed75832, with sampled peak hostRSS348.393GiB. Long host support ordering/
Frobenius, normalization and root-field scans now report advancing65536-element
chunks; arithmetic/check order and existing buffers remain unchanged. Independent
Go/native results and cross-chunk progress/error tests pass. Instrumentation
cost is included/unquantified, no new full data pass or slower memory mode.

## Incremental recovery

`pir-preprocess --checkpoint-interval 2h` enables optional partial-work recovery;
`--resume` loads the newest intact matching generation (falling back to the
previous one if damaged). Without this option there are no rolling checkpoint
writes, and completed support/root/code reuse remains unchanged. Backend choice
and code parameters do not change the recovery procedure.

The library exposes `PreprocessingOptions.Recovery`: supply loaded
`Discovery`/`Preparation` checkpoints and a synchronous `Save` callback. Stream
`checkpoint.Write(writer)` and atomically publish its returned `ProgressRecord`
with the bytes; restore using `ReadPreprocessingCheckpoint`. The callback borrows
arrays only until it returns; never retain or mutate them. A callback error stops
at that boundary. Persist support first, as for normalized-root reuse.

| Save boundary | Persisted work | Work repeated on restart |
|---|---|---|
| Completed random sample/interpolation/support union | N-bit union, completed sample count and found count | Exact monomial enumeration, then any unfinished sample; previous random draws need not be replayed |
| Completed root batch | Processed support offset, pending factors and occupied product-tree levels | Source validation/degree scan, unfinished batch and any later merges |
| Completed Newton reciprocal iteration | Completed G prefix, bound to the same support as the normalized root | Unfinished iteration, full independent P·G identity and final spectrum |

A snapshot is partial computation, not certified code data. Digests and source
bindings protect a trusted local cache against damage/mixups; they do not certify
arbitrary untrusted input. Final exact-dimension/support closure, full P·G identity
and canonical-file checks remain mandatory. Empty carry levels and pending leaves
are retained without forcing extra merges or transforms. No128 GiB FFT grid,
CUDA context or in-flight native work is serialized.

The CLI writes/hashes snapshots once through64 KiB buffers, fsyncs the payload and
manifest, then publishes the directory atomically. It keeps the latest two complete
generations **per phase**; only after successful publication does it prune older
ones. Failed unpublished recovery writes are removed because they cannot resume
and would exhaust the fixed storage budget. Ordinary failed canonical-stage
artifacts, all profiles and logs remain retained.

Periodic writes target about1% cumulative wall-time overhead. The first eligible
save is after the requested interval; expensive measured saves stretch later
intervals. A write exceeding its current allowance is abandoned between buffered
writes. Filesystem calls/fsync can block, so this is a target, not a guaranteed
bound. Full-run root snapshot writes now measure78.525s total across four saves;
large reciprocal-snapshot cost remains unmeasured. Logs include write byte counters,
phase timestamps and completed/failed recovery events; checkpoint time is included
in profiled workflow latency.

`SIGUSR1` asks the CLI to save and exit at its next recoverable boundary (exit75,
workflow event `checkpointed`), bypassing the periodic write allowance. A completed
normalized-root save is also a safe stopping point. A request arriving during the
last spectrum can save the completed G afterward. The full controller requests
this one hour before the earliest internal deadline, retaining hard TERM/KILL
and resource guards. SIGTERM is still an emergency stop, not a guarantee of a new
checkpoint. A long transform or blocked write may outlast the warning; only the
last completed snapshot is then recoverable. No supported save point exists inside
the final spectrum or an unfinished canonical-file publication.

At s17, snapshots model up to2 GiB for discovery, about14.44 GiB for the root
frontier (plus a bounded leaf buffer), and49.56 GiB for full G. Retaining two of
each adds about132 GiB to235.324 GiB canonical payload, within the approved512 GiB
retained-files guard before logs/profiles/partial-stage headroom. These are size
projections, not measured filesystem/RSS costs. Writes introduce no new transforms,
GPU transfers or grid copies. Restore may retain another2 GiB discovery bitmap;
normalizing an aliased restored root level can require a14.44 GiB copy, and the
root-frontier degree check scans the processed support once. Such restart costs
are separate from ordinary uninterrupted execution.

Small native interruption/resume, independent Go comparison, corruption/source
rejection, two-generation retention and fallback checks are recorded in the
[recovery validation report](../experiments/preprocessing-recovery/report.md).

## Full s17 result

Job75832 verifies the complete(s,m,t)=(17,2,5) public cache withK3876566203,
N17179869184. Four H200s,32 active workers/128 allocated cores;348.393GiB sampled
peak hostRSS and235.324GiB canonical payload. Resumed worker9h58min38s includes
persistence/readback and reuses certified support from75819 (discovery21min38s,
two samples). Construction/identity/spectrum9h16min46s includes checkpoint I/O.
The root product takes8h33min57s; the phase includes host and GPU work, not an
isolated GPU-kernel measurement. Native CPU-only full preprocessing remains
unmeasured. [Full evidence and boundaries](../experiments/gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088).

Root snapshot writes total78.525s (0.219% of worker wall); latest two occupy
23.426GiB, besides canonical data. Full-size snapshots were not used to resume
this successful attempt. Producer readback/hash checks pass; post-review verifies
seals/bindings/sizes without reading235GiB again. Full resident client loading,
CPU/GPU database encoding and packed persistence are separate validations.

## Explicit bounded-host canonical export

The resident `Prepare` API remains the default. For an explicit memory-saving
GPU server path, `PreparePolynomials` performs the same root/reciprocal/full P*G
verification and returns immutable P/G without a spectrum. `WriteCode` computes
and writes canonical header/P/spectrum through bounded GPU download tiles;
`WriteGenerator` preserves G. No caller-owned support/checkpoint is mutated.
Returned polynomials survive the producing session; a compatible resident GPU
session is required for `WriteCode`. No CPU or host-backed fallback is selected.

Inside a caller that already owns admitted sessions, certified support and fresh
atomic staging writers:

```go
polynomials, err := session.PreparePolynomials(support, reuse)
if err != nil { return err }
// A caller may now release its support/recovery references; keep persisted files.
if err = session.WriteCode(codeWriter, polynomials, 1<<20); err != nil { return err }
if err = polynomials.WriteGenerator(generatorWriter); err != nil { return err }
// Hash, fsync, verify and atomically publish both files only after success.
```

The CLI provides that persistence workflow with `--spectrum-tile-bytes 1048576`.
Omitted/zero retains resident mode; the explicit option requires CUDA and8 bytes
through64 MiB (rounded down to8-byte alignment). Selected size is logged with
command metadata; canonical bytes and strict manifest schema remain compatible.
The CLI releases its support/recovery references after P/G verification and final
P/G references after writing, before bounded readback. It preserves checkpoint
files and save callbacks; GC/RSS reclamation is not guaranteed by reference release.

Source-derived final s18 export arrays project256.001 GiB with1 MiB tile versus
889.446 GiB resident; this is not a peak measurement. Discovery, Newton/product
intermediates, device arrays, runtime and filesystem cache remain additional.
No extra transform or canonical output bytes are introduced, but chunk I/O and
GC timing can change runtime. Full preprocessing still stops at s17. This export
integration **passes complete bounded GPU validation76710**, including independent
canonical/G bytes, publication/restart/corruption checks and API/CLI memcheck;
the lower-level s17-width primitive passed76704. Full-size savings/costs remain
unmeasured. See [reproduction, tests and tradeoffs](../experiments/preprocessing-export/report.md).

## Internal root boundary for distributed prototyping

Shared preparation constructs and verifies normalized P before entering the
reciprocal/full-identity procedure. The generator helper takes P, K and N instead
of support indices; arithmetic, saved-root immutability, recovery binding and
canonical output are unchanged. A failed root-publication callback prevents
reciprocal work. Frobenius-closed subsets can use the same root builder and their
normalized products merge to P; this is tested without introducing a distributed
public API. [Bounded backend validation](../experiments/preprocessing-root-boundary/report.md)
passed in job77126. Caller support remains borrowed; this boundary alone does not force
GC, consume caller data or establish lower measured RSS. Server-only transport
experiments add no dependency to CPU-client builds.

The private `partitionRootSupport` helper keeps Frobenius partners together and
balances orbit degrees across sorted immutable copies. It adds one support-index
copy and is not a full-size distributed support representation. The optional
[independent-worker experiment](../experiments/parallel-root-workers/report.md)
reuses normalized-root/product helpers with separate executors and a joined merge;
it introduces no public scheduler or default parallel mode. Performance validation
is pending, and no end-to-end speedup follows from subset equivalence alone.

## Explicit subset membership experiment

`validateExponentSupportWith` shares cardinality, range, strict order and full
Frobenius membership checks across bitmap and sorted-search storage. Binary search
borrows the immutable ordered input, avoiding an N-sized bitmap with O(KlogK)
lookup work. The ordinary wrappers for discovery/cache/root preparation and
partitioning keep bitmap validation. Experimental root/partition helpers accept
an explicit choice; no trusted-validation bypass or new public API is introduced.
The optional test binary flags `-root-validation` and `-root-reference-validation`
select measurement strategies only. [Comparison and memory/time boundaries](../experiments/root-subset-validation/report.md)
are documented. Matched job77213 passes and measures substantial sparse-s17
worker RSS/latency reductions, while smaller cases show search overhead. The
explicit option is recommended for those sparse worker workloads; ordinary
full-support/cache/discovery defaults remain bitmap. No global default or
full-size speed/memory claim follows.

## Buffered persistence experiment

The [bounded persistence gate](../experiments/preprocessing-persistence/report.md)
reuses CLI checksum/fsync/atomic publication and restart readback with optional
4 MiB coalescing and controlled ownership of separate files. This is a server I/O
experiment; ordinary codecs/CLI defaults remain unchanged. Buffering must flush
successfully before fsync and publication, and synthetic probe payloads are not
usable preprocessing artifacts. Bounded HPC77352 passed (gnode02,2CPU/4GiB RAM,
0GPU): write calls1,536→24 per96MiB trial, but elapsed benefits are mixed and full
readback dominates. Persistence-phase max sampled cgroup charge362.695MiB;
all atomic/restart/corruption/concurrent race checks pass. Retain existing defaults
while integrating scalable rank ownership and publication. No full-size output
latency or power-failure guarantee follows from this probe.

## Experimental rank-owned root reduction

The [root merge-tree gate](../experiments/root-merge-tree/report.md) reuses the
canonical root/product kernels with bounded support routing and unique subtree
ownership. MPI is a separate explicit server-test feature; clients remain CUDA/
MPI-free. Only the coordinator holds the complete validated source/reference;
workers keep local orbit-closed shards. The shared schedule supports odd/empty
owners, with validation snapshots independent of compute/message inputs.
Four logical ranks on two physical H200s validate two levels; bounded job77704
is submitted, with measurements pending. This does not implement distributed discovery, large-product collective
transforms, resumable tree checkpoints or fulls18 preprocessing. Existing atomic
publication/verification helpers remain required for subsequent pipeline integration.
