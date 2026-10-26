# Lift-RS rate estimator

`pir-rate` estimates code rate by sampling reduced exponent tuples and checking
monomial membership. It uses [liftrs.GoodMonomial](../../encoding/liftrs/monomial.go),
not code preprocessing or database encoding. No field polynomial is needed.

From the repository root:

```sh
go run ./cmd/pir-rate \
  -catalogue pirconfig/configs_catalogue.json \
  -samples 1000000 -seed 20261026 -confidence 0.99
```

The command emits JSON containing counts, rates, confidence bounds, per-code seeds,
input SHA-256, Go/platform information, and sampling-only elapsed seconds. It does
not modify the catalogue. The [experiment script](../../experiments/liftrs-rates/run.py)
adds revision/source hashes and saves reproducible results with a
[report](../../experiments/liftrs-rates/report.md).

Each tuple contains `m` independent uniform integers in `[0,2^s)`, sampled with
replacement. The seed for a code is `root_seed + 1000*s + 100*m + t`, independent
of catalogue order. Go's explicitly seeded `math/rand.NewSource` is used for
reproducibility, not for protocol randomness.

`-confidence` is the simultaneous confidence target across all `C` configurations.
For `n` samples each, use `alpha=(1-confidence)/C` and the two-sided Hoeffding
half-width `sqrt(log(2/alpha)/(2*n))`, clipped to `[0,1]`. Each result's `confidence`
is `1-alpha`. The union bound does not require independence between configurations;
the sampling bound assumes independent uniform samples within each configuration.
The pseudorandom generator approximates that model; the intervals are conservative
statistical bounds, not code-dimension or security guarantees.

Supported inputs follow the membership test: `s=2..20`, `m=2..3`, `1<=t<q-1`.
Runtime grows with `samples*s*m*t`; high curve degrees can be expensive. Default
arguments match the command above. Sample count must be in `[1,2^53]`; confidence
must be strictly between zero and one. Errors produce a nonzero exit status.
