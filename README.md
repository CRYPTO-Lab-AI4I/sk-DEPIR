# sk-DEPIR: Lifted Reed-Solomon Secret-Key PIR

This repository implements the **Lift-RS online construction** from the accompanying [paper](paper.pdf), *Doubly-Efficient Secret-Key PIR with Low Storage Overhead*. It is being prepared as a readable, documented, and auditable open-source library. The paper also describes a Reed-Muller construction decoded along concatenated curves; that protocol is outside this repository's scope.

The client follows the permuted smooth-code blueprint: it queries the other `q-1` points on a degree-`t` curve through the requested location, using a secret permutation and masks. The Go implementation is `pir.LiftedRS`. `pir.NewPIR` remains as a compatibility alias for `pir.NewLiftedRS`.

Contributors and collaborating agents must follow [AGENTS.md](AGENTS.md), record every change in the [message board](message-board/README.md), and use the [experiment conventions](experiments/README.md) for new measurements. Check this README with every change.

The current development target is **TRL 5**: a usable, validated client and preprocessing library. The [roadmap](ROADMAP.md) gives the ordered tasks and acceptance criteria. Storage is treated as a plain I/O endpoint returning requested records; server-side protocol computation is outside the scope. TRL 6 scaling and advanced optimization work is deferred for a later decision.

## Implementation status and limits

The online query and decode path supports locality `ℓ=q-1`, dimensions `m=2` and `m=3`, prime fields, and the listed binary extension fields. `Params.Q`, `Params.K`, and `Params.M` denote the field size `q`, curve degree `t`, and database dimension `m`. Production `Query` and `QueryLocal` append 128 dummy positions, so the transmitted query length is `L=(q-1)+128`.

**A standalone small Lift-RS encoder is available in [encoding/liftrs](encoding/liftrs/README.md).** It provides reusable generator-matrix preprocessing and systematic field-element encoding for correctness fixtures over GF(4), GF(8), and GF(16), with at most 256 codeword positions. These are not the target security configurations. **The existing file-based API has no integrated Lift-RS encoder.** `LiftedRS.Encode` calls the external Reed-Muller encoder (`RMEncoding`) as a testing placeholder. `encoding/rme.cpp` contains an experimental `LRSE2D` helper, but the encoder command protocol does not select it. RM-backed tests therefore validate the Lift-RS online protocol only on compatible RM codewords. They do not establish Lift-RS encoding, rate, storage overhead, or encoding time.

A separate [small cryptographic reference](encoding/cryptoref/README.md) permutes codewords and applies storage-position XOR masks using the existing AES helpers. It tests independent masking keys, re-encoding, and genuine-codeword retrieval. Its in-memory architecture is a correctness reference; it does not select a production API or streaming design.

The [small application reference](encoding/appref/README.md) now executes manually supplied submatrix plans, pads and packs records, invokes both reference stages, and reconstructs through ordinary byte-range reads. It completes the small correctness pipeline; it does not choose a large-instance architecture or implement the automatic planner.

Other limits to account for when using or reviewing the code:

- General locality `ℓ` is not implemented. The decoder uses the full-field sum identity for `ℓ=q-1`; supporting other localities requires retaining curve parameters and interpolating over the field.
- The legacy Go client uses `math/rand` for curve sampling, dummy positions, and shuffling. A production security implementation needs a cryptographic randomness source.
- Local end-to-end tests replace the HTTP server with file-backed evaluation. `TestQueryFromServer` is a network smoke test; it skips when the configured server is unavailable and does not compare decoded values with an expected database.

## API

The Go module is `github.com/CRYPTO-Lab-AI4I/sk-DEPIR`. Import the client as
`github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pir`; preprocessing and configuration packages
use the same module prefix. Reed–Muller names identify only the retained testing
placeholder.

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

On macOS 26, Go 1.21.6 compiled this checkout but its generated test binary failed to load (`missing LC_UUID`). The current suite, including the reference encoder, was verified with Go 1.27.3 on that platform.

```bash
go build ./...
make -C encoding rme
go test ./...
```

The external encoder is needed by `BenchmarkEncode` and the tests whose names contain `WithRMEncoding`. The direct `GF(16)` online end-to-end test and online query/decode benchmarks do not need it. RM-backed tests write `input/db.csv`, `output/matrix.csv`, `input/example_db.csv`, and `output/example_matrix.csv`.

## Test coverage

| Test | Codeword | Server path | What it establishes |
|---|---|---|---|
| `TestLiftedRSWithRMEncodingEndToEnd` | Degree-2 RM codeword over `F_31`, used as a Lift-RS-compatible placeholder | Local CSV | Exhaustive query/decode check at all `q²` locations; does not test a Lift-RS encoder. |
| `TestLiftedRSWithRMEncodingEndToEndFromConfigKey` | Degree-3 RM codeword over `F_31`, with keys loaded from `config.json` | Local CSV | Same online check using configured keys; writes example CSV files. |
| `TestLiftedRSEndToEndWithGF2nCodeword` | Directly evaluated degree-2 polynomial over `GF(16)` | Temporary local CSV | Binary-field curve generation, permutation, masking, dummy handling, and decode; no offline encoder. |
| `TestReferenceLiftedRSEncoderOnlineRecovery` | Genuine systematic GF(16) Lift-RS codeword beyond degree-7 RM | In-memory reads | Generator-matrix encoding and online retrieval at every systematic position. |
| `TestCryptographicReferenceLiftedRSRecovery` | Genuine systematic GF(16) Lift-RS codewords | In-memory reads | XOR masking, shared permutation, independent database/re-encoding keys, and recovery with repeated curve positions. |
| `TestQueryFromServer` | Assumes a compatible database is served | HTTP | Network smoke test only; may skip and does not assert decoded values. |

Tests are in `planner/`, `cmd/pir-rate/`, `pirconfig/`, `pir/`, `utils/`, `encoding/liftrs/`, `encoding/cryptoref/`, and `encoding/appref/`. The [reference package tests](encoding/liftrs/reference_test.go) include exhaustive tiny-instance curve and message checks and a runnable API example. Run them from the repository root with `go test ./...`.

## Benchmarks

The benchmarks in [`pir/pir_test.go`](pir/pir_test.go) isolate client components; they are not end-to-end server measurements. Named Lift-RS query groups include curve generation, secret permutation, 128 dummy positions, and shuffle. Corresponding decode groups use mock responses and exclude query generation, server work, and the network. Older microbenchmarks remain and some explicitly use `noiseCount=0`; use the named Lift-RS groups to measure the current online path.

```bash
go test ./pir -run '^$' -bench 'Benchmark(QueryGenLiftedRS2D|DecodeLiftedRS2D|QueryGen3D_t5|Decode3D_t5|QueryGen3D_t6|Decode3D_t6)'
```

For encoding with the external RM placeholder:

```bash
go test ./pir -run '^$' -bench BenchmarkEncode
```

`BenchmarkEncode` includes file I/O, RM encoding, secret permutation, and masking. It is **not** a Lift-RS encoding benchmark. Database-size labels in query benchmarks are parameter-derived estimates that assume a near-rate-one Lift-RS storage regime; they are not measurements from the small reference encoder or a scalable Lift-RS backend. New performance claims belong in linked reports under [`experiments/`](experiments/).

## Field arithmetic and configuration

The agreed [data representation conventions](docs/data-representation.md) specify how an application matrix is partitioned into submatrices, each forming a PIR database with one field element per record and separate mask domains. The [encoder requirements](docs/encoder-contract.md) separate non-cryptographic codeword encoding from cryptographic permutation and masking. The format conventions use little-endian bit packing and a versioned JSON manifest. The small wrapper implements packed reads; durable manifests and the large-instance workflow remain to be implemented.

The [protocol I/O contract](docs/protocol-io.md) describes querying every PIR database through compatible query-sharing groups, storage responses, and client reconstruction. Query-sharing groups reuse client query generation; the client sends the positions separately for each PIR database, and communication counts every copy. Storage only serves the reads explicitly requested by the client.

The versioned [PIR configuration catalogue](pirconfig/README.md) and Go loader are
implemented. All 13 [configuration entries](pirconfig/configs_catalogue.json) have
[reproducible Monte Carlo rate estimates](experiments/liftrs-rates/report.md), with
confidence intervals and the four matching paper estimates retained. Target-size
encoding support is not claimed. A JSON Schema and consistency checks distinguish
estimates from verified capacities. The [plan evaluator](planner/README.md) validates
supplied partitions and calculates their logical storage/communication costs.
The [greedy planner](planner/README.md#greedy-planning) uses one-step lookahead to
construct compact row-block plans. It is deterministic and does not claim optimality.
The [cost-grid report](experiments/greedy-costs/report.md) covers 1 GB through 1 TB
with 1 KB through 10 MB records; these are model calculations, not encoded databases.

The [rate estimator](cmd/pir-rate/README.md) samples monomials using a deterministic
membership test that supports `s=2..20` without field arithmetic or generator
matrices. This does not extend the small encoder's supported sizes.

The PIR configuration catalogue includes binary extension fields `GF(2^s)` for every integer `s` from 8 through 20, initially with `m=2` and the [bounded-sample parameter profile](docs/parameter-planning.md#configuration-catalogue); prime fields are outside the current scope. Application records must be at least as wide as the largest PIR record in the supported catalogue; packing is outside the current scope.

The [planner design](docs/parameter-planning.md) describes joint selection of submatrix dimensions, code parameters, and query-sharing groups, including total storage and communication costs. Encoding-time and memory estimates may accompany plans but do not influence partition selection. The [small deterministic reference encoder](encoding/liftrs/README.md) now supplies the correctness baseline. The evaluator and greedy planner use catalogue evidence without running either encoder.

The [application encoder](docs/encoder-contract.md#application-encoding) consumes a selected plan, partitions and pads the input, converts it to field elements, and invokes the non-cryptographic and cryptographic PIR encoders for each database; [appref](encoding/appref/README.md) implements this workflow only as a bounded reference. [Code preprocessing](docs/encoder-contract.md#code-preprocessing) takes only code specifications and produces reusable encoding information, independently of any database. Durable manifest/key persistence and large-instance backends remain separate work.

- `utils/gf2n.go` implements `GF(2^n)` with log/exp tables for degrees 4, 8, 12, 13, 14, 16, 17, 18, and 20.
- Degrees 9, 10, 11, 15, and 19 remain to be implemented for the target range. Terabyte-scale application databases are a parameter-planning target, not a verified implementation capability.
- `utils/curve_generation.go` implements prime-field curve generation and `GF(2^n)` curve generation for the online Lift-RS path.
- `config.json` stores base64-encoded permutator and encryptor keys, a permutator tweak, and a server URL. The configuration helper fills in and persists missing key fields; `Gen` creates ephemeral in-memory keys.

## Project structure

```text
pirconfig/      JSON configuration catalogue, schema, Go loader, and validation tests
planner/        Greedy row-block partitioning, validation, and logical cost evaluation
cmd/pir-rate/   Monomial-sampling Lift-RS rate estimator
pir/            Go Lift-RS client API, tests, benchmarks, and RM encoder wrapper
utils/          Field arithmetic, curve generation, permutation, masks, and matrix I/O
encoding/       External RM placeholder and liftrs/, cryptoref/, appref/ reference backends
server/         Julia HTTP query handler
input/, output/ Example CSV data and generated matrix locations
experiments/    Scripts, reports, and results for new measurements
docs/           Construction and data representation specifications
message-board/  One file per agent message, with an index in README.md
```
