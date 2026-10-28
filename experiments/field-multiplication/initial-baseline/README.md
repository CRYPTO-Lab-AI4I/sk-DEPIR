# Excluded initial control

This first s=10 control overlapped the initial correctness tests, so it is
excluded from performance comparisons. The parent directory contains the fresh
control run without concurrent tests. Its exact pre-change library is revision
`a2e0352`, using the same cyclic-scaling driver and command; see the parent
[report](../report.md) and [reproduction script](../run-baseline.py).
