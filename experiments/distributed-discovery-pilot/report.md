# Partition discovery in the assembled genuine pipeline

Status: implemented and locally compiled; integrated GPU validation pending.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[focused reviewer tests](test_review.py), [readiness](readiness/checks.json).
Producer revision and complete source identities are recorded by the controller.
This follows [passed genuine pilot78898](../distributed-preprocessing-pilot/report.md).

## What changes

The same genuine s4/s6 pipeline now selects `PIR_PIPELINE_DISCOVERY=distributed`.
The prior coordinator mode remains the default when that experimental variable is
omitted. No public preprocessing default or CPU-client dependency changes.

Each rank enumerates only its contiguous monomial range through the existing
native criterion, retains its local good-index list, receives sampled coefficients,
and owns its grid and support-union bits. Rank0 supplies the same random byte
stream in global-good order using bounded tiles. Sampling byte width, decoding,
mask and default sample limit are shared with ordinary discovery; compare the
exact entropy-stream digest and sample count with the independent native oracle.

GPU tensor evaluation uses `RankStorage.BaseVectors`: axis2 selects the existing
Fq metadata and zero affine origin in the existing vector kernel. Both axes use
that same plan, with the maintained checked rank transposes between them. The
result stays resident through the existing inverse extension transform. Only the
owned interpolated coefficients return to the host; no global grid is gathered.
Axes0/1 and the ABI signature retain their existing meaning. The new base mode
validates full q-by-q shape and base-field range before arithmetic/mutation.
Older libraries reject the new explicit axis; no fallback is selected.

Local support unions use exact integer collective counts to make identical
stopping decisions. Discovered roots route directly to an owner determined by
`min(i, q*i mod (N-1)) mod ranks`; both Frobenius partners choose the same owner.
Each receiver repeats range/ownership/closure/uniqueness checks. This rule does
not promise equal loads: native preflight predicts s4 roots62/35 and s6 roots1159/751.
Admit actual loads and headroom before larger execution; do not assume halves.

The root/Newton/identity/spectrum/checkpoint/canonical-artifact stages reuse the
passed integrated pilot. Only afterward is a small complete support archive
gathered for independent comparison and persistence; it does not drive root
ownership. Final artifact assembly and the diagnostic CPU oracle still use the
coordinator. Those full diagnostic arrays are not a fulls18 memory layout.
The current integration is two ranks and one root-reduction level, with in-process
prefix reload rather than fresh-process/failure recovery. No s18 admission claim.

## Explicit memory/transfer/time costs

Let r be participating ranks and w extension-slot bytes (4 for these instances,
8 at s17). Payload models, not measured new peaks:

| Cost | Previous coordinator discovery | Partitioned prototype |
|---|---|---|
| Principal GPU grids | Three full N*w arrays on one GPU | Three N*w/r arrays per GPU, same aggregate grid payload |
| Host active grid payload | One in-place N*w grid | Per rank4*N/r sampled-base input +N*w/r returned coefficients |
| Host union bitmap | N/8 bytes on coordinator | AboutN/(8*r) per rank |
| Good monomial list | Complete list on coordinator | Local index-range lists; cardinalities can differ |
| Host/device grid traffic per sample | Upload/download for tensor and interpolation separately, about4*N*w | Upload once and download once, about2*N*w; tensor result stays on GPU |
| Node traffic | No distributed grid transpose | Two tensor plus two interpolation transposes, Taylor windows, bounded random/index packets |
| Diagnostic archive | Complete support on coordinator | Still retained for this small oracle/persistence gate |

For two ranks at w4, the two host grids together have the same per-rank payload
as the old coordinator grid, but twice the aggregate host grid payload. At w8,
per-rank payload is25% below the old coordinator grid while aggregate is50%
higher. GPU per-card grid payload halves; complete runtime is unknown because
communication and host enumeration/routing change. This is an explicit experimental
mode, not adoption of a slower memory strategy. Measure the assembled case before
selecting a production mode or admitting larger host/GPU budgets.

At largest current N4096: each coefficient grid/sample shard is8 KiB and each
local union bitmap256 bytes; sampled packet tile is at most16 KiB, index tile
32 KiB, random buffer at most16 KiB. CUDA context/runtime, plans, good lists,
returned support and checkpoint/publication arrays are additional. All remain
inside prior caps; no claimed physical peak until the new run is reviewed.

## Validation kept integrated

One allocation with two arms: normal genuine s4/s6 and smallest s4 CUDA
memcheck/leak. No new isolated kernel jobs. Existing CPU polynomial/sparse routing
regressions and focused sampling/orbit ownership tests protect shared helper
changes. Native-only preflight verifies actual good counts, sample counts and
closed-orbit loads; both instances require one sample with seed109.

Every canonical artifact must match both the independent native implementation
and the exact hashes retained by78898. Prefix commits/reloads and full P*G identity
are still checked. Existing transpose/Taylor reviewers cover every discovery
transform, independently of later root/Newton transforms. New sample/source/index
packets require complete descriptors and paired payload proofs; index payloads
are independently reconstructed from saved support. Final support/count/sample
and entropy digest must match the native oracle. Focused tests reject random-stream
and global-stop mismatches and guard the same two-arm launch configuration.
The new full pipeline only sanitizes the smallest case, not every existing kernel
and size again. The old coordinator mode remains available without a repeat run.

## Resources, measurement and reproduction

Same admitted **two HPC nodes / two H200s / four CPU cores / 8 GiB requested RAM**,
one GPU/two cores/4 GiB per node, **128 MiB owned GPU/card**, **1 GiB RSS/rank**,
no pinned buffers, **128 MiB cache / 2 GiB scratch / 32 MiB per file**,
**720-second controller / 15-minute scheduler**. Actual discovery/factor/executor
engines are sequential with64 MiB per-engine caps. Exclude gnode03; otherwise
eligible nodes allowed. Existing per-node executable/library/wrapper/cwd/TMPDIR
preflight, UCX controls, progress/resource logs and profile retention remain.
No resource increase.

78898 controller took129.545 seconds, normal arm3.925 seconds and smallest
sanitizer3.875 seconds including all diagnostic work. New cases retain the same
N/K and downstream operation counts. One extra resident discovery transform/sample
and bounded sample/root communication should fit **3–6 minutes estimated** for the
controller, excluding queue; no measured new runtime yet. Keep240-second launcher /
220-second test cutoff per arm. Prior proof payload bound<14 MiB/rank plus one
small discovery sample/instance remains below20 MiB projected, inside32 MiB/file.
Cache remains projected below32 MiB. Limits and deadline headroom are unchanged.

Timings include the native oracle, profiling/logging, support archive comparison,
prefix fsync/reload and final publication checks. They are integration measurements,
not unprofiled preprocessing latency or fullsize speedup estimates. Save all samples,
completed prefixes and artifacts; do not monitor after submission.

```sh
python3 -B experiments/distributed-discovery-pilot/test_review.py
python3 -B experiments/distributed-discovery-pilot/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-discovery-pilot/slurm-%j.log \
  experiments/distributed-discovery-pilot/run.sbatch discovery-01
# After the user's completion ping:
python3 -B experiments/distributed-discovery-pilot/analyze.py discovery-01
```

Raw results/logs/profiles: `runs/discovery-01/`. Reusable source/binaries and
coefficient/checkpoint artifacts: `.cache/liftrs/distributed-discovery-pilot/discovery-01/`
(outside Git, manifest hashes retained). Previous caches/profiles remain.
