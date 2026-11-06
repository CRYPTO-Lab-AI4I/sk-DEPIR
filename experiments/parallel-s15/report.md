# Guarded local s=15 first use

The [local s=15 run](initial/results.json) and [fresh cached run](cached/results.json)
**pass**. Median native encoding is 393.19 s; fresh cached encoding is 393.69 s
with 18.46 GiB peak RSS. First use peaks at 20.55 GiB. This demonstrates
standalone s=15 feasibility on the local 48 GiB host.
[run.py](run.py) drives [TestS15FirstUse](../../encoding/liftrs-shared/s15_firstuse_probe_test.go).
The starting point is the verified [s=14 bounded preparation](../preparation-checkpoint/report.md).

## Scope and budgets

Use GF(2^15), m=2, t=5, N=2^30; K is counted exactly during discovery, not taken
from the catalogue's Monte Carlo estimate. The code-preprocessing stages run in
Go with native transforms; all database-dependent encoding runs in C++.
The ordinary public discovery limit remains s≤13. Experimental import and
preparation/loader limits are extended only through s=15 after small-buffer tests.

On this Apple M5 Pro (18 logical CPUs, 48 GiB, macOS 26.6.2 arm64), use Go 1.27.3,
Apple Clang 21, generic C++17 `-O3`, 16 workers, a 24 GiB Go soft memory target,
32 GiB sampled process RSS guard and 20 GiB new-cache budget. Require at least
30 GiB free disk before starting. Allow four hours active compute: extrapolating
the measured s=14 preparation suggests more than the earlier two-hour ballpark.
The current run was subsequently extended to 5.5 hours as described below.
This is an estimate, not an s=15 timing claim. All policy values and power samples
are recorded in results.json. Battery ≤10% suspends work until AC returns;
ordinary power changes do not kill the process, and pauses do not consume the
active-time allowance. The Go test has a longer wall timeout to permit pauses.

## Reproduction and evidence

From the repository root (use a new run name to preserve old evidence):

```
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/parallel-s15/check.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/parallel-s15/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name smoke --smoke
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/parallel-s15/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name initial
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/parallel-s15/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name cached --online-only
python3 experiments/parallel-s15/analyze.py initial
python3 experiments/parallel-s15/analyze.py cached
```

Each run records revision, exact measured source hashes and content-addressed
snapshots, compiler/toolchain, command, binary hash, phase CPU profiles, native
samples, process RSS and power history. Large data is kept outside Git in
`.cache/liftrs/parallel-s15/s15-m2-t5/` (smoke uses s6-m2-t4). Support and prepared
constants are written atomically with checksums and reused on restart. A killed
incomplete preparation must be recomputed; completed discovery remains cached.
No old cache is deleted. Profiles and reports stay in this experiment directory.

## Method and checks

Separate setup, exact discovery, support persistence, preparation, prepared-data
persistence/release, trusted-hash loading, online workspace setup, three encodings,
verification, output packing/readback and one fresh cached encoding. Native and Go
phase clocks are recorded separately from aggregate CPU. CPU profiles stay enabled.
Heap checkpoints in preparation force GC, retaining stage-specific reachable heap
and scratch capacity. Those times include diagnostic overhead and are not pure
latency comparisons with older runs. RSS is sampled every two seconds; kernel peak
RSS is collected on exit. Heap checkpoints are not continuous peak measurements.

All message elements are checked at systematic coordinates, 16 degree-t curve
restrictions use independent carryless base multiplication, repeated canonical
word hashes must agree, and 1,027 packed positions are read back. Small tests
compare native arithmetic and encoding with existing independent reference paths.
Digest equality across repeats is reproducibility evidence, not an independent
proof of the entire huge codeword. Fresh cached execution must reproduce the
first-use canonical hash. Discovery's exact count remains separate from the
paper/catalogue estimates. No server computation is introduced.

## Automatic completion

[finish.py](finish.py) waits for the already running initial worker, analyzes its
terminal evidence, and starts a fresh cached process only after first use passes.
It refuses to overwrite existing cached results. Run with the same GOCACHE and
`--go` path as above. Its [summary metadata](summary.json) records current status;
on completion it writes `summary.md` and indexes a timestamped result message on
the board. A failed or guard-stopped first use is reported as incomplete and is
not automatically retried. This supervisor does not change current support claims.

## Results

[Admission validation](validation/verification.json) passes focused race checks,
vet and both full Go/native suites in an isolated copy. The small
[first-use](smoke/analysis.json) and [fresh cached](smoke-cached/analysis.json)
fixtures pass, with identical canonical word hashes, systematic/curve checks and
packed readback. These are small correctness fixtures, not s=15 measurements.
A test-harness compile error in the new clock call was corrected before validation;
no failed large attempt was launched.

The [large run](initial/results.json) started on AC at 100% battery, using revision
`3497bbd` plus its recorded source snapshots. Exact discovery and preparation completed. Its
[phase log](initial/events.log), [state](initial/state.json) and
[resource/power progress](initial/progress.jsonl) preserve the historical running observations. Full
[analysis](summary.md) passes, including all first-use and cached checks. No
application/cryptographic integration or production-security claim is added.


## Historical live time-guard extension

At the user's request, [extend_guard.py](extend_guard.py) increased the **current
initial run** from four hours to **5.5 active hours (19,800 s)** without restarting
worker PID 84546. [Override metadata](guard-extension/status.json) records the
new policy, source hash and transfer time. The original results.json policy and
source remain historical launch provenance; this override is authoritative for
the extended deadline. Fresh cached execution keeps its original independent
budget. RAM, cache and Go soft limits remain 32, 20 and 24 GiB respectively.

The original monitor had a compiled fixed deadline. Only that monitor was
suspended; the worker continued under replacement RSS/cache/power/deadline
monitoring. On worker exit the original parent resumes to collect wait4/kernel
peak RSS and run the existing finalization/follow-up pipeline. A guard stop is
recorded in the override metadata even if the old parent labels the nonzero exit
as failed. [New progress](guard-extension/progress.jsonl) and native samples are
separate from the original progress file, which pauses during the transfer.
The worker's CPU/heap profiling continues unchanged. The transferred elapsed
clock uses ps with one-second resolution; battery pauses remain excluded.

[Policy checks](test_extend_guard.py) cover elapsed-time parsing and battery
behavior; PID/parent relationships and the continued running worker were verified
at activation. The [root README](../../README.md) was checked; its capability
claims and resource measurements need no change for a time-only override.


## Manual sleep break

The user requested a pause at 2026-11-06 11:58 UTC. Worker 84546 and active
monitor 89916 were **suspended**, with idle-sleep prevention disabled and state
retained in RAM. [Pause metadata](manual-pause/status.json) captures 10,579 seconds
of consumed active time (one-second resolution). The then-unfinished preparation was
not serialized; normal sleep may preserve it, but restarting/shutting down or
terminating these processes loses that in-memory work. Existing caches are intact.

Resume only when the user asks, using a new enforcing monitor while preserving
the suspended parent-monitor chain for final process accounting:

```
python3 experiments/parallel-s15/extend_guard.py \
 --worker 84546 --monitor 84443 --hours 5.5 \
 --resume-from experiments/parallel-s15/manual-pause/status.json
```

The command verifies paused identities, starts the active clock from the captured
10,579 seconds, resumes the same worker, restores sleep prevention and records
its new `guard-resumed-<timestamp>/` location in the pause metadata. All intervening
awake/sleep break time is excluded from the guard; 9,221 active seconds remain.
Do not simply send SIGCONT to the old monitor, which would charge awake break
time. The old active monitor source is preserved in the content-addressed
[snapshot](../source-snapshots/bbd23c73d5d0403ac13c7fb7bbb11640db143481f76389a77006cefaac81b59a.txt).
The original four-hour monitor remains suspended until final collection.

Preparation timing spans this break, so its raw phase elapsed time must be labeled
as interrupted and cannot be used as an uninterrupted latency baseline. At that pause, online
encoding had not begun. Policy tests cover resumed active duration across different
clock origins and long breaks; resume was not exercised until the user subsequently requested it. Root README checked: no capability claim
changes for a pause. No automatic AC-triggered resume is allowed for this manual break.


### Resumed after manual break

At 2026-11-06 14:39 UTC the user explicitly requested resume. The same worker
84546 resumed on confirmed AC (100% battery). The new
[enforcing monitor](guard-resumed-20261106-143915/status.json), PID 99476, begins
from 10,579 consumed active seconds; the roughly 2h 41m wall-clock break is excluded
from the 19,800-second budget. Its
[progress log](guard-resumed-20261106-143915/progress.jsonl) records post-resume monitoring.
The previous monitor chain resumed after worker exit for orderly final statistics.
Sleep prevention is restored. No database state was regenerated or discarded.
The recorded resume command has already run; do not execute it again while active.
Preparation remains an interrupted timing measurement; later online phases are
separate. Existing CPU/heap profiles and automatic cached follow-up are preserved.


## Completed results and catalogue

[Summary](summary.md), [first-use analysis](initial/analysis.json) and
[cached analysis](cached/analysis.json) verify measured source/binary/evidence
hashes and retained code files. All three first-use words and the fresh cached
word share canonical SHA-256
`5c09282e5cd2a77bd9541b5023a487f8deaf794a516e7e6a77007e7f9d08b265`.

| Quantity | Result |
| --- | ---: |
| Exact K / N | 216,914,047 / 1,073,741,824 |
| Code rate | 0.2020169487 |
| First-use encoding repeats | 391.355 / 393.186 / 396.485 s |
| Fresh cached encoding | 393.690 s |
| Whole-process peak RSS, first use / cached | 20.546 / 18.461 GiB |
| Minimal P/spectrum online file | 4.808 GiB |
| Support + P/spectrum + G files | 8.808 GiB |

Discovery took 1,131.46 s. Raw preparation time spans the manual pause and is not
an uninterrupted setup baseline. Monitor observations put the complete initial
experiment at approximately 3h 33m active time, including three encodings/checks;
the last monitor sample is not an exact active-time high-water mark. Original
results.json reports the launch policy and raw clock boundaries; extension/pause
metadata separately records the effective policy and break. Its `paused_s=0` field
covers only the original monitor's automatic pauses, not the manual break.

[record-catalogue.py](record-catalogue.py) records exact standalone support only
after both analyses pass; run it after the analyzer commands above. The catalogue
retains historical MC K=216,920,545 and paper evidence. Exact K is lower by 6,498
(0.003%). Historical planner tables are not regenerated or silently changed.
Public discovery remains s≤13; the s=15 path uses experimental preparation and
trusted-hash loading, with no application wrapper or cryptographic integration.
The next design gate is the s=16 simultaneous-memory schedule and cardinality
representation, rather than an unguarded larger run.

## Publication verification and completion ordering

During publication, repeated analysis exposed that the first-use manifest had
captured an early derived `analysis.json`: the supervisor observed `passed`
before profile/evidence finalization completed. The original manifest is retained.
Its exact captured analysis bytes are preserved in
[analysis-at-finalization.json](initial/analysis-at-finalization.json), reconstructed
from the saved supervisor summary and verified against the original recorded hash.
The analyzer validates that historical copy while regenerating current analysis;
no raw measurement or output digest was changed. Future runs publish `finalizing`
until evidence hashes are sealed, and the supervisor waits for both running and
finalizing states. Derived analysis is excluded from future evidence manifests.
A new small publication fixture checks the corrected runner/analyzer path.

[archive-monitor-samples.py](archive-monitor-samples.py) compresses monitor-native
samples losslessly, verifies every original digest and preserves raw copies
outside Git. [Archive manifest](monitor-sample-archive.json) links compressed
samples and local raw paths; CPU and heap profiles remain unchanged.
