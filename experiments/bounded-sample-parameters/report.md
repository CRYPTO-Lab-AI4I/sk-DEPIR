# Bounded-sample parameter calculation

## Source and reproduction

This is an analytical calculation, not an encoding experiment or measured
security result. [run.py](run.py) generates [results.csv](results.csv) from
[Conjecture 6.1](../../paper.pdf#page=38), with the query lengths from Section 7
and comparison points from [Table 1](../../paper.pdf#page=64).

- Base revision: `c3ff381ad2be181915e09a6a630f2e66ddd53da8`; run with the
  DOC-035 script and user-replaced paper recorded alongside this report.
- Paper SHA-256: `205571fe7afb594d82d461a900bf0a5e2e210b8c688260224dbc6a5d9f0b8076`.
- Environment: Darwin arm64, Python 3.9.6; standard library only.
- Command from repository root: `python3 experiments/bounded-sample-parameters/run.py`.
- No input database, randomness, timings, or hardware performance measurements.

## Method

Fix `m=2`, enumerate all integers `s=8..20`, and set `q=2^s`.
Conjecture 6.1 gives

`S = min(q^(t-1), t^(t^2-1))`.

Select the smallest integer `t>2` for which `S >= 2^50`, using exact integer
comparisons. The CSV reports base-two logarithms of the two bounds and their
minimum, rounded to six decimals only for display. The CSV uses LF line endings. Equality is accepted because
the paper defines security for at most `S` samples.

For `ell=q-1`, `L=ell+128`, the script also checks both terms of the conjectured
computational bound `T=min(binomial(L,ell),q^(m*L))` against `2^100`, using
an exact binomial coefficient and the integer exponent `m*L*s`. All pass.
This checks the 100-bit target; it does not reproduce the larger bit counts
printed in Table 1 or claim those values as library security levels.

## Results and interpretation

- `s=8` requires `t=8`; `s=9` requires `t=7`.
- `s=10,11,12` require `t=6`.
- Every `s=13..20` requires `t=5`.
- At `s=10,t=6`, `S=2^50` exactly; there is no sample-budget margin.
- The choices at `s=12,15,18,20` match Table 1's bounded-sample entries.

Table 1 separately reports 712 GB plaintext and 2.75 TB encoded at
`s=20,m=2,t=5`, with rate about 0.259. Doubling these rounded paper values
produces 1.424 TB plaintext and 5.50 TB encoded, in decimal units. These are
paper-derived capacity estimates, not outputs of this script or an implemented
encoder. Actual layout capacity also depends on partitioning and padding.

## Limitations

The security bounds are conjectural and use the paper's constant distinguishing
advantage threshold of `1/e`. They do not validate implementations, cryptographic
key sizes, randomness, query sharing, or use beyond the sample budget. Encoder
capacity, runtime, memory, and usable layout rate still require their own evidence.
Lower `t` values do not satisfy this initial sample profile and are deferred as
lower-security alternatives. No rate is inferred for the other field degrees.
