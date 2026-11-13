# Native64-bit additive transform prototype

This implements the kernel part of step2 in the
[s17 execution plan](../../docs/encoding-scaling.md#current-execution-plan-for-s17),
after [scalar reference validation74290](../s17-arithmetic/report.md). Runtime
validation is submitted as **job74295**, source checkpoint6083f56;
[submission record](submission.json). No post-submit monitoring. No full s17 encoder, wide C ABI, matrix schedule or code
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
