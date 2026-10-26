# REF-006: Application correctness reference completed

- Timestamp: 2026-10-26T00:22:52.624262+00:00
- Agent: Codex
- Status: completed
- Base revision: 14dae98; publication authorized by the user.

## Implementation

Added [encoding/appref](../encoding/appref/README.md):
[plan validation](../encoding/appref/plan.go), [input conversion/encoding](../encoding/appref/encode.go),
and [query/read/reconstruction](../encoding/appref/retrieve.go). It executes a supplied
rectangle plan with exact coverage, verified capacities, deterministic coordinate IDs,
canonical layout order, and compatible fixed query-sharing groups. Code preprocessing
is reused per spec within the call. Encoding pads and converts input bits, invokes
both reference stages, packs stored field elements, and returns public data separately
from private state. Every encoding uses fresh group/per-database keys.

Retrieval uses crypto/rand for curve coefficients, dummies, and unbiased shuffling,
generates all queries before I/O, reads every member separately including repeated
positions, and reconstructs privately after the complete read phase. The caller's
ReadAt callback handles ordinary byte ranges with context and no group knowledge.
Private entry indices retain curve multiplicity, including real/dummy collisions.
Read failures, invalid inputs, and cancellation return no partial application record.

Bounds: 256 application rows, 256 bits per record, 64 PIR databases; records have
at least 4 bits for the fixture catalogue. Codes remain the bounded GF(4)/GF(8)/GF(16)
references. Public layout metadata is available as a copied Plan; durable JSON
manifest/key persistence and reopening remain separate work. This in-memory API and
read strategy do not constrain large-instance implementations. No planner, cloud
adapter, production-security claim, or server computation was added.

## Evidence and documentation

[Tests](../encoding/appref/appref_test.go) cover a general nonaligned six-rectangle
layout with different fields, golden bit slices and padding, packed boundaries,
all 256 one-byte input values, 3D and high-degree fixtures, irrelevant groups,
fixed complete participation, identical repeated group uploads, deterministic
repeated-curve fixtures, invalid plans/records, malformed/failed reads, entropy
failure, cancellation, ownership, fresh re-encoding, and concurrent retrieval.
The [external-package example](../encoding/appref/example_test.go) demonstrates
encoding and retrieval through bytes.Reader.ReadAt without a server.

Updated [README](../README.md), [roadmap](../ROADMAP.md), all relevant
[representation](../docs/data-representation.md), [encoder](../docs/encoder-contract.md),
and [I/O](../docs/protocol-io.md) contracts, and [repository map](../AGENTS.md).
Explicitly distinguish implemented reference behavior from persistence, planning,
large-scale architecture, and the legacy RM API. README reviewed for setup, scope,
randomness wording, package layout, and planned/implemented boundaries.

Validation used official Go 1.27.3 darwin/arm64 at
/tmp/sk-depir-go-toolchain/go/bin/go with temporary GOCACHE:

- go test ./encoding/appref -count=1: passed, including the runnable example.
- go test ./... -count=1: all five packages passed in isolated checkout
  /tmp/sk-depir-ref006.Whe0Md (existing RM executable). Repeated after the final
  read-phase separation change. Isolation preserves tracked CSV examples written
  by the legacy RM tests.
- go test -race ./encoding/appref -count=1: passed after the final code change.
- Current documentation/package links and heading anchors, gofmt, new-file
  whitespace, and git diff --check: all passed.

No performance experiment or benchmark results were produced. The user requested
commit/push after implementation; publication outcome will be recorded separately.
