| Workers | Encoding median | Taylor | Columns + transposes | Rows | Historical baseline / median |
|---:|---:|---:|---:|---:|---:|
| 1 | 0.043 s | 0.002 s | 0.013 s | 0.027 s | — |
| 4 | 0.015 s | 0.002 s | 0.004 s | 0.007 s | — |

Whole-process peak RSS: 14.86 MiB. Stage columns are separate medians and need not sum to the encoding median. Historical speedup is meaningful only for the s=13 fixture; small smoke timings are correctness evidence.
