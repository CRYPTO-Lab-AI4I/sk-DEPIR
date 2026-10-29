| Phase | Measured time |
|---|---:|
| executor-setup | 0.000 s |
| discovery | 0.003 s |
| save-basis | 0.011 s |
| preparation | 0.006 s |
| save-prepared | 0.011 s |
| load-online | 0.000 s |
| online-setup | 0.000 s |
| encode-0 | 0.007 s |
| verify-0 | 0.001 s |
| pack-output | 0.004 s |

Exact K=589; N=4,096; code rate K/N=0.143798828.
Median profiled encoding: 0.007 s.
Measured discovery/preparation + executor/online setup + one median encoding: 0.017 s (excludes serialization/loading/verification; reused or skipped stages contribute zero).
Whole-process peak RSS: 0.014 GiB.
Packed output: 3,072 bytes.
Canonical SHA-256: `97a47fcfcc6d8c558ce131d41be267a82db051305d96ddf7eded267c19cf9163`.
