# sk-DEPIR: Lifted Reed-Solomon Secret-Key PIR

This repository implements the **Lift-RS online construction** from the accompanying [paper](paper.pdf), *Doubly-Efficient Secret-Key PIR with Low Storage Overhead*. It is being prepared as a readable, documented, and auditable open-source library. The paper also describes a Reed-Muller construction decoded along concatenated curves; that protocol is outside this repository's scope.

The client follows the permuted smooth-code blueprint: it queries the other `q-1` points on a degree-`t` curve through the requested location, using a secret permutation and masks. The Go implementation is `pir.LiftedRS`. `pir.NewPIR` remains as a compatibility alias for `pir.NewLiftedRS`.

Contributors and collaborating agents must follow [AGENTS.md](AGENTS.md), record every change in the [message board](message-board/README.md), and use the [experiment conventions](experiments/README.md) for new measurements. Check this README with every change. During optimization,
[retain preprocessing and CPU profiles](experiments/README.md#profiling-and-preprocessing-retention)
by default; keep large local caches outside Git and report instrumentation overhead.

The current development target is **TRL 5**: a usable, validated client and preprocessing library. The [roadmap](ROADMAP.md) gives the ordered tasks and acceptance criteria. Storage is treated as a plain I/O endpoint returning requested records; server-side protocol computation is outside the scope. TRL 6 acceptance is deferred; evaluating encoding-efficiency improvements remains in scope.

Fast encoding is an explicit priority. Language, compiler, algorithms, hardware,
layout and parallelism remain open choices; select improvements through measured
comparisons while preserving PIR functionality, correctness and usability on
standard setups. Target commodity CPUs and GPUs and measure on locally available
hardware. Larger workloads may need larger machines, but hardware/software
requirements should remain as generic as practical.

## Implementation status and limits

The online query and decode path supports locality `ℓ=q-1`, dimensions `m=2` and `m=3`, prime fields, and the listed binary extension fields. `Params.Q`, `Params.K`, and `Params.M` denote the field size `q`, curve degree `t`, and database dimension `m`. Production `Query` and `QueryLocal` append 128 dummy positions, so the transmitted query length is `L=(q-1)+128`.

**A standalone small Lift-RS encoder is available in [encoding/liftrs-go](encoding/liftrs-go/README.md).** It provides reusable generator-matrix preprocessing and systematic field-element encoding for correctness fixtures over GF(4), GF(8), and GF(16), with at most 256 codeword positions. These are not the target security configurations. **The existing file-based API has no integrated Lift-RS encoder.** `LiftedRS.Encode` calls the external Reed-Muller encoder (`RMEncoding`) as a testing placeholder. `encoding/rme.cpp` contains an experimental `LRSE2D` helper, but the encoder command protocol does not select it. RM-backed tests therefore validate the Lift-RS online protocol only on compatible RM codewords. They do not establish Lift-RS encoding, rate, storage overhead, or encoding time.

**Paper and cyclic encoders now share the paper's information set** through the
[bounded orbit API](encoding/liftrs-go/README.md#paper-and-cyclic-encoders) for `m=2`.
In-memory experiments accept `s=2..13` (at most 67,108,864 positions); completed
catalogue benchmarks cover s=8..13. Discovery, method preparation,
public artifact loading, and database encoding are separate. Prepared artifacts
include cached solver data or cyclic polynomials/spectra, reused for many databases.
The [CPU comparison](experiments/encoder-comparison/report.md) reports setup, artifact
sizes, per-database encoding, process memory and output I/O separately. This does
not provide validated s=14..20, streaming, GPU or integrated file-based encoding.
The [table-free checkpoint](experiments/tablefree-encoding/report.md) now measures
s=9,m=2,t=7 with exact K=17200. The standalone field API covers every s=2..20
without full-field arithmetic tables; encoder support remains separately bounded.
[Stage-specific CPU profiles](experiments/encoder-profile/report.md) identify
FFT arithmetic and polynomial multiplication costs separately; these are
diagnostics, not new latency benchmarks.
The [polynomial-product comparison](experiments/fast-polynomial/report.md) measures
the hybrid FFT/Karatsuba backend against the preceding implementation, separately
for preprocessing and database encoding; codewords and artifacts are unchanged.
The [s=10 CPU checkpoint](experiments/encoder-scaling/report.md) adds exact
K=103556 at m=2,t=6, paper/cyclic agreement, independent sampled-curve checks,
separate encoding heap sampling and packed output measurements. JSON artifacts
remain in use.
The [current s=10 CPU profile](experiments/encoder-profile-s10/report.md) identifies
portable base-field multiplication inside the FFTs as an optimization target.
The [fixed-constant multiplication comparison](experiments/field-multiplication/report.md)
adds bounded reusable FFT tables while preserving the table-free general field API
and existing artifact format.
The [progressive cyclic benchmarks](experiments/cyclic-scaling/report.md) first
extend measurements through s=12: all measured cases fit the agreed one-hour
first-use and 12 GiB forecast policy. Large prepared caches stay in memory; JSON
size limits are unchanged and large-cache loading was not measured.
The [s=13 benchmark](experiments/cyclic-s13/report.md), run with a two-hour
allowance, adds exact K=11926435 at t=5: 29.42 min reusable preparation,
185.15 s median encoding and 5.34 GiB whole-case peak RSS. First use is 33.94 min;
application integration remains unmeasured. The subsequent
[profiled s=13 runs](experiments/encoder-profile-s13/report.md) retain a 432 MiB
local cache, reload it in 4.83 s and encode in 186.35 s, with identical output.
CPU profiles identify fixed-constant multiplication as the largest remaining
self-time share. Profiled/cache measurements are separate from this unprofiled
baseline; large local files stay outside Git.

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

Tests are in `planner/`, `cmd/pir-rate/`, `pirconfig/`, `pir/`, `utils/`, `encoding/liftrs-go/`, `encoding/cryptoref/`, and `encoding/appref/`. The [reference package tests](encoding/liftrs-go/reference_test.go) include exhaustive tiny-instance curve and message checks and a runnable API example. Run them from the repository root with `go test ./...`.

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

[Optimizing upload through storage layout](docs/protocol-io.md#storage-layout-optimization)
is now in scope for TRL 5: place each query-sharing group's same-position values
together so one ordinary read address retrieves them all. A grouped layout,
streaming writer and client adapter remain to be implemented and verified.
The current evaluator still counts repeated uploads. The [tradeoff tables](experiments/tradeoff-costs/options.md)
and [DP tables](experiments/dp-costs/options.md) now display inverse download rate
`D/B` and absolute download, including `D/B` for the reference configuration;
the searches now optimize storage versus download while retaining separate upload
and total-traffic accounting. The rerun comparison is in the
[objective-change report](experiments/download-objective-comparison/report.md).

The versioned [PIR configuration catalogue](pirconfig/README.md) and Go loader are
implemented. All 13 [configuration entries](pirconfig/configs_catalogue.json) have
[reproducible Monte Carlo rate estimates](experiments/liftrs-rates/report.md), with
confidence intervals and the four matching paper estimates retained. Target-size
encoding support is not claimed. A JSON Schema and consistency checks distinguish
estimates from verified capacities. The [plan evaluator](planner/README.md) validates
supplied partitions and calculates their logical storage/communication costs.
The [greedy planner](planner/README.md#greedy-planning) uses one-step lookahead to
construct compact row-block plans. `GreedyApplication` additionally searches split/stack
logical widths and optional small full-download suffixes bounded by a configurable
download ratio (default 1). These are planning APIs, not new encoder backends.
Selection is deterministic and does not claim optimality.
[`SearchTradeoffs`](planner/README.md#storage-and-communication-tradeoffs) returns
bounded storage/communication alternatives, including whole-download baselines,
using capacity-ordered batches, with explicit frontier, state and count truncation
flags. Repeated row blocks use compact metadata without reducing physical traffic. Feasible completed greedy
baselines retain equal-or-better witnesses through approximation and display selection;
protected witnesses can increase the returned option count. Communication rate means
requested record bits divided by upload plus download; higher is better.
The separate [cost-grid DP](planner/README.md#cost-indexed-dynamic-programming) uses
rounded storage/download cells rather than sampled row states. Its
[dp-costs experiment](experiments/dp-costs/report.md) records the same qualifying workload
grid, explicit numerical windows, and a comparison table. It has no global approximation
guarantee and supports storage/download budgets and rejects an upload budget.
Default cost grids omit workloads with fewer than **5,000 application records**,
leaving 21 specifications in the [tradeoff grid](experiments/tradeoff-costs/report.md).
This filters evaluation workloads, not the planner API or correctness fixtures.
The [cost-grid report](experiments/greedy-costs/report.md) covers 1 GB through 1 TB
with 1 KB through 10 MB records for the unsliced baseline; these are model calculations, not encoded databases.
The [updated width-search grid](experiments/application-costs/report.md) records the
new planner and indexed-only alternatives. Full-download limits use a shared
reference across logical widths; storage-first choices can still incur large indexed
communication for highly fragmented records.

The [rate estimator](cmd/pir-rate/README.md) samples monomials using a deterministic
membership test that supports `s=2..20` without field arithmetic or generator
matrices. This does not extend the small encoder's supported sizes.

The s=8..13 catalogue entries now record exact capacities K=3121, 17200,
103556, 455434, 1965739 and 11926435, respectively, with standalone encoding evidence.
The s=11..13 evidence covers in-memory cyclic encoding; remaining
entries have estimated capacities. Historical
experiment snapshots retain their earlier estimates.

The PIR configuration catalogue includes binary extension fields `GF(2^s)` for every integer `s` from 8 through 20, initially with `m=2` and the [bounded-sample parameter profile](docs/parameter-planning.md#configuration-catalogue); prime fields are outside the current scope. For indexed layouts, application records must be at least as wide as the largest PIR record in the supported catalogue; packing is outside the current scope.

The [planner design](docs/parameter-planning.md) describes joint selection of submatrix dimensions, code parameters, and query-sharing groups, including total storage and communication costs. Encoding-time and memory estimates may accompany plans but do not influence partition selection. The [small deterministic reference encoder](encoding/liftrs-go/README.md) now supplies the correctness baseline. The evaluator and greedy planner use catalogue evidence without running either encoder.

The [application encoder](docs/encoder-contract.md#application-encoding) consumes a selected plan, partitions and pads the input, converts it to field elements, and invokes the non-cryptographic and cryptographic PIR encoders for each database; [appref](encoding/appref/README.md) implements this workflow only as a bounded reference. [Code preprocessing](docs/encoder-contract.md#code-preprocessing) takes only code specifications and produces reusable encoding information, independently of any database. Durable manifest/key persistence and large-instance backends remain separate work. The [encoding algorithm review](docs/encoding-algorithms.md) compares the paper's systematic encoder, a cyclic polynomial formulation, and reusable transform implementations; it records the bounded implementations and remaining large-instance work. The
[five-step encoder plan](docs/encoding-algorithms.md#6-five-step-implementation-plan)
has completed table-free arithmetic, hybrid full/low polynomial products and CPU
cyclic measurements through s=13. One hour is a first-use ballpark; the s=13
trial used a two-hour timeout and the existing 12 GiB Go-memory guard. Larger
forecasts remain estimates, not benchmarks or feasibility proofs. Compact binary
artifacts, further field-kernel optimization, bounded-RAM
encoding and optional GPU acceleration remain planned, not supported large-instance
capabilities.

- `utils/gf2n.go` implements `GF(2^n)` with log/exp tables for degrees 4, 8, 12, 13, 14, 16, 17, 18, and 20.
- Degrees 9, 10, 11, 15, and 19 remain absent from that legacy backend; the
  separate `encoding/liftrs-go` table-free arithmetic supports every s=2..20. Terabyte-scale application databases are a parameter-planning target, not a verified implementation capability.
- `utils/curve_generation.go` implements prime-field curve generation and `GF(2^n)` curve generation for the online Lift-RS path.
- `config.json` stores base64-encoded permutator and encryptor keys, a permutator tweak, and a server URL. The configuration helper fills in and persists missing key fields; `Gen` creates ephemeral in-memory keys.

## Project structure

```text
pirconfig/      JSON configuration catalogue, schema, Go loader, and validation tests
planner/        Bounded tradeoff search, greedy partitioning, validation, and cost evaluation
cmd/pir-rate/   Monomial-sampling Lift-RS rate estimator
pir/            Go Lift-RS client API, tests, benchmarks, and RM encoder wrapper
utils/          Field arithmetic, curve generation, permutation, masks, and matrix I/O
encoding/       RM placeholders; liftrs-go/, liftrs-cpp/, cryptoref/, appref/ backends
server/         Julia HTTP query handler
input/, output/ Example CSV data and generated matrix locations
experiments/    Scripts, reports, and results for new measurements
docs/           Construction and data representation specifications
message-board/  One file per agent message, with an index in README.md
```


## Parallel encoding implementations

The Go Lift-RS package is now `encoding/liftrs-go` (Go package name `liftrs`);
update imports from `github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs` to
`github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-go`. The parallel
[C++ implementation](encoding/liftrs-cpp/README.md) currently contains comparison
kernels and an additive-transform stage, not a complete PIR encoder or integrated
native backend. The [language comparison](experiments/kernel-languages/report.md)
records matched work, output checks and compiler/profiling boundaries. Historical
source manifests retain their measured paths; active code and links use the new
language-labelled directories.
