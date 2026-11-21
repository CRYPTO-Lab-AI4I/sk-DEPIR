# GPU-owned Taylor targets and sources

Status: **job78460 submitted**, implementation `d327890`; [receipt](submission.json).
Local CPU/race, native/HostTiled/C/CPP/CUDA/MPI builds and reviewer checks pass.
GPU correctness/performance results remain pending. [Results destination](runs/local-01/),
[job log](slurm-78460.log). Actual revision/build inputs are saved by the controller. [Readiness evidence](readiness/checks.json).

[Matched products78432](../resident-product-scaling/report.md) passed all14 arms.
At N16384/tile4096, median net product work was0.355987–0.365759 seconds with
host-stored arrays and0.616580–0.619547 seconds with resident arrays. **Both use
GPU arithmetic**. Larger tiles improve resident work only3.6–4.1% at this size;
validation capture contributes less than0.1ms with4096-word tiles. The remaining
gap motivates removing locally owned Taylor data's repeated host round trips,
not a presumed large GPU speedup or a fullsize estimate.

## Shared arithmetic and source schedule

The [Taylor scheduler](../../encoding/liftrs-shared/distributed_taylor.go) now
asks the storage adapter to prepare its local snapshot. The host adapter and old
resident snapshot path share the original read helper. The explicit new adapter
leaves local target/source entries zero in the host workspace; remote receive/
send windows, trimming, ticket waves, headers/proofs, quotient/remainder barriers
and level order are unchanged. No remote source or correctness check is omitted.

[`LocalTaylorWindow`](../../encoding/liftrs-cuda/rank_storage.cuh) reads/writes an
owned logical index directly through the existing row/interleaved GPU view. Only
remote sources fall back to the existing bounded TaylorWindow. The same
`window_taylor` kernel, sparse coefficients and `taylor_update` recurrence operate
on either accessor. Shared shape checks and launch/progress/error handling remain
in the native stage helper; no arithmetic is copied. New `TaylorLocal32/64` APIs
are optional, synchronous and read-only toward the host snapshot. Dedicated
executor ownership is required, as for the existing rank-storage APIs.

Targets are committed directly on GPU, removing the separate target copy/sync.
The existing bounded GPU window still uploads the remote snapshot layout (local
entries are zero); uploads of unused zero padding are not optimized away here.
Outgoing remote sources still require GPU-to-host transport. This change removes
local target/source downloads, not required cross-rank communication.

## Canonical values and cache invalidation

Direct local reads must not bypass field-table bounds. Each slot has a private
canonical-value proof: successful whole-slot scans or zero reset establish it;
raw upload/scatter invalidates before mutation; local copy propagates both source
and destination validity. Arithmetic invalidates before launch and restores proof
only after successful closed-field computation. Failures leave a conservative
invalid proof. New local Taylor scans the full owned slot when its proof is
invalid, then safely reuses it across windows. Existing vector/pointwise boundary
checks still run. The proof assumes the dedicated owner is not mixed with raw
legacy engine operations, and no external pointer can mutate the private arrays.

The [small GPU guard](../../encoding/liftrs-shared/distributed_local_taylor_gpu_test.go)
warms local Taylor, then injects an out-of-field word using **upload, receive
scatter and local transpose**, in both physical layouts and both slot widths.
Each subsequent call must reject before changing any owned data or the host
snapshot. The guard restores zero/original views before product checks.
It runs only in small validation arms, not benchmark arms: prewarming metadata
must not make the comparison unfair. Full-extension patterns in the existing
FFT driver independently test local reads beyond base-field product coefficients.

## Validation and matched measurements

[Runner](run.py)/[Slurm entry](run.sbatch) reuse the validated named-arm controller
and shared test/oracle code. All checks use the same admitted two-node allocation:

1. Legacy host-shard products, both widths, s4d4/5/8 ands17d5/9/12 (24rows).
2. GPU-local full FFT normal and CUDA memcheck, same6 small domains: full-extension
   input patterns, independent native FFT plus scalar Horner points, exact inverse,
   third-slot preservation and stable allocation (24rank/pass rows/arm).
3. GPU-local products normal and CUDA memcheck, same6 domains: schoolbook every
   coefficient/base range/zero tail, direct high-component pointwise/rejection
   checks,12cache-invalidation guard events/arm, right spectrum and reuse
   (24rank/product rows/arm).
4. Nine normal matched arms: s17-field synthetic N4096/16384,4096-word tile,
   host-stored GPU /old resident /local-source resident,3samples/backend, with
   order alternated (8rank/product rows/arm). These larger arms are not sanitized.

[Independent review](analyze.py) reconciles all source/evidence/cache identities,
full output bytes/hashes, exact remote Taylor/transpose coverage and paired proofs,
canonical-proof guard flags, both clean CUDA summaries and three-sample timing
summaries. [Reviewer/selection tests](test_review.py) exercise the shared fixture,
reject missing invalidation evidence and ensure guards do not warm benchmark
plans. Existing named-profile path regression remains active.

## Resources, costs and reproduction

Same admitted **2 HPC nodes /2 H200s /4 CPU cores /8 GiB requested RAM**, one GPU,
2cores/4 GiB per node; **128 MiB owned GPU/card /1 GiB RSS guard/rank**, no pinned
buffers, **128 MiB retained cache /2 GiB scratch /32 MiB file**, **720-second
controller /15-minute scheduler**. gnode03 excluded; other eligible nodes allowed.
Shared path/TMPDIR/environment preflight precedes GPU work. No resource/default
increase or agent monitoring. Each arm keeps240 seconds (220-second test timeout).

No new principal array or large staging buffer is allocated. Existing three GPU
shards, transport, Taylor window and two bounded host buffers stay. Three host
proof flags and small kernel view metadata are added. At N16384 main GPU payload
remains192 KiB/card, transfer32 KiB, Taylor3 KiB, host Taylor6 KiB/rank, before
metadata/runtime/context. Field scans after raw writes add device reads/status
checks, while local addressing may be strided for an interleaved view. Net runtime
benefit is unknown; no hidden memory-for-time default is adopted. Local mode is
explicit and existing producer/library defaults remain unchanged.

Total/net/capture and constructor clocks retain the matched experiment's exact
boundaries; input upload/result readback, source exchange/proofs, field validation
and logging remain included. The diagnostic right-spectrum read is separately
timed and subtracted per rank for net work, without removing the check. CPU
profiles persist for all arms, with reports for sample0 of all three backends.
Native stage/proof/resource logs identify operation, pass/level and completed work.
Profiling/logging costs remain in reported diagnostics; these are not GPU kernel
or unprofiled latency measurements.

78432 took230.175 seconds including14 arms/builds/profiles; all small local shapes
have prior FFT/product sanitizer cost measurements. This14-arm suite replaces
some repeats with small validation; no new larger shape is added. Projected
controller200–400 seconds fits existing720 seconds, leaving headroom. Runtime/
peaks remain unknown until measured. Completed arms save before later phases.

```sh
python3 -B experiments/resident-taylor-local/test_review.py
python3 -B experiments/resident-taylor-local/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/resident-taylor-local/slurm-%j.log \
  experiments/resident-taylor-local/run.sbatch local-01
# After the user's completion ping:
python3 -B experiments/resident-taylor-local/analyze.py local-01
```

Exact revision/build inputs/commands/resources, incremental results, logs and
profiles persist in `runs/<name>/`; sealed reusable binaries/output/reference
cache stays outside Git at `.cache/liftrs/resident-taylor-local/<name>/`. Earlier
caches are retained. Next depends on measured local-access benefit before larger
products and preprocessing integration. Fullsize/s18 admission, distributed
discovery/root/Newton/generator/persistence remain separate. Stillstep1.
