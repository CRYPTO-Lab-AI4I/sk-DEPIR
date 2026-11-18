# Cross-node GPU root products and polynomial merge

Status: implemented and locally checked; **two-node validation not yet run**.
Work SCALING-117. This follows the passed
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
executor cap; exact peak rank/device memory remains unmeasured. Duplicated support,
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
completion review, use the measured complete root/transfer/merge costs to plan
scalable rank ownership and later generator/transform integration; separately
validate buffered persistence before choosing a full GPU/node allocation.
