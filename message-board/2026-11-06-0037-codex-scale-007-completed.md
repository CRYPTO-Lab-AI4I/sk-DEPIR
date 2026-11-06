# SCALE-007 completed

Timestamp: 2026-11-06T00:37:49.180504+00:00. Agent: Codex.

Implemented compact [support membership](../encoding/liftrs-shared/exponent_set.go),
[bounded flat factor batches](../encoding/liftrs-shared/factor_product.go), private
[convolution scratch reuse](../encoding/liftrs-shared/orbit_product.go) and
[Newton-result reuse](../encoding/liftrs-shared/orbit_polynomial.go).
[Preparation orchestration](../encoding/liftrs-shared/scaling_experimental.go)
releases scratch before spectrum generation; ordinary encoders retain no mutable
preparation workspace. Production arithmetic and outputs are preserved.

The [s=14 diagnostic](../experiments/preparation-checkpoint/report.md) passes full
P/G/spectrum hashes against the retained originals: preparation 34m 32.50s,
whole-process peak RSS 9.015 GiB, maximum checkpoint reachable heap 5.418 GiB.
Profiles, heap/RSS/power observations, exact measured source snapshots and
[analysis](../experiments/preparation-checkpoint/s14-retry/analysis.json) remain
available. This reused the support basis, forced internal GC checkpoints and
profiled execution; comparison with historical first-use RSS/time is not a
controlled speedup measurement. No database encoding or s=15 run was added.

The first large attempt stopped on AC loss at 40% battery due an overly strict
controller rule. Its incomplete evidence remains separate. The corrected
[runner](../experiments/preparation-checkpoint/run.py) continues through ordinary
power changes, pauses at low battery and resumes on AC, preserving memory.
The successful retry stayed on AC with no pauses. Original constants are untouched;
new local cache material is about 142 MiB, outside Git. Native samples are compressed
losslessly in the experiment with raw copies retained locally.

[Verification](../experiments/preparation-checkpoint/validation/verification.json):
both full Go/native suites in an isolated copy, focused race checks and vet pass.
Smoke fixtures and the full s=14 constants agree; four power-policy tests pass.
[Final checks](../experiments/preparation-checkpoint/final-verification.json)
resolve 69 measured source files and validate evidence, binary and artifact hashes,
Python compilation, changed-document links and diff whitespace. Full suites were
not repeated after diagnostic-only progress/controller changes.

README check: updated [root README](../README.md),
[scaling design](../docs/encoding-scaling.md),
[shared implementation README](../encoding/liftrs-shared/README.md) and
[experiment index](../experiments/README.md). The completed measurement supports
reviewing s=15 admission on the 48 GiB host, but does not prove a 32 GiB RSS bound.
Next: review guards and storage/time budgets, then run guarded s=15 discovery,
preparation and encoding with retained profiles/cache. No guards were extended
in this work item. Changes remain uncommitted.
