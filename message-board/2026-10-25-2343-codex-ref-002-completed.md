# REF-002: Small systematic Lift-RS encoder implemented

- Timestamp: 2026-10-25T23:43:58.890324+00:00
- Agent: Codex
- Status: completed
- Base revision: db73cd6; changes remain in the working tree for review.

## Implementation

Added [encoding/liftrs](../encoding/liftrs-shared/README.md) with
[preprocessing and encoding](../encoding/liftrs-shared/reference.go), independent
polynomial-basis [arithmetic](../encoding/liftrs-shared/field.go), and a runnable
[external-package example](../encoding/liftrs-shared/example_test.go).
Preprocess(Spec) enumerates good monomials, constructs the evaluation matrix,
and performs deterministic Gauss-Jordan elimination. Code.Encode consumes exactly
K field elements and returns an unmasked systematic codeword; the artifact can
be reused concurrently and exports copies of its metadata. The monomial DP tracks
positive paths separately so the constant monomial is not wrongly excluded.

Supports fixture fields GF(4), GF(8), and GF(16), dimensions 2/3 where N<=256,
and 1<=t<q-1. Bounds are checked before matrix allocation; dense element storage
is bounded by 256 KiB. The application wrapper, large configurations, packing,
and cryptographic encoding are outside this package. Existing file-based
LiftedRS.Encode still uses its clearly marked RM placeholder.

## Correctness evidence

[Reference tests](../encoding/liftrs-shared/reference_test.go) compare arithmetic with an
independent carryless-convolution/long-division oracle and check every nonzero
field element has an inverse. Monomial classification is compared against every
curve over GF(4) at t=1/2 and GF(8) at t=1. The GF(4), t=2 fixture has a hardcoded
information set and x-squared golden vector. Exhaustive recovery checks all
1024 messages, all 16 target positions, and all 256 degree<=2 curves through each
target, retaining curve-point multiplicity. Every accepted specification also
has deterministic all-grid recovery fixtures. Ownership, invalid inputs, exact
systematic recovery, deterministic setup, and concurrent reuse are covered.

[Online integration](../pir/liftrs_reference_test.go) encodes x^8*y^8 over GF(16)
with t=2 (beyond degree-7 RM), compares every codeword position against existing
field arithmetic, and retrieves every systematic message element using the
existing client, permutation/masks, 128 dummies, and plain in-memory reads.
This demonstrates prototype compatibility, not cryptographic qualification.

## Documentation and verification

Updated [encoder contract](../docs/encoder-contract.md), [roadmap](../ROADMAP.md),
[README](../README.md), and [repository map](../AGENTS.md). The docs checkpoint
was completed as DOC-049 before implementation. README now distinguishes the
implemented standalone reference from the RM-backed file API and planned wrappers.

Go was absent from PATH. Downloaded official Go 1.27.3 darwin/arm64 into /tmp,
verified SHA-256 dd3908cf20959d798c1e06ebe5b72213f0505ce01373b40ce3110975a6cd99c2,
and used a temporary build cache. Validation commands and outcomes:

- go test ./encoding/liftrs -count=1: passed, including example and exhaustive checks.
- go test ./pir -run '^TestReferenceLiftedRSEncoderOnlineRecovery$': passed.
- go test -race ./encoding/liftrs ./pir with only the new reference test names:
  passed (including concurrent reuse).
- go test ./... -count=1: all three packages passed in isolated copy
  /tmp/sk-depir-ref002.5mgp1N, using the existing RM executable; legacy tests write
  example CSV files, so running in the copy preserved workspace fixtures.
- gofmt and current docs/README/roadmap/package links and anchors: passed.
- git diff --check and whitespace check of new files: passed.

No benchmark, measured rate, or security claim was added. Tiny fixture dimensions
are automated-test assertions, not catalogue measurements. The next work is
cryptographic boundary adaptation/verification, then the application wrapper with
supplied plans. The paper's fast encoder remains needed for catalogue-size inputs.
