# Streamed constants and s17 representation audit

Status: allocation-free audit/design complete; no streamed encoder or new job.
The user requested publication of the validated compact work, then this next
scaling gate. Checkpoint **f9581d3** was committed and pushed before the audit.
[Script](run.py) and [results](results.json) record revision, source/reference
hashes, parameters, byte counts and numerical checks. Run from the repository root:

```bash
python3 -B experiments/streamed-constants/run.py
```

This reads source, catalogue and small completed-job manifests. It allocates no
field-domain arrays, touches no large cache, requests no GPU and performs no
hardware benchmark. Python3 standard library suffices. No profiles are produced
because there is no encoding/kernel execution to profile; existing profiles and
preprocessing remain retained. Byte counts below are projections unless explicitly
labeled measured; they exclude allocator/driver/pinned/page-cache overhead.
The script also emits hypothetical bandwidth scenarios, not latency predictions.

## Evidence and proposed s16 execution

The completed [GPU reuse](../cuda-buffer-reuse/report.md),
[compact GPU](../gpu-output-memory/report.md) and
[compact CPU](../cpu-output-memory/report.md) runs establish correctness and memory
baselines. At s16,m2,t5,N2^32,K918,672,421, reuse owns51.427 GiB GPU memory and
encoded in57.4/57.0 s. Compact owns35.429 GiB and encoded in101.5/93.3 s. Those full
runs are historical comparisons; matched small runs also show a compact penalty.

The [CUDA online kernel](../../encoding/liftrs-cuda/matrix.cu) reads P only to
initialize the first K prefix coefficients. It reads the full generator spectrum
only in the pointwise product between the full forward and inverse transforms.
This makes streaming possible without changing those five transforms:

1. Keep the same two-array reuse layout and immutable verified host plan. Clear
   prefix operands as before; upload the first K P coefficients directly into
   their arena destination, beside the separately uploaded message. P[K] remains
   part of verified code identity but the current prefix algorithm does not use it.
2. After the full forward transform, transpose scratch is dead until inverse.
   Borrow its first256 MiB as a spectrum upload tile. For each chunk, upload,
   multiply the matching arena range, then wait before overwriting the tile.
   Initially use synchronous pageable copies/default-stream ordering; asynchronous
   overlap/pinned staging is a separate optimization with its own memory cost.
3. Once every product completes, the inverse reclaims full transpose scratch.
   Canonical assembly and owning uint32 host output stay as the faster reuse path.

No full code array needs to stay on device. The projected owned count is the
measured reuse count minus its P/spectrum allocations: **about32.005 GiB**, including
the same measured table/plan/parity overhead. There is **no extra GPU allocation**
for the256 MiB tile because it borrows existing scratch. This is lower than the
measured35.429 GiB compact route and keeps fast rows; its actual timing is unknown.
Immutable P/spectrum remain on the host, so expected host usage stays near the
measured39 GiB, plus any copy/runtime effects. This is a VRAM reduction, not a RAM
reduction. It adds no disk reads after verified plan load and keeps8 GiB packed
output. No cryptographic/application semantics change.

The streamed call uploads16 GiB spectrum plus3.422 GiB prefix P per message:
**19.422 GiB code transfer**, in addition to the existing3.422 GiB message upload
and16 GiB output download. It uses64 spectrum chunks of256 MiB at s16, adding
copy/launch synchronization compared with one resident product. Current resident
setup uploads P/spectrum once, then amortizes that over messages. Streaming moves
that work into every encoding; first-use setup+encode must be measured separately
from reused encoding. The4-byte unused P tail explains the exact difference
between setup-code payload and per-call P upload.

If achieved aggregate code-transfer bandwidth were8/16/32 GiB/s, payload-only
transfer would take2.43/1.21/0.61 s. These are scenarios, not measurements or an
end-to-end forecast: pageable staging, synchronization, faults and kernel launches
can add time. Neither the57 s nor93 s previous encodings predicts the new result.
Retain resident fast execution as default when it fits admitted memory budgets.

A one-arena/direct-row streamed variant instead needs16 GiB arena plus256 MiB
staging before tables/parity at s16. It saves more VRAM but retains the measured
compact row-locality penalty. Its reusable staging can also serve final compaction;
do not count two mutually exclusive buffers. Host backing/upload costs remain.
This is a separate explicit alternative, not silently selected for lower memory.

## s17 width audit

Catalogue K3,870,933,765 is **estimated**, not verified support size. All s17
numbers depend on that estimate. N2^34 and L2^33 give2L=N; actual support and
preprocessing still need generation/verification. Preserve this distinction from
verified s16 K. Width changes are more than a larger count argument:

| Data | s16 | s17 requirement |
|---|---|---|
| Domain count N | 33 bits, already uint64 | 35 bits, fits existing wide count header |
| Coordinate/support exponent | 32 bits | 34 bits; uint32 files/values truncate |
| Base coefficient/message/canonical value | 16 bits | 17 bits; keep uint32 slots |
| Packed extension value | 32 bits | 34 bits; use uint64 or reviewed split representation |
| Tight persistent output | 8 GiB | 34 GiB at17 bits/value |
| One full extension array | 16 GiB | 128 GiB with uint64 slots |

[Go online loader](../../encoding/liftrs-shared/scaling_online.go) already reads
uint64 counts in LRSprobe2, but restricts s<=16 and decodes every P/spectrum value
as uint32. A wider spectrum needs an explicit compatible artifact discriminator;
old verified files/hashes must retain their meaning. Base P can stay uint32.
The [Go arithmetic adapter](../../encoding/liftrs-shared/orbit_arithmetic.go)
returns packed uint32 extension values, including generator/coordinate operations;
widen those paths rather than only changing accepted parameter limits. The base
field polynomial for degree17 already exists in [utils](../../utils/gf2n.go); that
does not make the packed extension/encoder paths support s17.
The [support checkpoint writer](../../encoding/liftrs-shared/s15_firstuse_probe_test.go)
casts indices to uint32; use a versioned64-bit support format for s17.

[Native field/transform](../../encoding/liftrs-cpp/transform.hpp) uses uint32
coefficients, coordinates, fixed-product tables and starts; its importer explicitly
rejects s>16. The CUDA field/kernel/bridge likewise borrows uint32 pointers,
uses32-bit generator/scatter positions and stores only32 coordinate bits per
plan lookup row. Review C ABI descriptions, limits and all casts; host size_t/Go
int are already64-bit on this machine, but that does not widen field values.
CUDA row plans also need18 basis levels for a17-bit split, beyond current17.
Base-field log/power table entries can remain uint32; widening all arrays would
waste memory. Bound degree/index shifts separately from their storage types.

The current uint16 compact code does not support s17. However, the same **disjoint
halves design can potentially generalize**: one uint64 arena is8N bytes, and each
half can hold N uint32 base values, enough for17-bit outputs. Check before narrowing,
scatter into the other half with64-bit coordinates, and download4N bytes directly
into the existing host uint32 result. This needs new kernels/ABI/layout proofs and
word tests; it is not implemented. GPU host expansion would no longer be required
for32-bit half values on the usual little-endian host, but conversion/portability
must be explicit. CPU ownership needs special care: returning a view of this
arena retains the entire128 GiB backing allocation, not only64 GiB visible output.
A separate64 GiB copy increases the peak and must not be omitted from accounting.

## Peak models and remaining capacity gates

| Principal arrays, GiB | s16 | s17 projection |
|---|---:|---:|
| Fast two-array GPU, streamed P/spectrum | 32.000 | 256.000 |
| One-arena GPU, streamed P/spectrum +256 MiB tile | 16.250 | 128.250 |
| Host verified full plan + message + owning output | 38.845 | 220.841 |
| Host file-backed constants + message/output + tile | 19.672 | 78.670 |
| Bounded CPU, resident constants | 39.345 | 286.841 |
| Bounded CPU, file-backed constants + tile | 20.172 | 144.670 |

File-backed routes require a new verified random/sequential-read code source and
separate plan metadata instead of today's fully materialized ScalingOnlinePlan.
Their RAM entries exclude page cache and driver/allocator/pinned staging. Reading
an mmap is not a RAM cap: cgroup/page-cache memory must be counted and guarded.
Resident-file backing is19.422 GiB at s16 /142.420 GiB projected s17, reused from
verified preprocessing; each encoding reads roughly that much if no full host
cache is retained. Final output adds8/34 GiB persistent storage; existing support,
root/G, old caches and atomic replacement overlap are additional, not deleted.
Do not describe reduced private RSS as a total host-memory saving without measuring
page-cache and cgroup behavior. Cold and repeated reads need separate timings. File-backed reads must preserve
the current trusted-manifest checksum checks and verified-code provenance. Initial
whole-file SHA verification costs another19.422/142.420 GiB sequential read if it
cannot be combined with trusted preparation; record that setup cost separately.
Keep verified backing stable across later reads (or authenticate chunks against
trusted metadata), and reject truncation/read errors before returning a word.
A pathname/stat check alone is not an integrity guarantee.

The available measured H200 NVL has140.40 GiB total capacity. A128.25 GiB single-
arena s17 projection may fit its physical capacity only after accounting for
wider tables, parity, driver/context and replacement headroom. It exceeds every
currently admitted compact40/reuse56 GiB device budget. The full-host-plan model
also exceeds the128 GiB request; CPU file-backed144.67 GiB exceeds it too.
**No s17 run or larger allocation is admitted by this model.** The fast256 GiB
GPU layout cannot fit this GPU even before overhead. Wider arithmetic, verified
preprocessing, streamed code and full peak/guard validation are still required.

## Proposed next prototype and validation

Recommended next route is **fast-layout GPU streaming at s16**. It reuses dead
scratch, isolates streamed-constant behavior from the already slower direct-row
layout, and has a projected32.005 GiB owned footprint under the current40 GiB cap.
Runtime remains unknown. The user subsequently selected this material trade-off;
[GPU-036](../gpu-streamed-constants/report.md) now implements its explicit prototype. Alternatives are compact GPU streaming or CPU file-backed
streaming, with the distinct costs above. No mode/default is changed by this audit.

For the selected route, retain resident implementations and add an explicit CLI
option/private prototype API, not an implicit fallback. Test full small CPU/Go
words, repeated/retained outputs, bad lengths/ranges, source immutability and
chunk tails. Exercise zero/one/exact-tile/multiple-tile boundaries, synchronous
copy dependencies and CUDA memcheck. Measure code upload and kernel stages, owned
allocations, driver/host peak, setup+first encode and reused encode on matched
inputs. Retain profiling and verified caches. Existing40 GiB device/128 GiB host
budgets need not increase for this s16 prototype; any long validation still uses
submit-and-handoff. The chosen route is fast-layout GPU streaming; its validation follows these boundaries.
