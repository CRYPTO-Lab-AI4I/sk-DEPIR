# Exact s=14 basis retained; preparation running

Timestamp: 2026-10-29 09:09 UTC. Agent: Codex. Work: SCALE-005. Status: update.

[Discovery phases](../experiments/parallel-s14/initial/phases.json) establish exact
K=50,976,325 at N=268,435,456 (m=2,s=14,t=5), in 254.898 s and two samples.
This corresponds to K/N=0.18990160897374153. The public basis is retained at
`.cache/liftrs/parallel-s14/s14-m2-t5/support.bin` (203,905,300 bytes), with SHA-256
`da599ca203bac67a501778d9aefb53ce1f83c0a6ff606342c8507fba54203dd1` and
`basis-metadata.json`. It can be resumed without repeating exact enumeration.

Preparation remains active with profiling, no guard failure, and observed peak
RSS near 5.9 GiB. This is not yet encoding/first-use feasibility evidence. The
[report](../experiments/parallel-s14/report.md) links the live manifest and progress;
source/binary provenance is retained. The README points to this in-progress run.
