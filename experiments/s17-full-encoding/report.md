# First full s17 CPU and resident GPU encoding

Status: **CPU76645 and full four-H200 GPU retry76706 passed, with identical
canonical and packed output hashes**. Earlier GPU76646 failed its host-memory
guard and is retained as a historical first attempt. User requests both
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

## Original first-attempt resource admission

User authorization: "let's launch encoding jobs for both cpu and gpu" following
presentation of the resident route and verified prerequisites. The chosen
per-job ceilings are below the previously reviewed full-run host/device/storage/
time upper bounds; these are new client runs, not reuse of preprocessing memory
measurements as proven client capacity. The table below records the original first attempts; the current GPU retry uses
the larger budget in the SCALING-098 section. Numeric plans and commands are recorded
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

## First submissions

[Exact jobs/budgets/commands](submission.json): **CPU76645**, **GPU76646**,
source frozen to `ed0c026` for both. sbatch accepted both independent submissions;
no scheduler/process/log/result checks afterward. Results:
`runs/first-01/cpu/`, `runs/first-01/gpu/`; Slurm logs `cpu-76645.log`,
`gpu-76646.log`. Outputs/profiles/binaries:
`.cache/liftrs/s17-full-encoding/first-01/{cpu,gpu}/`. Queue/runtime unknown;
48h is the scheduler limit, not predicted duration. Await user's completion ping.

## Completed first attempts (SCALING-093)

[CPU analysis](runs/first-01/cpu/analysis.json),
[GPU forensic analysis](runs/first-01/gpu/failure-analysis.json),
[combined measured summary](runs/first-01/summary.json),
[final accounting](runs/first-01/scheduler-completion.txt).
Sources frozened0c026. CPU seals554source/11evidence/37cache reconcile;
GPU failure seals554source/16evidence/34cache reconcile. The latter is not a
successful encoding or a CPU/GPU output match. Normal successful analyzer still
rejects failed records; explicit `--forensic` retains failure status and verifies
only saved seals, not mathematical success.

### CPU76645: full s17 mathematical encoding passed

On gnode02,32CPU workers,768GiB host requested,noGPU. SchedulerCOMPLETED0:0,
elapsed2h22min49s. Whole controller8568.295s; main pipeline8521.244s includes
process startup/cleanup and sampling boundaries, while individual phase clocks
exclude those. Exactly one complete encoding, no warmup/repeat; profiles/logging
enabled throughout. This is a first full correctness run, not an unprofiled
latency baseline. It preserves the CPU path without CUDA.

| CPU phase / resource | Measurement |
|---|---|
| Full142.441GiB code load/hash/validation |29min12.498s |
| Database generation |55.145s |
| Native plan/workspace setup |0.561s (lazy host page touching affects this boundary) |
| **Complete encoding** |**1h45min11.087s** |
| Full3,876,566,203 systematic checks |26.897s |
|16independent scalar curve checks |4.249s |
| All-value validation/canonical64GiB hash |58.425s |
| Tight34GiB pack/write/hash/fsync |2min22.346s |
| Full packed disk hash +1027decoded positions |1min50.054s |
| **Peak sampled process RSS** |**427.845GiB** |
| Periodic cgroup sample maximum |599.097GiB, including charged file cache |
| Cgroup recorded peak in periodic observations |602.416GiB, includes cache/other charges |
| Highest periodic file-cache observation |176.743GiB; not process RSS |

Native encoding stages: Taylor1095.009s, columns/transposes2099.162s,
rows2527.510s, buffers327.203s, pointwise171.857s, assembly90.345s. Those are
nested inside the complete call and not additional to its elapsed time. No full
s17 performance comparison with GPU is available. Cold-cache or other-node load
latency is not established by one source-file read.

Every symbol/curve/range/output check passes; output has36507222016bytes exactly,
17bits per canonical value without per-value type padding. It is resident
encoding followed by packing; this does not validate the distinct streamed-code
packed-file mode at s17. Canonical uint32 array uses64GiB logical bytes and128GiB
owning backing in the wide CPU encoder; those in-memory allocations remain part
of measured RSS. Packed file uses34GiB only.

Canonical SHA256:
`4fc38b3e28ade50b4de5923d0ee3685ccfa1a1dec51a29c8d51126b3a323554a`.
Packed SHA256:
`90bbdca5a34ca09b13bf44d274e7a9ec93b859d94b6778061ce62c2ebadb42f1`.
Retained word/profiles/executable are at
`.cache/liftrs/s17-full-encoding/first-01/cpu/`.
The packed producer readback already rehashes the full file; review does not
silently reread34GiB, nor repeat expensive encoding.

### GPU76646: successful setup, failed host-memory admission

On gnode04,fourH200/32CPU/384GiB requested host. SchedulerFAILED1:0 after36min56s;
controller2207.499s. Source loading1929.737s, database45.580s and encoder
setup69.057s complete. Full-sized resident GPU allocation reports**398.463GiB
aggregate owned device memory**, including metadata, under4×112GiB caps. This
confirms setup/allocation on four cards; it is not a measured GPU memory peak
through all runtime phases or proof of complete full-domain encoding correctness.

The call allocates its64GiB host output and then starts native message validation.
The controller triggers the **360GiB cgroup memory guard** before a full transform
is logged. Sampled process RSS reaches**221.577GiB**, below its320GiB RSS guard.
Periodic cgroup samples show about141–144GiB charged file cache in addition to
anonymous memory; a142.441GiB source read can therefore push total charged memory
over360GiB as the output is materialized. Last logged heartbeat is351.609GiB;
the actual guard-trigger sample exceeds360GiB but was not persisted separately.
Do not label351.609GiB as the final cgroup peak or claim a Slurm OOM: the saved
cause is our controller's cgroup guard, not an encoder/field validation failure.

[Incomplete persisted phases](runs/first-01/gpu/incomplete-phases.json),
[raw phase/native log](runs/first-01/gpu/encoding-pipeline.log),
[resource observations](runs/first-01/gpu/progress.jsonl), [Slurm log](gpu-76646.log).
TERM produces a flushed profile/termination event; after bounded TERM/KILL waits,
worker is unreaped and Slurm ultimately closes the failed allocation. The failed
attempt has no completed encoding phase, verified canonical word or packed result.
No CUDA OOM,reset or arithmetic fault is reported in these logs, but successful
full encoding is not established by absence of such a message. All failure
artifacts/profiles/plugin/source/code remain retained.

The original384GiBhost/360GiBcgroup plan allowed insufficient file-cache margin
for this workload; this is an experiment-admission mistake. No slower memory
mode is needed from current evidence. A concrete retry recommendation is
**512GiB requested host /480GiB cgroup guard**, retaining320GiB RSS/288GiB Go soft,
4H200112GiB/device,32cores and all correctness/output checks. That gives room
around~221GiB process +~143GiB source cache +34GiB output cache, with additional
headroom. This is a proposal, not a guarantee of GPU runtime/health or approval
to relax other resource guards. Full GPU encoding would still have to pass all
mathematical/output checks and match the now-known CPU digests. No replacement
job is submitted during this review; no CPU rerun is necessary.

Reproduce reviews (small sealed evidence only):

```sh
python3 -B experiments/s17-full-encoding/analyze.py first-01 cpu
python3 -B experiments/s17-full-encoding/analyze.py first-01 gpu --forensic
python3 -B experiments/s17-full-encoding/test_runner.py
```

Source/document checks pass; raw sealed logs retain original whitespace. The
reference now labels CPU measured/passed and GPU attempted/failed, preserving
s18–20 exploratory rows and mode-specific unvalidated status.

## GPU retry with conservative host headroom (SCALING-098)

The user authorizes the retry and asks that resource guards leave adequate room.
Use a fresh **gpu-retry-02** run and the **same encoding arithmetic/probe source
`ed0c026`** as CPU76645/GPU76646; the new spectrum prototype is not part of this
retry. The updated controller and its exact SHA256 are retained separately as
`controller.py.txt`. Source arithmetic hashes and controller evidence therefore
have explicit, separate provenance. No CPU rerun or preprocessing is needed.

| Resource | Retry budget / justification |
|---|---|
| Host RAM request | **640 GiB** on one HPC node (about907 GiB schedulable capacity) |
| Total cgroup guard | **608 GiB**, including charged filesystem cache |
| Process RSS guard / Go soft target | **320 GiB / 288 GiB**, unchanged; prior sampled process peak221.577 GiB |
| GPUs | **4 H200**, **112 GiB/card** owned cap; modeled maximum~104 GiB/card; contexts/driver use the remaining physical capacity |
| Active CPU cores | **32**, unchanged |
| New retained / scratch caps | **64 GiB / 16 GiB**, unchanged; tight output34 GiB, no duplicate code file |
| Worker / controller / Slurm deadlines | **46 h / 47 h / 48 h**, unchanged |

The earlier512 GiB host/480 GiB guard proposal is superseded for this retry.
Even if process RSS reaches its full320 GiB guard while all142.441 GiB source and
34 GiB output remain charged as file cache, the608 GiB total guard leaves
**111.559 GiB additional margin**. Observed process RAM was much lower. The32 GiB
gap between the cgroup guard and requested RAM allows controller/sampler and
scheduler headroom. This is an admission allowance, not a guaranteed peak or
runtime. Guard failures now flush the exact cgroup/RSS/storage observation before
termination, fixing the first attempt's missing trigger sample. Hardware/peer
preflight, phase/pass progress, profiles and failure retention remain active.
Shared-storage free space was12900.7 GiB and login-node `/tmp`401.9 GiB at recipe
review; these are observations, not quota or compute-node reservations. The job
rechecks assigned-node scratch and shared-storage headroom before work.

The controller accepts `--reference-run first-01`, preserves the reviewed CPU
reference in its evidence and requires both complete canonical and packed hashes
to match CPU76645 before reporting success. The analyzer repeats this cross-run
comparison, in addition to all systematic/curve/range/hash/packed readback checks.

```sh
python3 -B experiments/s17-full-encoding/test_runner.py
python3 -B experiments/s17-full-encoding/run.py gpu --check
module load slurm/slurm/25.05
sbatch --output=experiments/s17-full-encoding/gpu-%j.log   experiments/s17-full-encoding/gpu.sbatch gpu-retry-02   --revision ed0c026 --reference-run first-01
# After the user's completion ping only:
python3 -B experiments/s17-full-encoding/analyze.py gpu-retry-02 gpu
```

Four no-allocation policy/completion tests pass, including pessimistic host/cache
margin and rejection of mismatched reference hashes; controller manifest checks,
Python syntax, launcher shell syntax and whitespace checks pass. No full output
or large preprocessing data was rehashed for preparation. Runtime remains unknown
until this retry completes. Results will be `runs/gpu-retry-02/gpu/`, retained
output/profiles/binaries `.cache/liftrs/s17-full-encoding/gpu-retry-02/gpu/`.

### Retry submission

**GPU job76706 submitted**, fresh `gpu-retry-02`; [exact receipt](submission-gpu-retry-02.json).
Prepared controller9855919; frozen encoding/probeed0c026. Results
`runs/gpu-retry-02/gpu/`, scheduler log`gpu-76706.log`, retained data
`.cache/liftrs/s17-full-encoding/gpu-retry-02/gpu/`. No dependency, fixed node or
CPU rerun; four same-node GPUs with gnode03 excluded. No status polling, log
following or result analysis after submission. Await the user's completion ping.

## Completed full GPU retry (SCALING-102)

**76706 passed**, HPCgnode04,4H200 NVL,32 CPU cores,640 GiB requested host RAM,
608 GiB total-memory guard,320 GiB RSS/288 GiB Go,112 GiB/card owned cap.
Slurm **COMPLETED0:0**, **22 min11 s**. Frozen arithmetic/probe source
`ed0c0260c2eeea183658a788561517c96cd4a2de`, updated controller verified bytewise
against commit`9855919` and its recorded SHA256 sidecar. [Producer results](runs/gpu-retry-02/gpu/results.json),
[analysis](runs/gpu-retry-02/gpu/analysis.json), [review summary](runs/gpu-retry-02/gpu/review.json),
[accounting](runs/gpu-retry-02/gpu/scheduler-completion.txt), [scheduler log](gpu-76706.log).
**554 source /18 evidence /39 retained-cache identities** reconcile, with no
source overrides. The copied reviewed CPU reference equals the original76645
record, including full code identity. No expensive setup was regenerated and no
CPU rerun was needed. Actual toolchain/hardware/topology remain in environment
logs; GPU driver580.126.20, H200 NVL140.40 GiB/card, Go1.27.3, GCC14.2.0,CUDA12.9.

| Phase / boundary | GPU76706 measured runtime |
|---|---:|
| Load/hash verified142.441 GiB reusable code |318.948 s (5 min18.948 s) |
| Generate deterministic input, seed71 |36.369 s |
| GPU code/workspace setup |106.844 s (1 min46.844 s) |
| **Complete encoding call** |**434.930 s (7 min14.930 s)** |
| Check all3,876,566,203 systematic positions |28.338 s |
| Independent scalar curve checks,16 curves |4.629 s |
| All-value range and64 GiB canonical hash |66.937 s |
| Tight34 GiB packed write/hash/fsync |130.606 s (2 min10.606 s) |
| Full packed disk hash and1027 decoded positions |104.256 s (1 min44.256 s) |
| Worker pipeline |1232.023 s (20 min32.023 s) |
| Main test command including profile/cleanup overhead |1242.219 s |
| Controller |1330.502 s |
| Complete Slurm job, including build/smoke/provenance |1331 s (22 min11 s) |

All phases are profiled/logged. The synchronized native stage subtotal is
359.866 s inside the434.930 s encoding call; the remaining time is not classified
as additional transforms by this review. Keep the complete-call time as the
encoding figure. This is one successful encoding, not a repeated unprofiled
latency baseline.

### Memory and budget outcome

- **Peak sampled process-group RSS:221.591 GiB** (237931851776 bytes).
- **Owned device allocations:398.463 GiB aggregate**, observed at setup and
  encoding return (427846627660 bytes both times). The per-card maximum104.004 GiB
  is an array model, not a measured peak;112 GiB/card is the explicit cap.
  Whole-device peak, context/driver and transient overhead were not sampled.
- **Cgroup reported peak:261.071 GiB**, including charged file cache. Maximum
  periodic current sample259.139 GiB; largest periodic file charge37.384 GiB.
  Keep these separate from RSS and GPU-owned memory.
- No memory/storage/deadline guard triggered. The640/608 GiB admission leaves
  substantial margin. Lower charge than the failed first attempt reflects
  differing cache/ownership state, not a measured reduction in algorithm arrays.
  Do not tighten budgets automatically based on this run.
- The34 GiB packed word, profiles and reusable binary/plugin stay at
  `.cache/liftrs/s17-full-encoding/gpu-retry-02/gpu/`; the word is outside Git.
  Review checks producer full readback/size and sealed digests, without another
  hidden34 GiB reread. Original failed-attempt and CPU artifacts remain intact.

### Exact CPU/GPU agreement and observed ratio

Both CPU76645 and GPU76706 passed full correctness/output checks on the same
s17,m2,t5 code (K3876566203,N17179869184), verified preprocessing digest and seed71.
Complete canonical SHA256:
`4fc38b3e28ade50b4de5923d0ee3685ccfa1a1dec51a29c8d51126b3a323554a`.
Complete tightly packed34 GiB SHA256:
`90bbdca5a34ca09b13bf44d274e7a9ec93b859d94b6778061ce62c2ebadb42f1`.
The controller required equality before success, and the analyzer independently
checks the preserved CPU reference and every completed GPU phase.

CPU encoding6311.087 s versus GPU434.930 s gives **14.51× observed encoding-stage
ratio**. This compares single profiled/logged runs on different HPC nodes
(CPUgnode02, GPUgnode04). Code loading differs sharply (CPU1752.498 s versus
GPU318.948 s), with uncontrolled filesystem/cache state. Do not report a
controlled end-to-end speedup, linear multi-GPU scaling or general performance
bound. CPU/GPU memory figures describe different host/device placement and must
not be collapsed into a claim of lower total-system memory.

Full resident CPU and GPU s17 checkpoints are now demonstrated on HPC. Other s17
allocation/storage modes remain unvalidated, and fulls18 admission is separate.
Continue the s18 host-backed transform and split-product investigation; the
passed small export gate76710 does not remove those device/earlier-stage needs.
