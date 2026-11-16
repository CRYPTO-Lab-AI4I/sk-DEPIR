# Resident GPU client encoding integration

Status: **source implemented; bounded validation job76426 submitted, results pending**.
The reviewed CPU refactor comparison (job76055) does not block this integration.
Fulls17 preprocessing75832 is untouched and uninspected. No full GPU encoding run
or speedup, memory peak or latency result is claimed.

## Implementation and reproducibility

[API/ownership/design guide](../../docs/gpu-encoding.md),
[shared cyclic schedule](../../encoding/liftrs-cpp/cyclic_schedule.hpp),
[GPU client adapter](../../encoding/liftrs-cuda/resident_encoding.cuh),
[assembly kernels](../../encoding/liftrs-cuda/code_encoding_kernels.cuh),
[public code-data adapter](../../encoding/liftrs-shared/cuda_code.go),
[GPU checks/fixture generator](../../encoding/liftrs-shared/cuda_code_gpu_test.go).
The CPU/GPU ten-step schedule is shared; existing width-selected transforms,
field arithmetic, device slabs and peer leases are reused. Client callbacks only
supply views, movement, products and assembly. CPU defaults and the legacy s16
GPU modes are unchanged. Header extraction introduces no CPU arithmetic change.

[Runner](run.py), [client suite](suite.py), [Slurm launcher](run.sbatch),
[direct native checker](check.cu), [completion analyzer](analyze.py). Runner and
analyzer reuse the bounded complete-preprocessing workflow through optional suite/
review callbacks; default behavior is unchanged. This checks the shared executor's
server regressions as well as the new client, without duplicating controllers.
The source revision/hash of every copied source, commands and environment are
sealed in `runs/validation-01/results.json`. Results/logs/profiles remain beside
it; retained plugins/test/checker binaries and independent fixtures live in
`.cache/liftrs/gpu-code-encoding/validation-01/` outside Git. Exact submission
revision/budget will be recorded at launch.

```sh
python3 -B experiments/gpu-code-encoding/run.py --check
module load slurm/slurm/25.05 gcc/14.2.0 cuda12.9/toolkit/12.9.1
sbatch --output=experiments/gpu-code-encoding/slurm-%j.log \
  experiments/gpu-code-encoding/run.sbatch validation-01
# Only after the user's completion ping:
python3 -B experiments/gpu-code-encoding/analyze.py validation-01
```

Use a fresh name; preserve prior evidence. One eligible node,4H200,1GiB owned
cap/card (4GiB aggregate),4CPU,8GiB host request/6GiB sampled RSS/4GiB Go soft,
2GiB new retained cache and4GiB scratch; worker600s/controller900s/Slurm1200s.
These reuse the explicitly approved four-GPU small-test budget. No node pin,
exclude gnode03; no other job modification or GPU reset. Go1.27.3/GCC14.2/CUDA12.9,
CUDA architecture built for the actual node. Resource/cancellation/deadline guards,
peer preflight, profiles and persistent caches stay active. Submit then yield;
no agent monitoring before completion ping. No full resource increase is implied.

## Validation boundaries

The shared workflow repeats CUDA-free/native/race/ASan/vet, small complete GPU
preprocessing/restart/readback, peer preflight and fresh executable CUDA memcheck.
The additional client suite first generates genuine codes at(s,m,t)=(3,2,2),
(4,2,2),(4,2,3),(6,2,4),(8,2,5). Independent Go words are checked against the wide
CPU encoder before becoming references. Seed170087 fixes message values; the
existing fixture preprocessing uses its documented seed170058. Fixture generator
source and cache hashes preserve that provenance. Its internal binary format
is private to this checker, not another public code-data representation.

The Go client comparison covers1/2/4GPUs and both peer/local layouts:30 cases.
Check exact word equality with independent Go/CPU, zero/invalid/short inputs,
recovery, copied immutable constants, owning retained words after reuse/Close,
resource accounting, repeated calls without new device allocations, closed state,
and coexisting peer leases. Dense-prefix cases require2L>N and check the same
parameter-driven arena logic rather than a separate small/large implementation.
Small client construction/array admission also runs in CPU/CUDA-free builds.

The direct native checker runs60 cases: five genuine code fixtures, both
uint32/uint64 workspace specializations,1/2/4cards, peer/local. It checks output
canaries, overlap rejection, zero/invalid/retry, immutable copied constants,
metadata caps and reconfiguration rejection. CUDA memcheck repeats12 genuinely
small cases (s3/s4, both widths,4cards, both layouts), plus actual34-bit coordinate
checks, keeping sanitizer work bounded. Require zero CUDA errors and leaks.

A separate actuals17 kernel/peer-lifetime regression checks short transforms,
products, roots, high-bit values, lifetime and cap failures, using64MiB/card so
its negative enumeration test cannot allocate a full bitmap. Separately, true
GF(2^34) canonical-position kernels test indices on both sides of2^32 and high
positions/shard offsets against independent Go arithmetic. Forcing64-bit slots
on a small genuine code validates the complete wider storage schedule; it is
**not** a genuine fulls17 codeword or a truncated-domain substitute. Fulls17
encoding still needs verified preprocessing, complete capacity/budget admission,
full-domain correctness checks and actual measurements.

## Memory, communication and timing

For N=2^34,K3876566203,L=N/2, two work arrays plus resident P/spectrum and chunk
parity project398.457GiB device payload total,104.004GiB maximum/card with four
cards. Metadata/context/driver headroom is additional. Code P is not uniformly
spread because prefix slabs are shorter; per-card maximum is explicitly modeled.
The API returns64GiB owning canonical uint32 host output; retaining code/message/
output projects220.883GiB host arrays, plus runtime/staging and older words.
These are projections; the bounded job allocates only tiny fixtures.

The fast resident route copies constants once. It adds no transform or full
device output allocation, and does not introduce an implicit slower streamed
mode. Prefix promotion reshapes into dead scratch; canonical scatter writes
into the dead arena. Sharded routing/scatter and per-card parity synchronization
are real communication costs, not assumed free. No runtime estimate follows from
capacity. New streamed/compact/host-backed/packed sinks are outside this first
resident integration and remain later steps if justified.

Encoding uses synchronized host-wall stage clocks, explicitly labeled by
`TimingSource`, including launches/joins; legacy event timings keep their previous
boundaries. Go call timing includes allocating the owning host word. Constants
upload/configuration is separate setup. CPU profiles and logs are retained in
validation; validation latency is not an unprofiled benchmark or speedup baseline.

Common ten-step events identify the transform/pass; field checks report completed
blocks, logged setup copies use256MiB pieces for byte progress (ordinary calls
keep bulk copies), GPU transforms retain existing vector waves and assembly uses chunk
waves. Updates during long work throttle near30s. Controller resource heartbeats
are labeled separately with unknown remaining work. No per-element logging or
new full host transform array is introduced. Logging overhead is unquantified and
included in these validation timings. Native ABI errors drain device work before
releasing borrowed Go pointers.

## Local checks before submission

Pure-Go/CUDA-free resource/cap/unavailable tests pass. CUDA-only and combined
native/CUDA Go packages compile. Shared CPU module and wide encoding correctness
tests pass after the schedule extraction (~6s test time). The retained local
fixture generator compares genuine small codewords against independent Go and
CPU; its profiled run passes without GPU access. Shared controller tests and
no-allocation budget checks pass. [Local evidence](local-checks.json) and
[fixture test log](local-fixtures.log) record checks; binaries, fixture and CPU
profile remain under `.cache/liftrs/gpu-code-encoding/local-check/`.
Native CUDA compilation and execution/sanitizer checks run in the submitted
allocation and remain pending; source/Go-build success is not GPU validation.

## Submission

Job **76426**, source `bf2bffc`, accepted by sbatch under the unchanged small-test
budget ([receipt](submission.json)). Results/logs: `runs/validation-01/`,
`slurm-76426.log`; matching binaries/fixtures/profiles remain in
`.cache/liftrs/gpu-code-encoding/validation-01/`. No scheduler/log/process polling
or result inspection after submission; await the user's completion ping. Source
implementation and local CPU/build checks are distinct from pending GPU runtime
validation. Preprocessing75832 remains untouched/uninspected.
