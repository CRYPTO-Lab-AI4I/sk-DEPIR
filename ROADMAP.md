# Roadmap to TRL 5

The current target is a usable, validated Lift-RS client and preprocessing
library. TRL 5 is a planning target, not a claim about the present implementation
or a formal certification. TRL 6 may be considered later if needed.

The server is a storage I/O endpoint: it returns the records at the positions
requested by the client. Implement preprocessing, query generation, and client
decoding. Define the storage I/O contract and use simple readers in tests and
examples; server-side protocol computation and a production server are outside
this roadmap.

Upload optimization through [grouped storage layout](docs/protocol-io.md#storage-layout-optimization)
is explicitly in scope for TRL 5. Place each query-sharing group's same-position
values in a single readable record, so the client sends each address once using
ordinary storage reads. Implement the layout, streaming writer and client adapter;
verify unchanged reconstruction and measure actual upload and request overhead.
This goal does not require server-side protocol computation.

## Implementation sequence

1. Specify the encoder contract and input layout; sketch catalogue fields and
   analytical costs only.
2. Implement generator-matrix code preprocessing and systematic message encoding
   for small Lift-RS instances. Verify the dimension, information set, codeword
   constraints, and exhaustive tiny-instance retrieval before the application wrapper.
3. Test a small cryptographic encoding pipeline on genuine codewords. Use its
   behavior as a reference; do not impose its in-memory API on large-scale backends.
4. Use the small application wrapper with supplied plans to validate input
   conversion and end-to-end reconstruction before automatic planning.
5. Build the parameter planner using verified capacities, distinguishing them
   from paper and empirical estimates.
6. Implement faster preprocessing for useful configurations identified by the planner.

The [bounded paper/cyclic encoder comparison](experiments/encoder-comparison/report.md)
now covers the paper information set and exact codeword equality through N=65,536.
Reusable public preprocessing artifacts are separate from database encoding; these
small artifacts can be distributed with generation evidence. This is not yet a
large-instance streaming or GPU backend. The detailed
[five-step encoder plan](docs/encoding-algorithms.md#6-five-step-implementation-plan)
starts with table-free extension arithmetic and an s=9 encoding checkpoint, then
compact public artifacts and CPU measurements through s=12. Bounded-RAM encoding
and optional GPU acceleration are subsequent stages subject to this roadmap's
TRL gates; this plan does not declare TRL 6 achieved or automatically expand scope.

A complete planner or broad benchmark campaign must not delay the reference
encoder. Small-instance validation does not establish large-instance resource
feasibility. See the [encoder requirements](docs/encoder-contract.md) and
[planner design](docs/parameter-planning.md). The [PIR configuration catalogue and loader](pirconfig/README.md)
are available, with [Monte Carlo rate estimates](experiments/liftrs-rates/report.md)
for all 13 target configurations and supporting paper values. Target capacity
verification remains open. The [greedy planner and evaluator](planner/README.md)
now construct row-block plans and calculate the logical cost model without running
encoders. Logical split/stack width search and ratio-bounded full-download suffix
planning are implemented; automatic execution of those application plans remains open.
Capacity-ordered batched Pareto search over row-block choices is implemented with reported approximation
limits. A separate rounded cost-grid DP is available for comparison; its numerical windows
and approximation limits are reported. General rearrangement or globally optimal
partition search is not implemented.

## Ordered milestones

Complete the correctness foundations before optimizing or promising a stable
public API. Check off items only when linked code, documentation, and verification
evidence support completion. Record each work item on the
[message board](message-board/README.md).

### 1. Specify the supported construction

Accepted decisions: [polynomial-basis binary fields and submatrix input layout](docs/data-representation.md).
The design specifies bit packing, metadata conventions, initial candidate parameters,
and a deterministic small reference method. These contracts still need implementation
and validation; the concrete cryptographic suite and scalable encoder backend remain open.

- [ ] Document the paper-to-code correspondence, assumptions, and supported
  parameters. Initially retain locality `ell=q-1`; general locality is optional.
- [ ] Specify message capacity, record representation, padding, field-element
  encoding, coordinate order, and the mapping from requested records to codeword
  positions.
- [ ] Choose and document the initial supported field/dimension combinations,
  parameter validation, and limitations. Do not equate 128 dummy positions with
  a demonstrated security level.

Acceptance: preprocessing and client implementations can be reviewed against
one explicit specification, including the data and key lifecycle.

### 2. Complete the correctness foundation

- [x] Implement bounded generator-matrix preprocessing and systematic encoding,
  with [independent tests](encoding/liftrs/reference_test.go) and
  [online retrieval checks](pir/liftrs_reference_test.go).
- [x] Add a bounded [cryptographic reference](encoding/cryptoref/README.md) and
  [retrieval tests](pir/cryptoref_test.go) using separate encoding keys and XOR masks.
- [x] Complete a bounded [application correctness reference](encoding/appref/README.md)
  with [supplied-plan reconstruction tests](encoding/appref/appref_test.go).
- [ ] Integrate genuine Lift-RS encoding into the maintained large-instance application workflow. Keep the
  existing file-based RM placeholder clearly identified until replaced.
- [ ] Test retrieval from original records through preprocessing, storage reads,
  and decoding, including Lift-RS codewords beyond the current low-degree RM
  test subset.
- [ ] Add independent small test vectors and exhaustive small-instance checks
  for the supported parameter combinations.
- [ ] Use cryptographic randomness for curve sampling, dummy positions, and
  shuffling. Isolate deterministic randomness to explicit test support.
- [ ] Make tests self-contained, with temporary files and no dependency on a
  running server or changes to tracked examples.

Acceptance: a fresh checkout reproduces genuine Lift-RS retrieval, with evidence
covering each supported configuration and clear limits on security claims.

### 3. Design usable client and preprocessing APIs

- [ ] Separate preprocessing, query generation, and decoding responsibilities.
- [ ] Separate public query positions from private per-query decoding state.
- [ ] Decouple client algorithms from HTTP, CSV, and implicit configuration
  writes; document the request/response storage I/O contract.
- [ ] Return errors for invalid parameters, malformed responses, and I/O failures;
  define cancellation, ownership, cleanup, and concurrency behavior.
- [ ] Define a versioned encoded-database format and metadata that detect
  incompatible parameters or mismatched client/database configurations.
- [ ] Remove fixed intermediate filenames, working-directory assumptions, and
  implicit external-program discovery from the public interface.

Acceptance: an external application can preprocess a database and retrieve a
record using its own storage reader without modifying library internals.

### 4. Organize and package the library

- [x] Use the repository module path `github.com/CRYPTO-Lab-AI4I/sk-DEPIR`
  for external imports.
- [ ] Review dependencies, licensing, and attribution.
- [ ] Keep public client and preprocessing packages small; split `utils` into
  internal packages with clear responsibilities, such as fields, code
  construction, permutation, masking, and record layout.
- [ ] Choose the maintained preprocessing backend and document its build
  requirements. Retain other languages/backends only for an explicit purpose.
- [ ] Put small fixtures under `testdata/`, user examples under `examples/`,
  design/specification documents under `docs/`, and optional CLI code under
  `cmd/`. Move or remove the existing server demo as appropriate to the I/O scope.
- [ ] Keep experiment scripts, reports, and results together under `experiments/`
  using the existing [conventions](experiments/README.md).

Acceptance: developers can install, build, test, and understand the repository
without relying on the authors' workstation layout.

### 5. Validate a developer release

- [ ] Add CI for supported platforms and toolchains, including native
  preprocessing where applicable, formatting, static checks, and tests.
- [ ] Exercise invalid inputs, malformed responses, boundary parameters, and
  concurrent use where supported; use fuzzing and race checks where relevant.
- [ ] Provide a complete example from an external consumer module, API comments,
  installation instructions, format documentation, and known limitations.
- [ ] Add contribution and security-reporting guidance, release notes, citation
  information, and a documented versioning policy.
- [ ] Produce a reproducible integration report for a representative use case
  with stated parameters and environment, covering both preprocessing and the
  online client. Obtain review by someone outside the implementation work.

Acceptance: the documented workflow succeeds in a separate application and a
representative integration environment, and another developer can reproduce the
evidence. Any unsupported cases remain explicit.

## Deferred until a decision to pursue TRL 6

Large-scale external-memory preprocessing, advanced streaming and recovery,
specialized parallel/vectorized kernels, broad performance tuning, and large
database performance campaigns are deferred. After the CPU checkpoints in the
[five-step plan](docs/encoding-algorithms.md#6-five-step-implementation-plan), review
measurements with the user before starting external-memory or GPU work. Distinguish
resource fixes needed for a chosen TRL 5 use case from a decision to pursue TRL 6. Fix resource problems that prevent
the TRL 5 workflow from functioning, and preserve existing optimizations where
they remain correct and understandable.

Independent production security qualification, a stable 1.0 commitment, and
sustained operational deployment require additional evidence beyond this target.

## Current starting point

The repository has a bounded [Lift-RS reference encoder](encoding/liftrs/README.md)
with independent tiny-instance tests and online retrieval integration. This implements
code preprocessing and non-cryptographic encoding. The existing file-based API
still uses its RM placeholder and the online client uses experimental randomness.
See the [README](README.md) for current limitations.

A separate [cryptographic reference](encoding/cryptoref/README.md) now tests
permutation, XOR masks, fresh encoding keys, and genuine-codeword retrieval on
small instances. Its in-memory implementation does not constrain the future
large-scale API or storage strategy. The
[small application reference](encoding/appref/README.md) now completes this
correctness pipeline for supplied plans and ordinary packed reads. Automatic planning is implemented in `planner/`; executing those plans in the
application wrapper and durable persistence remain separate tasks. The orbit backend
encodes the standalone s=8,m=2,t=8 configuration with exact K=3121; s=9..20
still require further encoder and resource work;
the small reference provides comparison vectors for that work. Small test fields
do not inherit the catalogue's security claims.
