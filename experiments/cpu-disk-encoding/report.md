# Complete CPU disk-backed encoding

CPU-044 integrates the validated [disk transform component](../cpu-disk-transform/report.md)
into the complete five-transform mathematical encoder. Code remains verified/file-
streamed, two N-word work files hold mutable transform domains, and one owning
N-word host output remains. `OpenScalingCPPDisk` is explicit; faster resident/
file-fast defaults stay unchanged. This is a memory/time/storage trade-off, not
an expected speedup. Full s16 backing storage is not admitted; s17 unsupported.

## Reproduction and provenance

[Controller](run.py), [Slurm script](run.sbatch), [analyzer](analyze.py),
[resource/offset self-check](check.py), [shared API](../../encoding/liftrs-shared/cpu_disk.go),
[native schedule](../../encoding/liftrs-cpp/disk.hpp),
[small and complete tests](../../encoding/liftrs-shared/cpu_disk_online_test.go).

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-disk-encoding/slurm-%j.log experiments/cpu-disk-encoding/run.sbatch online-01 --memory-mode disk
# Only after the user reports completion:
python3 -B experiments/cpu-disk-encoding/analyze.py online-01
```

[Local compile/self-check evidence](compile-check.json) passed; runtime validation
**passed job74141**, run `online-01`, submitted2026-11-12.
[Results](runs/online-01/) and [Slurm log](slurm-74141.log) are populated by the
job; after the completion ping401 source identities and all cache/evidence
hashes reconciled ([analysis](runs/online-01/analysis.json)). Revision ce29557 plus recorded working sources;
exact HEAD/source hashes/snapshots, commands, CPU/compiler/filesystem, results,
profiles and cache hashes go into `runs/online-01/`. Retained cache
`.cache/liftrs/cpu-disk-encoding/online-01/` contains binaries, workspace files,
wide-header fixture and packed workflow outputs. Small fixtures from CPU-042
job74131 are checked against sealed hashes; complete s10 code reuses trusted
`.cache/liftrs/gpu-products/products-02/cpu-0/online.bin` with mandatory SHA256.
No previous preprocessing/profile is deleted or expensive discovery repeated.

16CPU, no GPU,16GiB requested host RAM,12GiB sampled group RSS/8GiB Go soft
limits;16GiB each new cache/scratch. Initial32GiB availability/runtime16GiB node
headroom;20min controller,10min checks,30min Slurm. GCC14.2/Go1.27.3/Slurm25.05.
Actual workspace filesystem recorded with `df -T`. No local SSD throughput is
assumed. Guards run inside the job; no agent monitoring until completion ping.

## API and algorithm

Constructor checks trusted code and creates exclusive0600 `data.bin` and
`scratch.bin` (4N bytes each) in the caller-selected directory. Existing files
are never replaced; failed construction may leave newly created files. Close
serializes and releases handles/metadata, preserving workspace artifacts for
caller management. They are not resumable checkpoints or a durable output.
Caller must exclude external changes/truncation; trusted code remains immutable.

Native execution initializes two L prefix operands in the data file, performs
forward/forward/product/inverse, retains K coefficients and zeroes the tail.
Full forward, streamed-spectrum product and inverse follow. All extension
arithmetic stays native CPU. Bounded reads of the final coefficients scatter
canonical values into the new owning host result, with joined zero-coordinate
parity and a checked primitive generator; no full orbit map exists. Products,
code copies and zeroing use RAM tiles. All descriptors/pointers are borrowed
synchronously. Inputs/earlier owning outputs remain unchanged. Errors return no
usable output; a subsequent call starts initialization again, after restoring
required file lengths/availability. There is no rollback/fsync/checkpoint promise.

Two product/assembly tiles remain live while a disk transform uses two more;
private buffer accounting conservatively includes4T words plus two64KiB I/O
staging buffers. It explicitly includes this nested liveness, rather than
claiming the component's two-tile bound for the full encoder. Metadata, runtime/
thread stacks, allocator retention and OS cache are additional. Code probe
headers are25/33 bytes; code offsets are allowed unaligned, work offsets align
naturally. Both artifact versions retain their existing bytes/semantics.

## Validation

Nine CPU-042 fixtures: constant subcodes s2..8 and nonconstant s6/s8 compare every
value with Go for nonzero/zero/repeated messages. The s6 case additionally uses
a wide33-byte header derived from the same trusted coefficients. Check five
transforms/one Go-C boundary, buffer/backing/output counters, no retained code/
full RAM work arrays, invalid messages, short workspace/error recovery, detached
output isolation, closed state and refusal to replace existing workspace files.
AddressSanitizer repeats. Ordinary Go/CPP and combined API build/vet run too.

Complete cached s10 file-fast and disk modes each perform two encodings with
the same seed71 message, all146641 systematic positions,16 sampled independent
curves per result, full canonical/packed hashes and1027 packed-file readbacks.
Both use the file loader without resident P/spectrum. Shared reference harness
keeps load, setup, encoding, final packing/writing and verification separate.
The result file is packed once after the repeated encodes, not once per encode;
this is an encoding comparison, unlike CPU-042's complete packed-file pipelines.

## Memory, storage and timing trade-off

File-fast needs8N private transform/output bytes. Disk mode retains only4N
owning output bytes plus4K message, conservative16min(T,N)+131072 native buffer
bytes, metadata/runtime; code is streamed in both. At s16, principal output+
message payload projects19.42GiB versus file-fast35.42GiB, with32GiB mutable
backing storage and19.422GiB reusable code extra. Workspace cache may use up to
both files, and code/output cache are additional. A packed disk-mode sink is
not implemented; later packing uses the owning output. No hidden full arrays
are substituted behind a bounded sink API.

At T65536 the conservative native buffer payload is1.125MiB. Small T also
restricts row batching and can worsen runtime; larger T uses more RAM and must
be selected explicitly. s16's32GiB backing alone exceeds current16GiB new-cache
admission, so no full run or runtime projection follows. s17 arithmetic/coordinate
changes remain required independently.

Five transforms add large repeated file reads/writes; products upload no GPU
data but read both operands and write results. Assembly adds sequential data
reads and random host output scatter. I/O counters are logical syscalls/bytes;
NFS/local cache can serve them. Nested read/write time overlaps transform/copy
phases; product/assembly read/write intervals are not all assigned to those
compute phases, so do not sum overlapping clocks. Native total includes buffer
allocation, map setup and full execution after admission; Go elapsed also
includes admission, owning output allocation and source metadata checks. Neither
includes fsync; filesystem durability is outside the timed contract.

CPU profiles retained for small/ASan/workflow phases. RSS/cgroup include build,
reference/verification arrays and cache, not just encoder private buffers.
End-state mincore samples are not a peak or evidence of zero cache during work;
job-scoped cgroup data is recorded when available. These are profiled fixed-order
samples on the recorded filesystem, not cold physical-disk throughput estimates.


## Measured result

Job74141 completed0:0 in8m33s on gnode17, using the recorded NFS workspace.
[Accounting](runs/online-01/accounting.log). All9 small fixture/Go cases,
constant words, wide33-byte header, input/error/workspace recovery, ownership,
Close/existing-file checks pass, including the ASan repeat. Ordinary builds/vet
pass. Complete s10 file-fast and disk modes each pass all146641 systematic
positions,16 sampled curves per encoding, repeated canonical/packed hash
agreement and1027 packed readbacks. All profiles/workspaces retained.

| Profiled s10 phase | File-fast | Disk-backed |
|---|---:|---:|
| Code load plus constructor seconds |0.010319|0.015592|
| Harness online setup seconds |0.0000017|0.0000025|
| First encoding seconds |0.331525|2.860369|
| Reused encoding seconds |0.310607|1.135627|

Constructors are called in `load-online`; tiny setup times are harness closure
installation, not actual encoder initialization. Disk mode is8.6x slower on
first call and3.7x slower on reuse in these fixed-order samples. Native writes
1.864s first/0.0765s reuse show strong filesystem/allocation/cache variability;
no physical-disk or cold-storage extrapolation. Native read time0.222/0.219s,
logical reads490210116bytes and writes163054404bytes per encoding via98637/
40040 syscalls. Same I/O volume in both repetitions. Product/kernel and random
canonical scatter costs are also included, not just disk syscall time.

At N2^20/T65536, owning host output4MiB and conservative private native buffers
1.125MiB replace file-fast's8MiB transform/output arrays; input, metadata,
allocator/runtime and filesystem cache remain additional. Two workspace files
8MiB persist. No full host code or transform array is retained. At s16 the
output/message projection remains19.42GiB plus buffers/runtime/cache, with32GiB
workspace files requiring separate admission; no full-size timing is measured.

Controller process-group peak1056096256bytes and job cgroup peak1520631808bytes
cover build, validation and caches, not isolated encoding memory. Final cgroup
file charge629231616bytes includes unrelated build/cache; end-state mincore for
43 retained files totals7446528bytes. These are separate observations, not proof
of a bounded total-memory peak. Slurm MaxRSS163672KiB is another per-task sample.

This completes small/full-s10 validation of the explicit owning disk route.
The four approaches now exist for both CPU and GPU, at the validation sizes in
[the consolidated status](../../docs/encoding-scaling.md#current-memory-route-status).
Combining bounded output with backed transforms, full s16 validation of the
new small-only modes, budgets and s17 arithmetic/preprocessing remain separate.
Faster defaults remain; slower modes require explicit selection.
