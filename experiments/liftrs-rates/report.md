# Lift-RS code-rate estimates

## Purpose and sources

Estimate `K/N` for all 13 current PIR configurations, with `s=8..20`, `m=2`, and
curve degrees selected by the [bounded-sample security calculation](../bounded-sample-parameters/report.md).
This measures monomial acceptance, not database encoding, storage bytes, or security.
The paper identifies rate with the fraction of good reduced monomials (Section
4.2.2, [p. 28](../../paper.pdf#page=28)) and describes Monte Carlo estimation in
[Section 4.2.3](../../paper.pdf#page=30).

- Runnable script: [run.py](run.py).
- Exact input catalogue snapshot: [pir-configs.json](pir-configs.json).
- Results: [results.json](results.json), including integer counts, seeds, intervals,
  input/source SHA-256 hashes, compiler/platform, and sampling timings.
- Estimator: [cmd/pir-rate](../../cmd/pir-rate/main.go).
- Membership implementation: [monomial.go](../../encoding/liftrs/monomial.go).
- Verification: [membership tests](../../encoding/liftrs/monomial_test.go),
  [curve oracle tests](../../encoding/liftrs/reference_test.go), and
  [estimator tests](../../cmd/pir-rate/main_test.go).

Base revision: `5400889e37e4e59c09f90a73a82cd3dda8947a59`, plus the uncommitted
EXP-001 implementation recorded with this report. The results record that the
working tree was dirty and hash the exact computational sources. No claim is made
that the estimator existed in that base revision. The input snapshot preserves the
catalogue as it stood before these estimates were incorporated.

## Reproduction and environment

Run from the repository root (the script also works from other directories):

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/liftrs-rates/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go \
  --samples 1000000 --seed 20261026 --confidence 0.99
```

Use `--go go` with a compiler on PATH. The recorded run used Go 1.27.3,
Python 3.9.6, macOS 26.6.2, Darwin arm64. CPU model and physical memory were not
available through the sandbox's `sysctl` calls and are explicitly marked unavailable.
They do not affect the monomial test or statistical model. No external encoder,
field tables, input database, or network service is used. The script builds a fresh
binary in a temporary directory before sampling. Prerequisites are Go, Python's
standard library, the repository's Go dependencies, and Git for revision metadata.

Counts and rate intervals are reproducible with the recorded source, parameters,
input, and pseudorandom generator. Timestamps, timings, revision/dirty metadata,
and platform fields may change on reruns. Source hashes allow changes to the
computational implementation to be distinguished from such metadata differences.

## Method

For each configuration, sample **1,000,000 exponent pairs with replacement**.
Each coordinate is uniform in `[0,2^s)`, obtained by masking a Go `math/rand`
`Uint32` to its low `s` bits. Use `rand.New(rand.NewSource(seed))` with
`seed=20261026+1000*s+100*m+t`; the full seeds are recorded below. This randomness
is for a reproducible experiment, not the cryptographic protocol.

A monomial is good exactly when no positive reachable exponent is divisible by
`q-1` (paper Lemma 4.8). At bit `k`, the permitted contribution is any integer
`a_k` in `[0,t*w_k]`, where `w_k` is the bit weight across the exponent tuple.
The implementation checks possible positive multiples `j*(q-1)`, `1<=j<=m*t`,
using binary carries in `sum a_k*2^k+j=j*2^s`. Starting with carry `j`, a reachable
interval `[lo,hi]` advances to
`[ceil(lo/2),floor((hi+t*w_k)/2)]`. Empty intervals are unreachable. Reaching carry
`j` after `s` bits makes the monomial bad. Excluding `j=0` preserves the constant
monomial. This exact reformulation uses `O(s*m*t)` time and constant auxiliary
space per test; it does not build a generator matrix. The independent residue DP
remains unchanged and supplies the membership-test oracle.

The point rate is `good/samples`. For simultaneous confidence `0.99` across
`C=13` configurations, allocate `alpha=0.01/13` to each and use the two-sided
Hoeffding half-width

`epsilon = sqrt(log(2/alpha)/(2*samples))`.

This is approximately **0.1983 percentage points**. Clip intervals to `[0,1]`.
The individual confidence level is `1-alpha`, approximately 99.9231%. The union
bound gives at least 99% simultaneous coverage under independent uniform sampling
within each configuration; independence between configurations is unnecessary.
These are conservative bounds. A fixed pseudorandom stream approximates the
independent-sampling model; it is not a source of information-theoretic independence.

Sampling timings in `results.json` include RNG draws, membership validation/testing,
and counting. They exclude compilation, catalogue parsing, JSON output, code
preprocessing, database encoding, permutation, masking, and transport. Seconds are
recorded only as run metadata; this is not a controlled performance benchmark.

## Results

All rows use `m=2` and 1,000,000 samples. Rate and interval columns are percentages;
the JSON records fractions and exact integer counts.

| s | t | Seed | Good monomials | Estimated rate | Confidence interval |
|---:|---:|---:|---:|---:|---:|
| 8 | 8 | 20269234 | 47,565 | 4.7565% | [4.5582%, 4.9548%] |
| 9 | 7 | 20270233 | 65,300 | 6.5300% | [6.3317%, 6.7283%] |
| 10 | 6 | 20271232 | 99,136 | 9.9136% | [9.7153%, 10.1119%] |
| 11 | 6 | 20272232 | 108,004 | 10.8004% | [10.6021%, 10.9987%] |
| 12 | 6 | 20273232 | 117,106 | 11.7106% | [11.5123%, 11.9089%] |
| 13 | 5 | 20274231 | 178,157 | 17.8157% | [17.6174%, 18.0140%] |
| 14 | 5 | 20275231 | 190,769 | 19.0769% | [18.8786%, 19.2752%] |
| 15 | 5 | 20276231 | 202,023 | 20.2023% | [20.0040%, 20.4006%] |
| 16 | 5 | 20277231 | 214,074 | 21.4074% | [21.2091%, 21.6057%] |
| 17 | 5 | 20278231 | 225,318 | 22.5318% | [22.3335%, 22.7301%] |
| 18 | 5 | 20279231 | 237,180 | 23.7180% | [23.5197%, 23.9163%] |
| 19 | 5 | 20280231 | 248,590 | 24.8590% | [24.6607%, 25.0573%] |
| 20 | 5 | 20281231 | 260,249 | 26.0249% | [25.8266%, 26.2232%] |

## Comparison with paper values

The relevant published estimates are 11.7% at `s=12,t=6` and 25.9% at
`s=20,t=5` from [Table 1](../../paper.pdf#page=64), and 20.24% at `s=15,t=5` and
23.67% at `s=18,t=5` from [Table 3](../../paper.pdf#page=66). All four lie within
the corresponding intervals above. These are comparison points copied from the
paper, not measurements made by this script. The paper does not report their
sample counts or uncertainty intervals, so agreement is a cross-check rather than
an independent correctness or coverage proof.

## Catalogue integration and limitations

The [PIR configuration catalogue](../../pirconfig/configs_catalogue.json) now uses
these observations for every point capacity: `k=round(n*good/samples)`, with
`rate=k/n`. Rounding uses integer counts and changes rate by at most `1/(2*n)`.
Each entry links this script, report, and result file and records sample count,
confidence, and interval. Previously recorded paper evidence remains separate.
All capacities remain **estimated**, and all target encoding statuses remain
**candidate**. The estimates enable provisional layout planning, not execution.

Intervals quantify sampling error under the stated model, not possible bugs in the
membership test, rate characterization, or generator. Evidence for correctness
comes separately from the independent residue and actual-curve oracles on small
instances. No exact target dimension, systematic information set, scalable encoder,
application storage rate, cryptographic security, or terabyte-scale implementation
has been established by this experiment.
