| Phase | Measured time |
|---|---:|
| executor-setup | 0.029 s |
| discovery | 254.898 s |
| save-basis | 0.157 s |
| preparation | 2017.985 s |
| save-prepared | 1.208 s |
| release-preparation | 0.051 s |
| load-online | 0.687 s |
| online-setup | 0.081 s |
| encode-0 | 90.521 s |
| verify-0 | 3.210 s |
| encode-1 | 90.533 s |
| verify-1 | 3.209 s |
| encode-2 | 91.987 s |
| verify-2 | 3.292 s |
| pack-output | 1.359 s |
| verify-packed | 0.001 s |

Exact K=50,976,325; N=268,435,456; code rate K/N=0.189901609.
Median profiled encoding: 90.533 s.
Measured discovery/preparation + executor/online setup + one median encoding: 2363.525 s (excludes serialization/loading/verification; reused or skipped stages contribute zero).
Whole-process peak RSS: 13.726 GiB.
Packed output: 469,762,048 bytes.
Canonical SHA-256: `b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c`.
