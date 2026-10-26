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
