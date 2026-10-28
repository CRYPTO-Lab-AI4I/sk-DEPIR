# Encoding responsibilities separated; RM archived

- Timestamp: 2026-10-28 22:30 UTC
- Work item: REF-004
- Status: completed

[Encoding guide](../encoding/README.md) now distinguishes three actual code paths:
`liftrs-go` contains pure-Go transform recursion/fixed-map execution;
`liftrs-cpp` contains C++ execution, C ABI and cgo ownership binding;
`liftrs-shared` contains the full public API, code discovery/preprocessing,
scalar arithmetic, polynomial strategy, assembly, oracle and native adapter.
Common arithmetic is in `liftrs-shared/field` and public types remain aliases;
prepared Go execution views borrow immutable tables. No per-element callback or
new database-sized copy was introduced. Public imports move to liftrs-shared;
all active Go callers/catalogue references/scripts and current docs follow it.
Both complete execution paths remain Go-orchestrated; this is not a full C++ port.

All previous root-level encoding Julia/C++ RM sources/helpers, Makefile,
dependency environment, benchmark, `encoding_time.png` and retained local binary
are under [rme_legacy](../encoding/rme_legacy/README.md). The RM subprocess driver
also moved there. A thin [PIR compatibility bridge](../pir/legacy_encoder_bridge.go)
keeps the old file-based RM placeholder functional, including zero-value process
initialization, common permutation/masking and serialized intermediate-file use.
Julia includes now follow source locations and fallback selection uses the
archived project; Julia execution itself was not tested. The legacy C++ binary
was rebuilt in isolated validation, not silently reused as correctness evidence.

[Verification report](../experiments/encoding-layout/report.md) links the runnable
checker, source manifest, logs, profile-bearing smoke events and all outcomes.
CGO_ENABLED=0 tests, native tagged tests, tagged vet and both complete Go suites
pass. Both backend smoke outputs match pre-move canonical/packed hashes, all 97
systematic elements and 16 curves. Original s=13/native and kernel analyzers pass
without repeating expensive benchmarks. An initial vet failure after field-type
aliasing was corrected with named A/B literals; earlier logs are retained.

Historical measured manifests/results and snapshot scripts remain unchanged;
[content-addressed source snapshots](../experiments/source-snapshots/README.md)
preserve the previous uncommitted native sources. Source resolution now follows
moves/snapshots/Git; current source capture includes shared fields and Go kernels.
Existing expensive caches/profiles/binaries are retained. No new timing claims or
larger-parameter support are inferred from moving files. Current source hashes,
Python syntax, changed-doc local link targets and diff whitespace checks pass.
Two unrelated historical message links to the deleted old sk_DEPIR.pdf remain
historical references rather than being silently retargeted to a newer paper.

README checked/updated: directory map, pure-Go/native build requirements,
legacy build location, API migration and historical timing boundaries. Updated
AGENTS repository map, design/code links, package guides, experiment index and
active runners. Previous native integration work remains intact and uncommitted;
no commit or push was requested for this item.
