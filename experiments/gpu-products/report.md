# Resident GPU convolution in preprocessing

Status: corrected job **72839 passes** product/memcheck and complete smaller-code
comparisons; regression job72841 passes. Starting revision
`715bac2` plus recorded changes. Follows the
[matrix composition experiment](../gpu-matrix/report.md) and the
[current plan](../../docs/encoding-scaling.md#completed-comparison-checkpoint-and-next-integration).
No full s16 run or forecast is admitted here.

## Reproduce and verify

[run.py](run.py), [run.sbatch](run.sbatch), [resident ABI](../gpu-matrix/gpu_api.h),
[CUDA implementation](../gpu-matrix/matrix.cu), [Go bridge](../gpu-matrix/bridge.go),
[product dispatch](../../encoding/liftrs-shared/orbit_product.go),
[independent dispatch tests](../../encoding/liftrs-shared/resident_product_test.go)
and [GPU checks](../../encoding/liftrs-shared/gpu_resident_test.go) define the work.

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-products/slurm-%j.log \
 experiments/gpu-products/run.sbatch products-01
python3 -B experiments/gpu-products/analyze.py products-01
```

Use fresh names. Initial bounds are one GPU, 16 CPU cores, 16 GiB requested host
RAM, 12 GiB sampled process-group RSS guard, 512 MiB explicit device allocations,
30-minute allocation, 25-minute controller deadline and at most 2^22 transform
values. Driver/context memory is separate from explicit buffers. CPU native
scratch remains allocated. Preserve every attempt and cache; nothing is deleted.

The new private resident executor hook replaces each eligible local convolution
inside the existing overlap-add loops. Full product degree determines transform
size even when only a prefix is returned. The ABI checks operand/prefix/domain
bounds, zeroes both device buffers, uploads actual inputs, performs two forward
transforms, pointwise products and inverse, then downloads only the requested
prefix. Host inputs are borrowed synchronously and unchanged. Small products and
unsupported executors keep the existing path. Workspace leases/diagnostic
callbacks stay in the same caller scope; resident products omit CPU x/y scratch.
The old equal-half `gm_convolution` delegates to the new general operation.

The optional ABI symbol preserves loading old transform-only binaries; requesting
a resident product on an old binary reports an error. Regression uses the exact
checksum-verified matrix fixture and old library from `gpu-matrix/matrix-01`.
Full matrix/convolution outputs are rechecked. New field-degree-4 GPU tests force
small resident products and overlap-add beyond the field domain, comparing all
coefficients with generic Go arithmetic. Unequal/truncated/zero products, changing
sizes, input immutability, invalid dimensions, missing plans and closed engines
are checked; CUDA memcheck runs those cases if available. CPU-only tests exercise
the same dispatch, mixed fallback and workspace recovery after an invalid result.

Complete s10,m2,t5 runs compare CPU, transform-only GPU, resident GPU at the current
cutoff, and resident GPU with product cutoff1024. Two first-use samples and a
fresh cached process per variant verify exact identity/support/constants/full
words, every systematic position,16 curves and1,027 readbacks. Variant order is
fixed, profiling and internal diagnostic GC stay enabled, and NFS persistence
is included by phase. Initial GPU setup is part of complete first-use clocks;
CUDA event stage times are separate. No unprofiled/cold-I/O or randomized claim.

The native compiler/toolkit remain GCC14.2.0/CUDA12.9, C++17 `-O3`, native device
architecture; Go1.27.3. Device/CPU/allocation/source hashes, binaries, profiles,
results and checksums are recorded. Private caches are retained under
`.cache/liftrs/gpu-products/<run>/`; no new large-code cache is created.
Default constructors, field coordinates, file formats and CPU online encoding
remain unchanged. CUDA is optional and this remains an experimental backend.

## Results

The [corrected raw record](runs/products-02/results.json),
[analysis](runs/products-02/analysis.json), [product checks](runs/products-02/product-checks.json)
and [CUDA memcheck](runs/products-02/memcheck.log) pass. The analyzer verifies269
sources plus binary/cache/reference/evidence hashes;100 product cases and invalid/
closed/old-ABI behavior pass, with zero memcheck errors. Twelve complete/cached
s10 runs match every support/constant/word/readback check across all variants.

Measured on gnode03, Xeon6710E, one H200 NVL, driver580.126.20, CUDA12.9.86,
GCC14.2.0 and Go1.27.3; see allocation/GPU/compiler logs. Median of two first-use
samples per backend, seconds:

| Backend | Preparation | Sum of all phase clocks, including setup/checks/I/O |
|---|---:|---:|
| CPU, product cutoff32768 | 6.647 | 9.756 |
| Transform-only GPU, same cutoff | 2.735 | 5.898 |
| Resident GPU, same cutoff | 2.400 | 5.666 |
| Resident GPU, cutoff1024 | 1.528 | 4.644 |

Resident integration at the unchanged cutoff gives1.14× preparation speedup over
transform-only GPU (12.3% less preparation time). Earlier GPU use gives a larger
combined effect:1.79× versus transform-only GPU and4.35× versus CPU preparation.
These are sequential, instrumented samples on a data-center GPU; no universal
cutoff, pure hardware advantage or full-s16 ratio follows. Complete phase clocks
include GPU setup, CPU online encoding, checks and persistence; polling/profile
shutdown makes worker process clocks different.

The first default resident run uses9 resident convolutions and3 standalone GPU
transforms, replacing30 transform-only calls. Cutoff1024 uses84 resident products.
Host CPU x/y scratch is avoided for eligible products; output prefix assembly and
outer overlap-add remain Go. Default library behavior remains CPU. Sampled command
RSS peak242.5 MiB; private cache86.63 MiB including binaries/constants/fixtures.
Profiles are retained separately per variant. All old caches remain intact.

[validation-05](../hpc-s16/runs/validation-05/results.json), job72841, passes full
Go/native repository suites, focused race/vet, legacy build and small fresh-cache
checks. Its evidence analyzer passes. CGO-disabled resident-dispatch/prefix tests,
Python/shell syntax, gofmt, links and whitespace checks pass. CUDA-specific checks
include full matrix regression, direct product comparisons and the old library.
The subsequent [bounded s16 intervals](../s16-resident-intervals/report.md) use
the verified resident library, without treating this small-code ratio as a forecast.

Initial attempt [products-01](runs/products-01/results.json), job72838, passed
matrix regression but stopped in the small-product test: its test-only Go
fallback called the full-field transform on a short buffer. Corrected that
fallback to `applySubspace`; the real preparation fallback was already the
length-aware native executor. Logs/source snapshots are preserved. The corrected
attempt in the fresh `products-02` directory passes the checks reported above.
