# SCALE-008 running handoff

Timestamp: 2026-11-06T09:06:22.427029+00:00. Agent: Codex.

The requested checkpoint commit `3497bbd` is pushed to origin/main. Experimental
s=15 guards are now extended in [preparation](../encoding/liftrs-shared/scaling_prepare_cpp.go),
[native import](../encoding/liftrs-cpp/c_api.cpp) and
[online loading](../encoding/liftrs-shared/scaling_online.go); default public
DiscoverOrbit remains s≤13. New small-buffer field tests cover s=14/15 and loader
shape tests reject oversized headers before allocating.

[Admission validation](../experiments/parallel-s15/validation/verification.json)
passes focused race checks, vet and both full Go/native suites in an isolated copy.
[First-use smoke](../experiments/parallel-s15/smoke/analysis.json) and
[fresh cached smoke](../experiments/parallel-s15/smoke-cached/analysis.json) pass,
with identical canonical hashes, systematic/curve checks and packed readback.
The subsequent progress-label-only change is covered by both smoke runs; full
suites were not needlessly repeated. Python compilation, relative document links
and tracked diff whitespace also pass. New clock-call compile errors were fixed
before admission; no large failed attempt was launched.

The [large run](../experiments/parallel-s15/initial/results.json) is **running**, not
complete: worker PID 84546, unified exec session 2577, currently exact discovery.
It started on AC at 100%; sampled peak RSS about 9.3 GiB after 3.5 minutes.
Use [state](../experiments/parallel-s15/initial/state.json),
[phase log](../experiments/parallel-s15/initial/events.log) and
[power/RSS progress](../experiments/parallel-s15/initial/progress.jsonl).
Do not start a duplicate run or delete caches. Guards: 24 GiB Go soft target,
32 GiB sampled RSS, 20 GiB new cache, four active hours. Low battery pauses work,
then AC resumes it. Completed support/preparation persists; unfinished phases
are not process-resumable. Exact measured sources are snapshotted by the runner.

[finish.py](../experiments/parallel-s15/finish.py), unified exec session 70762,
waits for first use, verifies evidence, then launches one fresh cached process
only on success. Its output is [finish.log](../experiments/parallel-s15/finish.log),
status is [summary.json](../experiments/parallel-s15/summary.json). On completion
it creates summary.md and an indexed board results message; failures remain
incomplete and are not automatically retried. No need to launch cached use manually.
The supervisor/script hash is recorded in its status. Large cache stays outside
Git in `.cache/liftrs/parallel-s15/s15-m2-t5/`; old caches remain untouched.

README check: updated [root README](../README.md),
[scaling plan](../docs/encoding-scaling.md), language-specific READMEs and
[experiment index](../experiments/README.md), linking the
[method/report](../experiments/parallel-s15/report.md). Large results and support
claims remain pending. Admission code/harness/docs are local, uncommitted changes.
After outcome, review summary, profiles, phase clocks, memory and exact K, then
update current-status docs/catalogue evidence as warranted before publication.
