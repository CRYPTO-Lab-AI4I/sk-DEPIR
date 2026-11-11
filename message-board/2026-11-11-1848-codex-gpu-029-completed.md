# GPU-029 — Two-work-array reuse validated

Timestamp: 2026-11-11T18:48:48.525429+00:00
Agent: Codex
Status: completed

User completion ping authorized result review of73576. [Analysis](../experiments/cuda-buffer-reuse/runs/reuse-01/analysis.json) reconciles323 exact source identities, sealed evidence/cache digests and both full CPU canonical/final packed hashes. CPU-only, CUDA-only, combined, CGO-disabled tests and vet pass;80 matrix regressions pass; small oracle/error/lifecycle/prefix/primitive/old-ABI rejection pass; memcheck reports0 errors and0 leaked bytes. All K systematic values,16 sampled independent curves per repeat and1027 packed readbacks pass. No code correction needed from this validation.

Measured owned GPU memory55,219,301,804 bytes/51.426982 GiB versus72,394,976,680 bytes/67.423076 GiB prior default:17,175,674,876 bytes/15.996094 GiB (23.724954%) saved, exactly one16 GiB array minus4,194,308 parity/error bytes. Both calls fit56 GiB cap. Host sampled39.030907 GiB, no host-plan-release change. gnode10/H200 NVL/driver580.126.20/Xeon6710E; old full baseline was gnode17.

Encoding57.417965/56.974810 s; setup3.701042 s, load309.385648 s, pack/write53.931674 s, verification157.852406/157.137699 s. Full worker805.764654 s, Slurm allocation19m05s, exit0:0, no guard tripped. Earlier35–50-minute forecast was conservative; verification was substantially faster here. Full timing comparisons to73257 are historical and cannot isolate reuse effects. Matched s10 default0.103897/0.103143 s versus reuse0.103885/0.103060 s differs by <0.1%; no observed encoding penalty in this small fixed-order comparison, not a general equivalence claim.

Updated [report](../experiments/cuda-buffer-reuse/report.md), root/shared/CUDA READMEs and scaling checkpoint. Persisted accounting, result analysis, profiles, raw evidence and exact source snapshots; previous caches/output retained. Checks: saved analyzer, admission/device-saving self-checks, documentation links and working whitespace pass. No redundant runtime suite rerun after documentation-only review. Recorded-cache reconciliation relies on completed-controller hashing of8 GiB packed payload, with physical size/metadata checks rather than fresh bulk reread.

Selected buffer-reuse validation step complete; keep opt-in mode and public/default GPU/CPU behavior as is. s17 still models~398 GiB principal GPU arrays with this schedule; new in-place/streamed-code versus host/disk-backed route remains an explicit memory–time choice before adoption. No new job/budget/default switch or commit/push requested/performed. Application/cryptographic/TRL5 limits remain open.
