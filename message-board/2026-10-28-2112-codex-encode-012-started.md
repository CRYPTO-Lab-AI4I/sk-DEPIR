# Integrate C++ transforms into the optimized cyclic encoder

- Timestamp: 2026-10-28 21:12 UTC
- Work item: ENCODE-012
- Status: started

The user authorizes thin experimental integration with an apples-to-apples encoding
comparison. Preserve all implemented optimizations: hybrid FFT/Karatsuba products,
low-product pruning, unequal-operand blocking, overlap-add, cached spectrum and
fixed-constant maps. Keep Go preprocessing and scalar/polynomial orchestration;
replace complete transform calls through an opt-in batched C ABI. Proposed but
unmeasured new kernel optimizations are not silently added to just one backend.

Use immutable C++-owned plans and borrowed pointer-free Go uint32 buffers, with
explicit lifetime/Close and a pure-Go default/fallback. Test small oracle outputs,
polynomial paths and concurrent/lifetime behavior, then compare full s=13 codewords
from the retained cache. Profile both paths, preserve cache/profile data, and report
initialization/plan transfer, calls/copies, full encoding and memory separately in
`experiments/native-encoding/`. Default Go builds must not require C++/cgo. Update
API/package/design documentation and README. No native default or preprocessing
rewrite is selected by this experiment.
