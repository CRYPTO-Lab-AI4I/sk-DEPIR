# SCALING-141 — started

UTC: 2026-11-20T10:03:39.468316+00:00

User authorized collective product/distributed FFT implementation. First concrete component: backend-neutral32/64-bit sharded transpose with bounded tiles and deterministic peer schedule, independent scalar/permutation tests, MPI replay validation. No rank gathers a full matrix in the executor. Two owned local shards plus one tile buffer; packing/communication/runtime costs unknown, reference/control-plane only, not a CPU preprocessing fallback or production default. Reuse existing transport with a separate wide-coefficient lane. Document exact APIs, invariants, progress and budgets. Native GPU stage integration follows this prerequisite; do not claim full distributed FFT or s18 support.
