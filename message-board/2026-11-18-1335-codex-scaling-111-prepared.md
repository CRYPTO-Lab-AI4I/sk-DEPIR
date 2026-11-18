# Explicit sorted membership comparison prepared

- Timestamp: 2026-11-18T13:35:31.191895+00:00
- Agent: Codex
- Work: SCALING-111
- Status: Prepared

[Experiment/report](../experiments/root-subset-validation/report.md) includes
runnable guarded controller, sbatch, reviewer and review tests. Shared validator
now selects bitmap or binary search only for membership storage; cardinality,
range, strict order, every q-Frobenius partner and progress remain common. Root
and partition helper variants take explicit mode and reuse the same arithmetic,
normalization/recovery and scheduling. All ordinary wrappers stay bitmap. CPU
client builds remain CUDA/MPI-free. No public API/default change or proof bypass.

Sorted membership avoids2GiB per invocation ats17(8GiB ats18projection), uses no
extra copied index or N-array, and adds O(KlogK) binary-search work. Actual latency
and RSS trade-offs unknown. No extra GPU transforms/transfers or canonical bytes;
existing K-index partition copy and runtime/context costs remain. Any later slower
production mode stays explicit; no fullsupport strategy adopted by equivalence.

Local4 tests pass0.107s: exhaustive65536subsets of16point domain compare acceptance
and error strings; malformed duplicates/range/order/missing partners reject;
wide>2^32 s17sorted input stays immutable, emits same progress and allocates<1MiB;
root partition/complete merge tests cover both modes and publication/admission.
Optional CPP+CUDA worker test compiles/skips absent explicit output env,0.004s.
Two reviewer tests reject reordered modes, changed oracle strategy and cross-mode
digest mismatch. Python AST, sbatch bash -n,runner --check,git diff --check pass.
GPU arithmetic/recovery/memcheck and performance are delegated to batch validation.

Opt-in test-binary flags -root-validation=bitmap|sorted and separately
-root-reference-validation retain bitmap defaults. New paired processes use
sorted CPU-reference validation equally in both modes to avoid N-sized reference
bitmaps contaminating RSS; this differs from historical77148 and is explicit
in report/README/plan. CPU root arithmetic/complete scalar witnesses stay
independent. Order bitmap/search/search/bitmap gives4timed samples per mode,
case and worker count, plus warm-ups; retain per-process CPU profiles and exact
binary/source/output hashes. Both modes get genuine small CUDA memcheck. Existing
complete CLI/native/reference/restart/race/ASan/vet/default-bitmap gates run too.

README checked/updated, as are actual scaling reference, active plan, preprocessing
internals and experiment index. Fullsize measurement cells unchanged. Existing
1node4H200/4CPU/8GiBhost/6GiBRSS/4GiBGo/4GiBowned/2GiBcache/4GiBscratch,
20min scheduler budget, no increase. New executors128MiB/card, at most1GiB summed
component caps. Commit source before submit, report job/results then yield without
monitoring. Sorted performance and process memory remain pending.
