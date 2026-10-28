# Lift-RS kernels in C++

Portable C++17 fixed-constant multiplication and additive-transform kernels for
comparison with [liftrs-go](../liftrs-go/README.md). This is not a full codeword,
cryptographic or application encoder. `transform.hpp` follows the same field,
tower coordinates, four-bit tables and recursion as Go. No vendor instructions,
GPU stack or third-party runtime is required.

[bench.cpp](bench.cpp) reads trusted versioned experiment fixtures and measures
batched products or complete transforms. I/O, allocations and inverse validation
are outside timed calls. The [language comparison](../../experiments/kernel-languages/report.md)
records build flags, shared inputs, correctness checks, CPU-profile availability
and measurements. The binary fixture format is internal to that experiment;
large fixtures and matching binaries remain in the ignored local cache.
