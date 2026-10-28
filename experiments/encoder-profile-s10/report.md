# CPU profiles of the s=10 cyclic encoder

## Purpose and reproduction

Identify the next optimization target in the current s=10,m=2,t=6 cyclic encoder,
after [measuring encoding and memory](../encoder-scaling/report.md). Library code
is committed as `c06375c`; this experiment adds profiling-driver arguments and
new evidence, without changing field arithmetic or the encoder algorithm.
[results.json](results.json) records revision, exact source/input hashes and profile
checksums. Prior s=9 profiles are preserved.

Run from the repository root:

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoder-profile-s10/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --seconds 10
```

The [script](run.py) builds the shared [profiling driver](../encoder-profile/main.go)
in temporary storage. Requires compatible Go, Python 3 and git. The driver now
accepts `-s` and `-t`; defaults remain 9 and 7 for compatibility. Field-element
sampling uses the selected width, not the former hardcoded GF(512) bound, and
artifact metadata must match the requested specification before sampling.

Each stage uses its own process and a separate CPU profile:

- **Preparation:** load the committed shared basis, then sample only reusable
  `PrepareCyclicFFT`, including generator construction and cached spectrum setup.
- **Encoding:** load the complete committed cyclic artifact and create a uniform
  message using seed 71, then sample repeated `Encode` calls. No discovery or
  code-only polynomial/spectrum preparation occurs inside these calls.

Both run whole calls until reaching at least ten seconds. Loading, input creation,
serialization, systematic checks, hashes and file I/O are outside sampling.
Allocation/GC and other runtime CPU work during each call remain included.
The final prepared JSON must match the committed artifact byte for byte; encoding
must match the checkpoint's canonical uint32 hash and all systematic symbols.
The script also checks N=1048576,K=103556 and the requested code specification.

## Saved evidence

| Stage | Calls | Sampled wall interval | Raw profile | Text views |
| --- | ---: | ---: | --- | --- |
| Reusable preparation | 1 | 26.252 s | [profile](prepare-cyclic-fft.pprof) | [flat](prepare-cyclic-fft-flat.txt), [cumulative](prepare-cyclic-fft-cum.txt), [call tree](prepare-cyclic-fft-tree.txt), [source](prepare-cyclic-fft-source.txt) |
| Database encoding | 4 | 10.644 s | [profile](encode-cyclic-fft.pprof) | [flat](encode-cyclic-fft-flat.txt), [cumulative](encode-cyclic-fft-cum.txt), [call tree](encode-cyclic-fft-tree.txt), [source](encode-cyclic-fft-source.txt) |

These intervals include profiler overhead and are **not new latency benchmarks**.
Use the earlier checkpoint for uninstrumented preparation/encoding times. Profile
CPU sample totals are 25.85 s and 10.52 s respectively; they differ from wall
intervals. Percentages below use all sampled process CPU stacks as the denominator.

## Findings

The [encoding source view](encode-cyclic-fft-source.txt) separates calls made by
`CyclicEncoder.Encode`:

| Operation | CPU sample attribution |
| --- | ---: |
| Prefix `H=(P*M) mod X^K`, including its three smaller FFTs | 23.0% cumulative |
| Forward full-size FFT for `G*H` | 38.5% cumulative |
| Inverse full-size FFT for `G*H` | 38.0% cumulative |
| Generator spectrum pointwise product | 0.1% cumulative |
| Coordinate recurrence field multiplication | 0.3% cumulative |

Rounding and unsampled/small work account for the remainder. Line attribution is
compiler/profiler-dependent; zero samples on a line do not establish zero cost.
The two full-size FFTs together account for **76.5%**. The prefix is also an FFT
product at this size: unlike the older s=9 profile, its 23.0% is not Karatsuba time.

A different, nested view in the [flat/cumulative report](encode-cyclic-fft-cum.txt)
shows the common arithmetic cost:

- `BinaryField.Mul`: **76.4% flat**, the portable bit-at-a-time base-field product.
- `QuadraticField.Mul`: **83.5% cumulative**, including its base-field calls.
- Transform recursion: **99.2% cumulative**, including both the prefix and full FFTs.

**Do not add these percentages to each other or to the operation table.** They
are different levels of the same sampled call stacks. Base multiplication is
mostly executing inside the FFTs; “76% FFT” and “76% base multiplication” are
compatible descriptions of the same work.

[Preparation](prepare-cyclic-fft-cum.txt) shows the same primitive as a hotspot:
base multiplication accounts for **74.4% flat**, polynomial division **86.9%
cumulative**, and hybrid low-product calls **93.7% cumulative**. Preparation remains
code-only and reusable, so these costs should not be weighted as per-database
latency. A common field-kernel improvement could benefit both stages.

## Recommended next implementation

Prioritize **portable base-field multiplication**, keeping the existing
polynomial-basis representation and current implementation as a reference/fallback.
The current scalar shift/reduce loop dominates samples across both stages. Compare
an improved carryless-product/reduction kernel, including common fixed multipliers
where useful, before changing the overall encoding algorithm or artifact format.

Acceptance for that next change:

1. Preserve every supported field polynomial, exact arithmetic, information-set
   position, public artifact byte sequence and complete encoded output.
2. Verify against independent exhaustive small-field/random larger-field arithmetic,
   dense fixtures, and saved s=9/s=10 encoding/curve checks.
3. Measure field kernels and complete cyclic encoding at s=9 and s=10 separately;
   retain code-only preprocessing, load, allocations and per-database boundaries.
4. Reprofile the improved backend, then advance s=11/12 using measured resource costs.

CPU-specific acceleration can be considered after a portable comparison, retaining
a portable fallback. This profile does not select a GPU backend, establish a
speedup for an unimplemented kernel, or justify removing an arithmetic check.
It does not make a case for prioritizing JSON cleanup: loading is excluded here,
and the separate checkpoint already measures its modest per-load cost.

## Environment and validation

Apple M5 Pro, MacBook Pro Mac17,9, 18 cores, 48 GB RAM; macOS 26.6.2 ARM64,
Go 1.27.3, Python 3.9.6, default GOMAXPROCS=18. Exact hardware/toolchain metadata
are in results. Arithmetic is single threaded; runtime/GC may use other cores.
No tests ran alongside the final profiling run. There is no CPU pinning, warmup,
cache flush, randomized order, independent repeated profile or confidence interval.
This is a local diagnostic, not a portable performance guarantee or memory/I/O
measurement. No library behavior or supported configuration changed.

The initial harness attempt completed preparation but failed when its source-view
filter requested `Encode`, absent from a preparation profile. The filter now selects
`PrepareCyclic` for preparation and `CyclicEncoder.Encode` for encoding. The complete
experiment was rerun; only the completed run is reported above.

Validation commands:

```sh
go test ./experiments/encoder-profile
go vet ./experiments/encoder-profile
go run ./experiments/encoder-profile -duration 1ms -profile /tmp/depir-profile-s9-compat.pprof
```

The package build and vet pass. The default s=9 compatibility run retains its
historical checkpoint output hash. The completed s=10 run reproduces the committed
prepared artifact and complete codeword, and checks every systematic symbol.
Source/input/profile/text-view hashes and report links are verified. The full
protocol suite was not rerun because only a profiling harness and documentation
changed; the committed s=10 checkpoint records the library's full-suite/race checks.

## Subsequent scaling decision

Before implementing the proposed field-kernel optimization, the user requested a
broader runtime/RAM baseline. The [progressive cyclic experiment](../cyclic-scaling/report.md)
now measures s=11/12 and records why larger fields were not admitted. The profile
above remains evidence for optimization priorities, not a claim that those
optimizations have been implemented.
