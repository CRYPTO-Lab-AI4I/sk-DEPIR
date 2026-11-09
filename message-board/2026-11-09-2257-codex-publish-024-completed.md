# Count-width checkpoint verified for publication

- Timestamp: 2026-11-09 22:57:00 UTC
- Agent: Codex
- Work ID: PUBLISH-024
- Status: completed (publication checks; commit and push follow)

Reviewed the pending WIDTH-001 implementation, tests, documentation and
[count-width validation evidence](../experiments/count-width-validation/report.md)
for the user's requested commit/push. The analyzer again passes 231 measured
sources, all sealed evidence, retained executable/output and native clock checks.
All 26 current source hashes in the verification record match. Focused boundary,
width-selection and example tests pass; Go formatting and shell syntax pass.
Source-snapshot content hashes and recorded environment names were checked.

Publish the Go/native count and versioned-header changes, tests, reports, both
validation attempts and exact measured sources. Large ignored toolchains,
preprocessing, binaries and packed outputs remain outside Git. Root README checked:
it correctly states s-dependent stored widths, full s15 compatibility, small s16
checks and unverified larger-size feasibility. Full suites, sanitizer and race/vet
already passed on Slurm; matching runtime source hashes make another large run
unnecessary for publication. Staged paths/whitespace are checked before commit.

Added only this publication message pair and board entries. Use the existing
authenticated SSH-over-443 route for the same GitHub repository, with verified
host keys and ordinary fast-forward semantics. Next: audit s16 preprocessing
lifetimes, peak RAM, scratch and time; establish enforceable job budgets; then
enable the experimental preparation path and admit a guarded first full s16 run.
No full s16 or s≥17 support is claimed by this checkpoint.
