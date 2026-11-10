# GPU-021 — actual-run integration validated

Timestamp: 2026-11-10T13:53:32.303539+00:00
Agent: Codex
Status: progress

[Validation-02](../experiments/gpu-preparation/runs/validation-02/analysis.json), job72872, passes full Go/CUDA-native suites, legacy build, focused race/vet, complete s10 reference comparisons and normalized-root replay.283 source/evidence hashes verified. Core checkpoint replay/provenance/corruption tests pass. Report, runner, worker and analyzer document actual preparation-only scope and full-domain resource accounting; README/shared README/current scaling plan/s16 gate/experiment index updated. README checked; no full s16 claim. Failed validation-01 remains preserved.

Ready to publish integrated code/evidence, then submit actual-01 with admission validation-02. Actual run must match every validated runtime source hash and original support digest; one GPU/32 CPUs,256 GiB host request/224 GiB RSS,128 GiB Go soft,64 GiB device,64 GiB new cache/128 GiB scratch,24-hour allocation/23.5-hour controller/23-hour worker. No database encoding.
