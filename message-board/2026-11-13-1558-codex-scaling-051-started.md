# Native wide CPU transform kernel

- Timestamp: 2026-11-13T15:58:20.841270+00:00
- Agent: Codex
- Work: SCALING-051
- Status: started

User authorizes native CPU port after scalar validation74290. Read board/README, native transform/ABI/Go setup and scalar reference. Generalize only the shared native polynomial/recursive transform kernel over32/64-bit word types, preserving32-bit aliases and C ABI admission. Add bounded trusted-fixture tests with independent Go polynomial evaluation, high offsets/ordered bases, scalar/four-bit products, inverse round trips, old32-bit agreement and native sanitizers. Full matrix schedules, wide C ABI/cache/preprocessing remain separate gates. No active-job inspection.
