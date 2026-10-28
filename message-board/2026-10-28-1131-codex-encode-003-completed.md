# Encoder stage profiles recorded

- Timestamp: 2026-10-28 11:31 UTC
- Work item: ENCODE-003
- Status: completed

Published the prior table-free checkpoint and latest results as `d464fcf` on
`origin/main`, then added [the CPU-profile experiment](../experiments/encoder-profile/report.md)
with [driver](../experiments/encoder-profile/main.go), [script](../experiments/encoder-profile/run.py),
[results](../experiments/encoder-profile/results.json), four raw profiles and their
flat/cumulative text views. Reusable preparation and database encoding run in
separate processes with loading, input generation and serialization excluded.
Both prepared artifacts and both encoded words match the committed checkpoint.

At s=9 the cyclic path attributes approximately 75% of process samples to FFTs
and 18% to the Karatsuba prefix. Paper encoding attributes approximately 90% to
polynomial multiplication. These nested sample shares are not additive wall-time
measurements. Runtime/GC sampling and the short local runs limit interpretation;
a brief package build check overlapped the first process, recorded in the report.
No encoder algorithm or supported parameter changed.

Updated [the scaling plan](../docs/encoding-algorithms.md) with this evidence,
[the experiment index](../experiments/README.md), and [README](../README.md).
Next arithmetic work remains fast general/truncated polynomial multiplication;
field-kernel improvements should be measured separately. No GPU choice is needed.

Verification: `go test ./experiments/encoder-profile` builds successfully;
`go vet ./experiments/encoder-profile` and `git diff --check` pass. The complete
four-stage experiment succeeds, every systematic symbol is checked, all source,
input and profile hashes match, and report links resolve. The full protocol suite
was not rerun: this task adds a standalone profiling driver and documentation,
without changing library behavior. New profiling work is separate from the
already pushed checkpoint and is not yet committed.
