# Wide scalar gate passed

- Timestamp: 2026-11-13T15:55:46.796506+00:00
- Agent: Codex
- Work: SCALING-050
- Status: completed

Job74290 passes both TestWideOrbitArithmetic and TestTableFreeFields on gnode17/Go1.27.3/Linux AMD64/CGO disabled, no guard failures. [Reproducible analyzer](../experiments/s17-arithmetic/analyze.py) reconciles187 Go source files to5b9e706,8 evidence files and the retained binary. [Analysis/results/report](../experiments/s17-arithmetic/report.md#reconciled-result) document commands and narrow scope; profiles/cache retained.

Updated scalar experiment report, scaling plan/reference and root/shared READMEs. Scalar correctness/order gate passes; native64-bit CPU arithmetic/plans and bounded transforms are the next implementation gate. Public encoder/cache/native/CUDA admission stays s≤16. No full-s17 runtime/memory claim or numeric table updates.

Verification: analyzer rerun passes, Python syntax/local links and authored-document diff checks pass. Raw archived compiler/Slurm logs retain original whitespace to preserve recorded digests. README checked/updated. No scheduler/process/log inspection of other jobs, new jobs, or rerun arithmetic tests; only completed74290 evidence reconciled after user ping.
