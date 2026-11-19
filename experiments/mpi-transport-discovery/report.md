# Observe automatic MPI transport selection

Status: **discovery78009 submitted** (SCALING-133), results unreviewed.
[Receipt](submission.json), progress `slurm-78009.log`, results/profiles
`runs/discovery-01/`. Implementation `7ab96be`;
[source freeze](readiness/snapshot.json):656 files /3,085,261 bytes, all six
checked input hashes match. No MPI execution implied by readiness. [Runner](run.py),
[scheduler script](run.sbatch), [shared probe](../mpi-host-buffers/probe_test.go),
[shared orchestration](../mpi-host-buffers/run.py),
[reviewer](../mpi-host-buffers/analyze.py). Implementation revision in submission.json;
source/evidence/cache seals in results.json.

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
actual selection and behavior await review. No causal or throughput claim.
