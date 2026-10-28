# Go/C++ encoding kernel comparison

## Purpose and reproduction

Compare the current Go fixed-constant kernel with a portable C++ implementation
of the same algorithm, then measure a complete additive-transform stage. Both
use shared actual s=13 field/transform tables, input values and batch sizes.
This is evidence about these compiled implementations, not a universal language
ranking or a complete native PIR encoder.

From the repository root (Go, Clang/GCC-compatible C++17 and Python 3):

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/kernel-languages/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --cxx /usr/bin/clang++ \
  --smoke --run-name smoke-new
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/kernel-languages/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --cxx /usr/bin/clang++ \
  --run-name comparison-new
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/kernel-languages/verify.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --cxx /usr/bin/clang++
python3 experiments/kernel-languages/analyze.py
```

[run.py](run.py) preserves named runs, shared fixtures and matching binaries;
choose unused names rather than deleting useful results. [analyze.py](analyze.py)
verifies the recorded `comparison` dataset and regenerates [table.md](table.md).
The recorded small [smoke results](smoke/results.json) use the
[original runner snapshot](smoke/run-snapshot.py); the later runner names binaries
per run so new measurements do not overwrite earlier binaries.

Implementations are [liftrs-go](../../encoding/liftrs-shared/README.md), with
[opt-in private-kernel helpers](../../encoding/liftrs-shared/language_bench_test.go),
and [liftrs-cpp](../../encoding/liftrs-cpp/README.md), with
[transform.hpp](../../encoding/liftrs-cpp/transform.hpp) and
[bench.cpp](../../encoding/liftrs-cpp/bench.cpp). The C++ side currently contains
field/table kernels and additive transforms, not code discovery, systematic
codeword encoding, cryptographic processing or application-plan execution.

## Environment, source and retained inputs

Measured on the same Apple M5 Pro, 18-core/48 GB machine: macOS 26.6.2 arm64,
Go 1.27.3, Apple Clang 21.0.0, libc++, default OS scheduling and Go runtime settings.
No core affinity was forced and no GPU or explicit target-specific ISA flag was
used. C++ flags are `-std=c++17 -O3 -g -fno-omit-frame-pointer`, plus
`-Rpass=loop-vectorize -Rpass-missed=loop-vectorize` diagnostics. The
[compiler log](comparison/compiler.log) records optimization remarks; the kernel/
transform loops generally report missed vectorization. This is not a SIMD-tuned
backend. Results are for this compiler/hardware combination, not measured x86/GPU
performance.

[comparison/results.json](comparison/results.json) records base revision
`0431f80` with uncommitted path migration/new comparison code, exact source hashes,
compiler commands, fixture/binary hashes, per-call timings and output checks.
The existing mathematical Go code is unchanged apart from import/path migration.
Normal container representations, inlining and compiler decisions differ between
the Go baseline and C++ port; this does not isolate a language effect independently
of those implementation choices. In particular, a direct native port need not be
optimal and the Go implementation is not maximally specialized.

Large fixtures, binaries, outputs and fixture-generation CPU profile are retained
under ignored `.cache/liftrs/language-s13/`; [retained-inputs.json](retained-inputs.json)
records local input/binary sizes and checksums. Seed 1313 generates canonical uniform
field elements. Shared tables come from actual base/extension transforms over
GF(2^13), polynomial 0x201b, and its fixed quadratic tower. Nontrivial table payloads
are 25,856 bytes (base) and 124,992 bytes (extension); both languages use the same
entries. Zero/identity constants are omitted from the kernel table list. This
is not a measurement of complete code preprocessing; the existing 432 MiB s=13
code cache is preserved separately.

`base.bin` and `extension.bin` contain the field size, table/input counts, each
constant/table and 4,194,304 uint32 inputs. `transform.bin` contains field/tower,
sparse-polynomial terms, offset values and fixed products; `transform-input.bin`
contains 67,108,864 dense extension-field coefficients. Fixtures are little-endian;
the native output writer uses the measured little-endian architecture. Fixture
checksums are recorded and reused. This internal format is not a public cache API.

## Timing and correctness methodology

Each kernel call performs **134,217,728 element products**, repeating batches of
262,144 or 4,194,304 elements (1 or 16 MiB). Fixed work selects one actual table;
rotating work selects successive actual tables per 4,096-element block. Both
languages materialize every output, use an uninlined batch entry, and keep the
same pass counts. C++ assembly inspection confirms the batch call remains inside
the pass loop ([check-codegen.py](check-codegen.py),
[assembly excerpt](comparison/cpp-pass-loop.txt)); output checksums prevent
accepting empty/incorrect results.

Each transform call is one forward additive FFT over 1,048,576 or 67,108,864
coefficients using the same field, tables and recursion. The input is dense
uniform extension-field data, not an actual sparse message-prefix/codeword.
Three repetitions are timed, with input copies, allocations, file I/O, hashing
and inverse validation outside the timers. Go performs forced GC before each
call and collects a separate CPU profile during each timed workload; profile
start/stop is outside the timer. C++ has no GC and measures with steady_clock.
These are standalone batches: no cgo/FFI, process startup or storage transfer
is included in the reported medians.

Both languages also apply the inverse and require exact recovery of the original
input. The smoke fixture checks the Go forward transform against independent
Horner evaluation at length 16. All complete forward outputs and batched outputs
must have identical canonical SHA-256 across languages. These checks are stronger
than comparing timing alone, but they do not constitute a full native encoder test.

## Profiling

Go CPU profiles and flat views for all three repetitions are retained in each
[comparison case directory](comparison/results.json). The fixture generator also
keeps its Go profile in the local cache. The macOS sampler initially could not
inspect benchmark processes inside the sandbox; per-case `sample.log` and sampler
return codes preserve those failures. This is the documented reason native
per-case profiles are unavailable, not an optimization-phase profiling opt-out.

A scoped sampler outside the sandbox successfully collected 20 seconds from the
running full-size C++ transform process, at a 10 ms interval:
[native sample](comparison/forward-extension-fixed-67108864/cpp/cpu-sample-external.txt),
[exact command/binary hash](comparison/forward-extension-fixed-67108864/cpp/external-profile.json).
[profile-native.sh](profile-native.sh) reproduces that sampling command with the
current benchmark PID and output path. No root/sudo or hardware change was needed.
This sample covers only a portion of the native case. Go and native profiling
coverage is not identical, so timing differences include that observer limitation;
no isolated profiling-overhead correction is applied.

## Results

[The full table](table.md) and [raw timings](comparison/results.json) give:

| Work | Go median | C++ median | C++ speedup |
|---|---:|---:|---:|
| Base tables, fixed, 16 MiB batches | 0.3790 s | 0.1593 s | 2.38× |
| Base tables, rotating, 16 MiB batches | 0.3902 s | 0.1598 s | 2.44× |
| Extension tables, fixed, 16 MiB batches | 0.4794 s | 0.2884 s | 1.66× |
| Extension tables, rotating, 16 MiB batches | 0.4819 s | 0.2888 s | 1.67× |
| Forward FFT, 1,048,576 elements | 0.6101 s | 0.5326 s | 1.15× |
| Forward FFT, 67,108,864 elements | 63.2141 s | 50.1112 s | 1.26× |

The 1 MiB batch variants give base speedups 2.28–2.48× and extension speedups
1.64–1.70×. All output hashes match; both inverse checks pass. The full-size FFT
uses about **21% less time** in C++ for this workload. This is a useful gain,
but the isolated kernel advantage does not transfer unchanged to the full stage.
No full-encoder speedup, production native integration or foreign-call overhead
has been measured. Peak RSS was not measured uniformly; input/output buffer
payloads and shared table bytes are explicit instead of inventing a RAM result.

This supports investigating a batched native FFT backend in an actual encoding
pipeline, comparing full codeword output, setup/ownership/interface costs and
portable deployment before selecting it. Go unrolling/table improvements and
other native algorithms remain eligible; the language choice is not settled by
one direct port. The existing Go backend remains the current codeword-encoding baseline.

## Validation and naming migration

[verify.py](verify.py), [verification.json](verification.json) and
[verification.log](verification.log) record passing renamed Go/package/catalogue/
planner tests, vet, independent C++ field/table checks through s=13 under
AddressSanitizer/UBSan, and the full Go suite in an isolated copy. The native
[oracle test](../../encoding/liftrs-cpp/test.cpp) uses carryless convolution and
long division rather than the ported shift/reduce code. Go's existing small-field,
Horner, dense-code and old-cache tests remain in that suite. No tests or other
benchmarks ran alongside the measured comparison; post-run validation is separate.

The Go package moved from `encoding/liftrs/` to `encoding/liftrs-go/`; its declared
package name remains `liftrs`. C++ lives under `encoding/liftrs-cpp/`. Active Go
imports, scripts, catalogue implementation paths and Markdown link targets are
updated. Earlier message text/commands, measured JSON hashes and snapshot scripts
retain their historical source identity. [source_evidence.py](../source_evidence.py)
verifies those exact bytes from current/mapped paths or committed history, without
rewriting original evidence. Renamed source lookup for retained pprof data was
also checked. This is an experimental import-path migration, documented in README.

The [evidence manifest](evidence-manifest.json) covers the small scripts, results,
profiles and checks. Retained large fixtures/outputs/binaries are outside Git and
located by metadata. No old preprocessing cache or useful profile was deleted.
