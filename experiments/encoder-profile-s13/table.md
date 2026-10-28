# s=13 profiles and retained cache

Profiled times include CPU sampling. Setup and encoding below come from separate runs after a monitor failure; there is no single combined process/RSS measurement.

| Stage | Run | Measured time |
|---|---|---:|
| discovery | setup | 85.109 s |
| preparation | setup | 1784.863 s |
| cache_save | setup | 1.498 s |
| cache_load | setup | 4.864 s |
| cache_load | reuse-01 | 4.834 s |
| encode_0 | reuse-01 | 186.352 s |

Cache: 453,243,425 bytes (432.25 MiB), retained under `.cache/liftrs/s13-m2-t5`.
Reuse-only process peak RSS: 2.257 GiB. Setup process peak RSS was not recovered after the monitor error.

## CPU sample shares

Flat shares are self CPU samples. Cumulative shares include callees and overlap; neither is a fraction of all wall time.

| Phase | Function | Flat | Cumulative |
|---|---|---:|---:|
| setup/preparation | `constantProduct.mul` | 38.46% | 38.51% |
| setup/preparation | `BinaryField.Mul (inline)` | 27.14% | 27.16% |
| setup/preparation | `(*additiveTransform).recurse` | 17.56% | 86.84% |
| setup/preparation | `polyMulLow` | 0.00% | 90.48% |
| setup/preparation | `polyDiv` | 0.00% | 67.54% |
| reuse-01/encode_0 | `constantProduct.mul` | 44.20% | 44.31% |
| reuse-01/encode_0 | `BinaryField.Mul (inline)` | 26.48% | 26.55% |
| reuse-01/encode_0 | `(*additiveTransform).recurse` | 17.51% | 97.67% |
| reuse-01/encode_0 | `polyMulLow` | 0.00% | 40.46% |
