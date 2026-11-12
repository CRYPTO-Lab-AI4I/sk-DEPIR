# Bounded CPU transforms and compact output

Status: CPU-034 validation job73818 completed and evidence reconciliation passed. The preceding GPU compact
prototype was published as f1546ed. This CPU counterpart was developed independently of GPU job73815; both completed
results are now reviewed. Ordinary native CPU execution remains
the fast default; the experiment requires explicit `--memory-mode compact`.

## Schedule and trade-off

[Shared adapter](../../encoding/liftrs-shared/cpu_compact.go) exposes
`NewCompactWorkspace` and `EncodeCompactWithStats` beside the normal methods.
[Native implementation](../../encoding/liftrs-cpp/c_api.cpp) replaces the full
transpose scratch with independent batches of up to64 columns per worker. Each
batch gathers, transforms contiguous columns, then scatters them back before the
next batch. Different workers own disjoint columns and scratch. Rows already
operate directly on contiguous data; no GPU code or dependency is needed.

After the prefix inverse, keep its first K values in place and clear only the
tail, eliminating saved H. After the full inverse, check the base-field range
before narrowing into two little-endian bytes/value in ascending order. These
writes cannot overwrite a future uint32 source. Canonical scatter reads the
lower arena half and writes unique positions in the disjoint upper half. Join
workers before writing parity at zero, copy that compact half to the front, then
expand backwards into uint32 values. Return ownership of the arena and detach
it from the workspace. Later calls allocate a replacement; retained outputs stay
valid through reuse/errors/closure. Failures return no word. Workspaces have one
caller; constants remain immutable. No masked/cryptographic behavior changes.

The explicit compact route requires s<=16 and 2L<=N. Other shapes use the normal
API; unsupported compact calls fail. The returned representation stays uint32,
and packed database files still use s bits/value. This intermediate representation
is internal to encoding, not a new persistent format or a source of s17 support.

At s16,m2,t5, K918,672,421,N2^32,L2^31 with32 workers, bounded columns need512 MiB.
Principal arrays change from `12N+12K+4` to `8N+8K+4+columns` bytes:
**58.267→39.345 GiB modeled**, versus58.351 GiB previously measured RSS.
Runtime/allocator/native-plan overhead and retained old outputs are additional.
The model includes message, immutable P/spectrum, arena and bounded columns, not
arbitrary caller retention. Extra compaction, copy, expansion and different column
locality may slow encoding. Matched s14 now measures a3.9–6.1% encoding penalty; full s16 timing is a historical comparison. Stage
counters separate compaction/move and expansion; columns include gather/scatter.
When memory fits the faster validated path, compact remains explicit opt-in.

## Reproduce

[Runner](run.py), [Slurm launcher](run.sbatch), [analyzer](analyze.py),
[self-check](check.py), [Go ownership/workflow tests](../../encoding/liftrs-shared/cpu_compact_test.go)
and [native sanitizer fixture](../../encoding/liftrs-cpp/online_test.cpp).

```bash
python3 -B experiments/cpu-output-memory/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-output-memory/slurm-%j.log experiments/cpu-output-memory/run.sbatch compact-01 --memory-mode compact
# Only after the completion ping:
python3 -B experiments/cpu-output-memory/analyze.py compact-01 --recorded-cache
```

Runner defaults to `--memory-mode fast`; it never silently selects compact.
Actual revision, dirty-source snapshots, compiler/OS/CPU/affinity, commands,
binary/source/evidence hashes, profiles, phase clocks and sampled process-group
RSS are saved in `runs/compact-01/results.json` and phase/log/profile files.
Retain binaries and one8 GiB full packed output outside Git under
`.cache/liftrs/cpu-output-memory/compact-01/`; no existing caches are deleted.
Controller seals payload digests; recorded-cache analysis checks physical size,
metadata and sealed digest without another8 GiB bulk reread. This is mathematical
encoding only; application/masking/TRL5 acceptance remains outside the benchmark.

Read-only reusable s16 input comes from completed preparation job72875:
`.cache/liftrs/gpu-preparation/s16-m2-t5/online.bin`,20,854,558,905 bytes,
SHA256 `f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874`.
Matched s14 uses `.cache/liftrs/parallel-s14/s14-m2-t5/online.bin`,
1,277,647,153 bytes, SHA256
`88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8`.
No discovery/preparation is rerun for those instances. Setup/cache validation,
encoding, independent verification, packing and readback are timed separately.

## Validation and budget

Small tests compare complete nonzero/zero/repeated words with the Go reference,
retain earlier outputs, check input immutability, errors/recovery, wrong shapes
and short/aliased workspaces. Native ASAN/UBSAN and TSAN exercise both existing
fast schedules and compact execution with1/4 workers at s9, including parallel
column/scatter work, high-byte values and corruption above bit15 that must be
rejected before narrowing. Go race checks cover the binding; native TSAN covers
native accesses. Broader native/Go-only suites and vet run inside the job.

Then compare fast followed by compact at s14, two encodes each on one allocation
and the same retained input. Fixed order/two samples do not establish a general
latency distribution. Run s6 smoke and **one full s16 compact encoding**, checking
all K systematic positions,16 sampled independent curves,1027 packed readbacks
and complete canonical/packed hashes against CPU job73131 and the reconciled GPU
reference. Repeated ownership is checked on smaller cases; full-size timing
comparison is historical. No second full-size candidate or baseline is added.

CPU-only32 cores, no GPU,128 GiB requested RAM,48 GiB RSS guard and44 GiB Go soft
limit;32 GiB minimum node headroom and128 GiB initially available. Two-hour Slurm
limit,110-minute controller,90-minute full worker; matched worker15 minutes.
New cache and scratch guards16 GiB each,20 GiB initially free. Profiles remain
enabled. Expected allocation roughly40–90 minutes; compact slowdown is unknown,
so this is a scheduling estimate, not a measured latency claim. Automatic guards
remain active; no monitoring after submission until the user pings.

Pre-submission focused compact/normal ownership, shape, alias and full-word tests
pass. Optimized standalone native fixture passes at s9, including high-byte
values and rejection before narrowing. Python admission, packed metadata,
modeled column memory and synthetic ownership/repeat checks pass; Python syntax,
Slurm shell syntax and diff whitespace checks pass. Native sanitizers and full s14/s16 runtime/RSS checks subsequently pass below.

Job73818 was reviewed only after the user completion ping; allocation log
`slurm-73818.log`. No additional job is submitted by this review.


## Completed measurements

[Manifest](runs/compact-01/results.json), [analysis](runs/compact-01/analysis.json),
[full phases](runs/compact-01/encoding/phases.json),
[ASAN/UBSAN](runs/compact-01/sanitizer.log), [TSAN](runs/compact-01/tsan.log)
and [final accounting](runs/compact-01/accounting.log) pass review. Reconciliation
resolves344 exact source identities and rechecks sealed evidence/cache digests;
recorded-cache mode retains the completed controller's full8 GiB payload digest,
checking its metadata and length without another bulk reread. Source revision is
f1546ed plus exact recorded dirty-source snapshots. Go/native suites, Go-only build,
race/vet, owning-output/error/overlap checks and both native sanitizer builds pass.
One full s16 encoding checks all918,672,421 systematic positions,16 independent
sampled curves and1027 packed readbacks. Canonical and packed SHA256 values match
retained CPU/GPU references exactly. Small and s14 repeated outputs also match.

Measured on gnode17, Xeon6710E,32 CPU cores, no GPU; GCC14.2.0 and Go1.27.3.
Slurm completed exit0:0 in **49m22s**, within the40–90-minute forecast and2h limit.
No resource guard tripped. Profiles, preprocessing and new output caches remain.
Accounting above was collected after completion and is additional evidence, not
part of the controller's earlier sealed evidence map.

| Memory | Fast reference | Compact |
|---|---:|---:|
| Principal arrays, modeled GiB | 58.267 | 39.345 |
| Peak sampled process-group RSS, GiB | 58.351 | 39.409 |

Observed RAM reduction is18.942 GiB (32.46%), consistent with removed scratch/H
minus512 MiB bounded column storage. Slurm MaxRSS is41,341,424 KiB. Compact fits
the48 GiB RSS guard. The returned arena is caller-owned; retaining additional
words still adds16 GiB each. A successful call leaves only512 MiB in its workspace.

| Full s16 phase | Seconds |
|---|---:|
| Load/check online input | 74.087 |
| Native setup and compact workspace | 25.549 |
| Encoding | 1289.759 (21m29.8s) |
| Independent verification | 806.780 |
| Pack/write8 GiB | 64.159 |
| Packed readback | 2.032 |

Encoding includes23.586 s compaction/move and6.064 s backwards expansion;
columns/gather/scatter443.127 s, rows552.599 s and Taylor215.183 s dominate.
Compared with previous fast calls1281.328/1247.173 s, the new single call is
0.66%/3.42% slower, but this is historical and uses a different allocation and
Go memory limit. Independent verification also differs materially from the prior
run (806.8 versus about470 s); do not infer unchanged workflow time from encoding
alone or attribute that difference to one kernel. Load/setup/checks/I/O are separate.

Matched s14, same allocation, fixed fast-then-compact order:

| Schedule | First encoding, s | Second encoding, s |
|---|---:|---:|
| Fast | 64.462 | 62.117 |
| Compact | 68.413 | 64.517 |
| Compact penalty | 6.13% | 3.86% |

Sampled s14 RAM decreases3.593→2.526 GiB. Two fixed-order samples establish an
observed penalty, not a general latency distribution. No transforms or disk
streaming are added, but packing/move/expansion and changed column locality cost
time. Compact stays explicitly selected; fast remains the default when admitted
memory fits. This closes the CPU counterpart validation, not s17 or TRL5 gates.
