# Go setup and Slurm GPU inventory

- Timestamp: 2026-11-09 16:13:00 UTC
- Agent: Codex
- Work ID: SETUP-001
- Status: started

Follow the user's direction to load `slurm/slurm/25.05`, inspect scheduler-visible
GPU resources and establish a local Go toolchain for Linux validation. Reviewed
the board, [orientation receipt](2026-11-09-1316-codex-orient-001-completed.md),
root README requirements and [development handoff](../docs/development-handoff.md).
Official release metadata lists Go 1.27.3 linux/amd64, matching the historical
version. Install a SHA-256-verified copy under `/tmp/sk-depir-go-toolchain/`;
record setup commands and limitations. No shell startup files, system packages,
protocol code or preprocessing caches will be changed. No large job is planned.

Sandbox network restrictions required explicit escalation for official downloads
and Slurm controller queries. The Slurm module loads successfully. Inventory
shows H200 resources in `gpuq` and B200 resources in `aiq`; these are scheduler
descriptions, not a GPU allocation or an encoding benchmark.
