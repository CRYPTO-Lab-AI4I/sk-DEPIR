# Resident GPU encoding from verified code data

`ScalingCodeData.WithCUDA(CUDACodeOptions)` is an experimental client encoder
through s17. It consumes the same immutable canonical code data as `WithWideCPP`;
preprocessing can be delegated elsewhere. CPU encoding and ordinary CUDA-free
builds remain available. Full s17 resident GPU encoding **passes76706** on4H200:7 min14.930 s,
221.591 GiB sampled host RSS,398.463 GiB aggregate owned device allocations.
Canonical and34 GiB packed output hashes match CPU76645. Bounded GPU
runtime checks pass76426 in the [integration experiment](../experiments/gpu-code-encoding/report.md).
The older `ScalingOnlinePlan.WithCUDA(CUDAOptions)` and its validated s≤16
memory modes retain their existing implementation/defaults.

## Use

Load code with `LoadScalingCode`, a trusted checksum, and explicit count/payload
caps as described in [code-data conventions](scaling-code-data.md). A checksum
alone does not certify an arbitrary supplied code; use verified preprocessing.
On a CUDA machine build the shared plugin including the resident executor:

```sh
nvcc -std=c++17 -O3 -lineinfo -arch=native -shared -Xcompiler=-fPIC,-pthread \
  encoding/liftrs-cuda/matrix.cu encoding/liftrs-cuda/preparation.cu \
  -o /path/to/libliftrs-cuda.so
go build -tags liftrs_cuda ./...
```

The Go CUDA binding needs cgo and dynamic loading, not CUDA headers/runtime in
ordinary CPU builds. `liftrs_cpp` independently selects CPU native execution.
The plugin must match the installed driver/GPU architecture; source deployment
requires the CUDA toolkit. Devices currently share one host with all-pairs peer
access; cross-node encoding is not implemented.

```go
// code is already loaded and verified. Inspect before admitting allocations:
resources, err := code.CUDAResources(4)
// Check err, admit resources plus metadata/context and aggregate host headroom.
encoder, err := code.WithCUDA(liftrs.CUDACodeOptions{
    LibraryPath: "/path/to/libliftrs-cuda.so",
    Devices: []int{0, 1, 2, 3},
    DeviceMemoryLimit: 112 << 30, // per GPU; used by full s17 validation76706
    VectorLayout: "local",       // explicit "peer" or "local"
})
// Check err; defer encoder.Close().
word, stats, err := encoder.EncodeWithStats(message)
```

Device count is explicitly1/2/4/8 visible ordinals. Initialization/cap/peer errors
are returned; there is no automatic CPU, streamed or disk fallback. The example
is API usage, not authorization for a large allocation. `CUDAResources` performs
no GPU allocation and reports arrays rather than peak memory. Construction copies
P/spectrum once to device without retaining their host pointers. Callers can keep
the source for CPU use or release it when no longer needed.

Each call returns an owning canonical `[]uint32`; subsequent calls, failures or
Close cannot alter previous results. Calls serialize, and Close waits for active
work. The input is borrowed only for the synchronous call and remains unchanged.
The new path currently provides canonical words; its direct packed-output,
streamed-code, compact and host-backed modes are not ported to s17. Tight packed
persistence remains a separate output step, not a claim of this integration.

## One sequence, separate execution

| Responsibility | Source |
|---|---|
| Code-data API, array model, optional binding | `liftrs-shared/cuda_code*.go` |
| Ten-step cyclic order, joining before reuse | `liftrs-cpp/cyclic_schedule.hpp` shared by CPU and GPU |
| CPU buffer adapters/products/output layouts | `liftrs-cpp/online.hpp` and existing CPU modules |
| GPU resident allocations, plans, peer leases and transforms | `liftrs-cuda/preparation.cu`, internal `Resident<Word>` executor |
| Client buffer views/transfers and schedule callbacks | `liftrs-cuda/resident_encoding.cuh` |
| Client promotion, canonical scatter and parity reduction | `liftrs-cuda/code_encoding_kernels.cuh` |
| Shared GPU field/vector recurrences and sharded addressing | `transform.cuh`, `preparation_kernels.cuh` |

Native paths are under `encoding/`. The existing `Preparation` Go binding name
and `pg_*` preprocessing symbols are retained for compatibility; client methods
use the same resident executor. New client symbols are optional when opening an
older preprocessing plugin, but required when configuring client encoding. The
server discovery/root/reciprocal mathematics stays separate. Independent Go
encoding remains a correctness reference rather than another shared backend.

Let L=nextpow2(2K−1), with minimum4, and A=max(N,2L). Buffer roles are:

| Stage | Arena A | Scratch N |
|---|---|---|
| Prefix product | Disjoint P/message views, then truncated H | Transpose scratch |
| Full product | Dead prefix storage becomes transpose scratch | Promoted/padded H, then Y |
| Canonical assembly | Dead arena receives uint32 canonical symbols | Read-only final coefficients |

Promotion reshapes shorter distributed prefix slabs into full slabs; every
stage joins before reusing memory. P[0:K] is stored in base-field uint32 slots;
P[K] is validated but unused by the truncated prefix product. Spectrum/workspaces
select uint32 through s16 and uint64 at s17, sharing the same kernels. Scatter
uses a verified primitive generator so nonzero output coordinates are distinct.
Each source chunk produces parity, reduced per GPU; the host combines one parity
symbol per GPU and writes coordinate zero after all cards join. No additional
full device output array or host transform grid is created.

## Resource and timing boundaries

At s17,m2,t5 with exact K3876566203, principal resident code/work/parity arrays
project **398.457GiB aggregate**, with **104.004GiB maximum on one of four GPUs**.
Metadata, CUDA contexts/driver/peer mappings and headroom are additional. Unequal
P placement is included. These are capacity projections, not measured peaks or
runtime guarantees. Keeping host code/message/output projects220.883GiB arrays;
host output alone is64GiB. Retained older words add64GiB each. Existing full-job
resource admission is separate from this API and the small validation budget.

Resident code stays on device: only a message, canonical result and small parity/
error observations cross the host boundary per call. There is no newly selected
slower memory-saving mode or extra transform; new peer routing/scatter costs and
full-size performance remain unmeasured. Both vector layouts are explicit,
using the existing executor; local routing adds transpose passes as documented
for preprocessing, and small measurements cannot select the fastest full encoding.

`CUDAEncodeStats.TimingSource` is `synchronized-host-wall` for this path. Native
stage clocks include launches and cross-card joins; input timing includes upload/
padding, and output timing includes parity/error observations and download.
Setup/copying constants is outside encoding. `ElapsedNS` also includes Go output
allocation/dispatch; old resident API kernel clocks retain their legacy CUDA-event
boundaries. Do not compare stage figures as identical measurement boundaries.

`LIFTRS_CPU_PROGRESS=1` emits the common ten operation names and advancing native
work. Field validation reports blocks, logged setup transfers report completed bytes;
transforms retain pass/vector counters;
canonical assembly uses bounded waves of chunks with periodic completed-work
updates. Logs expose public code dimensions/progress, not message values. Logging
and CPU profiling costs are included in validation timings, not claimed zero.
This is standalone mathematical encoding, not application/masking integration or
a production-security or constant-time claim.

## Full s17 first attempt

76646 successfully loads the genuine file and allocates398.463GiB aggregate
resident device state on4H200, then fails the360GiB host/cgroup guard during host
output allocation/message validation. ProcessRSS221.577GiB is within its320GiB
limit, but source-file cache adds~143GiB. That attempt produced no complete GPU encoding or verified GPU word. CPU76645 passes with the same input, providing canonical/
packed reference digests. [Full review/retry proposal](../experiments/s17-full-encoding/report.md#completed-first-attempts-scaling-093)
keeps fast resident mode and suggests more host/cache admission, not slower
computation. That first-attempt review did not launch a retry; **retry76706 subsequently
passed**, as recorded below.

## Full s17 resident GPU validation

**76706 passes**, HPCgnode04,4H200/32 CPU cores/640 GiB requested host RAM,
with112 GiB/card owned cap and608 GiB total-memory guard. Encoding434.930 s,
code load318.948 s, setup106.844 s; complete scheduler22 min11 s includes full
correctness and34 GiB packed output write/readback. All3876566203 systematic
positions,16 scalar curves, all-value range/canonical hash and packed hash/
readback pass; complete canonical and packed digests match CPU76645 exactly.
[Evidence and measurement boundaries](../experiments/s17-full-encoding/report.md#completed-full-gpu-retry-scaling-102).

Peak sampled process RSS221.591 GiB is separate from398.463 GiB device-owned
allocations observed at setup/return and261.071 GiB cgroup peak including cache.
Whole-device/driver/context peak was not sampled. Single profiled encoding-stage
ratio to CPU is14.51×; different nodes/cache states and setup costs preclude a
controlled end-to-end speedup claim. Other s17 memory modes and fulls18 remain
unvalidated. Retain explicit backend choice and CUDA-free CPU client builds.
