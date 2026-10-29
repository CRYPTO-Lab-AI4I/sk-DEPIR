| CPU workers | Median batch (s) | Speedup vs 1 |
|---:|---:|---:|
| 1 | 0.000 | 1.00× |
| 2 | 0.000 | 1.99× |
| 4 | 0.000 | 4.45× |

| Transfer block | Read MiB/s | Write MiB/s (excludes fsync) | Median fsync (ms) |
|---:|---:|---:|---:|
| 4096 B | 49.8 | 111.6 | 11.32 |
| 65536 B | 99.3 | 1020.9 | 4.52 |
| 1048576 B | 402.8 | 4416.6 | 5.21 |
| 4194304 B | 274.1 | 4577.5 | 4.20 |

| Pipeline slots | Median read/compute/write (s) | Buffer payload (MiB) |
|---:|---:|---:|
| 1 | 0.000 | 0.0 |
| 2 | 0.001 | 0.0 |
| 4 | 0.000 | 0.0 |
