# SCALING-191 prepared

UTC: 2026-11-23T10:39:12.586290+00:00. Agent: Codex. Status: prepared.

Verified receipt82100 and source/evidence/cache seals; all three arms passed,
including committed final manifest and native/pilot byte checks. Saved raw results,
profiles, Slurm log and analysis under [publication experiment](../experiments/distributed-publication-commit/report.md).
Controller158.073s; assembled8.510s/378.824MiB RSS per worker; timing boundaries differ.

Implemented coordinated durable assembled compact prefix commits, full source
metadata binding and sealed support hard-link alias; legacy manifests still load.
Existing kernels/numerical work/default observer path unchanged. Added source/hash/
immutability/exclusive-metadata tests and fresh-process file-reload adapter.
[Checkpoint report](../experiments/distributed-workflow-checkpoints/report.md) links
code, runner, independent reviewer and resource/storage/runtime trade-offs.
Updated primary scaling reference/current rows, plan/storage docs and README;
README check: public progress changed, so its current summary was updated.

CPU tests and race tests passed; CUDA/MPI tagged compile and CPU admission tests
passed without initializing either runtime. All24Python reviewer suites, script
syntax and authored diff whitespace checks passed. Exact frozen-source/native
input preflight remains before same-budget submission. No fullsize or arithmetic
resume claim. No new GPU job submitted at this point. Next validate committed
checkpoints under existing two-node/two-H200 caps and hand off, then integrate
assembled arithmetic resume and command/resource admission.
