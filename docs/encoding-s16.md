# s=16 representation and memory gate

The count/header extension and small-buffer boundary checks pass. Full experimental
s16 preparation, cached CPU encoding and repeated cached GPU encoding now pass
as recorded below; no
general memory bound, application execution or larger-size support is established. The
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
| Discovery/preparation | Experimental preparation now accepts s16 under the first-use controller; public discovery remains s≤13 | One full preparation/cached encoding configuration passes; public discovery and application scope remain limited |

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
Resident convolution now passes complete smaller-code checks and
[bounded s16 root-product intervals](../experiments/s16-resident-intervals/report.md)
using the saved support. Later operations require up to2^32-point transforms;
the experimental CUDA executor now passes exact operation comparisons through
2^28 points in the [bounded scaling experiment](../experiments/gpu-scaling/report.md).
Private version2 GPU metadata carries2^32 counts and17 row levels without a
full-domain workspace allocation. Explicit per-engine limits enforce each approved job cap (4 GiB for
the larger sweep); full-domain execution still requires a larger reviewed memory schedule
or a validated blocked design before a reliable forecast. Existing CPU/header
count validation remains separate; individual field values still fit32 bits.
No full s16 run is restarted.


The user-authorized [actual GPU preparation attempt](../experiments/gpu-preparation/report.md)
now replaces additional synthetic sweeps. It uses the retained support and a
restartable normalized-root checkpoint, with full identity checking before saving
constants. The report specifies the64 GiB device/256 GiB requested host budget
and independent guards. It performs preparation only; a submitted job is not a
validated full s16 checkpoint until final evidence passes review.


Actual job72875 now completed preparation successfully: full polynomial identity
verified, normalized root and final P/G/spectrum persisted,2h20m15s preparation
and2h24m15s total allocation elapsed. Peak sampled RSS72.575 GiB under the224 GiB
guard. See the linked report for hashes and reconciliation limits. This closes
the actual preparation gate on the measured machine. The later cached encoding
checkpoint below is separate; application/cryptographic execution remains open.


Cached CPU job73131 now passes [the encoding checkpoint](../experiments/cached-s16-encoding/report.md):
20m52.503s encoding,34m53s total job elapsed,32 workers/no GPU and74.370 GiB
sampled RSS. Every systematic position,16 independent sampled curve checks and
1027 packed readbacks pass;8 GiB packed output is retained. This is a single
profiled mathematical encoding with trusted prepared input, not independent
full-reference/repeat equality or application/cryptographic/TRL5 acceptance.


CUDA job73257 passes [full cached GPU encoding](../experiments/cuda-s16-encoding/report.md):
two same-message calls with one session,85.954/60.216 s,7.193 s reusable setup,
67.423 GiB owned device allocations and39.024 GiB sampled host RSS. Both full
canonical hashes and the final tightly packed8 GiB hash match CPU job73131;
all systematic positions, sampled curves and packed readback pass. The allocation
lasted30m02s including independent verification, builds and I/O. CPU comparison
is historical across allocations. This adds GPU repeat/cross-backend evidence;
application/cryptographic/TRL5 acceptance and s17 support remain open.
