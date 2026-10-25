# Encoder requirements

Encoding an application database involves selecting a partition, preparing each Lift-RS
code, and encoding the resulting PIR databases. The stages below define the separate
responsibilities of planning and encoding. These are requirements for the planned
implementation; see the [roadmap](../ROADMAP.md#implementation-sequence) for development
order.

## Planner boundary

The [planner](parameter-planning.md) returns one or more plans for partitioning the
application matrix into PIR databases, with rates and estimated costs from the
configuration catalogue. It does not run code preprocessing, encode application data, or
manage encoding keys. The caller selects a plan and passes it to the application
encoder.

## Application encoding

**Input:** the application database, a selected partition plan, and the key material
needed for cryptographic encoding.

**Output:** the collection of stored PIR database encodings and the layout and
configuration metadata needed for queries and reconstruction. Keep secret keys separate
from public metadata.

Validate the plan against the input dimensions, exact matrix coverage, supported
configurations, and query-sharing rules. Obtain reusable encoding data from [code
preprocessing](#code-preprocessing) and check the plan against its verified capacities.
If an empirical capacity estimate does not fit, report the mismatch; do not silently
change the plan or select another one.

Extract the planned submatrices and identify them by their application coordinates. For
each submatrix, append zero rows and columns to the selected PIR dimensions and convert
each row into a field element using the [representation
conventions](data-representation.md#pir-record-representation). Pass the resulting
message to the non-cryptographic encoder, then pass its codeword to the cryptographic
encoder. Supply shared permutation keys and separate masking keys according to the
[representation rules](data-representation.md#encoding). The application encoder
executes the selected layout; it does not optimize the partition.

Each PIR encoder processes one PIR database and knows nothing about application matrix
boundaries or the other submatrices. Application-level query orchestration and
reconstruction are described in [protocol I/O](protocol-io.md).

## Code preprocessing

The code is defined in [Definition 4.4](../paper.pdf#page=24). [Section
4.2.1](../paper.pdf#page=25) describes its monomial basis, with [Algorithm
1](../paper.pdf#page=28) testing monomial membership; [Section 4.2.4, Lemma
4.9](../paper.pdf#page=31) describes preprocessing and systematic encoding. The reusable
interface below is a library requirement.

**Input:** only the Lift-RS code specification, including the field representation, code
constraints, and coordinate conventions. This stage takes no database or secret keys and
is separate from parameter planning.

**Output:** reusable information for encoding under that code: its verified message
dimension `K` (input field elements), codeword length `N`, and a validated set of
coordinates that can hold the input unchanged (the systematic information set), together
with its index mapping and any auxiliary data needed by the encoding algorithm.
Specifying the code and constructing this information can itself require substantial
computation.

A systematic generator matrix is one possible artifact for small codes; its size may
make it impractical for larger codes. The interface must allow other representations
without requiring a dense matrix. Identify each artifact by its complete code
specification and preprocessing implementation version, and validate compatibility
before reuse across databases.

Measure code preprocessing time, peak memory, and artifact size separately from database
encoding. Report whether encoding measurements reuse an existing artifact.

### Small reference encoder

The bounded implementation is [encoding/liftrs](../encoding/liftrs/README.md).
`Preprocess(Spec)` constructs reusable code data; `Code.Encode(message)` applies
it to field elements. These are the first implementations of code preprocessing
and non-cryptographic encoding, before the application wrapper.

The deterministic method enumerates exponent tuples, classifies
monomials using the paper's membership criterion, and evaluates the accepted
monomials to form a generator matrix `G` with `K` rows and `N` columns. Verify
membership independently on exhaustive tiny instances.

Enumerate field elements by their polynomial-basis integer values. Flatten grid
coordinates as `x_0 + q*x_1 + ... + q^(m-1)*x_(m-1)`, matching the existing
[index conversions](../utils/indexTransformer.go). Enumerate exponent tuples in
the same order. Reduce `G` by Gaussian elimination, scanning columns in this
order and taking the first available pivot row. Pivot columns form the systematic
information set. The resulting matrix has an identity submatrix there, so message
row `j` maps to the `j`-th pivot column and encoding is message-vector multiplication.

Keep this dense artifact in memory, reusable across messages with the same code.
Check dimensions and allocation sizes before constructing it; reject instances
outside the reference limit `N <= 256`. The backend uses field degrees 2, 3,
and 4 as correctness fixtures, with evaluation dimensions 2 or 3 where they fit.
Its matrix size is `K*N` field elements, so this is a correctness reference rather than a scaling strategy.
A persistent artifact format is unnecessary for the first implementation; add one
when an encoder needs cached setup. See the package documentation for supported
combinations, the API example, and tests. The application wrapper and larger
field configurations remain to be implemented.

## Non-cryptographic encoding

**Input:** a message of exactly `K` field elements and a compatible code-preprocessing
artifact. Validate the message length and field-element values against the artifact.
Padding and bit conversion have already been performed by the application encoder.

**Output:** an unpermuted, unmasked systematic Lift-RS codeword of `N` field elements.
Here `K` is the artifact's verified message dimension and `N=q^m`, where `q=2^s` is the
field size and `m` is the evaluation-space dimension. The encoder uses no secret keys.

Systematic encoding preserves input elements at designated codeword coordinates. Use the
validated information set, coordinate enumeration, and mapping from local input index to
systematic coordinate supplied by code preprocessing; an arbitrary prefix of the
evaluation grid is not automatically suitable. The paper gives an interpolating set in
[Theorem 4.3](../paper.pdf#page=31) and the encoding steps on [page
32](../paper.pdf#page=32).

**Existing file-based API:** `LiftedRS.Encode` in [pir-lifted.go](../pir/pir-lifted.go) calls the
[RM encoder wrapper](../pir/rm_encoder.go), a testing placeholder. The unused `LRSE2D`
helper in [rme.cpp](../encoding/rme.cpp) is not an implementation of this binary-field
systematic encoder or its preprocessing.

## Cryptographic encoding

See the paper's [Figure 5](../paper.pdf#page=33) for masking and [Definition
5.2](../paper.pdf#page=34) for code permutation.

**Input:** the codeword, a permutation key, and the PIR database encoding's masking key.

**Output:** the stored database encoding, obtained by permuting the codeword positions
and masking each value at its resulting storage position. Define the permutation as a
map from codeword position to storage position, and use that storage position as the
mask's position identifier. The client reverses the mask before codeword decoding.

Apply the [mask-separation
rules](data-representation.md#identification-and-mask-separation). Keep cryptographic
randomness separate from coordinate identifiers; the concrete key-derivation and
randomness formats remain open. Databases sharing a query must also meet the
[compatibility rules](data-representation.md#query-sharing-groups). This stage preserves
the codeword's length and field-element width. It may be streamed with the first stage,
but its behavior and measurements remain distinct.

**Existing code:** `permuteAndEncryptMatrix` in [pir_api.go](../pir/pir_api.go) uses
[Permutator](../utils/permutator.go) and [Encryptor](../utils/encryptor.go). These are
prototype components; the current slice/index mask interface does not implement the
planned separation across PIR databases and re-encodings.

## Verification and open decisions

Target every integer field degree `s` from 8 through 20. Validate the field polynomial
and arithmetic before enabling a configuration. Verify that systematic positions
reproduce the input, generated codewords satisfy the code constraints, and permutation,
masking, and retrieval recover the original input. Use independent small vectors and
exhaustive small-instance retrieval checks. The existing
`TestLiftedRSEndToEndWithGF2nCodeword` in [pir_api_test.go](../pir/pir_api_test.go)
checks the online path on a directly evaluated codeword. The new
[reference encoder tests](../encoding/liftrs/reference_test.go) check genuine
systematic encoding; [online integration](../pir/liftrs_reference_test.go)
retrieves every systematic position of a lifted codeword beyond RM. The RM-backed tests in
[pir_test.go](../pir/pir_test.go) do not verify this Lift-RS encoder.

The initial `(s, m, t)` targets are in [parameter planning](parameter-planning.md).

**Methods supplied by the paper:** [Algorithm 1](../paper.pdf#page=28) tests monomial
membership; [Theorem 4.3](../paper.pdf#page=31) supplies an interpolating set; [Claim
4.2 and the encoding steps](../paper.pdf#page=32) describe randomized degree-set
preprocessing and systematic encoding through interpolation and evaluation.

**Reference implementation status:** the bounded dense method above is implemented
and independently tested on tiny fields. It establishes the correctness baseline
for future backends and wrapper integration. It does not implement the paper's
fast encoder or support the initial `s=8..20` catalogue. The faster method needs
an extension-field isomorphism, interpolation/evaluation algorithms, and validation
of randomized degree-set preprocessing; evaluate those against the reference.

**Cryptographic follow-up:** review and adapt the existing AES-based permutation
and masking components to this contract, including binary-field masks and separation
across encodings. Select optional master-key derivation only when needed. For the first
integration, independently generated per-database masking keys suffice; HKDF is
optional. Generate fresh masking keys on re-encoding and retain them separately
from coordinate metadata. Key generation and online sampling must use Go's
[`crypto/rand`](https://pkg.go.dev/crypto/rand); deterministic sources belong only in explicit tests. The current
custom permutation and integer-modulus masks are prototype code, not a selected
binary-field cryptographic suite.
