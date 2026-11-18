# Buffered persistence77352 submitted; no monitoring

- Timestamp: 2026-11-18T18:31:57.969240+00:00
- Agent: Codex
- Work: SCALING-119
- Status: Handoff

Submitted **77352**, [runner](../experiments/preprocessing-persistence/run.py),
[receipt](../experiments/preprocessing-persistence/submission.json),
[report](../experiments/preprocessing-persistence/report.md).
Implementation committedf493cb9. Actual frozen snapshot check628files/2972651bytes
matches all8 readiness implementation-input hashes;
[check](../experiments/preprocessing-persistence/readiness/snapshot.json).

Results/profiles `experiments/preprocessing-persistence/runs/io-01/`, progress
`experiments/preprocessing-persistence/slurm-77352.log`, cache
`.cache/liftrs/preprocessing-persistence/io-01/`. Within existing budget:1HPCnode,
2CPU/4GiB RAM,0GPU,2GiB RSS/3.5GiB cgroup guard when available,128MiB retained,
2GiB scratch/32MiB file/15min scheduler. Compare fixed96MiB per trial,64KiB direct
codec calls vs4MiB buffer and1/2/4 distinct-file owners, two mirrored samples.
Full small concurrency probe repeats under-race before measurement; race and
shared-storage performance are pending, not local results. CPU profiling and
code/payload retention/automatic guards enabled. Synthetic output is not valid
preprocessing; only its verified trial payloads are pruned to honor storage cap.

Updated main scaling reference/plan, README/preprocessing docs/experiment index,
report and board with submitted status. README checked; no public/default/format
change or fulls18 claim. Local build/unit/restart/reference/12-case small smoke
and3reviewer regressions pass, recorded in prepared message/readiness report.
No source changes after submission; receipt/documentation only.

User handoff: do not poll scheduler/process, follow logs or inspect results.
Await user completion ping. After review, evaluate full publication/readback and
cache/memory/time effects before adopting buffer/writer settings or seeking a
larger/multi-node I/O allocation. Fulls18 support/merge-tree/generator/transforms/
full publication/runtime estimates remain outstanding.
