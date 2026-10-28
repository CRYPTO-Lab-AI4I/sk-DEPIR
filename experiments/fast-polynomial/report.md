# Fast polynomial products and encoder comparison

## Purpose and reproduction

Implement the polynomial-multiplication part of [Step 1](../../docs/encoding-algorithms.md#step-1-remove-full-extension-field-arithmetic-tables),
then compare it with the table-free Karatsuba implementation. The [script](run.py)
keeps baseline and new measurements on the same machine; [results.json](results.json)
records exact source/input hashes, environment and raw samples. [table.md](table.md)
summarizes the whole-encoder measurements; [kernels.txt](kernels.txt) retains all
kernel timings and allocation counts.

From the repository root:

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/fast-polynomial/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

Requires compatible Go, Python 3 with POSIX wait4, git history containing baseline
`7e5a8ac`, and writable temporary storage. The new implementation was measured on
top of that revision with uncommitted ENCODE-004 changes; its source hashes identify
the actual measured files. The baseline is extracted from git and built in an
isolated temporary directory. Prior reports, results and public artifacts are
preserved. There is no database generation or network dependency during the run.

## Algorithm and reusable information

The [product backend](../../encoding/liftrs-go/orbit_product.go) uses the existing
extension-field additive FFT on power-of-two binary subspaces. Forward-transform
both polynomials, multiply their values and inverse-transform the result. Choose
a domain large enough for the **entire local convolution**, even when returning
only its low coefficients; otherwise discarded high terms could alias into them.
Products too large for the extension field use overlap-add with half-field chunks.
No larger field or change to code parameters is introduced.

Small products retain Karatsuba. The low-product routine skips recursive branches
that cannot reach the requested prefix, while retaining high-half terms needed
for Karatsuba's cross term. Very unequal operands are split into balanced blocks
when below the FFT cutoff, avoiding padding a short operand to the long operand's
full size. Schoolbook multiplication remains for an operand shorter than 24.
FFT dispatch requires **both operands** to meet the cutoff: 1024 coefficients for
extension-field products and 32768 for base-field products. These conservative
local choices are not claimed to be optimal for every field, input shape or CPU.
For example, forced FFT beats automatic dispatch on the tested full base-field
17200-by-17200 product, while pruned Karatsuba wins on its low half.

[Encoder integration](../../encoding/liftrs-go/orbit.go) reuses immutable transform
constants already prepared/loaded with the code. Newton inversion, reversed
quotient calculation and both encoders' prefixes use low products. The paper's
product/remainder tree and cyclic generator construction use the full backend.
Per-call operand/value/output buffers depend on the supplied polynomials and are
never serialized or shared between concurrent calls. No keys, masks or application
layout enter code preprocessing. Public artifact formats and bytes are unchanged.

Local FFT multiplication takes O(n log² n) extension-field operations, including
monomial-basis conversion. Returning a prefix does not turn it into a specialized
truncated FFT: transforms still cover the full local degree. Product-tree overhead,
scalar field multiplication, in-memory scratch and conservative dispatch remain.
This implements fast full/low polynomial products; it is not an optimal transform
implementation or evidence of terabyte-scale encoding feasibility.

## Correctness and validation

[New tests](../../encoding/liftrs-go/orbit_product_test.go) cover:

- Independent schoolbook convolution using carryless/tower arithmetic, including
  empty inputs, zeros, non-power-of-two lengths, partial/full outputs and unchanged
  operands. Force both FFT and Karatsuba paths.
- Products crossing the field-domain boundary and unequal operands in both orders.
- Every binary subspace size through 256, checked by direct Horner evaluation and
  transform round trips; reject invalid transform sizes.
- Series-inverse and division identities with forced FFT products.
- A forced-FFT paper/cyclic pipeline checked against the independent dense encoder
  on a small Lift-RS instance.

Existing [dense-oracle and encoding tests](../../encoding/liftrs-go/orbit_test.go),
[historical output/artifact and sampled-curve checks](../../encoding/liftrs-go/binary_field_test.go)
remain in place. The experiment independently rebuilds the old and new implementations,
reprepares both methods at s=8,t=8 and s=9,t=7, and requires byte-identical artifacts
to the committed [checkpoint inputs](../tablefree-encoding/artifacts/). It checks
all systematic symbols, old/new canonical output hashes, packed output hashes and
packed byte counts. Exact K remains 3121 and 17200, respectively; no catalogue
capacity or supported-size change is implied.

Verification completed after the final measurements:

```sh
go test ./encoding/liftrs-go -count=1
go test -race ./encoding/liftrs-go ./pirconfig
go vet ./encoding/liftrs-go ./pirconfig
go test ./...  # isolated repository copy: legacy tests write CSV files
```

All passed. `git diff --check`, result/source/input hashes and report links were
also checked. No protocol or security guarantee follows from arithmetic agreement.

## Measurement boundaries

The baseline and new binaries use the same unchanged
[checkpoint driver](../tablefree-encoding/main.go). Each preparation and encoding
stage runs in its own fresh process. Shared basis discovery is reused from the
saved input, not rerun. Preparation time excludes loading that basis; encoding
loads the complete prepared artifact once before three repetitions of the same
uniform message (seed 71). Report median encoding time and cumulative bytes
allocated per call; preparation has one observation per method/backend/config.
Cached spectrum generation belongs to preparation, not per-database time.

Packing, hashing and sync-writing a small canonical file occur outside encoding
time, using the unchanged driver. They are not benchmarks of grouped storage or
sustained I/O. Process peak RSS includes runtime, artifact loading, encoding and
output buffers; it is not isolated encoding workspace. Allocation totals are not
peak live memory. No warmup, CPU affinity or cache flush is used. Run old then new
at s=8 and new then old at s=9; the order is explicit but not randomized.

Kernel benchmarks use [BenchmarkPolynomialKernels](../../encoding/liftrs-go/orbit_product_test.go):
base and extension coefficients, lengths 1024/4096/17200, full/low output, and old
full-product-and-slice, pruned Karatsuba, forced FFT and automatic dispatch. Input
seed 207 and immutable transform setup are outside timing. Run three single-call
samples (`-benchtime=1x -count=3`). Raw allocation counts are retained. This is a
short local comparison without confidence intervals, not a stable performance gate.

Environment: Apple M5 Pro, 18 cores, 48 GB RAM, macOS 26.6.2 ARM64, Go 1.27.3,
Python 3.9.6, default GOMAXPROCS=18. Arithmetic is single threaded; runtime/GC
workers can run concurrently. CPU/model/memory and temporary-volume `df` output
are recorded; device/filesystem performance is not characterized. Final measurement
runs completed before race/full-suite tests; no test suite ran alongside them.

## Results and trade-offs

See the [complete table](table.md). At s=9:

- Paper encoding: **4143.95 → 1544.79 ms (2.68× faster)**; reusable preparation:
  **8522.68 → 3311.10 ms**. Allocated bytes per encoding drop from 279.75 to 33.97 MiB.
- Cyclic FFT encoding: **490.45 → 430.43 ms (1.14× faster)**; reusable preparation:
  **3904.11 → 3409.64 ms**. Allocated bytes drop from 59.17 to 21.35 MiB.

At s=8 paper encoding improves 1.32×. Cyclic encoding is effectively unchanged
(71.36 → 70.90 ms), while its preparation **regresses by about 15%**
(385.53 → 442.82 ms). Keep that regression visible: this change is not a universal
speedup, and further shape/field-specific tuning needs evidence. The user accepts
this trade-off because preparation is performed once per code and reused across
databases; it does not increase per-database encoding time. Artifact sizes and
codewords are unchanged. Cyclic remains the faster database encoder in both cases.

The median extension-field length-17200 low product falls from 1028.48 ms with the
old full-product-and-slice path to 120.88 ms with automatic dispatch. For the same
base-field case, pruned Karatsuba reduces 211.07 to 96.68 ms and is faster than
forced FFT (110.43 ms). Whole-encoder gains are smaller than these dense random
kernel ratios because their polynomial shapes, coefficients and remaining work differ.
The [earlier profiles](../encoder-profile/report.md) already showed FFT arithmetic
dominating cyclic database encoding; optimizing the prefix cannot remove that work.
No new CPU profile was collected, so its post-change percentage is not asserted.

### Initial dispatch trial

The first trial selected FFT from the **larger** operand's length. Its
[results](initial-results.json), [table](initial-table.md) and [kernels](initial-kernels.txt)
are retained because it regressed cyclic preparation at both sizes: 386.0 → 692.3 ms
at s=8 and 3903.4 → 4682.8 ms at s=9. That trial exposed expensive padding for
unequal operands. Final dispatch requires both operands to meet the cutoff and
uses balanced blocks below it. This removed the s=9 regression; the smaller s=8
regression remains as reported above.

The first trial's differing source files are archived as
[orbit_product.go](initial-orbit_product.go.txt),
[orbit_product_test.go](initial-orbit_product_test.go.txt) and
[orbit.go](initial-orbit.go.txt). Their bytes match the initial result's source
hashes; the remaining recorded source files are unchanged. To reproduce that trial,
restore those three files in an isolated copy and run the same script. Its
measurements informed dispatch selection, so the final sample is not an independent
holdout benchmark.

## Next work

This checkpoint implemented and validated the general/low-product backend within
s<=9. Scalar field arithmetic and transforms remain important costs. The later
[s=10 CPU checkpoint](../encoder-scaling/report.md) advances the comparison with
stage-specific memory and packed output measurements. Current work prioritizes
scaling measurements over artifact-format cleanup unless loading/size obstructs
those experiments. Field-kernel improvements can be evaluated separately. External-memory and
GPU work remain behind the existing scope decision gates.
