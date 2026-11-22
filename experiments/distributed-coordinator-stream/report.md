# Bounded coordinator diagnostics in the genuine pipeline

Status: **retry79213 passed**, `coordinator-02`, producing revision `53271bf`.
[Retry receipt](submission-retry.json), [scheduler log](slurm-79213.log),
[raw results](runs/coordinator-02/results.json), [independent review](runs/coordinator-02/analysis.json),
[summary](runs/coordinator-02/review-checks.json). Original79202 packaging failure
and receipt remain below; no math ran in that failed allocation.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/checks.json),
[reference manifest](reference-inputs.json).
Retry results: `runs/coordinator-02/`; reusable artifacts/snapshots/profiles at
`.cache/liftrs/distributed-coordinator-stream/coordinator-02/`, outside Git.
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
 experiments/distributed-coordinator-stream/run.sbatch coordinator-02
# After completion ping:
python3 -B experiments/distributed-coordinator-stream/analyze.py coordinator-02
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


## Failed launch79202 and packaging correction (SCALING-167)

79202 stopped after **13.676 seconds**, on HPC gnode10/gnode11, before builds,
GPU setup or either cold/resume arm. Its source snapshot intentionally excludes
experiment JSON; the newly required runtime reference manifest was therefore
missing. The local test supplied that file directly and missed the packaging
boundary. This is a launcher bug, not a GPU arithmetic/restart failure. Verified
785 producing source identities,8 original evidence files and0 cache files;
all completed environment/reviewer commands passed. No encoding/preprocessing
runtime or memory measurement was produced. Existing78917 arithmetic and78915
restart results remain valid.

Move the active manifest to the experiment root and explicitly include only that
reviewed JSON path in the shared snapshot selector. Other experiment JSON,
readiness metadata and results remain excluded; source size/file caps remain.
A regression checks the actual selector, and a pre-submission check creates the
real Git-blob snapshot, loads the manifest from it and verifies/copies every real
reference input. This covers the missing boundary without another GPU calibration
job. Python source/packaging changed only; Go/CUDA sources and prior checks remain
unchanged. Same resources and3–6-minute estimate for the retry, excluding queue.
Keep the failed run and its original receipt intact; retry uses a fresh run/cache.


The [real committed-snapshot preflight](readiness/retry-checks.json) passed on
source `f6e743e`:786 source files /3,623,064 bytes,including the manifest,then all
12 real reference inputs /69,746 bytes verified and copied read-only. Reviewer
tests executed successfully from that snapshot. No local GPU/MPI initialization.


## Completed cold/restart integration79213 (SCALING-168)

Both genuine s4/s6 cold and fresh-prefix8-resume arms passed, including retained
reference flags, streamed support, all route proofs, complete P*G identity,
canonical code/generator bytes and process identity changes. Producing revision
`53271bf`; reconciled786 source/41 original evidence/203 cache files. HPC
gnode09/gnode11,2H200/4CPU/8 GiB requested RAM,unchanged limits. Controller
**2 min15.180 seconds**; retained cache **15.355 MiB**. No guard fired.

| Arm | s4 complete profiled case | s6 complete profiled case | Peak process RSS | Observed post-call owned GPU/card |
|---|---|---|---|---|
| Cold | 1.656894 seconds | 0.511963 seconds | 373.082 MiB | 0.075245 MiB |
| Fresh-process prefix8 resume | 1.531600 seconds | 0.511681 seconds | 371.145 MiB | 0.075245 MiB |

Times include file checks/copies,setup,logging,proofs,persistence and publication.
Independent native input generation is separately measured reusable setup above.
GPU observations exclude context/driver/transients/earlier factor peaks. Tiny
single samples do not establish stable speedup or measurable fullsize RSS savings;
array removal is an ownership result. No complete coordinator native polynomial
oracle or support archive is created in this explicit mode. Rank-local copies
and the bounded small-reference format remain. This is not fulls18 support.
Next:four-worker FFT/product correctness using the same two GPUs,then more-rank
root/discovery integration and representative scaling.
