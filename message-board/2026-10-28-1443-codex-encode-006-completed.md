# Current s=10 cyclic CPU profiles recorded

- Timestamp: 2026-10-28 14:43 UTC
- Work item: ENCODE-006
- Status: completed

Published the s=10 implementation, exact catalogue evidence, artifacts and report
as `c06375c` on `origin/main`, then completed the proposed profiling step.

The shared [profiling driver](../experiments/encoder-profile/main.go) now accepts
field/curve arguments, samples the correct field width and checks artifact/spec
agreement. Its s=9,t=7 defaults and output are preserved. Added the
[s=10 report](../experiments/encoder-profile-s10/report.md),
[script](../experiments/encoder-profile-s10/run.py),
[raw results](../experiments/encoder-profile-s10/results.json), two CPU profiles
and flat/cumulative/call-tree/source views. Preparation and encoding use different
processes; loading, message generation, serialization, verification and I/O are
outside sampling. Prepared JSON and full encoding hashes match the committed
checkpoint; every systematic symbol passes.

The completed run samples one preparation call (26.252 s interval) and four
encoding calls (10.644 s). Encoding source attribution: 76.5% of sampled process
CPU stacks in the two full-size generator-product FFTs, 23.0% in the prefix
product's smaller FFTs. Base-field multiplication accounts for 76.4% flat across
these stages. These are overlapping algorithm/arithmetic views, not additive
wall-time shares or new latency measurements. Reusable preparation likewise has
74.4% flat base-multiplication samples. No new hot-spot claim is extrapolated to
larger fields or an unimplemented optimization.

The first harness attempt selected an encoding-only source filter for preparation
and failed after generating that profile. Corrected the stage-specific source-view
filter and reran the complete experiment. The report records this harness issue;
only the completed run supplies the published results.

Updated [README](../README.md), [scaling plan](../docs/encoding-algorithms.md),
[experiment index](../experiments/README.md) and the
[historical profiler report](../experiments/encoder-profile/report.md).
README was checked; library support remains s<=10. Next optimize portable base-field
multiplication, validate all field/artifact/output conventions and benchmark full
s=9/s=10 encodings before s=11/12. CPU-specific/GPU work is not selected here.

Verification: `go test ./experiments/encoder-profile` builds successfully (no unit
tests in that harness), and `go vet ./experiments/encoder-profile` passes. The
actual default s=9 execution reproduces its historical checkpoint hash; actual
s=10 preparation/encoding match committed artifacts/words. All source/input/profile/
text-view hashes and updated file links verify; `git diff --check` passes. No full
protocol-suite rerun: only experiment harness/docs changed; committed library
full-suite/race evidence remains in ENCODE-005. Profiling work is uncommitted.
