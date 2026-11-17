# Encoding scaling history

Historical milestone snapshots and superseded forecasts, moved from the active
reference on2026-11-17. These sections record what was known at each stage; their
pending/not-run language is **not current status**. Consult the
[maintained scaling reference](encoding-status.md) for current measurements:
s17 preprocessing75832 passed, resident CPU76645 passed, and resident GPU76646
was attempted but stopped at its host-memory guard. Job76704 is separate bounded
prototype validation; no completion is asserted here.

The sections below preserve the sequence of implementation, submissions and
reviews. Pending/unmeasured statements describe the stage at that time and are
superseded by the [current results and mode tables](encoding-status.md). The latest full-size
review is [SCALING-093](#full-s17-cpu-result-and-gpu-attempt-scaling-093).

## Bounded preprocessing components (HPC, not full table rows)

[GPU review](../experiments/gpu-s17-preprocessing/report.md#completed-corrected-peer-retry-scaling-073)
measures s17-width dense transforms at only2^22 values (32 MiB), on gnode02 H200
NVL cards with NV6 peer connectivity. Warm/profiled/logged whole-call timings,
one sample, include transfers. Full s17 preparation remains **Not run**.

| GPUs | Forward runtime | Inverse runtime | Owned GPU memory (aggregate) | Validation |
|---|---|---|---|---|
| 1 | 184.709 ms | 158.943 ms | 65.57 MiB | Complete bounded job74453 passes; comparison measured within74466 |
| 2 | 216.765 ms | 190.268 ms | 67.14 MiB | Corrected bounded job74466 passes, zero memcheck errors/leaks |

Two GPUs are17.4%/19.7% slower on this small component. Extra cards are not ruled
out for larger workloads. Driver/context memory is additional; sampled command
host RSS below1 GiB includes builds and is not isolated preparation memory.

§ [Completed22-case HPC batch](../experiments/encoding-table/report.md#completed-all-case-review-scaling-073)
uses32 allocated CPU cores and retained preprocessing, with original native
worker settings captured in the run logs. The jobs are profiled, at different
revisions/nodes from older baselines; these rows are not controlled speedups.
One encoding sample at s16, two at smaller sizes. New memory entries are sampled
**whole-workflow** process RSS (including verification/readback), not isolated
encoding peaks. File residency is additional and can overlap RSS; it is not an
exact physical-RAM sum. All22 cases complete on the HPC system under their
recorded budgets. These batch measurements stop at s16; the separate s17
resident CPU pass and GPU guard failure are recorded in the current tables.

The [local-vector comparison](../experiments/gpu-s17-preprocessing/report.md#completed-local-vector-validation-scaling-075)
passes bounded GPU validation74468, including zero memcheck errors/leaks. Same
s17-width32 MiB dense tile, one warm/profiled/logged whole-call sample per case:

| Layout | GPUs | Forward runtime | Inverse runtime | Owned GPU memory (aggregate) |
|---|---|---|---|---|
| Peer | 1 | 197.515 ms | 170.759 ms | 65.57 MiB |
| Peer | 2 | 233.538 ms | 197.709 ms | 67.14 MiB |
| Local | 1 | 194.450 ms | 183.013 ms | 65.57 MiB |
| Local | 2 | 189.407 ms | 178.511 ms | 67.14 MiB |

Local reduces two-card time18.9%/9.7% versus peer, but two local cards gain only
2.6%/2.5% over one local card. Single-card local inverse is7.2% slower. These
fixed-order component samples are not fulls17 forecasts or a settled layout
choice. Default peer and all capacity cells stay unchanged. Same2N transform/3N
product payload; extra8 transpose passes add16N word reads/writes (~2 TiB at
s17), with full-size runtime/link traffic unknown. A256 MiB repeated1/2/4GPU
comparison now has explicit4GPU/4 GiB resource approval; submission and results
are recorded in the experiment report. No capacity or full-size timing changes.

## Repeated GPU preprocessing calibration (HPC components)

[Job74635](../experiments/gpu-s17-preprocessing/report.md#completed-four-gpu-calibration-scaling-077)
passes complete bounded validation on four H200s (gnode02, NV6 peers), with zero
CUDA errors/leaks, full tile hash/inverse agreement and reconciled evidence.
256 MiB tiles (2^25 s17-width values),3 fixed-order warm/profiled/logged samples
per case, whole-call wall time including transfers. Not fulls17 preparation.

| Layout | GPUs | Forward median (range) | Inverse median (range) | Owned GPU memory, aggregate |
|---|---|---|---|---|
| peer | 1 | 0.865674 s (0.864669–0.865862 s) | 0.754588 s (0.754539–0.755842 s) | 513.57 MiB |
| peer | 2 | 0.995857 s (0.995585–0.996306 s) | 0.884177 s (0.883658–0.884894 s) | 515.14 MiB |
| peer | 4 | 1.088009 s (1.087734–1.089616 s) | 0.978584 s (0.976949–0.978659 s) | 518.28 MiB |
| local | 1 | 0.855847 s (0.855641–0.855908 s) | 0.830580 s (0.830201–0.831150 s) | 513.57 MiB |
| local | 2 | 0.915740 s (0.915194–0.915805 s) | 0.890322 s (0.889293–0.890732 s) | 515.14 MiB |
| local | 4 | 0.778385 s (0.778306–0.778699 s) | 0.753873 s (0.751693–0.754376 s) | 518.28 MiB |

Four-card local is28.5%/23.0% faster forward/inverse than four-card peer, but versus
one-card peer gains10.1% forward and essentially nothing inverse. Sum of reported
medians improves5.4%, not end-to-end preprocessing. Sampled whole GPU-test process-
group RSS peaks1.615 GiB; device context/driver overhead is additional. Current
small vector grids may limit scaling (source geometry, not measured occupancy).
Full-domain latency and peak host memory remain unknown. Default/capacity rows
are unchanged; finish persistent full-workflow admission before any fulls17 job.

## Persistent preprocessing readiness

Historical admission checkpoint, superseded by [completed full preprocessing](#full-s17-preprocessing-verified-scaling-088).

The [server CLI and complete-path validation](../experiments/gpu-preprocessing-workflow/report.md)
now implement atomic support/root/code cache stages and verified restarts, retaining
profiles/provenance/logs. Local native/Go-oracle and failure-path checks pass;
complete GPU executable validation is pending. No s17 full-run table cell changes.
Projected retained s17 public preprocessing files total~235 GiB, with additional
profiles, failed stages and storage headroom; fresh writes/readback add unmeasured
I/O latency. Full host peak and budget admission remain outstanding.

| s17 preprocessing route | Modeled product arrays per GPU | Implementation status | Full latency |
|---|---|---|---|
| 4 GPUs, one node | 96 GiB | Same-node kernels boundedly validated; persistent workflow validation pending | Unknown |
| 8 GPUs, at least two nodes here | 48 GiB | User-authorized option; cross-node execution not implemented | Unknown |
| 16 GPUs, at least four nodes here | 24 GiB | User-authorized option; cross-node execution not implemented | Unknown |

These divide the384 GiB array model evenly, excluding runtime/staging/headroom,
communication and possible host replication. They are not speedup estimates or
full-resource admission; CPU/GPU client encoding support stays unchanged.

Persistent workflow75194 passes the complete one-GPU s8 path and CPU/race/ASan/vet,
but four-GPU discovery stalls on gnode03 with a later GPU-reset report. It is
failed/incomplete, distinct from prior successful four-card kernels on gnode02.
No fulls17 cell changes. [Forensic report and unchanged-budget retry](../experiments/gpu-preprocessing-workflow/report.md#job75194-failure-and-bounded-retry-scaling-079)
retain unknown fault cause, partial caches and bounded-shutdown/preflight changes.

Persistent complete-path validation75342 passes ongnode04 in3min42s including
builds/tests. Peer/local1/4GPU fresh/root/completed-cache cases, CUDA-free
readback, independent Go bytes, peer preflight, race/ASan/vet and zero CUDA errors/
leaks pass;509/37/73 source/evidence/cache identities reconcile. This completes
the small persistent-runner gate, not any fulls17 measurement or capacity claim.
Full support/cardinality, concurrent RAM/device/storage/time budget and guarded
launch admission remain next. [Report](../experiments/gpu-preprocessing-workflow/report.md#completed-persistent-workflow-validation-scaling-080).

[Exact count-only admission](../experiments/s17-preprocessing-admission/report.md)
is implemented with local references passing; full s17 GPU count remains pending.
No full-preparation cell changes. Format-derived successful retained files are
12K+12N+61 bytes at s17 (≤288 GiB+61 forK≤N/2), before logs/profiles/partial stages
and headroom. New count-only mode avoids the large host list/grid; its~2.5 GiB
modeled GPU enumeration payload is not a full preprocessing memory reduction.
Final RAM/storage/runtime admission and cross-node support remain separate.

[Current configuration/free-space snapshot](../experiments/s17-preprocessing-admission/resource-snapshot.json) supports the admission draft; configuration, filesystem free space and blank QoS fields are not quota or resource approval.

## Exact s17 count and full admission readiness

Historical admission checkpoint, superseded by [completed full preprocessing](#full-s17-preprocessing-verified-scaling-088).

[Count75810](../experiments/s17-preprocessing-admission/report.md#completed-count-admission-scaling-082)
passes all checks, with exact K3876566203,N17179869184 and no full host grid/list.
Count-only executable6.276s includes setup/profiling/metadata; it is not discovery
or preparation latency.513source/41evidence/85cache identities reconcile, zero
CUDA errors/leaks. No s17 full-preparation/encoding row becomes measured.

Successful public preprocessing files calculate to235.324GiB: support28.883,
root14.441, canonical P/spectrum142.441 and G49.559GiB. Discovery principal host
model189.765GiB (two index lists/grid/two bitmaps); final retained host220.883GiB,
with product/Newton/identity/GC/runtime/driver/cache overhead additional and peak
unknown. Existing~220.84GiB/~235GiB estimates are superseded by these exact-K
calculations, not measured peaks.

The [full controller](../experiments/gpu-s17-full-preprocessing/report.md) uses
one exclusive128-core node/32workers,864GiB host request,4H200112GiB owned/card,
512GiB retained and16GiB scratch. Eight/sixteen GPU options require cross-node
support; no linear speedup is assumed.

The first full s17 local-layout trial is now approved under the concrete
[experiment budget](../experiments/gpu-s17-full-preprocessing/budget.json):
4H200112GiBowned/card, exclusive128cores/32workers,864GiBhost request,
512GiBretained/16GiBscratch,48h scheduler cutoff. Submission is recorded in its
report; table rows stay unmeasured until completion and evidence review.

Full s17 public preprocessing is now submitted as job75819 (local layout), with
[recorded approved limits](../experiments/gpu-s17-full-preprocessing/submission.json).
Status is **submitted, not yet verified**; all full s17 runtime/peak-memory entries
remain unmeasured until completion and reconciliation. No database encoding job.

The approved replacement extends the scheduler cutoff to100h (97h worker/98h
controller), with optional2h/adaptive recovery writes targeting about1% overhead.
Two snapshots per discovery/root/reciprocal phase project about132GiB additional
retained payload; compute/memory/storage limits remain unchanged. No FFT workspace
is saved, and an interrupted transform is repeated. These are recovery/capacity
changes, **not measured s17 preprocessing performance**. See the
[recovery contract](preprocessing.md#incremental-recovery) and full-run report.

Replacement **75832 submitted** with100h scheduler cutoff and incremental recovery;
original75819 received a cancellation request only after the replacement was
accepted. Existing completed support is reusable. Results/logs and exact budget:
[submission receipt](../experiments/gpu-s17-full-preprocessing/submission-resume-02.json).
No new-job monitoring or full result review; s17 remains unmeasured.

## Matched CPU refactor comparison at s14

[Job76055 passes](../experiments/cpu-encoding-regression/report.md#completed-comparison-scaling-086)
on the HPC hostgnode02,32 CPU cores, no GPU: four balanced pairs, identical
verified code/database, full canonical output hashes and separate profiles.
Both versions log progress, with the old671bf5e baseline receiving an explicit
logging-only patch; candidate revisione07f308. These matched measurements are
separate from older table entries with different revisions/methodology.

| Operation | Pre-refactor median | Refactored median | Difference |
|---|---|---|---|
| Resident encoding | Runtime60.346 seconds | Runtime61.119 seconds | +1.28% |
| Resident cache loading | Runtime1.894 seconds | Runtime2.540 seconds | +34.1%;0.646 seconds |
| File-backed encoder opening | Runtime1.501 seconds | Runtime2.042 seconds | +36.1%;0.541 seconds |

Encoding-process peak sampled RSS is3.593–3.595GiB in both implementations,
with filesystem cache additional. Workspace allocation varies2.6–282ms and the
first old code load is an I/O outlier; all samples/ranges remain in the report.
The workspace median ratio is unstable, not evidence of a fixed38x penalty.
The shared decoder is the setup profiling hotspot. This small larger-size
encoding difference does not block GPU client integration; setup optimization
remains a follow-up. No s17 performance or memory-capacity update follows.

## Resident GPU client integration (SCALING-087)

[Source/API](gpu-encoding.md) now implements resident code-data encoding through
s17 with explicit1/2/4/8 same-node GPUs, common CPU/GPU cyclic order and shared
width-selected GPU transforms. CPU defaults remain CUDA-free; legacy s16 GPU
modes are unchanged. [Bounded validation](../experiments/gpu-code-encoding/report.md)
is pending. Genuine s17 preprocessing/encoding measurements remain unknown;
new code availability does not change any measured table cell.

For exacts17K, principal resident code/work/parity payload projects398.457GiB
aggregate and104.004GiB maximum on one of four GPUs, before metadata/context/
headroom. Keeping source/message/canonical output projects220.883GiB host arrays;
output is64GiB uint32. This refines the earlier~398GiB resident model, not measured
peaks or full admission. The new path provides canonical words; s17 direct packed,
streamed-code, compact and host-backed modes remain unimplemented. Compare full
setup/transfers/peer routing/output latency after correctness and budget gates.

The bounded resident-client integration suite is submitted as **job76426**,
sourcebf2bffc, within the prior4GPU/4GiB test budget. [Receipt](../experiments/gpu-code-encoding/submission.json)
links results/logs. No monitoring or fulls17 encoding submission; runtime status
remains pending until the completion ping and evidence reconciliation.

## Full s17 preprocessing verified (SCALING-088)

**75832 passes**:4H200,348.393GiB peak sampled host RSS; worker9h58min38s,
including cache I/O/readback and reused support. Earlier discovery21min38s,
2samples; complete P·G identity and canonical-file readback pass. Final reusable
payload235.324GiB; clients need the142.441GiB `code/online.bin`. Four periodic
root saves add78.525s (0.219% worker wall), with23.426GiB latest snapshots retained.
Full device peak was not sampled;384GiB is the product-array model, while the
final post-release observed owned total is256.009GiB. Do not conflate these.

[Report/summary](../experiments/gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088)
reconciles523source/17evidence/14cache identities. This is a resumed run, not an
uninterrupted fresh baseline; profiled/logged root construction accounts for
about86% of worker wall. Raw3.162GiB progress remains locally retained outside Git,
with sealed digest/selected events; streaming review passes without a redundant
large-payload rehash. Encoding rows stay unmeasured at s17. The preprocessing
prerequisite is complete; full CPU baseline/GPU validation, output and resource
admission remain separate. GPU validation76426 was not inspected here.

## GPU client bounded validation verified (SCALING-089)

**76426 passes ongnode10,4min28s scheduler elapsed.** Genuine small-code client
checks cover1/2/4GPUs, peer/local, independent Go/CPU equality and ownership/error/
reuse/lease behavior. Sixty direct native32/64-bit cases and12sanitized cases pass;
actuals17 short kernels and34-bit canonical positions pass. Zero CUDA errors/leaks.
[Report/provenance](../experiments/gpu-code-encoding/report.md#completed-validation-scaling-089)
reconciles544source (three unrelated exact review-script snapshots explicitly
qualified),47evidence/76cache identities. GPU integration correctness gate is
complete; fulls17 encoding/time/peak/packed-output rows remain unmeasured. No
large-client job admission or runtime estimate follows from these small tests.

## Full s17 encoding launch plan (SCALING-090)

[CPU/GPU full-run recipe](../experiments/s17-full-encoding/report.md) uses the
verified75832 code and common seed71 database, with immutable committed sources.
CPU32workers/768GiB RAM request/noGPU; GPU32workers/384GiB host and4H200112GiB
owned cap/card. Each has64GiB new output/log/profile storage,16GiB scratch and
48hSlurm limit. Both select fast resident layouts, no implicit memory-time trade.
One full encode each, all systematic symbols,16independent curves, full canonical
hash and34GiB tight packed output with full disk digest/readback. All timings are
profiled and separate load/setup/encode/check/output boundaries. No full result
or new table timing is claimed until completion and evidence reconciliation.

Fulls17 encoding jobs submitted: **CPU76645 /GPU76646**, independent, both frozen
to revisioned0c026. [Exact submission/caps](../experiments/s17-full-encoding/submission.json).
Results `experiments/s17-full-encoding/runs/first-01/{cpu,gpu}/`; logs
`cpu-76645.log` /`gpu-76646.log` in that experiment directory. No monitoring or
measurement claims after submission; await completion ping. Existing caches remain.

## Initial s18 preprocessing admission study (SCALING-091)

[Exact count/memory audit](../experiments/s18-preprocessing-admission/report.md)
is prepared on oneGPU: sum existing good weight classes with2^ones multiplicity,
without an8GiB bitmap or512GiB grid. Validate independentGo small counts and
knowns16/s17 dimensions first. No fulls18 support/FFT/API admission is enabled.
At exactK, compute support/root/generator/cache formulas; current array model
needs1536GiB device products,512GiB host grid and8K+768GiB final host arrays,
with transient/headroom additional. At112GiB/card and current power-of-two layout,
16GPUs would be needed across at least4nodes; cross-node/host distribution is
not implemented. Initial count job stays within existing oneGPU small ceilings,
not authorization for full distribution or a slower memory mode. No table entry
becomes measured until the count completes and evidence is reviewed.

s18 dimension/audit **76647 submitted**, frozen sourcebb524d3,1GPU/small ceilings.
[Receipt/results locations](../experiments/s18-preprocessing-admission/submission.json).
No monitoring; await completion ping. Fulls18 preprocessing is not submitted;
CPU76645/GPU76646 unchanged/uninspected.

## s18 exact dimension verified (SCALING-092)

76647passes ongnode05,59s total suite,560source/29evidence/4cache identities
reconciled; Go small references/knowns16/s17 counts agree, zero CUDA errors/leaks.
ExactK16300262095,N68719476736. The37.4ms GPU weighted sum excludes initialization;
it produces no support/FFT/cache and is not a preprocessing latency estimate.
[Report/formulas](../experiments/s18-preprocessing-admission/report.md#completed-exact-count-and-memory-audit-scaling-092).

Current-layout canonical files950.170GiB (code572.723GiB), final host arrays
889.446GiB and discovery770.893GiB before temporary/runtime/cache/headroom. No
safe single-node admission follows. Device products1536GiB; power-of-two route
at112GiB/card projects16cards/fournodes, with cross-node/host distribution absent.
Two recovery generations per phase model another528.001GiB before overhead;
prior512GiB retained cap is insufficient. Propose distributed resident versus
explicit lower-memory tradeoffs before implementation/large resource admission.
Fulls18 runtime/peaks remain unknown; encoding76645/76646 not inspected.

## Full s17 CPU result and GPU attempt (SCALING-093)

[Completed review](../experiments/s17-full-encoding/report.md#completed-first-attempts-scaling-093):
**CPU76645 passes** ongnode02/32cores,1h45min11.087s encoding,427.845GiB peak
sampled processRSS; full142.441GiBcode loading29min12.498s. Packed write/hash/fsync
2min22.346s plus readback1min50.054s;34GiB output is tight. All3876566203systematic
positions,16scalar curves,all values/hash and packed readback pass. Scheduler
2h22min49s. Profiler/logging included; this is not an unprofiled repeat baseline.
Cgroup observation reaches602.416GiB recorded peak including cache, distinctfromRSS.

**GPU76646 failed** ongnode04:398.463GiB resident owned device setup succeeds,
then total-memory guard360GiB triggers while host output allocation/message
validation begins. ProcessRSS221.577GiB remains below320GiBguard, but source
file cache adds~143GiB. No complete encoding/output/hash or GPU speedup exists.
No SlurmOOM/arithmetic failure reported; bounded shutdown leaves worker unreaped,
Slurm recordsFAILED1:0. Preserve all failure evidence. Propose512GiBhostrequest/
480GiBtotal guard with unchangeddevice/RSS/Go/compute caps; no retry submitted.
CPU success/source554/11evidence/37cache reconcile; GPU failedsource554/16evidence/
34cache reconcile separately. Other s17 file/compact/disk/packed-stream modes are
still unrun; separate packing of resident CPU output does not validate those modes.

## Superseded s17 runtime forecast

* **Historical low-confidence forecast, superseded by the measured GPU row:** the
pre-trial scenario allowed
roughly12–24h after discovery, or roughly24–48h for fresh discovery plus
preparation/verification/persistence. Baselines are measured s16 GPU-assisted
preparation8415.469s (2h20m15s, excludes discovery) and CPU discovery10597.545s
(2h56m38s). N grows4x; an illustrative N·log2(N)^2 factor is
4(34/32)^2≈4.52, giving10.56h preparation and13.29h discovery before extra
width/communication effects. The12–24h construction range allows approximately
1.1–2.3x over that model for wider values and integration costs; this allowance
is a judgment, not a measured bound. The end-to-end range rounds their sum and
allows persistence/checks, not queue time. Actual runtime may exceed it.

This historical GPU scenario is not a runtime estimate for the new native
executor. Native weight-class enumeration replaces per-monomial carry evaluation,
so the old discovery extrapolation is no longer a calibrated model for that
implementation. Its GPU discovery/preparation is now measured by75819/75832;
CPU-only full preparation remains unmeasured.

This assumes a working wider server backend with adequate resident memory and
useful parallelism. The three-array GPU payload is~384GiB: three H200s are a
capacity floor, with four on one node a layout/headroom option, not implemented
support at the time of that forecast. Exact K and the four-card GPU workflow
are now measured; cross-node communication remains unmeasured. No linear
speedup from GPU count is assumed. Faster GPU-assisted discovery/native host
work could lower latency; streaming or poor inter-GPU scaling could raise it.
These scenarios are not a promise that the present code can run s17. Use bounded
wider preprocessing calibration before assigning full-run limits.

