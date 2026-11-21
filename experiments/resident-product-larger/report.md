# Larger structured GPU products with exact coefficient checks

Status: **78469 passed all 22 arms**, producer `5525143`, on HPC gnode13/gnode14
(two H200s, four CPU cores, 8 GiB requested host RAM). Controller runtime:
**4 min 13.462 s**. Independent review verified 738 source, 184 original evidence
and 108 cache identities, all 52 rank/product outputs and both small sanitizer
arms with zero CUDA errors/leaks.
[Submission](submission.json), [raw results](runs/larger-01/results.json),
[analysis](runs/larger-01/analysis.json), [review summary](runs/larger-01/review-checks.json),
[scheduler log](slurm-78469.log), [readiness](readiness/checks.json).

This follows [passed local Taylor78460](../resident-taylor-local/report.md): at
N16384, local-access resident median work0.222795–0.224370 seconds versus host-
stored GPU0.352875–0.361130 seconds, with unchanged resident payload. Both use GPU.

## Why change the large-case oracle

The old arbitrary-operand schoolbook oracle is quadratic and already dominated
much of the small run's CPU profile. Scaling that oracle would measure validation
rather than the distributed GPU product. The explicit new pattern keeps both
operands dense: A retains normalized pseudo-random coefficients; every valid B
coefficient is1. In characteristic two, each C[k] is an XOR window of A.

The producer's [streaming oracle](../../encoding/liftrs-shared/distributed_product_oracle_test.go)
uses constant auxiliary workspace and O(ac+ownedWords) work. The independent
[Python reviewer](run.py) builds XOR prefixes and checks two clipped endpoints for
**every saved coefficient**. CPU regressions cross-check the streaming oracle
against schoolbook for multiple fields/counts/shard origins; Python tests do the
same for prefix endpoints and reject mutually consistent bad producer files.
Arbitrary operands retain their exact schoolbook oracle and default behavior.
No FFT/arithmetic kernel is changed and no coefficient/source check is removed.

Structured inputs may produce different intermediate values/zero patterns and
costs from arbitrary root polynomials. Results must remain labeled structured;
**do not combine them with previous arbitrary-input timings or extrapolate full
preprocessing runtime**. Two arbitrary-input reference arms atN16384 allow a
same-allocation comparison. This is an experimental benchmark, not a library or
protocol behavior change.

## Exact workload and coverage

The shared product driver now uses explicit test options (backend, tile, pattern,
shape count). `TestScaledGPUProducts` admits only the declared s17-field sizes and
backends. A dimension18 grid contains262144 evaluation points, exceeding the s17
base-field cardinality131072 while remaining inside its quadratic extension.
This does **not** admit a PIR code with s18, or validate full s17 preprocessing.

[Runner](run.py)/[Slurm entry](run.sbatch) use one source/build/preflight/allocation:

- Two arbitrary-input references, N16384, host-stored GPU/local-access resident,
  both exact-domain and zero-tail shapes (4rank/product rows/arm).
- Two structured-pattern CUDA memcheck/leak arms, N4096, same two backends and
  both shapes (4rank/product rows/arm). Larger domains are not sanitized here.
- Eighteen normal arms: N16384/65536/262144, both backends,3samples/backend/size,
  order alternated by sample. Each arm has **one exact-domain geometry** and
  one size (2rank/product rows), to bound proof files. Zero-tail and geometry
  reuse remain covered by the small/reference arms.

The large geometry checks inverse workspace allocation reuse within the product;
it does not claim a second geometry ran. `inverse_reused` and
`geometry_reuse_checked` distinguish these. Every coefficient, canonical base
range, leading/constant1, high-component pointwise check, right-spectrum
preservation, all remote source/transpose routes and paired hashes remain checked.
No guard warmup runs in measured arms. Both backends retain the same product/FFT
sequence and independent public coefficient specification.

[Analyzer](analyze.py) reconciles exact source/evidence/cache identities, saved
bytes/hashes and all admitted cases, then reports three-sample median/range per
size/backend. Total/net/capture and constructor timings retain the previous
comparison's boundaries: resident input upload, result readback and proof/logging
are included; the diagnostic spectrum capture is separately timed and subtracted
per rank for net work. Inverse allocation queries are included where executed;
final preservation/oracle/fsync work is outside the product clock, inside complete
arm/controller time. Profiling/logging remain enabled; these are not GPU kernel
or unprofiled latency measurements.

## Preventing log and resource failures

[Focused tests](test_review.py) verify single-shape/subset route coverage, inverse
reuse flags, independent output rejection and actual arm selection. A conservative
s17 linearized-polynomial plan with all9 possible sparse terms gives20604 Taylor
proof events/rank/product atN262144. Budgeting800bytes/event, transpose proofs and
1 MiB overhead stays below24 MiB/file, with32 MiB/file guard. **Separate sizes and
one geometry per MPI arm** avoid exceeding that guard. No proof is suppressed
or sampled. Larger future domains need explicit bounded proof storage planning;
blindly enlarging this workload would risk the file cap before GPU memory fills.

Same admitted **2 HPC nodes /2 H200s /4 CPU cores /8 GiB requested RAM**,
one GPU,2cores,4 GiB/node; **128 MiB owned GPU/card /1 GiB RSS guard/rank**,
no pinned buffers, **128 MiB retained cache /2 GiB scratch /32 MiB file**,
**720-second controller /15-minute scheduler**. gnode03 excluded; other eligible
nodes allowed. Shared executable/wrapper/library/cwd and remote TMPDIR preflight,
explicit UCX cache control and guards reuse the validated controller. No resource
admission or default change.

AtN262144, three resident arrays total **3 MiB/card**, plus32 KiB transport,
12 KiB Taylor window and field/plan metadata; host Taylor payload24 KiB/rank.
The host reference's six-shard model is6 MiB/rank before runtime/diagnostics.
Streaming producer oracle adds no global array; the Python prefix reviewer uses
a small temporary prefix on the controller. Diagnostic operands/results are
additional and included in RSS. Prior measured peak418.320 MiB/RSS and1.791 MiB
owned GPU leave substantial room below unchanged guards. Projected retained
output/reference cache about32–50 MiB fits128 MiB; largest result file0.5 MiB.
No intermediate disk matrix or cache deletion is introduced.

78460 completed14 smaller arms/builds/profiles in200.380 seconds. Extending N by
16x from16384 to262144 suggests roughly16–30x arithmetic/transfer work per large
product, with fewer arms per size/one geometry and a linear oracle; this is a
conservative budgeting projection, not a fullsize prediction.22 arms including
build/profiles are projected250–500 seconds inside720 seconds, each arm retaining
240 seconds (220-second test cutoff). Completed phases/profiles save before the
next; partial results survive a later failure. Larger normal cases deliberately
reuse already validated kernels, with no new sanitizer workload size.

```sh
python3 -B experiments/resident-product-larger/test_review.py
python3 -B experiments/resident-product-larger/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/resident-product-larger/slurm-%j.log \
  experiments/resident-product-larger/run.sbatch larger-01
# After the user's completion ping:
python3 -B experiments/resident-product-larger/analyze.py larger-01
```

Raw revision/commands/resources, flushed stage/source/proof logs, CPU profiles
and completed-arm results persist in `runs/<name>/`. Sealed reusable binaries and
output/reference cache remain outside Git at
`.cache/liftrs/resident-product-larger/<name>/`. Earlier caches/profiles remain.
No agent monitoring after submission. If this gate passes, use the measured
large-component costs to move toward distributed root/Newton executor integration
and s18 field admission, rather than repeating small tile comparisons. Whole
pipeline discovery/generator/persistence/scaling remain distinct; stillstep1.

## Completed measurements (SCALING-159)

Three samples per backend/size; structured dense inputs, s17 field. Net product
work includes input upload, result readback, communication and proof logging,
but subtracts the separately recorded diagnostic right-spectrum capture.
Setup, CPU oracle, output checks and persistence are outside this product clock
and inside the controller runtime. These are profiled component measurements.

| Evaluation values | Host-stored GPU median (range) | Local-access resident GPU median (range) | Resident speedup |
|---|---|---|---|
| 16,384 | 0.349606 s (0.346083–0.350733 s) | 0.225291 s (0.224596–0.226843 s) | 1.55× |
| 65,536 | 1.245009 s (1.239719–1.247973 s) | 0.591319 s (0.590708–0.592592 s) | 2.11× |
| 262,144 | 4.911230 s (4.905237–4.914341 s) | 1.553557 s (1.552168–1.560198 s) | 3.16× |

Maximum logged normal host RSS: **383.594 MiB/process**; small sanitized RSS:
**414.480 MiB/process**. Maximum owned GPU allocation: **4.612 MiB/card**, excluding
context/driver memory. Retained cache: **32.303 MiB**. Largest proof stdout file:
**12.924 MiB**, below the unchanged 32 MiB/file limit; no proofs dropped.
The sample-0 largest resident Go profile spends 80% of sampled CPU time in opaque
cgo calls; this cannot distinguish kernel execution, GPU waits or MPI/native work.
The linear oracle no longer dominates that profile.

All three sizes passed the independent prefix oracle coefficient by coefficient.
Arbitrary-input references, inverse workspace reuse and small sanitizer checks
also passed. Larger inputs were not sanitized. These results support proceeding
to [local GPU roots plus a distributed final merge](../distributed-root-merge/report.md),
not a fulls18 runtime projection or production backend/default change. Partitioned
support, Newton integration, discovery and coordinated persistence remain.
