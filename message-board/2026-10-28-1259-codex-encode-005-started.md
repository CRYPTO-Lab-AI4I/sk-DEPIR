# Scale CPU encoding to s=10

- Timestamp: 2026-10-28 12:59 UTC
- Work item: ENCODE-005
- Status: started

Enable the next catalogue configuration s=10,m=2,t=6 with the current JSON artifact
format. Compute exact capacity, prepare/compare cyclic and paper encoders where
feasible, verify systematic positions and sampled actual-curve constraints, and
report discovery, preparation, load, encoding, memory and packed output I/O
separately under `experiments/encoder-scaling/`. Reuse public code-only constants
across databases. Record artifact-size or loader-limit obstructions before making
small compatibility changes; defer format redesign unless it blocks measurement.
Update README, catalogue evidence, encoder contracts and the plan to prioritize
scaling measurements before artifact cleanup. Keep larger-instance/production
claims narrow. Existing results and artifacts remain unchanged.
