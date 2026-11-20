# SCALING-153 — prepared

2026-11-20 17:20 UTC; Codex.

78392 passed independent review:208 normal/208 sanitized rank/pass outputs,
544 proof endpoints per mode, source/third preservation and reuse, clean CUDA
error/leak summaries.710source/33evidence/835cache identities verified.
Controller123.480s,normal4.924s/sanitizer4.024s;RSS376.137/413.750MiB/process,
1.583MiBowned/card,14.849MiBretained. HPCgnode14/15,2nodes/2H200/4CPU/8GiBhost.
[Raw/review evidence](../experiments/distributed-gpu-storage/report.md) retained.
Main scaling table, plan, design and README updated with completed status/units.

Implemented [resident vector composition](../experiments/distributed-resident-vectors/report.md):
existing vector launcher shared by both adapters; common host/resident separable
stage order. GPU view field scan rejects invalid raw slots before writes, adding
one read and4bytes status transfer per stage (cost unmeasured). Plan retained
under existing cap. Resident data/scratch avoid whole-host vector staging; third
slot preserved. No new FFT/field recurrence, production default or field admission.

Changed CUDA executor/rank/vector APIs and comments, shared stage adapters/order,
new test/experiment/reviewer, validated composition runner's optional baseline,
README/API docs, plan/reference and board. CPU regressions0.076s, race1.357s;
native/C/CPP/CUDA/MPI builds and Python reviewers/launcher/shell/gofmt/whitespace
pass. [Commands/hashes/warnings](../experiments/distributed-resident-vectors/readiness/checks.json).
No local GPU/MPI execution. Initial local compile/fixture mistakes corrected
before passing final checks. README checked/updated for API and measured status.

Ready to freeze source and submit old host-tile regression plus resident normal/
memcheck on the existing2node/2H200/4CPU/8GiBhost budget,128MiBowned/card,
1GiBRSS/rank,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler.
24rows per mode, independent native/scalar forward and exact inverse output,
range-error nonmutation, third-slot preservation and allocation reuse. Hand off
unmonitored. Resident Taylor/full FFT/product integration remains pending.
