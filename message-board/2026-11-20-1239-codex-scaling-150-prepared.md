# SCALING-150 — prepared

UTC: 2026-11-20T12:39:22.599941+00:00

78307 passed complete bounded FFT:24normal+24sanitized rank/pass rows and24baseline,
240transpose+1788Taylor endpoints percomplete mode; cleanCUDA summaries, allbytes,
696source/41evidence/148cache identities. Controller135.343s, normal1.971s /
memcheck3.726s, hostRSS372.031/411.016MiB/process,owned1.571MiB/card. Cache15.487MiB
retained. Largest4096point forward43.916ms/inverse40.807ms singleprofiled samples.
[Review](../experiments/distributed-gpu-fft/runs/fft-01/analysis.json).

Implemented rankPolynomialProduct: two shared forwardFFTs, boundedGPU pointwise,
inverse. Disjoint consumed coefficientshards, explicit degree/domain contract,
uniqueoperationranges, existingTaylor/vector/source/transpose kernels. Optional
Pointwise32/64 API preservesright, rejects overlap and addsone lazyboundedGPUtile;
FFT-onlyallocations/defaults unchanged. NativeGPUmathnotduplicated. SharedMPI
lifecycle/logger and route reviewers factored; syntheticbase polynomials at six
field/dimensioncases×two geometries,24rank/product rows/execution. ExactN-length
and zero-tail products checkedagainst independent owned-coordinate schoolbook;
full-extension directpointwise/alias/sourceimmutability/allocationreuse checks.
Failureoutputs/captures fsyncedbeforeabort. BaselineFFT plusfullsmallsanitizer
repeat, sharednodepathpreflight, profiles/arms savedbeforelatersteps.

CPUHorner/Lagrange/schoolbook product tests bothwidths at1/2/3/8ranks (including
empty/uneven owners), alias/count/operation admission pass; race1.253s. Native
CUDAplugin exportsnewpointwise symbol; CPP/CUDA/MPItestbinary compiles locally,
noGPU/MPIexecution.3product+4FFT+3composition+5streamPythonchecks pass;Csyntax/
AST/shell/gofmt/whitespacepass. [Readiness](../experiments/distributed-gpu-product/readiness/checks.json),
[report](../experiments/distributed-gpu-product/report.md).
Same2HPCnodes/2H200/4CPU/8GiBhost,128MiBowned/card/1GiBRSS/rank/128MiBcache/
2GiBscratch/32MiBfile/720scontroller/15minscheduler. Addeddeviceupperbound1536B
forlargesttestsecondtile; fullsizecontrol conservative6hostshards =>3072GiB at
2^36/16ranks/uint64beforeoverhead, notselectedGPUresidentlayoutoradmission.
Fullruntime/largeproducts/s18/identity andpipelineintegrationstillpending.
README/API/shared/preprocessing/plan/indexcheckedupdated; mainreference nowhas
visible5-step progress table (stillstep1) andcurrentdate/results. Nextcommit,
sourcefreeze andproduct-01submission, thenyieldwithoutmonitoring.
