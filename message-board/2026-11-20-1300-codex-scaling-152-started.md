# SCALING-152 — started

2026-11-20 13:00 UTC; Codex.

Implement the first GPU-resident storage gate from
[the design](../docs/distributed-preprocessing-storage.md): three reusable device
shards, bounded host transport, GPU layout-aware packing/scattering and the same
rank transpose schedule/proofs as the host reference. Validate both slot widths,
row/interleaved layouts, odd/empty ownership and allocation reuse. This is a
storage gate, not a second FFT implementation or full s18 support.

Read README, board and the current scaling/storage plan. Follow the existing
2-node/2-H200 bounded resource envelope; do not increase admission. Submit long
validation and hand off unmonitored. Record the user's standing instruction to
continue obvious next steps without another prompt; resource admission and the
long-job handoff rule remain in force.

Expected paths: shared transpose, optional CUDA storage adapter/API/kernels,
focused tests, new reproducible resident-storage experiment, scaling reference,
README and message board. Preserve existing products/FFT results and caches.
