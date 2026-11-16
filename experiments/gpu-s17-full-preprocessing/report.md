# First guarded full s17 GPU preprocessing

Status: **fulls17 GPU preprocessing75832 passed and its evidence is reconciled**.
It reuses certified support discovered in the intentionally cancelled75819, then
constructs/verifies P/G/spectrum and publishes the complete cache. Resumed
scheduler elapsed10h00m12s; sampled peak hostRSS348.393GiB. This is public
preprocessing, not database encoding or storage-side protocol computation.
[Measured results](#completed-full-preprocessing-scaling-088) supersede the
historical proposal/submission status below; original numeric limits remain in
[budget.json](budget.json).

## Scripts and validation

[Controller](run.py), [Slurm script](run.sbatch), [small controller tests](test_controller.py),
[completion analyzer](analyze.py), [actual preprocessing CLI](../../cmd/pir-preprocess/README.md).
[Exact sizes](../s17-preprocessing-admission/s17-sizes.json), [size calculator](../s17-preprocessing-admission/sizes.py)
and [count validation](../s17-preprocessing-admission/runs/counts-01/analysis.json)
bind K=3876566203,N=17179869184,m2,t5. Small fresh/root/final/CUDA-free/race/ASan/
memcheck workflows are independently verified, not full-size validation.
[Local check record](local-checks.json) links exact commands/source hashes and
[controller tests](controller-tests.log). Full-runtime behavior remains unmeasured.

```sh
python3 -B experiments/gpu-s17-full-preprocessing/run.py --check
python3 -B experiments/gpu-s17-full-preprocessing/test_controller.py
# ONLY after exact budget agreement is recorded in budget.json:
sbatch --output=experiments/gpu-s17-full-preprocessing/slurm-%j.log \
  experiments/gpu-s17-full-preprocessing/run.sbatch first-01
# A later explicit --resume reuses certified completed stages, never overwrites.
# Only after a completion ping:
python3 -B experiments/gpu-s17-full-preprocessing/analyze.py first-01
```

Results/logs/profiles/source hashes/budget identity go to `runs/<name>/` and
`.cache/liftrs/gpu-s17-full-preprocessing/`. The shared cache's `s17-m2-t5/` holds
atomic support/root/code stages. Earlier profiles, binaries and failed partial
stages remain. CLI cache locking rejects concurrent cooperating writers. No agent
monitoring/polling follows submission. The controller itself emits resource
heartbeats and preserves the latest actual worker phase/counter; unknown remaining
work is never an ETA. Host support ordering/Frobenius, normalization and reused-
root scans now emit completed counts per65536-element block with30s throttling,
without a new full pass or large array. GPU progress remains at existing kernel/
wave boundaries; a long or blocked kernel cannot report subkernel completion.
Instrumentation/GC overhead is unquantified and included in profiled timings.

## Full trial resource budget

| Resource | Proposal |
|---|---|
| Placement | One eligible exclusive node; exclude gnode03, no fixed node |
| CPU | All128 node cores allocated/billed;32 active host workers |
| Host RAM | 864 GiB request;800 GiB process-group RSS cap;720 GiB Go soft limit |
| Device | 4 H200s;112 GiB owned cap/card,448 GiB aggregate |
| New retained files | 512 GiB total cache + result/log guard |
| Local scratch | 16 GiB cap; initial32 GiB free required |
| Shared filesystem | Initial576 GiB free; live64 GiB free headroom guard |
| Wall cutoffs | Worker97 h,controller98 h,Slurm100 h; original job75819 used45/46/48 h; these are not runtime estimates |
| Profiles/checkpoints | Enabled and retained; immutable completed stage reuse |

An exclusive node prevents competing batch jobs consuming unaccounted host RAM;
it costs128 allocated cores even though only32 workers are selected. Node scheduler
RAM is about907 GiB, so request864 GiB leaves about43 GiB outside the request.
Controller requires832 GiB initially available and stops below32 GiB node available
headroom. Device H200 memory measured140.4 GiB/card;112 GiB cap leaves about28 GiB
outside it. Per-array host limit128 GiB/native128 MiB/code payload161 GiB are
separate from aggregate RSS. Memory-limit/GC cost if reached is unknown. This is
an admitted-budget first trial, not a proven bound on driver/GC/filesystem usage.
100 h uses the configured100 h partition limit; QoS/association/quota still
matter. Available shared filesystem space is not authorization or quota.

The GPU array model is3·8N=384 GiB (96/card) before metadata/staging. The initial
full command explicitly selects `--vector-layout local`, the faster measured
four-card large-component candidate; it adds eight transpose passes/FFT (~2 TiB
aggregate array traffic at full domain). Small whole-workflow cases do not settle
full speedup. Default library peer layout is unchanged and either explicit choice
is supported by the budget checker. No slower memory-saving path is forced.
Eight/sixteen GPUs are authorized options to evaluate, but current implementation
cannot span these four-GPU nodes; first trial uses the implemented four-card route.

## Exact file sizes and host lifetime audit

| Public artifact | Calculated size |
|---|---|
| Support | 28.883 GiB |
| Normalized root checkpoint | 14.441 GiB |
| Canonical P/spectrum cache | 142.441 GiB |
| Generator G | 49.559 GiB |
| Total successful payload | 235.324 GiB |

These are exact byte-count formulas rounded toGiB, not measured disk consumption.
A successful fresh run writes/hashes and reads back each once; logs, profiles,
metadata, allocator/filesystem overhead and retained failed stages add storage.
The512 GiB guard allows headroom/partial stages but does not allow unlimited
failed retries. Exact byte values/source identity are in s17-sizes.json.

[Discovery](../../encoding/liftrs-shared/preprocessing_core.go) uses a128 GiB
uint64 grid,~28.883 GiB good-index list,~28.883 GiB output support and two2 GiB
bitmaps that may coexist during validation:189.765 GiB principal model. The
serialized root/P has14.441 GiB and G49.559 GiB. Final support+P+G+spectrum are
220.883 GiB, with saved normalized root sharing P backing, not another copy.
The GPU product arrays are384 GiB independently of those host totals.

This is **not peak RSS**. [Newton products](../../encoding/liftrs-shared/orbit_polynomial.go)
retain the previous G and its squared operand while receiving a new result;
[overlap-add](../../encoding/liftrs-shared/orbit_product.go) can hold the accumulator
and a local convolution output together. Root product-tree intermediates, full
64 GiB identity output, earlier dead-but-uncollected arrays, runtime/native metadata,
CUDA staging and filesystem cache add memory. The proposed large request/RSS/GC
limits leave headroom around current models; no measured guarantee is claimed.
No large duplicate payload is introduced by this controller or streaming verifier.

Controller checks the published support K against the verified count during
execution, guards RSS/storage/free headroom/deadlines and uses previously tested
bounded TERM/KILL waits. It does not reset GPUs. Startup peer-copy/transpose probe
aborts promptly on device failure before large allocations. SIGTERM retains a
profile and completed checkpoints; a failed root tree before normalization is
not a reusable root checkpoint. The original75819 source restarts an interrupted sample/root tree; replacement
recovery at safe boundaries is described below.

## Evidence boundaries

CLI checks exact dimension, support ordering/Frobenius closure, normalized root,
full polynomial identity and canonical write/readback. Large payload digests are
recorded by the CLI and immutable stage manifests. Exit does **not** silently
read235 GiB again: it hashes source/binaries/plugin, profiles, metadata and logs.
The default completion analyzer verifies those seals, stage binding/header/size
and successful worker verification events; independent post-completion large-file
rehashing must be separately submitted as a CPU-only job if required, labeled
outside preprocessing latency. Failed/unreaped workers are not successful runs.
Full output cache and timings must not be published as measurements until review.

## Full trial approval (SCALING-083)

After the explicit numeric full resource proposal and local/peer trade-off
comparison, user directs "ok let's do local then". Budget approval is recorded
in budget.json for exactly the proposed4H200112GiB/card, exclusive128cores/
32workers,864GiBhost/800GiBRSS/720GiBGo soft,512GiBretained/16GiBscratch and
45/46/48-hour worker/controller/scheduler cutoffs. Local is explicit, peer
library default remains unchanged. Existing historical preparation checks remain
intact; no-allocation checks are rerun against the approved manifest. Approval
and submission do not establish a completed s17 cache or measured peak/latency.

## First full submission

Job75819 submitted, source0ac7da7 ([submission/exact approved budget](submission.json)).
Results/logs/profile metadata: `runs/first-01/`, `slurm-75819.log`; persistent
support/root/code/profile cache: `.cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5/`.
Compilation binaries and native plugin remain in the parent cache with jobID.
No node pin, excludegnode03. sbatch returned successfully; no scheduler/log/progress
polling, waiting or result inspection after submission. Await completion ping.
No full peak/latency/cache success is claimed merely from submission.

## Longer replacement and partial-work recovery (SCALING-084)

The user explicitly requests the available larger limit and progress persistence
that does not substantially affect runtime. The updated [budget](budget.json)
keeps compute/memory/storage caps unchanged and sets97h worker /98h controller /
100h Slurm. The controller requests `SIGUSR1` one hour before the earliest internal
deadline. A safe stop has exit75/status `checkpointed`, never `passed`; subsequent
execution needs `--resume` and a new result name. Running job75819 cannot acquire
these hooks or internal cutoffs live: its isolated source and deadlines are fixed.

Replacement command adds `--checkpoint-interval 2h --resume`. Rolling snapshots
save completed samples, unmerged root-tree frontiers and completed reciprocal
prefixes, retaining two per phase. Estimated extra retained payload is about132GiB
at s17; there is no extra FFT, device transfer or saved128GiB workspace. Periodic
writes target1% cumulative overhead with adaptive intervals and bounded-write
abandonment; forced shutdown saves bypass that target. Blocking filesystem calls
and an in-flight transform may exceed the notice, so the previous completed
snapshot remains the fallback. Full checkpoint I/O cost is unknown, not assumed
free. All write/profile/log time stays within workflow timing. Read the precise
[recovery contract](../../docs/preprocessing.md#incremental-recovery) and
[small-instance verification](../preprocessing-recovery/report.md).

The replacement performs native interruption/restart preflight before the full
GPU workload. Canonical artifacts, independent final identities and validation
methodology are unchanged. `progress.bin` is excluded from controller exit
rehashing just like other large payloads: its streaming digest is sealed in the
atomic stage manifest, and it is authenticated on restart. No full-size recovery
measurement is claimed from the small tests.

## Longer replacement submitted

[Submission receipt and exact budget](submission-resume-02.json): job75832,
submitted source dd9d86a, `resume-02 --resume`,100h scheduler cutoff. It depends on
`afterany:75819`; only after successful sbatch acceptance was75819 cancelled.
One-time pre-replacement inspection found completed support, no completed root/code,
so discovery can be reused. Results: `runs/resume-02/`; log: `slurm-75832.log`;
shared persistent cache remains `.cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5/`.
Old logs/results/profile/partial files remain retained. Submission and cancellation
requests succeeded; neither terminal old state nor new state/results were polled.
Await user's completion ping before checking the new job.

## Completed full preprocessing (SCALING-088)

Job75832 completes0:0 ongnode06 with4H200,128 allocated CPU cores/32 active workers
and864GiB RAM requested. [Scheduler receipt](runs/resume-02/scheduler-completion.txt),
[sealed results/commands](runs/resume-02/results.json),
[analysis](runs/resume-02/analysis.json), [derived totals](runs/resume-02/summary.json)
and [Slurm log](slurm-75832.log). The analyzer reconciles523 source,17 original
evidence and14 retained cache identities againstdd9d86a, including exact manifest
bindings/field identifiers/counts/sizes. ExactK3876566203,N17179869184. Support
uses two random samples and is inherited from0ac7da7/job75819; its earlier
[forensic review](runs/first-01/review.json) reconciles518source/16evidence/6cache
identities and correctly retains the intentionally cancelled run's failed status.

The complete P·G identity and final cache readback pass; all controller commands
succeed without resource/deadline/shutdown guards. Payload files were streamed,
hashed and read back inside the worker. This review verifies their sealed
manifests/current sizes, not an unnecessary independent235GiB rehash. Full
resident file loading and client encoding are separate gates; no end-to-end
s17 database encoding is claimed from preprocessing.

| Measurement | Result and boundary |
|---|---|
| Earlier genuine support discovery | Runtime21min38.257s in75819;2 samples |
| Resumed worker | Runtime9h58min37.848s; support loading, preparation, persistence/readback included |
| Resumed controller | Runtime9h59min58.563s; builds/preflight/metadata included |
| Scheduler allocation | Elapsed10h00min12s |
| P/G construction, identity and final spectrum | Runtime9h16min46.126s; includes root checkpoint I/O and support validation |
| Root product | Runtime8h33min57.088s; dominant stage, includes periodic checkpoint/log costs |
| Reciprocal G | Runtime24min22.147s |
| Full polynomial identity | Runtime9min05.096s |
| Final generator spectrum | Runtime4min03.751s |
| Final complete-cache verification | Runtime24min44.344s; part of worker total |
| Peak sampled host process-group RSS |348.393GiB; predecessor189.823GiB; filesystem cache additional |
| Peak device allocation |Not periodically sampled; product-array model384GiB aggregate/96GiB per GPU, metadata extra |
| Device owned at final observation |256.009GiB aggregate after dead product buffers released; **not peak** |

This is a resumed run after a deliberate replacement, not one uninterrupted
fresh-discovery-to-cache baseline. Do not add the phase rows to the worker total:
construction/identity/spectrum is inclusive of its component phases. The old
allocation elapsed33min51s and was deliberately cancelled after support was saved;
that includes more than discovery and is not another successful complete run.
The host/device limits are admissions, not peak measurements. These profiled,
logged timings are not an uninstrumented latency baseline.

Root construction accounts for about86% of worker wall time. This locates the
largest phase, not a proven CPU-versus-GPU/kernel bottleneck: the phase contains
both host work and GPU execution. The retained CPU profile/binary can support
later investigation. No profiling-based algorithm change is made in this review.

### Retained data and recovery cost

Canonical payloads (exclude manifests/profiles/logs/recovery):

| File | Size |
|---|---|
| `support/support.bin` |28.883GiB |
| `root/root.bin` |14.441GiB |
| `code/online.bin` (P/spectrum client data) |142.441GiB |
| `code/generator.bin` |49.559GiB |
| Total |**235.324GiB** (252677224705bytes) |

Cache: `.cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5/`. It is retained
outside Git with source manifests/checksums. Clients consume `code/online.bin`
with their explicit load/admission caps; other files retain reproducibility and
verification/restart data. Preprocessing is reusable for multiple databases of
the same code parameters, not part of every cached encoding call.

Four successful periodic root snapshots took9.042/14.841/21.196/33.446s,
**78.525s total,0.219% of worker wall time** (nested in root-product duration).
Latest two snapshots are retained,23.426GiB including manifests; older complete
generations were pruned only after replacement publication as documented.
No reciprocal snapshot became due before completion. Full-run partial snapshots
were written, not used for a restart here; restart correctness remains demonstrated
by the smaller tests. No full-size partial-restart or independent snapshot
payload rehash is implied. Checkpoint overhead met the1% periodic target in this
run; this is not a guaranteed rate on other filesystems/parameters.

### Large-log review and retention

The native worker emitted3394846591bytes (3.162GiB) of progress output. It remains
at `experiments/gpu-s17-full-preprocessing/runs/resume-02/preprocessing.log`,
SHA256 `709aa3b12cd4da992da3baa61acc045ae89010b333e17dfa24d4f592f5f46790`.
Worker progress logs are explicitly ignored by Git to avoid multi-gigabyte
objects; original logs remain locally intact. Their sealed hashes and compact
selected stage records remain in results/analysis. Other small raw evidence,
metadata, profiles/cache links and scheduler records are preserved. The earlier
12MiB worker log is likewise retained/ignored with its sealed record.

The original analyzer assumed small logs and completed its initial reconciliation
using whole-file reads. The reviewer was updated to stream hashes in1MiB blocks
and select only relevant event lines without retaining/decoding millions of
kernel records. The bounded streaming rerun passes the same source/evidence/cache
checks and explicitly requires identity completion. No benchmark boundary,
producer data, original seal or timing is changed. Three small
[analysis tests](test_analysis.py) check bounded hash reads, noisy event filtering
and repeated-checkpoint durations. The shared seal helper in
`experiments/s17-arithmetic/analyze.py` now also uses bounded reads, with identical
digests. Source/Go/native algorithm behavior is unchanged.

Logging overhead is unquantified and included in the reported times. This large
volume should be reduced in a future instrumentation follow-up while preserving
phase/pass/advancing-work visibility; it is not evidence that logging was free.
Do not silently delete the logs/profiles or claim a logging speedup without a
matched measurement.

Preprocessing is now available for fulls17 client encoding. CPU/GPU encoding
validation, budgets, packed persistence and end-to-end measurements remain
separate; job76426 was not checked or changed during this review.
