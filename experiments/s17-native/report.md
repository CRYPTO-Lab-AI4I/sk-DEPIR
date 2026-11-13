# Native64-bit additive transform prototype

This implements the kernel part of step2 in the
[s17 execution plan](../../docs/encoding-scaling.md#current-execution-plan-for-s17),
after [scalar reference validation74290](../s17-arithmetic/report.md). Runtime
validation **passed job74295**, source checkpoint6083f56;
[submission record](submission.json), [results](runs/kernel-01/results.json),
[reconciled analysis](runs/kernel-01/analysis.json). Results were inspected only
after the user reported completion. No full s17 encoder, wide C ABI, matrix schedule or code
cache is admitted by this work.

[Controller](run.py), [Slurm script](run.sbatch),
[native kernel](../../encoding/liftrs-cpp/transform.hpp),
[fixture generator](../../encoding/liftrs-shared/wide_native_fixture_test.go),
[native fixture check](check.cpp). Results/commands/exact revision/source hashes,
CPU profiles and environment are retained at `runs/kernel-01/`. Binaries and
trusted fixture are retained outside Git at `.cache/liftrs/s17-native/kernel-01/`.
Existing preprocessing and active-job files are neither read nor altered.

```sh
python3 -B experiments/s17-native/run.py --check
module load slurm/slurm/25.05
sbatch --output=experiments/s17-native/slurm-%j.log \
  experiments/s17-native/run.sbatch kernel-01
```

## Implementation and invariants

FieldT/ProductT/TermT/TransformT specialize the same native recursive kernel over
uint32_t or uint64_t. Existing Field/Product/Term/Transform aliases remain32-bit;
the C ABI and its s≤16 admission/word descriptions stay unchanged. Base-field
components, polynomial and delta remain32-bit. Cast the high extension component
to the selected Word before shifting; select64-bit trailing-zero scans for wide
offsets. Coordinates, linearized degrees, products/tables and mutable buffers use
the selected type. Full-cardinality storage and host addresses remain64-bit.

The64-bit specialization currently consumes trusted bounded experiment metadata,
not public cache data. Calls borrow caller-owned buffers and allocate no transform
scratch; plan metadata/tables are built separately. As before, direct kernel
callers must provide canonical elements and valid immutable plans. No claim of
constant-time or cryptographic behavior. The ordinary32-bit path has no new
allocation, data-layout or memory/time mode selected; regressions are tested.

## Bounded correctness methodology

Generate60 trusted fixtures at s8/16/17/18/20, with8/128/1024 coefficients and
canonical or ordered bases. Ordered bases include the highest extension bit;
nonzero offsets include high bits beyond32. All cases are extension-field plans.
Sparse subspace-polynomial setup allocates only bounded per-level metadata.
Inputs are dense through128 coefficients; at1024 they have16 low coefficients
plus a nonzero highest-degree coefficient to exercise all division levels.
Expected evaluations use independent Go carryless multiplication/reduction,
Horner evaluation and sparse high-degree exponentiation, not native recursion. Seed170051 is fixed. Plans and arrays never span a full domain.

Each fixture runs with scalar and four-bit fixed-product maps:120 cases each in
an optimized binary and ASan/UBSan binary. Compare every forward result with
independent evaluation, inverse round trips and caller-buffer end canaries.
At s≤16, compare the32-bit and64-bit specializations on identical inputs/metadata.
Also run the existing independent native scalar oracle and complete shared/native
Go tests with liftrs_cpp. This is bounded regression coverage, not another full
cached s16 encoding. Pure-Go field/scalar tests and fixture-generation CPU profiles
are retained; native nanosecond clocks are diagnostics, not stable benchmarks.

Use GCC14.2/C++17/Go1.27.3 on one CPU node,4cores,8GiB requested RAM/6GiB sampled
process RSS,4GiB Go soft memory,2GiB new retained cache and4GiB scratch. Worker
10minutes, controller15minutes, Slurm20minutes; resource guards stay active.
No GPU, communication test or budget increase. Build/run correctness inside the
job; do not monitor before the user's completion ping.

Passing would establish bounded native forward/inverse arithmetic, not the
34-level full transform schedule, parallel matrix kernels, full encoding or
preprocessing. Next integrate wide plans with explicit interface/ownership/range
checks, measure native overhead, then extend reusable code formats and preparation.
The maintained capacity/runtime table remains projected at s≥17.

## Reconciled results

[Analyzer](analyze.py), [optimized output](runs/kernel-01/native.log),
[sanitizer output](runs/kernel-01/sanitizer.log),
[shared CPU suite](runs/kernel-01/native-regressions.log),
[binding build](runs/kernel-01/binding-regressions.log).

All120 optimized cases and120 ASan/UBSan cases pass. Every fixture matches the
independent polynomial evaluation, inverse round trip and canaries; overlapping
s≤16 cases agree between32-bit and64-bit specializations. Invalid lengths/offsets
are rejected. The legacy native scalar oracle and shared liftrs_cpp Go suite pass.
The binding package builds and reports no test files; this is not a separate
binding test pass. The profiled shared-suite executable created by go test is
in the compute node’s isolated `/tmp/sk-depir-native-wide-74295/repo/` tree,
rather than the sealed shared cache. Its retention depends on node scratch
lifetime; the five sealed cache files cover the initial Go binary, native
binaries and fixture. All profile files themselves are archived. No recorded command failed or triggered a guard.

Source reconciliation verified410 files against6083f56,20 evidence files and all5
retained cache files (fixture and binaries). The reusable scalar hash verifier is
shared with the earlier experiment; rerunning its reconciliation still passes.

```sh
python3 -B experiments/s17-native/analyze.py kernel-01
```

The kernel gate passes for these bounded plans, not a complete34-level transform
or full encoder. No full-domain RAM/latency measurements or speedup claims follow.
Next implement explicit wide interface/plan ownership and parallel matrix
execution, compare against the verified recursive/scalar kernels, then widen
code-data formats and preprocessing. Existing CPU/GPU admission remains s≤16.
