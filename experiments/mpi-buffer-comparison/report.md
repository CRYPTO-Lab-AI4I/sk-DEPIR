# Direct versus MPI-owned staging

Status: prepared for submission (SCALING-127); no comparison result yet.
[Runner](run.py), [scheduler script](run.sbatch), [reviewer](analyze.py),
[review tests](test_review.py), [shared preparation/execution](../root-merge-tree/run.py).
Exact source revision will be recorded in submission.json and frozen inputs in
results.json. [Previous bounded staged result](../mpi-owned-staging/report.md):
normal validation and CUDA memcheck passed; original direct-buffer failures remain
unexplained. This experiment does not assume that staging fixes their cause.

Build the native library and tagged Go test executable **once**. Reuse these
identical files in four sequential arms on one HPC allocation:

| Arm | Payload mode | Extra owned host scratch per rank |
|---|---|---:|
| staged-a | MPI-owned chunk plus CPU copy per send/receive | 512 KiB normal; 8 KiB memcheck |
| direct-a | Direct Go payload | 0 bytes |
| direct-b | Direct Go payload | 0 bytes |
| staged-b | MPI-owned chunk plus CPU copy per send/receive | 512 KiB normal; 8 KiB memcheck |

Each arm starts fresh MPI processes, reruns CPU references/setup, and uses the
same fixtures, compiler flags, GPU kernels, routing, checks, chunk sizes and rank
placement. Only the explicit `-root-mpi-buffer` argument changes. Both modes use
the same C-local count/proof records and symmetric payload proofs. Check binary/
library hashes after each completed arm and in later review. All final arrays
must agree across arms, and physical node-to-rank placement must be identical.
Order reversal helps expose drift but is not replication across allocations.

Per arm: genuine s=8 (7,381 exponents) and synthetic s=17 (65,536 and 262,144
exponents). Smaller cases have two timed passes; largest has 12. Every fixture
has an excluded warm-up and excluded artifact pass. Thus each mode has 24 timed
largest-case samples across two arms. CPU equality, exact support coverage,
unique ownership, canonical bounds, transfer pairing and full retained-state
checks remain mandatory. Each arm also runs small-case four-rank CUDA memcheck
and retains all four CPU profiles. Shared offline profile summaries run after
timing; native symbol lookup is skipped but raw profiles remain available.

Per-pass clock includes routing, index/coefficient fingerprints and proof
traffic, GPU roots/tree, and CPU final equality. Setup, CPU reference construction,
warm-up, artifact snapshots and profiler summaries are excluded from those
medians, but command elapsed times, per-case setup, profiles and total workflow
latency are retained. Phase maxima are not additive. The reviewer reports two
staged/direct ratios of matched arm medians (below one means staged was faster).
These profiled diagnostic ratios do not isolate raw copy cost or establish
full-size throughput. MPI/OpenFabrics warnings are retained; RDMA use is not
assumed. Small fixtures and two logical ranks sharing each card limit scaling
conclusions. A clean direct run does not erase earlier failures.

Stop immediately on a failed command/check/guard. Preserve the raw failed status,
completed-arm summaries and failure payloads; do not produce a complete timing
comparison from failed arms. `comparison-progress.json` is saved before each arm
and after every completed arm, independently of the final results assignment.
Worker/native timestamped phase, progress, chunk and resource logs stay enabled.
This is checkpointed experimental evidence, not full-size restart support.

Resources remain inside the previously admitted envelope: **2 HPC nodes, 2 H200s
(one/node), 4 logical ranks/CPU cores (two ranks/node), 4 GiB requested host RAM
per node**, 1 GiB process RSS guard/rank, 64 MiB owned GPU/rank (at most 128 MiB/
card; contexts additional). Staging is at most **2 MiB extra host memory total**;
no extra GPU allocation, transform or production output bytes. The copies may
increase runtime; the penalty remains unmeasured. Retained-data cap 128 MiB,
scratch cap 2 GiB, file cap 32 MiB; controller cutoff 720 s, scheduler 15 min.
Projection from 77799's retained files: 12.810 MiB shared build plus four times
14.757 MiB arrays, approximately **71.838 MiB** retained cache; new failure captures
also remain subject to the 128 MiB guard. No earlier caches/profiles are deleted.
Expected elapsed **4–7 min after scheduling**, estimated from 77799's build and
per-arm command times, not a measurement or guarantee. No resource increase.

Reproduce on HPC with Go 1.27.3/GCC 14.2/CUDA 12.9.1/OpenMPI 4.1.8/Slurm 25.05:

```sh
python3 -B experiments/mpi-buffer-comparison/test_review.py
python3 -B experiments/mpi-buffer-comparison/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/mpi-buffer-comparison/slurm-%j.log \
  experiments/mpi-buffer-comparison/run.sbatch compare-01
# After the user's completion ping only:
python3 -B experiments/mpi-buffer-comparison/analyze.py compare-01
```

Results/profiles and the intermediate summary: `runs/compare-01/`; progress:
`slurm-JOBID.log` and arm-prefixed logs. Cache/binaries/arrays/failure captures:
`.cache/liftrs/mpi-buffer-comparison/compare-01/`, outside Git with file seals in
results.json. Source and evidence hashes also recorded. Run from the same revision
with a fresh name to regenerate. Guards/deadlines persist without agent monitoring.
No production default, CPU encoder behavior or full s=18 support changes.

Local readiness: five comparison tests cover actual mode/scratch, timing
exclusions, mismatched outputs/placement, fresh paths/small-fixture reset and
fail-stop durable summaries. Existing 3 tree / 4 proof / 4 distributed / 5 staged /
3 shared-runner tests pass; Python and scheduler syntax checks pass. Native kernels
and tagged Go harness are unchanged; their compilation/runtime checks run once
inside the submitted job. No local MPI/GPU execution. [Commands and hashes](readiness/checks.json).
