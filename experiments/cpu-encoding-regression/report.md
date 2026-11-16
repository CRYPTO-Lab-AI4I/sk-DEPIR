# Larger CPU encoding refactor comparison

Status: **job76055 passed; completed results reconciled after the user completion ping**. This follows the
[small s10 comparison](../cpu-encoding-modules/report.md#completed-corrected-validation),
which found encoding1–3% slower and resident/file setup about2–3ms slower after
modularity changes. The purpose is to determine whether the difference matters
at a representative larger size before GPU client encoding integration. No
production algorithm or memory mode changes here; fulls17 preprocessing is untouched.

## Reproduce, provenance and resources

[Runner](run.py), [Slurm launcher](run.sbatch), [identical old/new probe](probe.go.txt),
[baseline logging patch](baseline_progress.py), [static checks](test_runner.py),
[completion analyzer](analyze.py). Results/logs will be under `runs/comparison-01/`;
source identities, exact commands, compiler/CPU inventory, affinity, resource
observations, all profiles and matching test executables are recorded there or
in `.cache/liftrs/cpu-encoding-regression/comparison-01/`. Large inputs remain at
their original location outside Git; no new preprocessing is performed.

```sh
python3 -B experiments/cpu-encoding-regression/run.py --check
python3 -B experiments/cpu-encoding-regression/test_runner.py
module load slurm/slurm/25.05 gcc/14.2.0
sbatch --output=experiments/cpu-encoding-regression/slurm-%j.log \
  experiments/cpu-encoding-regression/run.sbatch comparison-01
# Only after the user's completion ping:
python3 -B experiments/cpu-encoding-regression/analyze.py comparison-01
```

Run on one eligible CPU allocation in gpuq,32 cores,32GiB requested RAM,24GiB
process-group RSS guard,30GiB cgroup guard including file cache,20GiB Go soft
limit,16GiB new retained cap and16GiB scratch. No GPU is requested; no fixed node,
exclude gnode03. srun binds the job to its allocated cores. Worker limit15min per
command, controller110min and scheduler120min. These are within the previously
approved [s14 encoding-table budget](../encoding-table/run.py), which the
no-allocation admission check verifies. Deadlines are limits, not time estimates.
Planning estimate25–45min execution plus unknown queue time, based on prior
roughly60–91s s14 calls and20 total warmup/measured calls across timing/profiling;
not a measured duration for this experiment.

Go1.27.3, GCC14.2/C++17 and the repository's generic compiler flags. The old CPU
libraries are exactly Git671bf5e before the explicitly recorded logging patch;
current source revision and every copied source hash are sealed at job start.
The same probe file is copied into both package trees. Source snapshot, pristine
baseline hashes, modified hashes and `baseline-progress.diff` allow independent
replay/reconciliation; no benchmark file is silently substituted or arithmetic
source change hidden.

## Inputs and correctness

Full case: s14,m2,t5,N268435456,K50976325. Read-only input:
`.cache/liftrs/parallel-s14/s14-m2-t5/online.bin`, SHA-256
`88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8`.
Database values use Go math/rand seed71 as in the
[original verified s14 run](../parallel-s14/report.md). Each of the warmup and
measured outputs is streamed through bounded little-endian uint32 hashing and
must match the previously verified canonical word digest:
`b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c`.
That historical word passed every systematic position and independently evaluated
curve restrictions; those curve checks are not rerun in this timing experiment.
Exact whole-word equality to that digest is required, not merely agreement between
the two implementations. No full output or packed file is retained; there is no
need for another1GiB word copy or repeated storage benchmark.

Before large work, generate a small genuine s4,m2,t3 fixture using independent Go
preprocessing/encoding and run the same resident probe in both binaries. The
fixture, code/message specification, digests, executables and profiles remain
retained. This prevents a build or logging-patch error from consuming the larger
benchmark budget.

## Measurement boundaries

Four paired samples use old/new, new/old, old/new, new/old process order on the
same allocated cores. Every fresh process separately times:

| Phase | Included |
|---|---|
| `load-resident` | File open, streamed read/hash/validation, complete resident code arrays; file close before timer ends |
| `open-file` | Verified file-backed encoder open and native metadata; close and subsequent GC outside timer |
| `native-plan` | Resident encoder metadata/plan construction |
| `workspace` | Initial reusable workspace allocation |
| `warmup` | One complete encoding, checked and reported separately |
| `encoding` | One subsequent complete resident EncodeWithStats call, including replacement-output allocation |

Message generation, output verification and explicit GC after warmup are outside
encoding timing. Allocation accounting and native stage clocks remain in the
record. The old returned output is released and GC completed before the measured
call in both versions; both retain warm workspaces. No compact/file/disk encoding
or packed-output timing is included. This is a targeted resident comparison;
other modes retain their earlier small correctness/performance evidence.

After the four unprofiled pairs, one separately profiled process per version
collects CPU attribution for setup, workspace, warmup and encoding. These timings
are marked `profiled:true` and excluded from median comparisons. Native C++ stage
clocks supplement Go CPU profiles, which may attribute native time to cgo rather
than individual C++ functions. Profiling overhead is included only in those
separate runs; it does not replace ordinary latency samples.

## Progress instrumentation and limitations

Future long jobs must emit useful progress. Pre-refactor671bf5e lacks that
instrumentation, so this experiment adds the current progress header, matching
completed-block counters in `parallel_jobs`, and phase observation around the old
resident schedule. It changes no arithmetic, buffer allocation, loop order or
thread scheduling. The exact three changed/added files and diff are retained.
Both binaries run with logging **enabled**, including inside measured encoding;
this intentionally differs from the earlier unlogged s10 comparison. Observer
clock/counter/log costs are included and not independently measured. Small
remaining differences cannot automatically be attributed solely to modularity.

Native events name the ten encoding operations and advancing joined blocks,
with UTC, completed/total units and elapsed milliseconds; long updates throttle
at30s. For example, a full-forward event identifies its parallel-block count.
Runner heartbeats additionally report RSS/cgroup/storage every30s and explicitly
leave remaining work unknown. Code loading and output verification use bounded
streaming; the latter reports advancing element counts for long verification.
Timers/report logs are flushed and retained. No agent monitoring follows submission.

OS/NFS cache state, memory placement, CPU frequency and other node load remain
uncontrolled; this is a same-node paired comparison, not a cold-cache benchmark
or a statistically conclusive1% test. The analyzer reports all samples, ranges,
medians and paired ratios; ratios above1.2 are flagged using the existing review
threshold, not treated as automatic correctness failures. A near1–3% encoding
difference with small absolute setup overhead should not block GPU integration;
a substantial repeatable larger slowdown warrants investigation first. No new
s14 equivalence, performance improvement or s17 capability is claimed in advance.

## Pre-submission validation

[Local evidence](local-checks.json), [current small probe](local-current.log),
[instrumented baseline small probe](local-baseline.log),
[baseline build](local-baseline-build.log). Static admission and exact-old-source
patch/idempotence tests pass. The current probe builds via a Go overlay, generates
a genuine independent-Go fixture and passes full-word checks; the instrumented
baseline compiles separately and matches that same fixture. The completion parser
checks both real logs, transform identifiers, successful ten-operation schedules,
word digests and exclusion of profiled samples. Local timings are s4 validation,
not s14 performance evidence.

Two local parser invocations initially used unsuitable log wrappers: ordinary
`go test` adds a trailing package summary, and stdout-only redirection omits native
stderr events. Direct test-binary invocation with merged stdout/stderr passes,
matching the submitted runner's existing behavior. The earlier logs remain
retained; no production or timing-probe bug was found from those wrapper errors.

## Submission

Job **76055**, submitted revision `e07f308`, accepted by sbatch; see the
[exact submission/budget receipt](submission.json). Results:
`runs/comparison-01/`; Slurm log: `slurm-76055.log`; retained binaries, fixture and
profiles: `.cache/liftrs/cpu-encoding-regression/comparison-01/`.
No queue/process/log checks after submission. Fulls17 preprocessing75832 was
neither changed nor inspected. Await the completion ping before analysis.

## Completed comparison (SCALING-086)

Job76055 completed successfully ongnode02 with32 CPU cores, no GPU,
[scheduler elapsed23min24s, exit0:0](runs/comparison-01/scheduler-completion.txt).
Controller elapsed1402.824s includes source preparation, builds, small correctness
checks, eight comparison processes, two profiling processes and profile summaries.
It is not the encoding latency. Submitted sourcee07f308; baseline671bf5e with
only the disclosed progress patch. [Results/commands/source identities](runs/comparison-01/results.json),
[reconciled analysis](runs/comparison-01/analysis.json), [Slurm output](slurm-76055.log).
The analyzer verifies521 source,31 original evidence,17 retained cache and141
pristine-baseline files, and exactly reproduces the logging-only modifications.
All native schedules, small independent-Go checks and20 fulls14 warmup/measured
output hashes pass. No deadline/resource guard fires.

Unprofiled medians over four samples per version:

| Operation | Before refactor | After refactor | Difference |
|---|---|---|---|
| Resident encoding after warmup | Runtime60.346 seconds | Runtime61.119 seconds | +0.773 seconds;1.28% slower |
| First encoding/warmup | Runtime60.660 seconds | Runtime61.311 seconds | +0.651 seconds;1.07% slower |
| Resident cache loading | Runtime1.894 seconds | Runtime2.540 seconds | +0.646 seconds;34.1% slower |
| Verified file-backed encoder opening | Runtime1.501 seconds | Runtime2.042 seconds | +0.541 seconds;36.1% slower |
| Native plan construction | Runtime5.813 milliseconds | Runtime5.869 milliseconds | +0.056 milliseconds;1.0% slower |

Sampled encoding-process peak RSS is3.593–3.595GiB in both versions; profiled
process peaks reach3.597GiB old and3.598GiB new. These include the process setup,
code and message/workspace/output arrays, excluding filesystem cache. The cgroup
and resource samples are retained separately; no memory-capacity model changes.

Encoding ranges are60.312–60.570s old versus61.009–61.158s new; paired ratios
show0.93–1.40% higher encoding times. This is a small repeatable difference in
this instrumented comparison, not proof that it is random noise or solely caused
by modularity. The first old resident load took21.457s, versus1.873–1.901s in
its remaining samples. All samples are retained, including that I/O outlier;
its cause was not established and the table reports the prescribed medians.

Workspace-only medians are3.093ms old and118.542ms new, but the samples are
bimodal: old2.946/3.047/3.138/228.169ms; new2.594/2.647/234.438/282.371ms
(sorted). The38.3x ratio crosses those two clusters and must not be presented
as a stable38x allocation regression. In the separate profiled samples,
workspace duration is264.992ms old and228.583ms new; nearly all CPU samples in
both are `runtime.memclrNoHeapPointers`, clearing allocated memory. These
[old](runs/comparison-01/workspace-old-top.txt)/[new](runs/comparison-01/workspace-new-top.txt)
summaries were derived during review from the already sealed profiles, not
another benchmark. The distribution/allocator effect remains a limitation;
absolute workspace cost stayed below0.283s in unprofiled samples.

The separate [old resident-load profile](runs/comparison-01/profile-old-load-resident.log)
and [new profile](runs/comparison-01/profile-new-load-resident.log) locate the
extra CPU work in the shared `readScalingCodeValues` decoder: roughly1.22s flat
versus0.52s in the previous loading loop. File-open profiles similarly show
roughly0.70s in the shared decoder versus0.18s in the previous opener. Read and
SHA CPU costs are comparable. These samples point to the decoder loop as the
setup follow-up; they are not a proven explanation of every elapsed difference.
Native encoding stage medians suggest the larger absolute differences are in
columns/transposes and rows, with a smaller canonical-assembly increase; no
single kernel cause has been isolated and no code fix is made here.

**Decision:** the bounded larger-size comparison is complete. About1.3% encoding
cost and0.5–0.6s additional load/open time do not justify delaying GPU client
encoding integration. Retain the decoder improvement and allocator variability
as follow-ups, with profiling evidence preserved. This is limited tos14 resident
encoding/setup on this node, logging enabled in both versions; it does not
validate larger memory modes or fulls17 execution. Preprocessing75832 was not
inspected or modified during this review.

Reproduce derived workspace summaries with the retained matching binaries:

```sh
export GOTOOLCHAIN=local
export GOCACHE="$PWD/.cache/liftrs/preprocessing-workflow-build-cache"
.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go tool pprof -top   .cache/liftrs/cpu-encoding-regression/comparison-01/old.test   .cache/liftrs/cpu-encoding-regression/comparison-01/profiles-old/workspace.pprof
# Repeat with new.test and profiles-new/workspace.pprof.
```

[Review seals and commands](runs/comparison-01/review.json) identify the derived
files. The two workspace summaries and scheduler-completion receipt are post-completion
review artifacts; the original sealed results/evidence map is unchanged. The
completion analyzer command above reproduces the measured comparison table.
