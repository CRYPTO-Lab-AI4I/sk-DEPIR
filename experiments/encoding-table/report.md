# Supported encoding table gaps on HPC

Measure the 22 unrun encoding cells through s16 in the
[scaling reference](../../docs/encoding-status.md) at checkpoint61b1857. Existing
CPU/GPU encoders and verified preprocessing are reused; only opt-in experiment
harness controls change. Laptop preparation results remain labeled L, HPC results
R. No s≥17 execution, new large preprocessing or multi-GPU implementation.

Status: prepared for submission; full build/runtime checks run inside each job.
[Submission manifest and result directory](runs/fill-01/submissions.json) records
exact case names, jobs and source revision. Per-case `results.json`, logs, phase
JSON and CPU profiles will be written beneath `runs/fill-01/<case>/`.

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
