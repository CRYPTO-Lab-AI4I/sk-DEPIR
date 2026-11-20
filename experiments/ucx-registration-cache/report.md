# UCX registration-cache control

Status: **78218 passed all four host-replay arms** (SCALING-136).
[Receipt](submission.json), [raw results](runs/no-cache-01/results.json),
[review](runs/no-cache-01/analysis.json), [accounting](runs/no-cache-01/scheduler-accounting.txt),
[progress](slurm-78218.log). Implementation `19e6143`; verified
**658 source / 33 original evidence / 3 cache hashes**, all commands exited zero.
[Runner](run.py), [scheduler script](run.sbatch),
[reviewer](../mpi-host-buffers/analyze.py). Raw evidence remains unchanged.

## Review of 78218

Every worker reported the requested cache-disabled settings. Actual-node UCX1.18
configuration contains `UCX_RCACHE_ENABLE=n`; native logs select UCX on all ranks
in every arm. All **1,024 handoffs / 2,048 endpoint proof checks** matched expected
root/index payloads; no mismatch captures. Both direct arms passed, including the
forced-reclamation stress that was previously unrun after direct failures.

| Arm | Result | Command elapsed | Largest logged process peak RSS/rank |
|---|---|---:|---:|
| staged-normal | 256 handoffs / 512 proofs passed | 2.071 s | 296.418 MiB |
| staged-reclaim | 256 handoffs / 512 proofs passed | 2.327 s | 264.297 MiB |
| direct-normal | 256 handoffs / 512 proofs passed | 2.124 s | 294.934 MiB |
| direct-reclaim | 256 handoffs / 512 proofs passed | 2.225 s | 265.898 MiB |

HPC gnode13/gnode14: 2 nodes / 4 CPU cores / 8 GiB requested host total / zero GPUs.
Scheduler elapsed **1 min 2 s**, controller **59.869 s**. Build17.223 s and offline
profile summary28.106 s are separate from replay commands. Timings include MPI
process startup, churn, synchronization, hashing, proofs and profiling; these are
not pure network/copy latency or a controlled full-size speed comparison. All CPU
profiles retained; source/cache/hash checks pass with no resource guard trigger.

The pass supports **registration-cache-dependent behavior in this reproducer**.
It does not establish a permanent UCX/Go fix, correctness at arbitrary sizes or the
cost of cache disabling in full preprocessing. Prior raw failures remain failed.
**Next:** validate the [GPU merge-tree crossover](../ucx-cache-gpu-tree/report.md)
with the same explicit setting, CPU equality, state/proof checks and CUDA memcheck,
using the previously admitted two-H200 envelope. No production default change;
full s=18 remains unrun.

[Discovery 78009](../mpi-transport-discovery/report.md#review-of-78009) selected
UCX on every rank in both passing staged arms and the failing direct arm. The
same cross-node direct payload disagreement reproduced without GPU arithmetic.
This control tests a narrower hypothesis: cached memory registrations may interact
with reused/reclaimed Go heap pages. That mechanism is **not established**.

Keep the observed UCX PML (`OMPI_MCA_pml=ucx`) and explicitly set
`UCX_RCACHE_ENABLE=n`. Retain the same PML/BTL/UCX diagnostic logging as discovery.
No TCP forcing, field/GPU changes, new payload, allocation pattern, route, chunk
size, GC control or proof logic. The four existing arms are staged-normal,
staged-reclaim, direct-normal and direct-reclaim; stop on failure and retain both
payloads. The ordinary arms still contain artificial allocation churn; forced
reclamation is a separate stress condition, not a production optimization.

The OpenMPI module's `mca_pml_ucx.so` resolves to **UCX 1.18.0**:
[library resolution](readiness/pml-libraries.log), [version](readiness/ucx-version.log).
The system `/usr/bin/ucx_info` is a different 1.20 installation and is not used.
The linked 1.18 tool documents registration-cache enablement as yes/no/try and
accepts n; [configuration](readiness/ucx-config.log). The runner records version
and effective configuration again on the allocated HPC node before replay,
requiring `UCX_RCACHE_ENABLE=n`. Worker telemetry records the requested setting;
review rejects missing/default cache settings. Native PML selection is recorded
separately from component availability. These are configuration/selection checks,
not a proof that every internal UCX operation is cache-free.

Explicit Go buffers/staging, GPU memory and output sizes do not change. Disabling
registration caching may increase registration/unregistration work and transfer
latency, and may change UCX-internal pinned memory use. **Magnitude is unknown**;
process RSS and resource guards remain measured. No speed or memory saving is
assumed. This is a diagnostic option only, not a production default. A pass supports
cache-dependent behavior under this replay, not a permanent fix or clearance of
earlier failures. Follow-up GPU-tree correctness and measured end-to-end costs
would still be needed before adopting it. Existing validated staged paths remain.

HPC budget unchanged: **2 nodes / 4 ranks / 4 CPU cores / 8 GiB requested host RAM
total / zero GPUs**; 1 GiB process RSS guard/rank, 16 MiB churn/rank plus sub-MiB
payload/hash buffers, at most 512 KiB owned staging/rank (2 MiB total). Caps:
128 MiB retained, 2 GiB scratch, 32 MiB/file, 720 s controller / 15 min scheduler.
Arm test/step/command deadlines remain 90/100/120 s; profile summary 90 s.
Bounded component/proof/capture/progress logs and all CPU profiles retained.
Estimated **1–3 min after scheduling**, not measured. No resource increase or
full s=18 capability claim; no monitoring after submission.

```sh
python3 -B experiments/ucx-registration-cache/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/ucx-registration-cache/slurm-%j.log \
  experiments/ucx-registration-cache/run.sbatch no-cache-01
# After user completion ping only:
python3 -B experiments/mpi-host-buffers/analyze.py no-cache-01 \
  --experiment experiments/ucx-registration-cache
```

Results/profiles/actual-node configuration: `runs/no-cache-01/`; progress:
`slurm-JOBID.log`; replay/binary/captures outside Git at
`.cache/liftrs/ucx-registration-cache/no-cache-01/`, sealed in results.json.
Fixture provenance and toolchain as in [host replay](../mpi-host-buffers/report.md).
Use the recorded revision and a fresh name to regenerate. No useful caches removed.

[Local checks](readiness/checks.json): tagged host build and hash/proof unit pass;
31 Python reviewer tests pass, including rejection of missing/default cache control
and distinguishing selected PML from available components. Python/scheduler syntax
and whitespace checks pass. No local MPI execution or causal/performance claim.
