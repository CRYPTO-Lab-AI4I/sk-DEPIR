| Phase | Measured time |
|---|---:|
| load-online | 0.597 s |
| online-setup | 0.004 s |
| encode-0 | 92.118 s |
| verify-0 | 3.191 s |
| pack-output | 1.360 s |
| verify-packed | 0.001 s |

Exact K=50,976,325; N=268,435,456; code rate K/N=0.189901609.
Median profiled encoding: 92.118 s.
Measured discovery/preparation + executor/online setup + one median encoding: 92.122 s (excludes serialization/loading/verification; reused or skipped stages contribute zero).
Whole-process peak RSS: 4.589 GiB.
Packed output: 469,762,048 bytes.
Canonical SHA-256: `b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c`.
