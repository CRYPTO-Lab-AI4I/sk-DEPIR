# Shared Lift-RS encoding

This package (Go name `liftrs`) supplies the public API used by the Go, C++ and optional CUDA execution
paths: code specifications, field arithmetic, discovery, reusable preprocessing,
polynomial multiplication strategy and codeword assembly. The dense oracle and
reference tests also live here. `field/` holds common scalar arithmetic used by
preprocessing, orchestration and the Go transform backend; native kernels perform
identical field operations in C++.

Go-only transform execution is in [liftrs-go](../liftrs-go/README.md); native
transforms and their binding are in [liftrs-cpp](../liftrs-cpp/README.md). The
shared adapter chooses whole-buffer native calls explicitly. Code preparation
stays Go. See the [overview and run commands](../README.md) and the examples below.
The import is `github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared`;
`encoding/liftrs-go` no longer contains the complete encoder API.

Experimental scaling preparation uses a private base-field lookup table for
public code coefficients (less than 768 KiB through s16), released before online
use. General field APIs remain table-free and artifact coordinates are unchanged.
The [arithmetic experiment](../../experiments/preparation-arithmetic/report.md)
records kernel and complete smaller-code checks; the later
[full s16 preparation](../../experiments/gpu-preparation/report.md) and
[GPU encoding](../../experiments/cuda-s16-encoding/report.md) checkpoints are separate.
The [subsequent CPU comparison](../../experiments/preparation-kernels/report.md)
also removes per-coefficient field dispatch inside private polynomial leaves,
with exact prefix checks and complete smaller-code regressions. It leaves the
general field path and global native-transform cutoff unchanged.


The Go reference implements code preprocessing and systematic encoding for
small curve-lifted Reed-Solomon codes with locality `q-1`. It takes field elements
directly; the application wrapper owns partitioning, padding, and bit conversion.
There is no permutation, masking, storage access, or secret state here. The
separate [cryptographic reference](../cryptoref/README.md) can consume its codewords.

```go
code, err := liftrs.Preprocess(liftrs.Spec{
    FieldDegree: 4, // GF(16), polynomial X^4 + X + 1
    Dimension:   2,
    CurveDegree: 2,
})
if err != nil {
    return err
}
message := make([]uint32, code.Dimension()) // exactly K field elements
// Fill message with integers in [0,16).
word, err := code.Encode(message)
if err != nil {
    return err
}
positions := code.SystematicPositions()
// word[positions[j]] == message[j]. A PIR query for message j targets positions[j].
```

Import `github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared`. The runnable
[example](example_test.go) demonstrates the complete preprocessing/encoding call
sequence.

## Monomial membership and rate estimation

`GoodMonomial(spec, exponents)` checks whether a reduced monomial belongs to the
code. It supports `s=2..20`, `m=2..3`, and `1<=t<q-1` without field arithmetic or
code preprocessing. The [example](example_test.go) tests a monomial in a target-size
code; this does not make that code available to `Preprocess`.

The [implementation](monomial.go) rewrites the paper's reachability condition using
binary carries. For each possible positive multiple `j*(q-1)`, track the interval
of reachable carries from initial carry `j` to final carry `j`. This uses
`O(s*m*t)` time, constant auxiliary space, and no heap allocations. The original
residue DP remains in reference preprocessing as an independent check.
[Tests](monomial_test.go) compare them exhaustively on small codes and on randomized
larger fixtures; the original tests additionally check membership against actual curves.

The [standalone estimator](../../cmd/pir-rate/README.md) samples uniform exponent
tuples. See the [experiment report](../../experiments/liftrs-rates/report.md) for
all 13 target rates, confidence intervals, and reproducibility details.

## Supported reference instances

| Field degree | Polynomial bitmask | Allowed evaluation dimensions |
|---|---|---|
| 2 | `0x7` (`X^2+X+1`) | 2, 3 |
| 3 | `0xb` (`X^3+X+1`) | 2 |
| 4 | `0x13` (`X^4+X+1`) | 2 |

Accept `1 <= t < q-1` and `N=q^m <= 256`. These small fields are test fixtures,
not members of the initial security catalogue. Larger inputs are rejected before
matrix allocation. The dense matrix has `K*N` uint32 entries, bounded by 256 KiB
of element storage, plus slice headers and temporary data. This backend provides
no evidence of resource feasibility at the target catalogue sizes.

Preprocessing classifies exponent tuples using the positive-residue reachability
criterion, evaluates the good monomials, and applies deterministic Gauss-Jordan
elimination. The first available pivot columns give the systematic positions.
Coordinates and exponents use `x_0 + q*x_1 + ...` order. Encoding multiplies the
message by this systematic matrix. See [the contract](../../docs/encoder-contract.md#small-reference-encoder)
and [paper, Section 4.2.1](../../paper.pdf#page=25) for the code definition and
membership criterion. This dense backend does not implement the paper's fast
encoder from Section 4.2.4.

The preprocessed `Code` can be reused concurrently. It copies exported position and
monomial slices; encoding neither changes the message nor shares output buffers.
A zero/uninitialized `Code`, incorrect message lengths, and out-of-field values
produce errors.

## Verification

From the repository root:

```sh
go test ./encoding/liftrs-shared
go test ./pir -run '^TestReferenceLiftedRSEncoderOnlineRecovery$'
go test -race ./encoding/liftrs-shared
```

[Reference tests](reference_test.go) check field arithmetic with an independent
polynomial oracle, monomial membership against all tiny curves, an explicit
GF(4) golden vector, every message and curve through every grid point for a tiny
code, all supported configurations on deterministic recovery fixtures, input
errors, artifact ownership, and concurrent reuse. Exhaustive tests include repeated
curve positions and codewords beyond the corresponding low-degree RM code.

The [online integration test](../../pir/liftrs_reference_test.go) uses the existing
client, permutation/masking helpers, and plain in-memory reads to retrieve every
systematic position of a genuine GF(16) lifted codeword. It does not qualify the
prototype cryptographic components for production. The existing file-based
`LiftedRS.Encode` still calls its clearly marked RM placeholder.

## Paper and cyclic encoders

The separate `OrbitBasis` API implements the paper's omega-power information set
and two equivalent encoding algorithms for `m=2`, `s=2..13`, `1<=t<q-1`, with
`N<=67,108,864`. This extends the comparison backend, not the dense `Preprocess` API
or the legacy file-based PIR encoder. Output remains unmasked canonical-grid
polynomial-basis field elements. See the [algorithm review](../../docs/encoding-algorithms.md)
and [initial table-backed comparison](../../experiments/encoder-comparison/report.md).
The [table-free checkpoint](../../experiments/tablefree-encoding/report.md) measures
the earlier implementation through s=9, including packed output and separate
preparation processes.
The [s=10 checkpoint](../../experiments/encoder-scaling/report.md) extends exact
encoding to N=1,048,576 and K=103556 for t=6, retaining JSON artifacts. The complete
artifact loader limit is 64 MiB because the paper solver occupies 35.74 MiB;
this input-size cap is separate from any publication or working-memory budget.
[Progressive cyclic measurements](../../experiments/cyclic-scaling/report.md) now
cover s=11,t=6 and s=12,t=6 with exact K=455434 and K=1965739. They reuse prepared
constants in memory; existing basis/complete JSON loader caps remain 2/64 MiB and
may reject larger serialized caches. No larger paper comparison is claimed.
The [s=13 cyclic benchmark](../../experiments/cyclic-s13/report.md) adds exact
K=11926435,N=67108864 at t=5, with three encodings and final-output systematic,
independent-curve and packed-readback checks. First use is 33.94 min, preparation
29.42 min and median encoding 185.15 s, with 5.34 GiB whole-case peak RSS.
It uses a two-hour case timeout and the same memory guard. That unprofiled run
did not measure large-cache JSON loading; the later profiling run below does.

### Reusable preprocessing and database encoding

See the [runnable example](orbit_example_test.go), which checks errors at each step:

```go
basis, err := liftrs.DiscoverOrbit(spec, liftrs.OrbitOptions{})
// Handle err, then prepare once (or load a published complete artifact).
encoder, err := basis.PrepareCyclicFFT() // alternatively PreparePaper / PrepareCyclic
artifact, err := encoder.MarshalBinary() // public, database-independent constants
// Handle errors and persist artifact with its generation provenance.

encoder, err = liftrs.LoadCyclicEncoder(artifact) // once per process/code
// Handle err, then reuse encoder for distinct databases of exactly K field elements.
word, err := encoder.Encode(message)
```

The three stages have different costs:

1. **Discovery:** exact good-monomial enumeration and random support discovery;
   complete only when the observed degree set has the exact dimension K.
2. **Preparation:** method-specific constants, independent of any database.
   The paper solver caches its product tree, reversed-divisor reciprocals and
   denominator inverses. Cyclic encoders cache P/G; the FFT variant also caches
   the fixed generator's spectrum. These constants are included in their artifacts.
3. **Encoding:** receives a database's K field elements. No discovery, product-tree
   construction, denominator inversion or fixed-generator transform occurs here.

Loading complete artifacts parses the constants and reconstructs field/transform
runtime tables, without rerunning discovery or polynomial preparation. Count that
load cost once per process/code. `OrbitBasis.MarshalBinary` is a smaller discovery
artifact; loading that alone still requires `PreparePaper` or `PrepareCyclic`.
All initialized encoders are immutable and can encode concurrently; each call
owns its message-dependent buffers and output.

Artifact version 1 fixes the documented field polynomials, tower and generator
selection; incompatible changes require a new version. Artifacts use JSON and contain
no database data, masking/permutation keys,
or randomness needed for PIR queries. They can be published on GitHub with their
exact code specifications and reproducible generation evidence. The experiment
includes [example public artifacts](../../experiments/encoder-comparison/artifacts/).
Load only independently verified artifacts: loaders check shapes, bounds and field
ranges, not mathematical provenance. Basis artifacts are limited to 2 MiB and
complete encoder artifacts to 64 MiB in this bounded backend.

### Algorithms, coordinates and limitations

For the paper path, set `beta_i=omega^i` for support exponents i. A product-tree
transposed-Vandermonde solve recovers the univariate coefficients; an additive FFT
evaluates the full extension field. The cyclic path computes `H=(P*M) mod X^K`
and `Y=G*H`, where `P*G=1+X^(N-1)`. Polynomial products now use a hybrid
[additive-FFT/Karatsuba backend](orbit_product.go); `PrepareCyclicFFT` additionally
caches the fixed generator spectrum for the large `G*H` product. Both paths may
use extension-field arithmetic internally, even for base-field coefficients.

The extension uses `u^2=u+delta`, with delta the first polynomial-basis element
of absolute trace one. The deterministic primitive generator is the first
primitive packed extension element at least q. `phi(a+b*u)=(a,b)`, so the packed
integer is also the canonical coordinate `a+q*b`. For message j, use
`basis.SystematicPosition(j)` to compute `phi(omega^(j+1))` without storing a
K-entry coordinate table. No discrete logarithm is required. Both algorithms
return the same word in canonical coordinate order, before permutation/masking.

The FFT uses general binary subspace polynomials and costs `O(N log² N)` field
operations, including monomial-basis conversion. The product backend uses
power-of-two subspaces sized for the full convolution, including when only a
prefix is returned. Outputs beyond the available field domain use overlap-add
chunks. Small products use Karatsuba, low products prune irrelevant recursive
branches, and very unequal operands are split into balanced blocks. FFT work
is `O(n log² n)` per padded local product; this is not a specialized truncated FFT
or the asymptotically best known multiplier. Immutable transform data are reused;
all database-dependent scratch is per call and excluded from public artifacts.
The [comparison](../../experiments/fast-polynomial/report.md) records thresholds,
correctness, allocation and whole-encoder effects. The backend remains in memory;
there is no external-storage, streaming or constant-time implementation in this
default path. The separate [GPU matrix experiment](../../experiments/gpu-matrix/report.md)
validates complete transforms, resident convolution and GPU-assisted s10
preprocessing under `liftrs_cuda`; it does not change ordinary constructors or
move the complete online encoder to GPU.
Its [resident product adapter](../../experiments/gpu-products/report.md) optionally
keeps both transforms, pointwise products and inverse on device for each local
convolution, returning only the required prefix. The existing overlap-add and
single-caller workspace contract remain in the shared layer. Default executors
keep the CPU path; the optional private hook accepts immutable inputs and checks
the returned prefix length. CUDA support remains experimental and bounded.
The accepted s<=13 range supports resource-limited experiments, not a guarantee
that every code or serialized cache fits RAM or loader limits. Completed catalogue
benchmarks cover the public cyclic path through s=13. The experimental
[s=14 first-use path](../../experiments/parallel-s14/report.md) additionally
validates exact K=50,976,325 and cached native encoding in about 91 s, with
separate preparation and fresh-process memory measurements. s≥16 and automatic
application-plan integration remain unverified.
Do not extrapolate the small comparison's speed or artifact sizes to those codes.

[Tests](orbit_test.go) check transforms against Horner evaluation, polynomial
operations against schoolbook multiplication, complete message bases against the
independent dense oracle, prepared-artifact reuse, malformed inputs and concurrency.
The oracle uses different systematic coordinates: compare codewords after matching
information-set values, not by reusing the same raw message across both conventions.

## Table-free field arithmetic

`NewBinaryField(s, polynomial)` validates an irreducible polynomial and provides
polynomial-basis XOR addition, multiplication/reduction, powers, inverses and
absolute trace for s=2..20. `BinaryPolynomial(s)` returns the fixed conventional
identifier, preserving existing prototype choices. `NewQuadraticField(base)`
constructs u²=u+delta using the first trace-one base element; a `QuadraticElement`
is the pair `A+B*u`, with two uint32 components even at s=20. See the
[runnable example](binary_field_example_test.go).

Arithmetic contains no log/exp table or cardinality-sized allocation. Its field
state is fixed-size; multiplication is portable shift/XOR reduction. Operand
addition/multiplication/power methods require canonical inputs; inverse validates
its operand and rejects zero. These methods are not constant-time cryptographic
primitives. The public arithmetic does not imply full validated encoder support at s=14..20.

The bounded orbit adapter still uses uint32-packed extension coefficients and
in-memory transforms, restricted to s<=13. It preserves version-1 artifact bytes,
polynomials and primitive-generator selection. Discovery now samples four bytes
per monomial coefficient for s=9, masking to 9 bits; a single byte would not yield
uniform coefficients. The [tests](binary_field_test.go) independently check all
supported field degrees, historical codeword hashes and sampled s=9 curve constraints.
The legacy `utils.GF2n` table-based backend and its field coverage are unchanged.


The bounded transforms additionally use small reusable
[fixed-constant tables](orbit_constant.go), capped at 256 KiB of row entries per
transform. These accelerate repeated level constants without changing the public
table-free field API, coordinate conventions or serialized artifacts. They are
rebuilt during discovery/loading and reused across encodings; changing offsets
still use general multiplication. Table indexing is not constant-time. See the
[portable multiplication experiment](../../experiments/field-multiplication/report.md).


### Retained local caches

`LoadCyclicEncoderWithLimit(data, maxBytes)` loads the same trusted version-1 JSON
with an explicit size cap for both the complete cache and its embedded basis.
The default loaders keep their 64 MiB complete / 2 MiB basis limits. Bound the
file read before calling this API (see the [runnable example](orbit_cache_example_test.go))
and budget decoded-memory use separately;
structural checks do not authenticate mathematical contents. Large local caches
are distinct from public artifact distribution. The
[s=13 profiling runner](../../experiments/encoder-profile-s13/report.md) retains
cache checksums/provenance under ignored `.cache/liftrs/` and can profile encoding
again with `--reuse-cache`, without repeating discovery/preparation.

The completed profile checkpoint retains a 432 MiB s=13 cache and reloads it in
4.83 s. Encoding from the reloaded cache takes 186.35 s under CPU sampling and
matches the complete unprofiled output. See the report for separate setup/reuse
runs, resource observations and raw CPU profiles.


## Parallel language comparison

This directory contains shared Go orchestration (package `liftrs`), used by both
backends. The full encoder import formerly named `encoding/liftrs-go` moved here.
Pure-Go transform execution is separate in `../liftrs-go/`. Parallel experimental
[C++ kernels](../liftrs-cpp/README.md) originally covered fixed products and additive
transforms; the opt-in matrix path now also provides complete native online
execution, described below. The historical
[comparison report](../../experiments/kernel-languages/report.md) records shared
fixtures, validation, language/compiler boundaries and measured stage gains.

## Optional C++ transforms

`CyclicEncoder.WithCPPTransforms` creates an experimental adapter for the same
codeword encoder. Build with `-tags liftrs_cpp`, cgo and a C++17 compiler; default
builds return `ErrCPPUnavailable` and retain the Go path. See the
[runnable example](cpp_example_test.go) and [ownership/build details](../liftrs-cpp/README.md).
Go preprocessing, cached generator spectra and all polynomial optimizations are
shared by both paths. `EncodeWithStats` exposes per-encoding boundary accounting;
`Close` releases native code metadata after active encodings finish.
The native plan importer accepts field degrees 2–16 (s=16 has small-buffer
validation only). The original orbit/artifact
API remains bounded to s≤13; the opt-in minimal-artifact matrix path below adds
validated s=14 execution.

## Scaling experiments (opt-in)

The [step-3 prototypes](../../experiments/scaling-prototypes/report.md) add
`PrepareCyclicScalingPrototype`, which uses sparse transform setup, root-only
Frobenius-paired multiplication and reciprocal-based G construction. It verifies
P·G and preserves the original information set and artifact format. Default
preparation remains unchanged. This is not a bounded-memory setup implementation;
roots/factors, division/products and discovery still need resource work.

`ScalingPlan` constructs minimal online P/spectrum state, omitting G/support.
`NewWorkspace` provides a reusable arena, aliasing prefix work with the final
convolution buffer; the final canonical output remains separately allocated.
Use a separate workspace per concurrent call. `ResidentBytes` and `Bytes` count
array payload, not RAM. The [example](scaling_example_test.go) shows the Go path.
The initial native prototype used an internal workspace; the opt-in public
matrix workspace is described below. No supported streaming backend exists.
High-rate cases
whose full prefix convolution exceeds the transform domain are rejected by this
prototype; the ordinary encoder retains overlap-add support.

The minimal binary loader and native matrix path are now exposed experimentally
below. Their file layout is not a public artifact format and requires an
independently trusted digest/code provenance. External bucket routing remains
test-only: it uses scalar disk calls and Go tile kernels and tests
exact decomposition, not high-throughput I/O or a full native external encoder.

## Experimental parallel matrix encoder

With `-tags liftrs_cpp`, a verified minimal `ScalingOnlinePlan` can prepare
`WithCPPMatrix(workers)`. Allocate one `encoder.NewWorkspace()` per concurrent
encoding and call `encoder.Encode(message, workspace)` (or `EncodeWithStats`).
The [runnable example](scaling_matrix_cpp_example_test.go) demonstrates this API.
Close the encoder after its calls finish. Code P/spectrum is shared immutably;
transpose scratch initially adds `4*N` bytes to the workspace. Complete native
encoding reuses it for final output and transfers the slice to the caller; the
workspace allocates fresh scratch on its next complete or hybrid call. Earlier
outputs remain valid, and `Bytes()` excludes transferred storage. Keeping old
outputs still costs their full size. The Go/hybrid paths retain separate outputs.
See [CPU memory validation](../../experiments/cpu-buffer-reuse/report.md) for the
pending measured memory/runtime comparison. This opt-in path
preserves the canonical codeword and information set. `Encode`/`EncodeWithStats`
now run buffer setup, five transforms, pointwise products and assembly in one
native call, without duplicating large preprocessing arrays. Preprocessing and
loading remain Go. `EncodeHybridWithStats` preserves the previous Go scalar/
assembly path for comparisons. The original `CyclicEncoder` and
`WithCPPTransforms` APIs/defaults are unchanged.

`LoadScalingOnline(reader, trustedSHA256, maxElements)` uses buffered blocks to
read the experimental minimal binary. The digest must come from independently
verified preprocessing; matching a checksum does not prove code validity. The
format is experimental; count/header handling accepts s≤16 (s=14 first-use evidence is tracked
[separately](../../experiments/parallel-s14/report.md)), and is not a public artifact-distribution
commitment. Usage and profiling commands are in the
[parallel integration report](../../experiments/parallel-encoding/report.md).

`WriteScalingOnline(writer, plan)` streams the same minimal constants. It selects
the count width from s: s≤15 retains exact `LRSprobe1` bytes and 32-bit K/N;
s=16 uses `LRSprobe2` and 64-bit K/N, adding eight header bytes. Both versions
store coefficients as uint32. Writers own atomic persistence and trusted digest
metadata (for example, hash an `io.MultiWriter` stream). The loader reads both
versions and checks wide counts against the caller's element cap and host byte
offset range before allocating. Internal count calculations use uint64; this
widens a few scalar fields, not the large arrays. Native descriptions likewise
retain versions 1/2 for smaller domains and use 3/4 only for the 2^32 count.
See [s16 readiness](../../docs/encoding-s16.md): small-buffer field/transform and
header checks do not establish a full s=16 discovery/preparation or encoding run.
Experimental preparation now admits s=16 only under the
[first-use resource controller](../../experiments/hpc-s16/report.md); full
feasibility remains unverified. Ordinary discovery remains bounded to s≤13, and
s≥17 still requires wider extension values. The controller persists completed
support/constants with source provenance and retains profiles; incomplete
preparation still requires recomputation. Its small first-use/cached cases match
the retained reference bytes before the large case is admitted.


The [complete native backend documentation](../liftrs-cpp/README.md#complete-native-online-encoding)
explains stage counters, borrowing rules and the profiled comparison. The
immutable plan is shared; each concurrent encoding needs its own workspace.
`Close` waits for active calls and releases only small native metadata.

### Bounded preparation workspace

Support validation and discovery membership now use one bit per exponent rather
than hash-map entries or a byte per position. The opt-in scaling preparation
builds roots from at most 32,768 factors in flat leaf storage, merging completed
polynomials with a balanced schedule. Its private FFT buffers are reused across
convolutions and released before spectrum construction/return. Prepared objects
retain no mutable workspace; ordinary encoding calls keep their existing buffer
ownership. Newton inversion reuses full-length fresh results, preserving zero
padding when the product is shorter than the requested precision.

The [preparation checkpoint](../../experiments/preparation-checkpoint/report.md)
compares exact generated constants and separately labels diagnostic GC/heap
snapshots. This is an in-memory improvement, not an s=15 support extension or a
process-RSS guarantee. Normal preparation does not force collection; only the
opt-in diagnostic observer does so.


### Guarded s=15 admission

The [local s=15 experiment](../../experiments/parallel-s15/report.md) extends only
experimental native import, preparation and minimal-file limits to s=15 after
small-buffer 30-bit-field and allocation-free shape checks. The ordinary public
DiscoverOrbit limit remains s=13. Large execution now passes exact discovery, bounded preparation with heap
checkpoints, three complete native encodings, packed output checks and a fresh
cached run. Exact K is 216,914,047; median native encoding is 393.19 s and
fresh cached peak RSS 18.46 GiB (20.55 GiB first use). Preparation timing spans
a manual pause; these are standalone results, not application integration.
Retained source, profiles and cache hashes distinguish its evidence from the
completed s=14 measurements. Preparation remains Go with native transforms;
database-dependent encoding uses the complete C++ path.


The optional [GPU scaling harness](../../experiments/gpu-scaling/report.md) uses
versioned private plan counts through2^32 while keeping field values at32 bits.
Allocation-free metadata checking is separate from explicit engine workload and
device-memory limits. Workspace growth releases disposable old arrays first;
CPU constructors and production preprocessing defaults remain unchanged.


The [actual preparation harness](../../experiments/gpu-preparation/report.md)
optionally saves a normalized root product bound to its support digest/spec.
Replay checks its shape, field, normalization and digest, and still performs the
full generator identity check. Default callers use no checkpoint or persistence
callback. Preparation-only mode streams complete constants, closes executors and
stops before database encoding; atomic metadata distinguishes reusable files from
unfinished writes. No intermediate root-batch checkpoint is implemented.


### Explicit CPU and optional CUDA client encoding

The same verified `ScalingOnlinePlan` supports ordinary Go `Encode`, native CPU
`WithCPPMatrix(workers)` (liftrs_cpp,cgo), and optional `WithCUDA(options)`
(liftrs_cuda,cgo). Preprocessing may be obtained elsewhere; it does not select
client hardware. CPU-only builds require no CUDA toolkit, library or device.
CUDA-only builds do not require the C++ CPU binding. No silent fallback occurs:
choose CPU explicitly when GPU capability/resources are absent.

```go
// Native CPU client (build with liftrs_cpp,cgo):
cpu, err := plan.WithCPPMatrix(32)
if err != nil { return err }
defer cpu.Close()
word, err := cpu.Encode(message, cpu.NewWorkspace())
```

```go
// Optional CUDA client (build with liftrs_cuda,cgo):
gpu, err := plan.WithCUDA(liftrs.CUDAOptions{
    LibraryPath: "/trusted/path/libliftrs-cuda.so",
    DeviceMemoryLimit: 96 << 30, // Explicit full-s16 cap, not a universal default.
})
if err != nil { return err }
defer gpu.Close()
word, stats, err := gpu.EncodeWithStats(message)
```

Both return identical canonical uint32 arrays; packed storage remains s bits per
symbol. `CUDAArrayBytes()` reports the plan/workspace array payload without GPU
initialization; budget additional table/plan/driver overhead. GPU initialization
copies immutable public code constants; its object does not retain host plan
arrays. Calls serialize and Close waits; CPU workspaces remain independently
owned. Reuse the GPU object to avoid repeating code upload/setup.

Current GPU arithmetic bounds s2..16 are parameter-driven.
[Full s16 GPU encoding](../../experiments/cuda-s16-encoding/report.md) passes two
repeats against retained CPU word hashes:85.95/60.22 s with7.19 s setup,
67.423 GiB owned device allocations and39.024 GiB peak sampled host RSS. See [CUDA deployment/ownership](../liftrs-cuda/README.md) and the
[validation report](../../experiments/cuda-encoding/report.md). GPU table arithmetic
is variable-time and does not add production-security or cryptographic/application
encoding guarantees. Stable public discovery/TRL scope is unchanged.

The [buffer-reuse prototype](../../experiments/cuda-buffer-reuse/report.md) is
selected only by opt-in tests (`PIR_CUDA_REUSE=1` with the CUDA cached/oracle
workers). It does not change `WithCUDA`, `CUDAArrayBytes` or CPU selection.
It requires2L<=N, shares two work arrays and adds small parity storage. Validation
now passes with s16 CPU hash agreement,51.427 GiB device ownership and57.418/56.975 s
encoding. The matched s10 comparison shows no observed encoding penalty; s16 timing
comparison is historical. No public-default switch or s17 support is implied.
