# Bounded coordinator diagnostics in the genuine pipeline

Status: **prepared; integrated validation pending**. Freeze implementation before
submission; actual revision/source identities are recorded by the shared controller.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/checks.json),
[reference manifest](readiness/reference-inputs.json).
Expected results: `runs/coordinator-01/`; reusable artifacts/snapshots/profiles at
`.cache/liftrs/distributed-coordinator-stream/coordinator-01/`, outside Git.
Follows [passed genuine restart78915](../distributed-restart-pilot/report.md) and
[passed bounded s18 arithmetic78917](../s18-field-admission/report.md).

## Ownership change and verification

The genuine s4/s6 pilot previously recreated a complete native CPU reference and
full exponent arrays on rank0. Although independent diagnostics are useful, those
copies prevent larger distributed trials from demonstrating sharded memory use.

`PIR_PIPELINE_REFERENCE_FROM` optionally selects retained native reference files.
Empty preserves native-oracle behavior. Retained mode requires partitioned discovery
and streamed publication, with unchanged deterministic seed109 and code parameters
m2/t2. It loads only bounded metadata: field/K/N/samples, entropy identity and exact
file hashes/lengths. Known filenames and canonical field/size admission reject
unknown/path-bearing entries; data files are checked using fixed buffers before
GPU work. Checksums detect corruption, not authentication: trusted native/source
provenance is required.

[Shared pilot integration](../../encoding/liftrs-shared/distributed_pipeline_gpu_test.go)
skips complete native discovery/P/G/spectrum reconstruction during the distributed
run. [Discovery](../../encoding/liftrs-shared/distributed_discovery_gpu_test.go)
streams sorted support intervals through the existing bounded tile and canonical
codec, validating strict increasing order/range/count and the full support digest.
It retains no complete coordinator support archive. Rank-local union/raw/orbit
lists remain. The old diagnostic gather is retained when no sink is selected.

Reference P/G/spectrum/code files are copied with16 KiB buffers; comparisons read
bounded blocks. The complete distributed P*G=1+X^(N−1), all support/sample/entropy
checks, per-prefix commits/reload, spectrum and canonical/generator comparisons
remain. Fresh MPI processes resume prefix8 and repeat the remaining checks without
discovery/root work. Existing route/proof reviewers are reused, with retained-mode
flags checked explicitly. No GPU kernel/recurrence or client API changes.

Two arms in one allocation: cold genuine s4/s6 streamed support, then new MPI
processes resume that cold prefix8. No repeated sanitizer:78915 validated the
maintained kernels and78917 additionally sanitized s18 arithmetic; changes here
are file/metadata/host ownership boundaries, covered by focused CPU/race checks.
This is still a two-rank small pipeline, not fulls18 or arbitrary node-failure
recovery. Larger ranks/multi-level root ownership and representative scaling remain.

## Memory–time and storage costs

At current32-bit extension width, the native oracle keeps P/G/spectrum payload
8N+4 bytes, two full exponent copies cost16K bytes, and support serialization costs
4K bytes. The explicit mode removes at least **8N+20K+4 coordinator bytes** of
these payloads, plus unmeasured native discovery/setup temporaries. It uses the
existing4096-word network tile, fixed16 KiB file buffers and the existing bounded
canonical codec buffer instead. Each rank still retains its own root/support/
coefficient shards. These are ownership-derived savings, not measured RSS gains;
CUDA context/runtime dominates these tiny instances.

Validation reference files become reusable inputs: **69,746 bytes (68.111 KiB)** total for both
cases (including bounded metadata; exact byte count in the manifest/runner resources). The job retains
one additional verified input copy outside the arm directories; each arm's ordinary
reference outputs already existed in the prior diagnostic layout. Input reads,
checksum scans and bounded copying replace native reference computation. Total
runtime effect is **unknown**, and first-generation native reference setup remains
part of the complete validation workflow. This does not promise fullsize native
reference availability or introduce this test layout as a production requirement.
Old native/gather defaults remain; no slower memory-saving mode is automatic.

## Reference setup and reproducibility

`TestExportPilotReferences` separately generated the native inputs in **0.19
seconds** (test elapsed, including reference construction, writes and fsync; local HPC
login environment, no GPU/MPI initialization). It uses the maintained native
preprocessing and saves code, P/G/spectrum, support and metadata. Every data file
matches the independently retained78915 cold native/support hashes; manifest
records exact sizes/hashes. Inputs are checked and copied into the frozen job
cache before launch, then made read-only. The allocation preflights readability
on both participating nodes; each worker admits metadata/files before GPU setup.
No plaintext/key is involved.

Build the tagged test binary as recorded in readiness, create a fresh empty export
directory and run:

```sh
PIR_PIPELINE_REFERENCE_EXPORT=.cache/liftrs/distributed-coordinator-stream/readiness/reference-checked \
 .cache/liftrs/distributed-coordinator-stream/readiness/coordinator.test \
 -test.run '^TestExportPilotReferences$' -test.v -test.timeout=20s
python3 -B experiments/distributed-coordinator-stream/test_review.py
python3 -B experiments/distributed-coordinator-stream/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-coordinator-stream/slurm-%j.log \
 experiments/distributed-coordinator-stream/run.sbatch coordinator-01
# After completion ping:
python3 -B experiments/distributed-coordinator-stream/analyze.py coordinator-01
```

Do not overwrite retained inputs; regenerate in a fresh directory and compare
all checksums with the tracked manifest before reuse. Focused checks cover
corrupt/short files, exact filename sets, metadata shape, source-preserving copies,
wide/narrow ordered support, duplicates/range rejection and short writes. Tagged
build and no-GPU native export pass before submission; ordinary builds remain
CUDA/MPI-free. Local OpenMPI compiler socket warnings are retained separately
from successful build/test outcomes.

## Unchanged admitted resources and measurement boundaries

**Two HPC nodes /two H200s /four CPU cores /8 GiB requested host RAM**; one GPU,
two cores and4 GiB RAM per node. **128 MiB owned/card /1 GiB RSS/rank**, no pinned
buffers, **128 MiB retained cache /2 GiB scratch /32 MiB file**, **720-second
controller /15-minute scheduler**. Exclude gnode03, otherwise any eligible nodes.
No resource increase. Previous cold/restart78915 completed in2 min2.635 seconds;
estimate **3–6 minutes excluding queue**, with unchanged240-second launcher /
220-second test limits and margin for metadata/file checks and fresh builds.

Normal profiled case timing includes reference-file I/O/checks, GPU setup, all
math/proofs, persistence and publication. Native input generation is separate
reusable setup above; it must not be silently excluded from first-use validation
cost. No stable speedup/RSS claim or fullsize extrapolation is made. Flush existing
timestamped phases/pass/prefix progress and resource logs; retain completed arms,
profiles and all source/evidence/cache identities before later work. Submit and
yield without monitoring; review only after the user's completion ping.
