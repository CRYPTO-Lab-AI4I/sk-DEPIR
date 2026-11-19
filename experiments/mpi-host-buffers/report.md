# Host-only MPI buffer replay

Status: **host-only job 78000 submitted** (SCALING-129); results unreviewed.
[Receipt](submission.json), progress `slurm-78000.log`, results/profiles
`runs/replay-01/`; implementation `c48ee74`.
[Source freeze](readiness/snapshot.json): 652 files / 3,080,148 bytes; all eight
checked input hashes match. No MPI execution is implied by local readiness. [Runner](run.py),
[scheduler script](run.sbatch), [probe](probe_test.go), [reviewer](analyze.py),
[review tests](test_review.py). Implementation revision is in submission.json;
source/evidence/cache seals in results.json. No CUDA library/toolkit or GPU is used.

Replay the public polynomial captured at the sender in
[77868](../mpi-buffer-comparison/report.md#captured-direct-transfer-failure):
65,537 uint32 words / 262,148 bytes, SHA256
`4b73c559c2d99fa3e27ed1c60b584a4184265204d9bb55bcb3b9dcd1e7218d1d`.
Original file is retained at
`.cache/liftrs/mpi-buffer-comparison/compare-01/direct-a-normal/rank-1/case-2-sample-5-root-1-failed-send-0.bin`.
Runner verifies its size/hash and keeps a new sealed replay copy. No client data.

Use the existing [World transport](../distributed-root/mpi/world.go), without
field arithmetic or CUDA. The symmetric proof/capture/fsync/acknowledgement helper
is shared in [mpi/proof.go](../distributed-root/mpi/proof.go); the GPU tree delegates
to it, preserving its previous checks and handshake. CPU-only building uses
`go test -c -tags liftrs_mpi ./experiments/mpi-host-buffers` with CC=mpicc/CXX=mpicxx.
Ordinary client builds retain no MPI/CUDA dependency. The common bounded runner
now accepts an explicit host-only workload, skipping GPU identity/setup commands;
GPU workloads retain the existing default.

Four fresh-process arms, in order: staged-normal, staged-reclaim, direct-normal,
direct-reclaim. Stop on any failure; later arms remain unrun. Per arm, 64 cycles
of four handoffs: root rank1→0 (same node), root rank2→0 (cross-node), uint64
packet rank0→1 and rank0→2. Root packets replay the exact captured bytes using
65,536-word chunks, including the one-word tail. Index packets contain 16,384
synthetic uint64 values with high bits set, derived from the root prefix; they
are transport test patterns, not a claimed valid code support set.

Before each handoff, every rank allocates/touches/releases 16 MiB scratch. Both
normal and reclaim arms have this artificial churn. Reclaim arms additionally
call `debug.FreeOSMemory()` before allocating live payloads. This is an explicit
stress control, not normal preprocessing or a proposed memory optimization.
Fresh sender slices and receiver allocations borrow the same blocking MPI API.
One pass/non-reproduction cannot establish the fault's cause or clear direct mode.
Failure under forced reclamation alone must not be reported as reproduction of
ordinary preprocessing behavior. No causal claim about MPI, Go GC or hardware.

Both endpoints prove pre/post stability and peer equality; successful proofs must
also equal independent expected payload hashes. Review verifies exact route,
cycle/rank coverage, staging allocation, host placement and 1 GiB RSS limits.
Mismatch payloads are saved/fsynced by both peers before acknowledgement/abort.
Per-handoff timestamped proofs and progress are retained; completed counts advance
only after a communicator barrier. A 30 s heartbeat identifies the waiting
case/cycle and count without claiming new work. Native chunk events stay enabled.
All rank CPU profiles are retained; first staged profile gets an offline summary.
Diagnostic hashing, churn, GC, barriers and logging preclude throughput conclusions.

HPC allocation: **2 nodes / 4 logical ranks / 4 CPU cores / zero GPUs**, 4 GiB
requested RAM per node, 1 GiB RSS guard/rank, GOMEMLIMIT 512 MiB/rank. Scratch:
16 MiB churn/rank plus sub-MiB payload/hash buffers; staged mode owns another
512 KiB/rank (2 MiB total). No device memory. Existing caps: 128 MiB retained,
2 GiB build scratch, 32 MiB/file, 720 s controller and 15 min scheduler. Test
90 s / Slurm step100 s / command120 s per arm; profile summary90 s. Same-node
pair placement uses two ranks/node; physical nodes may differ from 77868.
The shared cache is used as remote working directory to avoid node-local cwd
fallback. No process/network transport is forced. All actual MPI warnings retained.

Estimated duration **2–5 min after scheduling**, not measured; build, forced-GC
cost and possible early mismatch vary. No budget increase; staging's isolated
runtime penalty remains unknown. No GPU/preprocessing default change or full s18
capability follows. Existing preprocessing and failure captures remain intact.

```sh
python3 -B experiments/mpi-host-buffers/test_review.py
python3 -B experiments/mpi-host-buffers/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/mpi-host-buffers/slurm-%j.log \
  experiments/mpi-host-buffers/run.sbatch replay-01
# After user completion ping only:
python3 -B experiments/mpi-host-buffers/analyze.py replay-01
```

Results/profiles/fixture provenance and arm checkpoints: `runs/replay-01/`;
progress `slurm-JOBID.log`; sealed replay/build/failure files outside Git at
`.cache/liftrs/mpi-host-buffers/replay-01/`. Reproduce from the recorded revision
and original sealed fixture using a fresh run name. Toolchain: Go1.27.3/GCC14.2/
OpenMPI4.1.8/Slurm25.05. No agent monitoring after submission.

[Local checks](readiness/checks.json): host-only and tagged GPU-tree binaries
compile, host hash/high-bit/proof-symmetry test passes, original tree fingerprint
tests pass. Two Python host-review tests reject corrupt proofs, missing progress
and wrong buffer mode; existing comparison/tree/proof/common review tests pass.
No local MPI/GPU execution; actual replay and shared-handshake integration await
scheduler results. Native arithmetic/transport kernels unchanged.
