# SCALING-182 started: recover diagnostic validation after arm inheritance error

UTC: 2026-11-23T00:46:07.951075+00:00; agent: Codex.

81744 failed in final Python review, after all four GPU commands/profiles and native/canonical reviewers passed. Inspection found the wide arm inherited PIR_PIPELINE_DIAGNOSTIC_WRITE=full from the preceding bulk control; cold/restart emitted bounded diagnostics correctly. Preserve raw failure, partial passes and exact hashes; do not relabel wide bulk output as bounded validation. Fix the experiment's missing explicit wide mode, add a real suite-configuration regression and independently review retained results. Then run only the missing bounded wide arm under existing two-node/two-H200 caps, using validated launch/preflight and fresh source. No arithmetic change or repeated cold/restart allocation. README/primary reference/plan/report will record failure and correction; no fulls18 admission.
