# Genuine discovery feeding hierarchical root products

Status: prepared; GPU allocation not yet submitted. Parent revision5268f7b;
producing revision and source hashes will be sealed by the controller.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer regressions](test_review.py), [local readiness](readiness/).
After submission,the receipt links the exact output directory and scheduler log.
Reusable frozen source,binaries,polynomial shards and profiles stay outside Git
under `.cache/liftrs/distributed-discovery-hierarchy/<run>/`.

## Purpose and shared implementation

Connect genuine sampled support discovery to the now-validated four-worker
[subgroup root hierarchy](../distributed-root-hierarchy/report.md). This advances
both ownership stages together; it does not repeat an isolated kernel sweep.
Newton,complete identity,spectrum and restart/publication beyond two workers
remain the following integration boundary. Fulls18 preprocessing remains unrun.

The maintained [discovery implementation](../../encoding/liftrs-shared/distributed_discovery_gpu_test.go)
now exchanges phase/sequence/rank-bound integer counts across all rank pairs in
one global order. Grids,good-monomial lists and union bits remain rank-local.
Rank0 generates coefficients in exactly the original global-good order and sends
bounded tiles to each owner. Closed-orbit roots route directly from local source
lists to deterministic owners; canonical support streams to rank0 by source rank.
Both members of each Frobenius orbit keep one owner; load balance is not promised.
This replaces pair-specific loops without introducing a second discovery algorithm.
The original two-worker pipeline calls the same implementation.

The [shared hierarchy helper](../../encoding/liftrs-shared/root_hierarchy_gpu_test.go)
accepts these locally discovered roots. Synthetic s18 validation and genuine
s4/s6 share leaf construction,subgroup lifetime,shard redistribution,the product
executor and native leaf/subtree/final checks. Actual arithmetic slot width is
selected from s; this is not a separate size-specific algorithm. The maintained
product reviewer now uses4-byte extension slots through s16,8 bytes above it,
including the genuine s6 case; canonical base coefficients remain4-byte slots.

## Workload and correctness

| Genuine case | Grid | K | Samples | Leaf polynomial counts | Child domains | Final domain |
|---|---|---|---|---|---|---|
| s4 | 256 positions | 97 | 1 | 37,25,27,12 | 64,64 | 128 values |
| s6 | 4096 positions | 1910 | 1 | 681,462,480,291 | 2048,1024 | 2048 values |

These dimensions were measured by the native preflight,not assumed balanced.
Spec is `(field degree s,two evaluation coordinates,curve degree2)`; deterministic experimental RNG seed109
is a reproducibility fixture,not a production randomness claim. The native
reference consumes the same original random stream independently; validate exact
support,sample count and entropy digest. Each rank's owned closed support is checked
before GPU leaf construction. Three scalar root witnesses supplement independent
native leaf/subtree/final polynomials. Small replicated native references and the
diagnostic archive are explicitly bounded and are not a fullsize memory strategy.

The Python reviewer reconstructs every expected sample distribution,root route,
archive packet,transpose and Taylor route. It verifies both payload proofs and
actual inter-level source bytes,all coefficients,padding,degree and final shards,
plus communicator-close and parent-reuse events. It shares reviewers with the
existing two-worker pilot and synthetic hierarchy. Completed-job analysis seals
source,evidence and reusable cache hashes. Checksums detect corruption;they do
not authenticate maliciously rewritten artifacts.

## Resources and timing

Unchanged admitted envelope: **two HPC nodes/two H200s/four CPU cores/8 GiB total
requested RAM**,four workers/two per card,one CPU/worker; **64 MiB owned GPU per
worker =128 MiB/card aggregate**,**1 GiB RSS/worker**,zero CUDA-pinned allocation;
**128 MiB retained cache,2 GiB scratch,32 MiB/file**; **720-second controller and
15-minute scheduler cutoff**. Other eligible HPC nodes allowed,gnode03 excluded.
CUDA contexts/driver internals are additional to the owned-array cap.

One profiled normal arm covers both genuine cases and all tree levels. No CUDA
arithmetic changes;the prior genuine pipeline and hierarchy retain clean small
CUDA memchecks. Host/race/native preflight and reviewers cover new glue; no extra
sanitizer-only allocation. Estimated complete controller time **2–4 minutes excluding
queue**,based on79376 taking2 min20 seconds including builds/profiles. New composed
runtime remains unmeasured. Profiled complete case time includes native references,
discovery,communication,GPU/communicator setup,products,proof logging and I/O;
controller runtime additionally includes build/preflight and profile summaries.
No fulls18 runtime or memory prediction is derived from these tiny cases.

Memory–time costs: at most64 integer count words per rank and pairwise control
traffic (quadratic in worker count),but no global support is needed for routing.
The current bounded validation additionally replicates native references.
Intermediate shards return to host and redistribute/upload between merge levels;
each subgroup releases GPU/staging before the next. Communication/setup overhead
and balance effects at scale remain unknown. No production default or slower
memory-saving mode is selected; existing faster paths are retained.

## Reproduction and local evidence

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9,H200. Ordinary client builds still need
neither MPI nor CUDA. Compile all tagged interfaces without initializing MPI/CUDA;
run actual native support geometry and legacy routing tests before submission.
Freeze the committed source and run reviewer tests from that snapshot. The admitted
allocation checks shared executable,wrapper/library/input/output paths on every
participating worker before expensive phases. Durable timestamped phase,discovery
sample,tree-level/transform/pass/tile and resource records reuse the validated
runner and native MPI/GPU logging. Completed work/profiles persist before review.
Logging/proof capture is included in diagnostic timings; no isolated logging-cost
measurement or silent timing-boundary change is claimed.

```sh
python3 -B experiments/distributed-discovery-hierarchy/test_review.py
python3 -B experiments/distributed-discovery-hierarchy/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-discovery-hierarchy/slurm-%j.log \
 experiments/distributed-discovery-hierarchy/run.sbatch genuine-01
# After completion notification:
python3 -B experiments/distributed-discovery-hierarchy/analyze.py genuine-01
```

Local native preflight passed genuine shapes and legacy/synthetic geometry/routing;
[tagged compile log](readiness/build.log), [native log](readiness/native.log),
[discovery reviewer](readiness/discovery-review.log),
[hierarchy reviewer](readiness/hierarchy-review.log), [new reviewer](readiness/review.log).
GPU/MPI combination remains pending; bounded correctness is not fulls18 support,
full-pipeline generator validation or node-loss recovery.
