# Small application correctness reference

This package completes the small-instance pipeline from application records to
PIR encodings and back. It executes a supplied partition plan using
[liftrs](../liftrs-go/README.md) and [cryptoref](../cryptoref/README.md).
It provides a correctness oracle for future implementations, without selecting
their streaming strategy, storage layout, memory use, or public API.

[example_test.go](example_test.go) is a runnable example: partition one byte into
four PIR databases, encode them, serve ordinary reads with `bytes.Reader.ReadAt`,
and retrieve the original byte.

## Plan and encoding

`Encode(plan, records)` returns a private `Client` and independent public
`StoredDatabase` values containing coordinate IDs and packed ciphertext bytes.
`Client.Plan()` returns a copy of the layout with no keys or permutation tables.
Application data is not retained by the client.

A `Plan` specifies row count, record width in bits, and rectangular `Partition`
entries. Each entry contains half-open application coordinates, a Lift-RS code
specification, and a query-sharing group number. Entries must be ordered by
`(row_start,column_start,row_end,column_end)` and cover every application bit
exactly once. Group numbers start at zero in first-occurrence order. Invalid
plans are rejected; the wrapper does not repartition or choose parameters.

Groups require identical code specifications and pairwise disjoint row ranges
or equal starting rows. Row boundaries may differ across groups. Preprocessing
is reused for each distinct code within the encoding operation. Every group gets
a fresh permutation key, and each PIR database encoding gets its own fresh mask
key. A new `Encode` call refreshes keys, including on re-encoding unchanged
coordinates. Secret state remains in the client object rather than public data.

Input rows have exactly `ceil(RecordBits/8)` bytes, with unused high bits zero.
Bit `b` of byte `k` is application column `8*k+b`. Submatrix columns become field
coefficient bits starting at zero; trailing columns and message rows are padded
with zeros. Stored field elements use the agreed LSB-first packed bitstream.
See [representation rules](../../docs/data-representation.md).

## Retrieval and ordinary storage

`Client.Retrieve(ctx, row, read)` generates one curve/dummy list per fixed group
and sends a full copy of it to every member through the `ReadAt` callback.
It performs a byte-range read for every position occurrence, even repeated or
irrelevant positions. The callback sees only a database's coordinates, byte
range, and context. Storage performs no PIR computation and need not know groups.

The client uses a common local PIR index for needed members and a valid cover
index when none is needed. It prepares all queries before I/O and completes all storage reads before
unmasking and decoding needed pieces with curve multiplicity preserved, then restores their application
column ranges. Padding is omitted from the returned record. An error yields no
partial record. I/O errors are wrapped, cancellation is checked between reads,
and complete reads with `io.EOF` follow `io.ReaderAt` semantics.

Queries use `crypto/rand`: uniform coefficients for a degree-at-most-t curve
through the target at parameter zero, q-1 nonzero evaluations, 128 independent
dummy positions, and a shuffled order. Private entry indices distinguish real
and dummy occurrences even at the same address. The tiny field operations are
local reference helpers using the code's polynomial. The fixed dummy count and
tiny parameters carry no claimed security level.

## Bounds and current scope

Application dimensions are at most 256 rows and 256 bits per record, with at
most 64 PIR databases. Records must have at least 4 bits, the largest PIR record
size supported by this fixture backend. Codes retain the reference encoder's
`N<=256` bound and field degrees 2, 3, and 4. These are fixture limits, not the
initial `s=8..20` security catalogue or limits of the mathematical construction.

This wrapper holds application data during encoding, materializes codewords,
and keeps private state in memory. It exposes the public plan in memory; durable
JSON manifest serialization/reopening and private-key persistence remain separate
work. Ordinary-read adapters can be supplied by the caller, but no cloud adapter
is bundled. The legacy file-based RM API remains a separate prototype path.

This is a correctness reference, not production cryptographic qualification.
The custom AES permutation is inherited from the prototype; masks do not
provide authentication. Future backends may use different I/O and API designs
while being compared on matching small inputs, plans, and reconstruction.

## Verification

```sh
go test ./encoding/appref
go test -race ./encoding/appref
```

[Tests](appref_test.go) cover general nonaligned partitions, mixed fields,
row/column padding with explicit golden parts, packed bit boundaries, all 256
one-byte application values, 3D and high-curve-degree fixtures, fixed participation
(including groups with no useful data), repeated shared uploads, forced repeated
curve positions, invalid plans/records, read and entropy failures, cancellation,
plan ownership, re-encoding, and concurrent retrieval. They require no server,
external encoder, or tracked example files.
