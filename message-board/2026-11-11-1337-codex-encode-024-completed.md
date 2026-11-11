# ENCODE-024 — cached s16 CPU encoding passed

Timestamp: 2026-11-11T13:37:02.635860+00:00
Agent: Codex
Status: completed

User completion ping authorized inspection. Job73131 encoding-01 at runtimeae27de5 passes, Slurm COMPLETED/0:0 elapsed34m53s. [Report](../experiments/cached-s16-encoding/report.md), [analysis](../experiments/cached-s16-encoding/runs/encoding-01/analysis.json) and final accounting preserve exact evidence. One32-worker CPU encoding1252.503 s (20m52.5s), load262.332 s, initialization10.213 s, checks402.171 s, packing56.277 s, readback1.506 s. Worker elapsed2004.025 s.296 measured source hashes and sealed evidence/small cache/input provenance verify. Packed-payload metadata/size agree with the completed controller digest; analysis explicitly marks no fresh8 GiB bulk rehash. Default analyzer retains full-rehash mode; recorded-cache mode and metadata rejection self-checks documented.

All918,672,421 systematic positions,16 independent sampled curves, full canonical/packed hashes and1027 readbacks pass. Canonical SHA c538a1bc9b408dc4d283aba887f808825404a1c2b8dce693aa7f3015d1152a48; packed SHA1c68bfe60f3914b06e31995afc67d5bb3d0a6776c4750af836f8c9794ee18bb9. Packed output8 GiB retained; original19.42 GiB preprocessing untouched. Peak sampled RSS74.370 GiB<224 GiB. Native stages: Taylor200.282 s, columns/transposes451.469 s, rows552.214 s, buffer14.832 s, pointwise25.216 s, assembly7.190 s. No GPU, preparation or second encoding.

Updated report, README overview/current/historical scope, native README, scaling plan and s16 gate to record one successful experimental mathematical encoding. README checked; metadata self-checks, Python syntax, links and whitespace pass. No encoder runtime change or new job. All caches/binaries/source snapshots/profiles retained. No independent full-reference word/fresh-repeat equality/latency distribution/cryptographic application/TRL5 claim. No matched old-code s16 speedup inferred. Current implementation/evidence/doc updates staged, not yet committed/pushed. Any future>30-second job must be handed off under rule8.
