# Parameter planning

The proposed planner selects code parameters and an [input layout](data-representation.md)
together. It is separate from the [application encoder](encoder-contract.md#application-encoding):
it produces plans, not encoded data. The [PIR configuration catalogue and loader](../pirconfig/README.md)
and [supplied-plan evaluator](../planner/README.md) are implemented. Automatic
partition search includes [bounded tradeoff search](../planner/README.md#storage-and-communication-tradeoffs)
and a [single-choice greedy heuristic](../planner/README.md#greedy-planning).
No default field has been selected; the initial security profile is specified below.
Development order is in [ROADMAP.md](../ROADMAP.md#implementation-sequence).

## Inputs, objective, and output

**Inputs:** application-record count and size, desired security level, and optional
limits on encoded storage and upload/download.

**Objective:** find storage–download tradeoffs for [partitions into PIR databases](data-representation.md)
that meet the security and any supplied cost limits. Retain nondominated alternatives:
none should have both higher storage and higher download than another
feasible candidate. Users can inspect choices before setting communication limits.
The single-choice greedy API instead minimizes storage with download as a tie-break.

**Output:** one or more partition plans. Each contains the submatrix partition, a
configuration for each encoded PIR database drawn from the [configuration
catalogue](#configuration-catalogue), [query-sharing
groups](data-representation.md#query-sharing-groups) (databases using one query), the
effective storage rate (application bits divided by stored bits), communication rate
(requested application-record bits divided by upload plus download per lookup), and storage and
communication costs. Optional encoding-time and memory estimates are informational
outputs only: they do not affect eligibility, ranking, or tie-breaking between
partitions. Distinguish estimates, exact calculations, and measured results,
linking the evidence for each. Mark plans that still need encoder validation as
provisional. The caller passes a chosen plan to the [application
encoder](encoder-contract.md#application-encoding).

## Configuration catalogue

The versioned [JSON format](../pirconfig/schema.json) and [Go loader](../pirconfig/catalogue.go)
represent both candidate configurations used for planning and
configurations supported for encoding. A candidate may have an empirical capacity
estimate; execution requires implemented field arithmetic and an encoder with verified
capacity. Record this status separately from the rate evidence.

The [PIR configuration catalogue](../pirconfig/configs_catalogue.json) contains the
agreed security profiles and [Monte Carlo rate estimates](../experiments/liftrs-rates/report.md)
for every target configuration. The four matching paper estimates remain as
supporting evidence. The s=8, s=9 and s=10 entries additionally record exact capacities K=3121, K=17200 and K=103556
and bounded standalone encoding support; the other capacities remain provisional. An application plan
still needs a compatible execution backend, beyond catalogue metadata alone.
The [verified tiny fixture](../pirconfig/testdata/reference.json) exercises the format
without claiming security.

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

Choose layouts by capacity, code rate, padding, and download costs. Upload remains
reported and constrained when requested, but is not the optimization objective.
Byte alignment does not restrict the catalogue. Each degree needs implemented field
arithmetic and validated encoder capacity before use. Legacy client arithmetic is in [gf2n.go](../utils/gf2n.go); the separate
[table-free arithmetic](../encoding/liftrs/binary_field.go) covers every s=2..20
without implying a full encoder at all those sizes. The existing [Params type](../pir/params.go) names `q`,
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

Let the application matrix contain `n` records of `A` bits each. The following sums
first describe the identity logical layout; split/stack adjustments follow below.
For each encoded PIR database `d`, define:

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

With logical width `W`, let `f=ceil(A/W)`. Apply the sums to the logical matrix
`(n*f) x W` to obtain indexed storage `S_P`, upload `U_P`, and download `D_P` per
logical lookup. If full-download pieces contain `T` bits in total, costs per original
application lookup are `S=S_P+T`, `U=f*U_P`, and `D=f*D_P+T`.
Full pieces have identity storage and no index-dependent upload; fixed read descriptors
and framing remain excluded. Fetch them once per original lookup. The partition area
is `n*f*W`, but the rate numerator remains `n*A`, excluding fragment padding.

The effective storage rate is `n*A/S`; unlike catalogue code rate `K/N`, it includes
partitioning and padding. Download counts every database, including irrelevant
responses. Upload counts a separate position list for every PIR database, including
repeated positions within each list. Query sharing saves client computation, not
transmitted addresses. Use the actual `L_g`, not just the
curve length `q_g-1`, where `q_g` is the group's field size.

The [storage layout optimization goal](protocol-io.md#storage-layout-optimization)
places same-position values from each query-sharing group in one readable group
record. With that layout, logical indexed upload would be
`U_P = sum_g L_g*ceil(log2(N_g))`, eliminating the factor `J_g`; storage and
download are unchanged before byte alignment and metadata. This is a target model,
not the currently implemented evaluator. Report download communication rate `A/D`
(and its inverse `D/A` for cost comparisons) separately from total communication
rate `A/(U+D)`. Search now optimizes storage and download, including its dominance checks and
approximation coordinates. Upload and total traffic remain separately reported.

The formulas include zero-row and zero-column padding through the chosen configurations.
Add metadata, database identifiers, byte alignment, transport framing, and
implementation resource overheads separately. A group with `J` databases using one
configuration has storage `J*N*s`, download `L*J*s`, and upload `J*L*ceil(log2(N))`. These counts describe logical addresses and data;
backend request headers, byte-range alignment, and other transport overheads must
be added for actual remote-storage traffic. The minimal logical address width
`ceil(log2(N_g))` in this model is distinct from the library's uint64 position
representation; count actual serialized widths in any adapter-specific report.

## Evaluating a supplied plan

The implemented [planner.Evaluate](../planner/README.md) validates a supplied
partition and calculates the model above. It accepts estimated capacities and marks
such plans provisional; unknown capacities are rejected. Requirements specify the
minimum catalogue security profile and optional storage/upload/download bit limits.
All arithmetic is checked for `uint64` overflow. The result retains used capacity
and security evidence; encoding-time and memory measurements have no effect.

A query-sharing group references one catalogue configuration ID. This fixes its
padded dimensions and code/query conventions; equivalent aliases are not merged.
The [row-range conditions](data-representation.md#query-sharing-groups) still apply.
Coverage checks operate on rectangles, not an application-sized matrix. The current
pairwise checks are quadratic in the number of rectangles.

This evaluator does not generate or rank plans, invoke preprocessing, or establish
backend availability. The greedy planner below generates plans and invokes it.

## Greedy partitioner

[planner.Greedy](../planner/greedy.go) implements row-aligned partitioning with
one-step lookahead. At each row position, try every eligible configuration, slice
that block vertically, and score it together with a greedy completion. The
completion fills the largest-fitting capacities and uses the cheapest covering
configuration for the last remainder. Score repeated blocks in aggregate. Choose
minimum full-plan storage, breaking ties by download; enforce all cost
and security limits. Estimated capacities remain provisional.

The returned plan represents column slices implicitly in full-width row blocks.
The evaluator counts every database and query copy without expanding them. General
explicit partitions remain supported. Homogeneous row blocks simplify grouping by
configuration ID but restrict the search space. No global or row-block optimality
is claimed; `ErrNoPlan` means no feasible completion was found by this heuristic.
The [cost-grid report](../experiments/greedy-costs/report.md) records the unsliced baseline costs.

The broader selection procedure below describes possible future search strategies.

## Application width search

`GreedyApplication` tries the original width plus requested logical widths, or, by
default, all catalogue field widths and powers of two starting at 8 bits up to the
original width. For each width it runs the same row-block heuristic, charging indexed
traffic for all fragments. Choose minimum storage, then download, then
larger width for exact ties. This expands the search without arbitrary rearrangement
or an optimality guarantee. `EvaluateApplication` validates the mapping and costs.

Full-download suffixes are opt-in through `AllowFullDownload`. Set
`FullDownloadRatio` to bound their aggregate bits by a multiple of download from the
encoded parts, per original lookup. Zero selects the default **1**: added full-download
traffic cannot exceed encoded download. If there are no encoded parts, compare with
the indexed-only largest-fitting baseline for that width instead. A baseline is required.
Additionally cap full-download bits against `FullDownloadRatio` times the lowest
indexed download among no-full greedy plans over **all** searched widths, including
the original width. Compute this common reference using the security requirements
but without cost budgets, so changing width cannot inflate its own allowance.
`MaxFullDownloadBits` optionally adds an absolute cap; storage/upload/download budgets
still apply. Thus unrestricted identity storage cannot automatically win on storage.
The heuristic considers at most one terminal full-download row block, not arbitrary
collections of full pieces. Ratio checks constrain selection, not supplied-plan validity.

The `2^50` limit counts protocol queries. With `f` fragments, each indexed group uses
`f` queries per application lookup, so at most `floor(2^50/f)` such lookups fit that
sample budget. Full-only layouts need no indexed-query security profile.
No new execution backend or large-instance support is implied by these planner APIs.

## Bounded tradeoff search

[`SearchTradeoffs`](../planner/README.md#storage-and-communication-tradeoffs) processes
configurations in descending capacity order. Choose a batch of full blocks for each
configuration, then pass remaining rows to smaller configurations. This removes
redundant permutations of disjoint full blocks without discarding small residual codes.
A final padded block may use any configuration, or the suffix may be downloaded whole.

Prune only plans with the same remaining rows at the same configuration stage.
Default limits are 32 frontier points, 256 total live partial plans per stage, 8 count
choices per transition, and 256 constructed blocks per greedy baseline. If all count
choices fit, enumerate them; otherwise sample zero, one, filling counts and powers
of two near either endpoint. State overflow samples remaining-row keys and cost endpoints.
Repeated blocks have compact metadata, but every physical database still contributes
storage and traffic. See the API guide for validation and precise limit semantics.

The result reports frontier, state, count and baseline truncation. There is no row-boundary
admission cap in this search. Approximation can miss better or feasible plans; no global
or multiplicative guarantee is claimed. Completed feasible baselines bypass lossy
reduction and retain equal-or-better witnesses. Output and display targets may be
exceeded to preserve them. The [tradeoff report](../experiments/tradeoff-costs/report.md)
records the revised method and retains the previous boundary-search results for comparison.

Communication rate is `A/(U+D)` for one original record of `A` bits; higher is better.
It is not `(U+D)/A`, and not communication divided by total database size. Whole-database
download of `n` records has rate `1/n`. Present rate alongside actual per-lookup bytes,
storage, and approximation flags rather than requiring a rate constraint upfront.

## Cost-indexed DP alternative

[`SearchDPCosts`](../planner/README.md#cost-indexed-dynamic-programming) instead records
maximum row coverage by rounded storage/download coordinates. Binary groups
represent all full-block multiplicities, followed by an exact padded or full-download
completion. This changes the search approximation, not layout semantics or cost formulas.

DP uses explicit per-width numerical windows, with a grid per download scale.
Results retain exact bits, grid units and item/transition counts. Defaults are 64 cells
per axis; the [dp-costs report](../experiments/dp-costs/report.md) uses 128. Cost rounding
and finite exploration windows can still miss superior plans; an item-wise additive
rounding bound is not a global approximation theorem. Completed greedy witnesses stay
protected. Download and storage budgets are supported; upload budgets require another DP
dimension and are rejected by this initial API.

## Selection procedure and output

1. Enumerate catalogue configurations, submatrix partitions, and query-sharing
   groups. Check the [grouping conditions](data-representation.md#query-sharing-groups)
   for every logical row before accepting the query-sharing groups. Reject
   layouts that fail coverage, capacity, compatibility, the desired security
   level, or storage and communication limits; label estimate-based plans provisional.
2. Compute storage, upload and download; optimize storage and download. Remove choices that are no better
   in either metric and worse in at least one. Return the remaining tradeoffs; apply a storage-first ranking only after
   the caller selects communication constraints or explicitly requests that policy.
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
