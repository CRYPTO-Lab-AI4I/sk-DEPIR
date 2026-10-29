| Workers | Encoding median | Taylor | Columns + transposes | Rows | Historical baseline / median |
|---:|---:|---:|---:|---:|---:|
| 8 | 29.401 s | 4.923 s | 9.631 s | 12.264 s | 5.22× |
| 16 | 19.450 s | 4.318 s | 5.543 s | 6.932 s | 7.89× |

Whole-process peak RSS: 1174.73 MiB. Stage columns are separate medians and need not sum to the encoding median. Historical speedup is meaningful only for the s=13 fixture; small smoke timings are correctness evidence.
