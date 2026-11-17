# Bounded spectrum validation76704 verified

- Timestamp: 2026-11-17T08:54:49.003221+00:00
- Agent: Codex
- Work: SCALING-099
- Status: Completed

76704 COMPLETED0:0 on HPCgnode04,4H200 NVL,4CPU,8GiBhost;3min59s scheduler,
236.909s controller. Producer46d16d2. Reconciliation passes561source/43evidence/
74cache identities without overrides. Twelve scalar-Horner s8/s17 cases pass on
1/2/4GPU peer/local, including all points and writer/range/reuse/closed errors.
Stream one-GPU CUDA memcheck0errors/0leaks; inherited four-GPU workflow memcheck,
small canonical/cache/root-restart/CUDA-free reference checks, race/ASan/vet pass.

32MiB spectrum with1MiB tile: resident242.058/221.534ms; stream217.848/215.225ms,
means231.796/216.536ms (6.58% lower stream mean). Balanced two samples include
profiling/logging and transform/download/hash-write, not disk. All complete hashes
match. Different input preparation and retained resident comparison array prevent
claiming full preprocessing speedup or measured RSS saving. Sampled command RSS
max0.565GiB includes builds/validation; device caps are not measured peaks.

Updated [report/raw results/review](../experiments/preprocessing-memory/report.md#completed-bounded-validation-scaling-099),
[active scaling reference](../docs/encoding-status.md), [plan](../docs/encoding-scaling.md),
README (checked/updated), CUDA README and experiment index. Preserved sealed logs,
profiles and31,275,710byte reusable binary/cache outside Git. Review records
scheduler accounting plus hashes of scheduler log/analysis/accounting evidence.
All relative documentation file targets and numerical summary checks pass;
source-derived model self-check reproduces exact prior model; whitespace checks
pass for edited source/docs. Preserve raw job logs byte-for-byte, including any
original whitespace. No new experiment or inspection of76706.

Next is explicit canonical stream publication/scoped host ownership with atomic
manifests, independent identities and full byte/digest/readback equivalence.
Host-backed tiled transforms and split products remain following measurements;
no fulls18 admission, memory claim, time tradeoff selection or default change.
