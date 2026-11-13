# Complete wide CPU integration submitted

- Timestamp: 2026-11-13T23:08:09.288312+00:00
- Agent: Codex
- Work: SCALING-058
- Status: handoff

Published prior validated code-data checkpoint83c8469. Implemented explicit
`ScalingCodeData.WithWideCPP`, capped workspace/resource accounting and complete
wide native five-transform cyclic schedule, source0db6ec6. Existing CPU/GPU
defaults unchanged. Source spectra keep their stored width; output owns dead
scratch:4N logical bytes retain8N backing. Previous words require extra RAM.
Relaxed atomic byte assembly avoids reused uint64 alias/shared-location races;
its latency is unmeasured. No extra transform, transfer, I/O or compaction pass.

Changed paths: `encoding/liftrs-cpp/wide_api.*`, new `wide_online*.go`,
`encoding/liftrs-shared/scaling_code.go` (shared metadata validation), new
`scaling_wide_cpp*.go`, both package READMEs, root README,
`docs/scaling-code-data.md`, `docs/encoding-scaling.md`,
`docs/encoding-status.md`, and the [reproducible experiment](../experiments/s17-encoding/report.md).
README checked/updated for explicit API, backing size, pending validation and
unchanged full-size admission. Scaling reference updated without measured-cell
changes. No full s17 cache/preparation or run is admitted.

Static checks pass: GoFMT, C++17 syntax-only for schedule/direct checker, Python
AST/controller --check, sbatch bash syntax, changed documentation links and
`git diff --check`. Longer functional checks are delegated to job74412,
[submission](../experiments/s17-encoding/submission.json):4CPU/8GiB requested,
6GiB RSS guard,2GiB retained cache/4GiB scratch,600/900/1200s deadlines, no GPU.
Results: `experiments/s17-encoding/runs/encoding-01/`; retained small verified
preprocessing/binaries: `.cache/liftrs/s17-encoding/encoding-01/`.

Tests cover real small-code Go/dense/systematic/packed agreement, P·G identity,
repeat/Close/GC ownership and errors, full shared Go/native regressions, raw
large-count metadata and high-value rejection, Go lifetime race, Go ASan/C++
UBSan and12 optimized/ASan/UBSan independent polynomial schedule cases.
Synthetic edge inputs are arithmetic fixtures, not claimed PIR codes. No runtime
result is asserted. No scheduler/process/log/result checks after submission;
older jobs/results left untouched. Await completion ping, then reconcile source,
cache and evidence hashes before using results.
