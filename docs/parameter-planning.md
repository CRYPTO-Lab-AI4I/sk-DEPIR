# Parameter planning

This is a proposed planner design, not an implemented API. It selects code parameters
and an [input layout](data-representation.md) together. It is separate from the
[application encoder](encoder-contract.md#application-encoding): it produces plans, not
encoded data. No default field has been selected; the initial security profile is
specified below. Development order is in
[ROADMAP.md](../ROADMAP.md#implementation-sequence).

## Inputs, objective, and output

**Inputs:** application-record count and size, desired security level, and optional
limits on encoded storage and upload/download.

**Objective:** choose a [partition into PIR databases](data-representation.md) that
minimizes total encoded storage while meeting the desired security level and storage
and communication limits. Use total communication to break ties, and report alternatives when storage and
communication trade off.

**Output:** one or more partition plans. Each contains the submatrix partition, a
configuration for each PIR database drawn from the [configuration
catalogue](#configuration-catalogue), [query-sharing
groups](data-representation.md#query-sharing-groups) (databases using one query), the
effective storage rate (application bits divided by stored bits), and storage and
communication costs. Optional encoding-time and memory estimates are informational
outputs only: they do not affect eligibility, ranking, or tie-breaking between
partitions. Distinguish estimates, exact calculations, and measured results,
linking the evidence for each. Mark plans that still need encoder validation as
provisional. The caller passes a chosen plan to the [application
encoder](encoder-contract.md#application-encoding).

## Configuration catalogue

The catalogue contains both candidate configurations used for planning and
configurations supported for encoding. A candidate may have an empirical capacity
estimate; execution requires implemented field arithmetic and an encoder with verified
capacity. Record this status separately from the rate evidence.

Describe each configuration by `(s, m, t)`, where `s` is the field-element width, `q =
2^s`, `m` is the evaluation-space dimension, and `t` is the query-curve degree. Each
entry includes:

- Input capacity `K` and encoded length `N = q^m`, both in field elements,
  and code rate `K/N`, with the empirical estimate or verified capacity and
  supporting encoding evidence.
- Field polynomial, [code-preprocessing method](encoder-contract.md#code-preprocessing),
  and encoder implementation when available, with reproducible encoding evidence.
- Transmitted query length `L` in encoded positions, including extra random
  positions added to hide the curve (dummy positions), and the rule for their count.
- Estimated security level, supporting analysis, and assumptions.
- Available runtime and memory measurements, linked to their reports; separate
  reusable code preprocessing from database encoding and record artifact size.

The planner estimates a layout's security level as the minimum of its selected
configurations' security levels, assuming the documented masking and query rules are
followed. Security assumptions belong to catalogue entries, not user inputs.

Include `GF(2^s)` for every integer `s` from 8 through 20, with `m=2`. The initial
profile targets at least 100 bits of conjectured computational security against an
adversary observing at most `2^50` samples, following the bounded-sample setting in
[Table 1](../paper.pdf#page=64). A sample is one observed protocol query. Use `ell=q-1`
curve positions and `L=ell+128` total positions, as in Section 7.

For `m=2`, [Conjecture 6.1](../paper.pdf#page=38) gives the sample bound `S_max =
min(2^(s*(t-1)), t^(t^2-1))`. Selecting the smallest `t>2` with `S_max >= 2^50` gives:

| Field degree `s` | Curve degree `t` |
|---|---|
| 8 | 8 |
| 9 | 7 |
| 10, 11, 12 | 6 |
| 13, 14, 15, 16, 17, 18, 19, 20 | 5 |

Thus `t=5/6` covers degrees 10 through 20; degrees 8 and 9 require higher `t` for the
same sample budget. At `s=10,t=6`, the sample bound equals `2^50` exactly, with no
additional margin. These choices match Table 1's bounded-sample entries at
`s=12,15,18,20`. The [calculation
report](../experiments/bounded-sample-parameters/report.md) provides results for every
degree and checks the computational bound against `2^100`. These are conjectured bounds
for the paper's assumption, not a security validation of the current implementation.
Lower `t` values will be considered later as lower-security alternatives.

Choose layouts by capacity, code rate, padding, and communication costs.
Byte alignment does not restrict the catalogue. Each degree needs implemented field
arithmetic and validated encoder capacity before use. Current arithmetic is in
[gf2n.go](../utils/gf2n.go). The existing [Params type](../pir/params.go) names `q`,
`m`, and curve degree `t` as `Q`, `M`, and `K`; its `K` is not the message capacity used
in this document.

Table 1 reports 712 GB of plaintext at `s=20,m=2,t=5`, with code rate about 0.259 and
2.75 TB of encoded storage. Two full blocks therefore hold about 1.424 TB of plaintext
using about 5.50 TB of encoded storage (decimal units). This motivates the
terabyte-scale application target; partitioning and padding reduce usable capacity.
These are paper estimates, not capacities verified by this library's encoder. Validate
implementation resource costs before claiming support at this scale.

Monte Carlo rate estimates can be highly accurate and are useful for comparing
configurations and planning layouts before an encoder is available. Record the sampling
method, sample count, uncertainty, and reproducible evidence. An estimate does not
establish the exact message capacity required for encoding.

Distinguish three kinds of rate evidence:

- **Estimated capacity and rate:** empirical results, including Monte Carlo
  estimates, used to plan layouts whose capacities still require verification.
- **Exact calculated rate:** with verified capacities and a specified layout,
  calculate `K/N` and the application storage rate using the model below. No
  previous encoding of that identical application layout is required. State
  which storage overheads the calculation includes.
- **Measured storage usage:** actual bytes written by a documented encoding,
  including the implemented format's overheads.

Catalogue rates should ultimately be backed by actual encoding evidence. Follow the
[experiment conventions](../experiments/README.md) for empirical estimates and
measurements, identifying the code specification and measurement boundaries.

## Storage and communication model

Use the partition and query-sharing rules defined in [data
representation](data-representation.md). The single-database read/response model comes
from [Figure 5](../paper.pdf#page=33); the sums below account for the library's
partitioned layout.

Let the application matrix contain `n` records of `A` bits each. For each PIR database
`d`, define:

| Symbol | Meaning |
|---|---|
| `u_d`, `b_d` | Submatrix rows and bit columns before padding |
| `K_d` | Input-row capacity from catalogue evidence; verified before encoding |
| `s_d` | Field-element width and padded PIR-record size in bits |
| `N_d` | Number of encoded field elements |

Require `u_d <= K_d` and `b_d <= s_d`. Append `K_d-u_d` zero rows and `s_d-b_d` zero
columns before encoding. A complete partition satisfies `sum_d u_d*b_d = n*A`, with no
overlap or uncovered entries.

For query-sharing group `g`, let `L_g` be the transmitted position count, including
dummies, `J_g` its number of PIR databases, and `N_g` its common encoded
address-space size. Under full codeword storage,
tightly packed field elements, and explicit-address queries:

| Quantity | Bits |
|---|---|
| Total encoded storage `S` | `sum_d N_d*s_d` |
| Download per application-record retrieval `D` | `sum_g L_g * sum_(d in g) s_d` |
| Upload per application-record retrieval `U` | `sum_g J_g*L_g*ceil(log2(N_g))` |

The effective storage rate is `n*A/S`; unlike catalogue code rate `K/N`, it includes
partitioning and padding. Download counts every database, including irrelevant
responses. Upload counts a separate position list for every PIR database, including
repeated positions within each list. Query sharing saves client computation, not
transmitted addresses. Use the actual `L_g`, not just the
curve length `q_g-1`, where `q_g` is the group's field size.

The formulas include zero-row and zero-column padding through the chosen configurations.
Add metadata, database identifiers, byte alignment, transport framing, and
implementation resource overheads separately. A group with `J` databases using one
configuration has storage `J*N*s`, download `L*J*s`, and upload `J*L*ceil(log2(N))`. These counts describe logical addresses and data;
backend request headers, byte-range alignment, and other transport overheads must
be added for actual remote-storage traffic. The minimal logical address width
`ceil(log2(N_g))` in this model is distinct from the library's uint64 position
representation; count actual serialized widths in any adapter-specific report.

## Selection procedure and output

1. Enumerate catalogue configurations, submatrix partitions, and query-sharing
   groups. Check the [grouping conditions](data-representation.md#query-sharing-groups)
   for every application row before accepting the query-sharing groups. Reject
   layouts that fail coverage, capacity, compatibility, the desired security
   level, or storage and communication limits; label estimate-based plans provisional.
2. Compute total storage and communication. Remove choices that are no better
   in either metric and worse in at least one. Rank the remaining plans by the
   stated objective.
3. Optionally attach encoding-time and memory estimates from existing measurements
   to the selected plans, without changing the selection. Separate reusable code
   setup from per-database encoding and state whether setup is already available.
   The planner does not run either operation. Existing benchmarks are in
   [pir_test.go](../pir/pir_test.go); their measurement boundaries and RM-placeholder
   limitations are described in the [README](../README.md#benchmarks). If suitable
   measurements are unavailable, leave the estimates unspecified.
4. Return the [specified output](#inputs-objective-and-output), with supporting
   evidence, catalogue assumptions, and the reason for selection.

Smaller submatrices may reduce padding or permit cheaper configurations, but all
resulting databases contribute download. A layout may also change how many queries can
be shared. Compare these effects across the entire collection.

If no validated candidate meets the constraints, report that explicitly. Do not silently
lower the desired security level or treat empirical dimension estimates as verified
encoder capacity. Provisional plans may still be returned for exploration, with the
required verification stated explicitly.
