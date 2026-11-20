# Observe automatic MPI transport selection

Status: **78009 selected UCX and reproduced direct corruption** (SCALING-134).
[Receipt](submission.json), [raw results](runs/discovery-01/results.json),
[review](runs/discovery-01/analysis.json), [component evidence](runs/discovery-01/component-review.json),
[accounting](runs/discovery-01/scheduler-accounting.txt), [progress](slurm-78009.log).
Implementation `7ab96be`; raw failed status/logs/captures unchanged. Verified
**656 source / 25 original evidence / 5 cache file hashes**.
[Runner](run.py), [scheduler script](run.sbatch),
[shared reviewer](../mpi-host-buffers/analyze.py).

## Review of 78009

All four ranks explicitly selected the **UCX PML**, in both passing staged arms
and failing direct-normal arm (12 native selection records). OB1 was explicitly
not selected; OpenIB BTL initialization warnings do not establish the active data
path. This resolves the earlier uncertainty about automatic backend selection;
it does not establish which UCX mechanism caused the fault.

| Arm | Result | Completed handoffs | Command elapsed | Largest logged process peak RSS/rank |
|---|---|---:|---:|---:|
| staged-normal | Passed all 512 proof endpoints | 256 | 2.331 s | 293.758 MiB |
| staged-reclaim | Passed all 512 proof endpoints | 256 | 2.326 s | 262.832 MiB |
| direct-normal | Captured cross-node mismatch, cycle 0 / case 1 | 1 before failure | 1.118 s to abort | Not summarized as complete run |
| direct-reclaim | Not run | 0 | Unknown | Unknown |

HPC gnode13/gnode14: 2 nodes, 4 ranks/CPU cores, 8 GiB requested host RAM total,
zero GPUs. Scheduler **28 s**, controller **27.755 s**; no resource/time guard
triggered. Rank 2→0 sent 262,148 bytes; both captures match their locally stable
proofs but differ across peers. First zero-based byte/word 0: sender 1, receiver 0,
matching the earlier host reproduction. All paired paths/hashes are in the review.
This is a profiled, allocation-churn diagnostic, not a throughput baseline.

The selected module resolves UCX 1.18, whose installed tool supports disabling
its registration cache. **Next:** the [explicit cache control](../ucx-registration-cache/report.md)
keeps UCX and the host replay, sets `UCX_RCACHE_ENABLE=n` (job 78218 submitted), and records effective
configuration/worker settings. Repeated registration may cost time; internal-memory
and runtime effects are unknown. Diagnostic only, no production default change.
Full s=18 and a permanent direct-buffer repair remain unvalidated.

[TCP control78001](../mpi-transport-control/report.md#review-of-78001) stalled on
its first staged transfer; it is not a usable corruption or throughput control.
Return to automatic transport selection and record which components are selected
before changing any additional data path. No arithmetic/GPU work is involved.

Reuse the captured public polynomial, four host-replay arms, normal allocation
churn/explicit reclamation conditions, routes, chunks, expected hashes, paired
captures, profiles, guards and heartbeat logs from the
[host replay](../mpi-host-buffers/report.md). Enable only diagnostic logging:
`OMPI_MCA_pml_base_verbose=20`, `OMPI_MCA_btl_base_verbose=40`, `UCX_LOG_LEVEL=info`.
Installed OpenMPI metadata lists BTL40 as info-level logging. Clear the two
per-job forced PML/BTL environment overrides; site defaults still apply.
Every worker records the requested logging fields and PML/BTL override fields;
review checks all requested values, allowing additional recorded diagnostic keys.
Actual selected components must be read from native logs, not inferred solely
from requested environment settings. UCX may be unused; logging it does not
imply selection. No component cache, runtime reclamation setting or algorithm
is changed. No default decision follows from this diagnostic.

Stop on a failed arm, preserving native component logs, CPU profiles, both
payload captures and completed controls. Prior default host run78000 already
reproduced corruption, so clean execution here would only mean not reproduced
with logging. More logging perturbs timing/I/O; these are not latency baselines.
Native log details should guide the next control. Full s=18 remains unrun and
the explicit staged path retains its earlier bounded validation.

HPC allocation unchanged: **2 nodes /4 ranks /4 CPU cores /8 GiB requested host
RAM total /zero GPUs**; 1 GiB process RSS guard/rank,16 MiB churn/rank plus sub-MiB
payload buffers, at most512 KiB staging/rank. Caps:128 MiB retained,2 GiB scratch,
32 MiB/file,720 s controller/15 min scheduler;90 s test/100 s step/120 s command
per arm,90 s profile report. Additional bounded trace output is retained under
the existing file cap; overhead is unknown. No larger resources or slow default.
Estimated **1–3 min after scheduling**, not measured; possible early failure.

```sh
python3 -B experiments/mpi-transport-discovery/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/mpi-transport-discovery/slurm-%j.log \
  experiments/mpi-transport-discovery/run.sbatch discovery-01
# After user completion ping only:
python3 -B experiments/mpi-host-buffers/analyze.py discovery-01 \
  --experiment experiments/mpi-transport-discovery
```

Results/profiles and full native selection logs:`runs/discovery-01/`; progress:
`slurm-JOBID.log`; replay/binary/failure captures outside Git:
`.cache/liftrs/mpi-transport-discovery/discovery-01/`. Exact fixture provenance,
hashes and toolchain as in host replay. Fresh name plus recorded source revision
regenerates the run. No useful cache/profile removed; no agent monitoring.

[Local checks](readiness/checks.json): tagged host build/hash/proof unit passes,
29 Python review tests pass. New parser regression distinguishes Go timeout stack
struct arguments from JSON events while still rejecting malformed event JSON.
Native instrumentation adds explicit requested settings only. No local MPI run;
actual selection and failure behavior are reviewed above. No causal or throughput claim.
