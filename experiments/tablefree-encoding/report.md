# Table-free field arithmetic and s=9 encoding checkpoint

## Purpose and reproduction

Implement the first [encoder scaling checkpoint](../../docs/encoding-algorithms.md#step-1-remove-full-extension-field-arithmetic-tables):
remove log/exp arithmetic tables, preserve existing codewords/artifacts, then encode
the first larger catalogue configuration, s=9,m=2,t=7. These are measured CPU times
and exact dimensions, not extrapolated runtime or Monte Carlo capacity estimates.

Run from the repository root:

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/tablefree-encoding/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

Use any compatible Go toolchain via `--go` (default `go`). Requires Go dependencies,
Python 3 with POSIX wait4, git and writable temporary storage. The exact
[script](run.py), [driver](main.go), [packing test](main_test.go), [raw results](results.json),
[table](table.md) and [public artifacts](artifacts/) are kept together. Results
record source/input/artifact SHA256 hashes and timestamps, starting from revision
`08f9f5f` plus the uncommitted ENCODE-002 implementation. That revision alone does
not contain the new implementation; hashes identify the measured source.

## Implementation and correctness

[BinaryField and QuadraticField](../../encoding/liftrs-go/binary_field.go) implement
portable polynomial-basis multiplication/reduction, exponentiation, inverse and
absolute trace without cardinality-sized tables. The public base field supports
every s=2..20. Quadratic elements are pairs (a,b) of uint32 base-field elements,
with u²=u+delta; constructors use checked irreducible polynomials. Their arithmetic
state has fixed size, independent of 2^s or 2^(2s). This is not a constant-time backend.

The [bounded adapter](../../encoding/liftrs-go/orbit_arithmetic.go) packs those two
components into uint32 for its s<=9 transform arrays, preserving old coordinate
ordering. Arithmetic support up to s=20 is not encoder support at those sizes.
The orbit encoder now accepts s=2..9, m=2, N<=262144. Discovery uses one random
byte per coefficient through s=8, and four bytes masked to s bits at s=9, so
sampling remains uniform for the wider field. Reusing only one byte would be wrong.
No old artifact version or field polynomial changed. New degree 9 uses 0x211.

[Arithmetic and compatibility tests](../../encoding/liftrs-go/binary_field_test.go)
compare base multiplication with independent uint64 carryless convolution and
long division (exhaustive through s=8), extension multiplication with a separate
four-product oracle (exhaustive through s=4), and randomized inverse/Frobenius
identities for every s through 20. They load the historical prepared artifacts and
reproduce their pre-change complete output hashes. Existing dense-oracle tests
remain in [orbit_test.go](../../encoding/liftrs-go/orbit_test.go).

The s=9 prepared artifact also passes all systematic-symbol checks and eight
sampled degree-7 curve sum checks using independent base arithmetic. The two
encoding algorithms use the same public information set and support discovery;
agreement is not a fully independent proof at this larger size. There is no dense
oracle at N=262144. Exact dimension is established by complete monomial enumeration,
and support discovery stops only after reaching that exact count.

## Measurement boundaries

Evaluate (s,t)=(8,8) and (9,7), both with m=2. Discovery uses deterministic public
seeds 1000+s over complete good-monomial coefficient vectors; this is reproducible
code-only preprocessing, not PIR randomness. The input message uses seed 71 and
canonical field elements. Each method repeats encoding that message three times;
record all samples and report their median. No warmup, random method interleaving,
confidence interval or performance threshold is used.

Each stage uses a fresh child process:

1. **Discovery:** code-only exact enumeration, support discovery and runtime setup.
   Serialization and artifact file writes are outside discovery time.
2. **Prepare paper / prepare cyclic FFT:** load the shared basis once, then time
   method-specific polynomial preparation separately. Record basis read/load time,
   preparation and serialization independently. Each method's preparation gets
   its own RSS high-water mark; the other method is absent from that process.
3. **Encode paper / encode cyclic FFT:** load the complete prepared artifact once,
   then time database encoding. No discovery, polynomial preparation or fixed-
   generator FFT is performed per database. Record load/read time separately.

Per-child wait4 peak RSS includes the entire stage process: loaded artifacts,
serialization or output buffers, GC and runtime state. It is not isolated live
workspace inside a single function. Encoding additionally records cumulative
allocated bytes per call, which are not peak heap. This improves separation over
the [earlier experiment](../encoder-comparison/report.md), whose preparation
process held both algorithms, but does not yet measure encoding-only peak workspace.

Compute canonical uint32 output hashes for compatibility and verify both encoders
agree. Pack output using the documented bit layout (element i, bit b at file bit
i*s+b), count packing separately, write/sync/close one file, then hash a cached
readback. Record packed hashes, exact bytes and sync-write/readback times. Packing,
hashing and file I/O are outside encode time. The test sink is ordinary packed
single-database output, not grouped storage, permutation or masking. Small files
and cached readback do not measure sustained external-storage throughput.

No memory map, disk-backed polynomial intermediate or GPU is used. Preparation
still uses Karatsuba; the cached-FFT cyclic encoder still uses Karatsuba for
H=(P*M) mod X^K, with only G*H accelerated by FFT. The next arithmetic work is fast
truncated multiplication, not an assertion of complete quasi-linear encoding.

## Environment

The saved result records macOS 26.6.2 ARM64, Go 1.27.3 and Python 3.9.6.
Hardware inspection reports Apple M5 Pro, 18 cores, 48 GB RAM, MacBook Pro Mac17,9.
Only model/CPU/memory fields are retained, not serial numbers or hardware UUIDs.
Arithmetic is single-threaded; Go GC uses the default runtime configuration.
The temporary output resides on /System/Volumes/Data; df output and free space are
recorded. The DiskManagement framework is unavailable in the sandbox, so diskutil
could not identify the storage device/filesystem; the command and error are retained.
Do not treat these small output timings as fully characterized device benchmarks.
No caches are flushed, sync is included on output writes, and readbacks are cached.

## Results and interpretation

See [table.md](table.md). Exact capacities are K=3121 at N=65536 and K=17200 at
N=262144. The latter replaces the current catalogue's prior Monte Carlo point
estimate K=17118; the sampling evidence remains separately recorded and historical
planning snapshots are unchanged. Supported means the standalone bounded encoder,
not application-plan execution, a grouped-storage backend or production security.

At s=9, both methods produce identical canonical and packed outputs. Cyclic FFT
has a smaller prepared artifact and faster encoding than this paper implementation
in the local sample. Table-free scalar arithmetic is substantially slower than the
old table-backed arithmetic at s=8; memory scalability, not a throughput win, is
what this checkpoint establishes. Old and new runs differ in timing environment
and method setup, so their timing ratio is descriptive, not a controlled regression
benchmark. Existing public s<=8 artifact bytes and encoded outputs remain compatible.

The complete public artifacts remain moderately sized in these cases; nothing here
establishes s=10..20 encoder feasibility, publication sizes at those parameters,
or bounded-RAM external-memory support. Artifact loaders still perform structural
checks on trusted artifacts; provenance and the independent checks above are the
publication evidence. Future native/GPU acceleration and external-memory work remain
behind the plan's explicit decision gates.

## Subsequent implementation

The [polynomial-product checkpoint](../fast-polynomial/report.md) adds hybrid
full/low products and compares its effects against this implementation. The
measurements above remain a snapshot of the source revision recorded here.
