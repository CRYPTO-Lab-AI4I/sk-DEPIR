# CAT-001 — JSON catalogue implemented

- Timestamp: 2026-10-26 09:29:04 UTC
- Agent: Codex
- Status: completed

## Changes

- Added the [Go catalogue package](../pirconfig/catalogue.go), [strict JSON checks](../pirconfig/json.go),
  and [version-1 JSON Schema](../pirconfig/schema.json). Loading validates dimensions,
  rate/query arithmetic, polynomial irreducibility, evidence requirements, and support
  declarations without running preprocessing, encoding, or external scripts.
- Added all 13 [target configurations](../pirconfig/configs_catalogue.json), linked to the
  existing bounded-sample security calculation. Capacities remain unknown, field
  polynomials unselected, and encoding status candidate. No rounded paper capacity
  was promoted to an exact dimension and no experimental results were introduced.
- Added a [verified GF(4) fixture](../pirconfig/testdata/reference.json),
  [Go tests](../pirconfig/catalogue_test.go), [example](../pirconfig/example_test.go),
  and [schema tests](../pirconfig/schema_test.py). Monte Carlo metadata includes
  sample count, method, confidence, and rate interval; synthetic test metadata is
  explicitly not a measurement. Resource observations distinguish setup from encoding
  and do not affect selection.
- Documented the format, API, and remaining work in [catalogue/README.md](../pirconfig/README.md).
  Updated the [planner document](../docs/parameter-planning.md), [roadmap](../ROADMAP.md),
  and [agent repository map](../AGENTS.md). Checked and updated the root
  [README](../README.md) for the new package, status, and tests.

## Verification

Commands used Go 1.27.3 at `/tmp/sk-depir-go-toolchain/go/bin/go`, with
`GOCACHE=/tmp/sk-depir-go-cache`:

- `go test ./catalogue -count=1`: passed. The fixture dimension and polynomial are
  checked against actual reference preprocessing in tests, not during loading.
- `go test -race ./catalogue -count=1`: passed.
- `go vet ./catalogue`: passed.
- `go test ./... -count=1`: all six packages passed in isolated copy
  `/tmp/sk-depir-cat001.xUZFVF`, including a repeat with final code. Legacy tests
  write CSVs, so isolation preserved the checkout's input/output files.
- `go test ./catalogue -run '^$' -fuzz '^FuzzLoad$' -fuzztime=5s -parallel=2`:
  passed a short smoke run (four executions; not substantial fuzz coverage).
- `/tmp/sk-depir-catalogue-schema-env/bin/python catalogue/schema_test.py`:
  three tests passed, including schema self-validation and both catalogue documents.
  The temporary environment uses `jsonschema==4.25.1` on Python 3.9; initial sandbox
  networking and an unavailable requested package version were resolved before testing.
- `gofmt -l catalogue/*.go`, `git diff --check`, and local Markdown/catalogue evidence
  path checks: passed.

## Limits and next step

Validation establishes metadata consistency, not truth of evidence or backend
availability. Referenced code must still support a selected configuration before
encoding. Partition plans remain separate; cost evaluation, partition search, target
capacity estimation, and target-size encoding are not implemented by this change.
Next: implement layout cost evaluation and obtain reproducible capacity estimates
for planning. No commit or push was requested for this work item.
