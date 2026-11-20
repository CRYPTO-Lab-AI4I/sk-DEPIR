# UCX registration-cache control

Status: prepared (SCALING-135), unmeasured. [Runner](run.py),
[scheduler script](run.sbatch), [shared host probe](../mpi-host-buffers/probe_test.go),
[orchestration](../mpi-host-buffers/run.py), [reviewer](../mpi-host-buffers/analyze.py).
Implementation revision in submission.json; source/evidence/cache seals in results.json.

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
