# Full s17 GPU retry ready with conservative host/cache allowance

- Timestamp: 2026-11-17T08:48:44.653315+00:00
- Agent: Codex
- Work: SCALING-098
- Status: Prepared

Updated [runner](../experiments/s17-full-encoding/run.py) GPU policy and
[launcher](../experiments/s17-full-encoding/gpu.sbatch) to640GiB host request/
608GiB total guard. Unchanged:320GiB RSS/288GiB Go,4H200112GiB/card,32cores,
64GiB retained/16GiBscratch,46/47/48h deadlines. Full process cap plus entire
142.441GiBcode/34GiBoutput cache leaves111.559GiB guard margin. This supersedes
512/480GiB; one node advertises~907GiB. Shared filesystem12900.7GiB free;
compute-node scratch is checked by the job, not inferred from login `/tmp`.

Keep encoder/probe sourceed0c026; do not include the unvalidated spectrum changes
from76704. Record actual controller in a sealed sidecar. New optional
`--reference-run first-01` checks complete canonical/packed GPU hashes against
CPU76645 before success and during analysis. Guard errors persist exact resource
observations before stopping. Four no-allocation tests pass, including>100GiB
worst-case cache margin and rejecting each mismatched digest. Manifest/extent,
Python syntax, Slurm shell syntax and whitespace checks pass.

Updated [report](../experiments/s17-full-encoding/report.md#gpu-retry-with-conservative-host-headroom-scaling-098),
[reference](../docs/encoding-status.md), [plan](../docs/encoding-scaling.md) and
README (checked/updated). No new arithmetic/backend/memory mode. No expensive
rereads or CPU/preprocessing reruns. Submit fresh GPU-only `gpu-retry-02`, then
hand off without monitoring. No inspection of unrelated76704.
