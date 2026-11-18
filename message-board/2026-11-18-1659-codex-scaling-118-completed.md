# Cross-node GPU roots77285 passed; primary reference current

- Timestamp: 2026-11-18T16:59:38.688915+00:00
- Agent: Codex
- Work: SCALING-118
- Status: Completed

Reviewed after user completion ping. [Report](../experiments/distributed-root/report.md),
[analysis](../experiments/distributed-root/runs/roots-01/analysis.json),
[raw results](../experiments/distributed-root/runs/roots-01/results.json),
[accounting](../experiments/distributed-root/runs/roots-01/accounting.txt).
77285 COMPLETED0:0 in2min36s (155.737s controller), sourcef80a685, HPCgnode16/17,
1H200/2CPU/4GiB requested host RAM per node. All15commands exit0. Verified621Git
source/20evidence/35cache seals without overrides;18normal and6small sanitized
rank measurements reconcile. Historical four-root digests, partition/race tests,
complete CPU/scalar witnesses, cross-rank sent/received polynomials and full
receiving-GPU merge/reference byte equality pass; both CUDA memchecks0errors/0leaks.

Largest synthetic s17 subset262144 exponents:1.179829s median full pipeline,
1.010393s local root,23.986ms exchange/wait/peer checks,101.660ms GPU merge.
Two timed passes after warm-up; rank-max phases not additive. Exchange includes
arrival skew (rank-local events~0.8ms vs~24ms), not isolated link bandwidth.
CPU reference/setup/persistence/profile reporting remain outside pipeline clock
but in complete job duration. Profiles mostly show native work under cgocall,
including the CPU reference; no GPU-instruction/native-rewrite inference.

Normal rank RSS401.129/410.523MiB, small memcheck maximum409.246MiB;
ownedGPU15.241MiB/card, no new pinned buffers. Sealed cache20.498MiB at
`.cache/liftrs/distributed-root/roots-01/`, retaining35files incl matchingbinary/
library,wrapper and complete normal/sanitized public arrays. Four CPU profiles
and all raw logs/results retained bytewise; no aggregate job RSS invented.

Changed README, CUDA/experiment READMEs, report,
[actual primary scaling reference](../docs/encoding-status.md) and
[plan](../docs/encoding-scaling.md), plus evidence/accounting/analysis and board.
README check: bounded status/hardware/timing updated, ordinary builds/defaults
unchanged; fulls17 CPU/GPU passed and fulls18 unrun cells unchanged. No new code,
job or budget increase. Twelve Python reviewer regressions pass; analyzer verifies
all hashes/arrays/memchecks; Python syntax, local doc links and edited-text
whitespace pass. No GPU rerun/full test suite for this results/documentation review.

Next planned measurement: buffered sequential checkpoint/final publication with
controlled writer concurrency; keep intermediate compute/exchange in RAM.
Scalable support/rank ownership, multi-level merge, generator/Newton/distributed
transforms and fulls18 memory/runtime remain outstanding. No full-size speedup,
capacity, default-mode or final GPU/node-count claim from this bounded gate.
