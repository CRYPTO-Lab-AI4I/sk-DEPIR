# GPU-backed spectrum publication

Status: **passed81365**,source-01. Implementation **f6353af** plus matched baseline **98fce69**;parent4b29540.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/). Results will be
`runs/source-01/`; immutable source, native references, binaries and reusable
artifacts stay outside Git in `.cache/liftrs/distributed-spectrum-publication/source-01/`.
Exact implementation/producing revision is recorded before submission.

## Remaining fullsize allocation

[81286 passed](../distributed-pipeline-scaling/report.md#verified-larger-workload81286)
a complete genuine s8/curve5 cold run and fresh-process prefix8 restart:
11.146650/8.002511 seconds per arm,376.051/379.316 MiB peak RSS per worker,
0.244 MiB owned GPU per worker,two HPC nodes/two H200s. However,
`rankGPUExecutor.Spectrum` still creates a complete extension-field host shard
before publication. At s18 that allocation alone is512 GiB aggregate,32 GiB/worker
with16 workers. Other host arrays and GPU workspace are additional.

The explicit experimental `PIR_PIPELINE_SPECTRUM=stream` separates the existing
GPU transform from host ownership. The [executor](../../encoding/liftrs-shared/distributed_executor_cuda.go)
prepares the spectrum in GPU slot0,then the shared
[typed publisher](../../encoding/liftrs-shared/distributed_publication.go) reads
bounded tiles instead of a complete host array. One downloaded tile feeds both
canonical output and the retained diagnostic shard. Root/generator publication
uses only host/MPI data and must not reset GPU slot0. A shared
[source helper](../../encoding/liftrs-shared/distributed_publication_gpu_test.go)
checks ordered/exact-once reads,canonical width/range,full completion and sync.

The existing resident default still returns the same complete host result.
Array publication retains its original full validation (including zero padding),
source slices and zero-copy sends. Both source forms share operation identities,
packet descriptors,global ordering,receiver bounds and canonical serialization.
Callback sources require a complete section (`Count=Domain`);the array path owns
partial sections and padding validation. Failures propagate to the existing
collective abort boundary;partial artifacts are not success.

## Memory–time trade-off

No additional FFT,network payload,device→host spectrum bytes or retained output
bytes are intended for the genuine optimized path. GPU workspace is unchanged.
The complete host spectrum is replaced by one4096-word tile per source (16 KiB
narrow/32 KiB wide),plus the existing rank0 receive tile and bounded serialization
buffers. These are allocation calculations,not measured total RSS savings.

| s | Avoided aggregate host spectrum allocation | Avoided per-worker allocation with16 workers |
|---|---|---|
| 16 | 16 GiB | 1 GiB |
| 17 | 128 GiB | 8 GiB |
| 18 | 512 GiB | 32 GiB |
| 19 | 2 TiB | 128 GiB (exploratory) |
| 20 | 8 TiB | 512 GiB (exploratory) |

Sixteen workers are an illustrative division,not new implementation support,
physical allocation or experiment admission. Other host arrays,root imbalance,
complete discovery,device/workspace memory,storage and runtime still need fullsize
review. This prototype does not establish fulls18 feasibility.

Fullsize latency is **unknown**. Interleaving GPU reads,file writes and network
publication may change overlap/throughput;tile range checks and bounded encoding
also have CPU cost. Diagnostic save and canonical publication now occur together,
so their phase boundaries differ from the resident baseline. Retain resident as
default when it fits;source mode remains an explicit experimental command-line
environment selection. No slower production default is adopted. Compare complete
arm/controller latency and bytes,not just the transform phase. No old cache is deleted.

## Validation and resources

Reuse the [larger complete suite](../distributed-pipeline-scaling/run.py) and
its independently generated native code/support/generator/spectrum and every
root leaf/merge/final reference. Two source-mode genuine s8/curve5 arms:cold plus fresh-process
prefix8 restart. Require identical canonical output,all MPI route/proof checks,
completed-prefix/full identity verification,no resident actual host spectrum,
source tile≤4096words,and exactly local-word-count downloaded once per arm.

A third genuine cold arm retains the resident default in the same allocation.
It repeats exact native/artifact checks and requires the full resident spectrum
allocation. This verifies the extracted common transform's old behavior and gives
one matched source/resident latency sample. Both use identical code,inputs,proof
logging and compact checkpoint layout. Arm order is fixed,with no statistical
speedup claim;cache/node/ordering effects remain. This short extra arm is needed
to check the default and make the memory–time trade-off reviewable.

One bounded field18 arm reuses the existing wide tail with dimensions5 and12,
actual GPU-backed spectrum reads,independent native FFT/reciprocal/convolution,
and canonical8-byte output. Small field18 spectra can lie entirely in the base
field;the retained81190 fixture confirmed no high extension bits in those actual
spectra. Therefore its existing deliberate high-bit raw fixture is uploaded to
GPU slot0 **after spectrum publication**,then read through the same source and
compared byte-for-byte. This validation-only fixture adds one bounded upload and
read;it is not extra work in the genuine optimized path. Expected and diagnostic
native arrays remain intentionally allocated for that small wide oracle. The new
wide source claim does not hide those arrays from measured RSS.

Ordinary CPU/race tests cover both widths,nonzero high bits,uneven/empty ownership,
exact local offsets/read counts,source/field/writer errors and plan admission.
Historical array tests remain. Eleven Python reviewer suites and replay of81286's
saved cold/restart evidence pass. Actual frozen source/native inputs must pass
before allocation. No CUDA arithmetic changed;no sanitizer-only arm is repeated.
Prior small clean CUDA checks are retained,not claimed for this new composition.

Same admitted HPC allocation:**two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**,four MPI workers/two per card;**64 MiB owned GPU/worker=128 MiB/card**,
**1 GiB RSS guard/worker**,zero CUDA-pinned;**128 MiB retained cache/2 GiB scratch/
32 MiB per file**,**720-second controller/15-minute Slurm deadline**,eligible nodes
excludinggnode03. Driver/context/MPI internals are additional to owned bytes.
No increased budgets or fixed GPU-count design ceiling.

Estimate **3–6 minutes excluding queue**,based on81286's2min22.513s controller
and81190's2min49.562s for four arms. Cold/restart workload is unchanged except its
output source;the matched resident arm previously took11.147s and the bounded wide arm6.841s including setup
and verification. Fullsize latency remains unknown. Cache projected below48 MiB;
worker logs below12 MiB/file based on81286's6.14 MB cold log and81190's7.77 MB wide
log. Native/reference cache plus existing binaries/diagnostics remain well below
128 MiB. Changed source output adds only bounded high-fixture diagnostic bytes
(up to32 KiB aggregate at dimension12). Existing guards and per-arm220-second
test/240-second launcher deadlines retain headroom.

Timestamped/flushed sample,root group,prefix,pass/tile and resource observations
remain active. CPU profiles and completed-arm evidence persist before later work.
Arm time includes setup/transforms,communication,hash/sync/checkpoint/reload,
diagnostic/canonical publication and proof logging. Controller adds native export,
builds,shared CPU/race checks,independent reviewers and profile processing. Isolated
logging overhead and fullsize pipeline throughput remain unmeasured.

## Reproduction

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200. Read-only/shared-path preflight runs
on every worker before GPU phases. Frozen-source checks cover the publisher,source
helper,wide tail,wrapper and required native manifests. Native export initializes
neither CUDA nor MPI;all references are retained with hashes.

```sh
python3 -B experiments/distributed-spectrum-publication/test_review.py
python3 -B experiments/distributed-spectrum-publication/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-spectrum-publication/slurm-%j.log \
 experiments/distributed-spectrum-publication/run.sbatch source-01
# After registered completion notice:
python3 -B experiments/distributed-spectrum-publication/analyze.py source-01
```

Next:reconcile exact source/cache/evidence and both widths,update maintained
scaling reference with measured latency/memory/output bytes,then finish remaining
fullsize discovery/ownership/root/reciprocal working-set admission and production
artifact/checkpoint paths before separate fulls18 resource/storage/deadline agreement.


[Committed-source preflight](readiness/snapshot-checks.json) passed on98fce69:
**831 build inputs /3,823,058 bytes**,all required publisher/source/wide/
wrapper/manifest paths and11reviewer suites verified from frozen source. The
existing tagged native binary was compiled from **379 byte-identical code inputs**;
its hash and source equality are recorded. Frozen helper execution verified
108014 historical native bytes and actual genuine s8/curve5 inputs. A second
potentially>30s local rebuild was omitted;fresh frozen-source compilation runs
inside the admitted allocation before GPU phases. [Native preflight](readiness/snapshot-native.log),
[compiled native tests](readiness/native-tests.log),[CPU](readiness/cpu.log),
[race](readiness/race.log). No local CUDA/MPI initialization.


Submitted **81365** after readinessf1dcb47. [Receipt](submission.json),
[scheduler log](slurm-81365.log),results `runs/source-01/`. Four arms:source cold,
fresh-process source restart,matched resident cold,and bounded field18 GPU source/
high-bit regression. Same two-node/two-H200 caps;estimate3–6min excluding queue.
Local completion notification registered outside Git;no model polling.


## Verified result:81365

Independent reconciliation passed **831 source/94 evidence/496 retained cache
files**,producing revisionf1dcb47. Source cold/fresh restart,matched resident cold
and bounded field18 source/high-bit arms all passed exact native roots/prefixes/
identity/spectrum/canonical artifacts,packet proofs and source allocation/read counts.

| Arm | Profiled wall time | Peak RSS per worker | Owned GPU per worker |
|---|---|---|---|
| s8 source cold | 10.713813 seconds | 379.027 MiB | 0.244 MiB |
| s8 source fresh-prefix8 restart | 7.749624 seconds | 379.039 MiB | 0.244 MiB |
| s8 resident cold | 10.346982 seconds | 379.180 MiB | 0.244 MiB |
| Bounded field18 source/high-bit tail | 5.793278 seconds | 380.199 MiB | 3.125 MiB |

HPC gnode16/17,two H200s,four workers/four CPU cores,8 GiB requested host RAM;
existing caps unchanged. CUDA context/driver memory is additional to owned bytes.
Controller wall time **2 min48.928 s**;complete retained cache **37.413 MiB**.
Source mode uses **0 resident actual spectrum words**,one≤4096-word source tile,
and exactly16,384downloaded words/worker in each genuine arm. Resident uses
16,384host spectrum words. The source tile adds a bounded source allocation;
the existing coordinator receive tile and serialization buffers remain. Small
RSS differences do not measure the fullsize array saving. Wide actual spectra
and deliberately high-bit GPU source data match exact native/original bytes.

In this fixed-order,single-allocation sample,source cold took **0.366831 seconds
more (3.545%)** than resident. This is not a statistically established penalty or
a fullsize estimate. Both used the same compact checkpoint/root/code parameters;
all output and checkpoint bytes match. Cold compact checkpoint payload524284bytes
versus4456448padded;fresh restart524260versus3932160padded. Resident default stays
unchanged;stream mode remains explicit when memory requires it or deliberately
selected for measurement.

[Analysis](runs/source-01/analysis.json),[raw results](runs/source-01/results.json),
[scheduler log](slurm-81365.log). Next remove the full host discovery-result copy
per sample with optional tile consumption into the existing local support union.
This targets another phase's8N/r allocation;it must not be added to spectrum
savings as simultaneous peak-RSS evidence. Fulls18 remains unrun/unadmitted.
