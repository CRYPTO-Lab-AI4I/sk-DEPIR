# s=16 representation and memory gate

The count/header extension is implemented and small-buffer boundary checks pass;
full s=16 encoding and its memory bound remain unverified. The
[compatibility validation](../experiments/count-width-validation/report.md)
records current verification. The [handoff](development-handoff.md) requires this gate after the
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
storage, arithmetic intermediates or serialized headers. The implementation
retains four-byte coefficients while separating counts from element values.

## Representation status and remaining admission work

| Boundary | Implementation | Remaining limit |
|---|---|---|
| Go field domain | [orbit_arithmetic.go](../encoding/liftrs-shared/orbit_arithmetic.go) uses uint64 cardinality/products and a wide generator-search counter | Four-byte packed extension values permit s≤16; host lengths must fit |
| Go transforms | Sparse, ordered and dense descriptions derive levels from the wide count; coefficients/offsets remain uint32 | Full s16 arrays are not allocated by boundary tests |
| Native field/import | [transform.hpp](../encoding/liftrs-cpp/transform.hpp) uses a uint64 count; [c_api.cpp](../encoding/liftrs-cpp/c_api.cpp) reads versions 1/2 unchanged and 3/4 with one extra high count word; domain shifts use uint64 | Native field/value bounds reject s≥17 |
| Minimal online file | [scaling_online.go](../encoding/liftrs-shared/scaling_online.go) reads both old 32-bit and new 64-bit counts, validating caps and host byte offsets before allocation | The caller still supplies a trusted digest and an explicit allocation cap |
| Allocation/index boundaries | Counts remain wide during header checks and generator search; lengths use the host int/size_t | Full preprocessing lifetimes and allocation schedules still need memory/admission review |
| Discovery/preparation | Experimental preparation now accepts s16 under the first-use controller; public discovery remains s≤13 | Large feasibility remains unverified; see the guarded experiment's current status |

The existing online C ABI already uses size_t for K/N/buffer lengths, so it is not
necessary to widen every argument. Input/output base-field symbols at s=16
use 16 bits; internal extension arithmetic uses up to 32 bits. s≥17 is a separate
element-width project and must remain rejected by a four-byte extension backend.

Saved widths depend on s. `WriteScalingOnline` uses exact existing `LRSprobe1`
bytes for s≤15 (32-bit K/N), and `LRSprobe2` for s=16 (64-bit K/N). The latter
adds eight header bytes. Native canonical/ordered plans likewise use versions
1/2 for counts fitting 32 bits and 3/4 for the 2^32 domain. Small count variables
use uint64 in memory; widening these scalars does not change the large arrays.

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

Small-buffer tests now cover s16 field arithmetic, canonical/arbitrary-basis/affine
transforms, 32-level matrix metadata, 2^32 counts and malformed headers, including
old-version byte compatibility. The linked cached s15 regression also passes
complete output checks. Review simultaneous preprocessing arrays and baseline profiles, choose worker,
RAM, scratch and wall-time budgets, and enforce them independently if Slurm does
not configure a memory cgroup cap. A larger-RAM in-memory baseline and bounded-RAM
alternatives remain eligible; no s=16 full run is admitted by this document.
The subsequent [first-use audit and controller](../experiments/hpc-s16/report.md)
now define the initial admission: 32 workers, 256 GiB requested RAM, a 224 GiB
sampled RSS guard and a 24-hour wall limit. Small full-pipeline/cache checks pass.
These numerical limits are a bounded experiment policy, not measured requirements
or a guarantee of completion.

The initial full attempt was stopped after discovery; its support is retained.
The [expanded preparation-efficiency step](encoding-scaling.md#current-step-larger-memory-preparation-efficiency)
now separates targeted profiling, specialized/native CPU products, bounded batch
parallelism, root-generation reuse and measured GPU alternatives before full-run
readmission. The current lookup improvement is a smaller-code/kernel result.
