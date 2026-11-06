# Lift-RS kernels in C++

Portable C++17 fixed-constant multiplication and additive-transform kernels for
comparison with [Go transforms](../liftrs-go/README.md). The optional binding
integrates them with the [shared encoder](../liftrs-shared/README.md). It does not
implement cryptographic or application encoding. `transform.hpp` preserves field
and tower coordinates, four-bit tables and recursion. No vendor instructions,
GPU stack or third-party runtime is required.

[bench.cpp](bench.cpp) reads trusted versioned experiment fixtures and measures
batched products or complete transforms. I/O, allocations and inverse validation
are outside timed calls. The [language comparison](../../experiments/kernel-languages/report.md)
records build flags, shared inputs, correctness checks, CPU-profile availability
and measurements. The binary fixture format is internal to that experiment;
large fixtures and matching binaries remain in the ignored local cache.

## Optional Go integration

Build Go callers with `CGO_ENABLED=1`, `-tags liftrs_cpp` and a C++17 compiler
(`CXX=clang++` or `CXX=g++`). The compiler uses generic `-O3` code; there is no
architecture-specific flag. See the [runnable API example](../liftrs-shared/cpp_example_test.go).
`WithCPPTransforms` imports immutable code-only metadata once into a native plan.
Call `Close` to release it; concurrent encodings borrow distinct Go buffers, and
`Close` waits for them. No Go pointer or database buffer is retained by C++.

The [C ABI](c_api.h) performs one call per complete transform, borrowing a
pointer-free `[]uint32` array without interface copies. Go keeps cached generator
spectra, multiplication cutoffs, truncated Karatsuba, unequal blocking and
overlap-add unchanged. Preprocessing and cache loading remain in Go. This tests
the native FFT inside the actual encoder without replacing the remaining Go
arithmetic. The adapter remains explicit in the API; standard Make commands now select it
by default. Go execution remains selectable.

[The full integration comparison](../../experiments/native-encoding/report.md)
reports initialization, native metadata bytes, whole-encoding time and profiles.
`CPPEncodeStats` counts transform calls, elements, native kernel time and the
Go-observed boundary interval. Their difference also includes locks, validation
and timing, so it is not a calibrated cgo-overhead benchmark. Native allocation
is absent from Go heap statistics; process RSS is measured separately.

## Separation and execution

This directory contains native execution and its Go/cgo ownership binding only.
Shared preprocessing, polynomial strategy, artifact loading, API and adapter live
in [liftrs-shared](../liftrs-shared/README.md). The C++ path imports that API and
calls `WithCPPTransforms`, then `Encode` and `Close`:

```sh
CGO_ENABLED=1 CXX=clang++ go test -tags liftrs_cpp \
  ./encoding/liftrs-shared ./encoding/liftrs-cpp
```

The [encoding overview](../README.md) contrasts this with the compiler-free Go
path. No RM code is used by these kernels; it is archived in `../rme_legacy/`.

The [scaling prototypes](../../experiments/scaling-prototypes/report.md) exercise
the original native executor with minimal online state and a reusable shared-code
workspace. That historical checkpoint used unchanged C++ kernels. Sparse
preparation remains Go. Its scalar disk-backed matrix/Taylor probe is a separate
correctness reference for the newer in-memory native matrix path below.


Standard root commands `make test` and `make encode` enable native support;
`make encode` also selects C++ execution. `BACKEND=go` selects the reference path.
No caller API or preprocessing default has changed; enabling the build tag alone
still does not replace explicit adapter construction in arbitrary Go callers.

## Experimental parallel matrix execution

`Plan.ApplyAt` accepts an ordered basis and a packed affine field offset using
internal descriptor version 2. Version 1 canonical plans remain compatible.
`ApplyMatrix` combines sparse h-adic Taylor conversion, arbitrary image-basis
column transforms, affine row transforms and 64×64 blocked transposes. Workers
are explicit; buffers are borrowed for one call, and scratch must be disjoint.
`ApplyMatrixWithStats` reports elapsed native stages, including thread scheduling.
Code-only Taylor multiplication maps are built once in explicit plan preparation.
The binding compiles/links with `-pthread` for portable standard-library threads.
This is an in-memory backend, with an extra N-element transpose buffer, not a
bounded-disk implementation. See the [integration experiment](../../experiments/parallel-encoding/report.md).


The [s=14 first-use experiment](../../experiments/parallel-s14/report.md) extends
the native width guard through s=14 after independent scalar/table, affine Horner
and inverse checks on small buffers. Sparse discovery and paired/root-only
preparation remain Go orchestration, using the native matrix executor for FFTs.
The experiment's code-only executor prepares all needed convolution lengths once
and reuses one transpose buffer; default public discovery remains bounded to s≤13.

## Complete native online encoding

The opt-in `ScalingOnlinePlan.WithCPPMatrix(workers)` encoder now performs all
database-dependent mathematical encoding in C++: buffer preparation, the five
matrix transforms, pointwise products, and canonical output assembly. Reusable
code discovery/preprocessing and artifact loading remain Go. The earlier
`WithCPPTransforms` adapter still replaces transforms only.

Use the [runnable example](../liftrs-shared/scaling_matrix_cpp_example_test.go):
prepare or load a verified `ScalingOnlinePlan`, call `WithCPPMatrix`, allocate one
`NewWorkspace` per concurrent caller, then call `Encode` and finally `Close`.
`EncodeWithStats` adds elapsed counters for transforms, buffer preparation,
products and assembly. `EncodeHybridWithStats` retains the previous Go scalar/
assembly path for comparisons. This does not change default library APIs or
standard command selection of the original transform adapter.

One C ABI call borrows the message, immutable P/spectrum arrays, reusable
workspace and newly allocated canonical output. No large constants are copied
into native storage, and no Go pointer is retained after return. Mutable arrays
must be disjoint (the workspace arena deliberately aliases successive phases).
The same Taylor maps, fixed-constant tables, blocked transposes and explicit
worker scheduling remain in use. Pointwise products and output scatter use
independent chunks; primitive-generator validation ensures disjoint output
positions. Assembly uses a small fixed-omega lookup map and reduces chunk parity.

Run the profiled [native-online experiment](../../experiments/native-online/report.md)
for the cached large-instance comparison. The path is in-memory and mathematical
only: no masks, application partitioning or external-memory backend is implied.

## Timing diagnostics

`ReadClocks()` returns nanosecond readings of the exact C++ steady clock used
by stage counters, system wall time, POSIX monotonic time and all-thread process
CPU time. Compare differences within each clock domain; absolute origins differ.
The readings are sequential and introduce small skew. The
[timing harness](../../experiments/encoding-timing/main.go) brackets whole Go/native
encoding calls with this API so native stages can be checked against a total
from the same clock, while Go/getrusage readings provide independent diagnostics.
This observation API does not change encoding or retained preprocessing.


The [s=15 admission experiment](../../experiments/parallel-s15/report.md) extends
the native importer through s=15 (30-bit extension elements) after small-buffer
Horner/inverse tests. The completed first-use and fresh cached runs establish standalone s=15
execution: about 6.5 minutes per encoding and 18.46 GiB cached peak RSS. Ordinary public Go discovery limits are unchanged.
