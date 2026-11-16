# SCALING-090 — full encodingCPU76645 /GPU76646

UTC: 2026-11-16T23:21:51.934121+00:00; agent: Codex; status: handoff.

Committed recipeed0c026 andsubmitted independentCPU76645 andGPU76646. Bothsbatchaccepted; source pinneded0c026 viaGitarchive. [Receipt/resources](../experiments/s17-full-encoding/submission.json), [report](../experiments/s17-full-encoding/report.md). No scheduler/process/log/result polling orwaiting aftersubmission. CPU32workers/768GiB/noGPU; GPU32workers/384GiBhost/4H200112GiBowned/card. Each64GiBnewretained/16GiBscratch and46/47/48h limits; full caps/guards recorded. Noexclusive/nodenpin,excludegnode03. Userrequests both; predecessorpreprocessing/fullGPUvalidationalreadyverified, unchanged.

Results `experiments/s17-full-encoding/runs/first-01/cpu/` and`.../gpu/`; logs `experiments/s17-full-encoding/cpu-76645.log`, `gpu-76646.log`; output/profiles/matchingexecutables `.cache/liftrs/s17-full-encoding/first-01/{cpu,gpu}/`. Sameverifiedcode/seed71message,fastresidentpaths. Eachrun smallbackend smoke,onefullencode,all systematicpositions,16independentcurves,canonicalhash,34GiBpacking/full diskhash/1027positions. Timings/profile/logcost boundaries explicit. Fullruntime/peak/clientcapacity unmeasured;48h cutoff notforecast. NoCPU/GPU fullhashagreement claimed untilbothresultsreviewed. Noencodertransformcheckpointguarantee;completedphases/output/profilesretained.

LocalCPUoracle/9phase pipeline,17bitpackingnegativechecks/34bitfixedstep tests,GoCPU/CUDAtagbuild/vet,Pythonpolicy/assertions,shell/whitespacepassedbeforecommit. READMEchecked/updated tosubmissionstatus;scalingreferencekeeps s17encodingrowsunmeasured. Awaitcompletionping,thenreconcile source/evidence/cacheandcrossbackendoutputhashes. Noadditionaljobs automaticallysubmitted.
