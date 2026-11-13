# Supported encoding table gaps on HPC

Measure the 22 unrun encoding cells through s16 in the
[scaling reference](../../docs/encoding-status.md) at checkpoint61b1857. Existing
CPU/GPU encoders and verified preprocessing are reused; only opt-in experiment
harness controls change. Laptop preparation results remain labeled L, HPC results
R. No s≥17 execution, new large preprocessing or multi-GPU implementation.

Status: submitted as CPU jobs74194–74203 and GPU jobs74204–74215,
source checkpoint366f474. Full build/runtime checks run inside each job;
no scheduler/log/result monitoring has been performed.
[Submission manifest and result directory](runs/fill-01/submissions.json) records
exact case names, jobs and source revision. Per-case `results.json`, logs, phase
JSON and CPU profiles will be written beneath `runs/fill-01/<case>/`.

## Evidence publication checkpoint

The19 fill-01 cases previously confirmed completed with Slurm0:0 and saved
`passed` summaries are archived with their logs, phase JSON and profiles. This
publication does not perform full digest/cross-mode reconciliation or introduce
new performance claims. Completed CPU cases: compact s10; packed/disk s14;
reuse/compact/file/packed s15. All12 original GPU cases completed. Remaining
original CPU cases were cancelled/replaced as described below. Active fill-02
outputs and partial original files remain local for completion/reconciliation;
this checkpoint does not inspect their status. Inputs/large outputs remain in
the documented durable cache directories outside Git.

## Reproduction and parameters

[Controller](run.py), [Slurm script](run.sbatch), [batch submission](submit.py),
[analyzer](analyze.py).
The controller's CASES is the explicit 10 CPU /12 GPU case matrix at
s10/14/15/16. Sizes and m2,t5 are independent of backend; switches select slower
memory modes explicitly. The implementation checkpoint is recorded in the
submission manifest and each result, with hashes of the isolated source copy.

```sh
python3 -B experiments/encoding-table/run.py --check
module load slurm/slurm/25.05
# Submit the full bounded case matrix without monitoring:
python3 -B experiments/encoding-table/submit.py fill-01
# Or submit one case with appropriate memory and optional GPU.
sbatch --mem=32G --gres=gpu:h200:1 \
  --output=experiments/encoding-table/slurm-%j.log \
  experiments/encoding-table/run.sbatch fill-01 gpu-compact-s15
# After the user reports completion, on a compute job if reconciliation exceeds30s:
python3 -B experiments/encoding-table/analyze.py fill-01
```

Submit CPU cases in one `afterany` chain and GPU cases in another, so at most one
of each runs simultaneously; one failure must not prevent unrelated cells from
being measured. No dependent result analysis occurs before the user's ping.
Each job builds and checks its experiment controls before its workflow; a failed
build/check prevents that case's encoding. Runtime checks have not yet passed.

## Budgets and trade-offs

One node,32CPU per case, one H200 only for GPU cases. Host requests16GiB at s10,
32GiB at s14/15,128GiB at s16; sampled process RSS guards12/24/96GiB respectively.
Cgroup guards14/30/112GiB include charged cache; sampled file residency and
RSS-plus-tracked-cache are also recorded, without treating their sum as exact
physical usage. CUDA-owned cap40GiB (host-backed512MiB); driver/context extra,
with free-memory headroom checked. GCC14.2, Go1.27.3, CUDA12.9.1 for GPU builds;
actual host/device/compiler/allocation details are retained in logs.

Every case has a two-hour Slurm deadline,110-minute controller deadline and
90-minute command/encoding deadline. New scratch is capped at16GiB per case.
New retained cache/output is16GiB per case except **s16 CPU disk64GiB**. The user
conditionally approved this increase when the prior16GiB cap is our own guard:
it is recorded as such in the prior disk experiment, not a scheduler limit.
Filesystem inspection found about13TiB free; `quota -s` reported no limit.
Controller free-space checks remain active; this does not reserve that space.
About40GiB is expected for s16 disk work files plus packed output, excluding builds.
The whole batch can retain up to400GiB by per-case caps (typical use should be much
less); retained code input files are shared symlinks and never deleted or copied.

s16 packed CPU removes the owning canonical output; its runtime is unmeasured.
Host-backed GPU trades device memory for full host arrays and repeated transfers;
disk-backed CPU trades private arrays for files, cache and substantial reads/writes.
Full s16 backed-mode latency is unknown: bounded attempts may time out rather than
produce a complete entry. Ordinary faster defaults are unchanged. Profiles and
cache artifacts survive errors/deadlines where termination permits; no automatic
cleanup deletes useful evidence.

## Extended CPU retries

The user requested relaunch of the three remaining CPU jobs with adequate limits,
then selected **packed in parallel with a sequential disk queue**. Cancellation
was requested for old jobs74201/74202/74203; their logs, partial workspace and
results stay in fill-01. New attempts use fresh fill-02 output/cache paths;
[retry manifest](runs/fill-02/submissions.json) records replacement identities.
Submitted at source checkpointfb9bade: s16 packed74251, s15 disk74252, s16 disk74253.
No post-submission scheduler/log checks or result analysis performed.

| Retry case | Slurm maximum | Controller maximum | Go test/worker maximum |
|---|---|---|---|
| s16 CPU packed | 4h | 3h45m | 3h30m |
| s15 CPU disk | 12h | 11h45m | 11h30m |
| s16 CPU disk | 48h | 47h45m | 47h30m |

These are explicit retry ceilings, not latency forecasts. The initial s14 disk
case measured758/684s per encoding, while s15's first encoding was still active
at about80minutes. Scaling crosses file-cache regimes; extrapolation from the
small case is unreliable. A larger deadline does not establish completion.
Partition gpuq reports a100-hour maximum, so these requests fit. Memory, cores,
profiling, input checks and storage caps are unchanged (64GiB only for s16 disk).
Original attempts have no mid-encoding checkpoint; retries restart encoding.

`--extended-limits` is accepted only for these three CPU cases. The submitter
passes the matching Slurm limit, controller limit and Go test timeout from one
policy; historical two-hour defaults remain unchanged. Dry-run and allocation-free
policy assertions verify ordering and unchanged non-time resource caps.

```sh
module load slurm/slurm/25.05
python3 -B experiments/encoding-table/submit.py fill-02 \
  --extended-limits --scheduling split --after-job 74201
```

The old-job dependency only prevents overlap during cancellation cleanup; retries
are independent of the old results. The packed job and first disk job then run
independently; the second disk job waits via afterany. No scheduler/log monitoring
follows submission. Reconcile completed retry cases after the user's next ping.

## Inputs, correctness and measurement boundaries

Trusted online digests for s10/14/15/16 are pinned in the controller; metadata,
code size and m2,t5 dimensions are checked and loaders verify the full code digest.
Reuse the caches documented by [s10 products](../gpu-products/report.md),
[s14 preparation](../parallel-s14/report.md),
[s15 preparation](../parallel-s15/report.md) and
[s16 preparation](../gpu-preparation/report.md). Their paths are in REFERENCES;
regeneration/provenance stays with those reports. No discovery/preparation repeats.

The existing [whole-word verification harness](../../encoding/liftrs-shared/s15_firstuse_probe_test.go)
checks every systematic value,16 independent curves, canonical hashes and stored
packed readbacks. The existing
[packed verification harness](../../encoding/liftrs-shared/gpu_packed_cached_test.go)
now also selects CPU file-packed output while reusing the same independent
verification. It maps the packed file during verification, outside encoding.
Seed71 is fixed. Run two encodings through s15 and one at s16 to bound expensive
backed modes; repeat hashes must agree, and the analyzer checks across measured
modes at each s. Retain one packed output per case. Setup, load, encoding,
packing/storage and verification keep separate profiled phase timings.
CPU file/disk native worker count remains16 in its existing disk harness; other
CPU modes use32. All cases receive32 CPU cores; report that distinction.

Only the experiment configuration accepts PIR_TABLE_S=10/14/15/16; existing
public parameter guards and historical test defaults remain. Small Go cases and
native sanitizer coverage remain in the earlier mode reports; this experiment
adds scale coverage rather than replacing that evidence. The packed CPU workflow
uses both build tags to share verification; ordinary CPU library builds still
require no CUDA. CPU jobs do not load or build a CUDA library.

No new measured numbers are claimed here before reconciliation. After completion,
link the analysis and update the reference with HPC labels, memory boundaries,
thread counts and stored-output versus encoding-only latency. Failed or timed-out
cases remain unmeasured with their failure/deadline recorded.
