# Step 4: local CPU, storage and resource calibration

- Timestamp: 2026-10-28 23:50 UTC
- Work item: SCALE-003
- Status: started

Measure independent native FFT tile throughput at 1/2/4/8/16 workers, file transfer
block sizes with explicit cache policy, and sequential versus overlapped tile I/O.
Use small bounded files/buffers and existing C++ plans rather than rerunning a full
encoding baseline. Inspect local capabilities without choosing an HPC/GPU stack.
Also separate preparation and minimal-loader processes to obtain meaningful RSS
observations, retaining reused preprocessing and stage profiles. Record environment,
source/binary/cache hashes, logical bytes and read/write/sync boundaries; do not
claim independent tile throughput is full-transform or full-encoder parallelism.
Use no-cache file access only when supported and report it distinctly from durable
media bandwidth. Update study/README/report and verify outputs before hardware
recommendations. No default algorithm change or larger full encoding is planned.
