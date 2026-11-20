# GPU-resident rank storage and transpose

Status: first storage prototype implemented; local CPU/race, C/CUDA and complete
CPP/CUDA/MPI build checks pass. **MPI/GPU execution and CUDA sanitizer results
are pending.** Source starts after `3a47a10`; the committed implementation and
actual job-start revision will be recorded in the submission receipt/raw results.
[Readiness commands, hashes and limits](readiness/checks.json),
[CUDA build log](readiness/cuda-build.log), [MPI build log](readiness/go-mpi-build.log).
No fulls18 preprocessing, resident FFT/product, memory peak or speedup is claimed.

This implements the first gate in the [resident-storage design](../../docs/distributed-preprocessing-storage.md),
after [bounded distributed products passed78321](../distributed-gpu-product/report.md).
We remain on step1: distributed large products/FFTs. The completed full s17
preprocessing/CPU/GPU encoding results are unchanged.

## Implementation and invariants

[`transposeStored`](../../encoding/liftrs-shared/distributed_transpose.go)
now owns the single shared peer-round/tile schedule. Host and
[CUDA storage adapters](../../encoding/liftrs-shared/distributed_storage_cuda.go)
supply local copy, pack and scatter; descriptors, wire order, ownership, proof
checks and progress boundaries remain shared. No field or FFT recurrence is
copied. CPU rendezvous/serial tests cover both widths and1/2/3/4/5/7/8/16/64 ranks,
including uneven/empty bands. Focused storage-error tests ensure local/pack/scatter
failures never publish the final completion notice. CPU/race checks pass.

[`RankStorage`](../../encoding/liftrs-cuda/rank_storage.go) owns a dedicated
single-GPU executor. The [native adapter/kernels](../../encoding/liftrs-cuda/rank_storage.cuh)
use three reusable device arrays and one bounded device transfer tile, all counted
by the same per-card cap/OwnedBytes. Array capacity is the larger local band of
the original matrix and its transpose. View reset clears one slot. Row-major and
interleaved `[column*ownedRows+row]` physical addressing implement the same logical
matrix; GPU pack/scatter directly translates each view. Bounded host I/O uses
logical row-major offsets. Upload/read/pack/scatter reject invalid ranges; local
transpose rejects aliasing. No borrowed host pointer survives a call. Go serializes
calls and destroys the borrowing storage handle before its executor. The optional
ABI does not disable existing operations in older plugins. CPU client builds stay
CUDA/MPI-free; no production default or s18 arithmetic admission changes.

Source and third slot survive each transpose. A target is incomplete on failure;
reset/discard it rather than treating it as a result. Storage moves **raw slot
bits**, deliberately including values outside the field in the test: arithmetic
and field validation are not performed here. All initial arrays are zeroed, and
empty owners launch no zero-size kernels or transfers.

## Bounded coverage and independent review

The [MPI/GPU test](../../encoding/liftrs-shared/distributed_storage_gpu_test.go)
uses s4 four-byte slots and s17 eight-byte slots. Shapes for both widths:
1x1,1x3,3x1,7x11,3x17,16x16; eight-byte slots also64x64. These cover zero owners,
unequal row/column ownership, square/rectangular shapes and127-word tile tails.
All four source/target physical layout combinations run forward and inverse.
There are **104 passes per rank,208 rank/pass rows per execution**, repeated under
CUDA memcheck/leak checks. Owners/arrays are reused across all layouts and passes;
retained device allocation must remain unchanged. Each pass checks every output
word, unchanged source and third-slot sentinel. Invalid shape/slot/extent,
ownership mismatch and alias calls reject before a legitimate pass.

The [reviewer](analyze.py), using [run.py](run.py), independently reconstructs
output bytes by global coordinates, checks exact peer stream lengths/operation
IDs/paired proofs, all layouts, saved expected/output files, ownership and
completion. It reconciles source/evidence/cache seals and both stdout/stderr CUDA
summaries. [Four reviewer/launcher tests](test_review.py) reject corrupted bytes,
proof/metadata/preservation/reuse changes and accidental selection of the old
vector driver. They run without GPU/MPI, alongside the shared reviewer tests.

## Reproduction and retained evidence

[Slurm entry](run.sbatch) reuses the validated
[composition controller](../distributed-gpu-vectors/run.py), staged MPI transport,
shared-path remote preflight and explicit UCX cache control. Source/build-input
hashes, exact revision/commands, profiles, progress/resource logs and exit codes
persist in `runs/<name>/results.json`; normal and sanitized arm results are saved
in `runs/<name>/composition.json` before later phases. Checksummed binaries,
output/reference shards and caches remain outside Git at
`.cache/liftrs/distributed-gpu-storage/<name>/`. Existing caches remain intact.

```sh
python3 -B experiments/distributed-gpu-storage/test_review.py
python3 -B experiments/distributed-gpu-storage/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-gpu-storage/slurm-%j.log \
  experiments/distributed-gpu-storage/run.sbatch storage-01
# After the user's completion ping:
python3 -B experiments/distributed-gpu-storage/analyze.py storage-01
```

Same admitted HPC envelope: **2 nodes /2 H200s**, one GPU,2 CPU cores and4 GiB
requested host RAM per node (**4 cores /8 GiB total**); **128 MiB owned GPU/card**,
**1 GiB RSS guard/rank**, no pinned buffers, **128 MiB retained cache**, **2 GiB
scratch**, **32 MiB/file**, **720-second controller /15-minute scheduler**.
gnode03 excluded; other eligible nodes allowed. Compilation and executable/
wrapper/library/cwd/TMPDIR preflight on both nodes precede GPU work. No budget
increase, large workload, new disk intermediate or agent monitoring is authorized.

78321's complete product suite took2.172 seconds normal /7.084 seconds sanitized,
and its controller141.151 seconds. This storage suite has more small geometry/
layout cases but no field arithmetic or Taylor source traffic. Each new MPI arm
keeps240 seconds (220-second test timeout), within the existing720-second
controller, with substantial headroom over previous bounded suites; runtime is
unmeasured. All cases are deliberately tiny (maximum4096 words) rather than a
scaling or bandwidth measurement. Controller/command times include build,
setup/validation and file work; per-pass diagnostic clocks measure target reset,
local GPU copy, packing, transfers, MPI proofs and scattering, taking the maximum
rank time. They exclude constructor/input upload/oracle/readback/verification/
fsync. CPU profiling and per-tile flushed progress/resource logging are enabled,
so per-pass times are instrumented diagnostics, not unprofiled latency baselines.
Native local-copy launches are bounded and report increasing work for long copies.

## Memory and time boundaries

For each rank, capacity is
`max(owned_original_rows*columns, owned_transposed_rows*rows)`. Device payload is
`3*capacity*slot_bytes +127*slot_bytes`, plus executor field tables/metadata,
context/driver excluded from owned bytes. Empty owners still retain the transfer
tile and metadata. Largest test main arrays total **48 KiB/card** (64x64,2 ranks,
eight-byte slots), plus **1016 bytes/card** transfer workspace. The scheduler
uses at most127 host words (**508 or1016 bytes/rank**), not whole host shards.
Diagnostic input, sentinel, readback, source and expected arrays are small
additional host allocations; RSS includes them and MPI/CUDA runtime state.

The [fullsize design](../../docs/distributed-preprocessing-storage.md) projects
three GPU shards at1536 GiB aggregate for s18, versus up to3072 GiB aggregate host
principal arrays in the old reference. Those are payload models, not measured
peaks or complete pipeline budgets. Resident vectors, Taylor source access,
pointwise/product lifetimes, setup cost, complete I/O and larger measured
throughput remain next. There is no measured memory/time benefit yet. Retain the
fast validated production defaults; this optional storage prototype does not
silently select a slower memory-saving path or authorize a16-GPU allocation.
