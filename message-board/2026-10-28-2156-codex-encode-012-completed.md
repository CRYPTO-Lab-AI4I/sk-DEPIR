# Native transforms integrated and full encoding compared

- Timestamp: 2026-10-28 21:56 UTC
- Work item: ENCODE-012
- Status: completed

Added an opt-in batched [C++ ABI and Go binding](../encoding/liftrs-cpp/README.md)
and [cyclic encoder adapter](../encoding/liftrs-shared/cpp_encoder.go). Go still owns
preprocessing, cached spectra and all polynomial orchestration/optimizations;
whole transforms borrow buffers without interface copies. Native plans own
immutable code metadata, with explicit Close and concurrent independent buffers.
Default builds retain Go without requiring cgo/C++. New tagged/stub tests and a
runnable API example document build, lifetime and error behavior.

The [integration experiment](../experiments/native-encoding/report.md) includes
runner/analyzer, source manifests, raw events, per-stage profiles, native stack
sample, retained binary/cache links and verification logs. At s=13, m=2, t=5,
three-call medians are 190.657 s (Go) and 153.409 s (native transforms): 1.243×,
19.5% less encoding time. Complete canonical/packed hashes match the prior
checkpoint, all K=11,926,435 systematic positions and 16 independent curves pass.
Native initialization takes 0.155 ms, with 141,904 native bytes; each encoding
makes five calls, zero interface buffer copies, and about 47–57 microseconds of
observed boundary interval beyond native kernel time. Peak process RSS is 2.257
GiB for both paths; Go/native allocation accounting and profiling limits are
explicit. The 432 MiB preprocessing cache remains intact and was reused.

Verification: default and tagged full `go test ./...` pass in isolated copies;
CGO_ENABLED=0 fallback, tagged vet, Go race checks and native ASan/UBSan pass.
Small complete words match the dense oracle; tests cover polynomial strategy
paths, invalid calls, concurrent buffers and Close. The s=4 runner smoke and s=13
analyzer pass, including binary/cache/source/evidence checksums. Initial smoke
launches stopped before measurements for missing fixture provenance and sandboxed
CPU metadata; the completed smoke uses recorded provenance/read access. Go race
checks do not instrument C++ memory. Native support is currently s=2..13 on the
validated macOS/arm64 host; other architectures, larger instances and the
cryptographic/application pipeline remain separate work. No default-backend
selection or production-security claim is made.

README checked and updated with native build/API and measured results. Updated
[encoder contract](../docs/encoder-contract.md),
[algorithm notes](../docs/encoding-algorithms.md), package READMEs and
[experiment index](../experiments/README.md). Existing historical experiment
snapshots remain unchanged. Current native FFT work is the main measured tuning
target; preprocessing remains Go as requested.
