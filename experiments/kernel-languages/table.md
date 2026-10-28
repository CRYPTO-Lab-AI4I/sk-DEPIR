# Go/C++ kernel and transform comparison

Three repetitions, medians. Kernel times cover 134,217,728 element products per call; FFT times cover one forward transform. Allocation/input I/O/validation are excluded.

| Work | Elements per batch | Go | C++ | C++ speedup |
|---|---:|---:|---:|---:|
| kernel base fixed | 262,144 | 0.3875 s | 0.1564 s | 2.48× |
| kernel base fixed | 4,194,304 | 0.3790 s | 0.1593 s | 2.38× |
| kernel base rotating | 262,144 | 0.3775 s | 0.1655 s | 2.28× |
| kernel base rotating | 4,194,304 | 0.3902 s | 0.1598 s | 2.44× |
| kernel extension fixed | 262,144 | 0.4711 s | 0.2881 s | 1.64× |
| kernel extension fixed | 4,194,304 | 0.4794 s | 0.2884 s | 1.66× |
| kernel extension rotating | 262,144 | 0.4848 s | 0.2855 s | 1.70× |
| kernel extension rotating | 4,194,304 | 0.4819 s | 0.2888 s | 1.67× |
| forward extension fixed | 1,048,576 | 0.6101 s | 0.5326 s | 1.15× |
| forward extension fixed | 67,108,864 | 63.2141 s | 50.1112 s | 1.26× |
