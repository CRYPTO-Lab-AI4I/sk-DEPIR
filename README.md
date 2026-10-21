# sk-DEPIR: Lifted Reed-Solomon Secret-Key PIR

This repository implements the **Lift-RS online construction** from the accompanying [paper](sk_DEPIR.pdf), *Doubly-Efficient Secret-Key PIR with Low Storage Overhead*. It is being prepared as a readable, documented, and auditable open-source library. The paper also describes a Reed-Muller construction decoded along concatenated curves; that protocol is outside this repository's scope.

The client follows the permuted smooth-code blueprint: it queries the other `q-1` points on a degree-`t` curve through the requested location, using a secret permutation and masks. The Go implementation is `pir.LiftedRS`. `pir.NewPIR` remains as a compatibility alias for `pir.NewLiftedRS`.

Contributors and collaborating agents must follow [AGENTS.md](AGENTS.md), record every change in the [message board](message-board/README.md), and use the [experiment conventions](experiments/README.md) for new measurements. Check this README with every change.

## Implementation status and limits

The online query and decode path supports locality `ℓ=q-1`, dimensions `m=2` and `m=3`, prime fields, and the listed binary extension fields. `Params.Q`, `Params.K`, and `Params.M` denote the field size `q`, curve degree `t`, and database dimension `m`. Production `Query` and `QueryLocal` append 128 dummy positions, so the transmitted query length is `L=(q-1)+128`.

**There is no integrated Lift-RS encoder.** `LiftedRS.Encode` calls the external Reed-Muller encoder (`RMEncoding`) as a testing placeholder. `encoding/rme.cpp` contains an experimental `LRSE2D` helper, but the encoder command protocol does not select it. Encoder-backed tests therefore validate the Lift-RS online protocol only on compatible RM codewords. They do not establish Lift-RS encoding, rate, storage overhead, or encoding time.

Other limits to account for when using or reviewing the code:

- General locality `ℓ` is not implemented. The decoder uses the full-field sum identity for `ℓ=q-1`; supporting other localities requires retaining curve parameters and interpolating over the field.
- Curve sampling, dummy positions, and shuffling use Go's `math/rand`. A production security implementation needs a cryptographic randomness source.
- Local end-to-end tests replace the HTTP server with file-backed evaluation. `TestQueryFromServer` is a network smoke test; it skips when the configured server is unavailable and does not compare decoded values with an expected database.

## API

`LiftedRS` satisfies the `PIR` interface in [`pir/pir_api.go`](pir/pir_api.go):

```go
type PIR interface {
    Gen()
    Encode(input, output string) (Matrix, error)
    Query(index int, serverURL string) QueryResult
    Decode(QueryResult) ([]int, error)
    Close() error
}
```

Use `NewLiftedRS(params)` for new code. `Gen` creates ephemeral keys; `GenFromConfig` loads keys from `config.json`. Call `Close` when an encoder process has been started. `QueryLocal` is available for local tests and benchmarks.

## Requirements and build

- Go 1.21.6
- A C++17 compiler for the preferred external RM encoder, or Julia 1.10+ for its fallback. Julia dependencies are recorded in `encoding/Project.toml` and `encoding/Manifest.toml`.

On macOS 26, Go 1.21.6 compiled this checkout but its generated test binary failed to load (`missing LC_UUID`). The current test suite was verified with Go 1.27.1 on that platform.

```bash
go build ./...
make -C encoding rme
go test ./...
```

The external encoder is needed by `BenchmarkEncode` and the tests whose names contain `WithRMEncoding`. The direct `GF(16)` online end-to-end test and online query/decode benchmarks do not need it. Encoder-backed tests write `input/db.csv`, `output/matrix.csv`, `input/example_db.csv`, and `output/example_matrix.csv`.

## Test coverage

| Test | Codeword | Server path | What it establishes |
|---|---|---|---|
| `TestLiftedRSWithRMEncodingEndToEnd` | Degree-2 RM codeword over `F_31`, used as a Lift-RS-compatible placeholder | Local CSV | Exhaustive query/decode check at all `q²` locations; does not test a Lift-RS encoder. |
| `TestLiftedRSWithRMEncodingEndToEndFromConfigKey` | Degree-3 RM codeword over `F_31`, with keys loaded from `config.json` | Local CSV | Same online check using configured keys; writes example CSV files. |
| `TestLiftedRSEndToEndWithGF2nCodeword` | Directly evaluated degree-2 polynomial over `GF(16)` | Temporary local CSV | Binary-field curve generation, permutation, masking, dummy handling, and decode; no offline encoder. |
| `TestQueryFromServer` | Assumes a compatible database is served | HTTP | Network smoke test only; may skip and does not assert decoded values. |

Tests are in `pir/` and `utils/`. Run them from the repository root with `go test ./...`.

## Benchmarks

The benchmarks in [`pir/pir_test.go`](pir/pir_test.go) isolate client components; they are not end-to-end server measurements. Named Lift-RS query groups include curve generation, secret permutation, 128 dummy positions, and shuffle. Corresponding decode groups use mock responses and exclude query generation, server work, and the network. Older microbenchmarks remain and some explicitly use `noiseCount=0`; use the named Lift-RS groups to measure the current online path.

```bash
go test ./pir -run '^$' -bench 'Benchmark(QueryGenLiftedRS2D|DecodeLiftedRS2D|QueryGen3D_t5|Decode3D_t5|QueryGen3D_t6|Decode3D_t6)'
```

For encoding with the external RM placeholder:

```bash
go test ./pir -run '^$' -bench BenchmarkEncode
```

`BenchmarkEncode` includes file I/O, RM encoding, secret permutation, and masking. It is **not** a Lift-RS encoding benchmark. Database-size labels in query benchmarks are parameter-derived estimates that assume a near-rate-one Lift-RS storage regime; they are not measurements from an implemented Lift-RS encoder. New performance claims belong in linked reports under [`experiments/`](experiments/).

## Field arithmetic and configuration

- `utils/gf2n.go` implements `GF(2^n)` with log/exp tables for degrees 4, 8, 12, 13, 14, 16, 17, 18, and 20.
- `utils/curve_generation.go` implements prime-field curve generation and `GF(2^n)` curve generation for the online Lift-RS path.
- `config.json` stores base64-encoded permutator and encryptor keys, a permutator tweak, and a server URL. The configuration helper fills in and persists missing key fields; `Gen` creates ephemeral in-memory keys.

## Project structure

```text
pir/            Go Lift-RS client API, tests, benchmarks, and RM encoder wrapper
utils/          Field arithmetic, curve generation, permutation, masks, and matrix I/O
encoding/       External RM encoder (C++ and Julia), currently a Lift-RS testing placeholder
server/         Julia HTTP query handler
input/, output/ Example CSV data and generated matrix locations
experiments/    Scripts, reports, and results for new measurements
message-board/  One file per agent message, with an index in README.md
```
