# SCALING-149 — prepared

UTC: 2026-11-20T12:03:38.620831+00:00

78302 independentlypassed24normal+24sanitized rank/pass rows,480proofendpoints,
688source/32evidence/99cache hashes, bothcleanCUDA summaries. Complete126.691s,
normal2.722s/memcheck3.826s; hostRSS372.113/412.508MiBperprocess,owned1.571MiB/card,
cache13.369MiB. Largest4096-word forward16.899ms/inverse15.142ms singleprofiled
samples; nofullsize extrapolation. Retained allraw evidence/logs/profiles.
[Review](../experiments/distributed-gpu-vectors/runs/composed-01/analysis.json).

Implemented generic rankTaylor source-window waves, bounded reusable hostsnapshot,
O(ranks)cursors, operative-source-only exchanges, pairedheaders/proofs, pass/level
barriers and full additive forward/inverse composition around existingvectors/
transposes. GPU sourceaccessor/kernel extractedverbatim fromHostTiled, shared by
bothadapters; nativestage hasoptionalTaylor metadata/window ABI, dedicatedowner,
canonicaldegreeorder and sourceimmutability checks. No duplicated recurrence,
CPUclientdependency/default/fullfieldadmission change. Target updates parallel
whereownersactive; unusedremainderprefix omitted toavoid races withtargets.
Controlhostlayout stillcallerinput+two hostshards; devicecapacityreused. Extra
Taylor transfer/launch/barrier costs unmeasured untilboundedjob; notproduction
layout/admission or assumedfree memorysaving. Fulls18stillunrun/unadmitted.

CPUserialglobal-source vswindow tests pass32/64slots,1/2/3/5/8/16/32ranks, empty/
odd/tailbounds and >2^32virtual coords. Complete smallfield transforms matchGo
Horner/directsubspacepolynomial and Lagrangeinverse at1/2/3/8ranks bothwidths;
finalrace1.212s. CUDAplugin and fullCPP/CUDA/MPItestbinary compile locally without
GPU/MPIexecution; compilerprobe socketwarnings in sandbox retained (exit0).
4FFT+3composition+5streamPython tests pass, shell/AST/gofmt/Csyntax/whitespacepass.
[Readiness](../experiments/distributed-gpu-fft/readiness/checks.json),
[report/runner](../experiments/distributed-gpu-fft/report.md).
Newrunnerreusesexistingcontroller, repeatsseparablebaseline then24fullFFTnormal/
24sanitized rank/passcases withnativefullFFTgold (<=4096-wordlocaldiagnosticgrid),
allbytes plusindependent exactsourcewindow/proofreview. Sharedpath/environment
preflight, CPU/racebeforeGPU, oldhostadaptercompile, profiles/armspersisted before
sanitizer, bothsummary streams. Same2nodes/2H200/4CPU/8GiBhost,128MiBowned/card,
1GiBRSS/rank/128MiBcache/2GiBscratch/32MiBfile/720scontroller/15minscheduler.
READMEchecked/updated; primaryreference/plan/preprocessing/CUDA/shared/experiment
indexcurrent. Nextcommit/sourcefreeze andsubmit fft-01 thenyieldunmonitored.
