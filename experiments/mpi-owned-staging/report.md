# Bounded MPI-owned staging probe

Status: prepared for a diagnostic run (SCALING-125); no MPI/GPU result yet.
The [previous diagnostic](../root-transfer-diagnostics/report.md) failed on an
uncovered uint64 support packet after 78 matching polynomial proofs. The original
77704 coefficient-transfer disagreement and 77782 support failure remain unresolved.

[Runner](run.py), [scheduler script](run.sbatch), [shared suite](../root-merge-tree/run.py),
[test harness](../../encoding/liftrs-shared/root_tree_gpu_test.go),
[MPI implementation](../distributed-root/mpi/world.go),
[reviewer](../root-transfer-diagnostics/analyze.py). The checked source revision
will be recorded in submission.json; exact frozen build inputs and all file hashes
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
These checks execute no MPI/GPU workload and cannot validate the new buffer mode.
