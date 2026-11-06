# s=15 preparation-memory audit

**A 48 GiB commodity machine remains a plausible s=15 target, but the current
preparation path is not yet admitted under its 32 GiB RSS guard.** Its principal
simultaneous array allocation is about 22 GiB; that is not an RSS bound. Hash-map
validation, millions of small factor objects, repeated FFT-buffer allocation and
unmeasured garbage retention justify a memory cleanup/checkpoint before a full
s=15 run. This audit changes no encoding algorithms or support guards.

[model.py](model.py) produces [results.json](results.json) and the
[allocation table](tables.md). It combines current source lifetimes, catalogue
capacity evidence, the retained s=14 first-use measurements and small allocation
replicas. All s=15 values are estimates/scenarios, not measured preparation.

Here N is the encoded length and K the code's message capacity. Support is its
allowed univariate exponent set. Preparation forms the degree-K polynomial P,
the reciprocal-derived generator G (N−K coefficients), and the N-element transform
of G, called its spectrum. These are reusable public code data, not a database.

## Evidence and reproduction

```
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/preparation-memory/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name replicas-final
python3 experiments/preparation-memory/model.py
```

[run.py](run.py) builds [probe.go](probe.go) and retains the binary under
`.cache/liftrs/preparation-memory/<source-id>/`. The
[final replicas](replicas-final/results.json) record commands, source hashes,
revision, environment, live heap, allocation traffic and process RSS, with CPU and
post-GC heap profiles for each case. Exact measured dirty sources are preserved
in `../source-snapshots/`. Go 1.27.3, macOS/arm64, Apple M5 Pro, 48 GiB;
[host reference](../encoding-timing/environment-end.json). Each process has a
1 GiB Go soft target and at most ten million entries; measured peaks stay below
154 MiB. No large code preprocessing or database encoding was run.

The [pilot](replicas/results.json) placed the heap baseline before CPU-profiler
initialization and therefore included profiler buffers in the retained delta.
Its evidence is preserved but excluded from projections. The final source takes
its baseline after profiler initialization and GC. Remaining runtime/background
noise is not subtracted. These are allocator/layout measurements for this Go
version and host, not exact future map capacities or RSS predictions.

| Replica | Entries | Retained heap MiB | Allocation traffic MiB | Peak process MiB |
|---|---:|---:|---:|---:|
| `map[int]bool`, denser | 3,389,383 | 72.56 | 72.94 | 82.08 |
| `map[int]bool`, sparser | 2,542,037 | 72.18 | 72.18 | 81.75 |
| Factor slices | 1,694,947 | 74.02 | 265.26 | 153.86 |
| Membership bitset, 2^24 positions | — | 2.00 | 2.00 | 10.80 |

Different map densities have similar backing capacities, producing about
22.45–29.78 retained bytes per entry. Scaling that empirical range gives an
s=15 validation-map estimate of 4.54–6.02 GiB. A full-domain bitset has an exact
array payload of N/8 = 128 MiB. The bitset is a proposed replacement, not an
implemented support validator.

## What is live, and when

The s=15 catalogue has N=2^30 and estimated K=216,920,545. Its Monte Carlo
interval gives K=214,791,491..219,049,598 (see generated results for authoritative
integer endpoints). Capacity must still be established exactly before encoding.
The coefficient scenario with support-capacity slack changes only from about
22.000 to 22.040 GiB across those endpoints.

1. [Setup](../../encoding/liftrs-shared/scaling_prepare_cpp.go) allocates one
   N-element uint32 transpose scratch array: 4 GiB. Small native transform/Taylor
   metadata remains separate from the payload model.
2. [Discovery](../../encoding/liftrs-shared/orbit.go) builds `good []int`, random
   bytes, a Boolean `found` array and an N-element field grid. Sorted support is
   appended afterward. Validation then creates `map[int]bool`; that map is local
   and is no longer needed when validation returns. Large arrays/old slice
   backing stores can remain in the heap until collection/scavenging.
3. [Root construction](../../encoding/liftrs-shared/scaling_experimental.go)
   first collects all linear/quadratic factors in `[][]uint32`. There can be
   about 108 million leaf objects. All leaves remain reachable through the
   backing header array during the recursive root product, even after individual
   subproducts are complete. A 1.25× header-capacity scenario plus 16-byte leaf
   allocations is 4.65 GiB. The top-root multiply illustration is 13.88 GiB,
   including support, transpose scratch, leaves, polynomial inputs/output and
   its FFT buffers. This illustration is not a bound on every Go stack lifetime.
4. The last [Newton reciprocal](../../encoding/liftrs-shared/orbit_polynomial.go)
   step creates a square with backing length N-1. Truncating its slice view in
   `polyMulLow` does not shrink that backing allocation. The first local FFT
   convolution holds two length-N arrays in addition to the N-element transpose
   scratch, P, the square, result and support: about 21.62 GiB at length-only
   support, or 22.02 GiB with 25% support-capacity slack. If the previous g is
   retained, the separate scenario adds 2 GiB. `g = make(...); copy(g,next)` also
   makes a second result allocation after the convolution, though that is not
   the principal array peak.
5. Full P·G verification has the same 21.62 GiB length-only envelope: 4N bytes
   transpose scratch + 8K support + 4(N+1) P/G + 4N result + 8N FFT pair.
   [Products](../../encoding/liftrs-shared/orbit_product.go) allocate new x/y
   buffers per local convolution. Their dead storage can coexist with later
   allocations until reclaimed. Spectrum preparation follows verification.

These quantities are **not all additive**: the map and factor leaves are not
logically required during final identity verification. Likewise allocation
traffic is not live memory. The source-level scenarios count backing allocations,
including lazily touched pages, and exclude runtime/native metadata, allocator
fragmentation, GC-retained unreachable arrays, stacks and OS overhead.

## Comparison with measured s=14

The retained [s=14 first-use data](../parallel-s14/initial/results.json) measures
13.73 GiB whole-process peak RSS. Its preparation phase allocates 195.55 GiB
cumulatively, ends with 3.52 GiB Go heap and 14.48 GiB Go-managed address-space
accounting. The model's identity array scenario is only 5.38 GiB at s=14.
Multiplying the measured RSS by four gives about 54.9 GiB, but does not establish
an unavoidable live-memory requirement at s=15.

The original [phase records](../parallel-s14/initial/phases.json) are endpoint
heap observations, not in-stage live-heap peaks. The retained
[CPU profile](../parallel-s14/initial/preparation-flat.txt) samples native work,
Go field arithmetic and multiplication; it cannot locate heap high-water marks.
No full-preparation heap profile was captured in that run. Our small heap replicas
help quantify layout costs, but cannot close that missing stage-peak evidence.
Do not attribute the entire process RSS to one particular allocation or stage.

## Admission decision and next changes

Keep s=15 unadmitted for now; there is no evidence yet that HPC or a GPU is
required. On this 48 GiB host, the 22 GiB array scenario is plausible, but a
24 GiB Go soft target leaves little margin and does not impose a 24 GiB RSS cap.
The existing 32 GiB sampled process guard has not been validated for s=15.

Before extending guards and launching the full case:

- Replace map-based support membership with a compact bitset while preserving
  sorted/unique/range and Frobenius-closure checks. Consider the same compact
  representation for discovery's `found` array.
- Build factors in bounded batches and merge completed products with a balanced
  schedule, avoiding a retained list of millions of independent leaves.
- Reuse preparation-only FFT buffers with explicit single-caller ownership.
  Reuse a full-length fresh Newton result when possible, retaining zero-padding
  behavior for short results. Release temporary ownership at stage boundaries.
- Check identical support/P/G/spectrum on small references, then profile a
  targeted s=14 preparation from the retained basis, recording stage live heap,
  allocation traffic and process RSS. Reuse discovery and preserve all caches.
  Use that checkpoint to set the s=15 soft target and hard sampled guard.

The main FFT arrays still need about 22 GiB; these changes primarily reduce
object overhead, allocation traffic and retention risk. Do not promise a specific
RSS reduction before that checkpoint. The unchanged s=15 online array payload is
18.42 GiB; code-only support/P/G/spectrum files are about 8.81 GiB, plus 1.875 GiB
for a packed output. These are analytical payload/storage quantities, not measured
s=15 encoding performance or published artifact sizes.


[check.py](check.py) verifies all replica evidence/source/binary digests, reruns
model arithmetic assertions and checks the probe with `go vet`; results are in
[verification.json](verification.json). Python compilation and local file links
are also checked. Library encoding suites are not repeated because this audit
changes no production code. Raw profiler summaries are kept verbatim; their
narrow `.gitattributes` rule preserves recorded hashes when tools emit trailing
blanks. No mathematical or security claim is established by the layout probes.


Implementation follow-up: the [bounded preparation checkpoint](../preparation-checkpoint/report.md)
implements the compact/batched/reusable-memory path and measures it separately.
The audit above describes the pre-change source revision in its manifest; its
original measurements and scenarios are preserved rather than silently updated.
