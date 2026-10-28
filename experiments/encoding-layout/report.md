# Encoding directory separation verification

This is a correctness/build checkpoint, not a performance experiment. The
[encoding guide](../../encoding/README.md) defines shared orchestration, Go-only
execution, C++ execution/binding and archived RM code. Preprocessing remains Go
for both paths; field arithmetic is shared through `liftrs-shared/field`.
The extraction preserves recursion, fixed maps, polynomial strategies and artifact
formats. Go plans borrow prepared rows and values and add only small execution
views; no new database-sized copies or cross-language calls are introduced.

## Reproduce

[check.py](check.py) runs the two full suites in an isolated checkout because
legacy tests write CSV files, builds the archived RM encoder from source and
checks each backend against immutable pre-move seed-71 s=4,t=2 hashes. It collects
profiles during the small smoke preparation and both encodings; temporary tiny
preprocessing is disposable, while the expensive s=13 cache is retained intact.

```sh
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  PYTHONPYCACHEPREFIX=/tmp/depir-pycache \
  python3 experiments/encoding-layout/check.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

[results.json](results.json) records source hashes and commands; [checks.log](checks.log)
records outcomes. The base revision is c13d034 with the uncommitted native
integration and layout changes identified by the manifest. Environment: Go
1.27.3, Apple Clang 21, Apple M5 Pro arm64, macOS 26.6.2, 48 GiB RAM. Timings in
check logs are operational durations, not optimized latency comparisons.

## Checks

- CGO_ENABLED=0 builds/tests the complete Go path.
- Tagged shared/native tests cover oracle agreement, all polynomial strategies,
  concurrent independent buffers, malformed plans and Close.
- Tagged vet checks encoding packages and the PIR compatibility bridge.
- Default and tagged full suites run with a newly built archived C++ RM process.
- Fresh small preprocessing feeds both backend paths. Both canonical and packed
  output hashes must match the original smoke checkpoint, all 97 systematic
  symbols and 16 independent curves. [Go events](go-events.jsonl) and
  [native events](cpp-events.jsonl) retain the results.
- Original s=13/native and kernel-comparison analyzers verify their recorded
  source, binary/cache and result checksums. Historical manifests/results are
  unchanged. Content-addressed source snapshots preserve measured uncommitted
  files that changed during extraction.

The [initial vet check](initial-checks.log) found unkeyed quadratic-element
literals after the public type became an alias from another package; these were
changed to explicit A/B fields. [Initial results](initial-results.json) retain
that failure. A [first passing run](first-passing-results.json) preceded a final
adjustment preserving the zero-value RM bridge; the final results cover that
adjustment and expanded source provenance.

No s=13 timing rerun was needed for this layout change. The documented performance
numbers retain their original source versions and measurement boundaries. These
checks do not establish a post-move speedup, new parameter support or security.
The Julia fallback was relocated and its includes/project selection updated, but
was not executed: compatibility tests use the freshly built C++ RM process. The
only remaining RM-specific code in `pir/` is the explicit compatibility bridge;
RM algorithms and process handling are under `encoding/rme_legacy/`.
