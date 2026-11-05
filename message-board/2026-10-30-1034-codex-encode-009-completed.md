# ENCODE-009 completed

Timestamp: 2026-10-30T10:34:27Z. Agent: Codex.

Added the complete native online step before s=15 admission in the
[scaling plan](../docs/encoding-scaling.md#next-step-complete-c-online-encoding)
and implemented it. The [C++ path](../encoding/liftrs-cpp/c_api.cpp) now performs
buffer setup, all five parallel matrix transforms, pointwise products and
canonical assembly through one call from the
[shared adapter](../encoding/liftrs-shared/scaling_matrix_cpp.go). Reusable code
preprocessing/loading remains Go; P/spectrum are borrowed without duplication.
The earlier hybrid method remains available. Preserve its measurement boundary
in the [older harness](../experiments/parallel-encoding/main.go) and
[first-use harness](../encoding/liftrs-shared/parallel_firstuse_probe_test.go);
corresponding historical reports now explain the explicit hybrid calls.

The new [experiment report](../experiments/native-online/report.md) links scripts,
raw results, CPU profiles, compressed native samples, source/binary digests and
retained caches. All seven full s=14 hashes match. Awake repeats give native
312.683/327.190 s versus hybrid 346.112/336.556 s (descriptive median reduction
6.27%), with about 4.59 GiB peak RSS for both. Earlier native repeats were
86.592/159.782/308.252 s; native stage sums and outer timers disagree. Preserve
all data and treat speedup as inconclusive, not a reliable performance claim.
No preprocessing was regenerated; larger-size forecasts remain unchanged.

Verification: focused oracle/hybrid, unit-message, invalid-input, ownership,
closed-handle, overlap and concurrent-workspace tests; Go race checks; vet;
[isolated full Go/native suites](../experiments/native-online/validation-final/verification.json);
[native ASan/UBSan fixture](../experiments/native-online/native-validation-final/verification.json);
[source/evidence audit](../experiments/native-online/verification-analysis.json),
Python compilation, local documentation links and `git diff --check` all pass.
Go race checks do not instrument native memory. Profiles remain enabled, with
raw native samples retained locally and lossless gzip copies in Git-sized results.

README check: read and updated [root README](../README.md),
[encoding overview](../encoding/README.md),
[shared API documentation](../encoding/liftrs-shared/README.md) and
[native documentation](../encoding/liftrs-cpp/README.md). Updated the encoder
contract, algorithm checkpoint and experiment index. Removed obsolete claims
that the matrix path still performs its scalar/assembly work in Go. No changes
to cryptographic/application encoding, public discovery limits or server behavior.
No commit/push was requested in this work item.
