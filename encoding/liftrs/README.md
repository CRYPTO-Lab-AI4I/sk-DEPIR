# Small Lift-RS reference encoder

This pure-Go package implements code preprocessing and systematic encoding for
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

Import `github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs`. The runnable
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
go test ./encoding/liftrs
go test ./pir -run '^TestReferenceLiftedRSEncoderOnlineRecovery$'
go test -race ./encoding/liftrs
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
and two equivalent encoding algorithms for `m=2`, `s=2..10`, `1<=t<q-1`, with
`N<=1,048,576`. This extends the comparison backend, not the dense `Preprocess` API
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
there is no external-storage, streaming, GPU or constant-time implementation.
It does not support s=11..20 or integrate automatically with application plans.
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
primitives. The public arithmetic does not imply full encoder support at s=11..20.

The bounded orbit adapter still uses uint32-packed extension coefficients and
in-memory transforms, restricted to s<=10. It preserves version-1 artifact bytes,
polynomials and primitive-generator selection. Discovery now samples four bytes
per monomial coefficient for s=9, masking to 9 bits; a single byte would not yield
uniform coefficients. The [tests](binary_field_test.go) independently check all
supported field degrees, historical codeword hashes and sampled s=9 curve constraints.
The legacy `utils.GF2n` table-based backend and its field coverage are unchanged.
