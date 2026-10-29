# Encoding scaling study

Use the shared encoder with C++ transforms as both **baseline and starting code**.
Retain Go preprocessing and independent Go/dense correctness references. The
[current checkpoint](encoding-algorithms.md#current-checkpoint-and-next-step)
now validates an experimental s=14 path; public default discovery remains s≤13
and s≥15 remains unmeasured. Keep the current
information set and canonical field/grid representation, not just membership in
the same code. Begin on the local 48 GiB CPU machine; choose larger hardware only
when measurements identify a specific need.

Throughout, q=2^s, m=2, N=q² and K is the code's message dimension. Resource
figures describe **one field-element database encoded under one code**. Encoding
an application layout can invoke several such encodings and reuse code-only setup.

## Five-step plan

| Step | Work | Required output | Status |
|---|---|---|---|
| 1. Resource accounting | Audit arrays, lifetimes, widths, accesses and preprocessing/output costs at s=13..20 | Source-linked inventory and reproducible analytical model | Complete: source audit/model |
| 2. Execution designs | Compare blocked in-memory and external-memory transforms/products, selective preprocessing and spectrum choices | Concrete candidates, cost models and explicit unresolved dependencies | Complete: candidates/gates |
| 3. Focused prototypes | Implement uncertain operations at manageable sizes under constrained RAM; compare with the existing code | Verified decompositions and measured memory/I/O, not just timings | Prototype round complete; performance gates open |
| 4. Local calibration | Measure useful-core scaling, sustained storage and compute/I/O overlap; assess GPU suitability | Calibrated resource model and reasons for any larger hardware test | Bounded local calibration complete; sustained/integrated gates open |
| 5. Design selection | Select from measured evidence; integrate and validate progressively at s=13,14,15 when budgets permit | Canonical outputs, measured resource use and next hardware recommendation | Local s=13/14 CPU integration complete; s=15 gate open |

The first local feasibility target is s=15; the design must account for s=16..20.
No runtime/RAM/disk budget is increased by this plan. Agree numerical budgets
before larger runs. Reuse comparable baseline results, durable preprocessing and
profiles. Keep setup, cache loading, database encoding, output and validation
separate. Measure temporary storage and bytes transferred as well as peak memory.

The study is complete when measured prototypes justify a design and its target
sizes/hardware requirements; implementing every size through s=20 is not required.

## Step 1: resource accounting

The [reproducible model](../experiments/encoding-resources/report.md) audits the
current code at revision f2aedf8. It counts array payloads and selected schedules;
it does not predict peak RSS or new encoding times. K is verified at s=13 and
Monte Carlo estimated at s=14..20. N=2^(2s). Current slots are four-byte field
values and eight-byte Go support indices. Model rows above s=16 use eight-byte
packed extension values; that backend does not exist yet.

The current cyclic encoder computes `H=(P*M) mod X^K`, then `Y=G*H` and the
canonical codeword. M contains K input field elements. Code-only P and G satisfy
`P*G=1+X^(N-1)`; the cached spectrum is the additive transform of G. Omega is the
primitive extension-field element defining the systematic positions and final
coordinate mapping. These choices are shared by the baseline and candidate paths.

### Arrays and lifetimes

| Stage | Arrays retained or created | Access / lifetime |
|---|---|---|
| [Discovery](../encoding/liftrs-shared/orbit.go) | Good-monomial indices 8K bytes, sampled coefficients up to 4K, grid 4N, found flags N, column buffer 4q, eventual support 8K | Enumerates all N monomials; row transforms, gathered columns, full extension interpolation; may repeat samples. Exact K discovery remains required. |
| [Transform preparation](../encoding/liftrs-shared/orbit_fft.go) | Dense polynomial p and next, reaching 4(N/2+1)+4(N+1) bytes for the extension transform; retained constants/tables are small | Builds on discovery **and cache loading**. The last next polynomial is unused. Arrays are temporary; GC may retain earlier allocations. |
| [Cyclic preparation](../encoding/liftrs-shared/orbit.go) | K extension roots, full product tree, P of K+1 base coefficients, dividend N, G of N−K | The cyclic path needs only the product-tree root polynomial. The paper solver separately needs its full tree. |
| [Product tree/division](../encoding/liftrs-shared/orbit_polynomial.go) | 2K−1 nodes; all tree polynomials; reversed dividend/divisor, reciprocal iterates, quotient, product and remainder; nested FFT buffers | The tree retains every level. Division allocates several long arrays; their capacities/GC overlap need stage measurements, so summing every allocation is not a valid peak bound. |
| Complete prepared encoder | P+G: 4(N+1); spectrum: 4N; support: 8K | Immutable, shared between databases. Total logical array payload: **8N+8K+4 bytes**, excluding small metadata and spare capacity. |
| [JSON loading](../encoding/liftrs-shared/orbit_artifact.go) | Original JSON, copied raw basis JSON, decoded arrays, rebuilt transform state and a support-validation map | Whole-file decoding and map construction. Serialized bytes and runtime memory are different quantities. |
| [Prefix H=P·M mod X^K](../encoding/liftrs-shared/orbit_product.go) | Input M: 4K; output H: 4K; two extension arrays of length L=nextpow2(2K−1) | Three transforms, one point product. L must fit the full convolution, even though only K coefficients are returned. Large products use overlap-add when needed. |
| [Cached G·H](../encoding/liftrs-shared/orbit.go) | y: 4N, shared generator spectrum | Two full transforms and one point product; spectrum read sequentially during multiplication. |
| Canonical output | word: 4N, alongside y until assembly completes | Reads y sequentially, writes word at powers of omega; accumulates parity at position zero. This is a large scatter, not sequential output. |
| [Native boundary](../encoding/liftrs-shared/cpp_encoder.go) | Small immutable plan; borrowed Go work arrays | Five transform calls at measured s=13, zero interface buffer copies. Native transforms allocate no database-sized scratch. |
| Packing / crypto | Packed output Ns/8; current benchmark uses bounded output buffers | Packing is separate from encoding. Large-scale masking, secret permutation and grouped-storage writing are not part of this mathematical-encoder benchmark. Their future I/O must not be omitted. |

The final assembly maps polynomial coefficient j to the canonical coordinate
omega^(j+1), with a parity value at zero. The [paper's information set](../paper.pdf#page=31)
and [encoding steps](../paper.pdf#page=32) remain fixed. `SystematicPosition(j)`
needs no K-entry map. Storing only orbit order would instead require an efficient
inverse address map for arbitrary query points; a discrete-log lookup is not a
free replacement for canonical storage.

At that resource audit the bounded code/plan APIs stopped at s=13;
default complete/basis JSON loader limits are 64/2 MiB, and the retained-cache
benchmark uses an explicit 2 GiB read limit. They are API/harness constraints,
not mathematical memory bounds. A larger-size prototype needs reviewed format,
width and resource-limit changes rather than simply lifting the degree guard.

### Work and movement, beyond array sizes

Discovery performs N exact monomial tests; each sample performs 2q base-field
transforms of length q and one extension interpolation of length N. Root generation
currently performs K extension powers and inversions. Product construction has a
balanced tree of polynomial products, followed by Newton-based division and one
cached-generator transform. These costs remain even if buffers are on disk.

For one cached encoding, the principal transform work is three length-L transforms
for H and two length-N transforms for G·H. There are L+N pointwise extension
products, plus input validation, copies, field checks and N−1 orbit/scatter steps.
Current transforms include monomial-basis conversion and O(n log² n) arithmetic.
A smooth N-only forecast misses L's power-of-two jump, field-width changes,
cache effects and the output scatter. No new CPU-time prediction is made here.

### What changes the scaling assessment

1. **Setup currently allocates dense arrays to construct sparse constants.**
   S_(k+1)=S_k²+delta_k·S_k is linearized: only powers of two occur. Updating its
   nonzero coefficients directly can remove the N-sized setup arrays without
   changing field coordinates or transform output. The opt-in step-3 sparse
   constructor now verifies this; default constructors still use dense setup.
2. **The cyclic product tree retains information it will discard.** Logical tree
   coefficients satisfy C(1)=2 and C(K)=K+1+C(floor(K/2))+C(ceil(K/2)); node headers
   are 72 bytes on this host. At s=15, coefficients plus headers alone account for
   about **54 GiB**, excluding allocator/capacity overhead. A cyclic-specific
   consuming product builder can retain only a frontier; division scratch must
   also be bounded. Do not remove the tree from the paper reference solver.
3. **Cached online encoding does not read G or the support values.** It reads P,
   the spectrum, K, fields and the coordinate generator. A separately validated
   online plan can omit G/support, retaining the full artifact for provenance and
   independent verification. This needs an explicit API/artifact design; existing
   loaders and validation are not silently weakened. Step 3 now tests a minimal
   online plan and internal binary file; it has not selected a public file format.
4. **Domain widths and prefix padding have discontinuities.** s=16 needs wider
   cardinality/size accounting (N=2^32); s>=17 needs >32-bit packed extension
   elements. The s=20 point capacity gives L=N rather than N/2. At s=19 the Monte
   Carlo interval crosses K=N/4, so either prefix size is possible until K is exact.

| s | Current resident arrays | Principal current buffers | Proposed lean buffers | Packed online P+spectrum |
|---:|---:|---:|---:|---:|
| 13 | 0.59 GiB | 1.18 GiB | 0.88 GiB | 0.22 GiB |
| 14 | 2.38 GiB | 4.76 GiB | 3.57 GiB | 0.96 GiB |
| 15 | 9.62 GiB | 19.23 GiB | 14.42 GiB | 4.13 GiB |
| 16 | 38.85 GiB | 77.70 GiB | 58.28 GiB | 17.71 GiB |
| 20 | 14.08 TiB | 32.16 TiB | 27.12 TiB | 5.65 TiB |

These are **logical payload scenarios, not measured RAM or upper bounds**. The
lean scenario still keeps a separate canonical output array. JSON, support maps,
capacity slack, garbage retention, runtime and OS memory are excluded. Full tables
and formulas are in the [model report](../experiments/encoding-resources/report.md).
Measured s=13 loading/encoding RSS remains 2.26 GiB, above these payload counts;
first-use preparation peaked at 5.34 GiB. Earlier proportional RSS forecasts remain
historical estimates, not unavoidable memory requirements. The study improves
accounting; it does not establish that s=15 fits or how fast it will run.

Fixed-width packing of a full s=20 support/P/G/spectrum artifact is about 8.8 TiB;
a P+spectrum online artifact is still about 5.65 TiB. These are storage scenarios,
not compression lower bounds. Dense spectra are not moderately sized GitHub
artifacts. Small field/transform metadata can be public; huge optional caches need
separate storage/regeneration or exploitable structure. That distribution problem
is unresolved, and no public artifact-size budget is increased here.

## Step 2: execution designs

### A. Lean in-memory CPU path

Start from the current native transforms. Generate sparse transform constants;
build the cyclic root product without retaining all child trees; bound division
scratch. Replace whole-JSON online loading with a versioned sectioned binary plan,
loading only P and the cached spectrum. Reuse two prefix FFT workspaces and one
full convolution workspace; explicitly release/reuse allocations across stages.
Retain separate canonical output initially, so the design does not rely on a
free in-place permutation. Also test Frobenius-paired root factors: m=2 gives
orbits of length one or two under alpha -> alpha^q. A pair contributes
`X²+(alpha+alpha^q)X+alpha^(q+1)`, entirely over the base field; fixed roots give
linear factors. Their product is the same P before normalization. This derives
from the already-validated support closure, not a new code or information set.
It may reduce preparation work; its benefit and exact output require a prototype.

Generic division is another avoidable allocation source: since P[0]=1 and
P·G=1+X^(N−1), the first N−K coefficients of `1/P` are exactly G. A dedicated
Newton-series construction can omit the dense dividend and reversed-division
copies. It still needs bounded product scratch and independent verification of
P·G before publishing artifacts. Compare it with the current division in step 3;
this study has not changed preparation. Scalar arithmetic and output conventions stay unchanged.

For base-slot width b=4, extension width e and prefix length L, online resident
payload is R=b(K+1)+eN. The modeled principal schedule is the larger of
`R+2bK+2eL` (prefix) and `R+2bK+(e+b)N` (assembly). At s=15 this is about 14.42 GiB,
which makes a local prototype worth testing. Setup, allocator/GC and actual
lifetimes still need measurement. At s=16 the same schedule is about 58 GiB;
it does not fit this machine simply by tuning arithmetic. A hybrid retaining
in-memory transforms but routing output to disk removes the full word array;
its principal payload is about 42 GiB at s=16, **before** routing buffers and runtime
overhead. This borderline option deserves measurement, not a claim of feasibility.
The designs are composable; disk output does not require disk-backed FFTs.

**Gate:** reproduce constants/artifacts/codewords and demonstrate bounded setup
and workspace use. This is the first local candidate, not a guaranteed solution.
The setup/format work is reusable in the other designs, limiting wasted effort.

### B. Tiled external-memory CPU path

Keep the same polynomial operations but replace whole-array transforms with an
exact matrix/tiled additive transform. Use bounded in-RAM tiles, explicit temporary
files and blocked transposes. Stream the generator-spectrum point product. The
prefix convolution requires the same treatment; fixing only the final N-transform
leaves a large 2L workspace. Product generation/division also need external-memory
schedules if preprocessing itself exceeds the host budget. Discovery also needs a
bounded tensor-evaluation/interpolation schedule and streamed support unions;
a fast online transform alone does not remove its N-sized grid/flags or exact
monomial enumeration. Preprocessing may run on a larger machine and be reused,
but that deployment cost must be reported separately.

A relevant construction is the general-basis Algorithm 1 in
[Samanta–Badakhshan–Gong, §III](https://arxiv.org/html/2608.20855v1#S3): Taylor
expansion in a subspace polynomial followed by column and row additive transforms.
Its arbitrary ordered basis avoids assuming our fields have a full Cantor basis.
An inverse, exact coordinate ordering, Taylor-expansion memory/I/O and file schedule
are now checked by the small Go matrix/Taylor/file prototype. The current native plan builder assumes the
canonical packed basis, and native Apply starts at zero; matrix subtransforms need
arbitrary ordered-basis/affine-offset plan semantics and validated output ordering.
A tile API alone is insufficient. The scalar-I/O correctness probe is not yet
a practical external-memory encoder. Its specialized Cantor timings do not
predict this implementation's speed.

[Bailey's external-memory FFT](https://www.davidhbailey.com/dhbpapers/fftq.pdf)
provides the hierarchical-transfer motivation. Its multiplicative FFT is not a
drop-in binary-field algorithm, and its small pass counts cannot be assumed for
our additive transform with monomial-basis conversion.

The current native recursion performs ordered polynomial division before recursive
half-transforms. Long-offset updates have dependencies. Merely mmap-ing that code,
or independently transforming arbitrary tiles, supplies neither a RAM bound nor
a correct blocked factorization. An exact decomposition is the central prototype.

Canonical output needs its own schedule: route `(local offset, value)` pairs into
bounded-fanout grid buckets, then reorder each in RAM and write contiguous output.
Use multiple partition levels if too many buckets/buffers would otherwise be
open. Four-byte local offsets require each final bucket to span at most 2^32
positions; intermediate larger buckets need wider offsets and more traffic.
An alternative is a measured in-place permutation; it must account for cycle
tracking, random accesses and index computation. Crypto/grouped output can later
compose routing stages only if mask/permutation semantics remain intact.

**Gate:** prove the forward/inverse and output mappings against the reference,
then measure actual file passes and scratch. With p read/write sweeps per
transform, the five-transform scenario transfers `2pe(3L+2N)` bytes, before point
products, routing, cache reads and other intermediates. p=2,4,8 in the model are
sensitivity choices, not measured or guaranteed pass counts. At s=15, p=4 gives
112 GiB of transform traffic; at s=20 it gives 320 TiB. At an assumed sustained
1 GiB/s, the latter alone takes 91 hours. Disk fit is not practical throughput.

### C. GPU tile executor for the same design

Use A's bounded buffers or B's proven decomposition; transfer batches of tiles
and overlap transfer/compute only when measurements justify it. Preserve the CPU
fallback and exact field representation. Device workspace is limited by the tile
budget; the host still needs a complete preprocessing/output schedule. A GPU does
not solve the tree, cache distribution or canonical scatter by itself.

Compare kernel time plus uploads/downloads and synchronization against CPU tile
execution. This host's integrated GPU shares memory; it does not add a separate
RAM pool. A discrete GPU may supply a different bandwidth/capacity balance, but
no vendor stack, GPU requirement or performance gain is selected in this study.

**Gate:** demonstrate that offload saves end-to-end time within generic deployment
constraints. GPU kernels wait until tile semantics and movement costs are known.

### Cache and hardware choices common to all candidates

The current uncached-G path already exists: `PrepareCyclic` retains P/G and
uses general overlap-add. At these catalogue sizes G has more than N/2 coefficients:
it is split into two chunks, costing three length-N and three length-N/2 transforms
for G·H, versus the cached path's two length-N transforms. A dedicated full-domain
convolution could instead use three length-N transforms, but that is a candidate,
not current behavior. Removing the spectrum shrinks persistent storage but can
increase workspace, compute and I/O; it is not automatically better. Compare cached, regenerated and disk-streamed
spectra by setup frequency and actual transfers. Never change the full-product
padding rule to save memory; it would alias high coefficients into H.

A larger shared-memory CPU server can validate a candidate once its resource model
is measured. Distributed HPC is not the starting design. Kernel speedups are
welcome where reusable, but do not select a hardware architecture or resume a
long tuning campaign before these memory and access dependencies are resolved.

### Handoff to step 3

The first focused probes should be:

1. Sparse constant generation and cyclic root-only/Frobenius-paired product construction versus
   existing constants and root polynomials; measure setup/division lifetimes.
2. An online-plan/workspace prototype under controlled RAM, retaining a normal
   canonical output array; validate full small codewords and existing hashes.
3. General-basis matrix forward/inverse transforms at several splits, including
   full and prefix products, followed by tiled canonical output routing. Count
   every byte read/written and temporary file before extrapolating larger cases.

The [step-3 probe report](../experiments/scaling-prototypes/report.md) now records
these prototypes and their measured limits. Sparse constants match through s=13
and setup alone is exercised at s=14/15. Root-only paired preparation, reciprocal
G and minimal binary online state preserve P/G/spectra/words on bounded fixtures.
A native reusable workspace matches the complete s=13 baseline hash in 153.37 s.
The matrix forward/inverse, disk convolution and canonical bucket routing are
correct on small cases; scalar I/O is not a practical external backend yet.

Steps 1–2 and this step-3 correctness/probe round are complete. Before extrapolating
or selecting hardware, implement blocked I/O for Taylor/column stages, measure
large preparation scratch/GC lifetimes and a minimal-loader-only process peak,
and port general-basis/affine tile semantics into native execution as needed.
The old default constructors remain unchanged. No s=14/15 encoding, public binary
format, RAM bound or full native external-memory encoder is claimed by that
prototype checkpoint. The following sections record subsequent calibration and
local integration; sustained storage and larger-size resource gates remain.


## Step 4: local calibration checkpoint

The [calibration report](../experiments/local-calibration/report.md) records three
repeat CPU/file/overlap tests and fresh-process resource observations. Independent
native tiles scale 6.87–6.99× at eight workers and 12.42–12.77× at sixteen, across
4 MiB and balanced-matrix-sized 32 KiB tiles. This does not establish parallel
speedup of complete encoding or generalized image-basis matrix stages.

MiB-sized transfers substantially reduce file-call overhead compared with 4 KiB
calls. F_NOCACHE is explicitly requested on this host, but the 256 MiB tests are
short filesystem/device-path calibrations, not sustained or durable-media
bandwidth promises. Simple one-compute-stage overlap gives no material improvement.
Retain this distinction before extrapolating the external-memory model.

Fresh s=10 preparation reduces peak process RSS from 78.4 to 50.6 MiB and measured
phase time from 12.00 to 8.80 s, preserving the exact complete prepared artifact.
Cumulative allocation grows, so scratch/GC improvements remain necessary. The
fresh minimal s=13 loader peaks at 623 MiB RSS in 1.01 s, without workspace/output;
its arrays occupy ~302 MiB and per-element decoding still creates temporary
allocations. Larger-size feasibility remains unmeasured.

The GPU inventory is recorded, with no available Metal compiler through xcrun and
no GPU benchmark or installation. CPU-first remains the preferred implementation
candidate on current hardware. Step 5 should combine bounded state, native
arbitrary-basis/affine tiles and parallel scheduling with blocked Taylor/transpose
I/O; improve minimal binary decoding, verify the combined s=13 path, and then
review budgets before larger runs. Longer storage qualification and integrated
parallel I/O remain gates when required by the chosen design. No HPC/GPU platform,
higher memory budget, worker default or additional supported field is selected.

## Parallel integration work item

Following step 4, implement these three steps before larger runs:

1. Extend C++ transforms to arbitrary ordered bases and affine offsets, preserving
   canonical field coordinates. Integrate the verified matrix decomposition with
   parallel row/column tiles, blocked transposes and Taylor conversion. Account
   for all scratch explicitly and make concurrency configurable.
2. Load the minimal P/spectrum plan with buffered binary decoding and reuse
   database workspaces. Keep artifact generation, validation and loading separate
   from database encoding; code preprocessing remains Go.
3. Check complete small codewords against existing references, then encode the
   retained s=13 fixture with profiles. Measure initialization, loading, encoding,
   output validation, process RSS and scratch separately. Compare to the retained
   153-second native baseline without repeating expensive preparation.

A 3–6× end-to-end improvement is an exploratory target, not a measured prediction.
Local s=14/15 runs follow only after this integration and a resource-budget review.
Larger commodity hosts and disk-backed intermediates remain subsequent options;
GPU offload depends on measured compute/movement costs. This work does not claim
that independent-tile speedups apply to the whole encoder.

### Parallel integration checkpoint

Expanded steps 1–3 above are complete for the opt-in in-memory path. The
[report](../experiments/parallel-encoding/report.md) records arbitrary-basis/affine
native transforms, reusable Taylor tables, blocked matrix transposes, explicit
workers and buffered minimal loading. Six full s=13 outputs match the retained
canonical hash. Profiled medians are 29.8/20.2 seconds at eight/sixteen workers;
whole-process peak RSS is 1.15 GiB. The historical serial native median is 153.4 s.

This validates a practical local CPU schedule through s=13. Extra transpose
scratch is counted, and no disk/GPU implementation or larger support is implied.
Large discovery/preparation remains a separate measured gate: proceed to s=14
under reviewed time/RAM/storage budgets before admitting s=15. For larger sizes,
retain the same mathematical/tile semantics with larger commodity hosts or bounded
file schedules as needed, rather than assuming all state fits this host.

## Measured s=14 checkpoint

The [profiled first-use experiment](../experiments/parallel-s14/report.md) completes
exact discovery (K=50,976,325 in 254.9 s), reusable sparse/paired preparation
(2,018.0 s), binary save/load and three full native encodings (median 90.53 s).
Every systematic symbol, 16 independent curves, repeat hashes and packed readback
pass. A fresh cached process gives 92.12 s and 4.59 GiB peak RSS, versus 13.73 GiB
for first use. Code-only support/P/G/spectrum total 2.19 GiB; the online file is
1.19 GiB. Profiles, native samples, constants and provenance are retained.

Default public discovery and JSON basis loaders stay at s≤13. Experimental
preprocessing and the minimal native path establish this s=14 evidence; they do
not add application/cryptographic execution or a disk backend. s=15 admission now
needs a budget based on preparation lifetimes as well as online memory. Do not
multiply retained first-use RSS and treat it as an unavoidable online requirement.

## Next admission: s=15

1. Audit the s=14 preparation profiles and live array lifetimes at s=15 sizes.
   Online encoding suggests roughly 6–9 minutes and 18–22 GiB RAM on this host;
   these are planning estimates, not measurements. Multiplying first-use RSS by
   four gives about 55 GiB, so distinguish live scratch from retained garbage
   before deciding whether preparation fits the 48 GiB machine.
2. Bound preparation memory where the audit requires it: consume root-factor
   batches and reuse/release polynomial-product and reciprocal scratch. Verify
   any change against retained P/G/spectrum and small independent references.
   Make no change solely to fit a proportional RSS forecast.
3. Extend and test the experimental native/loader guards for s=15 (30-bit packed
   extension elements and N=2^30 still fit the current element representation).
   Review numerical time/RAM/storage limits before launching; retain checkpoint
   resume, profiling and sampled RSS guards.
4. Run exact discovery/preparation and three encodings, then one fresh cached
   process for online RSS. Persist constants/profiles and update evidence only
   after correctness checks. If live preparation exceeds the local budget, use
   the measured shortfall to choose bounded external memory or a larger commodity
   RAM machine. s=16+ architecture decisions follow those results.
