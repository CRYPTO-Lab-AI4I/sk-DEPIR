# Table-gap jobs prepared

- Timestamp: 2026-11-13T10:26:15.382708+00:00
- Agent: Codex
- Work: SCALING-046
- Status: handoff preparation

Added [experiment controller/report](../experiments/encoding-table/report.md), batch submission/analyzer and22-case manifest:10 CPU,12 GPU, s10/14/15/16. Reuses retained code and existing implementations. Experiment-only Go controls select supported sizes, one/two repeats, and reuse the independent full packed verifier for CPU file-packed output. No public implementation/defaults changed. Historical R/HPC and L/laptop distinction remains explicit.

Changes: encoding/liftrs-shared/{scaling_table_test,s15_firstuse_probe_test,cuda_cached_test,cuda_workflow_test,cpu_disk_online_test,gpu_packed_cached_test}.go; experiments/encoding-table/; README.md, experiments/README.md, docs/encoding-{scaling,status}.md; board/index. README checked and linked pending experiment, without claiming results.

The user authorized increasing the s16 CPU disk storage guard when it was our own limit. Prior report identifies16GiB as a project experiment cap; df shows~13TiB free, quota -s no reported limit. Select64GiB retained storage for that case,16GiB for others; shared preprocessing is excluded. RAM16/32/128GiB requested by size, RSS12/24/96GiB, cgroup14/30/112GiB,32 cores, GPU40GiB (host-backed512MiB),110-minute controller/90-minute command/two-hour Slurm limits. CPU disk native workers16; other CPU modes32. At most one CPU and one GPU case run concurrently via independent afterany chains.

Verification: controller --check,22-case submission dry-run, Python AST, bash -n, gofmt, local documentation links and git diff --check passed. Full build/runtime controls and correctness verification execute inside the jobs; not yet validated. No long local tests. Backed fulls16 may hit deadlines; keep unmeasured cells until completion/reconciliation. Submission IDs will follow in a separate handoff; no monitoring.
