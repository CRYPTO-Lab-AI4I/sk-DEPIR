# SCALING-166 — prepared

2026-11-22T13:13:47.428979+00:00; Codex.

78917 PASS:16 normal/8 small sanitized rank outputs,zero CUDA errors/leaks in4
sanitized processes;all780 source/51 evidence/52 cache identities verified.
Controller2min25.876s,normal maxRSS378.676MiB/process,sanitized414.059MiB,
maximum reported ownedGPU3.118MiB/card (contexts/transients additional),cache16.478MiB.
[Report/raw evidence](../experiments/s18-field-admission/report.md) retained;
updated primary scaling table/main s18 preprocessing cell and docs/README.
Fulls18 remains unrun.

[Bounded coordinator integration](../experiments/distributed-coordinator-stream/report.md)
prepared: optional retained native-reference inputs,streamed ordered support without
full coordinator archive/native P/G/spectrum arrays. Fixed bounded metadata/file
checks,exact SHA/size/order/count/entropy/sample/canonical comparisons;fullP*G
identity and all FFT/MPI route proofs remain. Existing native/gather defaults stay.
Changed shared GPU test drivers and a test-only bounded-reference helper;reused
restart controller with optional reference/reviewer arguments,focused input/cgo
launcher regressions,new reproducible experiment/report/readiness/source input seals.
No CUDA arithmetic change;no repeated sanitizer. One cold+fresh-process-resume job.
Native reference setup0.19s including fsync,files match78915;69,746bytes reusable
inputs separate from distributed runtime. Memory-time/I/O trade-offs documented;
no stable speedup/RSS claim or fulls18 native-reference requirement.

[Readiness](../experiments/distributed-coordinator-stream/readiness/checks.json):
CPU/race,new metadata/file/order/short-write checks,alltags build,native fixture/export,
reviewers and Slurm syntax pass. Same2HPCnodes/2H200/4CPU/8GiB RAM,128MiBowned/card,
1GiBRSS/rank,128MiBcache/2GiBscratch/32MiBfile,720s/15min,no increase.
README checked and updated;docs/preprocessing.md,encoding-status.md,scaling plan,
reports,board updated. Temporary automatic-review usage limit blocked final build,
then user resumed and approval/build succeeded;no remaining block.
Freeze source,submit coordinator-01,record receipt and yield without monitoring.
