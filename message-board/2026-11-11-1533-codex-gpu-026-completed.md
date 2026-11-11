# GPU-026 — Full s16 CUDA encoding passed

Timestamp: 2026-11-11T15:33:05.241142+00:00
Agent: Codex
Status: completed

User completion ping authorized review of job73257. [Analysis](../experiments/cuda-s16-encoding/runs/encoding-01/analysis.json) verifies317 source identities, completed evidence/cache digests and retained CPU canonical/packed digest agreement. All K918,672,421 systematic positions,16 sampled independent curves per repeat and1027 packed readbacks pass; two same-message GPU encodings agree. Completed allocation30m02s, exit0:0, no guard tripped. [Report](../experiments/cuda-s16-encoding/report.md) links exact runner, worker, results, phases, profiles and accounting.

s16 encoding85.954/60.216 s, session setup7.193 s, online load164.070 s, pack/write56.689 s. Native total85.941/55.368 s; first output download31.837 s versus1.152 s later, cause unproven. Second Go profile records4.85 s output clearing. Matrix rows/columns dominate warmed device work. Independent audit/readback takes1309.019 s, explaining most of the full validation allocation. Prior32-thread CPU encoding1252.503 s yields historical ratios14.57x/20.80x, not a matched benchmark. Device-owned67.423 GiB excluding context/driver; host sampled39.024 GiB (Slurm39.05G). Tight8 GiB packed output and input/cache/profiles retained.

Updated root/shared/CUDA READMEs, s16 representation status and scaling plan; no production code changes required. Python admission/reconciliation self-checks, documentation relative links, working whitespace and saved analyzer pass. Completed analysis uses the controller's sealed8 GiB payload digest rather than a fresh bulk rehash; physical size and metadata checked. No redundant encoding/Go suite rerun after documentation-only review. Full CPU-only/CUDA-only/combined/race/vet/memcheck evidence remains job73243. CPU stays a peer main path. Application/cryptographic integration, s17, smaller GPU feasibility and TRL5 remain open. No further job or commit/push requested or performed.
