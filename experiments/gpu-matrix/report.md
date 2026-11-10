# Complete experimental GPU matrix transforms

Status: job **72835 passed**, including complete transform/preparation comparisons
and small CUDA memcheck (zero errors). Implements the
[published composition plan](../../docs/encoding-scaling.md#completed-comparison-checkpoint-and-next-integration),
starting revision `612e491` plus exact recorded experiment sources. This remains
an opt-in CUDA experiment, separate from the default CPU library backend. No full
s16 discovery/preparation or automatic readmission is included.

## Reproduction and boundaries

[run.sbatch](run.sbatch), [controller](run.py), [CUDA engine](matrix.cu),
[native API](gpu_api.h), [driver](bench.cpp), [dynamic Go bridge](bridge.go),
[fixture generator](../../encoding/liftrs-shared/gpu_matrix_fixture_test.go) and
[preparation adapter](../../encoding/liftrs-shared/gpu_preparation_test.go) contain
the complete implementation. From the repository root:

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-matrix/slurm-%j.log \
 experiments/gpu-matrix/run.sbatch matrix-01
python3 -B experiments/gpu-matrix/analyze.py matrix-01
```

Use fresh names and the exact source snapshot for an old run. Each job requests
one GPU, 16 CPU cores, 16 GiB host RAM and 30 minutes; the controller guards
process-group RSS at 12 GiB and total execution at 25 minutes. Maximum transform
size is 2^22 values. Explicit device buffers/tables/plans, including temporary
allocations during growth, are bounded by 512 MiB. CUDA context/driver allocations
are additional. No existing caches are opened for writing or removed.

Sources (including CUDA), compiler/driver/GPU/CPU/OS/allocation, matching binaries,
fixtures, profiles and SHA-256 hashes are retained. Large artifacts live outside
Git at `.cache/liftrs/gpu-matrix/<run>/`; regenerate with a fresh name. Native build
uses CUDA 12.9, GCC 14.2.0, C++17 `-O3`, architecture selected from the allocated
device. Ordinary builds do not need CUDA: `liftrs_cuda` enables only the optional
experiment bridge/test, which loads the explicitly supplied shared library.

## Algorithm and correctness

The GPU uses the existing packed polynomial/tower basis, base-field lookup tables
and matrix factorization. Canonical affine row tiles and arbitrary-basis column
tiles reproduce CPU evaluation order. Columns already have coefficient-major
layout in the matrix buffer; row execution uses explicit GPU transposes.
No basis conversion or floating-point arithmetic replaces field operations.

Taylor conversion uses two parallel gather passes per level. In the linearized
divisor of degree H, every lower degree d is at most H/2. A quotient correction
at j reads only untouched upper-half quotient entries, so these corrections are
independent. A second pass gathers remainder contributions. Inverse reverses
the passes. Each thread owns one destination; separate kernel completion orders
the passes. This depends on linearized degrees and the canonical row schedule;
the parser rejects other shapes/degrees. Small fixtures check every point by
independent Horner, then compare complete GPU/native CPU outputs and inverses.

The engine owns immutable plans/tables and three persistent device buffers.
Calls borrow host input only until synchronous return. Workspace growth allocates
new buffers before swapping them in, preserving prior state on allocation failure.
The Go bridge serializes calls/Close; it retains no Go slice pointer in native
state. Resident convolution uploads two half-length inputs, clears padding,
executes both forward transforms, pointwise products and inverse on device, then
downloads the result. A zero-input follow-up checks stale padding. This does not
yet eliminate transfers from the Go preprocessing orchestration.

## Measurements

The fixed seed is 1601. Full-width extension coefficients at s16 are tested at
dimensions4/6/8/18/20/22. Three CPU forward/inverse samples at four and sixteen
workers, three complete GPU forward/inverse samples, and resident convolution at
2^18/2^20/2^22 compare exact full arrays. CPU convolution has one sample per worker
count; GPU convolution has three. GPU kernel stages use CUDA events; total calls
include launch/event overhead, synchronization and pageable host transfers.
Initialization, plan allocation/upload and workspace growth are measured separately.
These diagnostics are instrumented, not unprofiled latency baselines.

Complete s10,m2,t5 first-use runs compare CPU and GPU-assisted preprocessing twice
each, then fresh-process cached execution. The existing harness checks P·G, all
systematic positions, 16 independent curves, full hashes, packing/fsync and 1,027
readbacks. CPU native executor scratch is retained. Only extension transforms
of at least 4,096 values use GPU; base-field discovery transforms, root/product
orchestration, pointwise products outside the resident kernel and complete online
encoding remain CPU. Each orchestration-level GPU transform uploads/downloads.
Setup includes GPU initialization/plans. Shared NFS persistence and CPU profiles
remain enabled; internal diagnostic GC is included in preparation. Phase times
and complete process clocks differ by profiling setup/shutdown and other overhead.

The controller runs small CUDA memcheck if `compute-sanitizer` is available;
unavailability/failure is reported. Earlier native perf permission was denied;
CPU profiles and CUDA stage clocks are retained without native-instruction or
GPU occupancy claims. Results, stage measurements and analysis live in
`runs/<run>/`. Full s16 feasibility, consumer-GPU results, production security,
constant-time arithmetic and a public stable GPU backend remain unclaimed.

## Results and interpretation

[Raw record](runs/matrix-01/results.json), [analysis](runs/matrix-01/analysis.json),
[matrix/convolution samples](runs/matrix-01/matrix.log),
[fixtures](runs/matrix-01/fixtures.json) and
[memcheck](runs/matrix-01/memcheck.log) are retained. The analyzer verifies 264
recorded source files, every binary/cache/evidence hash and all case counts.
The host was gnode03, Xeon 6710E; one H200 NVL, driver 580.126.20, CUDA 12.9.86,
GCC 14.2.0 and Go 1.27.3. See recorded allocation/toolchain logs in the run directory.
This is a data-center GPU result, not a consumer-GPU measurement.

Complete forward transforms, median of three instrumented samples,
seconds (GPU includes upload/download and every conversion):

| Values | CPU, 4 cores | CPU, 16 cores | GPU |
|---|---:|---:|---:|
| 2^18 | 0.10781 | 0.04156 | 0.02334 |
| 2^20 | 0.49098 | 0.17308 | 0.05681 |
| 2^22 | 2.19455 | 0.75960 | 0.13747 |

At 2^22, the complete transform is 5.53× faster than the measured 16-core CPU
reference. The resident convolution takes 0.37926 s (three-sample median), versus
2.73664 s on 16 cores and 6.98238 s on four cores (one sample each): 7.22× versus
the 16-core observation. This compares actual backends, including GPU lookup
arithmetic versus current native field arithmetic, not hardware alone. Context/
table initialization is 0.6531 s, measured separately; each size's plan and buffer
setup is also separate in the raw samples. Smaller shapes can favor CPU once
launch/initialization costs are accounted for. No universal cutoff is selected.

Complete s10 runs, medians of two sequential samples per backend:

| Measured boundary | CPU | GPU-assisted |
|---|---:|---:|
| Preparation phase, seconds | 6.806 | 2.719 |
| Sum of all phase clocks, including setup/validation/I/O, seconds | 9.791 | 5.902 |
| Whole worker command including up to one-second controller polling, seconds | 10.743 | 6.790 |

Preparation is 2.50× faster in this trial; complete-worker observation is 1.58×.
These fixed-order, profiled samples are not randomized/load-frequency-controlled
latency baselines. First-use GPU setup is included in the complete phase/worker
figures, unlike warmed transform figures. CPU/GPU outputs, support, P/spectrum,
G and fresh cached codewords match exactly. Both cases establish K=146,641,
N=1,048,576; masks/application execution are not part of this experiment.

The first GPU-assisted run made 30 GPU transform calls, with 0.547 s total inside
the synchronous bridge and about 0.020 s host/device transfers. The Go
[preparation profile](runs/matrix-01/pprof-gpu-0.log) now attributes about 55% of
sampled CPU time cumulatively to Karatsuba and about 20% to cgo. That is this
small pipeline's CPU attribution, not a full s16 profile. Resident convolution is
validated separately but is not yet substituted into the Go polynomial-product
orchestrator. Lower product cutoffs, batching and moving whole convolutions into
the GPU are measured follow-up candidates; transform acceleration alone does not
remove the remaining CPU leaf work.

Explicit device ownership after all benchmark plans is 51,184,600 bytes
(48.81 MiB); growth temporarily retains old buffers and is guarded at 512 MiB.
This ownership count excludes driver/context allocations. Peak sampled process
RSS among commands is 257.18 MiB. Retained fixtures/binaries/code constants total
90.38 MiB. The Go bridge reports zero device ownership after Close; its engine
is single-caller/locked, and ordinary CPU constructors are unaffected.

Regression job 72837 in [validation-04](../hpc-s16/runs/validation-04/results.json)
passes both full repository suites, focused race/vet, legacy fixture build and
small fresh-cache checks; its analyzer passes. The independent Go Taylor-gather
test compares sequential long division, gather output and inversion without CUDA.
CUDA-tag focused tests, CGO-disabled preparation tests, Python/shell syntax,
gofmt, documentation links and whitespace checks pass. No full s16 run was launched.
