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

Domain counts now use uint64 while values, offsets and coefficient arrays remain
uint32. Descriptions select a count width: canonical/ordered versions 1/2 retain
their existing bytes for smaller domains; versions 3/4 add a high count word for
the s=16 extension domain of 2^32 positions. Small-buffer tests cover high-bit
values, affine Horner/inverse checks and 32-level matrix metadata. They do not
establish full s16 discovery, preparation or encoding; see
[count-width validation](../../experiments/count-width-validation/report.md).


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


## Within-block Taylor scheduling

The [CPU Taylor experiment](../../experiments/cpu-taylor/report.md) adds parallel
coefficient gathers when large Taylor stages expose fewer blocks than workers.
The existing recurrence remains the one-worker/small/coarse-parallel reference.
A join separates quotient correction and remainder calculation (inverse reverses
passes); linearized lower-degree bounds make source ranges stable and each
coefficient has one writer. It adds no field-value buffer or ABI requirement.
Correctness suites, independent division/matrix tests and native address/undefined/
thread sanitizers pass. Matched s14 encoding improves from87.9 to61.0 s at32
workers (1.44x), with Taylor35.7 to8.8 s. The hybrid schedule is selected; the
threshold/chunk choices and larger-size behavior remain subject to measurement.
The subsequent [cached s16 encoding](../../experiments/cached-s16-encoding/report.md)
passes in20m52.5s at32 CPU workers with74.4 GiB sampled RSS and complete
systematic/sampled-curve/packed-readback checks. This is one measured full-size
run, not a matched s16 speedup or repeated-latency result.

## Complete-encoding output ownership

`EncodeOnline` permits exact scratch/output identity after the final inverse;
partial overlap and other writable aliases remain rejected. Shared
`ScalingCPPEncoder.EncodeWithStats` transfers that scratch buffer as the caller's
owning result and replaces it on subsequent workspace use. Old words survive
workspace reuse and encoder closure. `Bytes()` counts only currently owned arrays;
Go reference/hybrid output scheduling remains unchanged. The
[CPU memory experiment](../../experiments/cpu-buffer-reuse/report.md) tests retained
outputs, alias boundaries, native sanitizers, memory reduction and runtime.

CPU job73586 now validates this ownership transfer on two full s16 calls:
58.351 GiB peak sampled RSS versus74.370 GiB previously,21m21.3s/20m47.2s
encoding with exact retained CPU/GPU hashes. Small matched timing differences
have both signs; no general time equivalence or penalty is established. Full
results and retained-output/race/sanitizer evidence are linked above.


## Explicit compact CPU workspace

The experimental [bounded CPU output route](../../experiments/cpu-output-memory/report.md)
retains normal owning uint32 results while reducing scratch and assembly memory.
Use separate compact methods; ordinary encoding remains the validated fast path:

```go
cpu, err := plan.WithCPPMatrix(32)
if err != nil { return err }
defer cpu.Close()
workspace, err := cpu.NewCompactWorkspace()
if err != nil { return err } // Requires s<=16 and 2L<=N.
word, stats, err := cpu.EncodeCompactWithStats(message, workspace)
```

`stats` adds compaction/move and backwards expansion timings. Bounded column
batches include gather/scatter in column timing. The workspace detaches its arena
on success, preserving earlier words across calls/closure; `Bytes()` excludes that
caller-owned result. Next use allocates a replacement. Slower execution is possible;
when normal memory fits, this route requires an explicit CLI choice. The experiment
provides `--memory-mode fast|compact` (default fast); production CLI integration is
still planned. s16 principal memory is modeled at39.345 GiB with32 workers, versus
58.351 GiB measured for the fast path. Compact validation now passes:39.409 GiB sampled RAM,21m29.8s full s16 encoding
and retained CPU/GPU hash agreement. Matched s14 compact is3.9–6.1% slower; full
s16 timing comparison is historical. No s17 support, default switch or new
cryptographic claim follows.


## Explicit CPU file-backed constants

[CPU file streaming](../../experiments/cpu-streamed-constants/report.md) retains
fast transforms while loading verified constants into existing work buffers:

```go
fileCPU, err := liftrs.OpenScalingCPPFile(path, trustedSHA256, maxElements, 32)
if err != nil { return err }
defer fileCPU.Close()
word, stats, err := fileCPU.EncodeWithStats(message)
```

Build with liftrs_cpp,cgo, without CUDA. The opened LRSprobe1/2 source must come
from verified preprocessing and remain immutable until Close. Open checks the
whole expected SHA256, shape/field ranges and builds native metadata; byte identity
alone is not mathematical validation. Descriptor identity survives path replacement;
size/mtime checks detect common changes but not hostile concurrent writes. Encoding
rejects read/truncation errors before returning a public word. Prevent external
writes. Requires s<=16,2L<=N; ordinary resident/compact methods stay unchanged.

P initializes prefix data directly. Spectrum tiles borrow dead transpose scratch
between full forward/inverse, then scratch becomes detached owning output. Old
words survive reuse/Close. WorkspaceBytes excludes detached words and file cache;
first encoding lazily allocates two N arrays and later calls replace output scratch.
No resident code/saved H or extra read tile is allocated. FileStats separates
logical reads/range-check time, bytes and chunks; prefix read time overlaps buffer
time. Reads can hit kernel cache, so counts are not physical disk traffic.

Model35.422 GiB principal RAM at s16 versus58.267 GiB resident, but cached code can
add19.422 GiB and packed output cache more. The experiment uses explicit
`--memory-mode file` (default resident), records RSS/cache/cgroup observations and
never drops shared caches. Validation passes:20m49.6s full s16 encoding,35.484 GiB
sampled RSS and retained CPU/GPU hashes. Sampled RSS+tracked file pages reaches
61.820 GiB; charged batch-task cgroup peaks around47.92 GiB, a different accounting
view. This is not a total35 GiB cap. Matched cached s14 costs0.9–1.2% extra encoding
time. No automatic default switch, s17 capability, cryptographic/application or
TRL5 claim follows.


`EncodeOnlineFilePacked` selects an explicit fast file-streaming output variant.
It leaves canonical s-bit little-endian bytes in the initialized prefix of dead
N-word scratch. Two full transform arrays are still needed; no lower peak is
claimed. Atomic byte OR joins neighboring bit fields after zero fill and joined
parity supplies coordinate0. Byte stores preserve format across host endianness.
Ordinary owning/resident entry points stay unchanged. The shared
`ScalingCPPFileEncoder.EncodePackedTo` holds the lock through bounded writer/hash
calls and reuses scratch; metadata/immutable-file and borrowed-pointer contracts
are unchanged. [Validation/cost boundaries](../../experiments/cpu-packed-output/report.md)
remain pending. Source checksums establish identity, not mathematical validity.


`ApplyMatrixFile` is an explicit disk-backed transform component: two distinct
exclusive read/write little-endian files and two bounded RAM tiles replace full
RAM arrays. File cache remains extra; it is not a total RAM bound. Caller offsets,
lengths, field ranges and descriptor access are checked. Joined Taylor file
passes, blocked file transposes and contiguous vectors preserve the resident
arithmetic/global row origins. Private buffers, logical reads/writes/syscalls and
nested timings are reported. No descriptors retained or fsync/checkpoint guarantee;
errors leave partial files, so restore verified input before retrying.
[Progressive validation/trade-offs](../../experiments/cpu-disk-transform/report.md)
pass through2^22 values; full online s10 validation also passes below. s16 needs32GiB backing files and at
least64GiB reads+64GiB writes per transform before Taylor passes, so it is not
admitted under current storage budgets. Resident methods remain default.


`EncodeOnlineDisk` integrates five file-backed transforms with tiled products,
zeroing/code copies and canonical assembly into a caller-owned N-word result.
All code/data/scratch files must be distinct and valid for the call; pointers/
descriptors are borrowed only until return. Code P/spectrum byte offsets may be
unaligned (existing25/33-byte headers). Native total excludes admission;
I/O counters are logical and overlap some phase clocks. Four simultaneously
live RAM tiles and conservative128KiB I/O staging are counted, plus owning output,
input, metadata/runtime and OS cache. Errors leave private output/work files
partial, not committed checkpoints. [Complete validation and trade-offs](../../experiments/cpu-disk-encoding/report.md)
pass small/ASan and complete s10 correctness. Disk2.86/1.14 s first/reused versus
file-fast0.33/0.31 s on NFS; filesystem effects vary. Full s16 storage/run is not
admitted; faster resident/default paths remain unchanged.
