# Fast polynomial multiplication and truncated products

- Timestamp: 2026-10-28 11:51 UTC
- Work item: ENCODE-004
- Status: started

Implement additive-FFT polynomial multiplication on variable-sized binary
subspaces, retaining Karatsuba for small products, plus a low-coefficient product
for prefixes and Newton/remainder computations. Reuse immutable transform data
from code preprocessing; allocate message-dependent scratch per call. Validate
against independent convolution, existing division/encoding tests, and committed
artifacts/output hashes. Benchmark kernels and both full encoder pipelines in a
new reproducible experiment; distinguish reusable setup from database encoding.
Check and update README, encoder documentation and experiment index. Preserve
prior results and the s<=9 encoder boundary. No CPU-specific or GPU kernel is
part of this change.
