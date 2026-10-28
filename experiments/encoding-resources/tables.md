| s | K evidence | Current arrays resident (GiB) | Current principal-buffer scenario (GiB) | Lean scenario (GiB) | Online packed cache (GiB) | Packed codeword (GiB) |
|---:|---|---:|---:|---:|---:|---:|
| 13 | verified | 0.59 | 1.18 | 0.88 | 0.22 | 0.10 |
| 14 | estimated | 2.38 | 4.76 | 3.57 | 0.96 | 0.44 |
| 15 | estimated | 9.62 | 19.23 | 14.42 | 4.13 | 1.88 |
| 16 | estimated | 38.85 | 77.70 | 58.28 | 17.71 | 8.00 |
| 17 | estimated | 220.84 | 441.68 | 363.26 | 75.66 | 34.00 |
| 18 | estimated | 889.44 | 1,778.87 | 1,462.15 | 322.15 | 144.00 |
| 19 | estimated | 3,581.11 | 7,162.22 | 5,883.67 | 1,367.14 | 608.00 |
| 20 | estimated | 14,419.96 | 32,935.92 | 27,773.94 | 5,786.24 | 2,560.00 |

Array scenarios exclude JSON, map validation, capacities, garbage retention, allocator and OS overhead; they are not RSS bounds. s>=14 is unimplemented; s>=16 needs size-accounting changes, s>=17 wider extension elements.

| s | Dense transform-setup pair (GiB) | Product-tree coefficients + node headers (GiB) | Full packed cache (GiB) | Transform I/O at p=4 (TiB) |
|---:|---:|---:|---:|---:|
| 13 | 0.38 | 2.78 | 0.34 | 0.007 |
| 14 | 1.50 | 12.34 | 1.48 | 0.027 |
| 15 | 6.00 | 53.95 | 6.38 | 0.109 |
| 16 | 24.00 | 235.76 | 27.43 | 0.438 |
| 17 | 192.00 | 1,525.40 | 117.32 | 3.500 |
| 18 | 768.00 | 6,672.42 | 500.31 | 14.000 |
| 19 | 3,072.00 | 29,016.51 | 2,126.29 | 56.000 |
| 20 | 12,288.00 | 125,953.55 | 9,012.47 | 320.000 |
