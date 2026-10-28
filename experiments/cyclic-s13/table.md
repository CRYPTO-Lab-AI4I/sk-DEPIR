# Cyclic encoding at s=13

Status: completed. Times below are measured, not forecasts.

| Stage | Time | Sampled peak heap | Sampled peak Go runtime |
|---|---:|---:|---:|
| discovery | 85.95 s | 0.834 GiB | 0.984 GiB |
| preparation | 1765.37 s | 4.711 GiB | 5.342 GiB |
| encode_0 | 185.15 s | 1.181 GiB | 5.342 GiB |
| encode_1 | 185.01 s | 1.181 GiB | 5.342 GiB |
| encode_2 | 185.20 s | 1.181 GiB | 5.342 GiB |

Exact K=11,926,435, N=67,108,864; code rate=0.17771773, N/K=5.6269.
First use: 33.94 min. Median encoding: 185.15 s.
Whole-case wall time: 40.15 min; process peak RSS: 5.337 GiB.
Packed output: 109,051,904 bytes (104 MiB).
