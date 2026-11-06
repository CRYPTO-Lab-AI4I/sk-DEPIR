# Encoding layout

Lift-RS code is separated by responsibility. Go provides the shared API,
discovery, reusable preprocessing and reference encoding. C++ provides transform
kernels and complete mathematical online execution for the opt-in matrix path;
Go remains its API/loading layer.

| Directory | Contents | Used by |
|---|---|---|
| [liftrs-shared](liftrs-shared/README.md) | Public API, code discovery, reusable preprocessing, field arithmetic, polynomial strategy, codeword assembly, dense correctness oracle and native adapter | Both paths |
| [liftrs-go](liftrs-go/README.md) | Pure-Go additive-transform recursion and fixed-map execution | Go transforms |
| [liftrs-cpp](liftrs-cpp/README.md) | C++ transforms, complete cyclic online execution, C ABI/cgo binding and native fixtures | Native execution |
| [cryptoref](cryptoref/README.md) | Small permutation/masking correctness reference | Reference pipeline |
| [appref](appref/README.md) | Small application-layout correctness reference | Reference pipeline |
| [rme_legacy](rme_legacy/README.md) | Historical RM sources, Julia dependencies, subprocess driver, benchmarks and plot | Old file-based test API only |

## Run either Lift-RS path

Import `github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared` (package `liftrs`).
Use `DiscoverOrbit` then `PrepareCyclicFFT`, or load a trusted prepared artifact.
Code preparation remains Go in both paths. `source.Encode(message)` uses Go
transforms; `source.WithCPPTransforms()` constructs the native adapter, whose
`Encode` produces the same canonical codeword. Close the native adapter afterward.
See the [Go example](liftrs-shared/orbit_example_test.go) and
[native example](liftrs-shared/cpp_example_test.go).

```sh
# Go path; no C++ compiler or cgo required.
CGO_ENABLED=0 go test ./encoding/liftrs-shared ./encoding/liftrs-go

# C++ transforms inside the shared encoder; requires C++17 and cgo.
CGO_ENABLED=1 CXX=clang++ go test -tags liftrs_cpp \
  ./encoding/liftrs-shared ./encoding/liftrs-cpp
```

On Linux, `CXX=g++` is also an intended generic compiler choice; native validation
so far is macOS/arm64. Field degrees and memory limits remain those in the shared
API; moving files does not add large-instance support. Kernel code performs exact
field operations but is not constant-time.

The public encoder import moved from `encoding/liftrs-go` to
`encoding/liftrs-shared`; `liftrs-go` now names only Go execution kernels. Artifacts
and retained preprocessing do not change format. Historical performance results
retain their original source manifests; see [source snapshots](../experiments/source-snapshots/README.md).

## Standard command defaults

From the repository root, `make test` enables `liftrs_cpp`; `make encode` builds
that tag **and selects C++ execution** in the profiling runner. Use
`make encode BACKEND=go` or `make test-go` for explicit Go, and
`make encode BACKEND=both` only for a requested comparison. The Go-only encoding
command does not compile C++ transforms (full legacy tests still build RM).
This does not change the shared encoder API or promote scaling prototypes.
`GO`, `PYTHON`, `CXX` and `RUN_NAME` can be set on Make's command line.


The [profiled s=14 path](../experiments/parallel-s14/report.md) uses experimental
Go discovery/preparation orchestration with native transforms. Its retained
minimal artifact loads through `LoadScalingOnline`, then encodes through
`WithCPPMatrix`. Default `DiscoverOrbit` and JSON basis loaders remain bounded
to s≤13. See that report for explicit first-use and cached-only commands.

The opt-in [matrix encoder](liftrs-cpp/README.md#complete-native-online-encoding)
executes complete database-dependent mathematical encoding in C++ with Go
preprocessing/loading. Its [example](liftrs-shared/scaling_matrix_cpp_example_test.go)
and [experiment](../experiments/native-online/report.md) distinguish it from the
original transform-only adapter and the Go reference path.

Further scaling development moves to a larger-memory CPU host after the verified
s=15 checkpoint. New agents should start with the [handoff](../docs/development-handoff.md),
including separate cache transfer, new-host validation and s=16 cardinality/memory
requirements. Current code does not yet implement or validate s≥16.
