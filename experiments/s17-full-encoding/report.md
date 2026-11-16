# First full s17 CPU and resident GPU encoding

Status: **prepared; full CPU/GPU encoding results pending**. User requests both
jobs after [full preprocessing75832](../gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088)
and [bounded GPU integration76426](../gpu-code-encoding/report.md#completed-validation-scaling-089)
pass. This measures standalone mathematical encoding, not application masking,
PIR upload/query performance or production security.

## Reproduce and provenance

[Shared runner](run.py), [CPU Slurm launcher](cpu.sbatch),
[GPU Slurm launcher](gpu.sbatch), [shared probe](probe.go.txt),
[admission/completion tests](test_runner.py), [completion analyzer](analyze.py).
The probe is copied unchanged into `encoding/liftrs-shared` in isolated job
sources. It calls existing CPU/GPU encoding APIs, not a new encoding algorithm.
The runner archives the **explicit committed revision** supplied at submission,
so later working-tree changes cannot alter the native code, tests or fixtures.
Source identities, probe hash, command arguments, host/compiler/GPU inventory,
profiles/logs and resource observations remain sealed in results.

```sh
python3 -B experiments/s17-full-encoding/run.py cpu --check
python3 -B experiments/s17-full-encoding/run.py gpu --check
python3 -B experiments/s17-full-encoding/test_runner.py
module load slurm/slurm/25.05
revision=$(git rev-parse HEAD)
sbatch --output=experiments/s17-full-encoding/cpu-%j.log \
  experiments/s17-full-encoding/cpu.sbatch first-01 --revision "$revision"
sbatch --output=experiments/s17-full-encoding/gpu-%j.log \
  experiments/s17-full-encoding/gpu.sbatch first-01 --revision "$revision"
# Only after the user's completion ping:
python3 -B experiments/s17-full-encoding/analyze.py first-01 cpu
python3 -B experiments/s17-full-encoding/analyze.py first-01 gpu
```

Jobs are independent, without a dependency or fixed node. Exclude gnode03 while
its reset fault remains unresolved; no attempt to reset hardware. Submit then
yield with identifiers/result locations, without agent monitoring. A fresh name
is mandatory; no existing cache/output is overwritten. No new preprocessing.
All local paths below resolve under the repository unless absolute.

## Explicit resource admission

User authorization: "let's launch encoding jobs for both cpu and gpu" following
presentation of the resident route and verified prerequisites. The chosen
per-job ceilings are below the previously reviewed full-run host/device/storage/
time upper bounds; these are new client runs, not reuse of preprocessing memory
measurements as proven client capacity. Numeric plans and commands are recorded
at submission; no other memory modes or automatic fallbacks are selected.

| Resource | CPU resident wide | GPU resident |
|---|---|---|
| Active host cores |32 |32 |
| GPUs |None |4H200 on one node |
| Requested host RAM |768GiB |384GiB |
| Process-group RSS guard |704GiB |320GiB |
| Cgroup memory guard (includes charged file cache) |744GiB |360GiB |
| Go soft memory limit |640GiB |288GiB |
| Explicit device cap |None |112GiB/card,448GiB aggregate; contexts/driver additional |
| New retained output/profiles/logs |64GiB/job |64GiB/job |
| Scratch |16GiB/job |16GiB/job |
| Main worker /controller /scheduler cutoffs |46h /47h /48h |46h /47h /48h |

Combined new retained caps128GiB, below the previously agreed512GiB envelope;
existing235.324GiB preprocessing and its recovery/profiles/logs stay retained.
Each output is34GiB plus tiny fixtures/binaries/logs/profiles, not another copy
of the142.441GiB code. CPU requires~427GiB principal host arrays; GPU projects
~220.883GiB retained host arrays and398.457GiB device arrays,104.004GiB maximum
one card. Runtime/allocator/native/context/file-cache headroom is additional;
none is a measured full-client peak. srun binds CPU workers to the allocation.
The two host requests cannot share one907GiB node; Slurm places them independently.
No exclusive node is required. Unavailable capacity queues rather than forces a
slower mode. There is no new architecture or vendor requirement for CPU clients.

Runtime is unmeasured. The48h limit is a conservative cutoff, not a48h forecast.
The complete workload includes loading142.441GiB, encoding,3.877billion systematic
checks, full-word hashing,34GiB packing/write/fsync and disk rehash/readback.
Those extra checks are deliberate and timed separately; encode time alone cannot
predict whole-job wall time. If a cutoff interrupts encoding, partial transform
work cannot resume. A successful packed publication and phase records persist;
useful verification/output artifacts/profiles remain, without claiming encoder
checkpoint support.

## Input and preparation gates

(s,m,t)=(17,2,5),K3876566203,N17179869184. Source:
`.cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5/code/online.bin`, with trusted
checksum from `code/manifest.json`. Before launch, admission checks the completed
preprocessing analysis, exact manifest identity/counts and payload extent.
Every worker then independently streams/validates/hashes the whole code through
`LoadScalingCode`; a readable header alone is not success. No full resident
loading was previously demonstrated; this is part of this experiment's gate.
Explicit code payload cap161GiB, CPU workspace cap320GiB; actual allocation
requirements are checked before workspace allocation. CPU keeps128GiB output
backing for64GiB logical uint32 values; the GPU output owns64GiB. Old results
are not retained across encodes because exactly one large encode is performed.

Use identical Go math/rand seed71 field values, as in earlier retained encoding
experiments. The plaintext occupies4Kbytes (~14.441GiB slots), with17bits/value;
packed plaintext capacity is smaller but not materialized. The fixed seed is
reproducible experiment input, not a recommendation for cryptographic randomness.

## Methodology, profiles and timing

Every named phase retains a unique CPU profile, duration and result fields in
`phases.json`; writes are atomically replaced. Profile costs are included and
all durations are explicitly labeled profiled. This first full correctness run
is **not an unprofiled latency baseline** or controlled CPU/GPU speedup claim.
Go profiles may attribute C++/CUDA work to cgo; native stage clocks remain in the
encoding record. Both client backends keep their existing validated layouts.

| Phase | Boundary |
|---|---|
| `load-code` | Full file read/hash/value/shape validation and resident arrays |
| `database` | Seeded message generation |
| `encoder-setup` | CPU plans/workspace or GPU metadata/workspace/constant copies |
| `encoding` | Exactly one complete public call, including owning-output allocation; no warmup/repeat |
| `verify-systematic` | Check every message symbol at its canonical coordinate |
| `verify-curves` |16fixed randomly chosen degree5 curves, independent scalar field arithmetic |
| `verify-canonical` | Check every result is a base-field value and hash64GiB canonical uint32 bytes |
| `pack-output` |17bits/value, bounded buffers, streamed SHA256, write/flush/fsync/rename |
| `readback-packed` | Rehash all34GiB from disk and compare1027decoded positions |

Before the large call each backend runs a genuine s4,m2,t3 smoke pipeline,
including all phases, comparing complete output to independent Go. The fixture
checks the packing helper against the existing canonical packer at4/17bit slots,
non-byte-aligned sizes, negative systematic/readback corruption, and fixed-omega
lookup maps at actual34-bit positions. GPU additionally builds/runs the existing
peer-copy/transpose preflight before large allocations. Compile flags/toolchain:
Go1.27.3,GCC14.2,C++17; GPU CUDA12.9 native architecture plugin. CPU builds omit
CUDA entirely and request no GPU.

The systematic scan avoids an N-sized coordinate table. Independent scalar tower
arithmetic prepares a small four-bit fixed-omega table, then32 workers scan
separate2^18-symbol blocks, with independently exponentiated starts. That map
is for verification only; production CPU/GPU encoding remains unchanged.
Curve restrictions use carryless base multiplication independent of native
encoding kernels, matching the earlier16-curve acceptance method. A full-word
SHA256 is retained in each backend; after both pass, the analyzer also requires
identical CPU/GPU canonical and packed hashes. Agreement alone is not the only
check: full systematic recovery, independently evaluated curves, range and disk
readback are individually required. These are mathematical checks, not a proof
of complete PIR security or every possible curve restriction.

## Storage, logs and guards

Results/logs: `experiments/s17-full-encoding/runs/first-01/{cpu,gpu}/`.
Outputs/matching executables/profiles:
`.cache/liftrs/s17-full-encoding/first-01/{cpu,gpu}/`. Final `word-packed.bin` is
34GiB, uploaded neither to Git nor an external service. Canonical words remain
only in process memory; retaining another64GiB file would add I/O/storage without
improving this cross-backend digest check. Partial packed `.tmp` is preserved on
failure and cannot be mistaken for completed output. Existing preprocessing,
profiles and earlier encoding outputs are never removed. Output digest/extent
is sealed in phases/results; controller exit and reviewer do not silently rehash
34GiB again after the worker's full disk readback.

Timestamped phase start/completion/failure and periodic advancing byte/symbol/
curve counters are flushed. Native encoding logs retain the shared ten-step
names plus completed transform blocks/vectors and GPU assembly chunks, roughly
30s updates inside long work. Controller records30s RSS/cgroup/storage heartbeats
separately, explicitly unknown ETA. CPU profile is flushed on SIGTERM before
exit; controller uses bounded TERM/KILL shutdown and leaves Slurm cleanup for an
unreaped worker. No agent polling occurs after submission. Logging costs are
unquantified and included; no per-element logs or verification coordinate array.

## Pre-submission checks

[Local evidence](local-checks.json), [final fixture log](local-final-fixture.log),
[full small CPU pipeline log](local-final-pipeline.log). The unchanged probe builds
via an isolated Go overlay; retained small Go oracle/CPU pipeline passes all9
phases.17-bit packing/corruption checks and34-bit fixed-step comparisons pass.
Two Python admission/completion tests pass, including damaged byte counts/digests/
systematic totals. No-allocation manifest/budget checks and shell syntax pass.
Local binary/fixture/output/profiles remain under
`.cache/liftrs/s17-full-encoding/local-check/`; no full domain is allocated locally.
GPU small backend smoke runs inside its submitted allocation; previous complete
bounded integration already passed76426. No full performance/capacity claim follows
from these checks or from sbatch acceptance.
