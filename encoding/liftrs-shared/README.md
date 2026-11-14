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

The private [wide packed scalar reference](orbit_arithmetic64.go) starts the
s17 port while retaining the existing tower and32-bit base values. Its
[bounded validation](../../experiments/s17-arithmetic/report.md) passes scalar/oracle
and legacy-ordering checks through s20. It does not extend
encoder, cache or native/CUDA support beyond s16.

The separate [ScalingCodeData APIs](../../docs/scaling-code-data.md) load/write
old code data without widening legacy values and add an explicit s17–20 spectrum
representation. LoadScalingCode requires trusted SHA-256 and count/byte caps;
VerifyScalingCode checks a stream without resident arrays. See the
[runnable example](scaling_code_example_test.go) and
[bounded validation](../../experiments/s17-code-data/report.md), passed as job74396
with legacy/header/slot checks and shared Go/CPU regressions.
This does not extend the shared encoder/cache-preprocessing admission beyond s16.

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
completed memory/runtime comparison: s16 sampled RSS58.351 GiB, two complete
outputs match CPU/GPU hashes and retained-output/overlap/race/sanitizer checks
pass. Timing samples are mixed, not a general runtime-equivalence claim. This opt-in path
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

The [bounded GPU output experiment](../../experiments/gpu-output-memory/report.md)
adds a private compact constructor to opt-in tests (`PIR_CUDA_COMPACT=1`, selected
by the runner's explicit `--memory-mode compact`). It returns the same owning
uint32 canonical word after bounded device compaction and backwards host expansion.
Normal WithCUDA and CPU paths are unchanged. Extra work and half-sized bulk
download are measured separately. Validation now passes at s16 with35.429 GiB
owned GPU memory and101.5/93.3 s encoding. Historical reuse calls took57.4/57.0 s;
matched s10 compact is9.6–10.7% slower. See the report for comparison limits.


## Explicit compact CPU workspace

The experimental [bounded CPU output route](../../experiments/cpu-output-memory/report.md)
retains normal owning uint32 results while reducing scratch and assembly memory.
Use separate compact methods; ordinary encoding remains the validated fast path:

```go
cpu, err := plan.WithCPPMatrix(32)
if err != nil { return err }
defer cpu.Close()
workspace, err := cpu.NewCompactWorkspace()
if err != nil { return err } // Requires s<=16 and 2L<=N.
word, stats, err := cpu.EncodeCompactWithStats(message, workspace)
```

`stats` adds compaction/move and backwards expansion timings. Bounded column
batches include gather/scatter in column timing. The workspace detaches its arena
on success, preserving earlier words across calls/closure; `Bytes()` excludes that
caller-owned result. Next use allocates a replacement. Slower execution is possible;
when normal memory fits, this route requires an explicit CLI choice. The experiment
provides `--memory-mode fast|compact` (default fast); production CLI integration is
still planned. s16 principal memory is modeled at39.345 GiB with32 workers, versus
58.351 GiB measured for the fast path. Compact validation now passes:39.409 GiB sampled RAM,21m29.8s full s16 encoding
and retained CPU/GPU hash agreement. Matched s14 compact is3.9–6.1% slower; full
s16 timing comparison is historical. No s17 support, default switch or new
cryptographic claim follows.


The [fast-layout streamed GPU experiment](../../experiments/gpu-streamed-constants/report.md)
is selected privately by `PIR_CUDA_STREAM=1`, driven by explicit
`--memory-mode streamed` (runner default resident). Its adapter retains immutable
host P/spectrum in Go until Close; C/CUDA borrow them per encode and retain no Go
pointer. This differs deliberately from ordinary WithCUDA, which copies code once
and need not retain the host plan. Both return owning canonical uint32 words.
Streamed P initializes prefix data directly and spectrum tiles borrow dead scratch
between full forward/inverse, retaining fast rows. Measured32.005 GiB device memory
at s16, plus host plan/input/output about39 GiB; extra19.422 GiB uploads per encoding.
Stream counters separate code bytes/time/chunks/tile without double-counting H2D.
Validation passes both full s16 encodes with retained CPU hashes:63.49/58.85 s,
about2.24 s code uploads, host RSS39.03 GiB. Matched s14 is5.6–6.7% slower; full
s16 timing comparison is historical. Resident remains default when it fits, and
ordinary CPU builds still require no CUDA. No s17/application/cryptographic
support is added.


## Explicit CPU file-backed constants

[CPU file streaming](../../experiments/cpu-streamed-constants/report.md) retains
fast transforms while loading verified constants into existing work buffers:

```go
fileCPU, err := liftrs.OpenScalingCPPFile(path, trustedSHA256, maxElements, 32)
if err != nil { return err }
defer fileCPU.Close()
word, stats, err := fileCPU.EncodeWithStats(message)
```

Build with liftrs_cpp,cgo, without CUDA. The opened LRSprobe1/2 source must come
from verified preprocessing and remain immutable until Close. Open checks the
whole expected SHA256, shape/field ranges and builds native metadata; byte identity
alone is not mathematical validation. Descriptor identity survives path replacement;
size/mtime checks detect common changes but not hostile concurrent writes. Encoding
rejects read/truncation errors before returning a public word. Prevent external
writes. Requires s<=16,2L<=N; ordinary resident/compact methods stay unchanged.

P initializes prefix data directly. Spectrum tiles borrow dead transpose scratch
between full forward/inverse, then scratch becomes detached owning output. Old
words survive reuse/Close. WorkspaceBytes excludes detached words and file cache;
first encoding lazily allocates two N arrays and later calls replace output scratch.
No resident code/saved H or extra read tile is allocated. FileStats separates
logical reads/range-check time, bytes and chunks; prefix read time overlaps buffer
time. Reads can hit kernel cache, so counts are not physical disk traffic.

Model35.422 GiB principal RAM at s16 versus58.267 GiB resident, but cached code can
add19.422 GiB and packed output cache more. The experiment uses explicit
`--memory-mode file` (default resident), records RSS/cache/cgroup observations and
never drops shared caches. Validation passes:20m49.6s full s16 encoding,35.484 GiB
sampled RSS and retained CPU/GPU hashes. Sampled RSS+tracked file pages reaches
61.820 GiB; charged batch-task cgroup peaks around47.92 GiB, a different accounting
view. This is not a total35 GiB cap. Matched cached s14 costs0.9–1.2% extra encoding
time. No automatic default switch, s17 capability, cryptographic/application or
TRL5 claim follows.


## Explicit GPU packed sink

With the optional CUDA build, `gpu.EncodePackedTo(message, destination, bufferBytes)`
returns `CUDAPackedStats` and writes the same canonical s-bit database format.
It allocates no full host canonical uint32 array. The native fast-layout kernel
packs directly into dead scratch, using atomic bit-field writes; streamed-code
adapters retain their existing code upload contract. Ordinary Encode/EncodeWithStats
remain owning APIs with unchanged defaults. Compact/direct-row constructors do
not support this new sink. CPU-only builds provide an unavailable stub and need
no CUDA dependency. Full s16 validation passes:65.5–66.2 s including writing/hash,
23.01 GiB encoding RSS, separate verifier31.01 GiB; GPU32.005 GiB unchanged. See
[GPU packed-output evidence](../../experiments/gpu-packed-output/report.md).

```go
stats, err := gpu.EncodePackedTo(message, outputFile, 8 << 20)
if err != nil { return err } // Discard/retain partial output according to caller policy.
// For files, Sync/Close and publish only after successful completion.
```

The sink buffer must be1 byte..64 MiB; destination-owned buffers/file cache are extra.
The writer must consume bytes synchronously and must not reenter the encoder.
Calls/Close serialize. Errors may leave partial output; the digest describes a
full result only on success, not a durable-storage transaction. NativeTotalNS
includes preparation+downloads; ElapsedNS includes buffer allocation/write/hash.
WriteNS/HashNS are separate, download time already belongs to D2H/native totals.
Keep this explicit slower-memory option optional when owning memory fits. s<=16
and canonical coordinates/cryptographic scope remain unchanged; s17 unsupported.


The next [host-backed GPU transform component](../../experiments/gpu-host-backed/report.md)
uses explicit low-level `liftrscuda.NewHostBacked`/`ApplyBacked`; it is not yet
integrated into ScalingCUDAEncoder. It retains the full host domain, bounds GPU
tiles, and introduces repeated transfers. The first component passes correctness
but costs60.5x resident latency at2^22 values with very small tiles;
[parallel work inside each tile](../../experiments/gpu-backed-parallel/report.md)
is being validated at unchanged memory/transfer bounds. These are component costs,
not full encoder estimates.
Resident shared encoder defaults remain unchanged; CPU output/backed gates follow.


## Explicit host-backed GPU online encoding

`plan.WithCUDAHostBacked(options,tileElements)` retains immutable host code and
uses bounded GPU tiles for the complete five-transform online schedule. Select
it explicitly for low device memory; `WithCUDA` and CPU defaults stay unchanged.
It requires2L<=N and s<=16, one host4N arena, a new owning4N result and at most4T
native staging, plus code/message. Products/transforms/orbit generation execute
on GPU; native CPU scatters canonical values and XORs zero-coordinate parity.
No full orbit map. It can be much slower than resident mode, even after the
[validated tile parallelism](../../experiments/gpu-backed-parallel/report.md).

```go
gpu, err := plan.WithCUDAHostBacked(CUDAOptions{
    LibraryPath: library, DeviceMemoryLimit: 512 << 20,
}, 65536)
if err != nil { return err }
defer gpu.Close()
word, stats, err := gpu.EncodeWithStats(message)
// stats reports retained host arena, peak native staging and transferred bytes.
```

Code stays immutable/alive until Close. Calls/Close serialize; failed calls
return no usable output and later calls initialize work again. Owning results
are detached across calls. This mode rejects EncodePackedTo; the validated
resident packed sink remains separate. CPU-only builds expose unavailable
behavior without CUDA dependencies. [Complete encoding validation](../../experiments/gpu-host-encoding/report.md)
passes full s10 correctness:1.61 s versus native CPU0.31–0.32 s, explicit GPU
560400 bytes. Full s16 time/RAM and s17 support are not demonstrated.


## Explicit CPU packed sink

With `liftrs_cpp,cgo`, `OpenScalingCPPFile` also supports
`EncodePackedTo(message,writer,bufferBytes)`. It packs canonical s-bit values
into dead scratch and reuses both N-word transform arrays across calls. No full
canonical uint32 result is detached or separately packed. The8N-byte transform
peak remains; output storage, file cache and caller-retained words are extra.
Atomic byte packing can cost runtime; owning methods remain unchanged/default.

```go
stats, err := encoder.EncodePackedTo(message, outputFile, 8 << 20)
if err != nil { return err } // Partial destination may remain.
// Sync/Close and publish separately after success.
```

The buffer must be1..64MiB. Writer consumption is synchronous; it must not reenter
this encoder. Calls/Close serialize through emission. Full digest is populated
only on success. Stats include workspace/read bytes, native stages, compute,
writer/hash and whole-call times. [CPU validation](../../experiments/cpu-packed-output/report.md)
passes small/reference/ASan checks and matched s10 file pipelines. The twoN
transform arrays remain counted. Full s16/s17 timing is not demonstrated.


## Explicit CPU disk-backed encoder

With `liftrs_cpp,cgo`, `OpenScalingCPPDisk(codePath,digest,maxElements,workers,
tileElements,workspaceDirectory)` streams verified code and creates exclusive
`data.bin`/`scratch.bin` workspace files (4N bytes each). No existing work file is
replaced. `EncodeWithStats` returns a new owning uint32 canonical word; it retains
no full host transform/code arrays. All arithmetic and orchestration are native.
Calls/Close serialize. Close preserves work files; callers manage their lifetime
and must exclude other access/truncation. They are not resumable checkpoints.

```go
encoder, err := OpenScalingCPPDisk(codePath, digest, maxElements, 16,
    65536, workspaceDirectory)
if err != nil { return err }
defer encoder.Close()
word, stats, err := encoder.EncodeWithStats(message)
```

This explicit memory mode is slower than resident/file-fast paths. Owning output
still costs4N bytes and message4K; native tiles conservatively16T+128KiB plus
metadata/runtime. Two backing files8N and filesystem cache are additional. At
s16, output/message project19.42GiB payload, backing32GiB; full storage is not
admitted under current budgets. Slower selection must remain explicit in CLI
programs; defaults stay unchanged. Stats label nested logical file I/O versus
compute; no fsync is timed. Errors discard the returned output; retry starts the
whole schedule after restoring workspace availability. [Full validation](../../experiments/cpu-disk-encoding/report.md)
passes small/ASan and complete s10 checks:2.86/1.14 s first/reused versus
file-fast0.33/0.31 s on NFS. Native buffers1.125MiB plus4MiB owning output and
8MiB backing files; cache/runtime extra. Full s16 storage/run is not admitted.
No bounded packed sink for disk mode or s17 support is added.


## Experimental wide CPU encoding

`ScalingCodeData.WithWideCPP(workers)` connects verified width-aware code data
to an optional complete native schedule; use `Resources`, capped `NewWorkspace`,
`EncodeWithStats`, then `Close`. The [example](scaling_wide_cpp_example_test.go)
shows the API. Source spectra stay32-bit through s16 and64-bit thereafter;
messages and canonical results remain uint32. Ordinary builds return
`ErrCPPUnavailable`; existing CPU/GPU defaults remain.

Each successful result owns the dead scratch allocation:4N logical bytes retain
8N backing bytes. Repeated calls allocate new scratch and preserve prior words.
Workspace caps exclude code/input, native metadata, transient parity, retained
older results and runtime headroom. Do not copy/concurrently reuse workspaces.
See the [full accounting](../../docs/scaling-code-data.md#optional-native-encoding)
and [bounded validation experiment](../../experiments/s17-encoding/report.md),
passed in bounded job74412. Full s17 preprocessing, encoding and latency are unvalidated.


## CPU implementation and validation

The [CPU implementation guide](../../docs/cpu-encoding.md) explains shared native
encoding order,32/64-bit kernels, resident/compact/file/disk/output modules,
Go/native ownership and optional job progress. Public constructors remain
compatible. Legacy resident and file loaders reuse the width-aware decoder/
verifier; native prefix/full metadata setup is common. Closed/uninitialized
resident encoders return nil/error from workspace creation instead of allocating
or panicking. Independent Go/dense/hybrid encoders remain correctness references.
[Cross-mode and baseline validation](../../experiments/cpu-encoding-modules/report.md)
is pending; this refactor does not establish s17 support or new memory savings.
