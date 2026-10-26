# Partition cost evaluator

This package evaluates a supplied application-to-PIR partition against the
[PIR configuration catalogue](../pirconfig/README.md). It validates the layout and
calculates logical storage and communication. Partition search remains to be
implemented; the evaluator never preprocesses codes, encodes data, or generates queries.

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

## Plans and validation

A `Plan` has `FormatVersion: 1`, application `Rows` and `RecordBits`, and a list of
`Database` rectangles. Each rectangle gives half-open row/column bounds, a catalogue
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
Checks are quadratic in rectangle count, with no allocation proportional to record
count or application bits. Very large numbers of rectangles may need a future sweep
algorithm; large application coordinates alone do not cause large allocations.

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
encoding/retrieval. A two-block target plan tests terabyte-scale coordinates using
estimated capacities, without encoding a database of that size.
