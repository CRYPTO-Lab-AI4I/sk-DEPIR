| Workers | Encoding median | Taylor | Columns + transposes | Rows | Historical baseline / median |
|---:|---:|---:|---:|---:|---:|
| 8 | 29.795 s | 4.974 s | 9.699 s | 12.477 s | 5.15× |
| 16 | 20.152 s | 4.380 s | 5.833 s | 7.308 s | 7.61× |

Whole-process peak RSS: 1175.34 MiB. Stage columns are separate medians and need not sum to the encoding median. Historical speedup is meaningful only for the s=13 fixture; small smoke timings are correctness evidence.
