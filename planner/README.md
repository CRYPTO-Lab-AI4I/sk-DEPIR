# PIR planning and cost evaluation

This package constructs greedy row-block partitions and evaluates supplied general
partitions against the [PIR configuration catalogue](../pirconfig/README.md). It
calculates logical storage and communication without preprocessing codes, encoding
data, or generating queries. Greedy planning is a heuristic, not an optimality claim.

## API

Import `github.com/CRYPTO-Lab-AI4I/sk-DEPIR/planner`. Supply a catalogue loaded with
`pirconfig.Load`, a `Plan`, and optional `Requirements`:

```go
result, err := planner.Evaluate(plan, catalogue, planner.Requirements{
    MinSecurityBits: 100,
    Log2Samples:     50,
})
if err != nil {
    return err
}
// result.Costs gives logical bit counts; result.Provisional identifies plans
// whose capacity or encoder still needs validation.
```

The [runnable example](example_test.go) constructs and evaluates a complete tiny
plan. For its three 4-bit application records, split columns into two GF(4)
databases using one query-sharing group. The evaluator returns 64 storage bits,
1,048 upload bits and 524 download bits per record retrieval, with effective rate
`12/64 = 0.1875`. The example makes no cryptographic security claim.

## Storage and communication tradeoffs

Use `SearchTradeoffs` to discover achievable choices before setting a communication
budget. It returns nondominated plans rather than choosing a storage-first winner:

```go
result, err := planner.SearchTradeoffs(rows, recordBits, catalogue, requirements,
    planner.TradeoffOptions{MaxFrontier: 32, MaxStates: 256, MaxCounts: 8})
// result.Alternatives contains plans and evaluations, ordered by storage.
// Evaluation.CommunicationRate = recordBits / (uploadBits + downloadBits).
```

Communication rate is the fraction of communicated bits that corresponds to the
requested application record: **higher is better**. It differs from storage rate
(application database bits / stored bits). Whole-database download has communication
rate `1/rows`. Costs always use upload plus download per original record lookup.

For each logical width, process configurations in descending capacity order (ties
by ID). Choose how many full blocks of each code to use, batching repeated blocks
rather than exploring permutations of them. Small codes remain available for residuals.
At every stage, finish using either a whole remaining suffix download or one padded
block of any configuration. The final padded block is not forced into capacity order.
Costs remain additive because all full-width row blocks are disjoint and grouping
saves client computation, not communication.

At the same configuration stage and remaining row count, prune dominated partial
plans. Compare storage/total communication normally, or storage/upload/download for
directional budgets. The defaults are:

- `MaxFrontier=32`: retained plans per remaining-row state and completed exploratory set.
- `MaxStates=256`: total live partial plans per configuration stage. This replaces
  the old historical boundary-admission cap, rather than exploring small row positions first.
- `MaxCounts=8`: count choices per state/configuration. Enumerate all when they fit;
  otherwise sample zero, one, filling counts, and powers of two from both ends.
- `MaxBaselineBlocks=256`: maximum blocks constructed for each greedy baseline.

Zero selects defaults. `MaxFrontier` must be at least 2, `MaxCounts` at least 4,
other limits positive. Repeated runs store a `RowBlock.Repeat` count: split the range
into that many equal-height consecutive blocks; zero means one. A final shorter padded
block is separate. Each repetition still has separate databases, masks and full traffic.
Full-download blocks cannot repeat. Compact cost evaluation does not expand repeats.

Frontier overflow uses geometric communication buckets with cost endpoints. Stage
state overflow samples remaining-row keys evenly and alternates their cost endpoints.
These approximations have no multiplicative error guarantee. `FrontierTrimmed`,
`StateLimitReached`, `CountChoicesTrimmed`, and `BaselineLimitReached` report losses;
`StatesExplored` counts processed partial plans across stages and widths. Report these
flags alongside costs. Without loss, exactness applies only to this layout family.

Whole download and completed feasible greedy baselines are protected separately.
Merge them after approximate exploration and apply exact dominance pruning only.
`BaselineWitnesses` links each to an equal-or-better final alternative; strict
improvement is not promised. The returned set can exceed `MaxFrontier`. Interrupted
baselines are outside the guarantee. `RepresentativeIndices(3)` preserves all witnesses
and endpoints, so its target size can also be exceeded. `ErrNoPlan` returns nonnil
search metadata if approximation may have missed feasibility. This API does not infer
a default communication budget or use full-download ratio exceptions.

[Tests](tradeoffs_test.go) compare untrimmed results with independent exhaustive tiny
enumeration, verify bounds, determinism, budgets and actual target coordinates. The
[21-case report](../experiments/tradeoff-costs/report.md) records the bounded search's
alternatives and limitations. The [runnable example](example_test.go) shows the API.

## Cost-indexed dynamic programming

`SearchDPCosts` is a separate algorithm; it does not replace `SearchTradeoffs`.

```go
result, err := planner.SearchDPCosts(rows, recordBits, catalogue, requirements,
    planner.DPCostOptions{StorageCells: 128, CommunicationCells: 128})
indices, err := result.RepresentativeIndices(3) // retains protected witnesses
```

For each logical width, binary decomposition represents every full-block count from
zero through `floor(rows/K)`, without sampling counts. Each item is a group of blocks
that can be selected once. The DP stores maximum covered rows per rounded storage and
communication cell at each item stage; equal coverage favors smaller exact costs.
Overshooting groups can finish using fewer blocks and one padded final block. Full
residual downloads and final padded blocks use exact costs. Compact repeated runs are
coalesced before independent evaluation.

Defaults are 64 cells per axis, 32 completed exploratory alternatives, and 256 blocks
per greedy baseline. Axis limits are 1..512 cells. `StorageWindowBits` overrides the
per-width exploration window; otherwise use twice the storage of the least-communication
one-code or greedy seed. `CommunicationWindowBits` overrides a single communication
window; otherwise solve geometrically spaced windows (factor 4) from seed communication
through application size. This keeps low-traffic precision without using database size
as the only cost scale. Whole-download, completed greedy, and one-code seeds are also
considered independently of DP windows. Only greedy/full seeds receive protected witnesses.

Each axis unit is `ceil(window/cells)` bits. Full-block group costs are rounded down
per item; exact costs are retained and checked against window/storage requirements.
`Windows` reports limits, units, item counts and transitions. A path's rounding loss
is less than `BinaryItems*UnitBits` per axis; this is an arithmetic accounting bound,
**not** a global Pareto approximation guarantee. With unit one, DP labels are exact
within the stated windows and layout family. With larger units, same-cell replacement
can lose tradeoffs or feasible completions. Output frontier reduction is also approximate.

For `M` binary items and `S,C` cells, the dense transition pass costs `O(M*S*C)` per
window. The grid has `(S+1)*(C+1)` slots; immutable reconstruction links can require
additional memory. There is no row-boundary or count-sampling cap in this algorithm.
The two-dimensional API explicitly rejects separate upload/download budgets because
supporting them correctly requires additional state dimensions. Exact storage budgets
remain supported. Output costs and communication rate retain the ordinary evaluator's
meaning, and protected baselines can exceed the output-size target.

See [DP tests](dp_test.go), [runnable example](example_test.go), and the independent
[dp-costs experiment](../experiments/dp-costs/report.md). These remain planning calculations,
not target-sized encoded databases or a validated production implementation.

## Greedy planning

```go
plan, result, err := planner.Greedy(rows, recordBits, catalogue, planner.Requirements{
    MinSecurityBits: 100,
    Log2Samples:     50,
})
```

For each next row block, `Greedy` tries every eligible configuration, covers up to
its capacity, and scores that block plus a greedy completion of the remainder.
The completion repeatedly fills the largest-fitting row capacity, batching repeated
blocks using division, and covers the last remainder with the cheapest covering
configuration. Candidates must satisfy the full-plan budgets. Choose minimum total
storage, breaking ties by total upload plus download, then greater next-block height
and configuration ID. Recompute lookahead after each chosen block. Catalogue order
does not affect the result. Unknown capacities and insufficient security profiles
are excluded; estimated capacities produce provisional plans.

The result uses compact `RowBlocks`: each covers all application columns and implies
`ceil(recordBits/s)` separate PIR databases. Slice `j` has columns
`[j*s,min((j+1)*s,recordBits))`. Only row blocks are allocated, even for millions of
column slices. `Evaluation.DatabaseCount` counts all physical PIR databases. This
compression changes metadata only: each database still has its own masks and full
address upload. Group databases by configuration ID across all blocks; row ranges
are identical within a block and disjoint between blocks.

For `B` selected blocks and `C` configurations, lookahead costs `O(B*C^2)`, followed
by `O(B^2)` compact evaluation. Memory holds block/catalogue metadata, not one entry
per column slice. Time and output still grow with the number of row blocks.
`ErrNoPlan` means this heuristic found no feasible completion, not proof that no
partition exists. It searches homogeneous row blocks and does not promise even the
optimum within that restricted family. Encoding resource measurements do not affect
selection. The [cost-grid experiment](../experiments/greedy-costs/report.md) reports
21 application-size/record-size combinations.

## Application width search

Use `GreedyApplication` to split large records into narrower logical rows before
partitioning. It preserves the original `Greedy` API and behavior.

```go
plan, result, err := planner.GreedyApplication(rows, recordBits, catalogue, requirements,
    planner.ApplicationOptions{
        AllowFullDownload: true,
        FullDownloadRatio: 1, // also the default when zero
    })
```

Without explicit `LogicalRecordBits`, search the original width, catalogue field
widths, and powers of two from 8 bits through the original width. An explicit list
replaces these alternatives; the original width is always included. Rank storage,
then communication, then larger width. `ApplicationPlan.Fragment(j,i)` maps fragment
`i` of record `j` to logical row `j*ceil(A/W)+i` and its source bit range. Zero-pad
the final fragment; narrow logical rows may also need field padding.

`AllowFullDownload` enables one terminal row block stored without code expansion.
Its entire contents are downloaded once per original lookup. The default ratio 1
limits those bits to indexed download in the selected plan; full-only plans instead
compare with the indexed-only largest-fitting baseline for that width. This prevents large
whole-database downloads from winning solely on storage. `FullDownloadRatio` must be
finite and positive (zero means default); `MaxFullDownloadBits` adds an optional
absolute cap. A shared cap also uses the lowest indexed download among no-full
greedy plans across all searched widths, including the original width. Baselines
use security requirements without cost budgets. Ordinary budgets also apply. No arbitrary cutoff or server computation
is introduced.

Call `EvaluateApplication` for a reshaped plan: indexed upload/download repeat
`f=ceil(A/W)` times, while full pieces are fetched once. Storage includes all padding;
application bits and effective rate use the original input size. `Evaluate` alone
costs one logical lookup. `Log2Samples` counts protocol queries, so `f` queries per
group per original lookup consume the sample allowance `f` times faster.

These APIs plan and evaluate layouts; they do not extend the runtime wrapper or
implement full-download cryptographic storage. [Tests](application_test.go) reconstruct
split records through the existing reference and check costs, ratio bounds, and budgets.
See the [runnable example](example_test.go).

## Plans and validation

A `Plan` has `FormatVersion: 1`, application `Rows` and `RecordBits`, and exactly one of a list of
`Database` rectangles or compact `RowBlocks`. Each rectangle gives half-open row/column bounds, a catalogue
`ConfigurationID`, and a query-sharing `Group`. Its coordinates identify it within
the application. These Go types have JSON tags for reviewing planning data; they
are not a durable stored-database manifest or a strict JSON-loading API.

Set `DownloadAll` on a rectangle or row block for a full-download piece: omit its
configuration ID and leave group zero. A full-download row block is one matrix, not
implicit field slices. `FullDownloadBits` reports their total area;
`FullDownloadOnly` means no indexed security profile applies (zero security fields
then are not a zero-bit computational security estimate). Full-only evaluation does
not assert a new runtime backend. The relative selection cap is an options policy,
not a restriction on supplied plans.

`Evaluate` checks:

- The catalogue is consistent and the application and rectangles are nonempty.
  Record width is at least the largest field width in the supplied catalogue,
  including estimated configurations (full-only plans are exempt). A caller can supply an intentionally scoped
  catalogue when evaluating a different field set.
- Rectangles are in lexicographic `(row_start,column_start,row_end,column_end)` order,
  stay within the application, and cover it exactly once. General partitions are
  allowed; row boundaries need not align across column portions.
- Each selected configuration has a known capacity. Rectangle height fits `K` and
  width fits `s`; the difference is zero padding.
- Group numbers start at zero in first-occurrence order. Members reference the same
  catalogue configuration ID, which fixes their padded/code/query profile. Every
  pair of row ranges is disjoint or has the same start; ending rows may differ.
  Configuration aliases are not merged automatically.
- Every selected security profile meets `MinSecurityBits` and `Log2Samples`.
  The result reports the minimum of each across the selected configurations.
- Optional `MaxStorageBits`, `MaxUploadBits`, and `MaxDownloadBits` are respected.
  Zero budgets mean unlimited. All products and totals must fit `uint64`; overflow
  is an error, never a wrapped cost.

Coverage uses pairwise overlap checks followed by an area check. Disjoint rectangles
inside the application cover it exactly once iff their summed area equals its area.
Checks are quadratic in explicit rectangle count, or in row-block count for compact
plans. Implicit slices are validated and costed in aggregate. There is no allocation
proportional to application bits or the implicit slice count. Very large explicit
partitions may need a future sweep algorithm.

## Cost model and evidence

For every indexed database, add `N*s` storage bits, `L*ceil(log2(N))` upload bits, and `L*s`
download bits. `L` includes dummy positions and repetitions. Every database is read
on every application-record retrieval, including databases with no relevant piece.
Full-download pieces add their area to both storage and download, with zero logical
index upload. Query sharing reuses client computation; it does not reduce these bit counts.

`Costs` also records application bits and total communication. `CommunicationRate`
is original record bits divided by total communication per original lookup. `EffectiveStorageRate`
is application bits divided by storage bits, rounded to `float64`. Counts are exact
for this logical model; byte alignment, identifiers, metadata, request framing, and
transport overheads are excluded. See the [full cost model](../docs/parameter-planning.md#storage-and-communication-model).

Estimated capacities are accepted for provisional planning. `Provisional` is true if
any selected capacity is not verified or its encoder is not declared supported.
Returned configuration assessments copy the capacity/security evidence from the
catalogue. Inputs are not changed and returned evidence does not alias them.
Independent calls are safe when callers do not concurrently mutate inputs.

A nonprovisional result still relies on the catalogue's claims: the application
encoder must check the backend and code, and security estimates remain conjectural.
Unknown capacities cannot size a partition. Resource measurements such as encoding
time and memory do not affect evaluation, and are not copied into the result.

## Verification

```sh
go test ./planner
go test -race ./planner
```

[Tests](evaluate_test.go) cover hand-calculated counts, exact budget boundaries,
mixed fields, security constraints, general partitions, group compatibility,
provisional evidence, ownership, concurrency, and overflow. They compare model
storage and read counts with actual [application-reference](../encoding/appref/README.md)
encoding/retrieval. [Greedy tests](greedy_test.go) compare small plans with a largest-fitting baseline
and an exact dynamic-programming lower bound, check deterministic choices and budgets,
and compare compact plans with fully expanded evaluation. A 10 MB record test counts
ten million databases without expanding them. A two-block target plan tests terabyte-scale coordinates using
estimated capacities, without encoding a database of that size.
