# Portable field multiplication and reusable FFT tables

## Purpose and reproduction

Reduce portable CPU encoding time without changing the field basis, information
set, artifact format or codewords. The baseline is revision `a2e0352`; optimized
runs start from that revision with the source hashes recorded in
[scaling-results.json](scaling-results.json). Historical snapshots remain intact.

From the repository root (compatible Go, Python 3, POSIX `wait4` and git):

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/field-multiplication/run-baseline.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/field-multiplication/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --stage kernels
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/field-multiplication/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --stage scaling
python3 experiments/field-multiplication/summarize.py
```

The original kernel run used the same [kernel driver](main.go), built with Go
1.27.3 and executed directly; see [its provenance](kernels-before-provenance.json)
and [raw measurements](kernels-before.json). The reproduction runner writes
`kernels-after.json` plus provenance; the general kernel is unchanged, so this is
not a separate before/after arithmetic implementation. Use
[run-baseline.py](run-baseline.py) for the exact pre-change s=10 control and
[run.py](run.py) for full optimized encodings. [summarize.py](summarize.py) checks
hash compatibility and regenerates the comparison and s=13 forecast.

Environment: Apple M5 Pro, 18 CPU cores, 48 GB RAM, macOS 26.6.2 arm64, Go 1.27.3,
Python 3.9.6, default GOMAXPROCS, portable Go without native instructions or GPU.
This is the same host as the [preceding scaling run](../cyclic-scaling/report.md)
and [recorded hardware metadata](../encoder-profile-s10/results.json). Packed
output uses temporary local APFS storage, sync and cached readback; it is not a
sustained storage throughput benchmark. No tests or other benchmarks ran alongside
the retained full-encoding measurements. The
[initial control](initial-baseline/README.md) overlapped initial correctness tests
and is excluded; a fresh s=10 control replaced it.

## Algorithm selection

The [kernel comparison](main.go) covers every s=2..20 using 4,096 seeded operand
pairs per field, three adaptive 80 ms benchmark samples per method, and separate
reusable reduction-table construction timing. It compares:

- The preceding variable-length shift/reduce algorithm and the unchanged public
  `BinaryField.Mul` method.
- A fixed-length masked loop replacing source-level branches with arithmetic.
- Four-bit Horner multiplication, with a 64-byte field reduction table. Its
  operand-dependent 16-entry table construction is included in every operation.

For the relevant s=9..12, neither alternative provides a convincing improvement
over the production method. Keep general multiplication unchanged. The four-bit
method improves some larger-degree kernel cases, but is not selected on that
basis: these measurements do not establish whole-encoder gains there. Wrapper/
inlining effects and the short repeated operand set limit kernel comparisons.
Source-level masking alone is not a constant-time security claim.

Instead, [constantProduct](../../encoding/liftrs/orbit_constant.go) prepares
four-bit rows for the GF(2)-linear map `x -> c*x`, preserving packed polynomial/
tower coordinates. [Transforms](../../encoding/liftrs/orbit_fft.go) reuse these
maps for sparse-polynomial coefficients, butterfly factors and their inverses.
Changing recursion offsets and general products still use scalar arithmetic.
Zero/identity maps allocate nothing. No full-field multiplication or log tables
are added, and the general field API remains table-free.

Each transform has a 256 KiB budget for row entries; base and extension transforms
therefore have a combined 512 KiB ceiling. Descriptors and allocator overhead are
additional. Exhausted budgets fall back to scalar multiplication. Tables are
immutable and database-independent. Discovery or artifact loading builds them;
per-database encoding does not. They are process-local runtime metadata, not new
serialized artifact fields. This is a CPU memory/cache tradeoff, not a change to
the encoding algorithm's mathematical output. Input-dependent lookups do not
provide constant-time arithmetic.

Measured row-entry storage, recorded by [the budget tests](../../encoding/liftrs/orbit_constant_artifact_test.go)
in [verification.log](verification.log), is 52.4 KiB combined at s=9, 64.1 KiB at
s=10 and 89.6 KiB at s=11/12. The s=13 table-construction test uses 147.3 KiB;
this is not an s=13 encoding benchmark. General tower-delta multiplication remains
unchanged; profile remaining arithmetic before extending the cache strategy.

## Full-encoding methodology

The runner reuses the unchanged [cyclic-scaling driver](../cyclic-scaling/main.go):
exact monomial enumeration/support discovery, reusable cyclic-FFT preparation,
three seeded database encodings, all systematic symbols, 16 independent actual
curve sums, canonical/packed hashes, streamed sync and cached readback. Seeds are
`1000+s` for discovery and 71 for uniform messages. A case encodes one full
codeword; it does not execute an application partition or grouped storage adapter.

GC before each phase is outside the phase timer. The 100 ms heap/Go-runtime
sampler and memory guard remain inside it; whole-process peak RSS is separate.
Encoding is the median of three calls. First use is discovery + preparation +
one median encoding. The 12 GiB guard/9 GiB soft GC limit and one-hour whole-case
timeout are unchanged. Heap samples are observations, not exact workspace bounds.
Large caches remain in memory: JSON serialization, loading and isolated phase RSS
are not measured by these scaling runs. Loader table construction is covered by
correctness tests, not a new loading latency claim.

The fresh [s=10 control events](baseline-s10-events.jsonl) and
[process metadata](baseline-s10-process.json) use the same driver from `a2e0352`.
The [s=11/12 baseline](../cyclic-scaling/results.json) uses the same driver on the
same host, measured earlier. The s=9 historical measurement used separate
preparation/loading processes, so it is context rather than a controlled speedup
comparison. Raw optimized events and their hashes are indexed by
[scaling-results.json](scaling-results.json); the runner records exact source,
revision, environment, timings and output verification data. The
[evidence manifest](evidence-manifest.json) also records hashes of retained
scripts, source driver, raw controls, results and verification logs.

## Results

The [generated comparison](table.md) reports:

| s / t | Preprocessing before → after | Encoding before → after | Encoding speedup | New first use | Peak RSS |
|---|---:|---:|---:|---:|---:|
| 9 / 7 | 2.35 s (new) | 0.249 s (new) | — | 2.76 s | 0.021 GiB |
| 10 / 6 | 27.45 → 12.70 s | 2.77 → 1.22 s | 2.28× | 14.71 s | 0.071 GiB |
| 11 / 6 | 145.26 → 64.36 s | 14.60 → 6.15 s | 2.38× | 77.82 s | 0.251 GiB |
| 12 / 6 | 677.19 → 307.59 s | 67.76 → 29.06 s | 2.33× | 354.27 s | 0.923 GiB |

Exact dimensions remain K=17200, 103556, 455434 and 1965739, respectively.
New discovery times are 0.162, 0.788, 7.315 and 17.623 s. First use includes this
work; preprocessing remains reusable across databases. The s=9 historical
cyclic measurement was 0.430 s encoding and 3.410 s preparation, but different
process/loading boundaries prevent presenting that as a controlled speedup.
All s=9..12 complete canonical and packed hashes match prior results. The s=12
RSS observation decreased from 1.19 to 0.923 GiB; allocator/GC differences mean
this is not a guaranteed memory reduction from adding the small tables.

The [updated s=13 forecast](s13-forecast.json) uses the unchanged
[model](../cyclic-scaling/run.py), s=12 measurements, catalogue-estimated
K=11955914 at t=5, and the 1.5 safety factor. First use is **78.0 minutes** with
**8.42 GiB** projected RAM. Time still exceeds one hour, so s=13 is not admitted
or benchmarked. This is a heuristic forecast, not proof that the field cannot
fit the policy. Optimizing this s=13 case would require further measurement or a
separate change to the policy; no larger support is claimed here.

## Validation

All retained validation passed; [verify.py](verify.py),
[verification.log](verification.log) and [verification.json](verification.json)
record the commands, outcomes and durations:

- Targeted liftrs/catalogue tests, including exhaustive small-field arithmetic,
  independent larger-field arithmetic, dense-oracle encoder comparisons and
  historical codeword fixtures.
- [Constant-map tests](../../encoding/liftrs/orbit_constant_test.go) against
  independent carryless convolution/long division: base fields s=2..20, packed
  towers through s=13, partial nibbles, zero/identity and budget fallback.
- [Artifact rebuild tests](../../encoding/liftrs/orbit_constant_artifact_test.go):
  byte-identical paper/cyclic-FFT artifacts at s=4,8,9; table-size checks through
  s=13. Existing loader/output tests include s=10.
- All systematic symbols and 16 independent actual curves for every measured
  s=9..12 case; complete canonical/packed hashes equal their prior fixtures.
- Race tests and vet for liftrs, pirconfig and the kernel driver.
- Full `go test ./...` in an isolated copy, preventing legacy tests from changing
  working-tree CSV files.
- Measured source/event hashes, forecast reproduction and documentation links;
  `git diff --check` passes.

The first general-kernel comparison and initial correctness checks informed
selection; final tests ran after timing completed. No native acceleration,
constant-time guarantee, larger paper-encoder performance, large-cache loading
latency or integrated application encoding is established by these results.
