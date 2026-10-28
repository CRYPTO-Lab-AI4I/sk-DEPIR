# Step-3 scaling prototypes

This checkpoint tests the three uncertain operations identified by the
[scaling study](../../docs/encoding-scaling.md). It adds opt-in paths; default
preparation, encoding, native kernels, artifact format and supported parameters
remain unchanged. These are correctness and resource probes, not a selected
large-instance backend. Steps 4–5 remain open.

## Reproduce and inspect

- [Runner](run.py), [analyzer](analyze.py).
- [Initial results/source manifest](initial/results.json), [phase measurements](initial/phases.json),
  [tables](initial/table.md), [validation log](initial/checks.log).
- [s=13 results/source manifest](s13/results.json), [phases](s13/s13-phases.json),
  [table](s13/table.md), [validation log](s13/checks.log).
- Shared experimental [setup/workspace API](../../encoding/liftrs-shared/scaling_experimental.go),
  [tests](../../encoding/liftrs-shared/scaling_experimental_test.go),
  [runnable example](../../encoding/liftrs-shared/scaling_example_test.go).
- Test-only [matrix/Taylor/routing probe](../../encoding/liftrs-shared/scaling_matrix_test.go),
  [sectioned online file](../../encoding/liftrs-shared/scaling_online_file_test.go),
  [profiling harness](../../encoding/liftrs-shared/scaling_probe_test.go),
  [native helper](../../encoding/liftrs-shared/scaling_probe_native_test.go),
  [s=13 validation](../../encoding/liftrs-shared/scaling_s13_test.go).

```sh
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  PYTHONPYCACHEPREFIX=/tmp/depir-pycache \
  python3 experiments/scaling-prototypes/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name initial
# One large correctness/resource validation; no baseline/preprocessing rerun:
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  PYTHONPYCACHEPREFIX=/tmp/depir-pycache \
  python3 experiments/scaling-prototypes/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name s13 --s13
python3 experiments/scaling-prototypes/analyze.py initial
python3 experiments/scaling-prototypes/analyze.py s13
```

Choose fresh run names to repeat; evidence is not overwritten. Base revision
685b334 with uncommitted prototype sources identified by SHA-256. Environment:
Go 1.27.3, Apple Clang 21 generic C++17 -O3 backend, Python 3.9, Apple M5 Pro arm64,
macOS 26.6.2, 48 GiB RAM, local APFS. All timed code is single-threaded arithmetic;
Go GC/runtime/profiling may use other threads. No CPU affinity or cache flush.
The runner's short probe uses a 2 GiB soft Go memory target, and s=13 uses 9 GiB;
these are not hard OS memory caps. Child commands have a ten-minute timeout.

Every measured stage has a CPU profile and flat/cumulative text views. Start/stop
and pre-stage GC are outside elapsed times, but allocation/heap observations include
profiler setup/teardown (roughly 2 MiB common overhead per stage). Short timings
are single diagnostic calls; profiles can have few/no samples. Initial heap peaks
are sampled every 5 ms plus the final observation; short stages can escape the
sampler. s=13 records before/after heap and Go Sys, **not isolated peak RSS**.
Do not compare allocation totals to resident memory or infer a peak from an endpoint.
Small workspace timings include complete word comparison; setup timings include
P·G identity verification. Cache writing/sync and loading are separately labelled.

## 1. Sparse setup and cyclic preparation

Sparse setup represents S_k by coefficients of X^(2^i), applies
`S_(k+1)=S_k²+delta_k*S_k`, and omits the unused final polynomial. The same table
budget/order and native plan description are produced. Dense versus sparse
byte-for-byte comparison covers s=2..13; only sparse construction is run at s=14,15.
Independent vanishing-subspace checks, small Horner evaluation and inverse recovery
cover s=14,15 setup; no full encoding at those sizes is claimed.

| Setup | Dense allocated MiB | Sparse allocated MiB | Sparse description KiB |
|---|---:|---:|---:|
| s=13 | 514.83 | 2.79 | 128.64 |
| s=14 | not run | 3.01 | 148.42 |
| s=15 | not run | 3.33 | 191.08 |

These allocation figures include profile/serialization overhead. Sparse setup's
persistent descriptions remain small, and setup does not need N-sized arrays.
The default builder is deliberately unchanged until an adoption decision; the
common table finalization was extracted without changing its behavior.

Cyclic preparation pairs roots under Frobenius: conjugate pairs give base-field
quadratic factors, fixed roots give linear factors. A balanced root-only product
retains no full product-tree graph. Normalize P and compute G as the first N−K
coefficients of 1/P; verify P·G=1+X^(N−1) before transforming G. The paper solver
retains its existing tree. Factors and generic Newton multiplication scratch still
allocate; no allocator/RAM bound or streamed discovery is implemented.

| s | Baseline setup (s) | Prototype setup (s) | Baseline allocation MiB | Prototype allocation MiB |
|---|---:|---:|---:|---:|
| 8 | 0.285 | 0.126 | 103.02 | 100.63 |
| 9 | 2.334 | 1.560 | 1326.83 | 1357.91 |

P, G and spectrum are identical. The prototype is faster in these single calls,
but s=9 cumulative allocations increase slightly and sampled peak heap does not
improve (the prototype also runs while baseline constants are retained). This
supports the algebraic construction, **not** a claim that large preparation now
fits local RAM. Newton-product buffer lifetimes remain a concrete next issue.

## 2. Minimal online plan and reusable workspace

The plan retains P, the generator spectrum and scalar code/field data. G/support
are omitted. A test-only sectioned binary file reads these arrays incrementally,
without whole JSON/raw-basis/map validation. It requires an externally trusted
SHA-256 from a verified prepared plan, bounds shapes/elements and rejects bad
checksums/limits. It is **not** a public artifact format or an alternative proof
of the degree set. Existing public artifact loaders are unchanged.

Online files are 274,657 bytes at s=8 and 1,117,405 bytes at s=9; their arrays are
274,632 and 1,117,380 bytes. The section header accounts for the remaining bytes.
File loading, saving/sync and plan initialization are separate phases. No compression
or GitHub distribution claim is made for large spectra.

One arena aliases the two prefix-product FFT buffers with the later full-domain
convolution buffer. H is saved before reuse; each call clears scratch and retains
a separately allocated canonical output. The workspace belongs to one plan and
cannot be reused concurrently; immutable plans allow separate workspaces. For
this prototype, the full prefix convolution must fit one extension domain;
higher-rate cases needing overlap-add are rejected. The normal encoder still
supports that strategy. Prefix FFTs are used even on tiny cases where the baseline
chooses Karatsuba, so these small comparisons are not identical algorithm-dispatch
latency benchmarks. Native kernels themselves are unchanged.

At s=9 repeated native workspace encoding allocates about 2.95 MiB per call
(including profiling and the 1 MiB final output), versus 23.35 MiB in the single
baseline call. The workspace arena/H payload is 1,117,376 bytes, allocated once.
Three complete words match for each input fixture. The s=13 validation below
checks the same schedule against the original full-word hash rather than merely
agreement between two new implementations.

## 3. Matrix transform and canonical output routing

The test-only disk path uses a general ordered basis and a split into dimensions
b and d−b. Taylor conversion expresses f as sum f_j(X)·S_b(X)^j. Column transforms
use the image of the high basis under S_b; row transforms use low-basis affine
cosets. Inverse reverses the steps. Sparse Frobenius powers give recursive Taylor
division/reconstruction. Column recursion offsets are expressed in the **image
basis**, not canonical packed coordinates; an initial small test exposed that
mismatch and the implementation now distinguishes the two. Canonical row cosets
and final output enumeration stay unchanged.

Tests cover s=2..4, domain dimensions 2..2s and every nonempty split. Forward
outputs match the existing transform; inverse recovers inputs. Disk-based ordinary
convolution covers both prefix/full products with correct full-product padding.
An s=4 complete codeword reconstructed through bucket routing matches the original.

For N=256, one 1,024-byte disk array is used. The reusable data tile is 64 bytes
at a 4+4 split and 256 bytes at 2+6 or 6+2, excluding small plan metadata and the
reference comparison vectors in the test harness.

| Split | Forward bytes read | Forward bytes written | Scalar I/O calls |
|---|---:|---:|---:|
| 2+6 | 11,264 | 8,192 | 4,864 |
| 4+4 | 6,144 | 4,096 | 2,560 |
| 6+2 | 8,192 | 7,168 | 3,840 |

These are counted ReadAt/WriteAt bytes, not SSD traffic; OS caching is uncontrolled.
The scalar transpose/Taylor implementation has **too many small calls** to serve
as a performance backend. It proves decomposition and ordering and gives concrete
movement counts for a blocked-I/O implementation to improve. Do not extrapolate
the toy millisecond timings or assume Bailey's classical FFT pass count.

Routing streams orbit-order coefficients into at most 64 bucket files, then loads
one output tile at a time. Tiles of 4,16,64 elements recover identical canonical
words and parity. Bucket write+read is 4,080 bytes; adding 1,024 bytes of sequential
output gives 5,104 logical bytes. The test uses an in-memory full output only as a
comparison sink; sequential output bytes are accounted, not physically written by
this helper. Larger runs require bounded-fanout multi-level routing and wider
intermediate offsets. Secret permutation/masking and grouped storage are not
implemented by this mathematical-codeword probe.

## s=13 validation

One profiled native workspace encoding reuses the original retained 432 MiB cache
and seed-71 input. Loading that original JSON remains a separate legacy cost;
creating the minimal online plan then releases the source G/support before the
next phase GC. Native initialization and workspace allocation are separate.
The original Go/native latency baselines are not rerun. One encoding took **153.374 s**, essentially the existing native baseline's
153.409 s median (different run/harness; no speedup claim). It makes the same five
native calls over 234,881,024 elements, with zero interface buffer copies. The
complete canonical SHA-256 is unchanged and all **11,926,435** systematic positions
match. No packed-output I/O was repeated in this validation.

| Stage | Profiled time | Relevant payload / observation |
|---|---:|---|
| Original cache load | 5.022 s | Existing 432 MiB JSON; source artifact unchanged |
| Online plan construction/native initialization | 11.24 ms | P+spectrum: 316,141,200 bytes; native metadata 141,904 bytes |
| Workspace allocation | 1.91 ms | Reusable arena+H: 316,141,196 bytes |
| One native workspace encoding | 153.374 s | 270,584,944 allocated bytes including profiler/final output |
| Full hash/systematic verification | 0.633 s | Original complete hash and every systematic position |

At encoding completion Go heap is 951,344,664 bytes (~907 MiB); Go Sys remains
~2.29 GiB from earlier loading. This is not a measured process peak or a guarantee
for a fresh minimal-file-only process. The old harness reported ~855 million
allocated bytes per native encoding; the current ~271 million supports the
workspace allocation reduction, but differing instrumentation and one-call timing
must be kept explicit. The primary remaining new allocation is the 256 MiB output.
Exact endpoints/timers are in [phase data](s13/s13-phases.json).

## Retention and validation

Small reusable files and matching probe binaries are retained under
`.cache/liftrs/scaling-step3/<source-id>/`, with `provenance.json` and per-file
checksums in each result manifest. They include sparse descriptions, s=8/9 cyclic
JSON and minimal online files, and small matrix/bucket fixtures. The original
`.cache/liftrs/s13-m2-t5/` cache is read-only and unchanged. Profiles/text views
remain beside results. Changed measured source bytes are preserved in
[content-addressed snapshots](../source-snapshots/README.md), without changing
historical manifests. Later test-only additions are separately validated.

Both default/cgo-disabled and tagged correctness tests, tagged vet and full suites
in isolated copies pass for the initial probes; see logs. Small outputs also
match the independent dense oracle, preparation coefficients match the existing
path, and invalid online checksum/shape/workspace cases are tested. Full suites
also pass after the final implementation and test additions; see the s=13 log.
A separate race check and runnable example verification are recorded in
[final checks](verification.json); no large baseline is repeated.

The prototypes address all three step-3 questions, but do not solve every scaling
issue: full-size preparation peak RAM, bounded discovery, efficient blocked I/O,
native arbitrary-basis tile execution, >32-bit packed fields, huge public cache
distribution and cryptographic output integration remain open. Step 4 must measure
core/storage scaling using a worthwhile blocked implementation before step 5
selects a design or admits s=14/15 full encodings.


The native kernels are unchanged, so detailed native stack attribution reuses the
retained [s=13 native sample](../native-encoding/s13/go-cpp/cpu-sample.txt).
New Go CPU profiles cover every prototype phase, including whole native calls;
`runtime.cgocall` attribution includes kernel execution, not interface overhead.
The two small source snapshots preserve the exact first measured versions before
additional shape/Horner checks; later examples are not claimed part of measured
binaries. Code failures encountered during development (image-basis offsets) were
corrected before the retained successful experiment runs.


### Publication command checks

[command-checks.json](command-checks.json) and [command-checks.log](command-checks.log)
record `make test` (native tag) and `make test-go` (cgo disabled) passing in an
isolated copy after the standard-command defaults changed. These commands do not
promote the scaling algorithms. The [native command report](../native-encoding/report.md#current-command-defaults)
records default-C++ and explicit-Go tiny driver runs. No large timing baseline
was rerun for the command change.
