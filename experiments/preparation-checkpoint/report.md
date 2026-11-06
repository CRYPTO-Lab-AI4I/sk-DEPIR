# Bounded preparation checkpoint

The [s=14 retry](s14-retry/results.json) passes complete P/G/spectrum digest
comparisons against the retained original constants. Preparation took **34m 32.50s**
and the process peaked at **9.01 GiB RSS**. This is a diagnostic memory measurement
with profiling and forced-GC checkpoints, not an unprofiled latency benchmark.
No discovery, database encoding or s=15 guard extension was included.

## Implementation and reproduction

[run.py](run.py) builds and runs the opt-in
[TestPreparationCheckpoint](../../encoding/liftrs-shared/preparation_checkpoint_test.go).
Preparation uses compact membership bits, bounded flat factor batches with balanced
root merges, private reusable convolution buffers and fresh Newton results where
full-length. Scratch is released before spectrum construction. Prepared encoders
retain no mutable preparation scratch; concurrent encoding ownership is unchanged.

Run from the repository root:

```
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/preparation-checkpoint/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name smoke --smoke
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/preparation-checkpoint/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name s14-retry
python3 experiments/preparation-checkpoint/analyze.py s14-retry
```

Use a new run name when reproducing, preserving existing evidence. The large run
uses s=14, m=2, t=5, K=50,976,325, N=268,435,456 and polynomial 0x402b. It loads the
pinned support file instead of rediscovering it. Source revision is
`d638a9f522af8f7b0f1659cd6a7fe3c528706080` plus the exact source hashes/snapshots
in [results.json](s14-retry/results.json); the measured binary hash is also recorded.
[analyze.py](analyze.py) verifies source, binary, artifact and evidence hashes.

Environment: Apple M5 Pro, 18 logical CPUs, 48 GiB RAM, macOS 26.6.2 arm64,
Go 1.27.3 and Apple Clang 21, generic C++17 `-O3`, 16 workers. The retry remained
on AC, with no pauses. Budgets were a 24 GiB Go soft target, 32 GiB sampled RSS
guard, two hours of active compute and 20 GiB for the new local cache.

## Measurements

[Analysis](s14-retry/analysis.json), [phase records](s14-retry/phases.json),
[heap checkpoints](s14-retry/checkpoints.json) and
[RSS/power progress](s14-retry/progress.jsonl) retain the measured data.

| Quantity | Measured value |
| --- | ---: |
| Setup | 0.064 s |
| Retained support load | 0.286 s |
| Preparation, native steady clock | 2,072.501 s |
| Preparation, aggregate process CPU | 7,663.087 CPU s |
| Full constant digest verification | 0.805 s |
| Whole test process wall time | 2,075.333 s |
| Whole-process kernel peak RSS | 9.015 GiB |
| Maximum reachable Go heap at checkpoints | 5.418 GiB |
| Maximum observed allocated Go heap | 6.495 GiB |
| Cumulative preparation allocation traffic | 208.911 GiB |
| Explicit checkpoint GC time, 15 checkpoints | 0.106 s |

GiB means 2^30 bytes. Allocation traffic is cumulative, not simultaneous memory.
The allocated-heap maximum combines 500 ms samples and pre-GC checkpoints; it is
an observed maximum, not an exact high-water mark. RSS includes native memory and
runtime overhead. Reachable heap after GC is sampled only at explicit checkpoints.

| Checkpoint | Reachable Go heap, GiB | Reusable convolution buffers, GiB |
| --- | ---: | ---: |
| Before root construction | 1.382 | 0 |
| Active root FFT, 2^26 elements | 2.299 | 0.5 |
| Before reciprocal | 2.108 | 0.5 |
| Largest reciprocal/identity checkpoints | 5.418 | 2 |
| Before spectrum | 2.418 | 0 |
| Complete | 3.418 | 0 |

CPU profiles span each top-level phase; heap profiles accompany the checkpoints.
Native samples and RSS are collected separately. Lightweight root progress every
2^20 support indices causes no additional GC or heap profiles. The Go and native
steady preparation clocks differ by about 5 microseconds; wall/POSIX elapsed
clocks differ from steady by about 9 ms over the 34.5-minute preparation.

Full serialized P/spectrum and G hashes match the original files. The polynomial
identity check also passes. Constants were not duplicated: the verified originals
remain in `.cache/liftrs/parallel-s14/s14-m2-t5/`, with hashes and regeneration
instructions in the [first-use report](../parallel-s14/report.md). New binaries,
source provenance and raw native sample copies remain under
`.cache/liftrs/preparation-checkpoint/`. Profiles, compressed native samples and
all measurement records are retained here; [analysis.json](s14-retry/analysis.json)
links their hashes and retained raw locations.

## Validation and limitations

[Validation](validation/verification.json) passes focused race checks, vet and both
full Go/native suites in an isolated repository copy. Small
[initial](smoke/results.json), [final](smoke-final/results.json) and
[guarded](smoke-guarded/results.json) fixtures pass full constant comparisons and
active-buffer checks. They force the FFT cutoff to zero to exercise scratch paths;
they are correctness tests, not production timing configurations.

The historical preparation took 2,017.98 s and its full first-use process peaked
at 13.73 GiB RSS. It included discovery, differently sized support storage and no
internal forced-GC checkpoints. The new result establishes a measured memory
profile, not a controlled algorithm speedup or an isolated RSS improvement.
The 0.106 s GC figure excludes other profiling and observation overhead.

The 5.42 GiB live-heap checkpoint is consistent with the roughly 22 GiB principal
array model for s=15 in the [audit](../preparation-memory/report.md). This supports
a guarded local attempt after reviewing s=15 limits; it does not prove a 32 GiB
RSS bound. s=15 has not been admitted or run. No new database encoding was needed
for this diagnostic because all prepared constants are byte-identical.

## Power interruption and recovery

The [first large attempt](s14/results.json) stopped after about 25 minutes when
AC disappeared at 40% battery. Its controller rule was too strict; its roughly
4 GiB partial RSS is not a full preparation result. Profiles are preserved, but
root construction had not completed and its in-memory work could not be resumed.
A [battery admission attempt](battery-admission/results.json) launched no worker.

The retry uses corrected control: start on AC or at least 30% battery, continue
through ordinary power changes, suspend at 10% on battery and resume on AC.
Pauses preserve memory and exclude paused time from the compute budget; idle-sleep
prevention is disabled while paused. [Policy tests](test_power.py) and their
[verification](power-policy-verification.json) cover continuation, pause, resume
and unknown readings. Memory/cache/time guards still flush profiles on termination.
No pause occurred in the successful retry.
