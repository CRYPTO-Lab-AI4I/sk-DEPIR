# CAT-004 — PIR configuration naming and publication checkpoint

- Timestamp: 2026-10-26 09:38:00 UTC
- Agent: Codex
- Status: completed; ready for commit and push

Renamed the new package directory to [pirconfig/](../pirconfig/README.md), its Go
package/import to `pirconfig` / `rme/pirconfig`, and its data file to
[configs_catalogue.json](../pirconfig/configs_catalogue.json). The filename describes
its purpose; per-entry status still distinguishes candidates from encoding support.
Updated source package comments/error prefixes, examples, Go and Python tests,
[package documentation](../pirconfig/README.md), [planner documentation](../docs/parameter-planning.md),
[data representation wording](../docs/data-representation.md), [roadmap](../ROADMAP.md),
[agent repository map](../AGENTS.md), and the [root README](../README.md).
README verification is complete.

Retargeted earlier CAT message links to current paths while preserving their
historical prose and commands. CAT-001 through CAT-003 and their files were not yet
committed; this publication will include the implementation, paper evidence
corrections, and rename together. There is no previously published catalogue import
path to maintain.

Verification with Go 1.27.3 and `GOCACHE=/tmp/sk-depir-go-cache`:

- `go test ./pirconfig -count=1`, `go test -race ./pirconfig -count=1`, and
  `go vet ./pirconfig`: passed.
- `go test ./... -count=1`: all six packages passed in isolated checkout
  `/tmp/sk-depir-cat004.50XxSO`; the legacy CSV-writing tests did not alter this checkout.
- `/tmp/sk-depir-catalogue-schema-env/bin/python pirconfig/schema_test.py`:
  all three schema tests passed.
- `gofmt`, `git diff --check`, active Markdown and retargeted board link checks,
  and JSON evidence path checks: passed. A search found no obsolete catalogue
  paths/imports or data filename outside historical board commands.

The next implementation step is a layout cost evaluator: validate a supplied
partition plan against configuration evidence and calculate storage, upload,
download, and effective rate. Use exact tiny fixtures and provisional paper-backed
plans to test it. Reproducible Monte Carlo capacity estimates for the missing target
entries and partition search follow; unknown capacity must not become an executable
plan. No encoder, planning algorithm, or security profile changed in this rename.
