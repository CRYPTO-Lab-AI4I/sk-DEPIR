# Full s17 GPU retry76706 submitted

- Timestamp: 2026-11-17T08:50:03.184384+00:00
- Agent: Codex
- Work: SCALING-098
- Status: Handoff

Submitted **76706**, fresh `gpu-retry-02`,4H200112GiB/card,32cores,
640GiB host/608GiB total guard,320GiB RSS/288GiB Go,64GiBnewretained/
16GiBscratch,46/47/48h cutoffs. Encoding/probe sourceed0c026 remains unchanged;
updated controller9855919 is retained and hashed separately. Reviewed CPU76645
reference is preserved and must match complete GPU canonical/packed digests.

[Receipt](../experiments/s17-full-encoding/submission-gpu-retry-02.json),
[budget/recipe](../experiments/s17-full-encoding/report.md#gpu-retry-with-conservative-host-headroom-scaling-098).
Results: `experiments/s17-full-encoding/runs/gpu-retry-02/gpu/`;
log: `experiments/s17-full-encoding/gpu-76706.log`;
cache/output/profiles: `.cache/liftrs/s17-full-encoding/gpu-retry-02/gpu/`.
Updated active reference overview and detailed GPU row to show retry submitted,
not a passed result, and updated plan/report. README checked/updated in9855919.
Four targeted runner tests, manifest policy checks, syntax/whitespace pass.
No monitoring of this job or inspection of76704. After completion ping, run
`python3 -B experiments/s17-full-encoding/analyze.py gpu-retry-02 gpu`, reconcile
source/controller/evidence/cache and compare to passed CPU digests. No new CPU
or preprocessing job was launched. Runtime/full GPU correctness remains pending.
