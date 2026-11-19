# Direct versus MPI-owned staging

Status: **77868 failed in direct-a; staged-a passed** (SCALING-128).
The comparison stopped on the first failed arm as designed. Later direct-b and
staged-b are **not run**; no complete matched timing ratio is reported.
[Receipt](submission.json), [raw results](runs/compare-01/results.json),
[review](runs/compare-01/analysis.json),
[accounting](runs/compare-01/scheduler-accounting.txt), [progress](slurm-77868.log),
[durable arm summary](runs/compare-01/comparison-progress.json).
Implementation `334cdfa`; **646 source / 29 original evidence / 136 cache hashes**
verified. Binary/library hashes are unchanged across the completed arm and review.
Raw outcomes, logs, payloads and profiles are preserved without editing.
[Runner](run.py), [scheduler script](run.sbatch), [reviewer](analyze.py),
[review tests](test_review.py), [shared tree runner](../root-merge-tree/run.py).

HPC: gnode13/gnode14, two H200s, four logical ranks/CPU cores, 4 GiB requested host
RAM per node. Scheduler elapsed **3 min 6 s**, controller **185.362 s**. No resource
or time guard caused the transfer failure. Retained cache: **34.574 MiB**.

| Round | Correctness outcome | Largest synthetic-case median pipeline | Normal largest logged process peak RSS per rank |
|---|---|---:|---:|
| staged-a | Passed all 22 normal passes, 516 transfer proofs; four-rank CUDA memcheck zero errors/leaks | 1.217932 s, 12 timed samples | 410.676 MiB |
| direct-a | Failed transfer fingerprints in largest case/sample 5; 52 earlier rank measurements passed | No validated full-round timing | 407.219 MiB observed before abort |
| direct-b | Not run after failure | Unknown | Unknown |
| staged-b | Not run after failure | Unknown | Unknown |

Staged-a smaller fixture medians: genuine s=8 **0.014448 s**, synthetic s=17 with
65,536 exponents **0.295641 s** (two timed passes each). Largest fixture has
262,144 exponents; owned GPU arrays peaked at **15.241 MiB per rank**, contexts
additional. Normal command **38.396 s**, memcheck **3.982 s**, profile summary
**28.060 s**; CPU references/setup, excluded warm-up/artifacts and CPU profiling
are retained and separated from per-pass medians. This repeats the earlier
[bounded staged pass](../mpi-owned-staging/report.md), not full s=17/s=18 support.

## Captured direct transfer failure

At largest case/sample 5, rank **1 → 0**, both on **gnode13**, sent 65,537 uint32
coefficients (**262,148 bytes**, two MPI chunks). Sender pre/post hashes agree;
receiver pre/post hashes also agree, but they differ across endpoints. Both
synced captures match their logged current-data hashes and exchanged proofs.
The paired files differ in **10,239 words**, all received as zero. First mismatch:
zero-based byte **40,960**, word **10,240**, sender **77,064**, receiver **0**.
See `failure_evidence` in the [review](runs/compare-01/analysis.json) for exact
paths and hashes. Payloads stay outside Git, under the sealed retained cache:
`.cache/liftrs/mpi-buffer-comparison/compare-01/direct-a-normal/rank-{0,1}/`.
The reviewer checks extents, proof pairing, snapshot hashes and first differences.

This is direct evidence of differing host payloads before receiver GPU work. The
failure is within one node, so inter-node network transfer is not required to
trigger it. The underlying MPI/runtime/buffer-lifetime cause is **not established**.
Do not claim that hardware is faulty, that a particular transport is selected,
or that staging has proved a permanent fix. Earlier direct-buffer gates remain
failed. Staging adds **512 KiB host scratch/rank (2 MiB total)** and CPU copies;
its isolated speed penalty remains **unknown**, since the direct round failed.

**Next:** keep the direct experimental path marked unsafe by these tests; retain
explicit staged mode for further bounded distributed work. Isolate the host-buffer
transfer fault with a small host-only reproduction, independently of arithmetic,
before any default decision. Continue full-size discovery/product/FFT integration
using only a validated explicit transfer mode. No new job submitted in this review;
full s=18 preprocessing remains unrun.

The offline profile command succeeded but took 28.060 s, close to its former
30 s deadline. Future shared tree runs allow **90 s for this reporting command**
inside the unchanged 720 s controller / 15 min scheduler limits; no timed pipeline
or resource allocation changes. Cold Go-tool startup/symbolization costs are not
separately measured, so the earlier timeout cause remains unconfirmed.

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

## Review verification (SCALING-128)

[Exact review commands and changed-source hashes](runs/compare-01/review-checks.json).
Six comparison reviewer/orchestration tests pass, including paired corruption
captures and rejection of a missing endpoint; 3 tree / 4 transfer-proof reviewer
tests also pass. Independent review reruns all completed staged CPU/state/proof/
memcheck checks and reconciles source/evidence/cache seals. No new MPI/GPU run or
native compilation: only Python review/reporting and documentation changed.
