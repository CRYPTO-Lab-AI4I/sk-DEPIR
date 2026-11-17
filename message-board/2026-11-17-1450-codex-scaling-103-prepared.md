# Bounded tiled-transform and split-product validation prepared

- Timestamp: 2026-11-17T14:50:57.990101+00:00
- Agent: Codex
- Work: SCALING-103
- Status: Prepared

Implemented diagnostic host-backed storage and split orchestration in
[CUDA headers](../encoding/liftrs-cuda/README.md#experimental-preprocessing-components),
reusing resident field/plan/admission and shared Taylor/vector/product arithmetic.
The resident source changes only to grant the adapter access; no ABI/default or
ordinary CPU build dependency changes. Added opt-in independent native CPU
fixtures, scalar/schoolbook witnesses, full-output checks/hashes, wide coordinates,
recovery/admission checks, and CUDA memcheck in the bounded component driver.

[Report/script/model](../experiments/preprocessing-tiles/report.md) describes
exact measurement boundaries, profile/log overhead, memory/time/transfer costs
and limits. Host tiling retains the full host array plus bounded staging; splitting
adds fourteen products and up to64GiB-minus4bytes projected host temporary while
reducing fulls18 final-product device arrays1536→384GiB. No complete fullsize
memory/runtime or feasibility claim. Earlier discovery/root/reciprocal gates
remain, and no slower mode becomes default.

Verification: two Python review tests pass; AST checks on all new Python scripts,
runner --check, sbatch bash -n, integer model reproducibility cmp and git diff
--check pass. CPP-tagged TestPreprocessingTileFixtures compiles/passes in0.004s
(default skip without explicit fixture destination). CUDA compilation, fixtures,
actual arithmetic equality, gprof output and GPU memcheck are deferred to the
batch job, not claimed passed. Existing guarded CLI regressions run there too.

README checked and updated with prototype scope. Updated active scaling plan,
main encoding-status reference, CUDA README and experiment index; completed s17
CPU/GPU results and untested-mode cells stay accurate. Reuse approved4H200,
4CPU/8GiBhost,6GiBRSS,4GiB aggregate owned device ceiling,2GiBretained/4GiBscratch,
20min scheduler budget. New component engines cap128MiB/card. Submit then hand
off without monitoring; exact revision and source hashes are recorded by runner.
