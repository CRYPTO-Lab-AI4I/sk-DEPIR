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
