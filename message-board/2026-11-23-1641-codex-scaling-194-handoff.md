# SCALING-194 handoff: resource agreement pending

UTC: 2026-11-23T16:41:50.063510+00:00. Agent: Codex. Status: handoff.

82181passed and all evidence/profiles/logs committed. Named host admission and
explicit reserve implemented; no math/transform/transfer change or forced GC.
[Prepared eight-worker plan](../experiments/distributed-host-admission/report.md)
and exact runner/Slurm/analyzer/reference are reviewable. Frozenaff14a1 has939build
inputs/414matching compiled hashes;27frozen reviewers, ordinary CPU/race/tagged
admission/reference checks pass. CPU-native preflight matches108014historical bytes
plus actual s8/curve5 inputs. README checked, primary reference and plan/storage/
preprocessing docs accurately show prepared/not submitted; fulls18 unrun/unadmitted.

Do not submit before resource agreement: proposal2HPCnodes/2H200,8workers/8CPUcores,
16GiBrequestedRAM total (4cores/8GiB per node),1GiBRSS/worker,768MiB declared
runtime/native/GC reserve per worker. OwnedGPU32MiB/worker=128MiB/card,zeroCUDA-
pinned,128MiBretainedcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm,
excludegnode03. Only core/requestedRAM totals increase from4/8GiB; preserve RSS
headroom rather than tighten it. Estimate3–7min excluding queue; not fullsize or
8GPU throughput. Exact resource table/report ready for user's decision.

On approval, record admission, submit experiments/distributed-host-admission/run.sbatch
host-01, save exact receipt/log/results/revisions, register ignored completion
notification, commit and yield without polling. On completion independently
verify eight-worker phases/rounded group counts/logical checkpoint hashes and
final native/pilot bytes, source/control/shard/alias checks and actual native work.
Fullsize reserve/node packing/storage/time admission follows representative data.
