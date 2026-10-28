# CPU checkpoint documented and verified for publication

- Timestamp: 2026-10-28 22:35 UTC
- Work item: PUBLISH-012
- Status: completed (publication checks)

Updated [algorithm status and next step](../docs/encoding-algorithms.md#current-checkpoint-and-next-step),
[encoder contract](../docs/encoder-contract.md#cpu-checkpoint-and-scaling-direction),
[roadmap](../ROADMAP.md) and [README](../README.md). The current CPU kernel phase
is complete through validated s=13; further CPU optimization remains possible.
Unimplemented lookup/unrolling/parallel variants are deferred, not reported as
completed. The next activity is the scaling design study: account for arrays and
artifacts at s=16..20, examine blocked/streaming/external-memory algorithms,
compare commodity CPU/GPU designs and prototype uncertain memory/I/O operations.
No architecture is selected by this documentation checkpoint.

s=14 is an unmeasured practical expectation on this 48 GiB host with a larger
setup budget (roughly 12–13 minute encoding, 10 GiB runtime RAM; 3-hour/23 GiB
preprocessing). s=15 is borderline for cached encoding only (55–63 minutes,
41 GiB); its 17-hour/97 GiB preprocessing estimate does not fit the host.
Formulas, measured anchors and forecast limitations are explicit. Neither size
is claimed supported, and old 12 GiB/two-hour guards are not silently increased.
TRL 5 remains incomplete. Profiling/cache retention continues during the broader
scaling/optimization work.

Publishing the prior [native integration](../experiments/native-encoding/report.md)
and [directory separation](../experiments/encoding-layout/report.md) with their
implementation, API/docs, historical source snapshots, results and verification.
Both complete test suites and tagged vet already passed after the last code
changes; no new timing run or redundant full-suite rerun was needed for this docs
update. Re-ran native s=13 and Go/C++ kernel evidence analyzers, checked the layout
source manifest, current checkpoint links and `git diff --check`: all pass.
README checked and updated. Caches/profiles/large local binaries remain retained
outside Git; legacy local binaries remain ignored. This message records checks
before commit/push; publication outcome and commit ID are reported to the user.
