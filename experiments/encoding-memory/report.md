# Encoding memory audit and s17 choices

Status: allocation-free audit/comparison complete; no memory–time design selected.
The memory-first plan was published as24bbad9. This work covers its accounting
and alternatives steps, not implementation or admission of s17.

## Reproduce and evidence

Run from the repository root, using Python3 standard library only:

```bash
python3 -B experiments/encoding-memory/run.py --check
python3 -B experiments/encoding-memory/run.py
```

[Script](run.py), [results](results.json). Execution takes under a second and
allocates no database, transform arrays or GPU memory. The result records
revision24bbad9d06775f16cfd512ef7f134219408873e9 and SHA256 for the script, inspected
sources, catalogue and measured reference manifests/phases/checkpoints. No new
performance or peak-memory measurements are generated. The script checks shape,
width/aliasing arithmetic, packed size and reconciliation against saved s16
CPU/GPU measurements. It rejects a K incompatible with the modeled prefix shape.

Measured inputs: [CPU job73131](../cached-s16-encoding/report.md),
[GPU job73257](../cuda-s16-encoding/report.md), and
[preparation job72875](../gpu-preparation/report.md). The CPU/GPU jobs used
Xeon6710E on gnode17; the GPU was H200 NVL143,771 MiB (140.40 GiB), driver580.126.20.
Saved logs describe GCC14.2.0, CUDA12.9.1 and Go1.27.3. Profiles/caches remain
retained at the locations in those reports. Their encoded-file hashes were
already reconciled; this audit reads small manifests and does not rehash huge
files. The reported node had about1008 GiB physical RAM, but that is not an
allocation or authorization to use it: the GPU job requested128 GiB and capped
sampled RSS at96 GiB. Existing96 GiB explicit device cap is unchanged.

## Accounting conventions and baseline reconciliation

N=2^(2s); L=nextpow2(2K-1). Both configurations here satisfy2L=N. K is exact
918,672,421 at s16, but the catalogue's3,870,933,765 at s17 is only an estimate.
Base coefficients/message/output use b=4-byte slots. Internal extension values
use e=4 at s16 and are modeled with e=8 at s17 because34 bits no longer fit32.
No current arithmetic/loader/kernel implementation accepts s17.

Let D=eN, P=b(K+1), M=bK and W=bN. Count large logical arrays, not RSS bounds:

- Online plan: R=P+D (P and the generator spectrum).
- Native CPU: R+2M+2D+W (input, saved H, aliased convolution arena, transpose
  scratch and separate canonical output).
- Current GPU: R+3D (two arithmetic arrays and transpose scratch, all full-sized).
- Current GPU host harness: R+M+W; it keeps the host plan for validation.
- Finished prepared data with support: b(N+1)+D+8K, before scratch/other live values.

| Quantity | s16 model, GiB | s16 measured, GiB | s17 model, GiB |
|---|---:|---:|---:|
| Native CPU online principal arrays / sampled RSS | 74.267 | 74.370 | 491.261 |
| GPU device principal arrays / owned allocations | 67.422 | 67.423 | 526.420 |
| GPU host principal arrays / sampled RSS | 38.844 | 39.024 | 220.841 |
| Minimal online file payload | 19.422 | 19.422 | 142.420 |
| Canonical uint32 output array | 16 | 16 | 64 |
| Final tightly packed output | 8 | 8 | 34 |

GPU measured minus modeled arrays is810,256 bytes of tables/plans; CUDA context,
pageable-copy staging and driver memory are outside OwnedBytes. RSS includes
runtime/allocator effects. s16 reconciliation supports this inventory, not a
universal overhead percentage. The previous rough530–550 GiB GPU estimate is
refined to526.420 GiB principal arrays if base coefficients stay32-bit; widening
those too would unnecessarily add14.420 GiB. Extra overhead still needs measurement.

## Lifetimes and independent preparation gate

| Phase | Live large values / issue | Memory action to evaluate |
|---|---|---|
| Discovery | `good` exponents, random bytes, found bitset, full interpolation grid, executor scratch, eventual support | Stream enumeration/random generation where compatible; bound grid/transform schedule separately. Transferred s16 support bypassed discovery; s17 has none. |
| Support load/validation | K host `int` indices and N-bit membership set | Current64-bit host indices are needed; s17 support serialization must become64-bit instead of the harness's uint32 format. |
| Root product | Support, retained sibling products, batched factors and product executor | Existing builder drops tree children; generic transforms still need two arrays. Preserve root restart checkpoint. |
| Reciprocal / identity | P, evolving G, square/next result, product output, generic FFT scratch, possible old allocations | Stage lifetimes matter; identity verification must remain. No preprocessing RSS bound follows from final P/G size. |
| Spectrum / save | P, G, support, N spectrum, possibly identity result; save loops use small blocks | Separate result ownership from actual compiler/GC liveness and device executor retention. |
| Release / reload | Harness drops preparation state, closes executor, frees OS memory, reloads only P/spectrum | Reuse this separation. `ScalingPlan()` copies P/spectrum; the full harness writes borrowed data then releases it instead. Avoid a second full host plan. |
| GPU setup | Host P/spectrum plus device constants/workspace | CUDA object does not retain host plan; caller can release it if no CPU/verification consumer needs it. |
| Prefix / full convolution | Prefix needs3L slots including transform scratch; full transform needs data+scratch, plus stored spectrum | Share two full work arrays across phases, instead of three. |
| Canonical assembly | Transformed result immutable, previous transpose scratch dead | Scatter canonical word into dead scratch, with a separate small parity/error buffer. Preserve owning host return value. |
| Packing / verification | Current word+message+plan retained; writing/hashing uses bounded blocks; random verification reads word | Streamed output needs an explicit sink/readback design. Do not silently change the canonical-array API or omit checks. |

The loader is already block-decoded; mapping it alone would not reduce device
workspace. Releasing the GPU caller's host plan would lower modeled host peak
from38.844 to19.422 GiB at s16, and220.841 to142.420 GiB at s17: the plan must
still exist during verified loading/setup. The validation harness intentionally
retains it, and a CPU consumer may also need it. This saves host retention without
extra transforms, not disk/PCIe traffic; actual RSS reclamation must be measured.

At s17, a full extension array alone is128 GiB. Discovery's grid and a full
scratch can therefore exceed256 GiB before exponent/random lists. Support is
28.841 GiB in memory and a proposed64-bit file; its membership bitset is2 GiB.
Final P/G/spectrum/support alone total220.841 GiB. Identity results add64 GiB
while live; a generic pair of full transform arrays adds256 GiB. These values
are inventories, not a simultaneous whole-preprocessing peak or proposed budget.

s16 preparation peaks at72.575 GiB sampled RSS; retained checkpoint live heap
maximizes at54.909 GiB and is54.909 GiB at `complete`, whereas final logical
P/G/spectrum/support is38.844 GiB. This gap is evidence against treating final
arrays as the peak. Heap allocation attribution and transient/device lifetimes
need further profiling before a full s17 preparation budget can be proposed.
Do not simply multiply s16 preprocessing RSS by eight: many base arrays stay32-bit.

## Alternatives and their costs

All figures below are proposed principal array payloads, not implemented capacities
or approved resource increases. Small tables, parity buffers, context, allocator
slack and overlap during replacement allocations require additional headroom.

| Route | s16 device GiB | s17 device GiB | Extra work / memory–time trade-off |
|---|---:|---:|---|
| Current resident layout | 67.422 | 526.420 | Baseline; cannot fit s17. |
| A: reuse two work arrays, keep spectrum | 51.422 | 398.420 | Same five transforms; no added host/device transfers. Removes one full array; replaces final output destination with dead scratch. Correctness/latency unmeasured. |
| A plus stream code;2 GiB tile | about34 | about258 | Reupload P/spectrum for each reused message; cannot fit s17 yet. Lower device use moves constants to host/disk. |
| B: one full in-place work array, stream code;2 GiB tile | about18 | about130 | Requires new in-place transpose/layout schedule, not just tile kernels. Can change locality/latency; code upload repeats. Tight headroom on H200 and exceeds current96 GiB cap; needs separate admission. |
| Compact34-bit storage alone, current four-array layout | — | 286.420 | Packing/unpacking and cross-word arithmetic cost; still cannot fit. Does not by itself solve capacity. |
| C: host/disk-backed transforms through bounded GPU tiles | configurable | configurable | Can cap device memory well below96 GiB, but adds repeated data movement and possibly recomputation. Host RAM or disk backs full arrays. Whole transform/Taylor dependencies and routing remain to be implemented. |

A's proposed schedule: put both L-sized prefix operands in one N-sized arena and
use the other N-sized buffer for transpose scratch. After the prefix inverse,
retain its first K coefficients in that arena and clear only the tail; there is
no need for a separate saved H on GPU. Run full forward/product/inverse in that
arena, then write canonical positions into the dead transpose buffer. Keep parity
separate and download into a caller-owned host array. This preserves transform
counts and removes an existing prefix-H copy; it is not an in-place permutation
of a live input. No claim of equal runtime is made until measured.

B would need full in-place transforms, including rectangular prefix layouts,
not the existing out-of-place transpose. It saves the second full work buffer,
but must also avoid retaining P on device at the prefix peak. Code streaming adds
142.420 GiB per reused s17 message, beyond14.420 GiB message upload and64 GiB output
download. The first code upload already exists in resident setup; the added cost
is repetition/lost amortization, not an extra first-use upload by definition.
Host backing can still require the142.420 GiB verified plan plus message/output
(and pinned staging); device savings do not remove host requirements. Streaming
from disk substitutes file reads and can lower host RAM at additional I/O cost.

For C, one explicit disk-backed inventory is two128 GiB work files, a64 GiB saved
prefix operand,64 GiB canonical routing file,142.420 GiB online code and34 GiB
packed output:560.420 GiB total, of which384 GiB is temporary work/routing. This
excludes retained support/G, old checkpoints, new validation copies and atomic
replacement overlap. It is one conservative schedule example, not a minimum
storage bound. Keeping the work files in RAM instead needs at least256 GiB host
workspace before constants/input/output. Disk-backed arrays/mmap are not an RSS
cap: page-cache/cgroup accounting and eviction must be guarded.

### Transfer sensitivity, not runtime forecasts

To expose C's potential penalty, charge p complete read/write passes per transform
across three L-sized and two N-sized transforms. This illustrative schedule moves
2pe(3L+2N)=896p GiB at s17. It excludes pointwise/code/output routing and does not
prove that p=1 is achievable. Taylor dependencies may require many passes/halos.

| Passes p | Additional staging traffic | Copy time using saved s16 directional rates | Disk time if aggregate throughput were1 GiB/s |
|---|---:|---:|---:|
| 1 | 896 GiB | 90 s | 14m56s |
| 4 | 3.5 TiB | 6m01s | 59m44s |
| 16 | 14 TiB | 24m03s | 3h58m56s |

The measured s16 warmed upload/download rates were7.736/13.887 GiB/s for different
sized pageable transfers. Applying them here is a sensitivity calculation, not
measured streamed bandwidth. Disk1 GiB/s is an explicit hypothetical, not measured
storage performance. Compute, synchronization and routing are additional;
overlap can reduce additive costs only if measured. The prior8–20-minute resident
s17 GPU projection cannot be reused as a streamed runtime prediction.

## Recommendation and decision gate

First validate A and host-plan release on s16: meaningful memory savings without
adding transforms, traffic or code recomputation. It is an incremental step,
not s17 feasibility. Keep the current layout available for matched comparisons
and test all saved hashes/lifecycle checks. A bounded prototype needs its own
experiment admission; no job is submitted by this audit.

For actually fitting s17 on one available GPU, choose between B's more complex
in-place schedule/tight device headroom and C's greater host/disk traffic. Neither
is selected here. Present that choice before adopting either penalty or raising
the96 GiB device/128 GiB requested host budget. CPU remains a maintained path;
a streaming GPU backend must not become a CUDA prerequisite. Application,
cryptographic encoding and TRL5 scope are unchanged.
