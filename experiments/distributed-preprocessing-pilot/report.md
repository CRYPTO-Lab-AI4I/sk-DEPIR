# Integrated genuine-code distributed preprocessing pilot

Status: **submitted as78898**, run `pilot-01`, implementation `32d0d12`.
[Submission receipt](submission.json), [scheduler log](slurm-78898.log);
results will be retained under `runs/pilot-01/`. Local checks passed; no GPU outcome yet.
[Runner](run.py), [Slurm entry](run.sbatch), [reviewer](analyze.py),
[boundary regressions](test_review.py), [readiness](readiness/checks.json).
Producer revision/source manifests are recorded at allocation start.

This work replaces a sequence of isolated component experiments with one assembled
pipeline gate. Existing FFT/product/transport kernels already passed their bounded
checks. Add validation for new integration boundaries and reuse those checks;
additional kernel calibration is warranted only for a concrete blocker.

## What is integrated

The [opt-in driver](../../encoding/liftrs-shared/distributed_pipeline_gpu_test.go)
runs genuine `Spec{s,2,2}` instances, deterministic discovery seed109:

1. Existing GPU discovery on the coordinator, compared with independent native
   CPU discovery. Route disjoint Frobenius-closed support subsets to workers.
2. Existing local GPU root-factor construction on each worker. Route coefficient
   ownership through the existing bounded operand scheduler.
3. A shared [resident executor](../../encoding/liftrs-shared/distributed_executor_cuda.go)
   computes the final root product, every Newton generator product, the full
   independent P*G identity, and the final canonical spectrum. P/G stay sharded
   throughout these operations; no complete intermediate polynomial is gathered.
4. Save each completed Newton prefix on both ranks with checksums and a
   support-bound commit manifest. Reload one interior prefix before continuing.
5. Only at final publication, assemble P/G/spectrum on the coordinator. Compare
   every coefficient and the complete canonical code artifact byte for byte with
   the independent existing CPU preprocessing implementation.

The ordinary host inverse and distributed generator share the completed-prefix
[Newton schedule](../../encoding/liftrs-shared/reciprocal_schedule.go). The host
arithmetic and callback/checkpoint behavior are preserved. Characteristic two
uses g² followed by a truncated P*g² product. The new
[sharded squaring adapter](../../encoding/liftrs-shared/distributed_square.go)
reuses existing base-field multiplication and checked operand transport to move
coefficient i to2i; odd output coefficients stay zero. It uses bounded tiles,
preserves the source and adds no extra FFT for squaring. Products always admit
the full convolution degree even when the returned prefix is truncated.

One admitted three-array GPU workspace is reused across all products/transforms.
Operations have distinct communication identities; expected route review spans
all repeated calls, not one synthetic final product. No new CUDA arithmetic,
second FFT implementation, public API or production default is introduced.
Ordinary client builds remain independent of CUDA/MPI.

## Necessary checks, grouped in one allocation

Two arms total:

| Arm | Genuine instances | Checks |
|---|---|---|
| Normal | s4,N256 and s6,N4096 | Complete discovery/support, P/G/identity/spectrum, every prefix/route and canonical artifact; Go CPU profiling |
| CUDA memcheck/leak | s4,N256 | Same complete pipeline for the smallest instance; no separate kernel sanitizer series |

An ordinary CPU test validates sparse squaring ownership for1/2/3/5ranks,
including odd and empty partitions, and the shared prefix schedule/error boundary.
Existing independent polynomial-product/division tests protect the host Newton
refactor. New artifact tests reject a corrupt published code, wrong support
binding, and a corrupt saved prefix even when its checksum is changed to match.
The actual native CPU fixtures pass local preflight without CUDA/MPI initialization.

The independent reviewer compares every final output/prefix coefficient with the
saved complete native oracle; checks the full identity and exact canonical code
bytes; verifies support ownership, all checkpoint commit hashes and bindings;
and reuses existing exact transpose/Taylor route reviewers for all executor calls.
Operand remap/publication packets must match the complete expected source/target
schedule and paired before/after payload proofs. Source revision, raw evidence
and cache hashes are reconciled by the existing analyzer.

Only the small instance is sanitized. This is correctness/integration validation,
not a CPU/GPU performance comparison or a new end-to-end latency baseline.
Timings include independent oracles, checksums, progress logging, profiling,
checkpoint fsync/reload and final artifact assembly. Later performance measurements
must separate those costs explicitly; no fullsize speedup is inferred here.

## Scope that remains deliberately bounded

Discovery uses the existing coordinator GPU implementation; it is not partitioned
across nodes. The CPU reference exists on the coordinator for this small pilot.
Final publication assembles small full arrays there. Neither is a fulls18 memory
layout. Two ranks give one root-reduction level; this does not validate a larger
multi-level reduction tree or more GPU ranks. Save/reload is within one allocation;
fresh-process/missing-node recovery is not validated. No s18 arithmetic admission,
fullsize support, or full distributed preprocessing feasibility is implied.

The goal is an assembled genuine pipeline first. After a pass, replace the
coordinator-only discovery/publication boundaries as required by the admitted
larger workload and extend arithmetic/rank/scale support without another broad
series of tiny kernel comparisons.

## Resources, logging and reproducibility

Same admitted **two HPC nodes / two H200s / four CPU cores / 8 GiB requested host
RAM** (one GPU/two cores/4 GiB per node), **128 MiB owned GPU/card**,
**1 GiB RSS/rank**, no pinned buffers, **128 MiB retained cache / 2 GiB scratch /
32 MiB per file**, **720-second controller / 15-minute scheduler**. gnode03 excluded;
other eligible nodes allowed. GPU discovery/factor engines close before the
resident executor is constructed, avoiding overlapping large workspaces.
No resource increase or slower default memory strategy is selected.

At the largest pilot size each host coefficient shard is8 KiB. Newton refinement
retains root/current generator, square, masked input and result shards; checkpoint
reload adds bounded file/readback buffers. Final gathered P/G/spectrum total48 KiB
before serialization/reference arrays. This is negligible beside the admitted
runtime/context headroom, and not a fullsize memory claim. Sparse squaring adds
one owned host output shard and at most16 KiB transport tile, plus metadata;
GPU arrays are unchanged. Communication/runtime costs are unmeasured until the
assembled run; no cost-free optimization is claimed.

Conservative full-term Taylor plans give at most1488 s4 and16555 s6 endpoint log
records/rank for all products/spectrum. At800 bytes/record their combined payload
is below14 MiB, leaving room for other records below32 MiB/file. Projected cache
below32 MiB fits128 MiB. These are bounds/projections, not observed new-run sizes.
The new native fixture test completes in0.06 seconds. Prior two-node suites took
about2–3 minutes including compilation/profile startup. Budget **3–6 minutes
estimated** for this integrated suite; keep the validated240-second launcher /
220-second per-arm test cutoff and controller/scheduler headroom.

Reuse node-by-node shared binary/wrapper/library/cwd/TMPDIR preflight and UCX
registration-cache controls. Logs flush phase starts/completions, actual Newton
prefix progress, transform/pass/packet identifiers and resource observations.
Completed prefix files are fsynced on every rank, then the support-bound commit
manifest is fsynced before proceeding. Completed outputs remain if later work
fails. No agent monitoring after submission; wait for the user's completion ping.

```sh
python3 -B experiments/distributed-preprocessing-pilot/test_review.py
python3 -B experiments/distributed-preprocessing-pilot/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-preprocessing-pilot/slurm-%j.log \
  experiments/distributed-preprocessing-pilot/run.sbatch pilot-01
# After the user's completion ping:
python3 -B experiments/distributed-preprocessing-pilot/analyze.py pilot-01
```

Raw results/profiles/logs: `runs/pilot-01/`. Reusable source/binaries and prefix/
artifact evidence: `.cache/liftrs/distributed-preprocessing-pilot/pilot-01/`, outside
Git with retained checksum manifests. Prior caches/profiles are preserved.
