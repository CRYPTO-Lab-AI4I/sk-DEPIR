# Cross-node GPU root products and polynomial merge

Status: **two-node validation 77285 passed**, reviewed as SCALING-118.
HPC gnode16/gnode17, one H200 and2 CPU/4 GiB requested host RAM per node.
Scheduler elapsed **2 minutes 36 seconds**; controller elapsed **155.737 seconds**.
Exact implementation revision `f80a6858f3e3c2155627ee983e867ba875ce6fad`.
[Submission receipt](submission.json), [raw results](runs/roots-01/results.json),
[verified analysis](runs/roots-01/analysis.json), [accounting](runs/roots-01/accounting.txt),
[progress/Slurm log](slurm-77285.log). This follows the passed
[staged GPU exchange](../gpu-staged-transport/report.md) and
[independent root workers](../root-subset-validation/report.md).
No full s=18 preprocessing run or latency is claimed.

## Implementation and inputs

[Runner](run.py), [scheduler script](run.sbatch), [rank/profile wrapper](rank.sh),
[completion reviewer](analyze.py), [review regressions](test_review.py),
[optional MPI binding](mpi/world.go),
[GPU root test](../../encoding/liftrs-shared/root_distributed_gpu_test.go),
[shared fixture/root helpers](../../encoding/liftrs-shared/root_fixtures_gpu_test.go).
The existing same-node worker test now uses those same helpers. All arithmetic
remains in the maintained Go/C++/CUDA root/product implementation; no second root
or merge algorithm is introduced. Ordinary CPU/client builds do not select
`liftrs_mpi` and acquire no MPI or CUDA dependency. This is an experimental server
validation path, not a public distributed-preprocessing API.

Each rank independently builds the same bounded immutable canonical-field
fixture: genuine s=8 support (t=5, seed109), then synthetic Frobenius-closed s=17
subsets of 65,536 and 262,144 exponents, including indices above2^32. Synthetic
seed is109+s+count, as in the earlier worker comparison. Synthetic subsets are
not full s=17/s=18 PIR support. Sorted membership is explicitly selected for these
sparse cases; ordinary full-support/cache/discovery bitmap defaults stay unchanged.

Every pass partitions complete Frobenius orbits deterministically between two
workers using the shared partition helper. Each rank builds its own normalized
partial root with its one-GPU executor. `MPI_Sendrecv` exchanges counts, validates
bounded lengths, then transfers the **entire uint32 coefficient polynomial**.
No array is written to disk for transport. MPI sees synchronous host pointers;
GPU root download and merge upload stay inside the existing executor. Unlike the
preceding transport probe, these host polynomials are ordinary Go-owned arrays,
not a new pinned staging allocation. No GPU-direct, asynchronous overlap or
preserved device-resident polynomial representation is claimed.

Each rank merges its own and received polynomials with the same GPU product
helper. Both ranks duplicate the final merge and CPU oracle deliberately to check
ownership and complete coefficient equivalence. This bounded gate still
replicates support/partition data and the reference on both nodes. It does not
implement scalable support discovery, arbitrary node counts, a multi-level merge
tree, distributed transforms, generator/Newton stages or streamed publication.
Those remain required before any full distributed s=18 admission.

MPI is initialized with `MPI_THREAD_FUNNELED`; all MPI calls stay on the same
locked Go OS thread. MPI failures, malformed lengths or a local test failure abort
the communicator so the other rank cannot be left waiting indefinitely. Slurm
rank must equal MPI rank; the reviewer requires two distinct hostnames and one
visible GPU per rank. Coefficients travel as `MPI_UINT32_T`; retained arrays use
canonical little-endian uint32 serialization. Both coefficients and peer degree/
normalization are checked before the GPU merge.

## Checks, timing and retained evidence

For every case, a complete native CPU root is computed using sorted membership,
and four independent extension-field Horner witnesses check selected roots. Every
warm-up/measured merged GPU coefficient is compared to the complete CPU result.
One warm-up and two timed passes run per case/rank. The final pass retains local,
peer, merged and CPU-reference polynomial files per rank. The reviewer checks
complete cross-rank sent/received bytes, merged/reference bytes, normalization,
canonical field ranges, counts, timing/resource bounds and source/evidence/cache
seals. A small s=8 case repeats under CUDA memcheck on both ranks, also retaining
its arrays; both sanitizer processes must report zero errors and leaks.

All four CPU reference hashes from job77213 are fixed regression expectations for
the shared fixture extraction (including the existing s=12 case). The readiness
check passed all four in16.93seconds; this is a correctness-check duration, not a
new performance baseline. [Saved log](readiness/fixture-regression.log), [exact readiness checks/source hashes](readiness/checks.json). The tagged
MPI/CUDA test binary compiles on HPC without running GPU/MPI work. Four Python
review tests and the existing five staged-runner tests pass. OpenMPI compiler
wrappers emit a sandbox socket warning on the login-node compile, but the build
completes. The scheduled build runs outside that local sandbox restriction.
[Actual immutable snapshot check](readiness/snapshot.json):621 build files,
2,941,820bytes; all10 locally checked implementation-input hashes match the
committed source `f80a685`. No GPU/MPI execution or second rebuild in this check.

Pipeline timing begins after a rank barrier and includes repeated partitioning,
local root construction, full polynomial exchange, receiving-GPU merge and full
CPU coefficient comparison. Setup/fixture creation and CPU reference generation/
Horner checks are separate. Reductions occur after stopping the local clock;
reported times are slowest-rank maxima, duplicated across rank rows and not
independent samples. Phase maxima must not be summed. Final array persistence,
CPU profile export, compilation and fixture regression are outside pipeline clocks
but inside complete job time. There is no matched whole-root latency baseline in
this new run, so do not infer a speedup over previous jobs or project full-size
runtime from these samples.

Per-rank Go CPU profiles cover normal and sanitized runs with distinct filenames;
matching executable/library and reusable public arrays remain in the sealed cache.
Go profiles include host/root/oracle/control work but do not attribute GPU
instructions; phase clocks and native progress records provide additional context.
Profiling/logging overhead is included; there is no unprofiled latency baseline.
Timestamped JSON logs identify rank, host, case, sample, phase and completed/total
work; native CPU/GPU progress is enabled via `LIFTRS_CPU_PROGRESS=1`. Rank RSS
high-water marks are recorded at phase/progress boundaries. There is no invented
command-group RSS or total-job-memory measurement. Stalled MPI/GPU work has
unknown remaining work and bounded deadlines, not a fabricated ETA.

## Same-budget resource admission and reproduction

This uses the already approved two-node GPU envelope; no increase is requested:

| Resource | Limit/allocation |
|---|---|
| HPC nodes / GPUs | 2 nodes,1 H200 per node; GPUq, excluding gnode03 |
| CPU cores / requested host RAM | 2 cores and4 GiB per node;4 cores/8 GiB total |
| Owned GPU cap | 128 MiB per card,256 MiB aggregate; one executor per rank |
| Rank RSS guard | 2 GiB/rank, checked at progress/phase boundaries |
| Pinned staging cap | Existing64 MiB/rank ceiling; this path adds no pinned buffer |
| New retained cache / scratch | 128 MiB /2 GiB on controller node;32 MiB per-file cap |
| Deadlines | 220s test /240s MPI step /260s command /720s controller /900s Slurm |

The largest retained merged polynomial is1,048,580bytes (262,145 uint32 values);
the largest local/remote polynomial is about512 KiB. Principal GPU arrays are
bounded by maximum524,288 elements, with allocations checked against the128MiB
executor cap; measured rank/device peaks are reported below (not a full-size projection). Duplicated support,
CPU references and final merges are test overhead, not a full-size memory-saving
design. The existing faster resident default remains unchanged. No full-size
memory/runtime trade-off is adopted by this gate.

```sh
python3 -B experiments/distributed-root/test_review.py
python3 -B experiments/gpu-staged-transport/test_review.py
python3 -B experiments/distributed-root/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-root/slurm-%j.log \
  experiments/distributed-root/run.sbatch roots-01
# Only after a user completion ping:
python3 -B experiments/distributed-root/analyze.py roots-01
```

Modules: gcc14.2, CUDA12.9, OpenMPI4.1.8; MPI test build uses
`CC=mpicc CXX=mpicxx CGO_ENABLED=1 go test -c -tags liftrs_cpp,liftrs_cuda,liftrs_mpi`.
The shared library uses the maintained matrix/preparation CUDA sources with
`-O3 -lineinfo -arch=sm_90`. Commands, exact Git revision, frozen build-input
hashes, stdout/stderr, progress, profiles and resource policy are saved under
`runs/roots-01/`. Binary/library/arrays remain outside Git at
`.cache/liftrs/distributed-root/roots-01/`, with checksums and regeneration via the
same runner/fresh run name. Prior caches/results are preserved.

After submission, hand off job ID and result paths without monitoring. Following
this passed completion review, use the measured complete root/transfer/merge costs to plan
scalable rank ownership and later generator/transform integration; separately
validate buffered persistence before choosing a full GPU/node allocation.

## Completed cross-node root review (SCALING-118)

The completion reviewer verified **621 frozen Git source files,20 evidence files
and35 retained cache files**, with no source overrides or modified producer
results. All15 commands exited zero: historical fixture hashes, ordinary partition
checks and race checks, tagged CUDA/MPI build, normal two-rank execution, both
CUDA memchecks and both profile summaries. Every one of18 normal measurements
(3 cases ×2 ranks ×warm-up/2 timed passes) and6 sanitized s=8 measurements passes.
Complete CPU roots, independent scalar witnesses, cross-rank saved local/peer
polynomial equality and full receiving-GPU merged/reference byte equality pass.
Both sanitizer processes report0 errors/0 leaked bytes. Explicit sorted membership
and unchanged canonical-field helpers are used throughout.

Medians below use two timed passes after a warm-up. Timings are slowest-rank
maxima; phase maxima may come from different ranks and must not be added. The
pipeline includes partition, local root, full host-polynomial exchange, GPU merge
and complete coefficient comparison; reference construction/setup and saved-array
writes are separate. Small sample count and profiling/logging prevent a controlled
speedup claim against earlier jobs.

| Field/input | Support exponents | Complete pipeline | Partition | Local root | Exchange/peer validation | GPU merge | Full coefficient comparison | Sent polynomial per node |
|---|---|---|---|---|---|---|---|---|
| s=8, genuine support | 7,381 | 12.816 ms | 0.996 ms | 6.090 ms | 0.326 ms | 5.392 ms | 0.154 ms | 14,764 bytes rank0;14,768 bytes rank1 |
| s=17, synthetic subset | 65,536 | 274.363 ms | 14.111 ms | 216.357 ms | 8.748 ms | 42.891 ms | 1.033 ms | 128 KiB +4 bytes per node |
| s=17, synthetic subset | 262,144 | 1.179829 s | 63.159 ms | 1.010393 s | 23.986 ms | 101.660 ms | 3.734 ms | 512 KiB +4 bytes per node |

Root construction dominates the largest pipeline; this result does not isolate
GPU kernel time from Go leaf/product work and copies. Exchange includes waiting
for the other rank to finish its root, count exchange, payload and peer checks.
In the largest timed passes, rank0 transfer start/completion events span0.780 and
0.824ms while rank1 spans23.900 and24.081ms. This rank-local evidence is consistent
with arrival skew; it is **not** evidence of a24ms intrinsic link transfer. No
network bandwidth estimate or bottleneck claim is derived from this phase clock.
The earlier matched transport probe remains the isolated communication evidence.

Largest case setup took75.1ms rank0/70.3ms rank1; complete CPU reference plus
scalar witnesses took13.920s/13.889s, outside pipeline timing. First s=8 setup
includes first executor/context-related work and took1.253s/0.979s. These and
compilation (19.598s CUDA,36.973s test binary), regression checks, profiling/export,
sanitized execution and persistence are included in the156s whole-job duration.
Do not interpret156s as root-only preprocessing latency. Profile summarization
alone took19.418s for the first rank; it is reporting overhead, not root work.

| Memory measure | Measured value |
|---|---|
| Normal rank0 peak process RSS | 401.129 MiB |
| Normal rank1 peak process RSS | 410.523 MiB |
| Small memcheck maximum rank RSS | 409.246 MiB |
| Largest owned GPU payload per card | 15.241 MiB |
| New pinned staging | None; host polynomials are Go-owned arrays |
| Sealed retained cache (35 files) | 20.498 MiB |

RSS is process high-water memory, not summed concurrent host/node/job peak.
Owned device payload excludes driver/context memory. Scheduler step MaxRSS is
launcher accounting and does not replace rank observations. All ranks stay well
within2 GiB RSS guards and128 MiB/card owned caps. Host request remains4 GiB/node.
The cache is `.cache/liftrs/distributed-root/roots-01/`, outside Git, with matching
binary/library/wrapper and complete normal/sanitized arrays. Checksums remain in
raw results. Four rank/run profiles are retained:
[normal0](runs/roots-01/normal-0.pprof), [normal1](runs/roots-01/normal-1.pprof),
[memcheck0](runs/roots-01/memcheck-0.pprof), [memcheck1](runs/roots-01/memcheck-1.pprof).
[Rank0 summary](runs/roots-01/profile-0.log) and
[rank1 summary](runs/roots-01/profile-1.log) mostly attribute native work to
`runtime.cgocall`, including the long CPU reference. They cannot isolate GPU
instructions or justify moving the whole host implementation to native code.

Verification: completion analyzer, four distributed reviewer regressions, five
staged reviewer regressions and three shared-review regressions pass; Python
syntax, documentation links and edited-text whitespace checks pass. Raw results,
profiles and logs are retained bytewise. No new job was submitted in this review.

Next planned measurement: buffered sequential checkpoint/final publication with
controlled writer concurrency, retaining compute/intermediate exchanges in RAM.
Scalable support/rank ownership, a multi-level merge tree, generator/Newton and
distributed transforms remain implementation/validation work before a full s=18
run. This bounded component does not update full-size memory/runtime projections
or change faster resident defaults; full-size communication/persistence costs
and final GPU/node-count selection remain unknown.

Subsequent gate: [buffered persistence77352](../preprocessing-persistence/report.md)
now passes. It verifies application-call coalescing and atomic/restart/race behavior,
but elapsed benefit is mixed and readback dominates; defaults stay unchanged.
The current next implementation is scalable rank ownership/multi-level merge and
full preprocessing/publication integration, as recorded in the scaling plan.

The subsequent [rank-owned merge tree](../root-merge-tree/report.md) now implements
bounded coordinator routing and uniquely owned subtree reductions using the same
arithmetic and extended optional MPI binding. Validation is pending on four logical
ranks sharing two physical GPUs; historical77285 timing/source evidence is unchanged.
