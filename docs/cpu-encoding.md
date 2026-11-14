# CPU encoding implementation

This guide describes standalone mathematical Lift-RS encoding. Application
partitioning, permutation/masks and storage publication are separate contracts;
see [encoder requirements](encoder-contract.md). The CPU refactor is implemented
and [bounded correctness validation passes](../experiments/cpu-encoding-modules/report.md).
Small s10 encoding medians are 1–3% slower; setup regressions remain under review. Existing
full-size evidence remains through s16; no full s17 preparation/encoding is implied.

## One encoding algorithm, several storage strategies

Let K be the input length, N=2^(2s) the codeword length and
L=nextpow2(2K−1). Verified code data contains P and the transform of G.
The shared native [sequence](../encoding/liftrs-cpp/online.hpp) performs:

1. Place P[0:K] and the message in two zero-padded L-element operands.
2. Transform both operands, multiply corresponding values, then inverse-transform
   to obtain H=(P·message) mod X^K.
3. Retain H's first K coefficients and zero-pad to N positions.
4. Transform, multiply by the stored spectrum of G and inverse-transform.
5. Scatter coefficient i to the canonical coordinate omega^(i+1), and place
   the XOR of the nonzero-coordinate values at coordinate zero.

Every operation finishes its worker threads before the next begins. This is
necessary because the same arena and scratch serve successive phases. The
primitive-generator check establishes distinct scatter destinations; a trusted
file hash establishes byte identity, not the mathematical P·G identity or code
membership. Those require verified preprocessing and independent tests.

The sequence is a C++ template over six concrete operations: prepare, transform,
prefix product, promotion, spectrum product and output assembly. Its callbacks
are compiled inline, with no virtual dispatch or extra domain-sized allocation.
Field arithmetic, constant multiplication, pointwise products and coordinate
scatter are also shared. Storage strategies retain their validated chunk sizes,
checks, padding rules, counter layouts and ownership boundaries.

| Module | Responsibility |
|---|---|
| [arithmetic.hpp](../encoding/liftrs-cpp/arithmetic.hpp), [transform.hpp](../encoding/liftrs-cpp/transform.hpp) | Shared32/64-bit arithmetic, primitive-generator checks, fixed products and recursive transforms |
| [matrix.hpp](../encoding/liftrs-cpp/matrix.hpp) | Joined Taylor conversion, columns/transposes and affine rows; full or bounded scratch |
| [online.hpp](../encoding/liftrs-cpp/online.hpp) | Complete encoding order, pointwise multiplication and scatter/parity reduction |
| [output.hpp](../encoding/liftrs-cpp/output.hpp) | Canonical native words, compact16-bit staging and packed s-bit output |
| [file_io.hpp](../encoding/liftrs-cpp/file_io.hpp) | Bounded little-endian read/write, range checks, EINTR/short-I/O handling and logical counters |
| [online32.hpp](../encoding/liftrs-cpp/online32.hpp) | Resident/compact/streamed-code/packed-output32-bit adapters |
| [wide_online.hpp](../encoding/liftrs-cpp/wide_online.hpp) | uint64 transform workspace with uint32 base inputs/results and source-width spectra |
| [disk_matrix.hpp](../encoding/liftrs-cpp/disk_matrix.hpp), [disk_online.hpp](../encoding/liftrs-cpp/disk_online.hpp) | File-backed transforms and the offset/tile adapter for the common online sequence |
| [progress.hpp](../encoding/liftrs-cpp/progress.hpp) | Optional job-side phase and work counters |

C ABI implementation headers are included once by c_api.cpp or wide_api.cpp.
This keeps their private metadata/error translation in one translation unit and
preserves existing standalone compiler commands. They are not independent public
headers. plan.hpp contains internal immutable32-bit metadata. Existing descriptor
import and the wide direct builder have different formats; their execution uses
shared arithmetic and matrix kernels.

## Width, memory and ownership

Extension values need2s bits: uint32 through s16, uint64 thereafter. Base values,
messages and canonical results remain uint32. Width changes the arithmetic and
buffer types, not the encoding sequence. Existing public Go constructors remain
compatible; the experimental wide constructor is still explicit.

The resident path transfers dead scratch to its owning result. The wide path's
logical4N-byte word retains8N backing bytes. Compact mode narrows into one arena,
scatters into a disjoint half, moves and expands backwards. Packed mode emits
s-bit values directly, retaining its two transform arrays. File mode streams code;
disk mode also puts transforms in two work files, but keeps an owning output in
RAM. These are distinct strategies; their savings cannot simply be added.

Atomic character stores protect compact/wide output that reuses wider objects;
packed bytes use atomic OR because different values may share bytes. Thread
joins precede plain parity completion, compaction moves or return. No typed
access to the former scratch representation may overlap output assembly.

Go owns all borrowed pointer-free arrays and keeps them alive for the native
call. Successful ownership transfer detaches the result backing from a workspace;
subsequent calls allocate replacement scratch/arena. A caller retaining older
results must count their full backing in live RAM. Concurrent calls require
separate workspaces; do not copy live workspaces. Close waits for active calls.

[plan.go](../encoding/liftrs-cpp/plan.go), matrix.go, online.go and disk.go contain
bindings by responsibility. A common fixed-size lock helper acquires each distinct
native plan once, avoiding recursive read locking when prefix/full metadata is
shared. The shared Go adapter constructs the same prefix/full metadata for resident
and immutable-file sources. Closed/uninitialized resident encoders allocate no
new workspace. File descriptors remain caller-owned/exclusive as documented by
their APIs; no fsync, hostile-write protection or resumable checkpoint is implied.

Resident and streamed loaders use the same width-aware header/value/digest
verification. Legacy admission still rejects wide files before domain allocation.
Resident legacy loading replays at most53 header bytes, reads the payload once
and takes ownership of its arrays without copying. File verification reads only
the header twice, then one payload pass, retaining no domain arrays. At s16 every
uint32 spectrum value is canonical, so stream verification can skip its element
scan while still reading/hashing every byte. Existing writing formats stay intact.

The independent Go/dense encoders and hybrid schedule remain correctness references.
Merging them into the native algorithm would weaken independent comparison.
Ordinary Go builds remain CUDA-free; client CPU support is retained. Faster
validated modes remain defaults; slower memory strategies are explicit choices.

## Diagnostics and verification

Set `LIFTRS_CPU_PROGRESS=1` in an experiment job to log flushed JSON to stderr:
UTC timestamp, phase, event, completed/total units and monotonic elapsed time.
Encoding reports operation starts/completions/failures. Native worker blocks,
bounded-column batches and disk Taylor/transpose/vector/product/assembly work
emit advancing counters at most every30seconds per operation. Their units describe
work within that operation, not overall percentage or an ETA. A blocked operation
may have no advancing counter; controller heartbeats are labeled separately.
Disabled logging performs no log I/O or per-work atomic counter updates; small
scope/environment checks remain. Profiled/logged timings are labeled separately.

The [validation runner](../experiments/cpu-encoding-modules/run.py) checks real
small codes through every CPU mode against Go/dense/systematic/packed oracles,
ownership, error recovery, caps/headers, lifetime/races, optimized/native
ASan/UBSan and high-width kernels. It also compares repeated s10 mode timings
with exact pre-refactor sources. This does not establish s17 feasibility or
full-size performance equivalence. Preserve caches and profiles; update the
[scaling reference](encoding-status.md) when new evidence is reconciled.

The first refactor validation stopped at a test-fixture admission error: the
s4,t2 code's padded prefix needs2L>N, which compact/file/disk modes correctly
reject. The corrected all-mode fixture uses s4,t3; the denser code retains
resident/wide correctness and explicit memory-mode rejection tests. Production
layouts/guards are unchanged. The [retry](../experiments/cpu-encoding-modules/report.md)
passes regression, race and sanitizer checks. Matched s10 encoding medians are
1–3% slower, while resident loading and file opening are 44% and 27% slower
(about 2–3 ms each). These setup flags need profiling before claiming equivalent
performance; small-code results do not establish full-size latency.
