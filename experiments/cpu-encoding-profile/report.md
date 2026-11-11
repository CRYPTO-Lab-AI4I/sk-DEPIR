# Cached s14 CPU encoding profile

[Runner](run.py), [Slurm launcher](run.sbatch), [analyzer](analyze.py) and short
[evidence self-check](check.py). Reuse the existing
[encoding probe](../encoding-timing/main.go) unchanged. Parent revision980aef4;
each run records the exact revision, source hashes/dirty snapshots, toolchain,
CPU/node/allocation, input/cache/binary hashes and commands.
[Profile-01 results](runs/profile-01/results.json) and
[verified analysis](runs/profile-01/analysis.json), job72955, now pass.

```sh
python3 -B experiments/cpu-encoding-profile/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-encoding-profile/slurm-%j.log experiments/cpu-encoding-profile/run.sbatch profile-01
# After the user's completion ping:
python3 -B experiments/cpu-encoding-profile/analyze.py profile-01
```

## Purpose and method

Identify which stages warrant CPU optimization before a full s16 database encode.
Use cached s14,m2,t5:K50,976,325,N268,435,456. Verify the original read-only
`.cache/liftrs/parallel-s14/s14-m2-t5/online.bin` against SHA256
`88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8`,
then retain a verified staging copy under `.cache/liftrs/cpu-encoding-profile/RUN`.
There is no discovery or preparation. A tiny s8 integration smoke first exercises
the same existing probe; its small generated cache is retained with hashes.

Run complete native encoding twice each at8,16,32 threads in that fixed order.
The database seed is71. Each full canonical output must match retained SHA256
`b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c`.
The native encoder and mathematical behavior are unchanged. The existing probe
checks native stage totals against an enclosing native steady clock. Timings
include output allocation and five transforms; hash verification is separate.
CPU profiles cover load, setup/workspace, encoding and hash phases. Retain all
profiles; generate pprof summaries for six encoding phases.

Stage counters separate Taylor conversion, columns with transposes, rows,
native buffer work, field point products and final canonical assembly. Columns
and transposes remain combined; Go output allocation is inside encoding wall time
but outside native stage counters. GPU kernels are not used. Existing host perf
permissions restrict native instruction sampling; the runner records
perf_event_paranoid. Go profiles may mostly show Cgo/unsymbolized native work.
This run provides stage and worker scaling, not a native instruction hotspot map.
No packing, masking, storage upload or complete application layout is timed.

Two samples in fixed order are a profiling screen, not a robust latency benchmark;
cache warming, worker order, memory/NUMA effects and shared-node load can matter.
It cannot reproduce s16's16x larger full-domain footprint. Choose optimizations
from measured stage costs and validate larger-size behavior separately.

## Resources, retention and handoff

One CPU-only Slurm job in gpuq (no GPU requested),32 CPU cores,16 GiB requested
host memory,12 GiB sampled process-group RSS guard,30-minute allocation and
25-minute controller deadline. Existing probe Go soft target12 GiB is not a hard
RSS limit. New durable cache is capped at8 GiB; original caches remain unchanged.
Retain reusable staging data, binary, exact source snapshots and all completed
profiles. Controller TERM/KILL stops guarded overruns. The reused standalone
probe does not install a profile-flush TERM handler: completed phase profiles
remain valid, but an interrupted active profile may be incomplete. Preserve such
attempts and do not treat them as successful latency samples.

Go1.27.3/pinned dependencies and GCC14.2.0 modules are reused. Source/build cache
is isolated in node-local scratch; stage/checksum and compilation time are separate
from measured encoding. Report sampled RSS, input staging, setup and per-stage
seconds after analysis. No performance conclusion is recorded before completion.

Submit and hand off immediately under AGENTS rule8. Do not poll status, follow
logs or wait for completion; the controller's automatic limits remain active.
Dependent analysis and any optimization wait for the user's completion ping.


## Completed profile and optimization priority

Job72955 ran on gnode02, Xeon6710E, GCC14.2.0/Go1.27.3 (exact environment in
run logs). The analyzer verifies286 source files, all sealed evidence/cache hashes,
six complete canonical outputs, five-transform/one-boundary native accounting,
and enclosing-clock bounds. Input staging and digest verification takes15.812 s,
outside encoding. Sampled process-group RSS peaks at5.594 GiB, below12 GiB.
Original preprocessing is unchanged and staged cache/binary/profiles are retained.

Two profiled samples per worker count; medians in seconds:

| Workers | Complete encoding | Taylor | Columns/transposes | Rows | Buffer | Pointwise | Assembly |
|---|---:|---:|---:|---:|---:|---:|---:|
| 8 | 248.675 | 49.921 | 84.324 | 107.436 | 0.444 | 5.237 | 1.177 |
| 16 | 140.746 | 40.937 | 42.153 | 53.884 | 0.320 | 2.631 | 0.519 |
| 32 | 91.164 | 38.933 | 22.189 | 27.613 | 0.324 | 1.474 | 0.312 |

Complete32-worker encoding is1.54x faster than16 and2.73x faster than8 in this
fixed-order screen. Rows and columns/transposes improve about3.9x/3.8x from8 to32,
while Taylor improves only1.28x. At32 workers Taylor is42.7% of total elapsed,
rows30.3% and columns/transposes24.3%. Native buffer, pointwise and assembly sum
to2.3%; total elapsed additionally includes Go output allocation and wrapper work.
Do not infer that all memory costs are only2.3%: accesses/transposes are also
inside the transform stages, and columns/transposes are combined.

The [native Taylor implementation](../../encoding/liftrs-cpp/c_api.cpp) parallelizes
independent blocks while preserving a sequential recurrence inside each block.
At high levels only one or a few blocks exist, limiting useful threads. This
source-based explanation is consistent with the scaling result, not an
instruction-sampling attribution. The [32-worker Go profile](runs/profile-01/pprof-encode-w32-r0-c0-native.log)
remains mostly unsymbolized native work and Cgo. perf_event_paranoid is4; no native
instruction hotspot claim is supported.

Prioritize CPU Taylor conversion: investigate within-block parallelism using the
already validated two-pass gather algebra from the optional GPU composition,
while retaining the current sequential recurrence as a correctness reference.
Verify full forward/inverse transforms and canonical words before selecting a
change. Separate column arithmetic from transpose time only if later evidence
requires that distinction; rows/columns are substantial but already scale well.
Do not prioritize buffer clearing, scalar point products or final assembly from
this profile. These are candidate changes, not implemented optimizations.

As a bound, halving Taylor alone at32 workers would reduce91.16 s to about71.70 s
(1.27x speedup), if all other costs stayed fixed. This is an illustrative arithmetic
scenario, not a measured gain or s16 projection. No new optimization/timing job
was launched after analysis. A larger s16 encode still needs its own memory and
correctness validation; this profile covers one complete cached s14 mathematical
encoding only, with preparation/packing/application work excluded.
