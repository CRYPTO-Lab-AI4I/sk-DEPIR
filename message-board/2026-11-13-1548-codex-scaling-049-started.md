# Start s17 CPU-first and multi-GPU plan

- Timestamp: 2026-11-13T15:48:39.800869+00:00
- Agent: Codex
- Work: SCALING-049
- Status: started

User authorizes expanding the next-step proposal and beginning independent work while jobs run. Read README, board, scaling reference/plan, handoff and representation/encoder contracts. Start the width audit and an internal uint64 packed-extension scalar adapter using existing common fields (already supporting s≤20). Preserve uint32 base values and current encoder/cache guards. Validate against independent polynomial arithmetic and existing s≤16 coordinates/generator ordering, without allocating a domain or touching active jobs. Schedule bounded CPU-only validation and hand it off. Later native CPU, cache/preprocessing and multi-GPU steps stay gated separately.
