# HPC and laptop context made explicit

- Timestamp: 2026-11-17T08:17:43.445578+00:00
- Agent: Codex
- Work: DOC-094
- Status: Completed

Replaced ambiguous machine/host references in maintained documentation and
experiment-report prose. Historical macOS measurements are labeled laptop;
Linux cluster measurements and capacity discussions are labeled HPC. Added a
compact hardware key and the full s17 CPU/GPU job allocations to the scaling
reference, distinguishing hardware capacity, requested resources and measurements.
Recorded the terminology rule in AGENTS.md for future work. Updated the capacity
heading link in the scaling plan. Historical board messages and sealed raw
measurement evidence are preserved.

Changed documentation:
- [AGENTS.md](../AGENTS.md)
- [README.md](../README.md)
- [cmd/pir-preprocess/README.md](../cmd/pir-preprocess/README.md)
- [docs/development-handoff.md](../docs/development-handoff.md)
- [docs/encoding-algorithms.md](../docs/encoding-algorithms.md)
- [docs/encoding-scaling.md](../docs/encoding-scaling.md)
- [docs/encoding-status.md](../docs/encoding-status.md)
- [experiments/encoder-profile-s13/report.md](../experiments/encoder-profile-s13/report.md)
- [experiments/gpu-scaling/report.md](../experiments/gpu-scaling/report.md)
- [experiments/hpc-cached-s15/report.md](../experiments/hpc-cached-s15/report.md)
- [experiments/local-calibration/report.md](../experiments/local-calibration/report.md)
- [experiments/parallel-encoding/report.md](../experiments/parallel-encoding/report.md)
- [experiments/parallel-s14/report.md](../experiments/parallel-s14/report.md)
- [experiments/streamed-constants/report.md](../experiments/streamed-constants/report.md)

Verification: reviewed contextual platform attribution against linked reports;
checked naming with a case-insensitive scan (including line breaks) over changed
documentation, checked the renamed capacity-heading reference, and ran
`git diff --check`. README checked and updated for laptop timing hardware and
HPC four-GPU nodes. Documentation only: no executable changes, tests or new jobs.
Existing performance results, resource approvals and implementation limits remain
as recorded; this work does not remeasure hardware or jobs.
