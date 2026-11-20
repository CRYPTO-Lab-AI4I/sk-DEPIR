# SCALING-148 — prepared

UTC: 2026-11-20T11:29:04.079202+00:00

Implemented optional dedicated CUDA VectorPreparation owner/API (typed synchronous
interleaved batches, immutable domain/capacity, optional ABI, no mixed resident
product/encoding storage), shared rank vector packing and two-transpose composition,
independent CPU ordering/ownership tests and two-node GPU/native-oracle experiment.
Existing vector/field/transport kernels reused. Moved stage definition inside its
Resident namespace for ordinary plugin compilation; component includes remain
pragma-once compatible. Input consumed; hostcontrol may own callerinput+two shards
plusboundedtiles, explicitlynotselectedproductionlayout/no default changes. At
full2^36/16ranks/64bits hostpayloadprojects1536GiB aggregate, extra2Nslots each H2D/
D2H perforward; runtimeunknown. FullFFTstillneedsdistributedTaylor conversion,
thenproduct/s18admission; nofullFFTclaimfromseparable stages.

CPUforward/inverse tests with noncommuting reference pass32/64bits across1/2/3/8/16
ranks, odd/empty/tailowners, invalidadmission/backenderrors; racepasses. CUDA-free
Go/Cstageowner interfacebuild passes, Cheader syntax passes.3new+5priorPython
reviewer/launcher/stream tests pass; AST/shell/gofmt/whitespace checked.
[Checks](../experiments/distributed-gpu-vectors/readiness/checks.json),
[report](../experiments/distributed-gpu-vectors/report.md),
[runner](../experiments/distributed-gpu-vectors/run.py).
NativeCUDA translation/composedMPI/GPU execution deferred to >30s admittedjob.
Six smallfield/gridcases×two directions×two ranks=24normalmeasurements; entire
small suite repeated under memcheck. Largestgrid4096words,longestvector64words,
capacity192words. NativeCPUoracle builds independent subspaceproductbasis, one
CPUcolumn+localexpectedrows, everyoutputchecked/saved; exactpairedproofcoverage,
errorcapture, profilesafter normal, completedarms persisted. Sharedrankwrapper/
binary/library/cwdpreflight on every node; remoteTMPDIR=/tmp, UCXcachecontrol,
stagedMPI, bothstreams sanitizer summaries. Same2nodes/2H200/4CPU/8GiBhost,
128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch/32MiBfile,720scontroller/
15minscheduler. README/APIusage/CUDA/shared/reference/plan/preprocessing/index
checked/updated. Nextcommit/sourcefreeze and boundedsubmission, yieldunmonitored.
