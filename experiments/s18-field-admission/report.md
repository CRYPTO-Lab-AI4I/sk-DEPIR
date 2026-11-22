# Bounded canonical s18 field admission

Status: **passed78917**, `field-01`, producing implementation `3a8f317`. [Receipt](submission.json), [scheduler log](slurm-78917.log). Actual
producing source/revision is recorded by the frozen allocation controller.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness checks](readiness/checks.json).
[Raw results](runs/field-01/results.json), [independent analysis](runs/field-01/analysis.json),
[review summary](runs/field-01/review-checks.json). Results retained under `runs/field-01/`; reusable snapshots, binaries and
artifact shards under `.cache/liftrs/s18-field-admission/field-01/`, outside Git.
This follows [passed genuine s4/s6 restart78915](../distributed-restart-pilot/report.md).

## Scope and implementation

Canonical s18 has18-bit base-field coefficients in4-byte slots and36-bit extension
values in8-byte slots. Existing native arithmetic/representations support these
values; explicit server/GPU parameter guards stopped at17. Admit s18 metadata in
`NewCodePreprocessor`, support persistence and the resident CUDA constructor,
without removing full-domain allocation checks. Legacy CUDA adapter stays s<=16.

[Bounded table construction](../../encoding/liftrs-shared/preparation_products.go)
is shared: ordinary CPU callers keep their existing s<=17 admission; explicit
GPU server callers admit18. Immutable log/duplicated-power payload is
3,145,720 bytes (less than3 MiB) per set, on host and each participating GPU,
versus1,572,856 bytes at17. Kernel recurrences, canonical polynomial/delta/omega,
CPU client behavior, backend selection and default modes do not change.

[Drivers](../../encoding/liftrs-shared/s18_field_gpu_test.go) reuse maintained
resident FFT/product implementations, independent native full-FFT/scalar Horner
oracles, independent schoolbook products, source-preservation and range/alias/
layout rejection checks. FFT patterns set the high extension bit (bit35), so
bounded subspaces do not accidentally test only base-field arithmetic. Direct
full-extension pointwise checks independently exercise high components as well.
Fixed deterministic patterns/seeds are in the shared drivers; all exact bytes
and route proofs are retained. Reviewers accept explicit case lists through a
shared optional argument; historical experiment defaults remain unchanged.

One allocation, four arms:

| Arm | Field | Evaluation values | Verification |
|---|---|---|---|
| FFT normal | s18 | 32 and4096 | Full forward/inverse, exact native/scalar output, two ranks |
| Product normal | s18 | 32 and4096 | Two nonaliasing product shapes, exact schoolbook, full-extension pointwise |
| FFT memcheck | s18 | 32 | Same FFT checks plus CUDA errors/leaks |
| Product memcheck | s18 | 32 | Same product checks plus CUDA errors/leaks |

These are synthetic component grids; genuine s18 has2^36 values. No fulls18
discovery, generator, canonical code artifact, fullsize runtime or memory result
is claimed. Genuine assembled pipeline is currently verified only at s4/s6.
More ranks/multi-level roots, partitioned diagnostic ownership and representative
scaling are separate remaining integration work.

## Resources, trade-offs and preflight

Same admitted **two HPC nodes / two H200s / four CPU cores /8 GiB requested host
RAM**, one H200/two CPU cores/4 GiB RAM per node. **128 MiB owned-device/card**,
**1 GiB RSS/rank**, zero pinned buffers, **128 MiB retained cache /2 GiB scratch /
32 MiB file**, **720-second controller /15-minute scheduler**. Exclude gnode03;
other eligible nodes permitted. No resource increase or fullsize launch.

Table memory increases by about1.5 MiB/set at the larger field; no memory-saving
mode, additional FFT, storage strategy or host/device transfer is introduced.
Setup/table-building cost at18 is not yet benchmarked; host correctness checks
are measured separately, and normal kernel timings exclude oracle/setup/capture
as specified by the shared drivers. No cost-free optimization or stable speedup
is claimed. Whole controller includes builds, references, profiling/logging,
proof files and persistence. Sanitized timings are labeled separately.

Prior s17 integrated jobs complete around2–4 minutes including builds; the s18
suite is bounded to fewer grids and only32-value sanitized grids. Estimated
controller **3–6 minutes excluding queue**; each launcher retains240 seconds /
220-second test timeout. Maximum4096-value GPU array payload is below0.1 MiB
before lookup/plan metadata; metadata stays well below128 MiB/card. Cache expected
below32 MiB. These are budgets/estimates, not s18 measurements.

Build both native/MPI test driver and changed CUDA source before submission.
Run independent native s18 full-FFT/scalar oracle locally without GPU/MPI init;
verify native full-discovery admission fails its1 MiB native cap before any huge
array allocation. Explicitly test wide support roundtrip on a tiny synthetic
closed orbit and rejection under inadequate host admission. Exhaustively compare
all s18 left operands with selected deterministic right operands against the
independent bitwise multiplication oracle, preserving the ordinary CPU default.

Reuse allocation shared-path/cwd/library/permissions preflight on every node,
staged UCX configuration and guarded controller. Save completed arms/profiles
before later arms. Timestamped per-pass/phase/tile logs plus bounded resource
observations are retained; timings include logging/proof overhead. No monitoring
after submission; results are reviewed only after the user's completion ping.

```sh
python3 -B experiments/s18-field-admission/test_review.py
python3 -B experiments/s18-field-admission/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/s18-field-admission/slurm-%j.log \
  experiments/s18-field-admission/run.sbatch field-01
# After completion ping:
python3 -B experiments/s18-field-admission/analyze.py field-01
```

CUDA12.9, GCC14.2/OpenMPI4.1.8, Go1.27.3 (local pinned toolchain); source/evidence/
cache hashes and actual revision, node identities and software resources recorded
by the controller. Ordinary builds remain CUDA/MPI-free. Local compiler probes
may print restricted socket warnings; distinguish these from build/test failures.

## Reviewed result (SCALING-166)

**All four arms passed** on HPC gnode14/gnode15, two H200s/four CPU cores/8 GiB
requested host RAM, within unchanged limits. Controller **2 min25.876 seconds**
including builds, CPU/race checks and profiles. All780 source identities,51
original evidence files and52 cached files reconciled; cache **16.478 MiB**.
16 normal rank/pass/product outputs and8 sanitized outputs matched exactly.
All four sanitized processes reported **zero CUDA errors and zero leaked bytes/
allocations**. Full-extension pointwise and every transpose/Taylor route passed.

| Arm | Launcher runtime | Peak process RSS | Maximum reported owned GPU/card |
|---|---|---|---|
| FFT normal | 2.923 seconds | 378.676 MiB | 3.118 MiB |
| Product normal | 2.522 seconds | 376.270 MiB | 3.118 MiB |
| FFT memcheck | 2.374 seconds | 414.059 MiB | 3.071 MiB |
| Product memcheck | 2.229 seconds | 411.086 MiB | 3.071 MiB |

Largest4096-value normal samples: forward **38.957 ms**, inverse **36.906 ms**;
product shapes **95.070 ms /93.402 ms**. These are profiled, single diagnostic
samples with the existing proof/capture boundaries, not full workflow latency or
stable speedup. GPU figures exclude CUDA contexts/driver and unobserved transients.
Native table/serialization checks and CPU/race orchestration also passed. No guard
fired. This establishes bounded canonical s18 arithmetic, **not** fulls18
preprocessing, discovery, root tree, generator or production support.

Next integrated step removes coordinator diagnostic full-array copies in the
small genuine pipeline using streamed support and retained native reference files;
then larger ranks/multi-level roots and representative size scaling remain.
