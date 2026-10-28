# Prepared encoder comparison

Measured CPU times; same codeword for both methods. Median of three encodings, with preparation excluded. P=paper, C=cyclic Karatsuba, F=cyclic cached FFT.

| s / t | N | K | Discovery ms | Prepare P / C / F ms | Artifact P / C / F KiB | Load P / C / F ms | Encode P / C / F ms | Peak RSS P / C / F MiB |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| 4 / 2 | 256 | 97 | 0.11 | 0.07 / 0.06 / 0.06 | 16.6 / 0.9 / 1.6 | 0.24 / 0.11 / 0.12 | 0.05 / 0.03 / 0.03 | 6.62 / 6.30 / 6.30 |
| 5 / 3 | 1,024 | 196 | 0.15 | 0.23 / 0.41 / 0.43 | 35.8 / 2.8 / 5.2 | 0.36 / 0.14 / 0.16 | 0.16 / 0.16 / 0.10 | 6.86 / 7.06 / 6.59 |
| 6 / 4 | 4,096 | 589 | 0.50 | 1.60 / 3.29 / 3.51 | 123.6 / 10.6 / 20.1 | 0.93 / 0.34 / 0.36 | 0.98 / 1.61 / 0.48 | 10.39 / 10.47 / 7.22 |
| 7 / 6 | 16,384 | 1,268 | 3.80 | 5.54 / 29.40 / 29.79 | 295.0 / 38.1 / 74.1 | 2.12 / 0.98 / 1.30 | 3.65 / 12.18 / 1.98 | 12.16 / 12.16 / 9.27 |
| 8 / 8 | 65,536 | 3,121 | 19.21 | 57.39 / 253.95 / 287.54 | 806.2 / 144.6 / 284.2 | 5.31 / 2.71 / 4.18 | 12.91 / 93.60 / 8.90 | 13.22 / 14.70 / 11.28 |

RSS is the complete method process peak, including loading and I/O; it is not encoding-only live heap. See results.json for allocations, artifact reads, sync writes and cached readbacks.
