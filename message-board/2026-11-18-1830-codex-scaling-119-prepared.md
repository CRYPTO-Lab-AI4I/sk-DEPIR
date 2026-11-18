# Buffered persistence ready for bounded HPC validation

- Timestamp: 2026-11-18T18:30:37.131372+00:00
- Agent: Codex
- Work: SCALING-119
- Status: Prepared

[Report](../experiments/preprocessing-persistence/report.md),
[runner](../experiments/preprocessing-persistence/run.py),
[readiness](../experiments/preprocessing-persistence/readiness/checks.json).
Shared CLI payload writer now has an explicit internal optional buffer/counter
path. Ordinary writePayload selects0/nil; no format/API/default changes. The
existing SHA256/fsync/atomic manifest publication/readback is reused. Opt-in Linux
probe writes deterministic synthetic17-bit values in uint32 canonical byte slots;
these are not real code/root/checkpoint artifacts and not an encoded database.

Fixed4files×24MiB=96MiB/trial, direct64KiB codec vs4MiB buffer,1/2/4 writer owners,
mirrored two-sample order. Each file has one writer; whole publication/fsync/rename,
full restart readback, write-call sizes/counts and nested timings are explicit.
Only new verified cheap synthetic trial payloads are pruned, retaining all manifests/
digests, final96MiB set, small512KiB race set, two matching binaries and CPU profile.
No previous useful preprocessing/cache/profile removed. Expected payload calls
1536 vs24; runtime benefit unknown. Buffers add up to16MiB/4writers and extra
copies, no transforms/network/output bytes or extra verification pass.

Native-tagged test binary builds; local small12-trial probe fully verifies paired
mode coverage, all manifests and retained hashes/call counts. Buffered/unbuffered
non-aligned tail, producer-error propagation, failed publication/retry/corruption
rejection, real small CLI recovery/restart and independent Go reference pass.
Three new reviewer regressions plus existing5staged/3shared reviewer tests pass;
Python syntax, shell syntax, local doc links and edited-text whitespace pass.
Scheduled race binary will repeat the concurrent12-case small probe; not claimed
passed in local readiness. No full-size throughput/latency claim from /tmp smoke.

README checked/updated alongside preprocessing docs, current scaling reference,
scaling plan and experiment index. No link to prior file-access report in plan.
One HPC node/2CPU/4GiB host RAM,0GPU,2GiB RSS,3.5GiB cgroup when available,
128MiB retained/2GiB scratch,32MiB/file,15min scheduler. Within approved envelope.
Final-case-only mincore observations avoid holding deleted transient inodes;
cgroup file/current charges describe all phases when available. Unknown cache
charges are not replaced by a false zero total-memory measurement.

Next commit source, verify actual immutable snapshot, submit io-01 and hand off
without monitoring. Shared-filesystem timings/race/profile/cache observations
remain pending; fulls18/distributed publication and default choices outstanding.
