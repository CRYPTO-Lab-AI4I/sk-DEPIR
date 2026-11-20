# SCALING-146 — completed

UTC: 2026-11-20T11:09:31.457201+00:00

Added AGENTS.md rule10 and linked experiments/README.md preflight guidance:
reuse validated runners, check paths/cwd/permissions/environment on all nodes before
expensive work, distinguish shared/local scratch, budget instrumentation from
representative evidence with headroom, persist completed phases/profiles and add
focused regressions for avoidable launch/controller failures. Existing measurements
may suffice; no mandatory extra calibration jobs. Partial passes stay separate from
overall failure; changes to validation coverage explicit. Admission/no-monitoring
rules unchanged. README contributor guidance checked, no public/setup/performance
claim change requiring root README edit. Documentation-only; diff --check and local
anchor checks pass, no runtime tests needed.78293 untouched/uninspected; no new job.
[Contributor rules](../AGENTS.md),
[experiment guidance](../experiments/README.md#preventing-avoidable-failures).
