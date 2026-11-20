# Rank-sharded transpose for collective preprocessing FFTs

Status: **MPI validation 78279 passed** (SCALING-142).
[Receipt](submission.json), [raw results](runs/transpose-01/results.json),
[review](runs/transpose-01/analysis.json), [accounting](runs/transpose-01/scheduler-accounting.txt),
[progress](slurm-78279.log). Implementation `e96ed6a`; verified
**672 source / 39 original evidence / 18 cache hashes**; all commands exited zero.
[Shared transpose](../../encoding/liftrs-shared/distributed_transpose.go),
[CPU/rendezvous tests](../../encoding/liftrs-shared/distributed_transpose_test.go),
[MPI harness](../../encoding/liftrs-shared/distributed_transpose_mpi_test.go),
[runner](run.py), [reviewer](analyze.py), [review tests](test_review.py).
This validates the data-movement/control prerequisite, **not** a distributed FFT,
collective polynomial product, GPU packing kernel or full s=18 preprocessing.

## Review of 78279

HPC gnode14/gnode15, 2 nodes / 4 ranks / 4 CPU cores / 8 GiB requested host RAM
total / zero GPUs. Scheduler **1 min 18 s**, controller **76.975 s**. Both modes
passed all eight shape/width operations: **16 global transposes / 64 rank outputs**.
All **7,312 endpoint proof records** match peer proofs and independently generated
tile values, with exact operation/route/offset/width/count coverage. Every output
word and all saved largest shards match independent coordinate references. Empty
owners, odd shapes, tiny tails, multi-chunk messages and original root/support
lanes pass. No mismatch captures, guards or recovery; all rank streams parse
strictly. Source/evidence/cache seals, profiles and UCX cache settings retained.

| Mode | Largest 32-bit transpose, 1024×256 | Largest 64-bit transpose, 1024×256 | Largest logged process peak RSS/rank |
|---|---:|---:|---:|
| Staged | 6.257 ms | 5.476 ms | 249.312 MiB |
| Direct | 8.865 ms | 4.697 ms | 248.145 MiB |

One profiled sample per shape/mode; slowest-rank durations include tile allocation,
scalar packing/scatter, metadata, hashing/proofs and transfer, excluding generation,
full validation, artifact/fsync and final collective aggregation. They are not
isolated network bandwidth, statistically stable mode comparisons or GPU timings.
Different shapes use different tile bounds; smaller shapes can take longer due
to more messages. Do not extrapolate full-size runtime from these samples.
Largest explicit shard+tile payload remains **1.125 MiB/rank**, plus optional
**8 KiB** MPI staging; measured RSS includes runtime/transport overhead. Retained
cache **12.666 MiB**; all eight rank CPU profiles and largest per-owner outputs
remain. Build **19.987 s**, race check **18.149 s** and profile report **28.267 s**
are separate from the two MPI commands (**1.572 s / 1.270 s**).

**Next:** expose stage-scoped GPU vector operations using existing `local_vectors`
and shared `vector_transform` recurrences. Global plan dimension and vector origin
must remain independent of the locally admitted buffer size; nonzero/wide global
origins affect row evaluation points. Current `HostTiled.rows_at` demonstrates a
diagnostic affine-row tile, not the production distributed executor. Validate
row/column forward/inverse stages at both widths against independent CPU results,
then compose transpose plus vector stages and distributed Taylor dependencies.
Reuse resident device arenas and bounded host exchange tiles, rather than adopting
the two-full-host-shard reference as a production layout. Extra transfers/runtime
and device workspace remain unmeasured; no layout/default decision or new large
budget follows. No new job submitted in this review.

## Implementation and invariants

A row-major Rows×Columns tensor has contiguous complete input rows assigned to
ranks; its transpose has contiguous complete output rows (original columns).
Quotient/remainder partitioning handles uneven shapes and empty owners without
size×rank overflow. Dimensions, offsets, operation IDs and word counts are64-bit;
Go array byte extents and caller admission are checked before allocation. A
virtual2^18×2^18 layout on16 ranks is tested without allocating the huge arrays.
Values are canonical-width storage slots, uint32 or uint64; every bit is preserved,
with no field arithmetic or representation change.

Each worker owns only its input/output shard and one reusable tile. Pairwise
rounds send to rank+round and receive from rank−round; rank<destination chooses
send-first, otherwise receive-first. Every nontrivial permutation cycle has both
orders, avoiding an all-blocking-send cycle. Unbuffered test channels exercise
this with1/2/3/4/5/7/8/16/64 ranks, odd shapes, tails and empty owners. Each
inter-rank word is sent once; local values copy directly to the owned output.
Packing/unpacking is a scalar host reference, not a GPU packing performance claim.
Per-round directed cycles can serialize some links; a future communication
schedule must be measured before choosing a production layout.

The MPI adapter checks a fixed64-byte descriptor before each tile: operation,
global shape, round, block offset, count, word width and source. Mismatches fail,
not silently reinterpret data. Separate coefficient tag pairs40/41 and50/51 keep
32/64-bit FFT tiles distinct from existing support10/11 and root20/21 transfers.
All lanes reuse the same bounded C chunk loops and optional MPI-owned staging.
Receive fills borrowed caller scratch directly, without another tile allocation.
The control exchange and symmetric fingerprint/capture acknowledgement are shared.
Both endpoints fsync/close mismatch payloads before abort. Hashing is bounded
and preserves high bits even in invalid received data; it is not authentication.
No concurrent MPI streams or asynchronous borrowed pointers are introduced.

## Validation boundaries

CPU tests compare gathered small outputs with an independent serial transpose,
check unchanged inputs, failure propagation, invalid admission and wide virtual
counts. Tagged descriptor tests detect operation/source/width aliasing and
32-bit truncation. The MPI test uses4 ranks and runs shapes1×3,7×11,64×128 and
1024×256 in both widths, under staged and direct modes, always with the validated
explicit UCX registration-cache control. No worker creates a full matrix/oracle.
Local inputs and every expected output are independently generated from global
coordinates; uint64 patterns exercise all bits (they are transport fixtures,
not a claimed field-support set). Small tiles7words exercise tails; largest tiles
16,384words exceed eager-sized buffering. MPI chunks1,024words force multi-chunk
messages. Every descriptor, tile proof and output word is checked; largest local
outputs are saved per owner and later verified independently in Python. Original
root/support lanes receive separate wide-index/multi-chunk regression checks.

No GPU is needed to validate this permutation/control prerequisite. Native GPU
kernel composition remains the next implementation: connect local FFT/vector/
Taylor stages and bounded transposes while retaining rank workers. This reference
must not silently become a CPU path inside GPU server preprocessing.

## Memory/time trade-off and budgets

Explicit storage per rank is `(InputWords + OutputWords + ScratchWords) × width`,
plus optional MPI staging, metadata/hash buffers, runtime/allocator and profiles.
In the largest actual case: input0.5 MiB +output0.5 MiB +tile0.125 MiB =**1.125 MiB**
for64-bit words, plus **8 KiB** MPI staging when selected;32-bit case uses half
the storage. Admission allows up to2^18 words per principal array (2 MiB at64-bit).
An empty owner may need no tile. At a projected full domain, these are full local
shards: this plan does not claim that adding another host shard is free or fits.
GPU-resident integration must account for host/device lifetimes and headroom anew.
For a2^36-element s=18 grid at64-bit working width over16 ranks, this host
reference would retain **64 GiB/rank /1 TiB aggregate** for input+output alone
(256 GiB/node with4 ranks/node), before all other state. One transpose moves
**480 GiB aggregate off-rank payload** (including intra-node rank transfers),
not a measured network volume or runtime. These are reference-layout projections,
not an adopted full-size budget. The fast GPU path should reuse resident device
arenas and bounded host staging instead of blindly retaining these host copies.

Compared with an already-local transform this adds host packing/scatter copies
and inter-rank traffic. Returning to original orientation adds another exchange.
There is no disk-backed intermediate exchange, added field transform, full gather,
or automatic slower fallback. Runtime and communication penalty are **unknown**;
this remains an explicit experimental reference, not an adopted production trade-off.

Resources: **2 HPC nodes /4 ranks /4 CPU cores /8 GiB requested host RAM total /
zero GPUs**,1 GiB process RSS guard/rank, GOMEMLIMIT512 MiB/rank. Existing caps:
128 MiB retained,2 GiB build scratch,32 MiB/file,720 s controller/15 min scheduler.
Each new MPI mode has180 s test /200 s step /220 s command cutoff; the controller
still bounds the total. Profile summary90 s. No resource increase. Expected
complete job **2–4 min after scheduling**, a projection from prior builds/transport
runs, not a measurement. No four-node/full-size bandwidth claim follows.

Timed transpose includes local output/tile allocation, packing, metadata/proof
traffic, exchange and scatter. Input generation, independent verification,
artifact/fsync, barriers/max aggregation and profile reporting are outside that
clock but remain in total job time. Scalar checks/hashes/logging perturb timings;
these are not GPU throughput or full preprocessing estimates. All rank CPU
profiles retained. Native chunk logs, phase starts/completion, advancing word
counts and30 s waiting heartbeat identify the operation; an unchanged heartbeat
is not forward progress. Strict separate rank streams prevent launcher warnings
from corrupting worker events. Stop on failure; failed modes stay failed.

```sh
python3 -B experiments/distributed-transpose/test_review.py
python3 -B experiments/distributed-transpose/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-transpose/slurm-%j.log \
  experiments/distributed-transpose/run.sbatch transpose-01
# After user completion ping only:
python3 -B experiments/distributed-transpose/analyze.py transpose-01
```

Results/profiles/arm checkpoints and rank streams:`runs/transpose-01/`; progress:
`slurm-JOBID.log`. Reusable binary/largest local outputs/failure captures outside
Git:`.cache/liftrs/distributed-transpose/transpose-01/`, all sealed in results.json.
Public inputs regenerate from the documented pattern; no costly preprocessing
is discarded. Existing caches/profiles stay intact. Record toolchain: Go1.27.3,
GCC14.2/OpenMPI4.1.8/UCX1.18/Slurm25.05. Regenerate from the committed revision and
fresh name. No agent monitoring after submission.

[Local readiness](readiness/checks.json): ordinary CPU tests, race check and
MPI-tagged unit/descriptor tests pass; native MPI/CUDA harness compiles and prior
fingerprint tests pass. Python reviewer tests reject missing packets and corrupted
saved bytes, alongside36 existing reviewer/logging tests. Compiler-wrapper sandbox
socket warnings exit zero. The local checks did not execute MPI/GPU work; the actual MPI results are reviewed
above. GPU FFT composition and full-size behavior remain unmeasured.
