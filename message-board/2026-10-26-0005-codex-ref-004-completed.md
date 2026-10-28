# REF-004: Small cryptographic reference implemented

- Timestamp: 2026-10-26T00:05:54.358148+00:00
- Agent: Codex
- Status: completed
- Base revision: 97bfb41; implementation remains in the working tree for review.

## Changes and boundaries

Added [encoding/cryptoref](../encoding/cryptoref/README.md), with its
[implementation](../encoding/cryptoref/reference.go), [unit tests](../encoding/cryptoref/reference_test.go),
and [runnable example](../encoding/cryptoref/example_test.go). It snapshots the
existing AES permutation over a bounded domain, verifies bijectivity, and uses the
existing AES mask-value helper with XOR rather than integer addition modulo q.
Masks bind to the resulting storage position. GenerateKey uses crypto/rand and
returns 32-byte keys; independent keys belong to permutation groups and database
encodings, with fresh masks on re-encoding. Coordinate identification is separate.

Lengths are 2..256; field degrees are 2, 3, and 4. Returned buffers are owned by
callers; initialized reference objects support concurrent access. Key reuse across
independent calls cannot be detected and remains the caller's responsibility.
No authentication or production cryptographic qualification is claimed. The
legacy RM/prime-field paths and existing query sampler are unchanged.

The reference defines correctness behavior only. Its materialized map, in-memory
codewords, layout, and API do not bind future streaming, chunked, or external-memory
implementations. Updated [encoder contract](../docs/encoder-contract.md),
[protocol references](../docs/protocol-io.md), [roadmap](../ROADMAP.md),
[README](../README.md), [reference encoder documentation](../encoding/liftrs-go/README.md),
and [repository map](../AGENTS.md) accordingly. Next is the small application
wrapper around supplied plans; catalogue-size encoding still needs a faster backend.

## Verification

Unit tests cover entire tiny permutation domains, masks compared with direct AES
block evaluation, a fixed AES zero-block vector, XOR round trips for every fixture
field value, separate database/re-encoding keys, repeated positions, invalid inputs,
key/output ownership, and concurrent use. Separate mask vectors can share individual
field values; tests compare full vectors rather than assert nonexistent per-position
uniqueness.

[Integration test](../pir/cryptoref_test.go) retrieves all systematic message positions
from genuine GF(16) lifted codewords beyond RM. Two databases and a re-encoding
share a permutation and generated query, with independent fresh masking keys.
The client repeats uploads and storage does plain reads. Explicit degree-2 curves
force repeated positions, and reconstruction preserves their multiplicity.

Go 1.27.3 darwin/arm64 from /tmp/sk-depir-go-toolchain/go/bin/go, with temporary
GOCACHE. Commands/outcomes:

- go test ./encoding/cryptoref ./pir with new reference test names and example:
  passed.
- go test -race ./encoding/cryptoref ./pir with the new unit/integration tests:
  passed, including concurrent masking and reads.
- go test ./... -count=1: all four packages passed in isolated checkout
  /tmp/sk-depir-ref004.9lrj2F using the existing RM executable. Isolation avoids
  legacy tests writing tracked example CSV files in the working tree.
- Current docs/package/README/roadmap links and Markdown anchors: verified.
- gofmt, new-file whitespace, and git diff --check: passed.

README reviewed and updated for scope, limitations, test coverage, and package
layout. No benchmark, performance experiment, or additional security claim was added.
