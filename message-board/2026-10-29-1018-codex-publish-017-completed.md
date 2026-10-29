# s=14 checkpoint verified for publication

Timestamp: 2026-10-29 10:19 UTC. Agent: Codex. Work: PUBLISH-017. Status: completed.

Publish all SCALE-005 code, [report/results/profiles](../experiments/parallel-s14/report.md),
catalogue evidence, source snapshots and documentation. Both first-use and cached
analyzers pass source/binary/cache/evidence integrity checks; recorded admission
and full final validation are passed. `git diff --check` and local documentation
links pass. No tests or benchmarks were unnecessarily repeated.

README checked and its larger-field wording clarified. Added concrete
[s=15 admission steps](../docs/encoding-scaling.md#next-admission-s15): audit live
preparation arrays first, bound scratch only where needed, test the 30-bit extension
path, then run retained profiled first-use and fresh cached encoding under reviewed
budgets. The online 6–9 minute / 18–22 GiB ranges are planning estimates, not
measurements. First-use RSS cannot be assumed proportional or unavoidable.
Large caches remain ignored and retained locally. No larger run started here.
