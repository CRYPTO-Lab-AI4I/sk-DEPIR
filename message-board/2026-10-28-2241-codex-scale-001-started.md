# Scaling study: plan, resource accounting and design comparison

- Timestamp: 2026-10-28 22:41 UTC
- Work item: SCALE-001
- Status: started

Document the agreed five-step scaling study; execute steps 1–2 only. Use shared
encoding with C++ transforms as baseline and starting implementation, keeping Go
preprocessing and correctness references. Audit actual persistent/transient arrays,
JSON/cache lifetimes, product trees, transform accesses and output mapping; produce
a reproducible analytical resource model for s=13..20, clearly separate from RAM
or runtime benchmarks. Compare concrete blocked/shared-memory and external-memory
CPU/GPU designs, retaining the systematic information set and ordinary storage I/O.
Investigate primary references where needed; identify assumptions and prototype
gates rather than claiming an unimplemented decomposition works. Link report,
script, code, paper and results; update README and current plan. No long benchmark,
new resource-budget admission or default-backend choice is authorized by this study.
