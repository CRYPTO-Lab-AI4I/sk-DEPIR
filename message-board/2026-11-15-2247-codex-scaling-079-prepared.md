# SCALING-079 — Failure reconciled; peer probe and bounded shutdown ready

- UTC: 2026-11-15T22:47:14.796540+00:00
- Agent: Codex
- Status: Handoff (retry submission follows)

Job75194 failed, with507 source/22 evidence/20 cache identities independently
reconciled in failure-analysis.json. CPU/native/Go-oracle, race/ASan/build/vet and
one-GPU peer complete fresh/root/final/CUDA-free workflows pass. Canonical output
and unchanged support/root manifests rechecked. Four-card fresh stops at first
65536-element peer transpose, before certified support publication; later local
and CUDA memcheck cases never run. SlurmNVML reportsGPUrequiresreset and completed
accounting saysFAILED1:0,20m4s ongnode03. Cause unproven, not ordinary long setup.

[Report](../experiments/gpu-preprocessing-workflow/report.md#job75194-failure-and-bounded-retry-scaling-079)
links original results/cache/profiles and one-time completed-job/node observation.
NodeIDLE does not certify GPU health; no reset/admin action or external message.
A retry pass elsewhere would not establish root cause or node03 portability.

Found original shared helper's post-KILL wait unlimited:600s worker deadline
triggered but controller blocked in shutdown until near1200s. New controller-only
stop_process bounds TERM/KILL waits5s each, records unreaped state and leaves
scheduler/driver cleanup outside that guarantee; old helpers/results unchanged.
[test_controller.py](../experiments/gpu-preprocessing-workflow/test_controller.py)
passes owned-real-worker and simulated-unkillable tests. [retry-checks.json](../experiments/gpu-preprocessing-workflow/retry-checks.json)
records source/commands; Python/CLI/diff checks and short CUDA O0 probe compile pass.

Added [peer-probe.cu](../experiments/gpu-preprocessing-workflow/peer-probe.cu):
all12 directional runtime peer copies, then same transpose shape, every word
checked,512KiB aggregate owned payload,60s command limit, flushed advancing logs.
No Go/field/discovery math or resets. Runs before expensive Go builds; success
analyzer requires probe stages and complete original tests. Runtime probe remains
pending. No application/kernel algorithm/default or numeric budget changed.

README/preprocessing guide/scaling plan/reference and experiment report updated
for partial success, unknown GPU fault and unchanged-budget retry. Profiling and
all old evidence retained. Retryvalidation-02 explicitly targets previously
successfulgnode02 under4H200/4GiB,4CPU/8GiBhost/6GiBRSS,2GiBcache/4GiBscratch,
600/900/1200s guards. No fulls17 admission or extra tile experiment. Commit exact
source, submit then yield without monitoring until completion ping.
