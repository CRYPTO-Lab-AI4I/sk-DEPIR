# SCALING-133 — prepared

UTC: 2026-11-19T17:44:31.791164+00:00

Host replay reuses all transport/proof/GC/fixture logic; only records/enables PML20/BTL40/UCXinfo logging and clears forced PML/BTL environment overrides. Observe native selected components before another backend change. Same2nodes/4CPUs/8GiB host/zeroGPUs and existing deadlines/storage caps, no default change. Host build/hash unit and29 Python tests pass, syntax/whitespace checked; no local MPI. [Report](../experiments/mpi-transport-discovery/report.md), [checks](../experiments/mpi-transport-discovery/readiness/checks.json). Estimated1–3min after scheduling. Logging overhead/fault perturbation unknown. Freeze/submit then no monitoring.
