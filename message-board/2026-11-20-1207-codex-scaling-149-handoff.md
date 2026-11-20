# SCALING-149 — handoff

UTC: 2026-11-20T12:07:56.674410+00:00

Submitted **78307**, fft-01, implementationbc6a88e. Exactjob-startsnapshotrevision/
hashesinrawresults; handoffchangesonlydocs/excludedmetadata.
[Report](../experiments/distributed-gpu-fft/report.md),
[receipt](../experiments/distributed-gpu-fft/submission.json).
Results/profiles experiments/distributed-gpu-fft/runs/fft-01/;
logslurm-78307.log; cache .cache/liftrs/distributed-gpu-fft/fft-01/.
Same2HPCnodes/2H200/4CPU/8GiBhost,128MiBowned/card/1GiBRSS/rank/128MiBcache/
2GiBscratch/32MiBfile/720scontroller/15minscheduler. CPU/race/sourcepathpreflight,
separablebaseline, complete small forward/inverse FFT/native+scalarHorner oracle,
fullsmallCUDA sanitizer, exactTaylor/transposeproof/outputreview, retainedprofiles/
phase logs andpassbarriers. NativeCUDAplugin andCPP/CUDA/MPItestbinaryprecompiled
locally; finalCPU/race/12Pythonchecks passed. NoGPU/MPIexecutionlocally orpostsubmit
monitoring/resultsinspection. README/mainreference/plan/preprocessing/CUDA/shared/
index current. Fulls18field/products/workflow/productionlayout stillpending.
Waitforusercompletionping beforedependentwork; nofullruntimeextrapolation.
