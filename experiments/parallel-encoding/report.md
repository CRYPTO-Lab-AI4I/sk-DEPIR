# Parallel matrix encoding integration

Completed local s=13 integration (SCALE-004). This implements expanded scaling-plan steps 1–3:
general-basis/affine native execution, parallel matrix/Taylor transforms, blocked
in-memory transposes and buffered minimal-plan loading. Preprocessing remains Go.
The default encoding APIs and standard command backend are unchanged.

## Reproduction and measurement boundaries

Use [run.py](run.py), [main.go](main.go) and [analyze.py](analyze.py). The runner records revision, exact
source hashes, compiler/platform, binary and cache hashes, wall time, process peak
RSS, and per-stage CPU profiles. Result directories are fresh and never overwritten.

```sh
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  python3 experiments/parallel-encoding/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name smoke \
  --smoke --workers 1,4 --repeats 1
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  python3 experiments/parallel-encoding/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name s13-prepared \
  --workers 8,16 --repeats 3
python3 experiments/parallel-encoding/analyze.py s13-prepared
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  python3 experiments/parallel-encoding/check.py --go /tmp/sk-depir-go-toolchain/go/bin/go
python3 experiments/parallel-encoding/native-check.py --cxx /usr/bin/clang++
```

Sources: revision 6a9f26f plus hash-identified uncommitted integration files.
Environment: Apple M5 Pro, 18 logical CPUs, 48 GiB, macOS arm64; generic C++17 -O3,
Go version and OS/compiler strings in each manifest. No affinity/frequency control.
The case limit is 30 minutes and Go soft memory target 9 GiB; these are not OS RAM
caps. No larger field workload is admitted by this experiment.

The s=13 input is the retained ~302 MiB minimal binary from local calibration,
SHA-256 f9992fa0a354d1154e73ccf110899abd96fb3f08b532fc2ae8a8aa1cfb0b7c47,
derived from the separately verified original cyclic cache. Its expected digest
is checked while loading. No discovery or large preparation is repeated. Small
s=8 fixtures generate and retain a complete code artifact if missing.

Loading, native plan creation, workspace allocation, database encoding and
verification are separate profiled phases. Profiles include native calls under
cgo; explicit native stage intervals distinguish Taylor, columns plus transposes,
and affine rows. These are elapsed times, not sampled CPU fractions. Pointwise
arithmetic and canonical output assembly remain Go. Output allocation is included
in encoding. Hash verification is outside encoding, and requires full canonical
word equality against retained s=13 evidence (seed 71) or the s=8 reference.
Packing, masking, cryptographic permutation and physical output writes are absent.

The matrix path owns an additional N-element uint32 transpose buffer, reused
across the five transforms. A 64×64 blocked transpose improves locality; it is
in-memory and does not constitute a disk-backed encoder. Taylor conversion keeps
within-block descending/ascending dependencies and parallelizes independent
siblings. Higher Taylor levels can remain serial bottlenecks. Native worker
launch/join is inside each call. Complete buffers and immutable plans are borrowed
across one native call; no per-element language crossing or interface data copies.
Small Taylor multiplication tables are prepared once with the native plans and
reused. Their construction is included in setup, outside database encoding.

Buffers are cleared/reused; one output at a time is retained, with GC outside each
phase. Phase endpoint Go heaps and cumulative allocation are not peak RSS. Process
RSS includes loading, runtime, all native threads, workspace and verification.
Comparisons to the historical 153.409 s native median are indicative: these are
new profiled measurements, with no fresh matched baseline or environment isolation.

## Evidence

- [Smoke manifest](smoke/results.json), [phases](smoke/phases.json).
- [Final s=13 manifest](s13-prepared/results.json), [phases](s13-prepared/phases.json),
  [table](s13-prepared/table.md), [native sample](s13-prepared/native-sample.txt).
- [Initial integration manifest](s13/results.json), [table](s13/table.md): same
  mathematical algorithm, but rebuilt small Taylor tables inside transforms.
  Preserve this intermediate evidence; final comparison uses reusable tables.
- [Validation script](check.py), [checks](validation/checks.log),
  [verification](validation/verification.json), [final checks](validation/final-checks.json); [native sanitizer script](native-check.py),
  [native checks](native-validation/checks.log), [native verification](native-validation/verification.json).
- Retained binaries, small preprocessing and provenance are under the ignored
  `.cache/liftrs/parallel-encoding/<source-id>/` path recorded in each manifest.
  The original preprocessing caches are read-only and preserved.

## Results and interpretation

Final profiled medians of three complete encodings, with code-only Taylor tables
prepared once:

| Workers | Encoding | Taylor | Columns + transposes | Affine rows | Historical baseline / encoding |
|---:|---:|---:|---:|---:|---:|
| 8 | 29.795 s | 4.974 s | 9.699 s | 12.477 s | 5.15× |
| 16 | 20.152 s | 4.380 s | 5.833 s | 7.308 s | 7.61× |

Stage columns are independent medians and need not sum to the encoding median.
All six complete canonical hashes equal
`6120f0388f6035dd480e391b26ef93c04ed368a77f2f93c02f6d4f56b99a7b79`.
The small fixtures additionally agree with direct Horner, independent dense-code
encoding, existing transforms and inverse recovery. Input and information-set
semantics do not change.

Process peak RSS is **1,232,437,248 bytes (1,175.34 MiB / 1.148 GiB)** across both
worker settings, including profiles, loader, message, native threads, workspace,
output and validation. Plan coefficient payload is 316,141,200 bytes; reusable
workspace including the transpose buffer is 584,576,652 bytes; input 47,705,740
bytes and output 268,435,456 bytes. This is measured process memory for this
schedule, not a larger-size bound or proof of bounded preprocessing.

Minimal loading takes **0.167 s**, with ~318.6 million allocated bytes including
profiling overhead, versus ~635 million in the earlier per-element loader probe.
The historical loader took 1.01 s; this comparison is not a matched cold-disk
benchmark. Native code-only setup is 1.93/3.39 ms at 8/16 workers; workspaces are
allocated separately. Cheap metadata can be regenerated; expensive preprocessing
and all measured binaries/profiles remain retained. Large s=13 discovery and
P/G/spectrum construction are not measured again or claimed faster here.

A five-second `/usr/bin/sample` collection (1 ms sampling, timestamp in the sample)
overlaps the first 16-worker encoding; this repeat is explicitly sampled as well
as Go-profiled. All three final 16-worker times are 20.152, 19.942 and 20.369 s.
Go profiles run on every phase and native stage timers are always included.
The initial run measured 29.401/19.450 s medians, supporting the same conclusion;
it is retained with its original setup boundary and source snapshots. Do not
interpret subsecond differences between runs as an optimization effect.

The local integration exceeds the exploratory 3–6× target at sixteen workers,
relative to the historical baseline. This gain combines algorithm decomposition,
parallelism and lean workspace orchestration; it is not a controlled measurement
of thread scaling alone. The earlier 12× independent-tile gain is not claimed for
the full encoder. Taylor still benefits much less from extra workers, and remains
partly sequential. No GPU, specialized hardware or new external dependency is used.

## Validation, remaining scope and next step

Go-only and native full repository suites pass in an isolated checkout; tagged
vet, focused Go race checks, independent native arithmetic and original ABI
address/undefined-sanitizer checks pass. A separate sanitizer fixture checks the
new general-basis matrix/Taylor path against Horner and inversion at every split
of small domains with one/four workers. Go tests check affine offsets, repeated
workspace clearing, immutable-plan concurrent use, lifecycle/alias errors, and
malformed/truncated/checksum-invalid minimal files. A final prefix-capacity
guard rejects plans whose ordinary convolution exceeds one field domain; focused
race, loader, example and vet checks were repeated after this guard. It does not
change the measured s=13 codeword path. Measured source versions remain recoverable through content-addressed snapshots
after comment clarifications and the final guard. The final binding also adds
explicit `-pthread` compile/link flags; focused tests pass with those flags on
this host. This is not Linux platform qualification or a new timing baseline. Go race instrumentation does
not cover native memory; address/UB sanitizers do not establish absence of races.
The original ABI fixture is retained at `.cache/liftrs/language-s13/transform.bin`;
regenerate it using the [kernel-language experiment](../kernel-languages/report.md).

Expanded integration steps 1–3 are complete for the in-memory s≤13 path. No stable
artifact publication format, integrated application encoder, external-memory
implementation or s≥14 support is added. The next gate is exact discovery and
reusable preparation/resource measurements at s=14, followed by full encoding
and budget review before s=15. Use measured buffer/lifetime accounting for those
runs rather than multiplying the former JSON-process RSS. Larger commodity hosts
or disk-backed execution remain options where capacity demands them.

