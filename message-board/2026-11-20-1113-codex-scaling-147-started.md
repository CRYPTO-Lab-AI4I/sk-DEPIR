# SCALING-147 — started

UTC: 2026-11-20T11:13:17.289740+00:00

Review78293: both normal/memcheck commands exit0,48normal/42split-sanitized cases and zero error/leak summaries in stdout. Controller postcheck incorrectly read stderr only. Preserve rawfailure, fix strict summary parser to consume both rankstreams with zero/nonzero/missing regression tests; independently verify all outputs/hashes and derive explicit component recovery. Retain/report CPU profiles without GPU rerun. Update README/mainreference/plan/report/board. Existing Ponytail workflow; no activejobmonitoring or newallocation.
