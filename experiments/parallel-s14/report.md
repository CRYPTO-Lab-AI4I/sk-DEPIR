# Profiled s=14 first-use and cached encoding

The run completed on the local 48 GiB Apple M5 Pro. At `m=2,s=14,t=5`, exact
capacity is **K=50,976,325**, with encoded length **N=268,435,456** and code rate **0.189901609**.
Three profiled encodings have a **90.53 s median**. A fresh cached process encodes
in **92.12 s**, with **4.59 GiB process peak RSS**. Full first-use peak is
**13.73 GiB**. These measure one field-element database, not an application layout.

## Results

| Operation | Measured time |
|---|---:|
| Exact discovery, two samples | 254.898 s (4 min 15 s) |
| Reusable preparation, including full P·G identity check | 2,017.985 s (33 min 38 s) |
| Save basis | 0.157 s |
| Save P/spectrum and G | 1.208 s |
| Load minimal online plan | 0.687 s; fresh process 0.597 s |
| First-use encoding repeats | 90.521 / 90.533 / 91.987 s |
| Fresh cached encoding, one repeat | 92.118 s |
| Full systematic + 16 curve checks + canonical hash | 3.21–3.29 s per output |
| Packed output write, hash and fsync | 1.36 s |

Discovery + preparation + executor/workspace setup + one median encoding totals
**2,363.53 s (39 min 24 s)**, excluding artifact save/load and verification.
The complete first-use experiment takes 2,564.96 s (42 min 45 s), including three
encodings, profiling overhead, saves/loads and verification. These boundaries
explain why the experiment wall time is longer than the first-use subtotal.

The original process retains high RSS after explicit preparation release; an
[auxiliary sampler](sample-online.py), [results](online-resource-samples/results.json)
and [samples](online-resource-samples/samples.jsonl) document this. It is not a
fresh-process online RAM bound. A separate cached-only process establishes the
4.59 GiB figure without repeating discovery or preparation. Its plan arrays occupy
1,277,647,128 bytes, workspace 2,351,388,948 bytes, input 203,905,300 bytes and
canonical output 1,073,741,824 bytes. Array payloads and process RSS remain distinct.

All four full output hashes match:
`b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c`.
Each output passes every systematic symbol and 16 randomly chosen degree-5 curve
restrictions, evaluated with independent carryless base-field arithmetic. Packed
output is **469,762,048 bytes (448 MiB)** with 1,027 readback positions checked.
Packed input capacity is 89,208,569 bytes; storage factor N/K is about 5.266×.
No full s=14 Go encoding or dense oracle is run. Small fixtures compare native
preparation with Go P/G/spectra and words; s=14 small affine transforms agree with
Horner/inversion, and independent native arithmetic/table sanitizers cover s=14.

## Scripts and evidence

- [Runner](run.py), [first-use/cached probe](../../encoding/liftrs-shared/parallel_firstuse_probe_test.go),
  [analyzer](analyze.py), [validation script](check.py).
- [Initial manifest](initial/results.json), [phases](initial/phases.json),
  [table](initial/table.md), [progress](initial/progress.jsonl), [events](initial/events.log).
- [Fresh cached manifest](cached/results.json), [phases](cached/phases.json),
  [table](cached/table.md).
- [Preparation CPU profile](initial/preparation.pprof), [flat view](initial/preparation-flat.txt),
  [early sample](initial/preparation-native.txt),
  [product-construction sample](initial/preparation-late-native.txt),
  [large-product/reciprocal sample](initial/preparation-large-native.txt).
  Each timed phase has its own `.pprof` and flat view beside the phase results.
  Native samples are retained for discovery, preparation and each long encoding.
- [Admission checks](validation/verification.json), [log](validation/checks.log),
  [final verification](final-validation/verification.json), [final log](final-validation/checks.log).

```sh
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
 python3 experiments/parallel-s14/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name initial
# Reuse verified cached constants in a fresh process:
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
 python3 experiments/parallel-s14/run.py \
 --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name cached --online-only
python3 experiments/parallel-s14/analyze.py initial
python3 experiments/parallel-s14/analyze.py cached
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
 python3 experiments/parallel-s14/check.py --go /tmp/sk-depir-go-toolchain/go/bin/go
```

Choose fresh result-directory names; the runner refuses existing names. Code caches
are reused after checksum/spec checks. To reproduce fresh discovery/preparation,
move the existing cache aside and preserve it, rather than silently deleting it.
Small end-to-end checks use `--smoke`; cached smoke adds `--online-only`.

## Environment, method and retention

Source: revision `9ac6ae8` plus manifest-hashed changes. Elapsed timings use
monotonic clocks; timestamps are wall-clock metadata. Intermediate measured
source versions remain in content-addressed snapshots when the cached-only harness
is added afterward. Environment: Apple M5 Pro, 18 logical CPUs, 48 GiB,
macOS 26.6.2 arm64, Go 1.27.3, Apple Clang 21, generic C++17 -O3 and 16 native
workers. Platform/compiler strings and binary/source hashes are in the manifests.
No affinity, exclusive-machine reservation or frequency control. A short cached
smoke/build overlapped the final first-use repeat; do not interpret its ~1.5 s
variation as an algorithm effect. All timing results are profiled.

Public discovery defaults remain s≤13. The [private preparation executor](../../encoding/liftrs-shared/scaling_prepare_cpp.go)
constructs sparse metadata and calls the [same exact discovery core](../../encoding/liftrs-shared/orbit.go)
with native base/extension FFTs.
Go orchestrates the [paired/root-only preparation](../../encoding/liftrs-shared/scaling_experimental.go)
and reciprocal G described in the [algorithm review](../../docs/encoding-algorithms.md); native matrix
transforms reuse code-only plans and one transpose buffer over every convolution
length. This changes execution, not the code or information set. It does not use
the obsolete complete-tree or dense-transform-setup schedule.

Policy: 24 GiB Go soft target, sampled 32 GiB RSS guard, 20 GiB cache budget and
two-hour case timeout. These are experiment controls, not OS caps or guarantees
for other machines. Process RSS uses wait4; the live guard polls ps. CPU profiles
span every phase, including discovery/preparation, save/load, encoding and packing.
Automatic native samples start after ten seconds in long phases; two additional
preparation samples cover later operations. Native worker CPU time exceeds wall
time: preparation records 5,780.89 sampled CPU seconds over 2,018 s wall time.
Flat percentages therefore are CPU fractions, not elapsed-time fractions. Go
scalar/Karatsuba work remains substantial; native frames also appear as `[probe]`
and require the retained native samples for attribution.

Cache: `.cache/liftrs/parallel-s14/s14-m2-t5/`:

| File | Bytes | Contents |
|---|---:|---|
| `support.bin` | 203,905,300 | Sorted public degree support |
| `online.bin` | 1,277,647,153 | Experimental header, P and generator spectrum |
| `g.bin` | 869,836,524 | Generator polynomial |
| `word-packed.bin` | 469,762,048 | Public seed-71 synthetic encoded output |

The three code-only files total **2.19 GiB**, excluding small metadata/binaries.
The online file is **1.19 GiB**. Metadata records each SHA-256, parameters and
identity checks; binaries and exact source provenance are under `builds/<source-id>/`.
The original s=13 caches are unchanged. Packing is separate from encoding and
writes/fsyncs a real file; no durable-device throughput guarantee follows from
this short write. No cryptographic masks/permutation or application-plan execution.
Checksums identify artifacts; they do not independently prove mathematical validity.

## Validation and limitations

The [fresh smoke](smoke-02/results.json), [resume smoke](smoke-03/results.json), and
[cached-only smoke](smoke-cached-02/results.json) pass. The first
[smoke attempt](smoke/results.json) was stopped by sandbox denial of process-RSS
access before discovery. The first [cached-only build](smoke-cached/results.json)
failed because the harness refactor scoped a variable incorrectly; no benchmark
ran, and the corrected attempt passes. These records are preserved.

Catalogue exact capacity now links to this encoding evidence, while keeping the
historical Monte Carlo source. Exact K is 0.455% below its former point estimate
51,209,164, and the exact rate lies within the recorded Hoeffding interval
[0.1887861653, 0.1927518347]. Supported status means standalone mathematical
encoding from the experimental cache; public default discovery, application
integration and production-security qualification are separate. Historical
planning/cost reports retain their original catalogue snapshots and are not rerun.

The measured s=14 path is practical on this machine under the stated budgets.
This is not an s=15 result or proof that its preparation fits. Next assess s=15
using measured first-use/online lifetimes and profiles before setting its budget.
