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

[encoding/appref](../encoding/appref/README.md) implements this workflow as a
bounded correctness reference using supplied plans. It generates fresh keys
internally, returns the public plan separately from private client state, and
supports packed ordinary reads. The planned general API may accept explicit
key material. Durable manifest/key persistence is not part of this reference;
its in-memory API does not constrain large-scale backends.

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

First apply the plan's [logical record mapping](data-representation.md#logical-record-layout),
if present: split records into fragments, stack them as logical rows, and zero-pad
final fragments. Extract the planned submatrices and identify them by logical
coordinates scoped to the plan. The current reference accepts direct layouts; it
does not automatically execute `ApplicationPlan` reshaping or full-download entries.
For each encoded submatrix, append zero rows and columns to the selected PIR dimensions and convert
each row into a field element using the [representation
conventions](data-representation.md#pir-record-representation). Pass the resulting
message to the non-cryptographic encoder, then pass its codeword to the cryptographic
encoder. Supply shared permutation keys and separate masking keys according to the
[representation rules](data-representation.md#encoding). The application encoder
executes the selected layout; it does not optimize the partition.

Full-download pieces bypass code preprocessing and Lift-RS expansion. Store their
matrix bits without expansion, using separate mask material where masked storage is
required; no secret permutation is needed. This specifies the boundary, not an
implemented full-download cryptographic suite.

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
it to field elements. These are the first reference implementations of code preprocessing
and non-cryptographic encoding, now used by the small application wrapper.

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
The dense fixture backend has no persistent artifact format. The separate orbit
backend now provides persistent public preprocessing artifacts. See the package documentation for supported
combinations, the API example, and tests. The [small application wrapper](../encoding/appref/README.md) uses these fixtures;
this wrapper remains limited to the dense fixtures and does not execute orbit
artifacts or automatic application plans.

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

[encoding/cryptoref](../encoding/cryptoref/README.md) implements this behavior for
small in-memory fixtures. Its permutation table and codeword slices are a
correctness oracle only. Future streaming or external-memory implementations may
use different storage, processing strategies, and APIs.

**Input:** the codeword, a permutation key, and the PIR database encoding's masking key.

**Output:** the stored database encoding, obtained by permuting the codeword positions
and masking each value at its resulting storage position. Define the permutation as a
map from codeword position to storage position, and use that storage position as the
mask's position identifier. The client reverses the mask before codeword decoding.

Apply the [mask-separation
rules](data-representation.md#identification-and-mask-separation). Keep cryptographic
randomness separate from coordinate identifiers. The reference key and mask
formats are documented in its package; future backend formats and optional
master-key derivation remain separate decisions. Databases sharing a query must also meet the
[compatibility rules](data-representation.md#query-sharing-groups). This stage preserves
the codeword's length and field-element width. It may be streamed with the first stage,
but its behavior and measurements remain distinct.

**Legacy code:** `permuteAndEncryptMatrix` in [pir_api.go](../pir/pir_api.go) uses
[Permutator](../utils/permutator.go) and [Encryptor](../utils/encryptor.go). These are
prototype components; the current slice/index mask interface does not implement the
planned separation across PIR databases and re-encodings. The separate reference
stage uses its mask-value helper with XOR, a shared permutation per group, and
fresh independent masking keys per database encoding; it does not call the
legacy integer-modulus Encrypt/Decrypt methods.

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
fast encoder or support the initial `s=8..20` catalogue. The [bounded orbit backend](../encoding/liftrs/README.md#paper-and-cyclic-encoders)
now implements the paper's information set, certified random degree discovery and
paper/cyclic encoding in an experimental m=2,s=2..13 range (N<=67,108,864). Its additive transforms, prepared
public artifacts and method-specific constants are independently tested against
the dense oracle. Its [hybrid polynomial products](../experiments/fast-polynomial/report.md)
now cover full and low-coefficient multiplication, including preprocessing and
per-database computations. Target-size encoder coverage, memory/I/O scheduling
and hardware acceleration remain open.

**Cryptographic reference status:** the bounded stage now tests permutation, XOR
masks at storage positions, independent database keys, and fresh masks on
re-encoding. See its [unit tests](../encoding/cryptoref/reference_test.go) and
[genuine-codeword retrieval test](../pir/cryptoref_test.go). The existing custom
permutation still needs cryptographic review. A large-scale implementation is
not selected by this reference. Independently generated masking keys suffice for
these fixtures; optional master-key derivation remains separate. Generate fresh masking keys on re-encoding and retain them separately
from coordinate metadata. Key generation and online sampling must use Go's
[`crypto/rand`](https://pkg.go.dev/crypto/rand); deterministic sources belong only
in explicit tests. The legacy integer-additive masking remains separate. These
fixture tests do not qualify the reference or the custom permutation for production.

See the [encoding algorithm review](encoding-algorithms.md) for the paper's pipeline,
the validated bounded cyclic formulation, and further transform implementation candidates. These
methods preserve the stage boundaries. Dense fixtures remain limited to N<=256;
the separate orbit backend accepts m=2,s=2..13, N<=67,108,864 for resource-limited experiments.

Complete public code-preprocessing artifacts can be distributed independently of
any database (for example, on GitHub). Store reusable solver constants and transform
spectra in those artifacts where their size is practical. Report discovery,
method preparation, serialization and process-local loading separately from
per-database encoding. No published size budget is established for the large
catalogue configurations; do not assume a full matrix or spectrum fits that budget.

The [five-step implementation plan](encoding-algorithms.md#6-five-step-implementation-plan)
specifies arithmetic, compact artifacts, moderate-instance CPU checks, bounded-RAM
encoding and optional GPU acceleration, with acceptance criteria for each stage.
Its proposed 64 MiB publication budget is configurable and distinct from local
working-memory or disk limits. It does not change current backend support.

Publication verification is distinct from loader validation. Follow the
[artifact acceptance checks](encoding-algorithms.md#step-2-define-compact-reusable-public-preprocessing-artifacts)
before publishing generated constants; a successful shape/range check does not
prove the artifact describes the intended code. Stage-specific memory and realistic
packed/grouped output measurements are required by the
[CPU scaling checkpoint](encoding-algorithms.md#step-3-scale-the-in-memory-cpu-encoder-through-s912).

The [table-free field API](../encoding/liftrs/README.md#table-free-field-arithmetic)
now covers s=2..20 and quadratic extensions without cardinality-sized tables.
The bounded encoder adapter separately remains limited to m=2,s<=13; its uint32
packed coefficients and in-memory arrays must not be used for larger encodings.
[The first larger checkpoint](../experiments/tablefree-encoding/report.md) records
exact s=9 capacity and separate code-only preparation, load and encoding evidence.
Version-1 s<=8 artifacts retain their conventions and complete output hashes.
The [s=10 checkpoint](../experiments/encoder-scaling/report.md) adds exact K=103556
with reusable JSON artifacts, paper/cyclic comparisons and separate memory/output
measurements. The [progressive cyclic measurements](../experiments/cyclic-scaling/report.md) now
verify the s=11/12 catalogue cases in memory, with exact capacities K=455434 and
K=1965739. The subsequent [s=13 measurement](../experiments/cyclic-s13/report.md)
adds exact K=11926435, 29.42 min reusable preparation, 185.15 s median encoding
and 5.34 GiB peak process RSS. It uses a two-hour timeout; one hour is a ballpark,
not a hard first-use limit. The s>=14 resource/correctness checkpoints remain
open. That unprofiled checkpoint did not serialize large caches; default JSON
loader caps remain separate limitations. The later profile experiment below
retains and reloads an s=13 cache through an explicit local limit.

The bounded CPU backend now caches four-bit multiplication maps for repeated FFT
constants, with at most 256 KiB of table entries per transform. Construction is
code-only runtime preprocessing (also performed on artifact loading); encoding
reuses the immutable tables. They do not change the field representation,
information set or version-1 serialized artifacts. The
[multiplication comparison](../experiments/field-multiplication/report.md)
separates setup, encoding and memory; table indexing is not constant-time.


During optimization, [experiment policy](../experiments/README.md#profiling-and-preprocessing-retention)
requires stage-specific CPU profiles and retention of reasonably sized reusable
preprocessing by default. Persist local caches with specifications/provenance and
checksums so later encoding profiles do not require repeating preparation. Keep
large caches outside Git, account for machine resources, and record exceptions.
Cache save/load is separate from encoding; profiled times remain distinct from
unprofiled baselines. Disable profiling once optimization is complete.


The cyclic JSON loader now offers `LoadCyclicEncoderWithLimit` for explicitly
bounded trusted local caches larger than the default limits. The opt-in byte cap
also applies to the embedded basis; defaults remain unchanged. Bound file reads
and decoded RAM separately. [The profiled s=13 experiment](../experiments/encoder-profile-s13/report.md)
retains a 432 MiB local cache with provenance/checksums, reloads it in 4.83 s
and profiles encoding separately (186.35 s). Its output matches the baseline.
This local cache policy does not raise the public distribution budget.
