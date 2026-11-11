# CPU-030 — Validation completed

Timestamp: 2026-11-11T22:24:24.907361+00:00
Agent: Codex
Status: completed

User confirmed all jobs finished. [Analysis](../experiments/cpu-buffer-reuse/runs/reuse-01/analysis.json) reconciles328 source identities and sealed evidence/cache hashes; both fulls16 canonical/final packed digests match retained CPU/GPU reference. All K systematic values,16 sampled curves per repeat,1027 packed readbacks, owning-output/partial-overlap/error/concurrency tests, native/Go suites, race/vet and ASAN/UBSAN pass. No runtime correction required.

Sampled RSS58.351486 GiB versus74.370125 GiB prior CPU,16.018639 GiB/21.539% reduction. Principal model58.266964 GiB, exact4N ownership transfer. Encoding1281.328327/1247.173287 s, setup0.085830 s, load93.362066 s, pack57.336773 s. Full job1h04m36s on gnode17/Xeon6710E32 CPU/noGPU, Slurm58.37G, exit0:0/no guards. Prior full encoding1252.502871 s is historical; first current+2.30%, second-0.43%. Matched s10 baseline0.255081/0.237360 s versus reuse0.261056/0.235616 s has mixed signs, not a consistent time penalty/benefit or universal equivalence claim. Extra caller-retained results still cost their storage; Go/hybrid paths retain separate outputs.

Updated [report](../experiments/cpu-buffer-reuse/report.md), root/shared/C++ README and scaling checkpoint. Final accounting, phases, profiles and source snapshots retained. Analyzer uses completed-controller8 GiB packed digest with metadata/size checks rather than fresh bulk rehash; other sealed artifacts verified. Self-checks/doc links/working whitespace pass. No redundant long suite after results/doc-only review. No new job/budget/publication requested. CPU ownership implementation was already published as pending in4b08075; this closes its validation gate.
