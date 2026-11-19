# SCALING-127 — prepared

UTC: 2026-11-19T13:46:45.465458+00:00

Prepared [same-build direct/staged comparison](../experiments/mpi-buffer-comparison/report.md).
Split shared tree runner into one preparation and reusable prefixed execution;
default suite preserves existing paths/clocks. Fresh normal execution clears
memcheck-only state. No Go/C/CUDA kernels or protocol changes.
New runner/reviewer reuses shared source/ownership/CPU equality/state/proof checks,
verifies mode/scratch/placement/build hashes and cross-arm byte hashes, computes
matched median ratios, saves arm summaries durably and stops on any failure.
Staged/direct/direct/staged sequence in one allocation; all profiles/caches kept.

5 comparison / 3 tree / 4 proof / 4 distributed / 5 staged / 3 shared runner tests
pass (24 total), including failure persistence, output/placement mismatch,
mode/scratch checks, timing exclusions and fresh-arm environment/path isolation.
Scheduler/Python syntax and whitespace checks pass.
[Readiness](../experiments/mpi-buffer-comparison/readiness/checks.json).
No local MPI/GPU execution. Native tagged build, historical fixture/reference
checks and shared/race tests run once in submitted job; no native source changed.

README checked/updated, main scaling reference/plan and experiment index current;
shared-tree and previous-stage reports link the new experiment. Same admitted
2 HPC nodes / 2 H200 / 4 logical ranks / 4 CPUs / 8 GiB host, 1 GiB RSS/rank,
64 MiB owned GPU/rank, 128 MiB retained / 2 GiB scratch, 15 min scheduler envelope.
71.838 MiB projected cache, 4–7 min projected wall time after scheduling. Staging
adds 512 KiB/rank and copies; runtime penalty unknown. No resource/default change.
Submit after freezing exact Git inputs, then hand off without monitoring.
