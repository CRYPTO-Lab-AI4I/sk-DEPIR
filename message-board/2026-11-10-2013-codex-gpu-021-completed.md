# GPU-021 — actual s16 preparation completed

Timestamp: 2026-11-10T20:13:44.438963+00:00
Agent: Codex
Status: completed

User completion ping authorized final inspection. Job72875 actual-01 passed, Slurm COMPLETED/0:0 with02:24:15 allocation elapsed. Worker preparation8415.469 s (2h20m15s), saved constants46.021 s; support load25.702 s reused original discovery. Full polynomial identity verified atK918,672,421/N2^32. Root checkpoint and final G/online files retained outside Git. Sampled process RSS72.575 GiB<224 GiB; new cache35.430 GiB<64 GiB. Explicit-device peak was not sampled;48 GiB is the array payload model,64 GiB cap enforced,zero ownership after close.

Updated [report](../experiments/gpu-preparation/report.md), README performance/completion scope, current scaling plan and s16 gate. README checked. Added [lightweight reconciliation](../experiments/gpu-preparation/reconcile.sh) and [result](../experiments/gpu-preparation/runs/actual-01/reconciliation.json):284 exact source hashes,42 sealed evidence files, small caches, metadata/digest agreement and actual sizes pass. Large payload digests were computed by the completed controller; no fresh bulk rehash is claimed. Final accounting retained separately. Links, shell syntax and whitespace pass. No runtime code changes or new hardware job. All existing profiles/caches remain intact.

This closes experimental s16 preparation on the measured machine, not cached database encoding/application/cryptographic execution or TRL5. The rough3–5-hour estimate was conservative. Next is a separately admitted cached s16 encoding with systematic/curve/packed-readback checks; none launched. Respect rule8: hand off all jobs expected>30 seconds and wait for user completion ping. Final report/evidence staged for review; published runtime remainsd9a80f0.
