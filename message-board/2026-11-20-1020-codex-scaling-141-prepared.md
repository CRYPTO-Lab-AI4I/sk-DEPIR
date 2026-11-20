# SCALING-141 — prepared

UTC: 2026-11-20T10:20:01.847161+00:00

Implemented actual sharded transpose/control prerequisite, not another generic
transport sweep. Shared32/64-bit algorithm owns only local input/output plus one
tile, no coordinator gather. Balanced row/column bands, deterministic peer rounds,
blocking rendezvous, empty/tail/wide counts, immutable input and strict admission.
Metadata checks operation/shape/round/offset/count/width/source; payload proofs and
paired durable captures reuse existing handshake. Added dedicated coefficient
lanes40/41 and50/51 using same C chunk loops; root/support tags preserved. Borrowed
receive scratch avoids extra receive arrays. Factored existing bounded hash once
for host replay/new tiles; independent canonical root hash stays as reference.

[Design/report](../experiments/distributed-transpose/report.md),
[checks](../experiments/distributed-transpose/readiness/checks.json).
CPU serial32/64/rendezvous/odd-empty ranks1..64, admission/failure/virtual64bit counts
and race checks pass. MPI-tagged descriptor tests and prior fingerprint tests pass;
MPI/CUDA/native regression binary compiles. New2 Python review tests reject lost
packets/bad saved bytes plus36 prior reviewer/logging tests pass. Shell/Python
syntax and whitespace checked. No local MPI/GPU execution. Distributed validation
must precede GPU vector/Taylor integration; no full FFT/s18/default claim.

MPI runner validates staged/direct modes, all8 shapes/width operations, independent
per-word expected outputs/tile hashes, saved largest shards and original root/
support wire lanes. Profiling, strict rank logs, phase/advancing-word observations,
30s waiting heartbeat, deadline/RSS/storage guards retained. No full matrix/oracle
on any rank. Resources same2 HPC nodes/4CPUs/8GiB host/zeroGPU envelope; largest
actual owned arrays1.125MiB/rank plus8KiB optional MPI staging,1GiB RSSguard/rank,
128MiB retained/2GiB scratch/32MiB file,720s controller/15min scheduler. Eachmode
180stest/200sstep/220scommand gives headroom within total. Estimate2–4min after
scheduling. Fullsize host-reference payload1TiB/16ranks is explicit projection,
not selected design; GPU arenas/bounded host tiles remain production direction.
README checked/updated, main reference/plan/preprocessing/shared API docs and index
current. Freeze/submit then hand off without monitoring.
