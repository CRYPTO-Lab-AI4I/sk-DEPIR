# sk-DEPIR: Lifted Reed-Solomon Secret-Key PIR

This repository implements the **Lift-RS online construction** from the accompanying [paper](paper.pdf), *Doubly-Efficient Secret-Key PIR with Low Storage Overhead*. It is being prepared as a readable, documented, and auditable open-source library. The paper also describes a Reed-Muller construction decoded along concatenated curves; that protocol is outside this repository's scope.

The client follows the permuted smooth-code blueprint: it queries the other `q-1` points on a degree-`t` curve through the requested location, using a secret permutation and masks. The Go implementation is `pir.LiftedRS`. `pir.NewPIR` remains as a compatibility alias for `pir.NewLiftedRS`.

Contributors and collaborating agents must follow [AGENTS.md](AGENTS.md), record every change in the [message board](message-board/README.md), and use the [experiment conventions](experiments/README.md) for new measurements. Check this README with every change. During optimization,
[retain preprocessing and CPU profiles](experiments/README.md#profiling-and-preprocessing-retention)
by default; keep large local caches outside Git and report instrumentation overhead.

The current development target is **TRL 5**: a usable, validated client and preprocessing library. The [roadmap](ROADMAP.md) gives the ordered tasks and acceptance criteria. Storage is treated as a plain I/O endpoint returning requested records; server-side protocol computation is outside the scope. TRL 6 acceptance is deferred; evaluating encoding-efficiency improvements remains in scope.

**Encoding scaling:** the [maintained results and capacity tables](docs/encoding-status.md)
are the reference for preprocessing and CPU/GPU encoding modes, including
multi-GPU estimates. The performance goal is **s≤18**; **s=19–20** remain
exploratory. Current standalone encoding is validated through s=16; larger rows
are projections pending implementation and validation. Update the reference as
scaling progresses; the [scaling plan](docs/encoding-scaling.md) tracks the work.

A [bounded HPC table-gap batch](experiments/encoding-table/report.md) measures
22 missing encoding-mode cells at s10/14/15/16 using retained preprocessing.
Submissions and results are separate from validated entries; the reference keeps
laptop and HPC evidence labeled independently.

The optional [GPU scaling experiment](experiments/gpu-scaling/report.md) separates
full-domain plan metadata from explicit per-engine allocation limits. Exact
CPU/GPU operation comparisons now pass through2^28 values; on the measured H200,
forward transform time is2.65 s versus65.11 s on16 CPU cores. This is an operation
comparison, not a full s16 encoding speedup. Ordinary builds and CPU constructors
retain their current setup requirements.

The [actual GPU-assisted s16 preparation experiment](experiments/gpu-preparation/report.md)
reuses saved support and can restart from a verified normalized root checkpoint.
It completed s16 preparation in2h20m15s (2h24m15s allocation elapsed), with the
full polynomial identity verified and72.6 GiB peak sampled process RSS. Saved
constants are retained and reused by the completed cached encoding below.

A [cached s14 CPU profile](experiments/cpu-encoding-profile/report.md) measures
249/141/91 s at8/16/32 threads on the HPC host, with exact reference-hash checks.
The initial profile identifies weak Taylor scaling and43% of total time at32
threads; these timings do not establish s16 encoding latency.
The [within-block CPU Taylor change](experiments/cpu-taylor/report.md) passes
correctness and native sanitizers. Matched s14 encoding improves from87.9 to61.0 s
at32 threads (1.44x), with the original recurrence retained for smaller/coarse
parallel cases. Full-size cached CPU encoding is recorded below.
A [cached s16 CPU encoding](experiments/cached-s16-encoding/report.md) now passes:
20m52.5s encoding,34m53s total job wall time and74.4 GiB peak sampled RSS at32
CPU threads, without a GPU. All systematic positions, sampled independent curves
and packed readbacks pass; the8 GiB packed word is retained. This is one standalone
mathematical encoding, not application/cryptographic integration or TRL5 acceptance.

CPU (Go/C++) and optional [CUDA client encoding](encoding/liftrs-cuda/README.md)
are maintained paths using the same verified code data. CUDA is opt-in and never
a prerequisite for CPU clients, regardless of where preprocessing runs. The
parameter-driven GPU online implementation passes [bounded validation](experiments/cuda-encoding/report.md)
through complete s10 workflows, with zero CUDA memcheck errors/leaks. The
[full cached s16 GPU run](experiments/cuda-s16-encoding/report.md) also passes:
86.0 s first encoding and60.2 s second encoding, plus7.2 s reusable setup, with
complete output hashes matching CPU. It uses67.4 GiB owned device memory and39.0 GiB
peak sampled host RSS on H200 NVL. Encoding is about14.6–20.8x faster than the earlier
32-thread CPU run (historical comparison, separate allocations). Cache loading,
packing and independent verification are timed separately; the full validation
job took30m02s. CPU-only clients remain supported.

The [next step is memory optimization for s17](docs/encoding-scaling.md#current-step-memory-optimization-for-s17),
before further standalone runtime tuning. Current full-array GPU storage would
exceed the available device memory. Compare bounded-memory schedules on s16 first;
make extra computation, transfers, host/disk storage and latency explicit so the
user can choose each material memory–time trade-off. s17 is not yet supported.
The [memory audit](experiments/encoding-memory/report.md) now reconciles s16
measurements and compares buffer reuse, in-place transforms and host/disk-backed
tiles. Its s17 sizes and transfer costs are models; no design or new run budget
has been selected by that audit. The user subsequently chose buffer reuse first;
the [opt-in prototype](experiments/cuda-buffer-reuse/report.md) now passes s16:
51.43 GiB owned device memory (23.7% less),57.4/57.0 s encoding under a56 GiB cap,
with matching CPU output hashes. Matched s10 encoding time is essentially unchanged;
full-size timing versus the prior run is historical. The public GPU default remains unchanged.
The [native CPU memory change](experiments/cpu-buffer-reuse/report.md) now passes
full s16 validation: sampled RAM falls74.37→58.35 GiB, with both owning-output
encodings matching CPU/GPU hashes. Encoding takes21m21s/20m47s; mixed small-run
timing differences do not establish a stable runtime penalty or benefit.
The two-encoding validation job took1h04m36s. CPU execution still needs no GPU.

Alongside that validation, the [next CPU/GPU memory steps](docs/encoding-scaling.md#next-part-shared-cpugpu-memory-schedules)
are reducing transform scratch, streaming constants/output, then evaluating
host/disk-backed transforms. **Slower memory-saving modes address insufficient
memory. When the faster path fits, they require explicit command-line opt-in and
must not become the default.** Production streaming/memory-mode CLI support remains
planned. The [GPU transform scratch study](experiments/gpu-transform-memory/report.md)
now provides an experimental `--layout fast|direct|compare` selector (default fast).
Direct mode now passes component checks through2^25 values: working buffers shrink
256→128 MiB, with forward/inverse time2.1%/2.8% slower at that size. It remains an
explicit option, not the default. Complete-encoding output memory is still a
separate gate; no fulls16 online memory saving is established by this component.
The [bounded GPU output mode](experiments/gpu-output-memory/report.md) now passes
two full s16 encodings:35.429 GiB owned device memory versus51.427 GiB in the
prior reuse mode, with matching CPU hashes. Encoding takes101.5/93.3 s versus the
prior57.4/57.0 s (historical comparison). Host RAM remains39.03 GiB. The matched
s10 comparison measures9.6–10.7% slower encoding; compact requires explicit
`--memory-mode compact`, and faster defaults stay unchanged.
The [CPU counterpart](experiments/cpu-output-memory/report.md) also passes:
39.409 GiB sampled RAM versus58.351 GiB previously,21m29.8s encoding and matching
CPU/GPU hashes. Matched s14 encoding is3.9–6.1% slower. Its explicit
`NewCompactWorkspace` / `EncodeCompactWithStats` API keeps owning outputs;
normal CPU encoding needs no CUDA. These are optional memory–time choices,
not s17 support. Validation jobs took28m59s GPU /49m22s CPU.
The [fast-layout streamed GPU mode](experiments/gpu-streamed-constants/report.md)
now passes full s16 validation:32.005 GiB owned GPU memory versus51.427 GiB resident
reuse (37.77% less), with63.49/58.85 s encoding and matching CPU/GPU hashes.
Host RAM stays39.03 GiB. Repeated19.422 GiB code uploads take about2.24 s per call;
matched s14 encoding is5.6–6.7% slower. Full-size timing versus prior57 s resident
calls is historical. Streaming remains explicit opt-in; resident defaults stay.
s17 still needs34-bit extension/coordinate handling and a one-arena/file-backed
design; the current two-array route would need256 GiB GPU memory. No s17
feasibility is claimed. The complete validation job took16m22s.
The [CPU file-streaming counterpart](experiments/cpu-streamed-constants/report.md)
now passes full s16 validation:20m49.6s encoding and35.484 GiB sampled process RSS,
versus58.351 GiB previously, with matching CPU/GPU hashes. File cache adds memory:
sampled RSS plus tracked resident file pages reaches61.820 GiB; charged batch-task
cgroup peaks around47.92 GiB. These are different accounting views, not a total
35 GiB cap. Matched cached s14 encoding is0.9–1.2% slower. The file route stays
explicit opt-in, resident stays default, and ordinary CPU builds need no CUDA.
The full validation job took36m45s; cold-disk behavior and s17 remain unproven.
The validated [GPU packed-output path](experiments/gpu-packed-output/report.md)
adds explicit `EncodePackedTo` without a full host canonical array. Direct GPU
bit packing uses existing scratch, then bounded writer/hash buffers download the
s-bit format. GPU workspace stays32.005 GiB in streamed-code mode; s16 host output
staging falls16 GiB→8 MiB, with OS cache and a separately mapped verifier extra.
Retry validation now passes the corrected parity regression and full s16 hashes:
65.5–66.2 s including packing/writing/hash, about23.01 GiB encoding RSS and31.01 GiB
separate verifier RSS. Matched s14 packed-file pipelines are27–29% shorter in two
fixed-order samples. Cache/driver memory is extra; GPU workspace is unchanged.
All four memory approaches now have CPU and GPU implementations; see the
[validation sizes and remaining combinations](docs/encoding-scaling.md#current-memory-route-status).
Owning/resident defaults remain. The
[host-backed GPU component](experiments/gpu-host-backed/report.md) keeps the full
host array and bounds device tiles, with measured component savings and costs: at2^22 values, device ownership
32.76→1.26 MiB but forward time0.128→7.73 s. [Parallel work inside tiles](experiments/gpu-backed-parallel/report.md) now passes:
that time falls to1.34 s with unchanged memory/transfers. It remains10.47x the
prior resident sample. The explicit `WithCUDAHostBacked` five-transform online path is now implemented
and [passes full s10 validation](experiments/gpu-host-encoding/report.md):
1.61 s with0.534 MiB explicit device memory versus native CPU0.31–0.32 s. It
retains host code and needs4N work+4N owning result, with repeated transfers;
the faster resident defaults stay. No full s16/s17 timing or backed packed sink
is claimed.







Fast encoding is an explicit priority. Language, compiler, algorithms, hardware,
layout and parallelism remain open choices; select improvements through measured
comparisons while preserving PIR functionality, correctness and usability on
standard setups. Target commodity CPUs and GPUs and measure on locally available
hardware. Larger workloads may need larger machines, but hardware/software
requirements should remain as generic as practical.

Standalone mathematical encoding now has an experimental **s=16** checkpoint;
large-instance discovery/preparation/encoding use the linked harnesses, while
default public discovery remains bounded to s≤13. The
[scaling study](docs/encoding-scaling.md) now includes resource accounting,
correctness prototypes, local calibration and an
[opt-in parallel CPU integration](experiments/parallel-encoding/report.md).
On the earlier48 GiB laptop, three profiled s=13 encodings have medians of
**29.8 s at eight workers and 20.2 s at sixteen**, with identical full canonical
outputs. Process peak RSS is **1.15 GiB**. The sixteen-worker time is about 7.6×
faster than the retained 153.4-second serial native baseline; this is a historical
comparison, not a freshly matched speedup trial.

The new path uses arbitrary-basis/affine C++ transforms, parallel matrix/Taylor
execution, blocked in-memory transposes, reusable workspace and buffered minimal
preprocessing loading. Existing reference/default APIs remain available. Reusable
large code preparation remains Go and is separate from these online timings.
There is no disk-backed backend. Historical [profiled s=14 first-use evidence](experiments/parallel-s14/report.md)
now gives exact K=50,976,325 at m=2,t=5: 4 min 15 s discovery, 33 min 38 s reusable
preparation and 90.53 s median encoding. First-use peak RSS is 13.73 GiB; a fresh
cached encoding takes 92.12 s with 4.59 GiB peak RSS. Its 1.19 GiB online plan and
other reusable constants/profiles are retained. These are historical laptop
mathematical-encoding measurements; later s15/s16 checkpoints are linked separately.
Cryptographic/application integration and larger-size feasibility remain separate gates. The [resource model](experiments/encoding-resources/report.md)
retains historical analytical payloads, not measured larger-size RAM bounds.

The opt-in matrix encoder now also has a
[complete C++ online path](encoding/liftrs-cpp/README.md#complete-native-online-encoding),
including point products and canonical assembly in one native call. Discovery,
reusable code preprocessing and artifact loading remain Go. The
[native-online comparison](experiments/native-online/report.md) records validation
and cached s=14 measurements separately from the historical hybrid timings above.
All full outputs match and peak RSS remains about 4.59 GiB; clock discrepancies
and host variance make the speed comparison inconclusive. The original transform adapter and Go reference remain available. A separate
[timing study](experiments/encoding-timing/report.md) now obtains agreeing clocks
and two complete AC pairs: 84.47 s full C++ versus 94.55 s hybrid, about 10.7%
less elapsed time with similar CPU cost and 4.59 GiB process peak RSS. The initial
power-transition pair is reported separately; these are profiled, warm-workspace
measurements on this host.

The [s=15 preparation-memory audit](experiments/preparation-memory/report.md)
now separates principal arrays (about 22 GiB) from allocation/GC overhead. A
48 GiB machine was judged plausible before the completed s15 admission below. Compact
validation, bounded factor batches and private preparation scratch reuse are now
implemented and pass small reference checks; the
[s=14 preparation checkpoint](experiments/preparation-checkpoint/report.md)
passes complete P/G/spectrum digest checks: 34m 32.5s preparation, 9.01 GiB
process peak RSS and 5.42 GiB peak reachable heap at diagnostic checkpoints.
Forced GC/profiling and retained-basis loading differ from the historical run;
this is not a controlled latency comparison. The guarded
[s=15 local experiment](experiments/parallel-s15/report.md) now passes exact
discovery, reusable preparation, three complete native encodings and fresh cached
execution: K=216,914,047, median encoding 393.19 s, cached encoding 393.69 s,
18.46 GiB cached peak RSS and 20.55 GiB first-use peak. Preparation was manually
paused/resumed, so raw setup elapsed time is interrupted. These establish
standalone local feasibility; application/cryptographic integration remains open.

## Development handoff

Local encoding development stops at the verified s=15 checkpoint. Further scaling
moves to a larger-memory commodity CPU environment, initially a single shared-memory
node (provisionally 256 GiB RAM). Experimental s16 preparation and one cached
CPU encoding now pass as linked above; larger sizes remain unverified. The
[agent handoff](docs/development-handoff.md) records the baseline, first actions,
portability gaps, artifact transfer/checksums and budget review. Large caches are
not in Git. The [destination transfer receipt](message-board/2026-11-09-1316-codex-orient-001-completed.md)
verifies the transferred s=13/14/15 reference data; preserve the source copies.
[HPC setup](docs/hpc-setup.md) records the local Go installation and Slurm GPU
inventory. Cached CPU encoding now passes through experimental s16 on Slurm;
complete GPU database encoding and larger-size feasibility remain unverified.
The [Linux compute-node validation](experiments/hpc-cached-s15/report.md) reuses
the retained s=15 constants under bounded Slurm resources, separating correctness
checks, staging, loading, encoding and output I/O. Both full suites and the complete
cached s=15 output checks pass. The [s=16 audit](docs/encoding-s16.md) explains
count/header handling without widening individual field values. Stored counts
remain 32-bit through s=15 and use 64 bits at s=16. The
[compatibility validation](experiments/count-width-validation/report.md) passes
small-buffer boundaries, sanitizer/full-suite checks and complete cached s15
output comparison; that boundary validation preceded the completed s16 runs above.
The [guarded s16 first-use experiment](experiments/hpc-s16/report.md) now audits
array/file lifetimes, validates the shared first-use harness on retained small
fixtures and defines a bounded single-node CPU run. Its report records live status;
experimental admission is separate from demonstrated feasibility.
The first full run was stopped after discovery because preparation would exceed
its time budget; discovery data is retained. A
[bounded arithmetic comparison](experiments/preparation-arithmetic/report.md)
measures 1.56× faster complete s10 preparation and a 3.55× faster s16 leaf kernel
using a small private multiplication table. Those initial kernel trials were
followed by the completed full preparation checkpoint above.
The [current scaling step](docs/encoding-scaling.md#current-step-larger-memory-preparation-efficiency)
records targeted stage profiling, CPU kernel/batch/root improvements, matched
GPU trials and subsequent preparation/cached-encoding admission outcomes.
Its first sub-step is the [bounded stage profile](experiments/preparation-profile/report.md),
using saved support and representative kernels with separate correctness checks.
The [CPU kernel trials](experiments/preparation-kernels/report.md) now integrate
direct private lookup loops and validate complete smaller-code outputs. An optional
[GPU tile prototype](experiments/gpu-transform-tiles/report.md) passes exact
forward/inverse comparisons on an H200 NVL: warmed, transfer-inclusive tiles are
19–25× faster than the four-core lookup CPU reference at the largest tested buffer.
Device initialization is separate. This is a transform component, not a complete
GPU encoder or a full s16 speedup; the library retains its CPU backend.
The [GPU composition experiment](experiments/gpu-matrix/report.md) now validates
complete matrix/Taylor transforms, persistent workspace and resident convolution.
At 2^22 values, complete GPU transforms take 0.137 s versus 0.760 s on 16 CPU
cores, excluding reusable setup but including transfers. GPU-assisted s10
preparation takes 2.719 s versus 6.806 s; complete worker times including setup,
checks and I/O are 6.790 s versus 10.743 s. These are instrumented sequential
measurements of preparation and transform components, not full s16 online results.
The optional primary CUDA client encoder is described above; default construction uses CPU.
The [resident product integration](experiments/gpu-products/report.md) now handles
unequal operands, low prefixes and overlap-add without returning intermediate
transforms to the CPU. Complete s10 preparation measures 2.400 s at the existing
cutoff, or 1.528 s with an earlier GPU crossover, versus 2.735 s transform-only
GPU and 6.647 s CPU. [Bounded s16 intervals](experiments/s16-resident-intervals/report.md)
match exact partial products and measure about 2.1× faster early root-product
work with resident GPU; later full-domain operations remain outside the measured
GPU cap, so these figures do not establish a full-s16 time or speedup.

## Implementation status and limits

The online query and decode path supports locality `ℓ=q-1`, dimensions `m=2` and `m=3`, prime fields, and the listed binary extension fields. `Params.Q`, `Params.K`, and `Params.M` denote the field size `q`, curve degree `t`, and database dimension `m`. Production `Query` and `QueryLocal` append 128 dummy positions, so the transmitted query length is `L=(q-1)+128`.

**A standalone small Lift-RS encoder is available in [encoding/liftrs-shared](encoding/liftrs-shared/README.md).** It provides reusable generator-matrix preprocessing and systematic field-element encoding for correctness fixtures over GF(4), GF(8), and GF(16), with at most 256 codeword positions. These are not the target security configurations. **The existing file-based API has no integrated Lift-RS encoder.** `LiftedRS.Encode` calls the external Reed-Muller encoder (`RMEncoding`) as a testing placeholder. `encoding/rme_legacy/rme.cpp` contains an experimental `LRSE2D` helper, but the encoder command protocol does not select it. RM-backed tests therefore validate the Lift-RS online protocol only on compatible RM codewords. They do not establish Lift-RS encoding, rate, storage overhead, or encoding time.

**Paper and cyclic encoders now share the paper's information set** through the
[bounded orbit API](encoding/liftrs-shared/README.md#paper-and-cyclic-encoders) for `m=2`.
In-memory experiments accept `s=2..13` (at most 67,108,864 positions); completed
catalogue benchmarks cover s=8..13. Discovery, method preparation,
public artifact loading, and database encoding are separate. Prepared artifacts
include cached solver data or cyclic polynomials/spectra, reused for many databases.
The [CPU comparison](experiments/encoder-comparison/report.md) reports setup, artifact
sizes, per-database encoding, process memory and output I/O separately. This does
not provide streaming, GPU or integrated file-based encoding; experimental s=14
validation uses the separate minimal-artifact/matrix path described above.
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

Standard project commands prefer C++ transforms:

```sh
make test                         # enables -tags liftrs_cpp
make encode                       # builds native support and selects C++
make encode BACKEND=go             # explicit Go encoding
make encode BACKEND=both           # explicit comparison, only when needed
make test-go                      # pure-Go build, CGO_ENABLED=0
```

`make encode` reuses the retained s=13 preprocessing and writes a fresh profiled
experiment run; see [requirements and evidence](experiments/native-encoding/report.md).
These are command defaults, not an encoder API change. The shared `Encode` method
and explicit `WithCPPTransforms` adapter are unchanged. Go remains the reference;
experimental scaling algorithms remain opt-in. C++ commands require cgo and a
C++17 compiler; ordinary `go test ./...` retains Go's normal untagged behavior.
Legacy RM tests in the full suite also need their archived native binary; both
Make test targets build it. Tests write CSV fixtures, so use an isolated checkout
when preserving local generated/example files matters.


- Go 1.21.6
- The standalone Go Lift-RS path needs no native compiler. The optional C++
  transform backend needs cgo and a C++17 compiler.
- Legacy file-based RM tests need the archived C++ encoder, or Julia 1.10+ as a
  fallback. Julia dependencies are in `encoding/rme_legacy/Project.toml` and
  `encoding/rme_legacy/Manifest.toml`.

On macOS 26, Go 1.21.6 compiled this checkout but its generated test binary failed to load (`missing LC_UUID`). The current suite, including the reference encoder, was verified with Go 1.27.3 on that platform.

```bash
go build ./...
make -C encoding/rme_legacy rme
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

Tests are in `planner/`, `cmd/pir-rate/`, `pirconfig/`, `pir/`, `utils/`, `encoding/liftrs-shared/`, `encoding/cryptoref/`, and `encoding/appref/`. The [reference package tests](encoding/liftrs-shared/reference_test.go) include exhaustive tiny-instance curve and message checks and a runnable API example. Run them from the repository root with `go test ./...`.

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

The [planner design](docs/parameter-planning.md) describes joint selection of submatrix dimensions, code parameters, and query-sharing groups, including total storage and communication costs. Encoding-time and memory estimates may accompany plans but do not influence partition selection. The [small deterministic reference encoder](encoding/liftrs-shared/README.md) now supplies the correctness baseline. The evaluator and greedy planner use catalogue evidence without running either encoder.

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
  separate `encoding/liftrs-shared` table-free arithmetic supports every s=2..20. Terabyte-scale application databases are a parameter-planning target, not a verified implementation capability.
- `utils/curve_generation.go` implements prime-field curve generation and `GF(2^n)` curve generation for the online Lift-RS path.
- `config.json` stores base64-encoded permutator and encryptor keys, a permutator tweak, and a server URL. The configuration helper fills in and persists missing key fields; `Gen` creates ephemeral in-memory keys.

## Project structure

```text
pirconfig/      JSON configuration catalogue, schema, Go loader, and validation tests
planner/        Bounded tradeoff search, greedy partitioning, validation, and cost evaluation
cmd/pir-rate/   Monomial-sampling Lift-RS rate estimator
pir/            Go Lift-RS client API, tests, benchmarks, and RM encoder wrapper
utils/          Field arithmetic, curve generation, permutation, masks, and matrix I/O
encoding/       liftrs-shared/, liftrs-go/, liftrs-cpp/; rme_legacy/; reference wrappers
server/         Julia HTTP query handler
input/, output/ Example CSV data and generated matrix locations
experiments/    Scripts, reports, and results for new measurements
docs/           Construction and data representation specifications
message-board/  One file per agent message, with an index in README.md
```


## Encoding implementation paths

See the [encoding directory guide](encoding/README.md):

- `liftrs-shared/`: common Go API, preprocessing, arithmetic, polynomial strategy,
  codeword assembly, correctness oracle and native adapter.
- `liftrs-go/`: Go-only transform execution.
- `liftrs-cpp/`: C++ CPU kernels, file-backed execution, C ABI and cgo binding.
- `liftrs-cuda/`: optional CUDA kernels and dynamically loaded cgo binding.
- `rme_legacy/`: archived RM sources, subprocess driver, dependencies and benchmark artifacts.

Import `encoding/liftrs-shared` (package `liftrs`) for the Go, C++ CPU and optional
CUDA mathematical encoding paths.
The Go path needs no C++ compiler. Enable `-tags liftrs_cpp` with cgo and a C++17
compiler, call `WithCPPTransforms`, and close the native adapter when finished.
Preprocessing remains Go. The [integration experiment](experiments/native-encoding/report.md)
measured s=13 medians of 190.66 s and 153.41 s before directory reorganization,
with identical codewords and about 2.26 GiB peak RSS. This move preserves algorithms
and artifact formats; it does not constitute a new performance measurement.


The explicit [CPU packed sink](experiments/cpu-packed-output/report.md) packs
canonical values directly into dead scratch and writes through a bounded buffer.
The two transform arrays still require 8N bytes. It avoids a detached canonical
uint32 result and a separate packing pass. Small/reference/ASan checks pass;
matched s=10 pipelines take 0.325–0.327 s versus owning-plus-packing 0.328–0.340 s
(two fixed-order samples). Atomic assembly itself costs more, so no robust
speedup is claimed. Owning/resident defaults remain.

The [CPU disk-transform component](experiments/cpu-disk-transform/report.md)
replaces full RAM arrays with bounded tiles and backing files. At 2²² values,
576 KiB private buffers replace 32 MiB resident arrays, but forward time rises
from 0.63 s to 4.72 s on NFS. Filesystem cache and runtime are extra.

`OpenScalingCPPDisk` integrates the complete disk-backed CPU schedule with
streamed code and an owning host output. [Full s=10 validation](experiments/cpu-disk-encoding/report.md)
passes: first/reused encoding took 2.86/1.14 s versus file-fast 0.33/0.31 s on NFS.
Native buffers of 1.125 MiB plus a 4 MiB owning result replace 8 MiB resident
arrays at s=10, with 8 MiB backing files and additional cache/runtime memory.
At s=16, workspace files alone require 32 GiB, exceeding the current experiment
budget; no full-size disk timing is claimed. A disk-mode packed sink and other
mode combinations remain separate. See the [consolidated plan status](docs/encoding-scaling.md#current-memory-route-status)
before larger validation or s=17 admission.
