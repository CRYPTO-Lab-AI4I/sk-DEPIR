# SCALING-161 — completed

2026-11-22T07:45:28.142828+00:00; Codex.

78662 independently passed5arms/10rank outputs, complete input/source proof and
native/scalar output checks, small CUDA memcheck/leak with zero errors/leaks.
Actual producerca57b77 (implementation868a580);750source/63original evidence/
76cache identities reconciled. Controller178.915s on HPC gnode08/gnode10,
2H200/4CPU/8GiB RAM. Normal RSS390.715MiB/process, small sanitizer415.520MiB;
post-call product/factor GPU2.356/3.103MiB/card are separate observations, not
aggregate physical peaks. Cache21.002MiB. Largest routed/peer net product
0.601303/0.597966s; input assembly32.968/30.703ms. One sample/arrival skew does
not establish a stable penalty/speedup. Removed131068-byte full peer allocation
per rank and halved aggregate input traffic262136→131068bytes.
[Report/evidence](../experiments/distributed-root-routing/report.md).

Updated main [reference](../docs/encoding-status.md), README, experiment index,
storage design and scaling plan. Retained raw results/logs/profiles and checksum
summary; caches outsideGit preserved. No arithmetic/code change or new job.
README checked/updated. Independent analyzer and3review regressions passed;
source/doc whitespace checked. No GPU rerun needed for evidence/docs updates.

User asked what remains: reusable distributed root-tree/repeated/truncated
Newton/generator executor; partitioned discovery/grid/support; explicit s18
arithmetic admission; coordinated restart/artifact publication; assembled genuine
small pipeline and representative scaling/resource validation before fulls18.
Current component gates still use s17 field and bounded synthetic supports.
No completion percentage or full runtime estimate inferred. Next concrete
implementation is complete rank-owned root/product execution and Newton/generator
composition through the existing shared schedules.
