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
passes job74466, including peer-link lifetime sanitizer checks. No full
s17 discovery, preparation, cache, latency or admission exists. The legacy CUDA32
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
small reusable caches, but no full s17 artifact or larger allocation is admitted.
