# CPU encoding checkpoint: s=10,m=2,t=6

## Purpose and reproduction

Advance the [CPU scaling plan](../../docs/encoding-algorithms.md#step-3-scale-the-in-memory-cpu-encoder-through-s912)
beyond s=9, retaining JSON artifacts and the paper's information set. These are
measured moderate-instance costs, not extrapolated terabyte-database runtimes.
The library now accepts m=2,s=2..10; s>=11 remains outside the supported encoder
boundary. Field arithmetic separately supports s through 20.

Run from the repository root:

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoder-scaling/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

Requires compatible Go, Python 3 with POSIX wait4, git, and writable temporary
storage. The [script](run.py) builds the existing
[timing/packed-output driver](../tablefree-encoding/main.go) and the new
[verification/heap-sampling driver](main.go). [results.json](results.json) records
all raw observations, exact source/artifact hashes, environment and revision
`f623d8f` plus uncommitted ENCODE-005 sources. That revision alone does not include
s=10 support; the recorded hashes identify the measured implementation.
[table.md](table.md) summarizes method costs. The complete public JSON artifacts
are retained under [artifacts/](artifacts/); the script regenerates them.

## Exact configuration and implementation changes

The field is GF(1024) in polynomial basis modulo **0x409**, m=2,t=6. Complete
monomial enumeration gives **K=103556** at **N=1048576**. Support discovery reaches
that exact count after one deterministic public sample (seed 1010). The prior
Monte Carlo point estimate K=103952 remains separate catalogue evidence; the
exact rate is 0.09875869750976562, about 0.38% below that point capacity estimate.
This is within the recorded sampling interval and is not a security estimate.

The [orbit backend](../../encoding/liftrs-shared/orbit.go) extends its degree bound from
9 to 10 with the existing fields, hybrid polynomial products and uint32 packed
extension coordinates. No information-set ordering or mathematical encoding
changed. Prepared artifact version 1 and JSON remain unchanged. The
[complete-artifact loader](../../encoding/liftrs-shared/orbit_artifact.go) cap increases
from 32 to **64 MiB** because the paper artifact occupies **35.74 MiB**. The basis
cap remains 2 MiB. These input-size caps are separate from publication and RAM
budgets; this is not a binary-format redesign.

The s=10 catalogue entry is now verified and supported by this standalone
non-cryptographic backend. Application-plan execution, cryptographic encoding,
grouped-position storage, streaming and GPU support are not added. Historical
planner experiment snapshots retain their original estimated capacities.

One encoded message contains 103556 ten-bit field elements: **129445 bytes**
(126.41 KiB) of input and **1310720 packed output bytes** (1.25 MiB). N counts
field positions, not bytes or application records of arbitrary width. This run
encodes individual PIR codewords, not a GB/TB application database.

## Measurement boundaries

Each stage is a fresh child process, measured using POSIX wait4 for peak RSS:

1. **Discovery:** exact good-monomial enumeration, degree-support discovery and
   runtime field/transform setup. Serialization and artifact writes are separate.
2. **Method preparation:** load the saved common basis once, then construct the
   paper solver or cyclic generator/spectrum. Record basis read/load time,
   method preparation and serialization separately. All of this is code-only,
   reusable across databases. No message or mask enters preparation.
3. **Uninstrumented database encoding:** load the complete prepared method once,
   create a uniform message with seed 71, then encode it three times. Report
   all samples and the median; allocated bytes are cumulative per-call allocation,
   not peak live memory. No discovery, product-tree setup or generator-spectrum
   preparation is charged to these calls.
4. **Encoding heap sampling:** use another fresh process, method and message seed
   7310. After loading and GC, record the live heap baseline containing the cached
   code and message. Sample HeapAlloc/HeapInuse every 2 ms while one call runs,
   then take a final sample. Record absolute observed peaks and the HeapAlloc
   increase above baseline. This call is instrumented and excluded from latency
   medians. Sampling/runtime overhead is included in its allocation observation.
5. **Verification:** another process compares complete paper/cyclic outputs for
   seed 7310 and independently evaluates actual curves. Its memory is separate
   from the encoding measurements.

The sampled heap increase **includes the 4 MiB unpacked output** and can miss
transient peaks between samples. It is not an exact bound on intermediate
workspace or physical RAM. The cached-code/message baseline is reported separately
in raw results. Process RSS includes artifact loading, runtime/GC, serialization
or output buffers, and all other work in that stage. Do not subtract one method's
baseline from another process's RSS. This distinguishes encoding allocations from
cached data more clearly than the earlier process-RSS-only reports, but is not
proof of bounded-RAM feasibility.

The timing driver checks all systematic symbols, hashes canonical uint32 output,
packs it little endian (element i, bit b at file bit i*s+b), then writes/syncs/closes
one packed file and hashes its cached readback. Packing, hashing and output I/O
are outside encoding time. Both methods emit exactly 1.25 MiB. This is a canonical
single-codeword sink, not permutation/masking, grouped storage or sustained disk
throughput. The independent [packing test](../tablefree-encoding/main_test.go) now
covers s=10 as well.

## Correctness and evidence

Both methods agree on their complete canonical and packed outputs for seed 71.
For the distinct message seed 7310, [the verification driver](main.go) checks:

- Complete paper/cyclic codeword equality.
- Every one of 103556 systematic symbols and uniqueness/range of its coordinate.
- 64 sampled degree-at-most-six curves, each evaluated at all 1024 field points
  using independent carryless convolution and polynomial long division. Their
  codeword-value sums are zero, as required for reduced degree at most 1022.

The [fixture test](../../encoding/liftrs-shared/s10_checkpoint_test.go) recounts all
1048576 monomials, encodes another message (seed 7311), checks all systematic
positions and tests 16 more sampled curves using the independent arithmetic oracle.
The [driver arithmetic test](main_test.go) compares 10000 random products with the
library field implementation. Existing tiny dense-oracle, polynomial-convolution,
historical artifact/output and s=9 checks remain in the suite. The
[catalogue tests](../../pirconfig/catalogue_test.go) check the exact capacity and
cross-encoder evidence while preserving the old Monte Carlo record.

Both large encoders share discovery and arithmetic; their agreement is not a
fully independent proof at N=1048576. Sampled curve tests are not exhaustive
membership or security proofs, and the dense oracle is only feasible on tiny codes.
Artifact loading still checks trusted data structurally; the reproducible generation
and cross-checks here provide the publication evidence, with these limits explicit.

## Results

See [table.md](table.md). Discovery takes **1.447 s** and is shared across methods.

| Method | Reusable preparation | JSON artifact | Artifact load | Median encoding | Preparation / encoding-process RSS |
| --- | ---: | ---: | ---: | ---: | ---: |
| Cyclic FFT | 26.475 s | 5.15 MiB | 53.45 ms | **2.682 s** | 74.28 / 55.88 MiB |
| Paper | 23.971 s | 35.74 MiB | 165.34 ms | **10.708 s** | 245.36 / 120.98 MiB |

Separate heap sampling observes an increase of **10.40 MiB** for cyclic and
**34.63 MiB** for paper, including output. The uninstrumented calls allocate
10.40 and 188.61 MiB respectively. Paper has five GC cycles in its instrumented
call; cyclic has none. This explains why allocation totals cannot be used as peak
live memory. It does not establish workspace bounds between sample points.

Cyclic packing takes 0.916 ms, output write/sync 4.520 ms and cached readback
0.517 ms. Paper packing takes 0.950 ms, write/sync 6.011 ms and readback 0.498 ms.
These small cached-file observations do not characterize device throughput.

Cyclic encoding remains approximately **4× faster** than this paper implementation,
although its one-time method preparation is slightly slower. Loading is about 2%
of one cyclic encoding call and is amortized across repeated database encodings;
artifact-format cleanup is not blocking this checkpoint. The current field/FFT
backend still needs performance work before practical large-scale claims.

Compared with the preceding s=9 run, N increases fourfold and K about sixfold;
cyclic per-call allocations decrease because the larger prefix crosses the
32768-coefficient FFT cutoff. The s=9 prefix used pruned Karatsuba. Thus simple
linear timing/memory extrapolation from either case would be misleading. Runs
were not interleaved across field sizes; this is not a controlled cross-size
performance ratio or a new hotspot profile.

## Environment and limitations

Apple M5 Pro, MacBook Pro Mac17,9, 18 cores, 48 GB RAM; macOS 26.6.2 ARM64,
Go 1.27.3, Python 3.9.6, default GOMAXPROCS. Arithmetic is single threaded; Go's
runtime/GC may use other cores. No CPU affinity, cache flush, warmup or randomized
method order is used. Each method has one preparation observation and three
encoding calls on one message, without confidence intervals or a performance gate.
The final measurement run completed before tests; tests did not run alongside it.

Temporary files reside on /System/Volumes/Data; df output is recorded. The
DiskManagement framework remains unavailable in the sandbox, so diskutil could
not identify the device/filesystem; its command/error are retained. Physical
storage performance is not characterized. Packed output is measured, but realistic
grouped application layout remains unimplemented. No external-memory work or GPU
measurement was performed; s>=11 encoding is not enabled.

## Preflight and validation commands

[Discovery preflight](discovery-pilot.json) and
[paper-preparation preflight](preparation-pilot.json) first established the exact
capacity and 37479757-byte paper JSON artifact. The latter exceeded the old 32 MiB
loader cap, motivating the minimal cap increase. These are preflight observations,
not the headline measurements. [pilot-source.json](pilot-source.json) records
commands and exact library hashes; its driver is from revision f623d8f. The one
subsequently changed library file is preserved as
[pilot-orbit_artifact.go.txt](pilot-orbit_artifact.go.txt). To reproduce the preflight,
use the s=10 source with that old loader file and the recorded timing-driver revision;
run discovery/preparation only, before attempting to load the oversized solver.

Validation after measurements:

```sh
go test ./encoding/liftrs-go ./pirconfig ./experiments/encoder-scaling ./experiments/tablefree-encoding -count=1
go test -race ./encoding/liftrs-go ./pirconfig ./experiments/encoder-scaling ./experiments/tablefree-encoding
go vet ./encoding/liftrs-go ./pirconfig ./experiments/encoder-scaling ./experiments/tablefree-encoding
go test ./...  # isolated repository copy: legacy tests write CSV files
```

Results/source/artifact hashes and report links are checked alongside these tests.
All passed. Source/artifact hashes and report links also pass verification, as does
`git diff --check`. See the [completion message](../../message-board/2026-10-28-1315-codex-encode-005-completed.md)
for recorded outcomes.

## Next step

Prioritize the cyclic encoder. Profile the current s=10 pipeline before choosing
its next field/transform optimization, then advance to s=11 and s=12 with the same
measurement boundaries. Keep current JSON until actual size/loading/memory costs
obstruct measurement; distinguish reusable cache costs from database work.
External-memory/GPU work still requires the plan's later scope decision.
