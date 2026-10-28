# Fast polynomial multiplication: encoder comparison

Same-machine runs; median of three per-database encodings. Reusable preparation is separate.

| s / t | Method | Preparation old → new (ms) | Encode old → new (ms) | Encode speedup | Allocated per encode old → new (MiB) | Encode-process RSS old → new (MiB) |
|---|---|---:|---:|---:|---:|---:|
| 8 / 8 | paper | 333.60 → 244.14 | 192.98 → 140.71 | 1.37× | 10.37 → 4.79 | 12.36 → 12.27 |
| 8 / 8 | cyclic-fft | 386.00 → 692.35 | 71.42 → 69.66 | 1.03× | 2.58 → 1.65 | 11.55 → 11.23 |
| 9 / 7 | paper | 8144.18 → 3211.46 | 4113.10 → 1493.13 | 2.75× | 279.74 → 33.98 | 26.28 → 26.34 |
| 9 / 7 | cyclic-fft | 3903.37 → 4682.83 | 490.58 → 424.58 | 1.16× | 59.17 → 21.36 | 18.58 → 17.19 |
