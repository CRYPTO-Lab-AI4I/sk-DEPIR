# Encoding clock consistency and alternating comparison

[run.py](run.py) builds the [Go harness](main.go), records exact
source/binary digests, retains dirty source snapshots, and uses the existing
verified s=14 online artifact. Code preprocessing is not repeated.

First run six 250 ms busy/sleep controls (`--clocks-only --smoke`). Read the same
C++ `steady_clock` used by native stage counters, POSIX monotonic time, wall time,
Go elapsed time and all-thread process CPU time. Fail the short check if elapsed
clocks differ by more than max(5 ms, 2%). These tiny calibration controls omit
profiling because profiler setup would distort their known-duration intervals.
Encoding phases retain Go CPU profiles and native samples.

Then run a small s=8 encoding fixture with alternating backend order. Totals and
native stages use the same native steady clock. Record Go/POSIX/wall differences
as independent diagnostics, plus getrusage user/system CPU time and peak RSS.
Native stages must fit the same-clock total; correctness uses full output hashes.

After those checks pass, run one cached s=14 process with ordered cases
`native,hybrid,hybrid,native,native,hybrid`, 16 workers and one shared immutable
plan/workspace. Retain the first cold-workspace case rather than silently removing
it. Later cases reuse touched workspace pages; report ordered and paired values,
not a guaranteed speedup from three samples. Inputs and profiling are identical.
Prevent idle sleep during the run; keep all profiles and existing constants.
Budget: 12 GiB sampled process RSS, 12 GiB Go soft limit, two hours total.

```
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoding-timing/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name clocks --clocks-only --smoke
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoding-timing/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name smoke --smoke --keep-awake
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoding-timing/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name s14 --keep-awake
```

The large input cache is `.cache/liftrs/parallel-s14/s14-m2-t5/online.bin`,
SHA256 `88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8`.
Regeneration/provenance is in the [first-use report](../parallel-s14/report.md).
New binaries and smoke preprocessing remain in `.cache/liftrs/encoding-timing/`.
This is mathematical encoding only, without masking or application/storage I/O.
The experiment does not assume a cause for earlier clock discrepancies or
extrapolate unstable timings to larger fields. Environment, commands and units
are recorded in each run's JSON; native elapsed counters are not CPU percentages.


The temporary Go toolchain had been removed before this work resumed. Restored
Go 1.27.3 from `https://go.dev/dl/go1.27.3.darwin-arm64.tar.gz` (archive SHA256
`dd3908cf20959d798c1e06ebe5b72213f0505ce01373b40ce3110975a6cd99c2`), matching the
prior benchmark version. The initial sandboxed build printed a nonfatal Go
module-stat-cache write warning; subsequent monitored runs use the normal cache
permissions. Existing preprocessing was retained throughout.

The [short controls](clocks/results.json) pass: busy wall intervals are about
250 ms with about 248–250 ms CPU time; sleep intervals are about 251 ms with
0.12–0.14 ms CPU. The [small encoding run](smoke/results.json) produces six
identical complete words, with agreeing elapsed clocks and all native stages
inside their enclosing native total. [analyze.py](analyze.py) checks sources,
binary/cache/evidence hashes, timing containment, boundary counts and paired
same-clock durations. These controls establish measurement consistency for short
intervals, not the cause of the earlier large-run discrepancy.


A read-only [environment observation](environment-midrun.json), collected by
[environment.py](environment.py), found the host on battery at 18% early in the
first hybrid case, about 12 seconds after the first native case ended. Power was
not directly sampled during the first native case. Hardware is the same Apple M5 Pro, 18 logical CPUs, 48 GiB;
no thermal/performance warning was reported by `pmset -g therm` (this is not a
measurement of frequency or proof that throttling was absent). The user connected
AC during the first hybrid case. [watch-power.py](watch-power.py) records power
and current phase every 15 seconds in `s14/power-observations.jsonl`.

Retain and label the first pre-AC/cold-workspace case and the transition case;
do not pool that first pair into the steady-AC speed comparison. The remaining
four cases form two ordered pairs on AC. This exclusion follows an observed
condition change rather than a timing threshold; all results remain reported.
Power observations are periodic, so transition timing is bounded by sample times.
A case is classified AC when the samples before, during and after it all report AC.
The midrun battery observation cannot establish the power state of the historical
October runs or explain their clock discrepancies.


## Results

The [large-run results](s14/results.json) and [verified analysis](s14/analysis.json)
record all six cases. Times below use the C++ steady clock around the entire
encoding call, including Go allocation/dispatch; CPU time aggregates all threads.
Two diagnostic clock reads bracket each phase and are separate from the encoding
boundary-call counts. RSS is the whole-process high-water mark, not a per-case
peak; `getrusage.Maxrss` uses bytes on this macOS host.

| Case | Backend | Power observation | Elapsed s | Process CPU s |
|---:|---|---|---:|---:|
| 0 | Full C++ | Before AC confirmation; cold workspace, power not directly sampled | 168.947 | 1809.538 |
| 1 | Hybrid | Battery → AC during case | 132.009 | 1314.049 |
| 2 | Hybrid | AC throughout observation bracket | 93.720 | 1082.200 |
| 3 | Full C++ | AC throughout observation bracket | 84.228 | 1097.931 |
| 4 | Full C++ | AC throughout observation bracket | 84.715 | 1104.408 |
| 5 | Hybrid | AC throughout observation bracket | 95.375 | 1094.926 |

For the two complete AC pairs, medians are **84.471 s full C++** and
**94.547 s hybrid**, a descriptive ratio of **1.119×**, or **10.66% less elapsed
time**. Individual pair ratios are 1.113× and 1.126×. This is a two-pair,
warm-workspace result with profiling, not a universal speedup guarantee or an
unprofiled latency baseline. The first pair remains visible and is excluded from
the AC comparison because its power conditions were unobserved/changing. Similar aggregate CPU
costs (means 1101 s full C++ and 1089 s hybrid) do not establish reduced CPU cost;
observed effective parallelism is about 13.0 versus 11.5 CPU-seconds per elapsed
second, respectively.

All six full outputs match the retained canonical SHA256
`b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c`.
For every large encoding, Go/native/POSIX/wall elapsed intervals agree within
**0.376 ms**, and native stage sums fit the enclosing same-clock total. Thus the
historical stage/total disagreement was not reproduced in this run. This does
not identify its historical cause. The current power transition is recorded;
it cannot retroactively explain the earlier October data.

Peak process RSS is **4.5875 GiB**. Load takes 0.995 s, native setup 3.99 ms and
workspace allocation 209 ms (allocation can be lazy). Arrays and preprocessing
sizes are unchanged from the [native online experiment](../native-online/report.md).
For AC full C++ cases, mean elapsed stage counters are Taylor 21.561 s,
columns/transposes 26.678 s, rows 34.508 s, products 1.419 s, buffers 0.025 s and
assembly 0.253 s. Transform stages remain about 98% of elapsed encoding time;
these are elapsed intervals, not sampled CPU percentages. This provides a
consistent baseline before the planned s=15 preparation-memory audit, without
admitting larger parameters or repeating preprocessing.

[check.py](check.py) records focused Go race checks, native experiment builds,
vet and the compiler-free Go encoding suite in
[validation](validation/verification.json). The clock C ABI is checked under
ASan/UBSan, including null output and four-value buffer bounds; the
[final clock check](clock-validation-final/verification.json) correctly allows
wall-clock adjustments while requiring elapsed/CPU clocks not to go backward.
Go race tests do not instrument native memory. Python compilation and local
file-link checks are also required before closing this work item.

Native text samples are retained losslessly as `.txt.gz`, with original text
retained beside the cached binary. Each analysis maps original digests to gzip
files and local originals. Go CPU profiles and flat summaries remain in the run
directories. The [end observation](environment-end.json) confirms continued AC
power. Power is sampled periodically; core frequency and complete host activity
are not controlled. No output packing, cryptographic stage or storage I/O is
included in the encoding times.
