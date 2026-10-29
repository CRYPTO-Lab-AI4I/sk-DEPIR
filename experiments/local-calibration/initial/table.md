| CPU workers | Median batch (s) | Speedup vs 1 |
|---:|---:|---:|
| 1 | 8.305 | 1.00× |
| 2 | 4.247 | 1.96× |
| 4 | 2.220 | 3.74× |
| 8 | 1.208 | 6.87× |
| 16 | 0.669 | 12.42× |

| Transfer block | Read MiB/s | Write MiB/s (excludes fsync) | Median fsync (ms) |
|---:|---:|---:|---:|
| 4096 B | 127.6 | 1916.8 | 7.09 |
| 65536 B | 690.2 | 2521.7 | 8.65 |
| 1048576 B | 4866.9 | 7874.6 | 5.39 |
| 4194304 B | 8089.2 | 11988.9 | 4.31 |

| Pipeline slots | Median read/compute/write (s) | Buffer payload (MiB) |
|---:|---:|---:|
| 1 | 8.341 | 8.0 |
| 2 | 8.320 | 16.0 |
| 4 | 8.332 | 32.0 |
