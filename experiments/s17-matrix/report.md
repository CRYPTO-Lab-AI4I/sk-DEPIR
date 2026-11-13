# Wide Go C interface and parallel CPU matrix transforms

Next stage of the [s17 implementation plan](../../docs/encoding-scaling.md#current-execution-plan-for-s17),
after [bounded native recursion validation](../s17-native/report.md). New wide
Go/C handles construct immutable ordered plans and canonical-grid matrix/Taylor
metadata. **Corrected retry job74330 passed**, source checkpointdc57f45;
[retry submission](retry-submission.json), [results](runs/matrix-02/results.json),
[reconciled analysis](runs/matrix-02/analysis.json). The original harness failure
and its unchanged evidence remain documented below. Full s17 encoding,
persistent wide code formats and preprocessing remain unimplemented.

[Controller](run.py), [Slurm script](run.sbatch),
[Go ownership/API](../../encoding/liftrs-cpp/wide.go),
[C declarations](../../encoding/liftrs-cpp/wide_api.h),
[native implementation](../../encoding/liftrs-cpp/wide_api.cpp),
[shared execution](../../encoding/liftrs-cpp/matrix.hpp),
[Go checks](../../encoding/liftrs-cpp/wide_test.go),
[usage example](../../encoding/liftrs-cpp/wide_example_test.go),
[direct C ABI sanitizer check](check.cpp).
Results, exact revision/source hashes, commands/environment and profiles will be
retained at `runs/matrix-01/` and `runs/matrix-02/`; their fixtures and explicitly
built/profiled binaries are retained under the corresponding
`.cache/liftrs/s17-matrix/<run>/` directories outside Git. Use a fresh name when
reproducing, for example matrix-03, to preserve both recorded attempts.

```sh
python3 -B experiments/s17-matrix/run.py --check
module load slurm/slurm/25.05
sbatch --output=experiments/s17-matrix/slurm-%j.log \
  experiments/s17-matrix/run.sbatch matrix-01
```

## Interface and implementation scope

NewWidePlan(s,polynomial,basis) validates the component field using the common Go
implementation, selects its deterministic tower and copies an independent basis
into native metadata. ApplyAt borrows uint64 data for one affine transform.
NewWideMatrix(s,polynomial,dimension,split) constructs the canonical binary grid
matrix decomposition: h-adic Taylor conversion, column transforms/transposes and
affine rows, reversing order for inverse. Apply requires two disjoint N-element
uint64 slices and1–64 workers. No caller pointer is retained; plans are immutable,
concurrent calls require disjoint buffers, and Close waits through Go locks.
OwnedBytes counts native metadata capacities and becomes zero after Close.

C handles are separate from the old32-bit ABI. The C boundary validates shapes,
ranges, dependent bases and buffer overlap; Go additionally validates field
irreducibility. Native field parameters are trusted code data, not authenticated
PIR code specifications. Shape/range failures precede mutation, execution errors
can leave private buffers partial, and successful validation is not a security
claim. No constant-time property is provided.

The32-bit and64-bit matrix paths share the existing parallel Taylor/transpose/
row/column algorithm in matrix.hpp. Original32-bit wrappers retain their shape,
compact/disk modes, native clocks and admission rules. Full-domain buffer sizes
use the selected word width; no new memory-saving mode/default is selected.
Allwide plans prepare fixed-product tables as code-only setup. New native64-bit
clocks include one canonical-input range scan; stage clocks omit validation.
Go dispatch, allocation, setup, packing/I/O and future encoder costs remain extra.
This check cannot establish whole-encoding speed or revise existing RAM models.

## Bounded validation

The controller repeats scalar/recursive oracle checks and120 optimized/120
sanitized recursion fixtures from the prior experiment. New Go tests compare
wide affine plans with component-field polynomial evaluation at s8/16/17/18/20,
including top-bit ordered bases/offsets and metadata-copy ownership. Matrix tests
check all splits through256 points at1/4 workers, independent dense polynomial
evaluation, inverse, range/shape rejection, exact/partial buffer overlap,
error recovery, accounting and closed handles. Fixed seeds170053–170055.

At32768/65536 elements, compare matrix against the verified recursive transform
with splits1/8/dimension−1. This triggers the within-block parallel Taylor gather
branch rather than only coarse small blocks. Build a34-level s17 matrix plan
without its2^34 buffers and reject short arrays, checking large counts separately.
Go race checks cover plan use/Close on disjoint buffers; they do not instrument
C++ worker memory accesses. Native workers use the already validated shared
schedule, with ASan/UBSan checks and immutable plans/disjoint writes.

The direct C runner independently evaluates112 s17/s20 small matrix cases with
carryless long division, inverse/canaries, overlap/range/dependent-basis rejection
in optimized and ASan/UBSan binaries (nonrecovering undefined-behavior checks).
Run the existing32-bit native matrix sanitizer test and shared CPU Go suite too.
The binding package now has actual tests/examples, unlike the previous build-only
checkpoint. Profiled Go binaries are explicitly stored in the shared cache so
node-local scratch cleanup does not discard their matching executables.

One CPU node,4cores,8GiB RAM request/6GiB sampled RSS,4GiB Go soft limit,
2GiB retained cache/4GiB scratch. Go1.27.3, GCC14.2/C++17, no GPU/CUDA.
Worker10minutes, controller15minutes, Slurm20minutes; no budget increase.
Static C++/Python/shell syntax and source-format checks precede submission.
No existing jobs are polled and no post-submit monitoring occurs.

These checks establish bounded interface ownership and parallel matrix execution.
Full-domain workloads, integrated wide encoding and preprocessing feasibility
remain unvalidated.
Next calibrate wider kernel overhead and integrate wide code data/encoding
lifetimes before larger-run admission. Capacity remains projected at s≥17.

## First run and corrected retry

Job74312 passes the120 optimized/120 sanitizer recursive cases, all wide plan,
small matrix, parallel Taylor and example tests, and the Go lifetime race check.
The retained shared regression executable was then launched from the repository
root; unlike go test, direct test binaries do not automatically switch to their
package directory. Relative example fixture paths consequently failed, ending the
suite with exit2. This is a controller error, not a completed full regression.
The optimized/direct C matrix, matrix ASan/UBSan and legacy matrix sanitizer
stages were after that failure and **did not run**.

[Analyzer](analyze.py) reconciles420 source identities to71db6a7,25 evidence
files and7 retained cache files. It preserves the failed status and explicitly
marks unrun stages; forensic hash verification never promotes a failed job.
The scalar/recursive analyzers continue to enforce success by default.

```sh
python3 -B experiments/s17-matrix/analyze.py matrix-01
```

The controller now launches shared regressions from encoding/liftrs-shared in its
isolated source copy, with assertions for the expected package-relative fixtures.
No test selection, algorithm, resource limit or timing boundary changed. Retry
uses fresh `runs/matrix-02/` and `.cache/liftrs/s17-matrix/matrix-02/`, preserving
the failed attempt, profiles and sealed digests. Its shared-suite profile may be
partial because the example panic stopped the original test executable.

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/s17-matrix/slurm-%j.log \
  experiments/s17-matrix/run.sbatch matrix-02
```

Corrected retry submitted as **job74330**, source checkpointdc57f45;
[retry submission](retry-submission.json). No post-submit monitoring.
The corrected retry passed the complete bounded validation; the failed first
attempt is preserved. Neither run admits full s17 encoding or a full-encoding
numeric scaling-table result.

## Corrected validation result

Job74330 passes all recorded commands without resource-guard failures:

- Wide ordered/affine Go/C plans, independent dense matrix evaluations,1/4-worker
  inverse checks, range/overlap/error recovery, lifecycle and example tests.
- Parallel Taylor checks at32768/65536 values and34-level metadata counts without
  a full-domain allocation; Go lifetime race checks.
- All112 optimized and112 ASan/UBSan direct C matrix cases, and the existing32-bit
  matrix sanitizer tests and shared CPU regression suite.
- Repeated120 optimized/120 sanitizer recursive cases from the preceding stage.

The analyzer reconciled421 source files againstdc57f45,31 evidence files and all10
cached fixtures/binaries. [Binding checks](runs/matrix-02/binding-checks.log),
[race check](runs/matrix-02/lifetime-race.log),
[shared regressions](runs/matrix-02/shared-regressions.log),
[direct C cases](runs/matrix-02/matrix.log),
[sanitized C cases](runs/matrix-02/matrix-sanitizer.log),
[legacy matrix checks](runs/matrix-02/legacy-matrix.log).

```sh
python3 -B experiments/s17-matrix/analyze.py matrix-02
```

This validates bounded interface/parallel-transform correctness, not performance
at full s17 size. Larger kernel calibration, persistent wide code loading,
preprocessing and full encoding integration remain separate gates. Profiles,
matched executables and first-failure evidence are retained; no other jobs were
inspected during this reconciliation.
