# Progressive cyclic encoder scaling

## Purpose and reproduction

Benchmark successive catalogue fields while predicted **first-use time**—code
discovery, reusable preprocessing and one database encoding—remains below one hour
and predicted process RAM stays within 12 GiB on the 48 GiB host. This policy was
confirmed by the user; preprocessing is still reported separately because it can
be reused across databases. [policy.json](policy.json) sets the hour, memory budget,
1.5 forecast safety factor and three encoding repetitions.

From the repository root:

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/cyclic-scaling/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

Requires compatible Go, Python 3 with POSIX wait4, git and temporary output space.
[The runner](run.py) builds [the driver](main.go), writes one event log per measured
field, and updates [results.json](results.json) and [table.md](table.md) after each
case. Raw events retain all timings, allocations, GC observations, exact capacities,
checks and output hashes. The table distinguishes measured cases from cases not run.
The driver does not implement a different encoding algorithm.

## Implementation and measurement boundaries

The library's in-memory orbit bound is extended to m=2,s<=13 for resource-limited
experiments. Admitting a specification is distinct from having a completed benchmark
or supporting its serialized cache. Existing basis/complete JSON loader limits
remain 2 MiB/64 MiB; this experiment keeps reusable prepared constants **in memory**
and does not serialize them. Artifact redesign and large JSON input limits are
therefore not prerequisites to measuring encoding. No GPU, streaming polynomial
intermediates, application-plan execution or server computation is added.

One child process per configuration runs:

1. Code-only exact monomial enumeration and support discovery, using public seed
   1000+s. Record exact K, N and discovery sample count.
2. Reusable `PrepareCyclicFFT`, including generator construction and spectrum setup.
3. Three database encodings of a uniform message (seed 71), reporting every call
   and the median. Before each call, drop the preceding word and collect garbage;
   GC is outside the encode timer. The prepared encoder and message are retained.
4. Verify all systematic symbols and 16 actual degree-at-most-t curves using
   independent carryless polynomial arithmetic. Obtain omega from the first public
   systematic position, iterate coordinates in linear time, and cross-check the
   public exponentiation mapper at the first/middle/last positions.
5. Stream canonical uint32 hashing and tightly packed output through bounded
   buffers. Pack element i, bit b at file bit i*s+b; write/sync one temporary file
   and check a cached readback. Do not allocate another full-size output copy.

Packing/hashing, verification and output I/O are outside encoding time. Discovery
and preparation contain no application database or keys. They run once per code,
then the same prepared constants are reused. JSON loading and artifact size are
**not measured** here. Earlier s<=10 reports retain those separate observations.

The parent records whole-case peak RSS using wait4. It does not isolate each
phase's process high-water mark. Each phase also records baseline heap, sampled
peak heap/runtime memory, cumulative allocated bytes and GC cycles. A 100 ms
runtime-memory sampler/guard runs inside phase timing; this small measurement cost
is included. Samples can miss short peaks. Allocation totals, sampled heap values
and process RSS have different meanings and must not be interchanged.

The Go soft memory limit is set to 9 GiB (75% of the 12 GiB budget); sampled Go
runtime Sys exceeding 12 GiB aborts the case. This is not an OS address-space cap:
Go reserves a large virtual arena. Physical RSS is reported after exit. The parent
also kills a case after one hour of wall time, covering repetitions, verification
and output I/O as well as first-use work. A timeout or guard failure is recorded,
not treated as a successful measurement or proof that the code cannot run elsewhere.

## Forecast and admission model

Begin with the committed s=10 cyclic measurement, then use the **largest completed
case** as the next anchor. Let its field degree, length and dimension be a,N_a,K_a,
and the proposed case use s,N,K. Define workload ratios:

- Transform work: `r_F = (N/N_a)*(s/a)^3`.
- Tree work: `r_T = (K/K_a)*(log2(K)/log2(K_a))^3*(s/a)`.

The transform proxy combines N log²N arithmetic with a linear degree-cost proxy
for scalar field multiplication. The tree proxy includes product-tree levels.
Predict discovery/encoding using r_F, and preprocessing using max(r_F,r_T).
Predict process RSS using max(N/N_a,K/K_a). Multiply each forecast by 1.5 for
admission. Candidate K comes from the current catalogue until exact enumeration
replaces it. Results record each anchor, policy and forecast used in the decision.

These are **heuristics**, not time/RAM bounds or statistical confidence intervals.
Discovery sample count, polynomial sparsity, FFT cutoffs, cache effects and field
polynomials can change costs. The initial s=11 discovery forecast was low because
it required two samples rather than s=10's one; the completed-case anchor then
informs later forecasts. Resource guards address unexpectedly expensive admitted
cases, but the model can also conservatively exclude feasible cases.

The runtime criterion uses discovery + preparation + one encoding, not just the
encoding column. The hard operational timeout covers the whole benchmark case.
Later fields are listed as estimates when the resource policy rejects them.
The packed uint32 adapter is bounded at s=13; much larger estimates additionally
require changed coordinate/cardinality representations and memory/I/O backends.
No extrapolated number establishes that those unimplemented backends are feasible.

## Correctness, environment and limitations

[Driver tests](main_test.go) independently check the generalized curve arithmetic
against field multiplication for every s=2..20, and streaming packing against the
bit definition for s=2..13, including empty and partial-byte output. Existing tiny
dense-oracle, historical output and s=10 curve checks remain applicable. Larger
runs validate exact counted dimension, every systematic symbol and sampled actual
curves; they do not compare with a dense matrix or a large paper-encoder run.
Sampled curve sums are necessary checks, not exhaustive code-membership or security
proofs. Deterministic seeds are public experiment data, not protocol randomness.

Host: Apple M5 Pro, 18 cores, 48 GB RAM, macOS 26.6.2 ARM64; exact hardware
metadata is retained in the [seed checkpoint](../encoder-scaling/results.json),
and this run records Go/Python/platform/CPU count. Arithmetic is single threaded;
GC/runtime workers may use other cores. No affinity, cache flush, randomized
order or confidence intervals. The memory guard and per-repeat GC differ from the
earlier s=10 timing harness, so cross-report ratios are descriptive, not controlled
optimization comparisons. Tiny packed-file sync/readback is not sustained storage
throughput or a grouped application layout benchmark.

## Results and stopping point

Completed s=11 and s=12. [results.json](results.json), [table.md](table.md),
[s11-events.jsonl](s11-events.jsonl) and [s12-events.jsonl](s12-events.jsonl)
retain all measurements and checks. Both cases passed all systematic symbols,
coordinate-map samples and 16 independent curve checks; no timeout or memory
abort occurred. Source revision is c06375c plus uncommitted ENCODE-007 sources
whose exact hashes are recorded. The catalogue input is the c06375c catalogue;
subsequent promotions preserve the original Monte Carlo evidence and do not alter
these recorded forecasts or exact dimensions.

| s / t | Exact K | Discovery | Reusable preprocessing | Median encoding | First use | Whole-case peak RSS |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 11 / 6 | 455434 | 15.07 s | 145.26 s | 14.60 s | **2.92 min** | **0.26 GiB** |
| 12 / 6 | 1965739 | 34.98 s | 677.19 s | 67.76 s | **13.00 min** | **1.19 GiB** |

Actual benchmark case wall times, including three encodings, verification and I/O,
were 204.82 s and 914.88 s. The policy's first-use quantity includes only one
encoding, not all three repetitions. s=11 needed two public discovery samples;
s=12 needed one, illustrating why the stage cannot be extrapolated from N alone.

Earlier completed cyclic measurements cover the remaining smaller catalogue fields:

| s / t | Preprocessing | Encoding | Source |
| --- | ---: | ---: | --- |
| 8 / 8 | 0.443 s | 0.071 s | [hybrid-product comparison](../fast-polynomial/table.md) |
| 9 / 7 | 3.410 s | 0.430 s | [hybrid-product comparison](../fast-polynomial/table.md) |
| 10 / 6 | 26.475 s | 2.682 s | [s=10 checkpoint](../encoder-scaling/table.md) |

These historical measurements were not rerun; their per-process/GC boundaries
differ as described above. Together with the new runs, all catalogue degrees
s=8..12 have measured cyclic encoding evidence. The s=11/12 catalogue entries now
use exact K and verified in-memory cyclic support, preserving their older Monte
Carlo points K=453002 and K=1964713. No larger paper comparison or persistent
large-cache loading is claimed.

The final s=12 anchor rejects s=13: **171.7 min first-use time** and **10.86 GiB**
RAM with the safety factor. Time alone exceeds the agreed hour, despite RAM fitting.
s=14 projects **1011.0 min** and **46.53 GiB**, exceeding both budgets. All larger
fields likewise fail; their forecasts and reasons are in the full table/results.
The packed adapter's s<=13 bound is an additional limitation, not the reason an
otherwise admitted case was left unmeasured. Even the transform-only workload
proxy at s=13 (ignoring the larger tree term and safety margin) gives about
66 minutes first use. None of these forecasts establishes impossibility: they
are conservative admission decisions under this implementation and host policy.
Farther extrapolations, particularly where coordinate representations must change,
are growth-model outputs rather than predictions for an implemented larger backend.

Peak Go heap sampled during preprocessing was about 0.93 GiB at s=12; process RSS
peaked at 1.19 GiB across the whole run. Each encoding allocated about 167.5 MiB
cumulatively and sampled about 279.7 MiB live heap including cached code/input and
output. These are distinct from peak RSS and are not isolated workspace bounds.
Packed outputs were 5767168 bytes (5.5 MiB) and 25165824 bytes (24 MiB), streamed
with sync and cached-readback verification. Raw events retain timing/hashes.

## Validation and next work

After the measurements, run:

```sh
go test ./encoding/liftrs ./pirconfig ./experiments/cyclic-scaling -count=1
go test -race ./encoding/liftrs ./pirconfig ./experiments/cyclic-scaling
go vet ./encoding/liftrs ./pirconfig ./experiments/cyclic-scaling
go test ./...  # isolated copy: legacy tests write CSV files
```

All tests, race checks and vet passed. Source/seed-input/event hashes verify;
policy consistency and all forecasts reproduce from the recorded anchors and
catalogue revision. Documentation file links and `git diff --check` also pass.
The [completion message](../../message-board/2026-10-28-1541-codex-encode-007-completed.md)
records final outcomes.
No tests ran concurrently with the benchmark cases.

This completes the requested resource-gated baseline without a format redesign.
Next use the measured scaling boundary and the [s=10 CPU profile](../encoder-profile-s10/report.md)
to prioritize arithmetic/transform improvements, then repeat this unchanged workload
before admitting larger first-use cases. Preprocessing remains reusable, so keep
its improvement distinct from per-database encoding gains. External-memory/GPU
work remains behind the existing scope decision gate.
