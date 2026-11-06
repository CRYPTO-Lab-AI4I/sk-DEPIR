# SCALE-007 handoff: AC power pending

Timestamp: 2026-11-05T22:37:38+00:00. Agent: Codex.

Implemented and documented compact support/discovery membership, bounded flat
factor batches with balanced merges, preparation-only FFT workspace reuse,
full-length Newton-result reuse and opt-in diagnostic heap checkpoints. Changes
are local and do not widen s=15 guards or alter the information set/format.

Focused correctness/race tests and both full isolated Go/native suites pass.
[Smoke runs](../experiments/preparation-checkpoint/report.md) verify complete
P/G/spectrum against the reference, including active buffer ownership and scratch
release. The final guarded smoke also passes after the profile-flushing signal
handler was added. Source/binary/evidence audits, Python compilation, local links
and whitespace checks pass. Original measured script versions are preserved in
content-addressed source snapshots; no old result or useful cache was deleted.

The large s=14 test has **not started**: its admission check found battery power,
then a separate query confirmed 17%. An asynchronous user question to reconnect
AC remains pending. No large computation/background benchmark is active. All
independent code, tests and reporting work is ready; current docs explicitly mark
the large measurement pending. The AC requirement is the experiment's resource
policy, not an approval requirement from AGENTS.md.

Resume: verify actual AC, then run
`GOCACHE=/tmp/sk-depir-go-cache python3 experiments/preparation-checkpoint/run.py --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name s14`.
It reuses the original pinned s=14 basis; it must verify full online/G digests,
retain CPU/heap profiles and report diagnostic GC overhead separately. Run the
analyzer, update measured results/admission recommendation and post completion.
Do not claim s=14 large-constant validation or a new RSS bound before that run.

README check: root/shared encoding README, scaling plan, experiment index and
historical audit follow-up now describe the implemented memory path and pending
checkpoint. Generated raw profiler whitespace is narrowly scoped in attributes.
