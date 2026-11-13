# Wide encoding baseline reconciled

- Timestamp: 2026-11-13T23:13:01.116030+00:00
- Agent: Codex
- Work: SCALING-059
- Status: completed

Job74412 passed. New [analyzer](../experiments/s17-encoding/analyze.py) reuses
sealed evidence verification and checks actual targeted/regression/direct C logs.
Command: `python3 -B experiments/s17-encoding/analyze.py encoding-01`. Verified
438 source identities against0db6ec6,24 evidence and18 cached binaries/fixtures.
All commands/guards pass;12 optimized and12 ASan/UBSan synthetic cases, real
s3/4/6/8 reference/dense/systematic/packed/ownership and Go race/ASan checks,
raw width/count checks and shared Go/native regressions pass.
[Report/results/analysis](../experiments/s17-encoding/report.md) retain exact
methodology and scope; caches/profiles preserved. No full s17 claim or forecast.

Changed analyzer/results/report, root README, both package READMEs,
`docs/scaling-code-data.md`, scaling plan/reference and board. README checked
and updated for bounded gate; changed documentation links and diff whitespace
pass. Recommended next is shared native CPU encoding sequence, preserving
width/memory-specific modules, APIs/defaults/accounting; validate both widths,
existing modes and representative timing before further integration. No refactor
or new validation job started by this advisory/results review. Other jobs and
unreviewed results untouched; no scheduler/process inspection performed.
