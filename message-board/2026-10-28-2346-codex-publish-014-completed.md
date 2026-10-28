# Native command defaults and step-3 checkpoint verified for publication

- Timestamp: 2026-10-28 23:46 UTC
- Work item: PUBLISH-014
- Status: completed (publication checks)

Root [Makefile](../Makefile) standard commands now enable native transforms:
`make test` runs tagged suites; `make encode` builds native support and selects C++.
`BACKEND=go` / `make test-go` explicitly select Go, and `BACKEND=both` explicitly
requests a comparison. The [profiling runner](../experiments/native-encoding/run.py)
defaults to cpp; Go-only runs compile without the native tag/compiler requirement.
Native commands reject cgo-disabled execution early, and binaries are retained by
source hash and backend to avoid overwrite. The direct driver defaults to cpp and
requires a new profile destination. Encoder APIs, explicit WithCPPTransforms and
ordinary untagged Go behavior remain unchanged. Scaling algorithms are still opt-in.

Publishing the prior [step-3 probes](../experiments/scaling-prototypes/report.md):
sparse setup, paired preparation, minimal online state/workspaces and exact small
disk/Taylor/output mappings. Existing measured baselines/caches are preserved.
[Native-default and explicit-Go smoke results](../experiments/native-encoding/report.md#current-command-defaults)
show actual backend selection and matching original small hashes. Initial command
checks stopped before encoding because the new profile guard rejected runner log
files; partial results/sources are retained, and the guard now rejects preexisting
CPU profiles only. [Command verification](../experiments/scaling-prototypes/command-defaults-verification.json)
links the standard native/Go full-suite checks in isolated copies, cgo guard,
analyzers, tagged driver vet, Python syntax, links and whitespace: all pass.
No large encoding benchmark was rerun for the command change.

README checked and updated with commands/requirements and API distinction;
encoding guides, native report and encoder contract agree. Historical result
manifests are unchanged; exact measured pre-documentation sources are retained as
content-addressed snapshots. Large caches/profile binaries remain ignored/local.
Next: improve blocked I/O in the exact matrix/Taylor prototype, measure preparation
scratch and fresh minimal-loader RAM, then calibrate useful CPU/storage scaling.
Native generalized basis/offset execution is a separate implementation gate. Do
not claim s14/15 practicality or select hardware from the toy disk timings. This
message records pre-commit publication checks; commit/push outcome follows to user.
