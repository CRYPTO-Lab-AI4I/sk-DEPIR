# Optimized cyclic encoding with optional C++ transforms

This experiment compares the full Go cyclic encoder with the same Go pipeline
using a batched C++ transform backend. Both s=13 cases completed verification. The native-backed median is **153.41 s**,
versus **190.66 s** for Go: **1.243× speedup**, or **19.5% less encoding time**.
This is a full-codeword measurement with matched preprocessing and algorithms.

## Matched implementation

The [Go adapter](../../encoding/liftrs-shared/cpp_encoder.go) and
[C++ ABI](../../encoding/liftrs-cpp/c_api.h) preserve all implemented encoding
optimizations: fixed four-bit multiplication maps, hybrid FFT/Karatsuba dispatch,
low-product pruning, unequal-operand blocking, overlap-add and the cached generator
spectrum. Go retains scalar arithmetic and polynomial orchestration. Only complete
transform calls move to C++; preprocessing remains in Go. Proposed wider-table or
loop-unrolling changes are not added to one side of the comparison.

Native plans copy immutable code metadata once. Each transform borrows a
pointer-free Go `uint32` buffer for the duration of the call; C++ neither copies
nor retains database buffers. `Close` waits for active encodings before releasing
the plan. Native initialization is separate from encoding; `CPPEncodeStats`
reports call counts, transformed elements, native kernel time and Go-observed call
time. Their difference includes locks, validation and timing and is not a
calibrated cgo-overhead measurement. Interface copy bytes exclude the algorithm's
unchanged work-array copies. Native vector-capacity accounting excludes allocator
overhead; Go heap statistics exclude native allocation. Whole-process RSS is also
reported.

## Reproduction and evidence

- [Runner](run.py), [driver](main.go), [evidence analyzer](analyze.py).
- [s=13 results and source manifest](s13/results.json), [small smoke results](smoke/results.json).
- [Verification script](verify.py), [check results](verification.json),
  [full check log](verification.log), [race log](race.log).
- [Runnable API example](../../encoding/liftrs-shared/cpp_example_test.go),
  [oracle, polynomial and lifetime tests](../../encoding/liftrs-shared/cpp_enabled_test.go),
  [native sanitizer checks](../../encoding/liftrs-cpp/c_api_test.cpp).

Both backends run sequentially in separate processes using the same tagged binary,
three encodings of seed-71 data and the same s=13, m=2, t=5 code. Exact K=11,926,435
and N=67,108,864. The measured working tree is based on `c13d034`; source SHA-256
values identify the uncommitted integration changes. Generic C++17 `-O3` code,
no ISA-specific flags; Go 1.27.3 and Apple Clang 21 on an Apple M5 Pro, 18 CPU cores,
48 GiB RAM, macOS 26.6.2 arm64, local APFS storage. The JSON records exact compiler
and platform strings. Execution is single-threaded arithmetic, with Go's normal
runtime/GC and profiling threads.

```sh
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  python3 experiments/native-encoding/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name s13 --backend both --native-sample
python3 experiments/native-encoding/analyze.py s13
```

Choose a fresh run name on repetition; the runner refuses to overwrite evidence.
macOS `sample` needs permission to inspect the runner's own child. It captures
20 seconds at a 10 ms interval during the first native encoding. Every stage has
a retained Go CPU profile; C++ frames can be opaque in Go profiles, so the native
sample supplements them. Profiled times include instrumentation and must not be
mixed with the earlier unprofiled latency baselines. Cache read/checksum/JSON load,
native initialization, each encoding and verification/packed-output I/O are separate.
Profile start/stop and pre-stage GC are outside elapsed encoding intervals.

The 12 GiB guard samples Go runtime memory every 100 ms, with a 9 GiB soft GC limit;
it is not an OS-wide RSS cap. Native plans add small code-only allocations. Each
case has a two-hour timeout. The final output is checked at every systematic
position, against sampled public index mappings and 16 independently evaluated
curves, then canonical and packed hashes are compared with the
[unprofiled s=13 checkpoint](../cyclic-s13/report.md). Repeated calls encode the
same message; detailed verification applies to the final returned word.

### Retained preprocessing and profiles

Reuse `.cache/liftrs/s13-m2-t5/cyclic.json`: 453,243,425 bytes (432.25 MiB), SHA-256
`a2a3bcd5913d821c99d08ceca05d6c5d73ce2abe6e3987910c0dcf5dfe97eb22`.
Its `metadata.json` and `provenance.json` remain intact. Generation and explicit
trusted-cache loader limits are documented in the
[original cache experiment](../encoder-profile-s13/report.md). Regenerate only if needed using the original Go driver, for example
`go run ./experiments/encoder-profile-s13 -s 13 -t 5 -cache NEW_CACHE
-profiles NEW_PROFILES -source-manifest SOURCE_ID`, with fresh paths and a
`provenance.json` recording source hashes. The original report describes the
manifest convention. Do not overwrite the retained cache. This comparison performs no discovery or expensive preparation.
Matching binaries are retained under `.cache/liftrs/native-encoding/`; each result
records its exact path and SHA-256. Profiles and raw events live in each run's
`go/` and `go-cpp/` directories; no useful preprocessing is deleted.

The s=4, t=2 smoke fixture was generated using the existing profile driver:

```sh
go run ./experiments/encoder-profile-s13 -s 4 -t 2 \
  -cache /tmp/depir-native-smoke-cache -profiles /tmp/depir-native-smoke-profiles \
  -source-manifest native-smoke
```

The smoke cache's provenance is the source manifest recorded in its results; its
temporary location is a disposable correctness fixture, not the expensive s=13
cache. The runner expects `provenance.json` beside its metadata (the driver writes only
metadata and the source-manifest identifier). For a newly generated cache, write
this file before running the comparison:

```sh
python3 - <<'PYPROVENANCE'
import hashlib, json, subprocess
from pathlib import Path
cache = Path('/tmp/depir-native-smoke-cache') # or the new s=13 cache directory
sources = list(Path('encoding/liftrs-go').glob('*.go'))
(cache/'provenance.json').write_text(json.dumps({
    'revision': subprocess.check_output(['git','rev-parse','HEAD'], text=True).strip(),
    'source_sha256': {str(p): hashlib.sha256(p.read_bytes()).hexdigest() for p in sources}
}, indent=2)+'\n')
PYPROVENANCE
```

Then run `run.py --s 4 --t 2 --repeats 1 --run-name NEW_SMOKE
--cache /tmp/depir-native-smoke-cache` for the small fixture. Initial smoke
launches stopped before encoding because this fixture lacked that file and the
sandbox blocked CPU metadata; these did not produce measured encoder results.
After supplying provenance and read access, both complete smoke cases passed.
Smoke sub-millisecond timings are correctness checks, not performance evidence.

## Measured results

| Backend | Cache load (s) | Native init (s) | Encoding calls (s) | Median (s) | Whole-process peak RSS (GiB) |
|---|---:|---:|---|---:|---:|
| go | 4.931 | 0.000000 | 188.340, 190.657, 190.863 | 190.657 | 2.257 |
| go-cpp | 5.046 | 0.000155 | 156.505, 153.409, 152.965 | 153.409 | 2.257 |

Cache-load time includes bounded reading, checksum verification and JSON loading.
Native initialization imports a 131,724-byte description and owns 141,904 bytes
(138.6 KiB) of native vectors/descriptors; it took 0.155 ms. No database elements
are copied across the interface. Each encoding uses five transform calls over
234,881,024 total elements (three prefix-product transforms and two full-domain
transforms). The Go-observed minus native-kernel interval is 47.1, 57.0 and 46.5
microseconds per complete encoding, including locks, checks and timers.

Whole-process peak RSS includes cache loading and all repeats, not just encoding;
it was 2,423,062,528 bytes for Go and 2,423,603,200 for the native adapter.
Sampled peak Go heap during encoding was 1,327,314,272 versus 1,324,541,560 bytes.
Native metadata are separate from the Go heap. These observations show no material
memory increase at this instance; they do not establish bounded-memory scaling
for larger fields.

Both final words match all 11,926,435 systematic positions, 16 independent curves,
and the earlier canonical/packed hashes:

- Canonical: `6120f0388f6035dd480e391b26ef93c04ed368a77f2f93c02f6d4f56b99a7b79`.
- Packed (109,051,904 bytes): `0b545a19ff25c1cdb26478ed2786c6ad76052f46c8a258bf11ad763c8f276980`.

See the [generated table](s13/table.md), [raw native events](s13/go-cpp/events.jsonl),
[Go CPU profile view](s13/go/encode_1-flat.txt),
[native-backed CPU profile view](s13/go-cpp/encode_1-flat.txt), and
[native stack sample](s13/go-cpp/cpu-sample.txt). In the native-backed Go CPU view,
98.25% of sampled CPU time is attributed to `runtime.cgocall`: this includes the
C++ transform execution, **not** 98% interface overhead. Independent native timers
place 150.85 of the second repeat's 153.41 wall seconds inside transforms. Native
FFT work remains the main measured optimization target. The native sample resolves
its recursion, while the retained Go profiles cover each full encoding.

The Go path remains the default and the native path is opt-in. The comparison
supports this integration's gain on the measured host and instance; it does not
select a universal language/backend or claim that all orchestration is native.

## Validation and limits

Tagged and default full Go suites pass in isolated copies (legacy tests write
CSV files), alongside the cgo-disabled fallback test, tagged vet, race checks,
and C ABI AddressSanitizer/UndefinedBehaviorSanitizer checks. Go's race detector
does not instrument C++ memory accesses; native concurrency relies on immutable
plans and distinct caller buffers. The tests compare
complete small codewords against an independent dense oracle and cover
truncation/overlap-add, concurrent independent buffers, invalid calls and lifetime.
The native sanitizer fixture comes from the existing shared Go fixture exporter:

```sh
PIR_LANGUAGE_FIXTURES="$PWD/.cache/liftrs/language-s13" \
  go test ./encoding/liftrs-shared -run TestExportLanguageFixtures -count=1
python3 experiments/native-encoding/verify.py --go go --cxx clang++
go test -race -tags liftrs_cpp ./encoding/liftrs-shared ./encoding/liftrs-cpp
```

The fixture exporter retains a large shared input as well as the small transform
description; see the [kernel comparison](../kernel-languages/report.md) for its
resource requirements. Native importer support is currently limited to field degrees 2–13. This is an
experimental mathematical encoder adapter, not cryptographic/application pipeline
integration, production security evidence or a selected default backend. Arithmetic
and lookup tables are not constant-time. Runs are sequential on one machine;
thermal effects, filesystem cache and CPU-frequency changes are not controlled.
The native binding has been tested on this macOS/arm64 host; Linux and other
architectures have not yet been validated. The native stack sample covers only
the first 20 seconds of the first encoding, not the complete mix of transform
calls; it supplements the full-duration Go CPU profiles.


## Directory separation after measurement

The full API/preprocessing/orchestration moved to `encoding/liftrs-shared`; pure
Go transforms are in `liftrs-go`, C++ transforms/binding in `liftrs-cpp`. Active
commands and links follow this layout. Measurements above remain the original
pre-move runs; exact dirty-worktree sources are retained in
[content-addressed snapshots](../source-snapshots/README.md). The
[layout checks](../encoding-layout/report.md) validate small pre-move hashes and
both build paths without repeating these large timing runs.


## Current command defaults

The current runner defaults to `--backend cpp`; it compiles `liftrs_cpp` and runs
only C++ transforms. Select `--backend go` for a pure-Go build, or `--backend both`
for the matched comparison above. Original result files remain the historical
comparison; changing the command default does not rerun or relabel them. Binaries
are retained separately by source hash **and backend selection** to prevent an
explicit Go build from overwriting a native binary's profiling evidence.

Root `make encode` invokes this runner; `make test` enables native tests.
`make encode BACKEND=go` and `make test-go` keep Go explicit. The driver itself now
also defaults to C++; direct use requires `-profiles NEW_DIRECTORY` and rejects
existing CPU profiles. The API remains unchanged. Shared `Encode` does not silently
change backend; command drivers select `WithCPPTransforms` explicitly.

[Native-default smoke](default-cpp-smoke-02/results.json) and
[cgo-disabled explicit-Go smoke](explicit-go-smoke-02/results.json) verify actual
backend selection on the retained s=4 fixture. They reproduce the original smoke
hashes and retain profiles. Timings are tiny correctness checks, not new performance
claims. Initial command checks stopped before encoding when the profile guard also
rejected runner metadata files; [native](default-cpp-smoke/results.json) and
[Go](explicit-go-smoke/results.json) partial results are preserved. The guard now
checks existing CPU profiles rather than unrelated log/metadata files. Measured
failed-build source bytes are content-addressed snapshots. Root native/Go full-suite
checks are recorded in [command checks](../scaling-prototypes/command-checks.json).
