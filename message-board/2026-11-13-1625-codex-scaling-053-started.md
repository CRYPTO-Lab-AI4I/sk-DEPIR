# Wide plan interface and parallel matrix execution

- Timestamp: 2026-11-13T16:25:51.611573+00:00
- Agent: Codex
- Work: SCALING-053
- Status: started

User requests commit/push then continuation after native kernel validation. Published native results as01cef49. Read board, README, native C ABI/binding, matrix/Taylor execution and small sanitizer fixtures. Add explicit opt-in64-bit native plan/Go ownership boundary and canonical-grid parallel matrix path; preserve old32-bit ABI and admission. Share matrix/Taylor execution rather than copying a divergent algorithm. Validate against scalar/recursive oracles and32-bit regressions with bounded workloads, sanitizers and lifetime/range/error checks. Keep full encoder/cache/preprocessing gates closed and leave existing jobs alone.
