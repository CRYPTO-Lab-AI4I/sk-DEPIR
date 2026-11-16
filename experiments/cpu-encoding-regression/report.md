# Larger CPU encoding refactor comparison

Status: **prepared; full s14 results pending**. This follows the
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
