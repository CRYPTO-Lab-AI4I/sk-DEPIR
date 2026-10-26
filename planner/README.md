# Greedy planner and partition evaluator

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
24 application-size/record-size combinations.

## Plans and validation

A `Plan` has `FormatVersion: 1`, application `Rows` and `RecordBits`, and exactly one of a list of
`Database` rectangles or compact `RowBlocks`. Each rectangle gives half-open row/column bounds, a catalogue
`ConfigurationID`, and a query-sharing `Group`. Its coordinates identify it within
the application. These Go types have JSON tags for reviewing planning data; they
are not a durable stored-database manifest or a strict JSON-loading API.

`Evaluate` checks:

- The catalogue is consistent and the application and rectangles are nonempty.
  Record width is at least the largest field width in the supplied catalogue,
  including estimated configurations. A caller can supply an intentionally scoped
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

For every database, add `N*s` storage bits, `L*ceil(log2(N))` upload bits, and `L*s`
download bits. `L` includes dummy positions and repetitions. Every database is read
on every application-record retrieval, including databases with no relevant piece.
Query sharing reuses client computation; it does not reduce these bit counts.

`Costs` also records application bits and total communication. `EffectiveStorageRate`
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
