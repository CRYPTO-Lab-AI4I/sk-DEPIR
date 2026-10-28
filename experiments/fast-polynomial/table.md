# Fast polynomial multiplication: encoder comparison

Same-machine runs; median of three per-database encodings. Reusable preparation is separate.

| s / t | Method | Preparation old → new (ms) | Encode old → new (ms) | Encode speedup | Allocated per encode old → new (MiB) | Encode-process RSS old → new (MiB) |
|---|---|---:|---:|---:|---:|---:|
| 8 / 8 | paper | 344.15 → 259.56 | 194.98 → 147.44 | 1.32× | 10.37 → 4.79 | 12.80 → 11.84 |
| 8 / 8 | cyclic-fft | 385.53 → 442.82 | 71.36 → 70.90 | 1.01× | 2.57 → 1.65 | 10.98 → 11.50 |
| 9 / 7 | paper | 8522.68 → 3311.10 | 4143.95 → 1544.79 | 2.68× | 279.75 → 33.97 | 26.20 → 25.05 |
| 9 / 7 | cyclic-fft | 3904.11 → 3409.64 | 490.45 → 430.43 | 1.14× | 59.17 → 21.35 | 17.86 → 16.84 |
