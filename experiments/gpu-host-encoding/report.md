# Complete host-backed GPU encoding

GPU-041 integrates the five-transform online schedule using validated
[parallel host-backed transforms](../gpu-backed-parallel/report.md).
`WithCUDAHostBacked(options,tileElements)` is explicit; resident `WithCUDA` stays
the default. The small-tile component remains10.47x the prior resident sample,
so this is a memory alternative with a substantial possible runtime penalty.
No full s16/s17 run is admitted by this experiment. CPU output/disk-backed work
still follows completion of the GPU route.

## Reproduction

[Controller](run.py), [Slurm script](run.sbatch), [analyzer](analyze.py),
[self-check](check.py), [API](../../encoding/liftrs-shared/cuda_host.go),
[native schedule](../../encoding/liftrs-cuda/backed.hpp),
[small tests](../../encoding/liftrs-shared/cuda_host_test.go),
[complete workflow](../../encoding/liftrs-shared/cuda_workflow_test.go).

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-host-encoding/slurm-%j.log experiments/gpu-host-encoding/run.sbatch online-01 --memory-mode host-backed
# Only after the user reports completion:
python3 -B experiments/gpu-host-encoding/analyze.py online-01
```

[Local compile/self-check evidence](compile-check.json) passed; GPU runtime
validation **passed job74129**, run `online-01`, submitted2026-11-12.
[Results](runs/online-01/) and [Slurm log](slurm-74129.log) are populated by the
job. After the completion ping,382 source identities and all cache/evidence
digests reconciled; see [analysis](runs/online-01/analysis.json). Revision is ce29557 plus recorded working sources;
results record exact HEAD, SHA256 of all sources and content-addressed snapshots,
commands, device/CPU/compiler environment and sealed results/cache hashes.
New cache `.cache/liftrs/gpu-host-encoding/online-01/` retains the library, binary
and packed workflow outputs. Existing small fixtures from GPU-038 job74094 are
verified against its sealed hashes; cached s10 code comes from
`.cache/liftrs/gpu-products/products-02/cpu-0/online.bin` with trusted metadata and
mandatory full digest verification. No discovery/code preprocessing repeats.
CPU profiles are retained separately for tests, setup, encoding and verification.

One H200,16CPU,GCC14.2/CUDA12.9/Go1.27.3. Same reduced limits:16GiB requested host
RAM,12GiB sampled process-group RSS,8GiB Go soft limit,512MiB explicit device cap,
16GiB each new cache/scratch;20min controller,15min worker,30min Slurm deadline.
Initial32GiB available host/runtime16GiB headroom and1GiB free device checks.
Driver/context allocations are extra. Guards execute within the job; no agent
monitoring until completion ping. Only small/cached s10 jobs are authorized here.

## Schedule and correctness

The immutable plan retains P and spectrum in host RAM. One mutable host N array
contains two L prefix operands, requiring2L<=N. GPU-backed transforms compute
both forward transforms, their tiled product and inverse. Retain K coefficients,
zero the rest and compute the full forward transform, spectrum product, inverse.
All extension multiplication/transform arithmetic and orbit point generation run
on GPU. Native CPU scatters immutable coefficient values to canonical positions
and XORs the coordinate-zero parity. A primitive generator is checked before
writes; no N-sized orbit/permutation map or visited bitmap exists. Two device
tiles and at most one host tile serve every phase. C++ retains no borrowed Go
pointer. Inputs are immutable; errors may leave private work/output partial.

All calls/Close serialize through the shared encoder lock. Arena persists across
calls; the owning result is new each time. Older libraries reject the optional
online ABI. Pointer-overlap, field, shape, primitive-generator and closed-state
checks guard native boundaries. Checksums authenticate byte identity only;
trusted mathematically verified code remains a requirement. Current s2..16
arithmetic/cryptographic scope is unchanged. No CUDA dependency in CPU builds.

Small s2/s3/s4 constant subcodes and s6/s8 retained codes use two tile sizes and compare every output to Go across
nonzero/zero/repeated messages. s2/s3/s4 retained fixtures exceed this schedule's
prefix arena and must reject, followed by one-symbol constant subcodes that
exercise successful small-field encoding and every canonical position. Invalid length/range, recovery, unsupported packed
sink, overlapping work/output, old ABI, stable allocation and Close are checked;
repeat under compute-sanitizer memcheck/leak checking. Eighty prior matrix
regressions and Go-only/CUDA-only/combined build checks, race/vet also run.
Complete s10 CPU and host-backed GPU workflows each perform two profiled encodes,
all systematic positions,16 sampled independent curves, full canonical/packed
hash agreement and1027 packed-file readbacks. CPU comparison is native, not Go.

## Resource and timing boundaries

Host principal arrays:4(K+1+N) bytes code,4K message,4N retained arena,4N owning
result,<=4T temporary native staging. Runtime, caller-retained earlier outputs,
file cache and validation arrays are extra. Device arrays8min(N,T) bytes plus
field tables and two matrix plans, with driver context additional. There is no
full code/spectrum retained on GPU. Products upload both operands per tile;
transforms repeat transfers for Taylor corrections and row/column passes;
canonical points add4(N-1) D2H bytes plus random host scatter. No production disk
workspace; packed artifacts add Ns/8 storage and corresponding file-cache use.
The host-backed path currently rejects EncodePackedTo: its result is owning,
while the existing resident packed sink remains available. Bounded output for
this backed schedule is a further integration issue, not silently emulated by
allocating a hidden full canonical array behind the packed API.

For s16 the above arrays project54.84GiB host payload plus staging/runtime/cache,
including32GiB arena+result, versus the resident streamed-code path's23.01GiB
measured encoding RSS/GPU32.005GiB. These are different memory schedules, not a
matched measurement. T65536 projects roughly1.3MiB explicit GPU arrays/tables/
plans. Device context is extra. Full s16 runtime and total RAM are unknown;
these allocation formulas do not establish s17 feasibility or admit a full run.

Stats: first10 native fields follow ordinary online H2D/Taylor/columns/rows/layout/
D2H/product/native-total/buffer/assembly seconds. Extra fields count H2D,D2H
bytes, peak native host staging, CPU routing seconds, five transforms and arena
bytes. Native total excludes argument/code/generator admission, while Go ElapsedNS
includes native admission and host output/first arena allocation. AssemblyNS is
GPU point generation; CPU scatter/XOR appears in HostRoutingNS. Workflow setup,
code loading, packing/writing and verification are separately timed. Timings are
profiled fixed-order samples, not an unprofiled full encoding speedup claim.


## Measured result

Job74129 completed0:0 in4m25s on gnode17/H200. Every small/reference/constant
case, memcheck (zero errors/leaks),80 matrix regressions, ordinary build/race/vet
and complete s10 CPU/GPU workflows pass. Two hashes agree, all146641 systematic
positions and16 sampled curves per encoding pass;1027 packed readbacks pass.

| s10 profiled phase | Native CPU | Host-backed GPU |
|---|---:|---:|
| Setup seconds |0.003249|1.058850|
| Encode 1 seconds |0.319386|1.616659|
| Encode 2 seconds |0.312775|1.606209|

This GPU memory mode is5.1x slower than native CPU here, not a GPU speedup.
Explicit device ownership560400bytes (0.534MiB), retained host arena4MiB,
new owning output4MiB, native staging256KiB, plus code/message/runtime. Traffic
per encoding456069120 H2D/128913404 D2H bytes. Second call Taylor0.508s,
H2D0.324s,D2H0.127s,row/column GPU0.0248s,CPU routing0.0102s, remainder includes
synchronous launch/event overhead. Full s16 latency remains unmeasured.
Peak controller-sampled process-group RSS1054543872bytes includes all checks;
Slurm MaxRSS241528KiB uses a different task/process sampling view. Neither is
an isolated encoder RSS measurement. Profiles/code/artifacts retained.

The complete owning GPU route is validated through s10. A backed packed sink
and larger-run admission remain separate integration/measurement choices; the
already validated resident GPU packed sink remains. Continue the authorized CPU
packed-output gate, then CPU disk-backed transforms. Faster resident defaults
remain; no material slower mode is enabled automatically.
