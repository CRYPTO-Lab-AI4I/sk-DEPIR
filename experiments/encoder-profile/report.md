# Encoder CPU profiles at s=9

## Purpose and reproduction

Identify arithmetic bottlenecks separately in reusable code preprocessing and
per-database encoding, following the [table-free checkpoint](../tablefree-encoding/report.md).
The encoder implementation is revision `d464fcf`; this experiment was added on top
of it. Exact source and input hashes are in [results.json](results.json).

From the repository root:

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoder-profile/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --seconds 4
```

The [script](run.py) builds the [driver](main.go) in a temporary directory,
then runs four fresh processes. Requires Go, Python 3 and git. Each process loads
one checkpoint artifact before sampling. Preparation repeats only method-specific
code preprocessing, reusing the loaded basis. Encoding loads the complete prepared
encoder and generates its input before sampling. It repeats the same uniform
GF(512) message with seed 71. All use s=9,m=2,t=7, N=262144, K=17200.

CPU sampling starts immediately before the first call and stops after completing
the call that reaches the requested minimum duration. Allocation and GC inside
these calls are included. Artifact loading, message generation, serialization,
verification, hashing and file I/O are excluded. Preparation's final serialized
artifact must match the checkpoint byte for byte; encoding checks every systematic
symbol and compares the canonical uint32 output hash to the saved checkpoint.
There are no keys, masks or database-dependent computations in preparation.

## Results

Saved profiles and their flat/cumulative text views:

| Stage | Calls | Sampled wall interval | Raw profile | Flat / cumulative views |
| --- | ---: | ---: | --- | --- |
| Paper preparation | 1 | 8.287 s | [profile](prepare-paper.pprof) | [flat](prepare-paper-flat.txt) / [cumulative](prepare-paper-cum.txt) |
| Cyclic FFT preparation | 2 | 7.937 s | [profile](prepare-cyclic-fft.pprof) | [flat](prepare-cyclic-fft-flat.txt) / [cumulative](prepare-cyclic-fft-cum.txt) |
| Paper encoding | 1 | 4.171 s | [profile](encode-paper.pprof) | [flat](encode-paper-flat.txt) / [cumulative](encode-paper-cum.txt) |
| Cyclic FFT encoding | 9 | 4.470 s | [profile](encode-cyclic-fft.pprof) | [flat](encode-cyclic-fft-flat.txt) / [cumulative](encode-cyclic-fft-cum.txt) |

The following percentages refer to **all sampled process CPU stacks**, including
runtime stacks. Cumulative percentages include descendants; nested rows must not
be added. They are sample attribution, not fractions of elapsed wall time.

| Stage | Polynomial multiplication, cumulative | Additive FFT, cumulative | Base-field multiplication, flat |
| --- | ---: | ---: | ---: |
| Paper preparation | 88.6% | below displayed threshold | 64.6% |
| Cyclic FFT preparation | 57.0% | 4.7% | 31.4% |
| Paper encoding | 90.5% | 4.1% | 72.0% |
| Cyclic FFT encoding | 18.5% | 74.9% | 66.4% |

Cyclic encoding calls polynomial multiplication only for H=(P*M) mod X^K;
G*H uses two transforms and a cached spectrum. Paper encoding uses polynomial
multiplication both for the prefix and the remainder-tree solver; the 90.5%
figure is their combined cost. Cyclic preparation spends a material portion of
samples in the runtime (including GC coordination and OS calls); it would be
misleading to treat every sample outside multiplication as useful transform work.

Both final prepared artifacts and both complete encoded words match the checkpoint.
These are profiles, not replacement latency benchmarks: profiling adds overhead,
and there is only one independent run per stage. Use the earlier checkpoint for
unprofiled timing measurements.

## Implications for the scaling plan

Portable scalar base-field multiplication is a common hotspot across both paths.
A faster multiplication backend can help the transforms and polynomial kernels.
At this size, even eliminating the cyclic prefix product entirely removes only
about 18.5% of sampled CPU work; it cannot explain a several-fold whole-encoder
speedup. Fast truncated multiplication remains necessary for the intended
asymptotic scaling and for reducing paper/preparation costs. Its benefit should
be measured separately from any field-kernel change.

Continue the [existing plan](../../docs/encoding-algorithms.md#step-1-remove-full-extension-field-arithmetic-tables)
with a verified fast polynomial multiplication/truncation backend, retaining
Karatsuba as a small-input baseline. Compare kernels and complete encodings,
including reusable setup, before increasing supported instance sizes. CPU-specific
arithmetic remains a separate choice; no GPU or external-memory scope decision
follows from these profiles alone.

## Environment and limits

Measured on the same local Apple M5 Pro, 18-core, 48 GB MacBook Pro as the
[checkpoint environment](../tablefree-encoding/results.json). This run records
macOS 26.6.2 ARM64, Go 1.27.3, Python 3.9.6 and GOMAXPROCS=18 in its results.
The arithmetic path is single threaded; runtime/GC workers may run concurrently.
Profile files are written outside the sampled computation. Storage hardware does
not enter an I/O throughput claim here. No cache flushing or CPU affinity is used.
A short package build check overlapped the first profiling process; none of the
other profiles ran alongside tests. These observations are local diagnostic
samples, with no confidence intervals, extrapolation to s=20, function-specific
peak RAM measurement, or production-security claim.

To inspect another view, rebuild the driver and run `go tool pprof -top` on a
saved `.pprof` file (functions are symbolized in the saved text views). Profiles
and source/input SHA256 hashes are retained for review. Existing performance
reports and historical measurements are unchanged.

## Subsequent implementation

The [polynomial-product checkpoint](../fast-polynomial/report.md) adds hybrid
full/low products and compares its effects against this implementation. The
measurements above remain a snapshot of the source revision recorded here.

The shared driver now accepts code-degree arguments with the same s=9,t=7 defaults.
[The s=10 profile](../encoder-profile-s10/report.md) measures the later hybrid
polynomial backend; it does not replace the historical samples above.
