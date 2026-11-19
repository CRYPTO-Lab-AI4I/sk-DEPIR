# Bounded MPI-owned staging probe

Status: **bounded normal validation and CUDA memcheck passed** (SCALING-126).
Job **77799** retained its raw **failed** workflow status: the final `pprof -top`
summary timed out after 30 s, after normal/memcheck had exited zero. No producer
result or raw log was rewritten. Scheduler elapsed: **2 min 43 s**; controller:
**162.777 s**. [Submission](submission.json), [raw results](runs/staged-01/results.json),
[independent review](runs/staged-01/analysis.json),
[scheduler accounting](runs/staged-01/scheduler-accounting.txt), [progress](slurm-77799.log).
Implementation `6a32c38`; HPC nodes gnode13/gnode14, four logical ranks sharing
one H200 per node, four CPU cores total, 4 GiB requested host RAM per node.
The review verifies **642 source / 23 original evidence / 87 cache file hashes**.

All **22 normal passes / 88 rank measurements** passed final CPU equality,
unique ownership, source coverage, canonical coefficients and retained state-chain
checks. All **516 normal proof events** matched (384 support-index and 132
polynomial endpoints, with paired identities and exact support coverage).
Small-case memcheck passed eight rank measurements and 24 proof events; all four
processes reported **zero errors and zero leaked bytes**. Both chunk widths and
actual staging allocations were checked. No disagreement snapshots were triggered.
The [earlier direct-buffer failures](../root-transfer-diagnostics/report.md) remain
unexplained; this staged test is a validated bounded candidate, not proof of cause.

| Fixture (partial support) | Timed passes | Median route + GPU roots + tree + verification |
|---|---:|---:|
| Genuine s=8, 7,381 exponents | 2 | 0.014324 s |
| Synthetic s=17, 65,536 exponents | 2 | 0.318970 s |
| Synthetic s=17, 262,144 exponents | 12 | 1.201005 s |

Whole normal-run largest logged process peak RSS: **411.309 MiB per rank**;
largest owned GPU arrays: **15.241 MiB per rank** (driver/context allocations
additional). Cache retained: **27.567 MiB**. Normal command wall time: **38.258 s**,
including setup, references, warm-up and artifact passes; memcheck command:
**4.283 s**. Per-pass medians exclude reusable setup/CPU oracle, warm-up, validation
snapshot pass and profiler summary; they include routing, hashing, proof traffic,
GPU roots/tree and final equality. Rank maxima are measured per phase and are not
additive. These are profiled diagnostics on shared cards, not a full s=17/s=18
preprocessing measurement or a controlled direct-versus-staged speed comparison.

All four saved CPU profiles were independently read with `pprof -symbolize=none`
under a 20 s command deadline; [recovered summaries](runs/staged-01/profile-recovered-0.log),
[commands and hashes](runs/staged-01/review-checks.json). This offline summary skips
external/native symbol lookup; full raw profiles remain available. The reason
for the original timeout is unconfirmed. Future shared tree runs explicitly use
this offline summary option after timing, preventing dependence on external
symbolization; no arithmetic, benchmark clock or stored profile is changed.

Logs retain OpenFabrics initialization warnings and a peer-node fallback from
the node-local source working directory to `/tmp`. Executable/payload paths were
absolute and all checks passed; RDMA transport use was not established. Do not
interpret these small shared-card timings as fabric or GPU scaling throughput.

Next: the [controlled direct/staged comparison](../mpi-buffer-comparison/report.md)
uses identical new proof instrumentation and one build before choosing a transport
default or assigning a runtime penalty. Job 77868 passed staged-a, then captured
a direct-a same-node payload disagreement; later arms were unrun. No speed ratio
or underlying cause is established. See the comparison report for paired evidence.
Keep staging explicit; the bounded pass allows further validation with that mode,
not clearance of the earlier direct-buffer failures or full-size integration.
Distributed discovery, larger distributed products/FFTs and publication remain
required for full s=18 preprocessing. No new job submitted during review.

[Runner](run.py), [scheduler script](run.sbatch), [shared suite](../root-merge-tree/run.py),
[test harness](../../encoding/liftrs-shared/root_tree_gpu_test.go),
[MPI implementation](../distributed-root/mpi/world.go),
[reviewer](../root-transfer-diagnostics/analyze.py). The checked source revision
is recorded in submission.json; exact frozen build inputs and all file hashes
are recorded by the runner in results.json. No arithmetic kernel changes.

This experiment explicitly selects `-root-mpi-buffer=staged`. One reusable
`MPI_Alloc_mem` chunk is copied before each synchronous MPI send and after each
receive. Payload buffers are not retained by MPI after those blocking calls.
Both uint64 support indices and uint32 coefficients use this implementation.
Count and diagnostic proof records also pass through C-local buffers. Direct Go
payloads remain the default; staging is an experimental option, not a proven fix.
The legacy two-rank Peer exchange is unchanged and is not used by this probe.

Normal chunks: 65,536 words; memcheck chunks: 1,024 words. Scratch is sized for
uint64, **512 KiB per rank / 2 MiB total** normally, **8 KiB per rank** under memcheck.
Each send and receive adds one CPU copy; no transforms, GPU allocations or
production disk output are added. MPI internal memory is not included in owned
scratch. Runtime/setup penalty is **unknown**, and hashing/control exchanges also
change timing. Even a pass cannot establish a root cause or a speedup versus the
failed direct runs. No slower path becomes a production default.

Both packet widths now share the symmetric pre/post hash handshake and durable
paired-capture acknowledgement. Index fingerprints preserve every bit, including
invalid high bits, before range validation. Captures include phase, per-peer
packet/stride sequence, width, both proofs and complete current bytes. Empty
support terminators carry no payload proof. Pairing checks use phase/sequence to
avoid mixing multiple packets. Original range/degree/union/CPU equality checks stay.
The genuine s=8 and synthetic s=17 fixtures are unchanged. Smaller cases run two
timed passes plus warm-up/artifact; largest 262,144-exponent case runs 12 timed
passes plus warm-up/artifact. Four profiles and small-case four-rank CUDA memcheck
are required for a complete pass. A failure may prevent those final steps.

HPC resource envelope: two nodes, one H200/node (two total), two logical ranks/node,
one CPU/rank (four total), 4 GiB requested host RAM/node (8 GiB total), 1 GiB process
RSS guard/rank. Owned device cap: 64 MiB/rank, at most 128 MiB/card; contexts and
runtime overhead are additional. Existing limits: 128 MiB retained data, 2 GiB
scratch, 32 MiB/file, 720 s controller and 15 min scheduler. Staging remains inside
the host budget. No resource increase or full-size run is requested. gnode03 is
excluded as in the existing runner; other accessible GPU nodes are eligible.

Reproduce from the committed implementation on HPC (Go 1.27.3, GCC 14.2,
CUDA 12.9.1, OpenMPI 4.1.8, Slurm 25.05):

```sh
python3 -B experiments/mpi-owned-staging/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/mpi-owned-staging/slurm-%j.log \
  experiments/mpi-owned-staging/run.sbatch staged-01
# Review only after the user reports completion:
python3 -B experiments/root-transfer-diagnostics/analyze.py staged-01 \
  --experiment experiments/mpi-owned-staging
```

Results/logs/profiles go to `runs/staged-01/` and `slurm-JOBID.log`; reusable build,
source and payload/failure snapshots remain outside Git at
`.cache/liftrs/mpi-owned-staging/staged-01/`, with checksums in results.json.
Existing caches/profiles are retained. Timestamped phase/progress/resource logs
and native MPI chunk completion logs remain enabled; no agent monitoring.

Local readiness: [native/MPI/CUDA build](readiness/build.log) passed; the two
[fingerprint tests](readiness/proof-tests.log) passed, including the 131,073-word
coefficient tail and invalid uint64 high bits. Three Python review tests cover 32-bit tails, 64-bit corruption and missing/
incomplete support proofs. Existing 3 tree / 4 distributed / 5 staged /
3 common-runner reviewer tests passed. [Exact commands and input hashes](readiness/checks.json).
These local pre-submission checks execute no MPI/GPU workload. Actual bounded
normal/memcheck results and their review are recorded above.
