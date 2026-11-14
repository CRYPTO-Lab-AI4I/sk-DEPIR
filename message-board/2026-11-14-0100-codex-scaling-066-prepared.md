# Wide preprocessing implementation ready for bounded validation

- Timestamp: 2026-11-14T01:00:44.072485+00:00
- Agent: Codex
- Work: SCALING-066
- Status: handoff (implementation checkpoint before validation submission)

Implemented [one shared preprocessing procedure](../docs/preprocessing.md) for
exact discovery, root construction, normalization, reciprocal and full identity
verification. Native32/64 executor selects width from s, reuses sparse plan
construction and existing base-polynomial algorithms, batches roots using gap
recurrence/trace/norm, and enumerates ternary weight classes once. Existing
GPU32 scaling preparation uses the shared root/reciprocal flow; the new explicit
CUDA32 adapter shares discovery and public-code orchestration. CUDA17 remains a
separate wider/distribution gate and is rejected explicitly, without fallback.
No full s17 support/preparation/cache or faster-server claim.

Changed native plan/API/preparation files in `encoding/liftrs-cpp/`, discovery,
polynomial dispatch/product ownership, small lookup cap17 and new preprocessing
session/artifacts/tests in `encoding/liftrs-shared/`. Support and normalized-root
artifacts bind trusted provenance/digests and preserve canonical representations.
One local native product returns its owned result directly instead of allocating
and copying a duplicate full Go result; overlap-add remains accumulated. No added
transforms/transfers or slower memory mode. Product operands are released before
final spectrum allocation. Explicit per-array/native caps are not aggregate RSS.

[Validation report](../experiments/s17-preprocessing/report.md) links the runner,
independent native driver, analyzer and [local checks](../experiments/s17-preprocessing/static-checks.json).
Local C++ syntax, Go formatting, Python/Bash syntax and diff checks pass; both
native widths pass independent complete small-domain tensor/Horner/product/good
set checks with one/four workers. Binary retained outside Git. Go/native suites,
CUDA build, race, ASan/UBSan and profiles await bounded Slurm validation. No full
s17 runtime estimate follows. Small native check is not profiled; job profiles
cover preparation and regressions. Existing caches and unreviewed jobs unchanged.

README checked and updated with source support/validation/CUDA limits; shared
and native READMEs, encoder contract, code-data guide, scaling plan/reference and
experiment index updated. Fast native s17 three-array payload384GiB, additional
Go arrays/overhead and unknown peak/time are explicit. Existing table timings
remain historical; native timing and weight-class discovery calibration unknown.
No change to client defaults, protocol, database packing, security or TRL5 claims.

Next submit the CPU-only4core/8GiB bounded validation, record ID/results and yield
without monitoring. Job74419 remains uninspected. Full CUDA17 and reviewed full
preprocessing admission still follow successful bounded evidence.
