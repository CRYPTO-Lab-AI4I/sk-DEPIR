# s=16 representation and memory gate

This is a source-based readiness audit, not implemented s=16 support or a measured
memory bound. The [handoff](development-handoff.md) requires this gate after the
verified s=15 checkpoint. The [Linux baseline](../experiments/hpc-cached-s15/report.md)
reproduces the existing encoder before changing its representations.

## Why the parameter is currently bounded

For m=2, the codeword and quadratic-extension domain have N=2^(2s) positions.
An individual packed extension element is a coordinate in `[0,N)`; the domain's
cardinality is N. They need different integer ranges at s=16:

| s | N | Largest packed extension value | Fits uint32? |
|---|---:|---:|---|
| 15 | 1,073,741,824 | 1,073,741,823 | Both count and value |
| 16 | 4,294,967,296 | 4,294,967,295 | Value only |
| 17 | 17,179,869,184 | 17,179,869,183 | Neither |

`s` is an input argument within implemented bounds. Increasing it does not widen
storage, arithmetic intermediates or serialized headers. The s=16 change should
retain four-byte coefficients while separating counts from element values.

## Changes required before enabling s=16

| Boundary | Current evidence | Required work |
|---|---|---|
| Go field domain | [orbit_arithmetic.go](../encoding/liftrs-shared/orbit_arithmetic.go) stores `orbitField.q` as uint32 and computes `q*q` before converting to int | Widen cardinality and products before multiplication; audit generator search, inverses, bounds and conversions |
| Go transforms | [scaling_experimental.go](../encoding/liftrs-shared/scaling_experimental.go), [scaling_matrix_cpp.go](../encoding/liftrs-shared/scaling_matrix_cpp.go) and [scaling_prepare_cpp.go](../encoding/liftrs-shared/scaling_prepare_cpp.go) infer levels using `bits.Len32(q)` | Use a count representation supporting 2^32; keep field offsets and coefficients uint32; audit degree/shift expressions and allocation products |
| Native field/import | [transform.hpp](../encoding/liftrs-cpp/transform.hpp) stores field cardinality as uint32; [c_api.cpp](../encoding/liftrs-cpp/c_api.cpp) imports one 32-bit cardinality word and checks `1u<<width` | Version the native description, widen count fields and use defined shifts at width 32; preserve versions 1/2 and existing fixtures |
| Minimal online file | [scaling_online.go](../encoding/liftrs-shared/scaling_online.go) reads four uint32 header words and bounds s≤15 | Add a distinguishable version with wider count fields; keep `LRSprobe1` interpretation and digest checks intact; validate before large allocations |
| Allocation/index boundaries | Go slices use int and native buffers use size_t, but transforms, byte products, chunk bounds and exponent loops cross narrower representations | Require a 64-bit host; check count-to-length/offset conversions and multiplication overflow; distinguish maximum index from element count |
| Discovery/preparation | Experimental preparation explicitly accepts s≤15; public discovery remains s≤13 | Keep public bounds unchanged until separately justified; enable an experimental s=16 path only after arithmetic, header and shape tests pass |

The existing online C ABI already uses size_t for K/N/buffer lengths, so it is not
necessary to widen every argument. Native field cardinality, descriptors and
derived shift logic still need changes. Input/output base-field symbols at s=16
use 16 bits; internal extension arithmetic uses up to 32 bits. s≥17 is a separate
element-width project and must remain rejected by a four-byte extension backend.

## Memory and admission

The current workspace aliases its two prefix-convolution arrays with the full
convolution buffer. With L=nextpow2(2K−1) and the s=16 catalogue estimate
K=919,440,829, 2L=N. Simultaneous principal online payload is:

- Immutable P and generator spectrum: `4(K+1)+4N` bytes.
- Message and saved prefix H: `8K` bytes.
- Shared prefix/full convolution arena, transpose scratch and canonical output:
  `12N` bytes.

Total is approximately `16N+12K` bytes, **74.3 GiB**. This is an analytical
payload scenario, not measured RSS, exact K, a preprocessing requirement or a
safe allocation budget. Code discovery, support/factor construction, reciprocals,
identity validation, GC retention and temporary copies have separate lifetimes;
the approximately 4× domain growth from s=15 does not establish their RSS.

First validate s=16 field arithmetic and arbitrary-basis/affine transforms using
small buffers in the 32-bit extension domain. Test the 2^32 count and malformed
headers without allocating a full codeword, including old-version compatibility.
Then compare changed s≤15 codewords and retained constant hashes before a large
run. Review simultaneous preprocessing arrays and baseline profiles, choose worker,
RAM, scratch and wall-time budgets, and enforce them independently if Slurm does
not configure a memory cgroup cap. A larger-RAM in-memory baseline and bounded-RAM
alternatives remain eligible; no s=16 full run is admitted by this document.
