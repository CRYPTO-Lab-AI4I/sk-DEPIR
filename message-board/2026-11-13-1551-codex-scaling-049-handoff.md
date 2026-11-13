# s17 scalar port started; validation submitted

- Timestamp: 2026-11-13T15:51:51.544588+00:00
- Agent: Codex
- Work: SCALING-049
- Status: handoff

Expanded [the active execution plan](../docs/encoding-scaling.md#current-execution-plan-for-s17) with width/scalar audit, native CPU transforms, explicit cache/coordinate formats, independently verified preprocessing admission, resident CPU baseline, measured multi-GPU sharding/selection, full GPU comparison and progression toward s18. Updated the scaling reference, shared/package/experiment README links. Corrected obsolete no-s16-disk-submission wording to the already-approved retry report; no new budget increase.

Implemented private wideOrbitField in encoding/liftrs-shared/orbit_arithmetic64.go with uint64 packed extension values and unchanged uint32 base components/tower. No public size guard changed. New tests cover high34/40-bit coordinates, independent polynomial oracle, inverse/Frobenius identities, generator order witnesses and exact existing s≤16 ordering. [Audit/controller/report](../experiments/s17-arithmetic/report.md), [submission](../experiments/s17-arithmetic/submission.json).

Committed implementation checkpoint5b9e706. Submitted job74290,4CPU/noGPU/8GiB requested RAM/6GiB sampled RSS,2GiB retained cache/4GiB scratch,10m worker/15m controller/20m Slurm. Results/logs/profiles at experiments/s17-arithmetic/runs/scalar-01/, Slurm log experiments/s17-arithmetic/slurm-74290.log, binary .cache/liftrs/s17-arithmetic/scalar-01/. No full-domain data or preparation created; all earlier caches untouched.

Verification: Python syntax/controller --check, shell syntax, gofmt, documentation links and git diff --check passed. Full compilation/runtime checks pending in job74290; scalar passing would not establish native/full s17 support. README checked/updated. No running retry status/log/results inspection or post-submit monitoring; wait for user completion ping before dependent implementation/testing.
