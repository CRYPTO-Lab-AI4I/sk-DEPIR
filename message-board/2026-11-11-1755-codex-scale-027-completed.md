# SCALE-027 — Memory optimization is the next plan step

Timestamp: 2026-11-11T17:55:00.250170+00:00
Agent: Codex
Status: completed

Expanded [scaling plan](../docs/encoding-scaling.md#current-step-memory-optimization-for-s17) with phase/lifetime memory audit, bounded-memory alternatives, explicit user choice for material memory–time trade-offs, s16 prototype verification and an evidence-based s17 arithmetic/resource gate. Marked CUDA integration complete. Record rough wide-port memory models as estimates, not measurements/budgets. Defer standalone packed-writing/kernel runtime work unless required by the selected memory schedule. Preserve CPU/GPU main paths, baseline hashes, profiles and caches.

Updated [README](../README.md) next-step pointer and [contributor instructions](../AGENTS.md) with the user's standing trade-off decision rule. Comparisons must expose host/device memory, disk, transfers/recomputation, setup, first/reused encoding time, deployment cost and uncertainty before adopting a material trade-off. No option is selected, implementation changed, experiment launched or budget increased in this documentation item.

Verification: relative documentation paths/current-step anchor and working whitespace checks pass. No runtime tests needed for documentation-only changes. No commit/push requested. s17 remains unsupported; application/cryptographic integration and TRL5 remain open.
