# Go/C++ kernels and full-size transform comparison completed

- Timestamp: 2026-10-28 20:59 UTC
- Work item: ENCODE-011
- Status: completed

## Publication, naming and provenance

Committed/pushed prior profile results, monitor fix and optimization/hardware/
language-label guidance as `0431f80` before this work. Renamed the existing Go
implementation to [encoding/liftrs-go](../encoding/liftrs-go/README.md), retaining
its declared `liftrs` package name, and added
[encoding/liftrs-cpp](../encoding/liftrs-cpp/README.md). Current Go imports,
examples, scripts, catalogue implementation paths and live Markdown targets now
use language-labelled paths. Historical message text/commands, measured JSON
hashes and snapshot scripts keep their recorded identity; only old Markdown
link targets follow the move. [Source verification](../experiments/source_evidence.py)
resolves exact old/mapped files or committed blobs instead of rewriting hashes.
README documents the experimental import-path migration. Go mathematical behavior
and serialized cache format are unchanged.

## Implementation and measurements

Added [Go opt-in fixture/benchmark helpers](../encoding/liftrs-go/language_bench_test.go),
[C++ kernels/FFT](../encoding/liftrs-cpp/transform.hpp),
[native driver](../encoding/liftrs-cpp/bench.cpp) and
[independent native tests](../encoding/liftrs-cpp/test.cpp).
The [report](../experiments/kernel-languages/report.md),
[runner](../experiments/kernel-languages/run.py),
[raw results](../experiments/kernel-languages/comparison/results.json),
[table](../experiments/kernel-languages/table.md),
[analysis script](../experiments/kernel-languages/analyze.py) and
[evidence manifest](../experiments/kernel-languages/evidence-manifest.json)
record source/compiler/host, shared-input/binary hashes, boundaries and limitations.

- Matched actual s=13 tables and seed-1313 inputs; each kernel call covers
  134217728 products, using 1/16 MiB batches, fixed/rotating tables, three repeats.
  C++ speedup is 2.28–2.48x for base products and 1.64–1.70x for extension products.
- Complete forward FFT at 1048576 elements: Go 0.6101 s, C++ 0.5326 s (1.15x).
- At 67108864 elements: Go 63.2141 s, C++ 50.1112 s (1.26x, ~21% less time).
  Complete output hashes match and both inverse round trips pass. Dense extension
  inputs are not a full systematic-codeword encoding workload.
- No allocation/I/O/process/FFI costs are included in these medians. No full
  native encoder, cgo integration, production-security or other-hardware gain is
  claimed. Container/compiler/inlining details differ; language choice stays open.

CPU profiles are retained for Go fixture generation and all timed Go calls.
Sandboxed native sampling was denied; logs preserve each attempt. A scoped
20-second external sample of our full-size native process succeeded; its
[profile/command metadata](../experiments/kernel-languages/comparison/forward-extension-fixed-67108864/cpp/external-profile.json)
and [reproduction script](../experiments/kernel-languages/profile-native.sh) are
linked by the report. Coverage is not identical to Go, and observer limitations
are explicit. Compiler vectorization diagnostics and an
[assembly check](../experiments/kernel-languages/check-codegen.py) of the actual
measured binary preserve evidence that calls were not removed from the pass loop.

## Validation, retention and documentation

[Small shared smoke checks](../experiments/kernel-languages/smoke/results.json)
match, with Go Horner evaluation and both inverse checks. The
[verification script](../experiments/kernel-languages/verify.py),
[commands/outcomes](../experiments/kernel-languages/verification.json) and
[log](../experiments/kernel-languages/verification.log) record passing renamed
Go/catalogue/planner tests, vet, independent native arithmetic/table checks through
s=13 under ASan/UBSan, and the isolated full Go suite. Validation ran after timing.
The production Go core only moved/imports changed; no redundant race rerun was
needed after its earlier passing checkpoint.

Large fixtures, outputs, binaries and export profile remain under ignored
`.cache/liftrs/language-s13/`, located/checksummed in results and
[retained-inputs.json](../experiments/kernel-languages/retained-inputs.json).
The earlier 432 MiB s=13 code cache is preserved and checksum-verified. Validation
copies now exclude `.cache` so reusable large data are not duplicated. No useful
cache/profile was deleted. Uniform peak RSS was not measured; reports state buffer
payloads rather than inventing a memory result.

README checked/updated, as were AGENTS map/naming, design/package/application docs,
experiment index/migration guidance, runners and catalogue paths. Historical
source verification, retained pprof source lookup after migration, changed Markdown
file targets, complete output hashes, evidence hashes and `git diff --check` pass.
New comparison and naming work remains uncommitted; the requested pre-work
checkpoint was already pushed. Next measure native batched-transform integration
in actual encoding, including setup/ownership/interface cost, before choosing a
backend; alternative Go/native optimizations remain eligible.
