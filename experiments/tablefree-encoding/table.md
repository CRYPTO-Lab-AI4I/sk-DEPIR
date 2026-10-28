# Table-free encoding checkpoint

Times measured on CPU; median of three encoding calls. All preparation is code-only and excluded from encoding.

| s / t | N / exact K | Method | Discovery ms | Preparation ms | Artifact KiB | Load ms | Encode ms | Preparation / encoding-process RSS MiB |
|---|---:|---|---:|---:|---:|---:|---:|---:|
| 8 / 8 | 65,536 / 3,121 | paper | 69.57 | 322.68 | 806.2 | 5.09 | 195.34 | 14.89 / 12.44 |
| 8 / 8 | 65,536 / 3,121 | cyclic-fft | 69.57 | 388.64 | 284.2 | 3.93 | 72.56 | 14.17 / 11.92 |
| 9 / 7 | 262,144 / 17,200 | paper | 263.76 | 8162.43 | 5281.3 | 27.41 | 4114.23 | 48.66 / 26.28 |
| 9 / 7 | 262,144 / 17,200 | cyclic-fft | 263.76 | 3918.22 | 1204.1 | 14.51 | 508.10 | 34.00 / 17.98 |
