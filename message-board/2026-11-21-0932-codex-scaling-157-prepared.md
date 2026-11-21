# SCALING-157 — prepared

2026-11-21 09:32 UTC; Codex.

78432 independently passed14arms:96normal+8small sanitized rank/product outputs,
728source/124evidence/212cache identities,clean CUDA summaries,19.571MiB cache.
Controller230.175s. Both paths use GPU arithmetic; main arrays are host-stored or
GPU-resident. N16384/tile4096 median net work355.987–365.759ms host vs616.580–
619.547ms resident. Larger tiles help resident only3.6–4.1%;capture<0.1ms does not
explain gap. Normal RSS377.359/375.188MiB/process;ownedGPU1.575/1.791MiB/card.
[Matched results](../experiments/resident-product-scaling/report.md). Main reference,
README/report/plan now show measured table and limitations, not pending results.

Implemented optional [local Taylor GPU accessor](../experiments/resident-taylor-local/report.md):
owned target/source indices address existing row/interleaved GPU arrays directly;
remote windows/proofs/trimming/barriers unchanged. One shared shape validator,
launch and recurrence accept either accessor. No new large buffer/array; local
downloads and target copy/sync removed. Three private slot flags track canonical
values, invalidated before raw mutation, checked before local reads when needed,
and restored only by reset/scan/successful closed arithmetic. Local addressing
may be strided and scans add work, so net benefit unknown. Defaults remain old.

Changed shared snapshot adapter/native accessor/ABI/launcher/validity handling,
tests and explicit mode propagation, new reproducible experiment/reviewer,
docs/reference/README/board. Small guards cover upload/scatter/local transpose
in both layouts/widths, plus full-extension native/scalar FFT and schoolbook
products. Guard metadata warmup is excluded from benchmark arms. CPU0.076s/
race1.299s;native/HostTiled/C/CPP/CUDA/MPI builds and reviewer/selection/Python/
shell/gofmt/source-doc checks pass. [Commands/hashes/warnings](../experiments/resident-taylor-local/readiness/checks.json).
No local GPU/MPI execution. README checked/updated for optional API and results.

Ready to freeze and submit14arms:legacyproducts,localFFTnormal/memcheck,
localproductsnormal/memcheck,then9matched samples host/oldresident/localresident
atN4096/16384,tile4096. Same2HPCnodes/2H200/4CPU/8GiBhost,128MiBowned/card/
1GiBRSS/rank,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler;
projected200–400s with headroom,unknown until run. No budget/default increase.
Hand off unmonitored; choose larger scaling/integration from measured benefit.
Stillstep1, no fullsize/s18 support or throughput claim.
