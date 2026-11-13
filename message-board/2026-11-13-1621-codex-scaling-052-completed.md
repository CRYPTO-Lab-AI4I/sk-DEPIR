# Native64 bounded kernel gate passed

- Timestamp: 2026-11-13T16:21:50.988223+00:00
- Agent: Codex
- Work: SCALING-052
- Status: completed

Job74295 passes120 optimized and120 ASan/UBSan cases, independent polynomial evaluation/inverse/canaries, overlapping32/64 agreement and invalid-shape checks. Legacy native scalar oracle/shared liftrs_cpp Go suite pass. Binding package builds but has no test files; report explicitly distinguishes this.

[Native analyzer/report/results](../experiments/s17-native/report.md#reconciled-results) verify410 source files to6083f56,20 evidence files and5 cache files. Extracted shared hash verification from scalar analyzer for reuse; scalar74290 reconciliation rerun passes unchanged187 source/8 evidence/1 binary identities. No arithmetic tests rerun. Profiles, fixtures and binaries retained at reported paths.

Updated root/native READMEs, scaling reference and execution plan to mark bounded native recursion passed. No new numeric full-encoding cells; wide ABI, parallel matrix schedule, code-data/preprocessing and full s17 execution remain pending. Verification: both analyzers pass, Python syntax/local links and authored-document checks pass; raw original logs retain whitespace/digests. README checked/updated. No other jobs/results inspected, no new job submitted or runtime budget changed.
