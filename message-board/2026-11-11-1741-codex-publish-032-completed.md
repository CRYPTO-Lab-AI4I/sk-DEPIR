# PUBLISH-032 — CUDA checkpoint verified for publication

Timestamp: 2026-11-11T17:41:58.564635+00:00
Agent: Codex
Status: completed

User-authorized publication includes primary optional CUDA API/backend, maintained independent CPU builds, shared matrix descriptions, historical GPU experiment shims, completed CPU job73131 and GPU jobs73243/73257, exact source snapshots, reasonably sized profiles/reports, README/contracts/scaling status and board records. Large caches and packed outputs remain local outside Git. Total staged payload before publication messages is about1.48 MB; no cache directory or large generated binary is staged.

Verification: saved CPU encoding analyzer with recorded-cache option, complete modular GPU validation analyzer and fulls16 CUDA analyzer with recorded-cache option all pass; metadata/digest/repeat/cap rejection self-checks pass. Staged/working whitespace checks pass. Source identities and sealed evidence match; fulls16 packed digests use completed-controller hashing rather than a fresh bulk reread. Existing runtime CPU-only/CUDA-only/combined suites, race/vet and CUDA memcheck evidence are retained; no redundant long job. Root/shared/CUDA README checked and performance/setup/limits agree with [full report](../experiments/cuda-s16-encoding/report.md). Remote main fetched over SSH443 and matches local ae27de5 before commit; no merge needed. Commit/push are the remaining publication commands for this verified tree.

Next recommended work: improve shared packed-output writing using blocks rather than individual-byte calls. The saved pack profile spends31.39 s cumulative in bufio.WriteByte of56.69 s wall; hashing, packing and I/O must stay separately accounted for. This can benefit both CPU/GPU consumers. Then evaluate reusable output buffers/first-download costs and GPU matrix rows/columns (about45.4 s combined warmed). Preserve canonical/packed hashes, generic CPU builds, code parameters independent of backend and profiling. No new experiment or implementation launched by this publication request. s17, smaller GPU feasibility and application/cryptographic/TRL5 acceptance remain open.
