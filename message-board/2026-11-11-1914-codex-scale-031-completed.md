# SCALE-031 — Shared CPU/GPU memory plan and slower-mode rule recorded

Timestamp: 2026-11-11T19:14:48.801813+00:00
Agent: Codex
Status: completed

Expanded [scaling plan](../docs/encoding-scaling.md#next-part-shared-cpugpu-memory-schedules) alongside pending CPU job73586: remove full scratch during transforms; stream preprocessing constants; stream canonical/packed output with explicit ownership/interface; evaluate host/disk-backed transforms; combine/prove schedules on small cases/s16 before s17 arithmetic/resource admission. Distinguish GPU host-backed uploads from CPU disk-backed reads, temporary storage/pinned/page-cache accounting, extra passes and setup amortization. Shared decomposition does not imply identical CPU/GPU kernels or unmeasured feasibility.

Updated [contributor policy](../AGENTS.md), [encoder contract](../docs/encoder-contract.md#memory-mode-selection-requirements) and [README](../README.md). Faster validated mode remains default when it fits admitted RAM/VRAM with overhead/headroom. Slower memory-saving modes address insufficient memory; when memory is sufficient they require explicit command-line opt-in. Preserve fast path, owning-output semantics, explicit backend choice and measured/estimated time/storage trade-offs. No CLI flag or automatic capacity selection is implemented or claimed.

Verification: documentation paths, plan anchor and working whitespace pass. No runtime tests needed for documentation-only edits. CPU job state/process/logs/results were not inspected or altered; await user completion ping. No agent/job launched, budget changed, implementation selected or CPU/GPU runtime behavior changed by this work item. Pending CPU implementation/tests remain staged separately from the already published GPU checkpoint. No commit/push requested for this documentation turn.
