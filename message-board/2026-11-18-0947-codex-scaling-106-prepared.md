# Shared root boundary and transport gate prepared

- Timestamp: 2026-11-18T09:47:44.134154+00:00
- Agent: Codex
- Work: SCALING-106
- Status: Prepared

Updated active plan and actual scaling reference for distributed prototyping:
RAM/direct messages for intermediate exchange, separate large buffered persistent
writes, per-file ownership, controlled writer concurrency, measured write/fsync/
readback/calls/bytes before full publication. No link to the investigation report
was added. README, preprocessing docs and experiment index checked/updated.

[Shared root/generator split](../encoding/liftrs-shared/preprocessing_core.go)
retains arithmetic, root/checkpoint identity, normalization, immutable saved roots,
full independent identity and input-recovery snapshot semantics. Generator helper
uses K/P instead of support; captured support checksum avoids a support closure.
Public callers still own support, CLI release before Newton is not implemented,
no forced GC or measured memory saving. Existing CPU/GPU execution stays default.
[New tests](../encoding/liftrs-shared/preprocessing_root_test.go) pass0.046s: s3/4/6,
1/2/4 complete Frobenius-orbit subsets merge to whole P, immutable input and failed
root-publication callback prevents reciprocal. Partition tests are sequential,
not distributed execution or speed measurements.

[Root validation runner/report](../experiments/preprocessing-root-boundary/report.md)
reuses single-node4H200/4CPU/8GiBhost/6GiBRSS/4GiBGo/4GiBowned-device/2GiBcache/
4GiBscratch,20min scheduler admission. Adds native root/recovery/export/race tests,
GPU small/export and memcheck; retains matching native/GPU binaries and profiles.
Inherited complete CLI/race/ASan/vet/restart/sanitizer gates remain. Submit this
within existing budget, reportID/results then yield; do not monitor.

[Two-node probe/report](../experiments/preprocessing-transport/report.md) compiles
with installed OpenMPI4.1.8 (Slurm resource plugin). Module query emits sandbox
socket-permission warning; compile succeeds, runtime/fabric untested. Zero-GPU,
2nodes/1rank+2CPU+4GiBrequest per node,64MiB data/rank,1GiBRSSguard/rank,
300sMPI/10min scheduler. No array files, direct Alltoall and inverse, full34bit
pattern/order checks,1/8/32MiB and1warm+3samples each, remote-host/coverage checks.
Local output file caps32MiB; bounded metadata/binary storage only. MPI stays optional
server experiment, CPU builds unchanged. Two review tests, AST, bash syntax,
runner checks and whitespace pass. MPI execution is pending explicit second-node
budget agreement per AGENTS.md; async question sent with exact runner/limits.
No actual multi-node/GPU transport or root parallelism claimed. The separate
persistence workload remains a later prepared/reviewed gate, not silently launched.
