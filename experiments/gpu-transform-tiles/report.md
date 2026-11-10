# CPU/GPU affine transform tiles

This implements the bounded GPU comparison in the
[current plan](../../docs/encoding-scaling.md#next-comparisons-cpu-products-and-gpu-transform-tiles).
It measures a component of large transforms, not a complete GPU matrix/Taylor
backend or full s16 preprocessing. Starting revision `4cb6193` plus exact recorded
trial sources. CPU remains the library backend.

## Reproduction and method

[run.sbatch](run.sbatch), [shared controller](../preparation-kernels/run.py),
[CUDA prototype](tiles.cu), [fixture reader](../preparation-kernels/fixture.hpp),
and [analyzer](../preparation-kernels/analyze.py) define reproduction:

```sh
module load slurm/slurm/25.05
# First complete the CPU fixture job described in the CPU report.
sbatch --output=experiments/gpu-transform-tiles/slurm-%j.log \
 experiments/gpu-transform-tiles/run.sbatch gpu-02 --cpu-run cpu-01
python3 -B experiments/preparation-kernels/analyze.py gpu gpu-02
```

Use fresh names; reproduce old/current variants from recorded sources. The job
requests one GPU, four CPUs, 16 GiB, 30 minutes; controller guards sampled process
RSS at 12 GiB and runtime at 25 minutes. Explicit device buffers/tables are capped
at 512 MiB (measured maximum 34,342,556 bytes). CUDA context/driver memory is
additional and is not bounded by this explicit-allocation cap. No large caches
are changed. Checksum-verified CPU fixtures and matching binaries remain outside
Git at `.cache/liftrs/{preparation-kernels,gpu-transform-tiles}/`.

The GPU uses the same GF(2^16) polynomial and packed quadratic coordinates as
the CPU. No floating-point substitution or field isomorphism occurs. A thread
owns each independent affine tile; coefficient-major layout makes simultaneous
tile accesses contiguous. Sparse division and butterfly dependencies remain
sequential within each tile, following the CPU transform. Forward/inverse are
matched against the native CPU for all values, using fixed-seed full-width random
coefficients, tile sizes 16/64/256 and total sizes 2^18/2^20/2^22.

CPU reference uses the current cached-constant native transform with four threads.
GPU uses bounded base-field log/power tables. The initial trial compares these
different arithmetic strategies; the second additionally times a CPU tile path
with the same log arithmetic as the GPU, checking every output against the native
reference. That separates an important arithmetic advantage from hardware benefit.
CPU lookup is an experiment reference, not a new integrated library backend.

Three repetitions record CPU forward/inverse wall time, GPU forward/inverse CUDA
events, synchronous host/device copies, both layout conversions and complete
forward host-to-host wall time. CPU timings include four-thread launch/join;
GPU kernel-only times exclude event creation/launch overhead, which is included
in complete forward time. Host buffers are ordinary pageable allocations.
Copies use CUDA runtime synchronous calls. Output comparisons are outside timers.
Input initialization is outside all kernel times; the same coefficients are used.
GPU context, table upload and buffer allocation are measured separately and are
reusable setup, excluded from steady-state forward times. No performance claim
includes them silently. No Nsight/perf hardware profile was collected; CUDA events
provide stage instrumentation and the CPU trial retains its Go profile.
Earlier allocations denied native perf profiling (`perf_event_paranoid=4`);
no privileged setting was changed. These clocks do not provide native instruction
or GPU occupancy attribution. Further tuning needs a separate supported profiling
tool run; no such fine-grained attribution is claimed here.

## Evidence and findings

[Initial job 72821](runs/gpu-01/results.json) / [analysis](runs/gpu-01/analysis.json)
passes all native CPU/GPU forward and inverse comparisons. This justified a
second job with a stronger CPU lookup reference, rather than interpreting all
of that initial gain as a GPU hardware effect.

[Matched-arithmetic job 72825](runs/gpu-02/results.json),
[analysis](runs/gpu-02/analysis.json), [GPU inventory](runs/gpu-02/gpu.log) and
[compiler](runs/gpu-02/nvcc.log) record NVIDIA H200 NVL, 150,109,880,320 bytes
reported device memory, driver 580.126.20, CUDA 12.9.86 and GCC 14.2.0 on Linux.
Both GPU jobs ran on gnode15 with Xeon 6710E host CPUs; only one GPU and four
CPU cores were allocated. Peak sampled process-group RSS among commands was
252.38 MiB initially and 211.54 MiB in the matched-arithmetic job. Standalone
binaries occupy about 1.1 MiB each outside Git; the shared fixture and CPU code
constants remain in the CPU trial cache.
This is a data-center GPU measurement, not evidence of consumer-GPU performance.

For 2^22 values, median forward times in seconds were:

| Tile | Native CPU, 4 cores | Lookup CPU, 4 cores | GPU kernel | GPU incl. transfers/layout | Lookup CPU / GPU incl. transfers |
|---|---:|---:|---:|---:|---:|
| 16 | 0.38799 | 0.07389 | 0.001023 | 0.003836 | 19.3× |
| 64 | 0.53603 | 0.11545 | 0.001550 | 0.004738 | 24.4× |
| 256 | 0.75045 | 0.17845 | 0.004127 | 0.007185 | 24.8× |

Context initialization was 0.6448 s and table allocation/upload 0.00965 s.
A first call paying that setup can lose to the CPU even though the warmed tile
kernel wins. Reuse and workload size therefore matter. Kernel/transfer medians
alone must not be labeled first-use latency. The analyzer verifies all 27 cases,
source/binary/evidence hashes, shared fixture and device-allocation bound.

This supports extending the GPU prototype to matrix/Taylor conversion and larger
tile sizes, keeping data resident across successive transforms/products. It does
not predict a 19–25× whole-encoder speedup. Large-level merges, full transform
composition, persistent workspace, reciprocal/identity, database I/O and full
s16 memory remain unmeasured. Also compare more CPU cores and native lookup
integration before selecting a deployment backend. CUDA is an optional experiment
dependency; the generic CPU path remains usable.
