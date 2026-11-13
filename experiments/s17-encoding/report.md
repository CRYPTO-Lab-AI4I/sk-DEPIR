# Complete wide CPU encoding integration

Connects width-aware code data to the complete native cyclic encoding schedule,
continuing steps2–3 of the [s17 plan](../../docs/encoding-scaling.md#current-execution-plan-for-s17).
[Controller](run.py), [Slurm script](run.sbatch), [direct C checks](check.cpp),
[shared implementation](../../encoding/liftrs-shared/scaling_wide_cpp.go),
[native schedule](../../encoding/liftrs-cpp/wide_api.cpp),
[shared tests](../../encoding/liftrs-shared/scaling_wide_cpp_test.go),
[raw API tests](../../encoding/liftrs-cpp/wide_online_test.go),
[runnable example](../../encoding/liftrs-shared/scaling_wide_cpp_example_test.go).
Bounded validation **passed job74412**, source checkpoint `0db6ec6`;
[submission](submission.json). No runtime, speedup or full s17 result is claimed.
The source revision, exact commands and source/cache/evidence SHA-256 values
are recorded in [results](runs/encoding-01/results.json);
[reconciled analysis](runs/encoding-01/analysis.json).

## Reproduce and budgets

```sh
python3 -B experiments/s17-encoding/run.py --check
module load slurm/slurm/25.05
sbatch --output=experiments/s17-encoding/slurm-%j.log \
  experiments/s17-encoding/run.sbatch encoding-01
```

Use a fresh run name. One CPU-only job requests4 cores and8GiB RAM; process-group
RSS guard6GiB, Go soft limit4GiB, new retained cache2GiB and scratch4GiB.
Worker600s/controller900s/Slurm1200s deadlines remain active. This matches the
previous bounded integration budget; no full-domain s17 allocation or GPU is
requested. Go1.27.3/GCC14.2, generic C++17 optimization, no architecture-specific
flags. The controller records hostname/CPU/job inventory and profiles each
retained Go test executable. It copies source to guarded scratch, retaining
matching binaries and small reusable preprocessing in
`.cache/liftrs/s17-encoding/encoding-01/` outside Git; existing caches/profiles
are preserved. Logs/results/profiles stay under `runs/encoding-01/`. Submit then
yield, without scheduler polling or log following. Other jobs remain untouched.

## Correctness boundaries

Real Lift-RS fixtures use (s,m,t)=(3,2,2),(4,2,2),(6,2,4),(8,2,5), deterministic
seed170058 and existing reference preprocessing. Independently verify P·G at
all N field points (equals X^(N−1)+1), then load the old serialized bytes through
the width-aware API. Retain online data, G and checksum/identity metadata.
Compare every canonical value with the existing Go encoder, all systematic
positions, dense monomial membership for s≤4, packed bytes and their hashes.
Exercise workers1/4, repeated/zero/invalid messages, cap rejection, input/code
immutability, recovery, output ownership across subsequent calls/Close/GC, and
concurrent distinct workspaces versus Close. Run full shared Go-only/native
regressions, raw native API tests, Go lifetime race checks and Go ASan with
C++ UBSan. Go's race detector does not instrument native C++ accesses.

Raw Go/C checks accept either32/64-bit spectrum slots for the same small field,
and reject high64-bit values rather than truncating them. Metadata-only tests
at s16/17/18/20 check large counts, prefix discontinuities and byte accounting
without allocating a domain. Standalone optimized/ASan/UBSan checks cover12
combinations: K=1/3/7 in GF(16), two spectrum widths and workers1/4; compare
with independent carryless multiplication, naive convolution and coordinate
assembly. Check end canaries, range/alias/generator rejection and recovery.
K=1 exercises the two-transform constant schedule; K=7 exercises2L>N. Synthetic
K=3/7 inputs test arithmetic only and are not asserted to be real Lift-RS codes.
Full8-byte s17 payload loading/preprocessing/encoding remains unvalidated.

## Memory and timing boundaries

[API/accounting](../../docs/scaling-code-data.md#optional-native-encoding):
P/message/H/results stay uint32, spectra retain source width; transform arena
and scratch use uint64. With L=nextpow2(2K−1), arena=max(2L,N), principal workspace
is8arena+4K+8N bytes. Dead scratch becomes an owning uint32 result, avoiding an
extra output allocation/compaction pass: logical4N bytes retain8N backing bytes.
Retained older outputs are additional live memory. Chunk parity adds
4ceil((N−1)/4096) bytes, plus thread/metadata/allocator/runtime headroom.

The s17 resident model stays approximately427GiB principal arrays including code,
message and workspace/result backing; K remains a catalogue estimate. This is
not measured peak RAM, run admission or feasibility evidence. Packed output still
uses s bits/value; packing is outside the native schedule. No new transform,
transfer, I/O or full compaction pass is introduced. Parallel assembly uses four
relaxed atomic byte stores per value to safely share the old uint64 object
representation; its latency and overhead are unknown. Existing validated faster
paths remain defaults; this explicit experimental API is not a selected slower
memory mode. Measure representative wider kernels and the complete schedule
before forecasting full-size latency or making a memory–time choice.

Native stage clocks separate validation/buffers, Taylor, columns, rows, products
and assembly. Native total includes checks but excludes Go dispatch, allocations,
metadata setup, loading and packing. Validation command elapsed times include
profiling/test overhead; they are not unprofiled encoding benchmarks. Genuine
s17 code identities, preparation cost, admitted whole-process budgets and
representative performance calibration remain subsequent gates.


## Reconciled results and next step

Reviewed after the user's completion ping. The [analyzer](analyze.py) reconciles
438 source files against0db6ec6,24 sealed evidence files and18 cached binary/
fixture identities. All commands pass without resource-guard failures.

```sh
python3 -B experiments/s17-encoding/analyze.py encoding-01
```

The [real-code schedule/ownership checks](runs/encoding-01/encoding.log) pass
at s3/4/6/8, as do Go ASan/C++ UBSan, lifetime race, metadata-only large-count
and raw width/rejection checks. The optimized and ASan/UBSan direct C checkers
each pass12 independent synthetic arithmetic cases. Full shared Go-only and
native regression suites pass. Profiles, matching binaries and small verified
preprocessing remain retained. These are bounded correctness results, not full
s17 encoding, preprocessing admission or representative performance evidence.

**Recommended next: deduplicate the native CPU encoding sequence before further
wide preprocessing or additional modes.** Arithmetic/transforms already share
templates; the old and wide online sequences still duplicate mathematical steps.
Extract one shared sequence, leaving width-specific buffers, code access and
output assembly in small implementation modules. Keep current APIs, allocation
counts, defaults and explicit memory-mode behavior intact. Validate both widths,
existing memory modes and representative timing against this checkpoint. This
review does not start that refactor or a new job. CPU/GPU kernels and independent
Go correctness references need not be identical implementations.
