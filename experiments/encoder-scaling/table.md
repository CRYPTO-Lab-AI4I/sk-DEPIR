# CPU encoding at s=10,m=2,t=6

Exact N=1,048,576; K=103556. Times are CPU-host wall times; encode is the median of three calls.

| Method | Preparation s | Artifact MiB | Load ms | Encode s | Allocated MiB/call | Preparation / encoding process RSS MiB | Sampled encoding heap increase MiB |
|---|---:|---:|---:|---:|---:|---:|---:|
| cyclic-fft | 26.475 | 5.15 | 53.45 | 2.682 | 10.40 | 74.28 / 55.88 | 10.40 |
| paper | 23.971 | 35.74 | 165.34 | 10.708 | 188.61 | 245.36 / 120.98 | 34.63 |

Heap sampling is a separate instrumented call, includes its 4 MiB output, and can miss peaks. RSS includes the whole stage process.
