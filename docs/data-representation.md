# Application input to PIR input

This document describes how to convert an application's database representation into the
input representation of a collection of PIR databases. It defines the layout
transformations, field-element representation, and security requirements that this
conversion must preserve. Parameter choices belong in [parameter
planning](parameter-planning.md); implementation status is in the
[README](../README.md).

## Application vs PIR database

An application database is the input to the application and its dimensions, i.e. number
of records and size of records, are determined by the user. We model it as a matrix:
each row is a fixed-size application record, and each column is a bit position within
those records.

A PIR database is an input to the PIR protocol and its dimensions must align with a
[configuration catalogue](parameter-planning.md#configuration-catalogue). It is also
modeled as a matrix similarly to an application database. By PIR database we refer to
the plaintext database in a PIR instance. Its encoded form is the database encoding. The
paper uses this field-element input model in [Section 5.1](../paper.pdf#page=33); the
submatrix layout below is a library design choice.

Application records must be at least as large as the largest PIR record size in the
supported catalogue for layouts containing encoded PIR databases. Full-download-only
layouts need no field padding. Packing multiple application records into a PIR record
is outside the current scope.

## Logical record layout

Before partitioning, records may be split into fixed-width fragments and stacked as
logical rows. For `n` application records of `A` bits, choose logical width `W <= A`
and let `f = ceil(A/W)`. The logical matrix has `n*f` rows and `W` columns.
Application record `j`, fragment `i`, maps to logical row `j*f+i` and source bits
`[i*W, min((i+1)*W,A))`; zero-pad the final fragment. With `W=A`, the matrix is unchanged.
This rearrangement lets large application records use higher-capacity configurations.

Partitioning, local indices, and query-sharing conditions below refer to this logical
matrix. Retrieve its `f` fragments to reconstruct one application record. The
[application planner](../planner/README.md#application-width-search) searches widths;
`ApplicationPlan.Fragment` exposes the mapping without copying data.

## Partitioning the application database

To fit supported PIR configurations and reduce storage and communication overhead,
divide the logical matrix into submatrices. Their dimensions may differ, and row
boundaries need not align across different column portions. Each submatrix becomes a
separate PIR database; append zero rows and columns to fit its selected configuration.
Padding contributes to storage overhead.

The submatrices must cover every logical matrix entry exactly once, without overlaps
or gaps.

A [planner](parameter-planning.md) proposes one or more such partitions and estimates
their storage and communication costs. An [application
encoder](encoder-contract.md#application-encoding) applies the selected plan to the
input database. The [small application reference](../encoding/appref/README.md)
executes these transformations on supplied fixture plans; it does not optimize them.
The [plan evaluator](../planner/README.md) validates supplied partitions and calculates
their costs. The [greedy planner](../planner/README.md#greedy-planning) constructs
row-aligned partitions; general supplied partitions remain supported.

## Indexing

Use zero-based indices and half-open ranges: `[a, b)` includes `a` and excludes `b`. A
submatrix starting at logical row `a` maps logical row `j` to row `j-a` within that PIR
database (its *local PIR index*).

The existing [matrix coordinate helpers](../pir/db.go) and [flattened-index
conversions](../utils/indexTransformer.go) address encoded grids; they do not implement
this application-to-PIR mapping.

## Encoding

Each encoded PIR database undergoes two encodings: a non-cryptographic encoding that produces a
systematic codeword (preserving input elements at designated positions), followed by a
cryptographic encoding that permutes and masks it. Their inputs and outputs are defined
in the [encoder requirements](encoder-contract.md), with paper and implementation
references.

### Query-sharing groups

To retrieve a PIR record, the client sends a query listing positions to read from the
encoded database. Several PIR databases can share query generation: the client
computes one position list and sends a separate copy for each PIR database. Storage
returns the requested values independently for each database. Query sharing reduces
client computation; upload still includes every copy of the position list.

PIR databases are eligible to form a *query-sharing group* when:

1. **Same encoding and query configuration:** the PIR databases use the same
   field representation, code parameters, padded dimensions, mapping from PIR
   row indices to codeword positions, and query-generation
   procedure, including the dummy-position count.
2. **Compatible row ranges:** for every logical row `j`, all group members
   containing part of that row must map it to the same PIR row index. Under the
   mapping `j-a`, this means every pair of unpadded logical row ranges must either be
   disjoint or have the same starting row. Their ending rows may differ.

For example, `[0,100)` and `[100,200)` satisfy the row-range condition, as do `[0,100)`
and `[0,80)`. The overlapping ranges `[0,100)` and `[50,150)` do not. These restrictions
apply within a group; the overall partition remains general.

Once groups are chosen, encode all members of each group using the same secret
permutation so that one query addresses corresponding codeword positions in every
member. Keep group membership fixed for the encoded layout; the query procedure in
[protocol I/O](protocol-io.md#query-groups-and-participation) must not change it
according to the requested record.

### Identification and mask separation

Each PIR database uses its own masking key, regardless of whether it is in a
query-sharing group. This is extremely important: Reusing masks across PIR databases
exposes relations between their data. Keys may be derived from a master key using HKDF
with distinct database context labels.

Identify each PIR database by its submatrix's row and column ranges within the
logical matrix, scoped to its application plan.

Associate each PIR database encoding with separate cryptographic material for masking.
Re-encoding must not reuse masks from a previous encoding, even if the submatrix
coordinates are unchanged. The key-derivation and randomness formats belong to the
[cryptographic encoder](encoder-contract.md#cryptographic-encoding) and remain to be
specified.

### Full-download databases

A trivial PIR database stores a submatrix without Lift-RS expansion. Fetch its entire
contents once on every application-record retrieval, whether or not it contains a
requested fragment. It has no code configuration or query-sharing group. Keep its
mask material separate just as for encoded databases; downloading everything removes
the need for a secret permutation. These are planning and I/O requirements, not a new
implemented cryptographic backend. Selection rules are in [parameter planning](parameter-planning.md#bounded-tradeoff-search);
the tradeoff API treats full downloads as ordinary costed choices.

## Query

Each logical fragment may span several PIR databases. Each submatrix's row and column
boundaries determine whether it contains part of the record, the local PIR index, and
where that part belongs in the logical fragment.

Reconstruct each logical fragment by removing PIR padding and joining its column pieces.
Join fragments in order and remove final-fragment padding to recover the application
record. [Protocol
I/O](protocol-io.md#query-groups-and-participation) specifies how to query every
query-sharing group with fixed participation and handle responses that contain no part
of the requested record.

## PIR record representation

Each PIR record is one element of a binary extension field `GF(2^s)`, represented by `s`
bits. The PIR configuration catalogue includes every integer `s` from 8 through 20. A submatrix
assigned to this field therefore has at most `s` data columns; append zero columns to
reach `s` when necessary. Wider portions of an application record are represented by
additional submatrices. Prime fields are outside the current scope.

An element's bit `i` represents the coefficient of `X^i`, for `0 <= i < s`, modulo a
fixed irreducible polynomial `P(X)` of degree `s`. Addition is XOR; multiplication is
polynomial multiplication followed by reduction. See [RFC 5510, section
8.1](https://datatracker.ietf.org/doc/html/rfc5510#section-8.1).

Use the polynomial identifiers in [gf2n.go](../utils/gf2n.go), including `0x11D` for
degree 8 and `0x1002D` for degree 16. Identify both `s` and `P(X)` in the metadata so
all parties use the same representation. That file also implements `GF2n.Add` and
`GF2n.Mul`; it does not yet support every target degree.

## Layout metadata and serialization

Retain original application dimensions, logical width and fragment mapping, and, for
each submatrix:

- Its half-open row and column ranges before padding, which identify it within
  the logical matrix.
- Whether it is a full-download piece; otherwise its PIR configuration and padded
  dimensions, including the field polynomial.
- Its query-sharing group for encoded PIR databases.

Planner output may represent a homogeneous full-width row block compactly, with its
row range, configuration ID, and group. Its vertical slices are implicit at multiples
of the field width, with the last slice truncated to the logical width. Each
slice still identifies a separate PIR database by its full row/column coordinates
and requires separate masking material. The compact planner representation is not
an implemented stored-database manifest or a change to the packed bit layout.
A compact row block may also give `repeat`: divide its row range into that many equal
consecutive blocks, each with the same configuration and column slices (zero means one).
Each physical repetition has separate coordinate identifiers and masks. Store a shorter
final padded block separately; full-download pieces do not repeat.

The ranges locate returned pieces within logical fragments; the fragment mapping
places them within the application record. Comparing the logical and padded dimensions identifies which zero
rows and columns to discard. Secret keys are not public metadata.

Each PIR record occupies `s` bits when tightly packed, regardless of the integer type
used for arithmetic. Field-element widths that are not multiples of eight require bit
packing.

Application rows retain their input order. Within a row, column `8*k+b` is
bit `b` of byte `k`, with bit 0 the least significant bit. For a submatrix
starting at logical column `c`, its column `c+i` becomes coefficient bit `i`
of the PIR field element. Append padding at the ends of rows and columns.

Store each PIR database's encoded field elements in codeword-position order.
Element `i`, coefficient bit `b`, occupies bit `i*s+b` in the data file; file bit
`8*k+b` is bit `b` of byte `k`. This defines little-endian bit packing without
per-element byte padding. Zero unused high bits in the last byte. The file has
`ceil(N*s/8)` bytes. For example, two 9-bit elements with integer values `1,1`
pack as the three bytes `01 02 00` (hexadecimal).

Use a UTF-8 JSON manifest with integer `format_version: 1`. It records application
row count and record width in bits, the logical width and mapping when reshaped,
and a list of PIR databases in lexicographic
order of `(row_start, column_start, row_end, column_end)`. Each entry records these
half-open ranges, its configuration (including `s`, `m`, `t`, `K`, `N`, polynomial,
coordinate/systematic mapping convention, and cryptographic suite when selected),
and a query-sharing group number for indexed databases. Full-download entries instead
identify that mode and store their tightly packed matrix bits without code expansion.
Number groups from zero in first-occurrence order
in that list. Keep secret material separately. Reject unknown format versions or
unsupported configurations; format version 1 describes the layout conventions and
does not imply an implemented or validated cryptographic suite.

The existing [CSV readers and writers](../utils/CSVConnector.go) remain prototype
adapters. CSV is not the packed storage format; compatibility requires explicit
conversion rather than interpreting CSV bytes as encoded field elements.
