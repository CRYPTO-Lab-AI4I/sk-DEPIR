# Primary CPU/CUDA online encoding validation

[Runner](run.py), [launcher](run.sbatch), [analyzer](analyze.py),
[shared API](../../encoding/liftrs-shared/cuda_encoder.go),
[CUDA implementation](../../encoding/liftrs-cuda/README.md).
Parent runtime revisionae27de5; all dirty source snapshots/hashes, exact revision,
compiler/driver/GPU/CPU/node/allocation, commands and artifact hashes are recorded
per attempt. [Validation-01 results](runs/validation-01/results.json) and
[verified analysis](runs/validation-01/analysis.json), job73243, now pass.
Full s16 GPU encoding remains a separate gate.

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/cuda-encoding/slurm-%j.log experiments/cuda-encoding/run.sbatch validation-01
# After the user's completion ping:
python3 -B experiments/cuda-encoding/analyze.py validation-01
```

## Backend contract and scope

Both CPU and optional CUDA are main client encoding paths. CPU clients need no
GPU, even when preprocessing is delegated elsewhere. The same trusted online
plan feeds either backend. No masks, protocol/server computation, application
layout, s17 arithmetic or production-security claims are added. GPU selection is
explicit; constructor/device errors do not silently choose CPU. CUDA remains an
optional vendor-specific backend with a practical generic Go/C++ CPU path.

Reusable matrix/lookup/orchestration code moves from gpu-matrix experiment to
encoding/liftrs-cuda. Historical experiment includes/imports forward to that
implementation; old source provenance/hashes remain recoverable. Shared ordered
basis/matrix construction becomes neutral Go code so CUDA-only builds do not
link the C++ CPU binding. Legacy kernels and new online ABI are separate optional
capabilities; missing online symbols fail clearly.

CUDA copies code constants once, serializes mutable device work, and performs
three prefix and two full transforms, point products and canonical assembly
without round-trip intermediate arrays. Three full-domain workspace arrays are
reused, including parity/error scratch after the final inverse. Native/input
validation and primitive-generator checks precede unsafe table/scatter access.
Repeated calls clear padding and flags. Host output is returned only after
base-field range checking. No Go pointer is retained and Close waits for Encode.
The high-level GPU object retains only scalar shape and device backend, so host
code arrays can be released independently or kept for CPU use.

## Verification and methodology

Full CGO-disabled Go, CPU-only, CUDA-only and combined suites, optional-tag with
CGO disabled, focused race/vet and historical matrix regression run before full
workflow comparisons. CPU runtime selection remains available throughout. Small
unit tests exercise admission/byte accounting and lifecycle without a GPU.

[Direct GPU tests](../../encoding/liftrs-shared/cuda_online_test.go) compare full
words to an independent Go path at s2,3,4,6,8 (m2, respective t1,2,2,4,5). Seed2501
builds code; seed2502 builds input. Repeat nonzero/zero/nonzero inputs, reject wrong
length/out-of-field messages, then rerun valid data to detect stale/error state.
Check input immutability, explicit ownership, double close/use-after-close and
old bounded-library online rejection. Keep small prepared-code files in the
private durable cache. Run these tests again under CUDA memcheck/leak checking;
report any missing tool/failure, not an assumed pass.

Historical field16 matrix fixtures compare complete forward/inverse/product
outputs through2^22 using the new library, retaining80 regression records and
source/cache checksums. Wide2^32/17-row-level descriptions have allocation-free
metadata checks; they are not a new full-domain execution claim.

[Workflow](../../encoding/liftrs-shared/cuda_workflow_test.go) reuses the same
load/encode/systematic/independent-curve/packed-readback harness for CPU and CUDA.
It loads trusted retained s10,m2,t5 code from gpu-products/products-02/cpu-0,
compares two16-worker native CPU encodings and two GPU encodings with setup,
transfers, reusable device state, Go profiles and per-stage clocks separate.
Message seed71 and complete canonical/packed hashes must match across backends
and repeats. Existing warm GPU kernel ratios are not applied to this sequence.

GPU event stages measure Taylor, columns, rows, layouts, pointwise, buffer and
assembly; transfer/native-total and whole Go clocks include different host work.
Do not sum overlapping boundaries as independent latency. Initial code upload is
in setup and reused. Go/native instruction profiling limitations remain explicit;
this experiment provides full output/stage/resource evidence, not production
security, a robust latency distribution or full s16 performance. GPU arithmetic
uses input-indexed tables and is not constant-time, as is the prototype's native
lookup arithmetic; no secret-key material is moved into code preprocessing.

## Resources, retention and next gate

One GPU/16 CPU cores,16 GiB host request/12 GiB sampled process-group RSS,
512 MiB explicit device cap,30-minute allocation/25-minute controller. CPU
references/small fixtures are bounded by s10; no s16 code arrays or actual
s16 encoding allocated. Driver/context overhead is separate from explicit arrays.
The controller monitors only to enforce job limits; the agent hands off and
waits under rule8. Completed profiles/caches remain intact; no old artifact is
deleted. New code fixtures/libraries/binary and s10 packed outputs live at
`.cache/liftrs/cuda-encoding/RUN`; logs/results/profiles and source snapshots are
retained in this experiment. Exact cache paths/checksums appear in the manifest.

After validation, separately admit full s16 GPU encoding against the retained
CPU canonical/packed digests. Array payload is approximately67.42 GiB (48 workspace
plus19.42 code), so the prior64 GiB preprocessing cap is insufficient. Proposed
full-run cap96 GiB device with explicit host/RSS/time/I/O bounds; no automatic
full job follows this validation. Keep CPU as a main path and disclose hardware
requirements rather than claiming commodity GPUs can hold s16 in memory.


Before submission, short CPU-side admission/lifecycle and wide-description tests
pass in CUDA-only and combined builds. A compile-only CUDA sm90 library is
retained at `.cache/liftrs/cuda-encoding/build-check/` with command/source/binary
hash provenance; it was not executed and is not a measured GPU result. The
submitted runner rebuilds for its actual allocated GPU and records that binary.
The shared harness refactor's initial combined compile caught stale KeepAlive
names; replacing those with the captured session/closer fixed it before submission.
Full runtime comparisons/memcheck remain pending.


## Completed modular validation

Job73243 passes all312 exact source identities and sealed output/cache hashes.
CPU-only/CGO-disabled/CUDA-only/combined builds, full suites, focused race/vet,
80 historical matrix records and full small-word Go/CUDA comparisons pass.
Tested s2/3/4/6/8 cover20 successful repeated/zero/error-recovery word matches,
invalid lengths/field values, host input immutability, close/release, and missing
online-ABI rejection from the verified prior bounded library. CUDA memcheck/leak
checking reports **0 errors and0 leaked bytes**. Pure metadata checks include
full-width2^32 counts/17 row levels without allocating full arrays.

Complete s10,m2,t5 workflow (N2^20) passes both repeats' all-systematic,16-curve,
canonical digest and1027 packed-readback checks. Packed hashes agree across CPU
and CUDA. Seconds, fixed order CPU then CUDA:

| Backend | Reusable session setup | Encoding1 | Encoding2 |
|---|---:|---:|---:|
| Native CPU16 workers | 0.002745 | 0.310053 | 0.308229 |
| CUDA | 0.850500 | 0.104040 | 0.103621 |

Steady encoding is about2.98x faster here, including message/final-word transfer.
Setup dominates this tiny case: setup+two encodings is about1.058 s GPU versus
0.621 s CPU. This reinforces keeping CPU as a peer main path and separating cold
setup from reuse. Do not project that ratio or break-even point to s16. The
measured H200 NVL/driver580.126.20, Xeon6710E, GCC14.2.0, CUDA12.9/Go1.27.3
are recorded in the logs; peak sampled process RSS0.982 GiB, below12 GiB.

Implementation/historical shims and APIs are retained with no runtime correction
needed from this validation. No full-size claim is made by this small run. The
next explicitly admitted [cached s16 GPU experiment](../cuda-s16-encoding/report.md)
compares against retained CPU hashes, separates init/upload/encoding/checking
and reuses one device session. CPU/Go/C++ remain usable without CUDA. No backend
selection depends on where code preprocessing was obtained.
