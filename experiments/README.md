# Experiment conventions

Place each new experiment in `experiments/<descriptive-name>/`. Keep unrelated
measurements in separate directories. Commit small, useful raw results; for large
artifacts, link to a durable location and record checksums in the report.

Each experiment must contain:

- A runnable script such as `run.sh` or `run.py` with all parameters explicit.
- `report.md` linking to that exact script and its results (local relative links
  or durable external links).
- Results in the same directory or a durable link recorded in `report.md`.

The report must state the purpose, repository revision, prerequisites, commands,
input data and parameters, random seeds if applicable, hardware/software
environment, timing boundaries, units, results, and known limitations. Explain
how to reproduce and interpret each figure or table. Separate actual measurements
from extrapolations and values copied from the paper. When an experiment changes,
update its script, results, and report together, then post an update on the
[agent message board](../message-board/README.md) and check the README.

Do not put unreported measurements into the README. Link each new performance
claim there to its experiment report.

## Preventing avoidable failures

Reuse working launch/controller patterns before introducing new ones. Check
remote prerequisites at the beginning of the admitted job, before expensive work:
all participating nodes must be able to access the executable, wrapper, inputs,
working directory and output paths, with the expected permissions and environment.
Keep shared rank entry points separate from node-local source/build scratch.

Budget new instrumentation from measured representative costs; ordinary execution
latency does not establish sanitizer latency. Account for case/pass counts, setup
and verification within both step and overall deadlines, leaving headroom. Use
existing evidence when sufficient; otherwise calibrate a bounded representative
case before committing to a large suite. Any resource increase still needs admission.

Persist completed phase results, profiles and verification status before later
phases. A later timeout must leave earlier passes reviewable while the overall
job remains failed. State any reduced validation coverage explicitly. Cover a
preventable launcher/controller bug with a focused regression before retrying.
These are job-side checks, not permission to monitor submitted jobs or to launch
extra calibration jobs without need. Existing jobs are not retroactively modified.

## Long-job progress logs

Future jobs expected to exceed30 seconds must write durable, flushed progress
logs beside their results. Record UTC timestamps and monotonic elapsed time,
phase start/completion/error events and progress inside long phases, including
native CPU/GPU transforms. Identify the transform/pass and completed/total
blocks, tiles or elements, with units. “Encoding started” plus memory samples
cannot establish progress; instrumentation must reach the long-running work.

Rate-limit periodic updates to roughly30–60 seconds during long work, alongside
resource observations and guard failures. Distinguish a heartbeat from advancing
work counters. ETA is optional: report it only from measured progress with known
remaining work, label its assumptions and leave it unknown otherwise. Keep log
volume bounded and report instrumentation overhead and timing boundaries; avoid
synchronizing GPU work per element or adding full data passes just for logging.
Document a representative sample/schema in each new runner's report and check
that progress records are actually emitted before admitting larger runs.

This requirement applies to future runs; existing running jobs are not retrofitted
by the policy change. Submission still ends with a handoff: do not poll state or
follow logs until the user requests a check or reports completion, as required by
[AGENTS.md](../AGENTS.md).

## Profiling and preprocessing retention

During optimization, collect stage-specific CPU profiles and retain reusable
preprocessing by default. A strong reason to omit either must be recorded in the
report; disable profiling when optimization is complete. Keep reasonably sized
local caches across runs, with code specifications, source revision/hashes,
checksums, size and regeneration/load commands. Large caches belong outside Git.
Check the project's machine budget before retaining them and document size-based
omissions. Report discovery, reusable preparation, cache save/load and encoding
separately. Profiled times include instrumentation overhead and must not silently
replace unprofiled latency baselines. Historical scripts/results retain their
original methodology; apply this policy to new optimization runs.

## Recorded experiments

- [Integrated genuine-code preprocessing pilot](distributed-preprocessing-pilot/report.md):
  discovery, sharded roots/Newton/identity/spectrum, prefix commit/reload and canonical
  artifact comparison; one normal suite and one smallest-case sanitizer, prepared.

- [Bounded root operand routing](distributed-root-routing/report.md): coefficient
  ownership tiles replace full peer polynomials in an explicit comparison;
  generic CPU/race tests and exact source/target proofs passed78662 (5arms);
  largest shape halves input traffic with no full peer polynomial retained.

- [Local GPU roots and distributed final merge](distributed-root-merge/report.md):
  shared product driver, bounded synthetic closed supports, complete native/scalar
  references and paired input ownership checks; passed78508 through65532 roots
  with largest distributed product0.603 s, controller2min28.668 s.

- [Larger structured GPU products](resident-product-larger/report.md): exact
  streaming/prefix oracle,N16384/65536/262144,host/local-resident comparison;
  passed78469; at262144 values resident net1.554 s versus host-stored GPU4.911 s,
  with unchanged caps and complete bounded proof logs.

- [GPU-owned local Taylor data](resident-taylor-local/report.md): same recurrence/
  remote source schedule, local GPU access and raw-write validity checks;
  78460 passes correctness/sanitizers and matched speedup atN16384, same resident memory.

- [Matched resident/host product scaling](resident-product-scaling/report.md):
  N4096/16384,127/4096-word tiles and3samples/pair;78432 passed, resident still slower.
  Total/net diagnostic clocks and sanitizer coverage are explicitly separated.

- [Resident GPU polynomial products](distributed-resident-product/report.md):
  shared product/FFT order and pointwise kernel, three reusable arrays; local
  78399 passes24 normal/24 sanitized outputs and CUDA checks.

- [Complete resident GPU FFT](distributed-resident-fft/report.md): shared Taylor
  storage/wave/level/pass and FFT order; CPU/race/build checks pass, bounded GPU
  validation78396 passes24 normal/24 sanitized outputs and CUDA checks. Fullsize scaling remains separate.

- [Resident GPU vector composition](distributed-resident-vectors/report.md):
  shared launcher/order and resident interleaved shards; local CPU/race/build
  78394 passes24 normal/24 sanitized rows and CUDA checks; no fullsize claim.

- [Distributed GPU resident storage](distributed-gpu-storage/report.md): three
  reusable shards, GPU row/interleaved pack/scatter and shared transpose schedule;
  78392 passes208 normal/208 sanitized rank/pass outputs and CUDA checks. No
  full resident FFT/product or fulls18 result yet.

- [Distributed GPU polynomial product](distributed-gpu-product/report.md): shared
  forward/pointwise/inverse composition, independent schoolbook/alias checks;
  78321 passed bounded MPI/GPU products and CUDA checks. No fulls18 or selected resident layout.

- [Distributed GPU additive FFT](distributed-gpu-fft/report.md): shared Taylor
  source windows and rank vector/transpose composition; local independent CPU/
  race tests and78307 full bounded MPI/GPU gate passed. No fulls18 result.

- [Distributed GPU vector composition](distributed-gpu-vectors/report.md): shared
  rank transpose plus dedicated GPU vector owner;78302 passed all24normal/
  24sanitized rank/pass results and480proof endpoints. Full-size costs unmeasured.

- [Persistent GPU vector stages](gpu-vector-stage/report.md): local storage/global
  coordinates separated using existing kernels; bounded CPU-oracle/reuse/error/
  78293 independently passed48 normal /42 split-sanitized cases, zero errors/leaks.
  Raw reporting failure preserved; stream parser fixed. No complete distributed FFT
  or full-size throughput claim.

- [Rank-sharded transpose](distributed-transpose/report.md): bounded32/64-bit
  data-movement prerequisite for collective FFTs, independent scalar/rendezvous
  tests; host-only MPI validation78279 passed (16 operations /7,312 endpoint
  checks), no distributed GPU FFT/default or full-size throughput claim.

- [GPU tree with UCX cache control](ucx-cache-gpu-tree/report.md): same-build
  78261 passed the full crossover: CPU/state/proof checks and 16 clean CUDA
  memcheck processes. No consistent mode speed winner; explicit cache setting,
  not full-size support. Earlier78249 logging-failed evidence remains preserved.

- [UCX registration-cache control](ucx-registration-cache/report.md): retain
  78218 passed all host arms (1,024 handoffs / 2,048 checks), including direct
  buffers. Supports cache interaction; full preprocessing cost/correctness unknown.

- [Automatic MPI component discovery](mpi-transport-discovery/report.md): original
  78009 selected UCX on all ranks; staged controls passed and direct corruption
  reproduced. Paired evidence retained; no underlying cause proved.

- [TCP-only MPI transport control](mpi-transport-control/report.md): explicit
  78001 stalled on the first staged payload and hit90 s test cutoff. No completed
  transfer/proof or speed/corruption conclusion; raw failed evidence retained.

- [Host-only MPI buffer replay](mpi-host-buffers/report.md): captured public
  payload, shared transport/proofs, same-node/cross-node routes and explicit
  reclamation controls.78000 passed both staged arms then reproduced direct
  cross-node corruption; paired captures retained, no underlying cause proved.

- [Direct versus MPI-owned staging](mpi-buffer-comparison/report.md): one build,
  staged/direct/direct/staged order on one existing-budget allocation. Same checks,
  77868 passed staged-a then captured a direct-a same-node transfer disagreement.
  Later rounds unrun; no valid speed ratio or default change. Paired failure
  payloads and completed staged CPU/proof/memcheck evidence retained.

- [Root-transfer diagnostics](root-transfer-diagnostics/report.md): targeted
  proof exchange and paired failure snapshots after77704's transfer hash mismatch;
 job 77782 failed later on an uncovered support-index packet. All 78 completed
  polynomial proofs matched; no paired failure capture. Cause remains unknown.

- [MPI-owned chunk staging](mpi-owned-staging/report.md): explicit bounded-buffer
  probe with proof/capture for both index and coefficient packets; same two-node/
  two-H200 budget, extra 512 KiB host scratch per rank and CPU copies. Runtime
  penalty unmeasured; 77799 passed normal validation and CUDA memcheck (516
  normal proofs). Workflow failed later on profile-summary timeout; raw profiles
  recovered offline. Earlier direct-buffer failures remain open; no default change.

- [Rank-owned support and GPU merge tree](root-merge-tree/report.md): shared
  bounded routing, deterministic unique ownership and complete state-chain/CPU
  reference checks. Four logical ranks share2H200/2HPC nodes;77704 failed a largest-case transfer hash/CPU equality check,
  no4GPU speedup, distributed discovery/fullsize transform or default claim.

- [Buffered preprocessing persistence](preprocessing-persistence/report.md): shared
  payload/fsync/atomic publication with explicit4MiB buffering, fixed96MiB trials
  at1/2/4 writers, full restart readback and file-call/cache accounting. Bounded
  HPC77352 passed:4MiB coalescing cuts calls1,536→24; elapsed benefit mixed,
  readback dominates. Atomic/restart/full concurrency race checks pass; no full-size
  throughput or default change.

- [Cross-node GPU root products](distributed-root/report.md): shared normalized
  roots, complete host-polynomial MPI transfer and GPU merge, independent CPU
  coefficient oracle and retained cross-rank array checks. Implemented; bounded
  two-node validation77285 passed (2H200/2HPC nodes). Largest synthetic s17 case:
  pipeline1.180s, peak normal rank RSS410.523MiB, ownedGPU15.241MiB/card. Complete
  cross-rank/CPU array checks and both memchecks pass; no fulls18 inference or default change.

- [Explicit streamed preprocessing export](preprocessing-export/report.md): shared
  verified polynomial stage, bounded canonical output with atomic publication,
  support/recovery lifetimes and complete stream/resident correctness validation
  (76710 passed); full-size savings/costs unmeasured, resident defaults and s18
  admission unchanged.

- [Preprocessing memory prototypes](preprocessing-memory/report.md): host-lifetime
  audit and bounded GPU spectrum export, with independent scalar/failure checks
  and transfer-inclusive comparison; full export integration, tiled transforms
  and split products follow separately. Bounded primitive validation76704 passes;
  full export integration and full-size memory/runtime remain pending.

- [Host-backed transforms and split products](preprocessing-tiles/report.md):
  shared-kernel bounded comparison, independent CPU fixtures and memory/transfer
  counters; bounded validation76756 passes. Measured lower-memory options are
  slower; full-size costs remain unknown, with no admission or default change.
- [Root/generator boundary validation](preprocessing-root-boundary/report.md):
  shared arithmetic, immutable support and partitioned-root equivalence; bounded
  backend/recovery validation77126 passed, no measured memory saving.
- [Two-node staged GPU exchange](gpu-staged-transport/report.md): shared field
  kernels, pinned host staging/direct MPI and receiving-GPU ownership with full
  CPU/roundtrip checks and both CUDA memchecks;retry77281 passed on2H200/2HPC
  nodes. At32MiB/rank, median pipeline11.827ms versus2.269ms host-only exchange;
  actual root-polynomial transfer/merge is next, fullsize runtime unknown.
- [Root-subset membership comparison](root-subset-validation/report.md): explicit
  bitmap versus sorted-search validation sharing every check; paired worker/merge
  profiles/memory measurement77213 passed. Sorted membership improves measured
  sparse-s17 workers; smaller cases show overhead. Ordinary defaults stay bitmap.
- [Independent GPU root workers](parallel-root-workers/report.md): bounded
  same-node1/2/4-worker root construction and complete merge versus shared schedule,
  independent CPU/scalar references; comparison77148 passes. Largest synthetic
  s17case shows1.165 observed ratio including merge, with substantial validation
  bitmap overhead; not fullsize/cross-node speed evidence.
- [Two-node RAM redistribution](preprocessing-transport/report.md): compiled
  optional MPI host-memory probe, no shared-file intermediates; job77127 passed
  on two HPC nodes.32MiB inputs yield6.385GiB/s remote payload/node; zero GPUs,
  not a distributed transform/root or four-node bandwidth benchmark.
- [s18 preprocessing route comparison](s18-preprocessing-routes/report.md):
  reproducible whole-workflow host/device/Newton/storage models and hypothetical
  communication/work scaling; distributed host/root/GPU design recommended for
  speed, distributed prototyping selected, with network/integration unmeasured. No new job.
- [s18 preprocessing admission](s18-preprocessing-admission/report.md): one-GPU
  exact weighted-class dimension count, independent/known-size comparisons and
  explicit host/device/storage lifetime models; full preprocessing remains deferred.

- [Full s17 CPU/GPU encoding](s17-full-encoding/report.md): same verified code and
  database, frozen source, full mathematical/output checks and separate profiled
  load/setup/encoding/packed-output phases; CPU76645 and four-H200 GPU76706 pass
  with identical canonical/packed hashes.

- [Resident GPU client encoding](gpu-code-encoding/report.md): shared CPU/GPU
  schedule, explicit same-node resident code/workspaces, width/ownership/cap
  checks and bounded Go/CPU/CUDA-sanitizer validation (pending).

- [Larger CPU refactor comparison](cpu-encoding-regression/report.md): matcheds14
  resident encoding and separate setup, balanced old/new order, canonical-hash
  checks, explicit progress-only baseline instrumentation and separate profiles.

- [Resident GPU s17 preprocessing](gpu-s17-preprocessing/report.md): shared width
  kernels, exact GPU classification/root/transform/product execution and same-node
  peer shards; bounded CUDA/sanitizer validation pending.

- [Shared s17 preprocessing](s17-preprocessing/report.md): width-selected native
  executor, shared sampling/root/reciprocal procedure, exact small references,
  high-bit kernels and support-bound checkpoint checks; validation pending.

- [CPU encoding modules](cpu-encoding-modules/report.md): shared native32/64-bit
  schedule, storage/output helpers, common data verification, cross-mode/
  lifetime/sanitizer checks and matched s10 old/new timing validation.

- [Width-aware code data](s17-code-data/report.md): legacy byte/slot compatibility,
  explicit wide headers, trusted-digest/cap/range validation and bounded stream
  verification; no full s17 cache or preprocessing admission.

- [Wide CPU interface/matrix](s17-matrix/report.md): separate64-bit handles, Go
  ownership, independent small-matrix oracles, parallel Taylor branch, count-width,
  race/sanitizer and32-bit regressions; full encoding remains pending.

- [Native64-bit transforms](s17-native/report.md): bounded high-bit/ordered-basis
  forward/inverse checks, independent Go polynomial evaluations, sanitizer and
  existing32-bit CPU regressions; wide ABI/full encoder remain pending.

- [Wide s17 scalar arithmetic](s17-arithmetic/report.md): bounded CPU-only
  uint64 packing/primitive-generator checks through s20 with independent scalar
  oracles and s≤16 coordinate compatibility; full encoder limits unchanged.

- [HPC encoding table gaps](encoding-table/report.md): bounded CPU/GPU runs for
  22 implemented but unmeasured cells through s16; retained preprocessing, explicit
  memory modes, separate HPC/laptop labels and whole-output correctness checks.

- [Resident GPU products](gpu-products/report.md): unequal/truncated products,
  overlap-add/old-ABI/memcheck and complete CPU/transform-only/resident s10 trials.
- [Bounded s16 resident intervals](s16-resident-intervals/report.md): verified
  support census and exact CPU/GPU partial root products; later-operation inventory
  exposes the unmeasured full-domain gap instead of claiming a whole-run forecast.

- [Composed GPU matrix transforms](gpu-matrix/report.md): exact full CPU/GPU
  matrix/Taylor and resident convolution comparisons, small Horner/memcheck,
  persistent device state and complete GPU-assisted s10 preparation/cache checks.

- [CPU preparation kernels](preparation-kernels/report.md): specialized Go/C++
  polynomial loops, native-transform crossover and complete s10 comparisons.
- [GPU transform tiles](gpu-transform-tiles/report.md): exact CPU/H200 forward/
  inverse comparisons with same-arithmetic CPU reference, setup and transfers.

- [Targeted preparation profile](preparation-profile/report.md): read-only saved
  s16 support sampling, root generation, lookup leaf products and bounded native
  convolution clocks/profiles; no whole s16 preparation or GPU claim.

- [Preparation arithmetic](preparation-arithmetic/report.md): bounded s16 leaf
  profiles and matched complete s10 runs comparing private base-field lookup
  with table-free preparation after the first full s16 run was stopped.

- [Guarded Linux s16 first use](hpc-s16/report.md): current array/storage audit,
  explicit controller budgets, small first-use/cache admission and a guarded
  full CPU experiment. See the report for current status; no full result is assumed.

- [Count-width compatibility](count-width-validation/report.md): s-dependent stored
  count widths, 2^32 boundary/sanitizer checks and a cached s15 regression; see
  the report for current status. No full s16 workload is included.

- [Linux HPC cached s15 validation](hpc-cached-s15/report.md): bounded Slurm CPU
  validation, isolated Go/native suites and cached mathematical encoding from
  transferred constants; see the report for current status and resource limits.

- [Bounded-sample parameters](bounded-sample-parameters/report.md): analytical
  security-profile selection, not rate measurements.
- [Lift-RS rates](liftrs-rates/report.md): uniform monomial sampling for all 13 PIR
  configurations, with reproducible counts and confidence intervals.

- [Greedy planner costs](greedy-costs/report.md): logical storage and per-record
  traffic for a 21-case application/record-size grid, with complete compact plans.

- [Application width-search costs](application-costs/report.md): updated 21-case grid,
  split/stack plans, full-download policy results, and indexed-only alternatives.

- [Bounded tradeoff costs](tradeoff-costs/report.md): nondominated alternatives for
  the 21-case grid, with communication rate and explicit approximation flags.

Default application-cost grids require at least **5,000 records**; the retained grid
has 21 cases. Code-parameter experiments and correctness fixtures are separate.

- [Cost-grid DP](dp-costs/report.md): separate maximum-coverage DP on rounded
  storage/communication coordinates, with explicit cost windows and the 21-case table.

- [Paper/cyclic encoder comparison](encoder-comparison/report.md): exact bounded
  encodings with reusable public preprocessing artifacts; separate setup, load,
  encoding, process memory and output I/O measurements.

- [Download-objective comparison](download-objective-comparison/report.md): rerun
  all four cost grids with storage/download ranking; preserve prior results and
  compare endpoints, displayed options and approximate intermediate frontiers.

- [Table-free encoding checkpoint](tablefree-encoding/report.md): portable
  arithmetic through s=20, preserved old codewords, and exact s=9 encoding with
  separate discovery/preparation processes, prepared artifacts and packed output.

- [Encoder CPU profiles](encoder-profile/report.md): separate s=9 paper/cyclic
  preparation and database-encoding profiles, with checkpoint output verification.

- [Fast polynomial products](fast-polynomial/report.md): independent convolution
  checks, kernel timings and same-machine old/new preprocessing and encoder costs.

- [CPU encoding at s=10](encoder-scaling/report.md): exact capacity, reusable JSON
  artifacts, paper/cyclic agreement, independent curve checks and separate heap/output measurements.

- [Current s=10 cyclic CPU profiles](encoder-profile-s10/report.md): separate reusable
  preparation and database encoding, with stage/source attribution and field-kernel priorities.

- [Progressive cyclic scaling](cyclic-scaling/report.md): measured s=11/12 with
  an explicit first-use/RAM policy and labelled forecasts for larger catalogue fields.

- [Portable field multiplication](field-multiplication/report.md): general-kernel
  comparison and bounded reusable FFT constant tables, with full cyclic scaling.

- [s=13 cyclic benchmark](cyclic-s13/report.md): a two-hour case allowance with
  separate discovery, reusable preprocessing, encoding and resource observations.

- [s=13 CPU profiles and retained cache](encoder-profile-s13/report.md): profiled
  discovery/preparation, separate cache save/load and one database encoding.

- [Go/C++ kernel languages](kernel-languages/report.md): shared four-bit tables
  and inputs, batched kernels and full-size additive-transform comparison, with
  language-labelled implementations and preserved historical source evidence.

- [Native transform integration](native-encoding/report.md): matched full cyclic
  encoding using Go or batched C++ transforms, retained Go preprocessing, ownership
  and interface accounting, full output checks and stage-specific profiles.

- [Encoding layout verification](encoding-layout/report.md): separated backend
  builds, legacy compatibility, pre-move fixture hashes and historical evidence.

- [Encoding resource model](encoding-resources/report.md): analytical array,
  preparation, packed-cache and I/O scenarios for the scaling study; no new
  encoding/RSS measurement or larger-size implementation.

- [Scaling prototypes](scaling-prototypes/report.md): sparse setup, paired cyclic
  preparation, lean online files/workspaces, exact matrix/Taylor and canonical
  routing probes, with profiled s=13 hash validation and explicit I/O limitations.

- [Local scaling calibration](local-calibration/report.md): native tile/core
  scaling, explicit-cache block transfers, compute/I/O overlap and fresh-process
  preparation/minimal-loader RSS, with GPU inventory and interpretation limits.

- [Parallel encoding integration](parallel-encoding/report.md): full canonical
  s=13 comparison with general-basis/affine native tiles, parallel Taylor/matrix
  execution, blocked transposes and buffered minimal-plan loading.

- [Profiled s=14 first use](parallel-s14/report.md): exact discovery, reusable
  lean Go/native preparation, retained binary constants, complete encoding and
  output validation with Go/native profiles. See the report for current run status.

## Encoding implementation paths

The complete API, preprocessing and orchestration now live in
`encoding/liftrs-shared/`, used by both backends. `encoding/liftrs-go/` contains
Go-only transforms; `encoding/liftrs-cpp/` contains native transforms and binding.
The [directory guide](../encoding/README.md) gives both execution commands. Active
runners collect shared, field and backend source dependencies. Historical results,
snapshot scripts and source hashes retain their original measured bytes/paths.
[source_evidence.py](source_evidence.py) resolves current/moved paths,
[content-addressed snapshots](source-snapshots/README.md) and Git history without
rewriting measurement hashes. Archived RM benchmarks/plot are isolated in
`encoding/rme_legacy/` and are not Lift-RS performance evidence.

- [Complete native online encoding](native-online/report.md): cached C++ encoding
  versus hybrid Go/C++ execution, with identical canonical words, profiled stages
  and separate initialization/process-memory accounting.

- [Encoding timing consistency](encoding-timing/report.md): short clock controls
  and alternating cached native/hybrid execution, with same-clock totals, CPU
  time, profiles and resource accounting.

- [Preparation-memory audit](preparation-memory/report.md): small allocation-layout
  replicas and source-based s=15 preparation scenarios, distinguishing live
  payload, allocation traffic and measured process RSS.

- [Bounded preparation checkpoint](preparation-checkpoint/report.md): compact
  membership, batched factors and reusable scratch, with exact constant checks
  and diagnostic live-heap/process-memory measurements from retained support.

- [Guarded local s=15 first use](parallel-s15/report.md): exact discovery, bounded
  preparation and complete native encoding with retained profiles/cache and
  explicit memory, time and power guards. First use and fresh cached execution
  pass: about 6.5 minutes per encoding and 18.46 GiB cached peak RSS.

- [GPU domain metadata and bounded scaling](gpu-scaling/report.md): allocation-free
  full-domain descriptions and progressive exact CPU/GPU operation comparisons.

- [Actual GPU-assisted s16 preparation](gpu-preparation/report.md): saved support,
  normalized-root restart checkpoint, complete identity checks and guarded resources.

- [Cached s14 CPU encoding profile](cpu-encoding-profile/report.md): native stage
  attribution and8/16/32-thread comparisons using retained preprocessing.

- [Within-block CPU Taylor parallelism](cpu-taylor/report.md): independent native
  division checks, sanitizers and matched baseline/candidate cached s14 encoding.

- [Cached s16 CPU encoding](cached-s16-encoding/report.md): retained GPU-prepared
  constants, selected CPU Taylor path and systematic/curve/packed-readback checks.

- [Primary CPU/CUDA online encoding validation](cuda-encoding/report.md): optional
  modular client backend, CPU-only builds and complete cross-backend output checks.

- [Full cached s16 CUDA encoding](cuda-s16-encoding/report.md): two encodings with
  reused device buffers, retained CPU digest checks and separately measured setup.

- [Encoding memory audit](encoding-memory/report.md): allocation-free phase
  inventory, s16 measured-memory reconciliation and explicit s17 memory–time choices.

- [CUDA buffer reuse](cuda-buffer-reuse/report.md): opt-in two-work-array schedule,
  exact device-memory savings, oracle/memcheck checks and explicit latency comparison.

- [Native CPU buffer reuse](cpu-buffer-reuse/report.md): owning scratch/output
  transfer, alias boundaries, sanitizer checks and memory/runtime validation.

- [GPU transform scratch study](gpu-transform-memory/report.md): explicit direct-row
  versus full-scratch layouts, exact memory counts and measured runtime trade-offs.

- [Bounded GPU canonical output](gpu-output-memory/report.md): explicit compact
  online mode, staged narrowing/disjoint assembly and memory–time validation.

- [Bounded CPU canonical output](cpu-output-memory/report.md): explicit compact
  owning arena, bounded column scratch and measured memory–time validation.

- [Streamed constants and s17 widths](streamed-constants/report.md): allocation-free
  source/peak audit, scratch-borrowing GPU design and explicit transfer/I/O alternatives.

- [Fast-layout GPU streamed constants](gpu-streamed-constants/report.md): explicit
  P/spectrum uploads through dead scratch, matched s14 and full s16 validation.

- [Fast CPU file-backed constants](cpu-streamed-constants/report.md): verified
  descriptor reads through dead scratch, RSS/file-cache accounting and matched validation.

- [GPU output to a bounded packed sink](gpu-packed-output/report.md): direct
  atomic bit packing, bounded download/writer, matched output pipelines and
  separately measured packed-file verification.

- [GPU host-backed transform component](gpu-host-backed/report.md): bounded
  device tiles, full host domain, dependency-safe Taylor batches and measured
  transfer/allocation costs before full online integration.

- [Parallel work inside host-backed GPU tiles](gpu-backed-parallel/report.md):
  per-element stage kernels to address measured low parallelism, fixed-input
  comparison at unchanged device/host staging and transfer counts.

- [Complete host-backed GPU encoding](gpu-host-encoding/report.md): explicit
  five-transform schedule, retained host code/workspace, bounded device tiles,
  owning canonical output and matched full s10 CPU/reference validation.

- [Direct CPU packed output](cpu-packed-output/report.md): canonical bit scatter
  in dead scratch, bounded writer/hash, mixed owning/error checks and matched
  cached s10 pipelines. Two full transform arrays remain counted.

- [Bounded private-RAM CPU disk transforms](cpu-disk-transform/report.md): joined
  file Taylor passes, blocked transposes, progressive oracles, I/O counters and
  separate cache/cgroup accounting before full online integration.

- [Complete CPU disk-backed encoding](cpu-disk-encoding/report.md): five native
  file transforms, bounded products, owning canonical output and full matched
  file-fast/disk s10 workflows; backing/cache costs remain explicit.
