# REF-001: Lift-RS specialization completed

**Timestamp:** 2026-10-21 09:31 UTC
**Agent:** Codex
**Status:** Completed

Removed the concatenated-curve RM client (`pir/pir-conc.go`) and its exclusive `GF(q²)` arithmetic and tests (`utils/gf2n_ext*.go`, `utils/curve_generation_test.go`). Removed RM-Conc-only fields and degree validation from the [PIR API](../pir/pir_api.go), [parameters](../pir/params.go), and [RM encoder wrapper](../pir/rm_encoder.go), plus dedicated cases in the [API tests](../pir/pir_api_test.go) and [benchmarks](../pir/pir_test.go). The [README](../README.md) and [agent instructions](../AGENTS.md) now describe the Lift-RS-only scope.

Kept the RM encoder, including its 2D and 3D implementations, because `LiftedRS.Encode` still uses it as a testing placeholder. Also kept base-field arithmetic, curve generation, permutation, masking, local and HTTP query handling, and Lift-RS tests. The README continues to state that Lift-RS encoding and production security are not established.

Verification: `make -C encoding rme`, `go test ./...`, `go build ./...`, Markdown link checks, and `git diff --check` passed. Tests ran with Go 1.27.1. Go 1.21.6 compiled but macOS 26 rejected its generated `pir.test` binary with `missing LC_UUID`. The tests rewrote two tracked example CSVs, which were restored to their pre-test state. No experiment was run or added. Future work: implement a genuine Lift-RS encoder and replace experimental randomness before making stronger claims.
