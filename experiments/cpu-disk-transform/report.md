# Bounded private-RAM CPU disk transforms

CPU-043 prototypes the last memory route after [CPU packed output](../cpu-packed-output/report.md).
It is a **transform component**, not yet a complete disk-backed online encoder.
Two mutable little-endian N-word files replace two in-memory arrays. Two RAM
tiles plus small serialization staging execute the same Taylor/column/row
arithmetic. Resident CPU remains default; disk mode is explicit and can be much
slower. Filesystem cache is extra and can approach the full backing files.

## Reproduction

[Controller](run.py), [Slurm script](run.sbatch), [analyzer](analyze.py),
[transpose/dependency/resource self-check](check.py),
[native component](../../encoding/liftrs-cpp/disk.hpp),
[Go binding](../../encoding/liftrs-cpp/plan.go),
[validation](../../encoding/liftrs-shared/cpu_disk_test.go).

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-disk-transform/slurm-%j.log experiments/cpu-disk-transform/run.sbatch disk-01 --memory-mode disk
# Only after the user reports completion:
python3 -B experiments/cpu-disk-transform/analyze.py disk-01
```

[Local compile/self-check evidence](compile-check.json) passed; runtime
validation **passed job74134**, run `disk-01`, submitted2026-11-12.
[Results](runs/disk-01/) and [Slurm log](slurm-74134.log) are populated by the
job; after the completion ping395 source identities and all cache/evidence
hashes reconciled ([analysis](runs/disk-01/analysis.json)). Exact HEAD (ce29557 plus recorded working sources),
source hashes/snapshots, commands, machine/compiler/filesystem and evidence/cache
hashes are recorded in `runs/disk-01/results.json`. New retained backing files,
binaries and profiles use `.cache/liftrs/cpu-disk-transform/disk-01/`. No prior
cache/profile is overwritten or deleted. Input/forward oracle fixtures from
GPU-039 job74121 are verified against its sealed cache hashes; those fixtures
already passed full Go and resident GPU comparison. This job also compares
resident native CPU against each retained oracle; full Go work is not repeated.

16CPU, no GPU,16GiB host request/12GiB sampled RSS/8GiB Go soft limit; new cache
and scratch16GiB each. Initial32GiB node availability,16GiB runtime headroom.
GCC14.2, Go1.27.3, Slurm25.05. Controller20min, workers10min, Slurm30min.
No larger storage budget or full s16 backing allocation is admitted. Guards run
within the job; no agent monitoring before completion ping. Record `df -T`:
this measures the actual shared cache filesystem, not an assumed local SSD.

## Algorithm and file contract

`liftrs_matrix_apply_file`/`ApplyMatrixFile` borrow two distinct read/write regular
file descriptors, aligned byte offsets, count N, tile T, immutable prepared
matrix plans and worker count. The caller must keep descriptors open and exclude
other access/truncation. All ranges and file lengths are checked before mutation.
No descriptor is retained. Errors can leave both files partially transformed;
restore verified input before retrying. There is no checkpoint/resume transaction.
Native execution does not fsync; validation Syncs after timed transforms.

Taylor passes read bounded destinations/sources, apply precomputed products and
write joined results. Quotient sources stay in the unchanged high half; remainder
reads the completed quotient in the disjoint upper half; inverse reverses passes.
Blocking preserves these dependencies across tile boundaries. Columns use a
blocked transpose to the scratch file, contiguous column transforms, then a
blocked transpose back. Rows use contiguous vector batches and global affine
origins. Each complete vector must fit T. Same scalar/table arithmetic and
canonical field/matrix descriptions as resident CPU, no protocol changes.

Files explicitly decode/encode little-endian uint32 values. Two T-word vectors
plus a64KiB serialization buffer suffice; matrix metadata/runtime/thread stacks
are extra. No full array, mmap or N-sized permutation map is allocated by native
execution. Byte-range prefix/suffix canaries validate nonzero offsets. Private
buffer size is reported, not inferred from RSS. File cache can retain all backing
bytes even when process buffers are small; memory pressure may evict them.

## Validation and measurement

Eight cases match GPU-039: `(s,dimension,T)`
`(4,7,48),(4,8,48),(6,11,320),(6,12,320),(8,15,1536),(8,16,1536),`
`(16,20,65536),(16,22,65536)`. Rectangular bases, non-power-of-two tiles and
row/column batch/block tails are covered. Native resident CPU forward, file
forward/inverse and restored-input recovery compare every value to reference.
File aliasing, unaligned offsets, short tile/file, read-only descriptors, field
range, closed plans and canaries are checked. Six smaller cases repeat with
AddressSanitizer. Ordinary Go/CPP, combined API build and vet also run.

Timings seconds/ns as labeled; I/O bytes and syscalls are logical operations,
not physical device traffic. Stats separate Taylor, columns (including both
file transposes), rows, nested read/write time, read/write bytes/syscalls,
private buffers, native total and transpose count. I/O times overlap phase
clocks and must not be added twice. Go elapsed includes admission/dispatch;
native total starts after admission and includes allocation. Test input setup,
file reads for oracle validation, hashing/Syncs and resident comparison are
outside transform timing. File cache is warm/dirty; kernel I/O timings may
largely measure cache. No physical-disk throughput extrapolation.

Profiles retained for full/sanitized components. Process RSS includes full
reference/resident validation arrays, not only production buffers. Final retained
backing-file cache residency uses existing mincore observation without faulting
pages; it is an end-state sample, not an encoding peak. Job-scoped cgroup data
(current/peak/file/anonymous) is recorded when available and labeled unavailable
otherwise. These accounting views are separate, not interchangeable totals.

## Trade-off before larger adoption

Resident transform workspace8N bytes becomes at most8min(N,T)+65536 private
buffer bytes plus metadata/runtime, with8N bytes of backing storage. At N2^22,
T65536:32MiB arrays become576KiB private buffers, with32MiB two-file backing
(plus test canaries). Reference/test arrays and file cache remain additional.
At s16, the same T projects576KiB private buffers plus metadata/runtime and
**32GiB backing files**, before code, message, output, snapshots or cache. This
exceeds the current16GiB new-cache budget and is not a full-run admission.

Columns alone read/write3N words each direction; rows addN each, so at least16N
bytes read and16N written per transform, before Taylor passes. At s16 that is
at least64GiB read+64GiB written per transform. Taylor adds repeated reads/writes
per coefficient/level; input field validation occurs on reads. Blocked transpose
creates many subrange syscalls and cache dirtying. Small T also restricts row
parallelism (only one row per batch at full s16/T65536). Larger T may improve
runtime but uses more private memory; it must be chosen and measured explicitly.
Runtime penalty, physical storage performance and full online/code/output
lifetimes are unknown. Full CPU online integration follows component validation;
s17 field/coordinate changes remain separate. No faster default is replaced.


## Measured component result

All8 forward/inverse/resident/retained-reference cases and restored-input/error/
canary checks pass;6 small cases pass ASan. Ordinary builds/vet pass.

| s16 component,T65536 | Resident seconds | Disk forward seconds | Disk inverse seconds |
|---|---:|---:|---:|
| N2^20 |0.149945|0.747427|0.531708|
| N2^22 |0.630953|4.720644|2.599884|

At2^22, native buffers589824bytes (576KiB) replace32MiB resident arrays;
metadata/runtime extra. Backing33554496bytes includes64 canary bytes. Each
transform reads1220497408 and writes205475840 logical bytes via83621 reads/
38464 writes. Forward Taylor1.921s, columns2.508s, rows0.270s; nested write2.317s
versus inverse0.167s. This is NFS on `/hpc/home`, with variable cache/allocation/
writeback behavior; no physical throughput claim. Forward7.5x resident, inverse
4.1x, not an online encoding estimate. Small T and sync I/O impose real costs.

Controller-sampled process-group peak1057570816bytes includes build/tests;
job cgroup peak1524035584bytes, final file658202624bytes, anon11468800bytes
include build/cache and are not isolated component RAM. End mincore samples
all28 work files at0 resident pages. That single post-Sync sample is **not**
evidence of no cache consumption during transforms. Private buffers, observed
RSS, final cache residency and cgroup charges are distinct views. Profiles and
fixtures retained. Full CPU online integration is next; larger storage/runtime
admission and s17 changes remain separate.

Final accounting: gnode17, COMPLETED0:0,3m34s, MaxRSS203544KiB;
[accounting](runs/disk-01/accounting.log). This per-task sample differs from
the controller group/cgroup values and is not an isolated encoding peak.
