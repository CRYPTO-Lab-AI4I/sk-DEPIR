# CPU-023 — within-block Taylor selected after matched comparison

Timestamp: 2026-11-11T11:30:03.423749+00:00
Agent: Codex
Status: completed

User completion ping authorized review. Job73099/taylor-01 passes; [analysis](../experiments/cpu-taylor/runs/taylor-01/analysis.json) verifies292 exact sources, sealed evidence/cache hashes, eight complete canonical outputs and same-clock stage accounting. Baseline native source independently checked against exact e29258d blob. Full Go/native suites, focused Go race/vet, s8 independent oracles, ASan/UBSan and TSAN pass. Both sanitizer drivers pass72 independent forward/inverse/reference cases; TSAN genuinely passes rather than taking the unavailable fallback.

[Report](../experiments/cpu-taylor/report.md) records complete encoding142.178→118.538 s at16 workers (1.20x),87.928→61.042 s at32 (1.44x/30.6% less time). Taylor35.675→8.768 s at32 (4.07x); rows/columns remain close. Whole-job sampled RSS4.593 GiB<12 GiB. Retain the hybrid gather/recurrence implementation; no additional runtime-code edit after the measured source. Existing recurrence remains one-worker/small/coarse-parallel reference. Threshold/chunk choices are measured initial settings, not universal optima.

Updated report, native README, root README and current scaling plan to select the change and retain the s16 gate. README checked; evidence self-check, doc links and whitespace pass. All source snapshots, preprocessing caches, old/new binaries and profiles retained. No extra job or monitor. s16 encoding/performance, broader latency distribution and native instruction attribution remain unverified. Next is separately admitted cached s16 encoding with correctness checks; do not auto-launch. Current implementation/evidence/docs staged, not yet committed/pushed.
