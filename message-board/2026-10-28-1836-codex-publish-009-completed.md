# s=13 checkpoint and standing profiling policy verified for publication

- Timestamp: 2026-10-28 18:36 UTC
- Work item: PUBLISH-009
- Status: completed

The [s=13 evidence manifest](../experiments/cyclic-s13/evidence-manifest.json) and
passing [verification record](../experiments/cyclic-s13/verification.json) verify.
No encoder code changed since the completed run/full tests. README was checked
and updated. Added the user's standing policy to [AGENTS.md](../AGENTS.md),
[experiment conventions](../experiments/README.md#profiling-and-preprocessing-retention)
and the [encoder contract](../docs/encoder-contract.md): retain reasonably sized
reusable local caches and collect CPU profiles during optimization, document
exceptions, keep large data outside Git, and separate cache/instrumentation costs
from encoding. Disable profiling when optimization is complete.

`git diff --check` passes. Publish this and all completed ENCODE-009 results,
catalogue evidence/tests and documentation before starting the profiled s=13 run.
That next item must make large caches actually reloadable; current default JSON
loader caps are too small for the s=13 cache. It will keep existing default caps
and distinguish opt-in local-cache limits from public artifact distribution.
