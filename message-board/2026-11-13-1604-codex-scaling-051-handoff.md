# Native64 transform validation submitted

- Timestamp: 2026-11-13T16:04:37.966789+00:00
- Agent: Codex
- Work: SCALING-051
- Status: handoff

Implemented the shared native FieldT/ProductT/TermT/TransformT32/64-bit specializations in encoding/liftrs-cpp/transform.hpp. Existing aliases, layout and C ABI admission remain32-bit. Wide multiplication casts before component packing; offset scans/degrees/products use the chosen width. No wide ABI, matrix schedule, cache loader or full s17 encoder added.

Added a bounded independent Go fixture generator and [native experiment](../experiments/s17-native/report.md):60 fixtures at s8/16/17/18/20 with sizes8/128/1024, high-bit ordered bases/offsets, dense or highest-degree witness coefficients. Check120 scalar/table cases in optimized and ASan/UBSan builds against independent polynomial evaluation, inverse/canaries and32-bit agreement; run legacy native scalar and Go shared/binding regressions with profiles. Root/native/experiment README, scaling plan/reference updated; no numeric full-encoding table cells changed.

Static C++ syntax checks pass for both fixture executable and full existing C ABI translation unit; Go fixture gofmt, Python controller --check/AST, shell syntax, local links and authored-source whitespace checks pass. Full runtime pending. Scalar74290 evidence from preceding completed work is preserved in source checkpoint6083f56.

Submitted job74295,4CPU/noGPU/8GiB requested RAM/6GiB RSS,2GiB retained cache/4GiB scratch,10m worker/15m controller/20m Slurm. [Submission](../experiments/s17-native/submission.json); results/profiles experiments/s17-native/runs/kernel-01/, Slurm log experiments/s17-native/slurm-74295.log, binary/fixture .cache/liftrs/s17-native/kernel-01/. No full-domain allocation or preprocessing. No active older-job inspection or post-submit monitoring; await user completion ping. Passing will establish bounded kernel correctness, not full s17 execution or a speedup.
