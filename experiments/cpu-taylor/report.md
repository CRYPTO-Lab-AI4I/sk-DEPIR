# CPU Taylor parallelism inside blocks

[Runner](run.py), [launcher](run.sbatch), [analyzer](analyze.py), short
[evidence self-check](check.py). [Taylor-01 results](runs/taylor-01/results.json) and
[verified analysis](runs/taylor-01/analysis.json), job73099, pass. The hybrid
within-block scheduling is selected for the native CPU path on this evidence;
larger-size behavior remains unverified.
Baseline is published revisione29258d. The runner copies the same candidate
source tree for both binaries, replacing only baseline c_api.cpp with that exact
published blob. It records both hashes plus all measured source snapshots,
binary/cache checksums, CPU/compiler/toolchain, allocation and commands.

```sh
python3 -B experiments/cpu-taylor/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-taylor/slurm-%j.log experiments/cpu-taylor/run.sbatch taylor-01
# After the user's completion ping:
python3 -B experiments/cpu-taylor/analyze.py taylor-01
```

## Change and correctness

The [native implementation](../../encoding/liftrs-cpp/c_api.cpp) retains the
original quotient-block recurrence for one worker, fewer than32768 values or
sufficient coarse blocks. If blocks are fewer than workers, a two-pass gather
exposes independent coefficient ranges inside each block. Dynamic chunks have
8192 coefficients; this is a provisional scheduling choice to measure.

Every lower degree d in a linearized polynomial is at most half/2. Quotient
correction q[j] reads only untouched upper quotient entries q[half+j-d] where
d>j. Only j<max_degree can change. After a join, each remainder coefficient reads
the completed quotient. Inverse conversion reverses these two self-inverse XOR
passes. Destinations have single writers; source/destination ranges are disjoint
within each pass. Existing fixed-constant maps are reused. No new field-value
buffer, public API, serialized format, field basis or output order is introduced.

[Native test](../../encoding/liftrs-cpp/taylor_parallel_test.cpp) includes the
implementation to exercise private scheduling without a test-only ABI. Compare
forward and independent inverse inputs with sequential polynomial division using
generic field multiplication, plus round trips. Cases cover base/extension
coefficients, zero/one/general constants, varied split sizes, below/at/above
admission, and1/2/8/32 workers. It runs with ASan/UBSan and ThreadSanitizer; a
specific unsupported TSAN memory-mapping startup failure is recorded as
unavailable, never as a passing race check. Other TSAN failures stop the job.

[Matrix tests](../../encoding/liftrs-shared/taylor_parallel_test.go) compare complete
forward/inverse transforms against the unchanged single-worker recurrence at
s8/s16, dimensions14/15/16 and2/8/32 workers. Existing small Horner/reference,
full Go/native suites, focused Go race/vet and a legacy RM test prerequisite run
before performance comparisons. Go race instrumentation does not cover C++:
separate native sanitizer evidence and the disjoint-write proof matter.

## Matched measurements and boundaries

Reuse [existing probe](../encoding-timing/main.go) unchanged. Load a verified
read-only retained s14,m2,t5 cache (K50,976,325,N2^28) into a durable staging copy.
Digest matches88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8.
Both old/new binaries first run a small s8 independent oracle check. Run two
complete s14 native encodings per binary at16 and32 workers (eight encodings).
First repeat runs old then new; second reverses that order. Each is a fresh
process, includes separately profiled loading/setup/workspace, uses message
seed71, and must reproduce full canonical hash
b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c.
No code discovery/preparation is repeated for s14. Preserve all existing caches.

Measure complete encoding and native Taylor/columns-transposes/rows/buffer/
pointwise/assembly seconds, memory and CPU profiles. Encoding includes output
allocation; validation/hash is separate. Columns and transposes stay combined.
Host perf restrictions limit native instruction sampling; Go profiles may mainly
show Cgo/unsymbolized native work. Fixed-order two-sample medians screen for
regressions and whole-encoding benefit; they are not a robust latency distribution
or an s16 forecast. More passes/scheduling or memory access can offset parallel
benefits; do not claim a gain before results.

## Resource policy and handoff

One CPU-only32-core job,16 GiB host request/12 GiB sampled process-group RSS,
30-minute allocation/25-minute controller,8 GiB durable cache cap. No GPU or
s16 encoding. Probe Go soft target12 GiB is not a hard memory bound. Controller
TERM/KILL enforces guards; completed profiles survive but an active standalone
probe profile may be incomplete on interruption. Go1.27.3 and GCC14.2.0,
C++17/-O3 for probes; sanitizer unit builds use-O1/-g and their named sanitizers.
Exact environment is in run logs, not assumed from the login node.

Retain verified input, old/new binaries, sanitizer binaries and smoke caches under
`.cache/liftrs/cpu-taylor/RUN`; record all hashes/size. Sources/CPU profiles/logs
and results are retained in this experiment or source-snapshots. No masks,
packing, storage upload or application/cryptographic layout is part of this
mathematical encoding comparison. Select the change only after full correctness
and end-to-end timing review; normal CPU builds require no new dependency.

Submit then hand off immediately under AGENTS rule8. Wait for the user's
completion ping before any scheduler/log monitoring, analysis or dependent work.


## Completed comparison and selection

Job73099 ran on gnode17 with the compiler/CPU/affinity recorded in its logs.
The analyzer verifies292 exact source files, all sealed evidence/cache hashes,
eight full s14 canonical hashes and enclosing native clock bounds. The recorded
baseline native source hash matches the exact e29258d blob. Complete Go/native
suites, focused Go race/vet, both s8 oracle checks, ASan/UBSan and ThreadSanitizer
pass. Both native sanitizer runs pass72 independent Taylor cases; TSAN is a real
passing run, not the documented runtime-unavailable fallback. These checks cover
the exercised cases, not a general proof of race freedom for every caller.

Two interleaved profiled samples per worker/variant; medians in seconds:

| Workers | Old encoding | New encoding | Encoding speedup | Old Taylor | New Taylor | Taylor speedup |
|---|---:|---:|---:|---:|---:|---:|
| 16 | 142.178 | 118.538 | 1.20x | 39.431 | 15.430 | 2.56x |
| 32 | 87.928 | 61.042 | 1.44x | 35.675 | 8.768 | 4.07x |

At32 threads, encoding time drops30.6%; at16 it drops16.6%. Individual complete
samples are142.189/142.167 vs118.873/118.203 s at16, and88.165/87.690 vs
61.205/60.880 s at32. Other stages stay close: at32 workers, old/new rows are
27.084/27.003 s and columns/transposes22.626/22.717 s. The gain tracks Taylor
rather than an unrelated row/column change. Peak sampled process RSS across the
whole validation/comparison job is4.593 GiB, below12 GiB. No extra field-value
buffer is introduced; thread/task metadata remains small and temporary.

Select the hybrid gather/recurrence implementation. Threshold32768 values and
chunk8192 are the measured initial choices, not universally optimal constants.
The current one-worker/small/coarse-parallel reference remains in code. The
comparison demonstrates a whole-encoding CPU gain on this cached s14 instance;
no GPU or special compiler extension is required. Rows now account for44.2% and
columns/transposes37.2% at32 threads; Taylor is14.4%. Do not extrapolate those
shares or the1.44x ratio directly to s16 memory pressure or field/basis changes.

Original and measured staging caches, both probe binaries, sanitizer binaries,
source snapshots and all profiles remain retained at the locations in the result
manifest. The source/API/default-format behavior is unchanged apart from native
scheduling. Next gate is a separately budgeted cached s16 encoding and its full
correctness checks; no further timing or hardware job was launched during result
reconciliation. Script self-check, syntax/links and whitespace pass. More robust
latency distributions, native instruction attribution, s16 encoding and
application/cryptographic integration remain outside this measurement.
