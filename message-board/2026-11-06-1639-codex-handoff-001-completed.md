# HANDOFF-001 completed: larger-machine continuation documented

Timestamp: 2026-11-06T16:42:05.127456+00:00. Agent: Codex.

Created the tracked [development handoff](../docs/development-handoff.md) and made
it required reading in [AGENTS.md](../AGENTS.md). Root README, roadmap, scaling
study and encoding directory README now identify the move: stop large laptop
work at the verified s=15 checkpoint; continue on a single shared-memory commodity
CPU node, provisionally 256 GiB RAM. s≥16 remains unverified, not promised HPC/GPU
support. TRL 5, client/preprocessing-only scope and generic/reference paths remain.

The handoff records code/results checkpoint d721182, measured s14/s15 baseline,
C++ online versus Go preprocessing boundaries, public/experimental guards and
unintegrated application/cryptographic work. It lists toolchain/tests and macOS-only
runner controls that require adaptation on Linux/HPC, s16 size/header/index changes,
resource/profiling admission, NUMA/job allocation accounting and larger-RAM versus
bounded-memory design choices. No unbounded large run or specialized architecture
is selected. Historical laptop time/RAM budgets are not new-node defaults.

Large caches remain intact outside Git. The source-host cache paths, data versus
binary portability, exact committed checksum verification and optional retained
outputs/raw profiling provenance are documented. Git clone alone does not migrate
those files. No transfer, destination setup or new benchmark has been performed.

Verification: updated relative links and embedded checksum Python syntax pass;
diff whitespace passes. Documentation-only change: no code or benchmark changed,
so backend suites were not repeated. Root README explicitly checked and updated.
Publishing the handoff ensures the next checkout and its agents receive the same
instructions; preserve local caches until the destination confirms verified transfer.
