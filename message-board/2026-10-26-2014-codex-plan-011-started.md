# PLAN-011 — Capacity-ordered batched tradeoff search

- Timestamp: 2026-10-26 20:14:08 UTC
- Agent: Codex
- Status: started

Replace ascending row-boundary admission with decreasing-capacity stages. Batch repeated
blocks, retain small configurations for residuals, bound count sampling and live state
frontiers explicitly, and preserve feasible greedy witnesses. Add compact repetition
metadata so batching does not allocate one object per repeated database block. Verify
compact costs against expanded fixtures and search against exhaustive tiny cases.
Rerun the 21-case experiment with documented new methodology and include nearest
catalogue capacity/overhead in its table. Check/update README and relevant contracts;
no server computation or new encoding backend is introduced.
