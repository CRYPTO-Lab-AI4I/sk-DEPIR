# Protocol I/O requirements

These are intended client/storage requirements, not an implemented API or wire format.
See the [README](../README.md) for implementation status and [data
representation](data-representation.md) for the submatrix layout and masks.

The single-database protocol is given in [Figure 5](../paper.pdf#page=33), with the
Lift-RS instantiation in [Section 5.2](../paper.pdf#page=34). The grouping and
reconstruction rules below extend it to the library's application layout; they are not
yet implemented by the legacy single-database API. The
[small application reference](../encoding/appref/README.md) implements fixed
participation, repeated byte-range reads, unmasking, and record reconstruction
for supplied tiny-instance layouts.

## Storage layout optimization

Reducing upload through the physical storage layout is in scope for TRL 5.
For each query-sharing group, store the field elements at the same encoded position
next to one another as a single group record. The client can then request that
record once per position and receive all group members' values through ordinary
reads. This removes repeated addresses across the group without server-side
protocol computation. Preserve the group's common permutation, independent masks,
fixed participation, and response ordering.

Implement a grouped layout and matching client storage adapter, documenting how
positions map to byte ranges. Measure upload including range descriptors, framing,
and byte alignment; do not assume it is negligible for every backend. The layout
must support streaming construction without requiring the whole group in memory.
The current reference reads and evaluator still use independently stored databases
and repeated uploads, as described below. Grouped reads are a goal, not an
implemented storage format. Server-side linear combinations remain outside scope.

## Query groups and participation

For each logical-fragment request, generate one position list per
[query-sharing group](data-representation.md#query-sharing-groups). For a group
`g`, the list contains `L_g` positions, including dummy positions that hide the curve.
The client sends a separate copy of the list for each PIR database in the group.
It requests every listed position, including repeated occurrences, through ordinary
storage reads. Storage need not know about query-sharing groups or repeat a query
across databases; it only returns the values addressed by the client.

Use the [application-to-PIR mapping](data-representation.md#query) to identify the
required PIR databases and their common row index within each query-sharing group. Query
that index; other members return irrelevant rows.

Keep group membership fixed for the encoded layout. Never split or merge groups, skip
databases, or change participation or message sizes according to which submatrices
contain the target.

For groups needing no data, use the same query-generation procedure at a valid local
index. The [planner](parameter-planning.md#configuration-catalogue) uses the minimum
selected configuration security level, under these query rules and the documented mask
separation.

**Existing code:** `GenerateClientQuery` in [pir-lifted.go](../pir/pir-lifted.go) uses
[curve generation](../utils/curve_generation.go) and `PrepareQuerySequenceWithNoise` in
[pir_api.go](../pir/pir_api.go). The latter appends and shuffles dummy positions using
`math/rand`; this is a prototype path, not a cryptographically secure sampler.

For a split/stack plan, perform the same fixed number `f=ceil(A/W)` of logical
lookups for every application-record request, including cover reads in every indexed
group. Each full-download piece is read entirely once per application request,
regardless of relevance; reuse those bytes locally across fragments. Do not assume
caching across requests. Fixed whole-piece reads need no indexed address list;
range descriptors, byte rounding, and transport overhead must still be counted by adapters.
The sample limit counts protocol queries: each indexed group consumes `f` samples per
application request. The existing reference wrapper does not automatically consume
`ApplicationPlan` or implement its full-download mode; [planner tests](../planner/application_test.go)
exercise fragment mapping and indexed retrieval through that reference.

## Storage response

For each query-list entry `i` of group `g`, database `d` returns one field element. The
client assembles the logical response as `response[g][i][d]`; grouping is client-side
bookkeeping, not a storage response format. Repeated positions retain one
response per occurrence; internal read deduplication must preserve that correspondence.
Every database also returns field elements for dummy positions.

The current single-database `QueryResult` and local read helpers are in
[pir_api.go](../pir/pir_api.go); they do not implement the application-level response assembly above.

## Client validation and reconstruction

Validate response shapes and field-element ranges. Missing, extra, malformed, or
unreadable data produces an error rather than silent padding or a partial application
record. These checks do not authenticate storage contents.

Retain private query state identifying which databases contain the requested record,
which query entries participate in decoding, and how to unmask them. Discard irrelevant
responses and unmask and decode the required PIR records. Reassemble the record using
the [layout reconstruction rules](data-representation.md#query).

The existing `decodeIndividualResponse` in [pir_api.go](../pir/pir_api.go) checks
response counts, selects real positions, unmasks, and combines values. It does not
implement all validation or application reconstruction required here. The separate
[cryptographic reference](../encoding/cryptoref/README.md) supplies XOR unmasking;
its [retrieval test](../pir/cryptoref_test.go) combines values with curve-point
multiplicity after ordinary storage reads. This is a small test pipeline, not
a selected large-scale I/O architecture.

## Storage adapter and errors

Use ordinary byte-range reads from the independently stored PIR databases. The
client adapter uses the [packed format](data-representation.md#layout-metadata-and-serialization)
to translate element `i` into bytes from `floor(i*s/8)` through
`ceil((i+1)*s/8)` (exclusive), then extracts its `s` bits. Reading extra bits in
those bytes does not change the logical response. Count byte alignment and
backend request overhead separately from logical communication.

Represent positions and dimensions as unsigned 64-bit values at the library
boundary. Validate bounds, arithmetic overflow, and conversion to backend offsets
before reading. A local adapter can use Go's [`io.ReaderAt`](https://pkg.go.dev/io#ReaderAt); remote adapters map
those ranges to their provider's ordinary read API. No custom server wire format
or query-sharing group identifier is required by storage. Backend transport and
authentication remain the adapter's responsibility.

Keep response values in the requested order, including one entry per repeated
position. Return Go errors for invalid input, incomplete or malformed responses,
and storage failures. Wrap underlying I/O errors so callers can inspect them with
[`errors.Is` and `errors.As`](https://pkg.go.dev/errors); do not return a successful partial application record.
Remote operations accept the caller's [`context.Context`](https://pkg.go.dev/context#Context) for cancellation and
deadlines. Leave retry policy to the storage adapter; the query state and requested
positions remain unchanged by a retry.
