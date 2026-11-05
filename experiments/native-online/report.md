# Complete native online encoding

This experiment compares the complete
C++ database-dependent path with the retained hybrid Go/C++ path, using the same
code, constants, message, worker count and machine. Reusable preprocessing stays Go.

Run [run.py](run.py) with `--backend native` or `--backend hybrid`; the exact
benchmark is [main.go](main.go). Both collect Go CPU profiles for every phase and
native samples for large encodings. Each backend runs in a fresh process.
The script records revision, source hashes (uncommitted sources retained in
`../source-snapshots/`), compiler, environment, binary digest and raw results.

```
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/native-online/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name smoke --smoke --repeats 2
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/native-online/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name s14-native --backend native
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/native-online/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name s14-hybrid --backend hybrid
```

The s=14 input is the existing verified `online.bin` under
`.cache/liftrs/parallel-s14/s14-m2-t5/` (1,277,647,153 bytes, SHA256
`88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8`).
Regeneration is documented in the [first-use experiment](../parallel-s14/report.md).
No repeated discovery/preparation is needed. New binaries and small-fixture
preprocessing remain under `.cache/liftrs/native-online/<source-id>/`.

Limits: 30-minute process timeout, 12 GiB RSS guard, 12 GiB Go soft-memory target,
16 workers on the Apple M5 Pro/48 GiB host. All encoding timings include CPU
profiling, output allocation and interface entry, but exclude load, native-plan
initialization, workspace allocation and output hashing (reported separately).
There are no large copies across the FFI. Native buffer setup, products, assembly
and the existing three matrix stages have separate elapsed counters; those are
not CPU percentages. Process peak RSS includes the loaded constants and all
encoding repeats. Full canonical hashes use seed 71 and must match retained s=14
evidence. This measures mathematical encoding, excluding masks/permutation,
application layout and packed output I/O; it is not a production security claim.


## Implementation boundary and validation

The [shared adapter](../../encoding/liftrs-shared/scaling_matrix_cpp.go) now uses
[EncodeOnline](../../encoding/liftrs-cpp/plan.go) and the
[whole-call C ABI](../../encoding/liftrs-cpp/c_api.cpp). `Encode`/`EncodeWithStats`
on the opt-in matrix encoder perform complete native online arithmetic;
`EncodeHybridWithStats` retains the previous scalar Go orchestration. The
original `WithCPPTransforms` adapter and ordinary Go API remain unchanged.

The same fixed-constant transform maps, precomputed Taylor maps, parallel native
workers and blocked transposes are retained. Native point products are chunked;
canonical assembly uses a small fixed-omega multiplication map and per-chunk
parity reduction. The small omega map and primitive-generator validation are
currently built/checked within the measured call; their cost is charged online.
No N-sized coordinate lookup table or duplicate P/spectrum allocation is added.
There is one language-boundary call instead of five; stage counters still count
five mathematical transforms. Residual outer-call time includes validation,
allocation and dispatch, and is not a calibrated measurement of cgo overhead.

[Small-instance tests](../../encoding/liftrs-shared/scaling_matrix_cpp_test.go)
compare complete words against Go/hybrid execution and the independent dense
oracle. They cover unit/zero/deterministic inputs, repeated and concurrent
workspaces, invalid input/ownership, overlap, closed handles and nonprimitive
scatter generators. The [native fixture](../../encoding/liftrs-cpp/online_test.cpp)
checks a known generator identity with one/four workers, including shared
prefix/full-domain plans, under address/undefined-behavior sanitizers.

The initial [repository validation](validation/verification.json), run by
[check.py](check.py), passes the focused Go race checks, vet, and both native and
compiler-free full suites in an isolated copy. The initial
[sanitizer verification](native-validation/verification.json) is produced by
[native-check.py](native-check.py). Native race freedom is not established by the
Go race detector or by ASan/UBSan; it relies additionally on disjoint workspace
ownership and the validated primitive-generator scatter.

## Timing anomaly and repeat methodology

The first native s=14 run produced correct complete hashes at 86.592, 159.782 and
308.252 seconds, with 4.584 GiB peak RSS. In repeat 1, sequential native stage
counters total 328.493 seconds while the outer Go interval is 159.782 seconds.
Repeat 2 totals 316.747 seconds versus 308.252 seconds. These clocks therefore
cannot all be interpreted as one consistent elapsed-time measurement.
The original [results](s14-native/results.json) and profiles are preserved.

A subsequent native/hybrid comparison uses fresh processes, two profiled repeats
per backend, and `--keep-awake` (macOS `caffeinate -i -w <benchmark-pid>`). This
prevents idle sleep for the process lifetime; it does not prove suspension was
the cause or control manual sleep, thermal state, frequency, power mode or other
host activity. Native repeat times remain much slower than the initial result,
and native stage sums still differ from the outer intervals. Do not use the
initial 86.6-second result alone as evidence of a speedup. The same executable
bytes are used for the repeated comparison; checksums identify this explicitly.

Reproduce the repeat policy by adding `--repeats 2 --keep-awake` and using
`--run-name s14-native-awake` or `--run-name s14-hybrid-awake`. The smoke runs
[smoke](smoke/results.json) and [smoke-final](smoke-final/results.json) also compare
full words with the retained Go oracle; their setup includes cache creation when
missing and is not target-scale preprocessing evidence.

[analyze.py](analyze.py) verifies recorded sources, binaries, artifacts, full
hashes, boundary counts and evidence digests, and writes `summary.json` and each
run's `analysis.json`. Large native text samples are retained unchanged beside
the local binary and stored losslessly as `.txt.gz` in the report directories;
`analysis.json` maps the original raw SHA256 to the compressed file and local
original. The raw result manifests remain unchanged. Run:

```
python3 experiments/native-online/analyze.py
```

Fresh processes do not imply cold filesystem caches. Timings include profiling
and are not unprofiled latency baselines. This experiment does not establish the
cause of the host/timer variation, large-s feasibility, public artifact validity,
or application/cryptographic encoding performance.


The earlier [parallel-encoding](../parallel-encoding/report.md) and
[parallel-s14](../parallel-s14/report.md) harnesses now explicitly call the retained
hybrid method so rerunning those scripts preserves their documented measurement
boundary. Their historical source manifests/results are unchanged; this
experiment alone compares the new complete native path.


## Recorded s=14 results

The [summary](summary.json) is calculated from the raw phase records below.
All seven complete canonical outputs match SHA256
`b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c`.
The input is K=50,976,325 elements, the output N=268,435,456 elements,
and the code is m=2,t=5 over GF(2^14). Both repeat-policy backends use the
identical executable SHA256
`3d8892a1ac1e65d82ecee2080b37fb4bb84464cdf4ceed65ae8a8353285d0f87`.
Hardware inventory was checked with `sysctl`: Apple M5 Pro, 18 logical CPUs,
51,539,607,552 bytes RAM (48 GiB); local macOS/APFS environment as recorded.

| Run | Profiled encoding seconds | Median seconds | Process peak GiB |
|---|---|---:|---:|
| [Initial complete C++](s14-native/results.json) | 86.592, 159.782, 308.252 | 159.782 | 4.584 |
| [Complete C++, idle sleep prevented](s14-native-awake/results.json) | 312.683, 327.190 | 319.936 | 4.588 |
| [Hybrid, idle sleep prevented](s14-hybrid-awake/results.json) | 346.112, 336.556 | 341.334 | 4.587 |

The repeat-policy comparison has a **descriptive** median ratio of 1.067 and
6.27% less outer-call time for complete C++. This is not a dependable speedup
claim: the native stage sums are 325.002/335.374 seconds versus outer intervals
312.683/327.190, and the earlier result was much faster. The analyzer flags a
native stage sum exceeding the enclosing interval by more than 1%; this is a
consistency diagnostic, not an explanation or corrected latency estimate. Hybrid
stage sums of 321.478/321.659 seconds omit its Go scalar/buffer/assembly work and
are not full encoding times. No outlier is silently dropped.

Native repeat-policy mean stage counters are Taylor 83.845 s, columns/transposes
103.113 s, rows 136.022 s, buffers 0.153 s, products 6.120 s and assembly 0.934 s.
These are recorded clock intervals with the discrepancy above, not CPU shares;
do not sum them into a corrected outer-call measurement. They identify transform
execution as the main remaining work but do not establish an optimization order.

Load/native-plan setup/workspace allocation take 1.615 s/6.37 ms/0.90 ms for the
native repeat process and 1.703 s/6.34 ms/290.38 ms for hybrid. Allocation is lazy
and does not guarantee all pages are touched at allocation time. The same arrays
are used by both: 1,277,647,128 bytes P+spectrum, 2,351,388,948 bytes reusable
workspace, 203,905,300 bytes input and 1,073,741,824 bytes output. Process RSS counts
native metadata/thread memory as well as Go payloads; there is no added large
preprocessing copy. Neither load measurement is a cold-disk benchmark.

The reusable artifacts, binaries, original native samples and small smoke cache
are retained locally. Compressed native samples and every Go profile/flat summary
are linked through each run's `analysis.json` and the raw result manifest.
Do not change the larger-size forecasts using these unstable timings.


Final source-matched checks pass in
[validation-final](validation-final/verification.json) and
[native-validation-final](native-validation-final/verification.json), using the
commands in [check.py](check.py) and [native-check.py](native-check.py) with the
corresponding `--run-name` values. The final run includes the explicit historical
hybrid harness calls and the readable C++ source. The
[verification audit](verification-analysis.json) resolves available source
manifests for the verification runs as well as the benchmark evidence.
