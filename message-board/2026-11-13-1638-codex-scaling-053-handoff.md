# Wide interface/matrix validation submitted

- Timestamp: 2026-11-13T16:38:37.381328+00:00
- Agent: Codex
- Work: SCALING-053
- Status: handoff

Committed/pushed validated native results as01cef49, then implemented source checkpoint71db6a7. New encoding/liftrs-cpp/wide.go + wide_api.h/cpp provide separate WidePlan/WideMatrix handles with copied immutable metadata, canonical validation, overlap/count checks, borrowing/no retained Go pointers, lock-protected lifetime/accounting and canonical-grid matrix execution. Existing32-bit c_api.cpp shares the same extracted Taylor/transpose/row/column code in matrix.hpp. Public comments and runnable example added; existing old ABI and shared full encoder/cache admission unchanged.

[Experiment/controller/report](../experiments/s17-matrix/report.md), [submission](../experiments/s17-matrix/submission.json). Tests cover independent affine/small dense matrix oracles, error/close/recovery/partial overlap, 32K/64K within-block parallel Taylor branch,34-level metadata without domain allocation, Go lifetime race,112 optimized and112 direct C sanitizer matrix cases, previous recursion120-case checks and32-bit regressions. Explicitly retained matched profiled Go executables in shared cache. Validation pending; no whole-s17 or performance claim.

Static C++ syntax passes for old/new ABI and direct C check, Go files gofmt, Python --check/AST, shell syntax, local links and authored diff checks pass. Root/native/experiment README and scaling reference/plan updated with the public API and pending scope. No protocol/cryptographic or preprocessing changes.

Submitted job74312,4CPU/noGPU/8GiB requested RAM/6GiB RSS,2GiB retained cache/4GiB scratch,10m worker/15m controller/20m Slurm. Results/profiles experiments/s17-matrix/runs/matrix-01/, log experiments/s17-matrix/slurm-74312.log, binaries/fixture .cache/liftrs/s17-matrix/matrix-01/. Existing jobs untouched; no post-submit polling/log reads/result analysis. Await user completion ping.
