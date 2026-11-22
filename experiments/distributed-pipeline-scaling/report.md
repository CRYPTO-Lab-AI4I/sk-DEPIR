# Larger genuine distributed preprocessing

Status: **passed81286**,scaling-01. Implementation revision **7395079**;parent9139e9e.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/). Results will be
`runs/scaling-01/`; reusable source, native references, binaries and artifacts
remain outside Git in `.cache/liftrs/distributed-pipeline-scaling/scaling-01/`.
Exact implementation/producing revisions are recorded at commit and submission.

## Scope and shared implementation

Measure one complete genuine **s=8, dimension=2, curve degree=5** cold run and
fresh four-process prefix8 restart: distributed discovery, GPU leaf roots and
hierarchical owned root merges, Newton reciprocal, full polynomial identity,
spectrum and canonical code/generator publication. This is **65,536 grid values**,
16 times the previous complete s6 domain. It is not field18 or fulls18.

Explicit `PIR_PIPELINE_CASES=8` and `PIR_PIPELINE_CURVE=5` parameterize the existing
[pilot](../../encoding/liftrs-shared/distributed_pipeline_gpu_test.go),
[bounded native inputs](../../encoding/liftrs-shared/distributed_reference_test.go)
and independent Python reviewers. Historical defaults remain s4/s6, curve degree2.
Cases stay bounded to s4–8 and curve degrees2–5; actual support/domain admission
and correctness checks still apply. A native preflight caught that curve degree2
at s7/s8 exceeds the existing K≤N/2 preparation contract. Those rejected shapes
are retained in readiness logs; no GPU job was submitted for them. Curve degree5
matches the target configuration and fits without weakening that invariant.

Use explicit owned-root execution and compact checkpoints as validated81190;
legacy padded default is unchanged. Native CPU export prepares independent
leaf/every-merge/final root files and complete code/generator/spectrum/support
references before distributed timing. GPU arithmetic consumes only discovered
local roots/shards. After each arm the controller compares exact native bytes,
all packet routes/fingerprints, discovery entropy/sample counts, completed-prefix
payload/hash/extent/identity, final artifacts and genuinely new restart processes.
Shared code/review helpers own all arithmetic and common checks; no second pipeline
implementation. Existing wide-field/sanitizer results remain applicable; this
suite does not repeat unchanged wide arms or a sanitizer-only arm.

## Preflight and resource admission

Native CPU preflight passed in **1.08 seconds**, with **K=7,381**, two samples,
leaf coefficient counts **[3,442,1,507,1,541,895]**, first-level domains8192/4096,
final root domain8192, final reciprocal prefix58,155. Complete native root/code
files are retained under `.cache/liftrs/distributed-pipeline-scaling/readiness/checked/native/`.
[Checked native log](readiness/native-export-checked.log), [CPU preflight](preflight.py),
[reference manifest](readiness/root-references.json). Native CPU reference setup
will be repeated/profiled in the allocation and counted separately from GPU arm
latency, inside total controller time. Historical default exports match **108,014 bytes** of pinned references, including
69,746 bytes of native code/support metadata and all narrow leaf/stage/final hashes.
The reproducible preflight caught and corrected a balanced-partition assumption
in the new exporter: genuine discovery uses stable orbit ownership. Checked native
files now match every historical stage byte;the initial mismatch is retained in
[readiness](readiness/initial-ownership-mismatch.log). No GPU allocation failed.

Same admitted HPC resources: **two nodes, one H200/node (two total), two CPU
cores/node (four total), 4 GiB requested host RAM/node (8 GiB total)**. Four MPI
workers share the cards two each. **64 MiB owned GPU/worker, 128 MiB/card**, **1 GiB
RSS guard/worker**, zero CUDA-pinned buffers; driver/context/MPI internal memory
is additional. **128 MiB retained cache, 2 GiB scratch, 32 MiB/file**, **720-second
controller /15-minute scheduler deadline**, eligible nodes excludinggnode03.
GPU/node count is an admission for this measurement, not a design limit.

Capacity projections (not measured peaks): three resident extension arrays need
**192 KiB/worker** for this 32-bit domain; tables/staging add memory. Budget below
**4 MiB owned/worker**, **512 MiB RSS/worker** based on the previous approximately
381 MiB process baseline plus small enlarged arrays. Source snapshot is below
8 MiB, complete cache projected below48 MiB; retained checkpoint payload for the
cold arm is **524,284 bytes versus4,456,448 padded bytes** (root plus16 Newton
prefixes). Other diagnostics/native copies/profiles/binaries are additional.
No useful earlier caches or profiles are deleted.

Measured81190 largest cold worker log was2,774,234 bytes for s4/s6 combined.
A conservative projection using fourfold square-root domain growth,16/12 prefix
ratio and1.5× extra margin is **22,193,872 bytes**, below32 MiB/file. This is an
extrapolation, not a proven bound. Known parameters/native geometry are checked
before allocation; retained file guards still apply. No increased resource cap.

Runtime estimate: **3–6 minutes excluding queue**, including build, native
reference preparation, shared CPU/race checks and both profiled GPU arms. For
context81190 took2min49.562s including four smaller/bounded arms; historical
65,536-value structured GPU product net median was0.591319s on two workers
([measurement](../resident-product-larger/report.md)), with different ownership
and excluded setup/oracle boundaries. Do not multiply that into a fullsize ETA.
Each arm keeps the existing220-second test/240-second launcher deadline, inside
the720-second controller deadline. Measured fulls18 cost remains unknown.

## Timing and memory–time policy

Timestamped/flushed stage, sample, root group, prefix, transform/pass/tile and
resource logs remain active. CPU profiles and completed arm results persist
before later steps. GPU-arm timing includes save/hash/sync/reload, setup,
communication, proof logging and publication; controller time additionally
includes native export, compilation, independent review and profile processing.
Isolated logging/codec overhead is not measured. Compare phases and total cost,
not kernel time alone. Compact mode reduces zero-padding I/O and auxiliary restore
memory; its fullsize zero-tail scan/small-read throughput cost remains unknown.
Padded stays the default when resources favor it; explicit experimental compact
selection does not establish a faster production default. Resident GPU arrays and
final encoded database packing are unchanged.

## Reproduction

HPC Go1.27.3, GCC14.2/OpenMPI4.1.8, CUDA12.9/H200. Run focused ordinary CPU/race,
reviewer and native/export compatibility checks, then verify the committed source
snapshot and required wrapper/manifests before submission. Every participating
node verifies shared reference/executable/library/wrapper paths and cwd before
GPU phases. No MPI/CUDA initialization is needed for native preflight export.

```sh
# Build the existing tagged test binary (no CUDA/MPI initialization):
CGO_ENABLED=1 CC=mpicc CXX=mpicxx go test -c -tags liftrs_cpp,liftrs_cuda,liftrs_mpi \
 -o .cache/liftrs/pipeline-scaling.test ./encoding/liftrs-shared
# Use a fresh destination; verifies pinned legacy code and every native root stage:
python3 -B experiments/distributed-pipeline-scaling/preflight.py \
 .cache/liftrs/pipeline-scaling.test .cache/liftrs/pipeline-scaling-preflight
python3 -B experiments/distributed-pipeline-scaling/test_review.py
python3 -B experiments/distributed-pipeline-scaling/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-pipeline-scaling/slurm-%j.log \
 experiments/distributed-pipeline-scaling/run.sbatch scaling-01
# After registered completion notice:
python3 -B experiments/distributed-pipeline-scaling/analyze.py scaling-01
```

Next: inspect larger measured phase/memory/I/O/checkpoint costs, finish scalable
fullsize discovery/ownership/working-set and artifact paths, then review fulls18
resources, storage, headroom and deadlines. Fulls18 remains unrun and unadmitted.


[Committed-source preflight](readiness/snapshot-checks.json) passed on7395079:
**827 files /3,803,974 bytes**, all required inputs/wrapper paths and ten reviewer
suites checked from the actual frozen source. Built that source,ran its native
export/admission and codec tests,and verified **108,014 historical native bytes**
plus genuine s8/curve5 support/code and every native leaf/merge/final reference.
[Build](readiness/frozen-build.log), [native regression](readiness/frozen-preflight.log),
[native tests](readiness/frozen-native-tests.log). GPU results passed81286 (see measured outcome below).


Submitted **81286** after readiness50f4867. [Receipt](submission.json),
[scheduler log](slurm-81286.log),results `runs/scaling-01/`. Same resources and
3–6 minute estimate excluding queue. Receipt-bound local completion notification
registered outside Git;no model polling. Analyze only after completion notice.


## Verified larger workload:81286

Independent reconciliation passed producing revision50f4867: **827 source files,
59 evidence files and274 cache files** verified. Both genuine s8/curve5 arms passed
all discovery/native root/packet/complete prefix/identity/spectrum/canonical-code
checks. Fresh MPI processes restored prefix8 and produced identical artifacts.

| Measurement | Cold | Fresh-process restart |
|---|---|---|
| Profiled arm wall time (launcher included) | 11.146650 seconds | 8.002511 seconds |
| Maximum-rank pipeline time | 9.631938 seconds | 6.729504 seconds |
| Peak host RSS per worker | 376.051 MiB | 379.316 MiB |
| Owned GPU memory per worker | 0.244 MiB | 0.244 MiB |
| Compact checkpoint payload | 524,284 bytes | 524,260 bytes |
| Equivalent padded checkpoint payload | 4,456,448 bytes | 3,932,160 bytes |
| Largest worker stdout log | 6,140,382 bytes | 3,226,233 bytes |

Total controller wall time **2 min22.513 s**,including reference export/build,
CPU/race checks,compilation,profiling and independent review. Native export took
**1.473112 seconds** in this allocation. Total retained cache **33.965 MiB**.
HPC gnode16/17,two H200s,four workers/four CPU cores,8 GiB requested host RAM;
existing caps unchanged. CUDA driver/context memory is additional to owned bytes.
Measured log sizes remained below the32 MiB/file cap and the22.2 MB projection.
No fullsize timing extrapolation is established by this one larger case.

Cold rank0 CPU profile spent95.14% of samples in cgocall,including69.69% cumulative
in local Taylor calls;cgocall also includes blocking/synchronization,so this is not
a pure kernel-time estimate. Save/reload/publication still allocate full resident
root/generator/identity/spectrum host shards. Next prototype bounded GPU spectrum
publication to avoid the full host spectrum allocation;retain resident default
and document unknown fullsize latency. Complete discovery/ownership and fullsize
resource/storage/headroom/deadline admission remain.

[Analysis](runs/scaling-01/analysis.json),[raw results](runs/scaling-01/results.json),
[cold profile](runs/scaling-01/cold-profile-0.log),[scheduler log](slurm-81286.log).
