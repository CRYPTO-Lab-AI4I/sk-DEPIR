# s18 preprocessing: exact dimension and memory admission

Status: **exact s18 dimension76647 passed; full preprocessing remains unimplemented/unrun**. User authorizes the initial dimension/memory audit after
fulls17 preprocessing passes, while CPU/GPU encoding jobs run separately.
No existing job is inspected/changed; no fulls18 preprocessing run is authorized
or submitted here.

## One exact membership criterion, less counting memory

The production monomial criterion is `good_weights` in
[monomial.hpp](../../encoding/liftrs-cpp/monomial.hpp). For m2, each binary digit
across exponentsx,y has weight0,1,2. A ternary class with `ones` digits equal1
represents exactly2^ones monomials, because either x or y can carry each such bit.
Weights0/2 have one assignment. Therefore exact K is the sum of2^ones over good
classes. This counts the same monomials as expanding a full bitmap; it is not
sampling, a rate estimate or a changed construction. The carry criterion still
checks **positive** multiples ofq−1, preserving the constant monomial.

A small shared `good_weight_class` helper decodes a class and optional multiplicity.
The existing production GPU class-classification kernel delegates to that helper;
its membership result is unchanged. [New count tool](count.cu) reduces weighted
class values on GPU using that same criterion, with bounded waves and one8-byte
explicit accumulator. No class table, q² bitmap, host index list or transform is
allocated. s18 has387420489classes, versus68719476736monomial positions and an
8GiB bitmap. Runtime/driver/kernel spill memory is extra;8bytes is explicit owned
payload, not whole GPU memory usage. No memory-saving full-preprocessing mode is
adopted and no runtime tradeoff is hidden: this is a different exact count-only
algorithm; it cannot produce the support needed by encoding.

Public `CountCodeMonomials`, preprocessing CLI/session/artifact limits, resident
GPU transforms and full discovery remain through s17. The standalone experiment
can count s18 because it needs only the existing membership recurrence. Extending
this diagnostic does not certify larger field/FFT buffers or relax public full
preprocessing admission.

## Reproduction and evidence

[Runner](run.py), [Slurm launcher](run.sbatch), [GPU tool](count.cu),
[independent Go reference](reference.go.txt), [host class self-check](class_test.cpp),
[memory formulas](memory.py), [admission tests](test_admission.py),
[completion analyzer](analyze.py). Results/logs will be in `runs/counts-01/`;
matching binaries/CPU reference profile remain outside Git at
`.cache/liftrs/s18-preprocessing-admission/counts-01/`. Explicit source revision
is frozen with `git archive`; later working edits cannot alter the copied code.
Every copied source/probe identity, environment, GPU/compiler information and
command is sealed in results; retain the command logs and reusable executables.

```sh
python3 -B experiments/s18-preprocessing-admission/run.py --check
python3 -B experiments/s18-preprocessing-admission/test_admission.py
module load slurm/slurm/25.05 gcc/14.2.0 cuda12.9/toolkit/12.9.1
revision=$(git rev-parse HEAD)
sbatch --output=experiments/s18-preprocessing-admission/slurm-%j.log \
  experiments/s18-preprocessing-admission/run.sbatch counts-01 --revision "$revision"
# Only after the completion ping:
python3 -B experiments/s18-preprocessing-admission/analyze.py counts-01
```

One H200,4CPU,8GiB requested RAM/6GiB sampled RSS/4GiB Go soft limit;512MiB
explicit-owned device ceiling (tool only owns8bytes),2GiB new retained/4GiBscratch.
Worker10min/controller15min/Slurm20min, finite TERM/KILL waits. These reduce the
previously approved small-test GPU/memory ceilings. One node, no pin, gnode03
excluded; no peer transfers or four-GPU reservation needed for counting.
Go1.27.3/GCC14.2/CUDA12.9; toolkit builds actual node architecture. CPU profiling
covers independent reference generation and is retained with its binary. The
GPU-only tool has synchronized wall time around its count stage; there is no
meaningful Go CPU profile for a CUDA reduction. Whole-command/setup/build/validation
latency stays separate from kernel/reduction wall time.

The tool processes2^20classes per wave, joining before advancing count progress;
start/completion/error logs and periodic completed/total classes distinguish real
work from30s controller resource heartbeats. Logs flush on each record. Runtime
is unknown until measured; the20min limit is not a prediction. No monitoring
follows submission. No expensive reusable preprocessing cache is created or
deleted: count artifacts, profiles, manifests and executables are preserved.

## Validation sequence

1. Enumerate independent Go `GoodMonomial` dimensions for s2..7,t1 and a denser
   t=q/2−1 (deduplicate s2). Compare every GPU weighted count against them.
2. Host C++ self-check s2..8: weighted class sum matches direct exponent membership;
   total multiplicity over all classes must equal4^s. This validates the class
   bijection, while the Go reference supplies the separate implementation.
3. CUDA memcheck a small s8,t5 count; require zero errors/leaks.
4. Count s10/16/17,t5 and require historical exactK146641/918672421/3876566203,
   then count s18,t5. These are complete class sums, not truncated domains.
5. Derive memory/file models from exactK and reconcile all evidence. A failed
   reference, range, complete-class or seal check rejects the result.

[Local evidence](local-checks.json), [reference log](local-reference.log),
[class log](local-class.log), [CUDA compile log](local-cuda-build.log).
Host class tests and independent Go reference generation pass; CUDA compilation
for available H200 architecture(sm90) passes without executing GPU work. Python
model/parser tests pass, Go membership/reference/count-admission regressions pass.
The original Go build attempt supplied a `.go.txt` filename and was rejected;
copying to the same `.go` form used in the job fixes the invocation. An initial
regression test selector matched no tests; the corrected explicit membership/
count-admission selector passes. No fullGPU or s18 dimension claim from localchecks.

## Current memory audit: what exact K will resolve

The [model](memory.py) remains the current resident dense-array algorithm;
it proposes no silent memory-time tradeoff. For s18, N=2^36 and field workspace
slots remain8bytes, so no new arithmetic slot width is required.

| Quantity | s18 formula/value (capacity model, not measured peak) |
|---|---|
| Discovery interpolation grid |8N =512GiB |
| Membership/union bitmap |N/8 =8GiB each |
| Discovery principal live arrays |16K+8N+N/4bytes; two index lists, grid and two bitmaps |
| Final retained support/P/G/spectrum |8K+12N+4bytes; **more than768GiB** even before temporaries |
| Full P·G identity result |4N =256GiB, additionally live at its stage |
| Three GPU product arrays |24N =1536GiB aggregate |
| GPU capacity floor at140.4GiB/card |11cards before runtime/headroom |
| At112GiB admitted owned cap/card |14arithmetic floor; power-of-two sharding would need16cards |
| Canonical reusable files |12K+12N+61bytes; **more than768GiB**, before profiles/recovery/logs |
| Two recovery generations per phase |About528GiB plus bounded frontier padding/metadata |

These are overlapping lifetime models, not numbers to sum into a claimed peak.
Host product/Newton temporaries, GC retention, staging and filesystem cache remain
extra. The previous fulls17 sampled348.393GiB peak cannot simply be declared to
scale4x; multiplying it is a warning scenario, not a measured s18 requirement.
ExactK supplies the support/P/G/file sizes needed for a concrete admission plan.
No fulls18 runtime forecast is justified yet; s17 root construction8h34min is
only a baseline to investigate, not a linear GPU-count scaling law.

Sixteen GPUs here span at least four4GPU nodes. Current peer memory backend has
no cross-node transport/collectives and an eight-device compile-time limit; simply
raising it cannot make remote pointers work. Host distribution is also needed
if full concurrent arrays exceed a node; a multi-node GPU rewrite alone does not
solve Go host-array lifetimes. Alternative streaming/recomputation/disk strategies
need their own explicit memory/time/transfer/storage plan before adoption.
Neither capacity nor willingness to use more GPUs admits those allocations.

After the exact count, present a concrete choice between distributed resident
preprocessing and a lower-memory design, including headroom, data movement,
persistence/checkpoint cost and unknown runtime. Fulls18 implementation and
large-budget agreement remain subsequent gates. s19/20 stay exploratory rows.

## Submission

Job **76647**, immutable source `bb524d3`, accepted by sbatch. [Exact receipt/caps](submission.json).
Results `runs/counts-01/`, log `slurm-76647.log`, retained tools/reference/profile
`.cache/liftrs/s18-preprocessing-admission/counts-01/`. No scheduler/process/log
polling or result analysis after submission. Wait for the user's completion ping.
Encoding76645/76646 were not inspected/changed. Fulls18 preprocessing remains blocked
on reviewed exact sizes, design choice/implementation and explicit large budgets.

## Completed exact count and memory audit (SCALING-092)

Job76647 completes0:0 ongnode05,59s scheduler elapsed,59.180s controller including
builds, all references and CUDA memcheck. Frozen sourcebb524d3.
[Results/commands](runs/counts-01/results.json), [analysis](runs/counts-01/analysis.json),
[model](runs/counts-01/memory-model.json), [accounting](runs/counts-01/scheduler-completion.txt),
[Slurm log](slurm-76647.log). Source/evidence/cache reconciliation passes560/29/4
identities. IndependentGo11small(s,t)cases match; host multiplicity/direct checks
pass; CUDA memcheck has zero errors/leaks. s10/16/17counts exactly match verified
historical dimensions. No resource/deadline failures. All tools/profile retained
at the submitted cache path.

**s18,m2,t5 exactK16300262095,N68719476736.** Weighted GPU class sum takes
0.0373797s (~37.4ms) over387420489classes, after device initialization; the complete
count command is sampled at1.298s by the one-second controller loop. Those are
different measurement boundaries, not evidence that preprocessing is fast.
Whole-suite peak sampled processRSS266.133MiB includes compilation, not full
preprocessing or an isolated GPU count peak. Explicit device accumulator8bytes;
CUDA context/runtime/kernel spills additional, unmeasured. No spectrum/support
or encoding artifact is produced by this exact count.

Exact-K dense-layout projections:

| Resource | Size / practical implication |
|---|---|
| Support file |121.446GiB |
| Normalized root |60.723GiB |
| Canonical P/spectrum file |572.723GiB |
| Generator G file |195.277GiB |
| All canonical reusable files |**950.170GiB** (1020236866033bytes), before other files |
| Discovery principal host arrays |**770.893GiB**, before transients/runtime/cache |
| Final retained host arrays |**889.446GiB**, before runtime/headroom |
| Full identity result at its stage |256GiB; do not sum indiscriminately across phases |
| GPU product arrays |**1536GiB**, before metadata/driver/headroom |
| Devices at112GiB owned cap/card |14arithmetic floor;16cards under power-of-two sharding, at least4nodes |
| Conservative two snapshots per phase |About528.001GiB plus filesystem/manifests/logs |

Final host arrays alone exceed the prior864GiB request/800GiB RSS limit and leave
only~17.6GiB under a907GiB schedulable node, before Go/native/runtime/cache
requirements. This is insufficient admission headroom; merely requesting a
slightly larger single-node budget is not a demonstrated solution. Newton/root
intermediates are additional at earlier phases. Canonical files alone exceed
512GiB previously authorized retained storage. With full two-generation snapshot
retention, payload/storage model approaches1478.171GiB (~1.44TiB), before headroom;
that storage is not approved by this count job.

### Design choice still required

| Route to prototype | Expected memory effect | Added cost / implementation gap |
|---|---|---|
| Distributed resident preprocessing | Split1536GiB device arrays and host-owned code/work data across16GPUs/fournodes; per-card product arrays96GiB before overhead | New cross-node transforms/communication, host ownership/distribution and distributed persistence; latency/communication cost unknown |
| Lower-memory preprocessing on fewer nodes | Stream/release selected host data, reduce device workspace or use recomputation/disk; exact savings depend on design | Additional reads/transfers/recomputation and possibly disk workspace; no quantified runtime or validated s18 layout yet |

Prefer investigating the distributed resident route for delegated preprocessing
speed, while auditing easy host-lifetime savings shared with existing code.
Neither option is adopted here. Do not assume16cards gives linear speedup or
that GPU distribution solves Go host-array pressure. Before coding a material
memory-time tradeoff, present concrete stage lifetimes, host/device savings,
network/transfers/I/O, storage/checkpoints and latency estimates/unknowns for user
choice. Keep faster paths/defaults when they fit. Fulls18 preprocessing and new
multi-node/storage budgets remain separate approvals. Runtime remains unknown;
s17root8h34min is a useful baseline but not an s18 estimate.

The memory/model and completion tests are rerun for publication. No new GPU run,
full allocation, encoding-job inspection or algorithm tradeoff is performed.
