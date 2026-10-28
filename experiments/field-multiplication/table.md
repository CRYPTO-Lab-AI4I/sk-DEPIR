# Fixed-constant cyclic encoding comparison

Encoding is the median of three calls; memory is whole-case peak RSS. Baselines for s=11/12 are historical runs of the same driver; s=10 is a fresh control.

| s / t | Setup before → after | Encoding before → after | Encoding speedup | New first use | New peak GiB |
|---|---:|---:|---:|---:|---:|
| 9 / 7 | 2.35 s | 0.249 s | — | 0.05 min | 0.021 |
| 10 / 6 | 27.45 → 12.70 s | 2.77 → 1.22 s | 2.28× | 0.25 min | 0.071 |
| 11 / 6 | 145.26 → 64.36 s | 14.60 → 6.15 s | 2.38× | 1.30 min | 0.251 |
| 12 / 6 | 677.19 → 307.59 s | 67.76 → 29.06 s | 2.33× | 5.90 min | 0.923 |
