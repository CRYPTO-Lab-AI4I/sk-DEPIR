# Matched resident/host product scaling and transfer tiles

Status: prepared; local build/reviewer checks pass. No new GPU measurements yet.
Source follows `a96d7f6`/`845ce07`; receipt/raw results identify the frozen producer.
[Readiness evidence](readiness/checks.json). This follows
[passed resident products78399](../distributed-resident-product/report.md).

78399's largest4096-point products took0.2347–0.2367 seconds resident, versus
0.1080–0.1111 seconds host reference, single profiled samples. The resident clock
also includes initial upload, final readback and a right-spectrum validation
capture. Its CPU profile assigns78.69% of samples to foreign calls; transfer
wrappers appear prominently, but that category includes GPU/MPI operations and
waits and is not a GPU kernel profile. Neither a fullsize speedup nor the cause
of the difference follows from those samples. The validated faster production
defaults remain unchanged.

## Exact comparison

[Runner](run.py) uses one admitted allocation and the same shared driver,
coefficients, dense schoolbook oracle, FFT/product order, field checks and paired
proofs for both backends. A dedicated `TestMatchedGPUProducts` accepts only
s17 field, dimensions12/14 (**N4096/16384**), host/resident, and127/4096-word
transport tiles. This is a synthetic polynomial test using the s17 field,
**not full s17 preprocessing or encoding**. No field parameter admission changes.

Three samples of each backend/tile pair run at both sizes and both previously
validated product shapes (exact domain and zero tail). Backend order alternates
by sample. There are12 normal MPI arms,8 rank/product rows each, plus2 sanitizer
arms at N4096/tile4096 only,4 rows each. New N16384 and repeated127-word arms are
not sanitized in this experiment; earlier small127-word sanitizer coverage is
retained. Each arm has a fresh process/owner per dimension; shape0 includes lazy
plan/window setup, shape1 reuses it. These are three repeated samples, not a
steady-state hardware-throughput claim. Constructor time is measured separately
as GPU owner/library/table initialization; Go field-table/oracle construction is
outside that setup clock.

The existing test driver takes explicit tile size. Initial resident uploads,
logical readback and diagnostic captures now use that bound; Taylor/transpose
transport uses the same bound. Default tests retain127-word tiling. The shared
controller admits named arms rather than rebuilding/submitting each comparison
separately. Frozen source/native build and remote path preflight happen once;
completed arms/profiles persist before later work.

Every output coefficient/base-field range/zero tail remains checked against
schoolbook multiplication; direct pointwise/alias/layout/range/right/scratch and
post-inverse right-spectrum preservation checks remain. No validation is removed
for speed. [Reviewer](analyze.py) checks exact case/tile-dependent packet coverage,
all Taylor routes/proofs, source/evidence/cache seals, resident/reuse claims and
CUDA summaries. [Reviewer/launcher tests](test_review.py) verify subset/merged-tile
coverage, arm selection, fresh owner sample plan and named profile paths. The
original fixture tests also rerun. Results summarize each geometry/backend/tile
with three samples and median/range; no fullsize extrapolation.

## Timing boundaries

`seconds` retains the old maximum-rank product boundary: reset/clear, resident
input upload, all three FFTs/pointwise, final readback, proof/logging and diagnostic
right-spectrum capture. `validation_capture_seconds` separately measures the
pre-pointwise right-spectrum capture. `work_seconds` is the maximum of each
rank's own total minus its capture, not the difference of two unrelated maxima.
Per-rank local values are retained and independently checked. Host capture is0.
This separates a validation readback without disabling any check. Final right
comparison, schoolbook oracle, output checks/fsync, constructor/direct pointwise
checks and profile reports remain outside product clocks, inside complete arm/
controller time. Profiling/logging and field/source validation remain enabled;
net work is not GPU kernel time or an unprofiled latency baseline.

CPU profiles are retained for every arm; text summaries are generated for the
first sample of all four backend/tile pairs. Named profile paths match the actual
rank wrapper output. Logs identify arm/backend/tile, dimension, product shape,
operation/FFT pass, Taylor level and increasing completed work. Resource guards,
deadlines and flushed progress remain job-side, not agent monitoring.

## Resources and projected costs

Same admitted **2 HPC nodes /2 H200s /4 CPU cores /8 GiB requested RAM**, one GPU,
2 cores and4 GiB/node; **128 MiB owned GPU/card /1 GiB RSS guard/rank**, no pinned
buffers, **128 MiB retained cache /2 GiB scratch /32 MiB file**, **720-second
controller /15-minute scheduler**. gnode03 excluded; other eligible nodes allowed.
[Slurm entry](run.sbatch). No allocation/admission increase or default change.

| Payload at eight-byte slots |127-word tile |4096-word tile |
|---|---|---|
| Host transpose buffer per rank |1016 bytes |32768 bytes (32 KiB) |
| Device transfer buffer per card |1016 bytes |32768 bytes (32 KiB) |
| Change per buffer |Baseline |+31752 bytes, about31 KiB |

Larger tiles are expected to reduce transfer/pack/proof call count; runtime
benefit is unknown. At N16384, three resident arrays total192 KiB/card, plus
32 KiB transfer,3 KiB Taylor window, metadata/tables/context. Two host Taylor
buffers total6 KiB/rank. The old conservative six-host-shard model is384 KiB/rank.
These are payload models, not RSS or whole-workflow savings. Prior measured
RSS414.418 MiB and owned GPU1.618 MiB leave substantial room below unchanged
1 GiB/128 MiB guards. Runtime/driver/context, oracle and diagnostic arrays are
additional; actual peaks will be reported.

78399 took153.300 seconds total; largest normal product0.237 seconds and
sanitized1.325 seconds. A fourfold domain increase suggests roughly4–6 times
normal product work/transfer, while dense schoolbook oracle work can grow16fold.
These are budgeting projections, not runtime estimates for full preprocessing.
N4096-only sanitizer arms retain measured shapes.12 normal arms plus2 small
sanitizer arms and one build/profile setup are projected around250–400 seconds,
with720-second controller and240-second individual arm limits unchanged. Each
arm saves before the next; a late failure retains prior evidence. Cache/output
payload is projected well below128 MiB; no large intermediate disk matrix is used.

```sh
python3 -B experiments/resident-product-scaling/test_review.py
python3 -B experiments/resident-product-scaling/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/resident-product-scaling/slurm-%j.log \
  experiments/resident-product-scaling/run.sbatch matched-01
# After the user's completion ping:
python3 -B experiments/resident-product-scaling/analyze.py matched-01
```

Raw results, exact revision/commands/resources, logs/profiles and incremental
composition are in `runs/<name>/`; sealed reusable binaries/output/reference cache
is `.cache/liftrs/resident-product-scaling/<name>/`, outside Git. Earlier evidence/
preprocessing caches remain untouched. Next depends on measured batching/setup/
communication cost; then larger products and preprocessing integration. Fulls18
admission, distributed discovery/root/Newton/generator/persistence remain separate.
