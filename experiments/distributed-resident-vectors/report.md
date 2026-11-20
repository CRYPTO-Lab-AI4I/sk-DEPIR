# Resident GPU vector and rank-transpose composition

Status: implemented; CPU/race and native/CPP/CUDA/MPI build checks pass. Bounded
GPU validation is prepared, not yet measured. Source follows `f360747`/`3c3a482`;
submission/raw results record the exact frozen revision.
[Readiness checks and retained artifacts](readiness/checks.json).
This follows [resident storage78392](../distributed-gpu-storage/report.md), which
passed208 normal +208 sanitized rank/pass outputs and all paired proofs, with
zero CUDA errors/leaks. We remain on step1: distributed products/FFTs.

## Implementation and checks

The new [`RankStorage.Vectors`](../../encoding/liftrs-cuda/rank_storage.go)
transforms complete vectors in place, using the same
[`Resident::vector_stage`](../../encoding/liftrs-cuda/preparation.cu) launcher and
`local_vectors` arithmetic as the existing host-tile VectorPreparation. The latter
retains its upload/extent validation and failure invalidation. No arithmetic or
FFT recurrence is copied. Both adapters share the bounded4096-vector launches,
progress and error events. Resident vectors require a canonical balanced global
matrix and an interleaved view; row affine points use the global owned origin.

Raw storage accepts arbitrary slot bits, so arithmetic first scans the selected
GPU view for values outside the field. It rejects before any slot mutation or
plan allocation. The existing cap-accounted flag stores the result. A valid call
lazily retains the shared GPU plan under the same cap. This adds one read of the
resident slot and a4-byte status transfer per stage, not a full host readback.
Runtime cost is unmeasured; it is not assumed free. Device arrays stay resident.

[`rankResidentVectors`](../../encoding/liftrs-shared/distributed_storage_cuda.go)
composes resident columns/affine rows and the shared transpose scheduler, reusing
one data slot and one scratch slot, preserving the third. Forward returns the
data slot in interleaved physical layout; inverse consumes that view, returning
interleaved coefficients. Chunked logical row-major I/O stays unchanged. Forward
accepts a row-major or interleaved input because its first step is transpose.
The host reference and resident path share `rankSeparableSchedule` for ordering:
forward columns then rows, inverse rows then columns. Input is consumed, and a
failed result is incomplete and must be discarded/reset.

This is a **separable stage**, not a Taylor-complete additive FFT. Resident Taylor
source access and pointwise/product composition remain next, followed by larger
matched products and preprocessing integration. Ordinary CPU clients have no CUDA
or MPI dependency, and no production default/field admission changes.

## Bounded experiment and independent validation

The [driver](../../encoding/liftrs-shared/distributed_resident_gpu_test.go) uses
s4 dimensions4/5/8 and s17 dimensions5/9/12 on2 ranks (N16–4096). Every forward
word must match the existing independent native/scalar oracle; inverse must equal
the original pattern. It checks unchanged third-slot sentinels and stable retained
allocation after first forward, including inverse reuse. Invalid physical layout
rejects; an out-of-field value uploaded through raw storage must reject arithmetic
without changing any input byte. Both slot widths are exercised.

[Runner](run.py), [Slurm entry](run.sbatch) and [reviewer](analyze.py) reuse the
validated composition launch, exact ownership/operation/proof review, independent
output hashes/bytes, source/evidence/cache seals and dual-stream CUDA summaries.
The [reviewer/selection tests](test_review.py) also exercise the shared fixture's
three tests and reject missing preservation/reuse/range flags. The suite first
reruns the old host-tile separable vectors (24 rank/pass rows), then the resident
normal and sanitized cases (24 rank/pass rows each). It does not repeat full FFT
or product benchmarks. No full-field table or recurrence is replaced.

Existing two-node admission: **2 HPC nodes /2 H200s /4 CPU cores /8 GiB requested
host RAM**, one GPU,2 cores and4 GiB/node; **128 MiB owned-device cap/card**,
**1 GiB RSS guard/rank**, no pinned buffers, **128 MiB retained cache /2 GiB
scratch /32 MiB file**, **720-second controller /15-minute scheduler**.
gnode03 excluded; other eligible nodes allowed. Shared executable/wrapper/library/
cwd and remote TMPDIR checks precede GPU work. No budget increase or monitoring.

```sh
python3 -B experiments/distributed-resident-vectors/test_review.py
python3 -B experiments/distributed-resident-vectors/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-resident-vectors/slurm-%j.log \
  experiments/distributed-resident-vectors/run.sbatch resident-01
# After the user's completion ping:
python3 -B experiments/distributed-resident-vectors/analyze.py resident-01
```

Normal results/profiles persist before sanitizer work in `runs/<name>/`, alongside
exact revision, commands, resource observations, phase/proof logs and checksums.
Reusable binary/library/output/reference cache is
`.cache/liftrs/distributed-resident-vectors/<name>/`, outside Git. Existing caches
are retained. Phase logs identify dimensions, passes/operations and increasing
work; native launches report progress about every30 seconds during long stages.
Timing includes logging/profiling. Per-pass clocks include lazy plan setup,
field checks, GPU vectors, target resets, packing, MPI/proofs and scattering;
oracle/input construction/upload, final readback/checks/fsync and profiles are
outside those clocks but inside command/controller time.

78392's larger storage-geometry suite took4.924 seconds normal /4.024 seconds
sanitized; its controller took123.480 seconds. Earlier separable host-stage
composition passed under the same limits. New resident stages add no new
arithmetic workload size (N<=4096). Each MPI arm keeps240 seconds (220-second
test timeout), within720 seconds controller, leaving headroom for setup and
sanitizer. Runtime/peaks are unknown until completion, not projected speedups.

## Memory and runtime boundaries

Three arrays and the127-word device transfer tile are unchanged. One GPU plan is
retained lazily; no new full GPU scratch array is needed. Intermediate host
transport stays bounded to127 words, plus the existing MPI runtime/proof buffers.
Diagnostic input/reference/output/sentinel arrays remain small additional host
allocations and are included in RSS. CUDA driver/context is excluded from owned
bytes. The largest4096-word case has48 KiB/card main arrays and1016 bytes/card
transfer tile before metadata/plan/context.

Removing whole-shard vector staging is an implementation fact; a measured speedup
or complete preprocessing memory reduction does not follow from it. Fulls18 field
admission, resident Taylor/FFT/products, discovery/Newton/root/generator state,
coordinated persistence and end-to-end scaling remain separate gates. Preserve
fast validated defaults; this experimental path does not introduce an automatic
slower fallback.
