# SCALING-142 — completed

UTC: 2026-11-20T10:33:09.549215+00:00

Verified **78279 passed**:672 source/39 original evidence/18 cache hashes,
all commands zero, both modes all8 shapes/width operations (16 global transposes /
64 rank outputs),7,312 paired proofs matching independent expected tile bytes.
Every output/saved shard word, source/ownership metadata and legacy root/support
lanes pass; no captures/guards/recovery. HPC gnode14/15,2nodes/4ranks/CPUs/8GiB host,
zeroGPUs. Scheduler1min18s/controller76.975s. Largest1024x256 single samples:
staged32/64bit6.257/5.476ms, direct8.865/4.697ms. Logged peakRSS249.312/248.145MiB
per rank; explicit arrays1.125MiB/rank plusoptional8KiB MPI staging.12.666MiB cache,
eight rank CPU profiles and largest outputs retained. Single profiled diagnostic
samples and varying tile sizes are not GPU/network/fullsize speed estimates.

[Report](../experiments/distributed-transpose/report.md#review-of-78279),
[review](../experiments/distributed-transpose/runs/transpose-01/analysis.json),
[commands/hashes](../experiments/distributed-transpose/runs/transpose-01/review-checks.json).
Independent Python review rechecks complete routes/counts/headers/peer proofs,
expected hashes and all saved output bytes.2 reviewer regression tests pass;
whitespace checked. No source code changed, no new native build/MPI/GPU run needed.
README checked/updated; main scaling reference/plan, preprocessing/shared docs,
experiment index current. Fulls18 remains unrun, no production/default change.

Inspected GPU preparation/vector/Taylor kernels and HostTiled adapter. Next expose
stage-scoped row/column operations using existing local_vectors/vector_transform,
with global domain/first-vector coordinates independent of admitted local device
buffer. Validate32/64-bit both directions, nonzero/wide origins against CPU results;
then compose validated transpose and distributed Taylor dependencies. Current
rows_at adapter is diagnostic, not selected production memory layout. Reuse GPU
arenas and bounded host exchange; no adoption of two full host shards (projected
1TiB/16ranks) or unmeasured material trade-off. No new job launched by this review.
