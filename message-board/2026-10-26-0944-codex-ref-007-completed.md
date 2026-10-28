# REF-007 — Secret-key PIR module namespace established

- Timestamp: 2026-10-26 09:43:41 UTC
- Agent: Codex
- Status: completed

Replaced `module rme` in [go.mod](../go.mod) with
`github.com/CRYPTO-Lab-AI4I/sk-DEPIR`, matching the repository remote. Updated all
internal, test, and example imports under `pir/`, `encoding/appref/`,
`encoding/cryptoref/`, `encoding/liftrs/`, and `pirconfig/`. Grouped the imports in
[pir/db.go](../pir/db.go) and [pir/pir_test.go](../pir/pir_test.go) consistently with
standard-library/external-package grouping; protocol behavior is unchanged.

Updated the [root README](../README.md), [reference encoder usage](../encoding/liftrs-shared/README.md),
[PIR configuration usage](../pirconfig/README.md), and [roadmap](../ROADMAP.md).
The README check is complete. The module-path milestone is now separate from
outstanding dependency/licensing review. No dependency versions or sums changed.

Remaining `rme` names identify the actual external Reed–Muller placeholder:
`encoding/rme.cpp`, build targets and ignored binaries, and process discovery in
[RMEncoder](../pir/legacy_encoder_bridge.go). These remain accurate component names. Historical
board commands retain their original module paths.

Verification used Go 1.27.3 and `GOCACHE=/tmp/sk-depir-go-cache`:

- `go list -m` reports `github.com/CRYPTO-Lab-AI4I/sk-DEPIR`.
- `go test ./... -count=1` passed all six packages in isolated copy
  `/tmp/sk-depir-ref007.y53Shs`. Legacy CSV-writing tests did not alter the checkout.
  Subsequent changes only grouped imports in the two files named above.
- An external temporary module at `/tmp/sk-depir-import.xGmXXi`, with a local
  `replace` directive, passed `go test -mod=mod ./...`. Its test imported `pir`,
  `encoding/liftrs`, and `pirconfig` through the new path, verified the tiny reference
  dimension, and checked catalogue initialization. This verifies local external
  imports, not availability of an unreleased remote version.
- `gofmt -l`, `git diff --check`, updated documentation link checks, and obsolete
  namespace searches passed. Remaining nonhistorical `rme` occurrences are limited
  to the actual RM component.

No commit or push was requested for this work item. The earlier DOC-050 naming
instruction remains an independent uncommitted change; it was preserved.
