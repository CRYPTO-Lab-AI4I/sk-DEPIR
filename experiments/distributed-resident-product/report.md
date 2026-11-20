# Bounded resident GPU polynomial products

Status: implemented; CPU/race, native/C/CPP/CUDA/MPI builds and reviewer tests
pass. Bounded GPU product validation is prepared, not yet measured. Source follows
`2ec5458`/`6b6019e`; exact frozen revision is recorded by submission/controller.
[Readiness commands, hashes and warnings](readiness/checks.json).

[Complete resident FFT78396](../distributed-resident-fft/report.md) passed24 normal
and24 sanitized rank/pass results, plus24 old complete host-tile reference rows,
with clean CUDA summaries and exact Taylor/transpose coverage. This gate composes
resident products without introducing a second FFT or field recurrence. We remain
on step1 (distributed products/FFTs), not distributed discovery/fulls18 support.

## Implementation and invariants

The [shared product controller](../../encoding/liftrs-shared/distributed_product.go)
now admits three disjoint FFT operation spans and fixes forward-left, forward-
right, pointwise and inverse-left ordering for both storage backends. Existing
host reference behavior, bounded multiplication tiles and ownership checks remain;
CPU independent Horner/Lagrange/schoolbook, overlap/operation admission and race
tests pass. A focused failure check prevents inverse/result completion after a
right-transform or multiplication error.

The [resident adapter](../../encoding/liftrs-shared/distributed_storage_cuda.go)
uses **three distinct GPU slots**: left0, right1 and scratch2. Left transforms
using scratch, then right transforms reusing scratch while left's spectrum stays
resident. Pointwise replaces left; inverse reuses scratch while right retains its
spectrum. Operands are consumed. Caller must choose a padded domain larger than
the sum of degrees; otherwise evaluation/interpolation yields a product modulo
the domain vanishing polynomial. No genuine Lift-RS P/G identity is inferred.

[`RankStorage.Pointwise`](../../encoding/liftrs-cuda/rank_storage.go) uses the
existing CUDA pointwise kernel on matching views. It rejects alias/layout mismatch
and scans both operands for out-of-field values before any arithmetic write.
The scan reuses vector validation and the existing flag. No fourth device array
or full host staging is needed; right and the third slot are preserved. Launches
are bounded and report increasing word counts during long operations. Optional
ABI keeps older storage operations usable; missing pointwise support errors
explicitly. No CPU-client CUDA dependency, default or field-admission change.

## Bounded validation and common reference

The [shared product driver](../../encoding/liftrs-shared/distributed_product_gpu_test.go)
uses exactly the host reference's coefficients, schoolbook oracle and proof/log
schedule; an explicit resident mode only changes storage/launches. For both
four/eight-byte slots, s4 dimensions4/5/8 and s17 dimensions5/9/12 (N16–4096),
two operand shapes fill the domain exactly or leave a zero tail. Every output
coefficient, base-field range, constant/leading1 and tail0 are checked. Canonical
diagnostic coefficient files are32-bit even when internal slots are64-bit; this
is not packed client database encoding.

[Direct resident pointwise checks](../../encoding/liftrs-shared/distributed_resident_product_gpu_test.go)
use independent scalar extension-field products with high components, right/third
preservation and zero tails. Aliased slots, mismatched physical views and an
out-of-field right operand reject without changing either arithmetic operand.
The checked nonzero pattern count is **min(ownedWords,capacity,17)**, rather than
the old host-tile min(capacity,17): a resident owner cannot exceed its actual
shard. Every owned word including padding is verified. Per case, one direct check
produces12 rank/check rows. During each polynomial product, the test captures
right's spectrum before multiplication and checks it again after inverse.
Retained device allocation must stay stable across the two product geometries.

[Runner](run.py)/[Slurm entry](run.sbatch) first execute the existing **host-shard
product reference** (24 rank/product rows), then resident normal and all small
CUDA memcheck/leak cases (24 rows each). [Independent review](analyze.py) reuses
schoolbook output/base/tail and three-FFT route verification, including exact
Taylor metadata/source windows, paired transpose proofs, extra resident/reuse/
preservation flags, source/evidence/cache seals and dual-stream CUDA summaries.
[Reviewer/selection tests](test_review.py) reuse three existing product fixtures,
reject missing resident checks and explicitly select both product drivers.
The source/transport arithmetic is not silently changed for comparison.

## Reproduction and resources

```sh
python3 -B experiments/distributed-resident-product/test_review.py
python3 -B experiments/distributed-resident-product/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-resident-product/slurm-%j.log \
  experiments/distributed-resident-product/run.sbatch product-01
# After the user's completion ping:
python3 -B experiments/distributed-resident-product/analyze.py product-01
```

Same admitted **2 HPC nodes /2 H200s /4 CPU cores /8 GiB requested host RAM**,
one GPU,2 cores,4 GiB/node; **128 MiB owned GPU/card /1 GiB RSS guard/rank**,
no pinned buffers, **128 MiB retained cache /2 GiB scratch /32 MiB file**,
**720-second controller /15-minute scheduler**. gnode03 excluded; other eligible
nodes allowed. No resource increase or monitoring. Shared executable/wrapper/
library/cwd and remote TMPDIR preflight, explicit UCX cache control, guards and
profiles reuse the validated runner. Each MPI arm keeps240 seconds (220-second
test timeout).78396 took141.091 seconds total, normal2.371/sanitizer3.691 seconds;
old same-size products78321 took2.172/7.084 seconds normal/sanitized. This suite
uses the same small sizes/sanitizer shapes; no calibration or budget expansion is
needed. New product runtime remains unknown until completion.

Exact revision/build identities, commands/exits, resource/phase/proof logs and
CPU profiles persist in `runs/<name>/`; completed arm results save before later
phases. Sealed reusable library/binary/output/reference cache remains outside Git
at `.cache/liftrs/distributed-resident-product/<name>/`. Previous caches remain.
Logs name operand/FFT pass, operation, level, quotient/remainder and completed work.
Profiling/logging remain enabled; results are diagnostic, not unprofiled baselines.

## Timing and memory boundaries

Resident product clocks include slot reset/clear, initial operand uploads,
three complete FFTs with source snapshots/communication, pointwise validation/
math, final result readback and the diagnostic pre-pointwise right-spectrum
capture. The final right-spectrum comparison, coefficient oracle/checks/fsync,
constructor and direct pointwise tests are outside the product clock but inside
command/controller time. The host reference also starts with host coefficient
operands and returns host results. Both use the same synthetic coefficients and
route/proof logging; tiny measured comparisons will not establish fullsize speedup.

Three GPU arrays, plan,127-word device transfer tile and one bounded Taylor window
are reused. Pointwise adds **no array**, but reads both operands for canonical
validation and returns two4-byte status flags. Taylor retains two C-word host
buffers and one C-word device window, C=min(N,3*max(rows,columns)) <=1Miword.
At largest N4096/2ranks/uint64, main arrays are48 KiB/card, transfer1016 bytes,
Taylor1536 bytes/card; host Taylor payload3072 bytes/rank, plus runtime/metadata.
Diagnostic operands/oracles/results/right-spectrum snapshots are additional small
host allocations, included in RSS. Context/driver is excluded from owned bytes.

The production adapter does not download complete intermediate spectra: only the
validation driver does so for right preservation. Local Taylor sources still
stage through bounded host snapshots; repeated transfers/field scans/barriers are
explicit costs, not assumed free. Fullsize memory/time and the complete workflow
remain unknown. Next after passed products: matched larger resident/host costs,
then root/Newton/generator/discovery/persistence integration and s18 admission.
No slower memory-saving mode is selected automatically; faster validated defaults
remain, and fulls18 preprocessing is unrun.
