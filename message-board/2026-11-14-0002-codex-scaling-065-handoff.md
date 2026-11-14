# CPU modularity refactor submitted for validation

- Timestamp: 2026-11-14T00:02:38.360790+00:00
- Agent: Codex
- Work: SCALING-065
- Status: handoff

Implemented source checkpointc3544fd. [Guide](../docs/cpu-encoding.md) maps the
common cyclic sequence and storage/output/width modules, barriers, ownership,
representation, code verification, diagnostics and intentional independent Go
references. Resident32-bit, compact/file/packed, wide64-bit and disk native
paths call one shared sequence. Reused power/primitive/fixed-product, pointwise,
scatter/parity and little-endian I/O. Kept C ABI compilation entrypoints using
private implementation headers; split Go bindings into plan/matrix/online/disk
with distinct-plan locking. Shared resident/file code decoder/verifier and
prefix/full metadata setup; no second payload read/domain copy.

Review fixes: closed/uninitialized resident workspace methods now return nil/
error without allocating/panicking; compact16-bit output writes use atomic
character stores when sharing original32-bit objects. Added no domain array,
transform, transfer, I/O pass or memory-mode/default switch. Runtime effects
await tests, including any loader/header overhead. Existing Go/dense/hybrid
references, cached data, encoding functionality and security limits preserved.
Optional job progress includes timestamped phase start/done/failure and
advancing block/batch/tile/vector counters; logging disabled in latency samples.
Formatting is whitespace/readability plus explicit helpers; temporary clang-format
18.1.8 in/tmp adds no dependency, exact Go build tags protected by configuration.

Changed native modules/bindings under encoding/liftrs-cpp; shared legacy loader,
code verifier, resident/compact/file adapters and module/code tests; relevant
READMEs, scaling plan/reference, code-data docs, new guide and
[experiment](../experiments/cpu-encoding-modules/report.md). Root/package READMEs
checked/updated. Earlier staged user logging/server-preprocessing/estimate/status
records were preserved in the checkpoint. No running-job artifacts inspected.

Static verification passed: GoFMT, C++17 -Wall/-Wextra syntax-only for C ABI/wide/
Taylor/helper drivers, Python AST/controller --check, sbatch bash syntax, changed
documentation links, Go-native build tags and staged diff whitespace. Functional
checks are pending job74419:4CPU/8GiB requested/6GiB RSS guard,4GiB Go soft limit,
2GiB new retained cache/4GiB scratch,600/900/1200s deadlines, no GPU.
[Submission](../experiments/cpu-encoding-modules/submission.json).

Validation covers complete CPU modes on real small codes and25/33-byte headers;
Go/dense/systematic/packed references, reuse/errors/caps/lifetime, Go-only/native
encoding/PIR/configuration suites, vet/race, Go ASan/C++ UBSan, direct native
helper/online/matrix/Taylor/wide oracles. Matched unprofiled s10 comparisons use
exact671bf5e CPU sources plus the identical benchmark file, alternating three
samples of three timed iterations per mode, plus code-load/open measurements.
Separate candidate profile retained. Scripts retain verified small preprocessing
and matched binaries. No fulls16 performance equivalence or s17 claim.

Results: `experiments/cpu-encoding-modules/runs/modules-01/`; retained cache:
`.cache/liftrs/cpu-encoding-modules/modules-01/`. Await user completion ping, then
run analyzer, reconcile source/baseline/cache/evidence and review timing ratios
before adoption. No polling, waits, log following or result inspection after
submission. Job74253 untouched.
