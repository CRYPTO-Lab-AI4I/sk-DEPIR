# Encoding scaling study

The [encoding scaling reference](encoding-status.md) is the maintained source for
preprocessing and CPU/GPU results, memory/runtime estimates and multi-GPU capacity.
The performance goal is **reasonable performance through s=18 inclusive** on this
HPC system; retain **s=19–20 as exploratory targets**. Update its tables with every
relevant result or design change. Current standalone mathematical encoding is
validated through s16; s≥17 requires implementation and validation. Numerical
latency targets and larger experiment budgets still require agreement.

Retain CPU and optional GPU paths, Go preprocessing and independent correctness
references. Preserve the information set and canonical field/grid representation.
The sections below retain the study's chronological plans and evidence, including
the laptop s≤15 checkpoint and [larger-host handoff](development-handoff.md).
Use the reference for current status and the
[current memory-route status](#current-memory-route-status) for implementation work.

Throughout, q=2^s, m=2, N=q² and K is the code's message dimension. Resource
figures describe **one field-element database encoded under one code**. Encoding
an application layout can invoke several such encodings and reuse code-only setup.

## Current execution plan for s17

The user approved the [CPU-first/multi-GPU proposal](encoding-status.md#aim-for-s17).
Proceed with independent implementation work while the existing table-gap jobs
run; their results remain untouched until the user's completion ping. Preserve
s≤16 CPU/GPU paths and code-cache compatibility throughout.

| Step | Work and acceptance gate | Status |
|---|---|---|
| 1. Width audit and scalar reference | Inventory coordinates, counts, field values and narrowing boundaries. Establish uint64 packed extension arithmetic through s20 against an independent oracle and s≤16 coordinate ordering, without a full-domain allocation | Scalar reference validated: job74290; width audit continues |
| 2. Native CPU arithmetic and transforms | Port extension kernels/plans to uint64 while retaining uint32 base inputs/outputs. Compare small dense/Go/native cases, >32-bit boundary witnesses, sanitizer checks and s16 regressions; measure kernel overhead before whole-run forecasts | Native recursion validated (74295); wide API/matrix/race checks pass, regression harness cwd fixed after74312; full corrected validation pending |
| 3. Representation and reusable code data | Extend coordinates/counts/cache formats and loaders explicitly. Keep old32-bit artifacts readable; reject truncation, unsupported widths and incompatible specifications. Verify hashes and round trips; preserve existing caches | Pending audit/native design |
| 4. s17 preprocessing admission | Discover exact K, verify polynomial identities, budget peak host/device memory, storage and setup time. Compare CPU and multi-GPU/streamed preparation; obtain reviewed numeric budgets before full setup | Pending wider implementation; capacity is not admission |
| 5. First full CPU baseline | Prefer resident reuse (~427GiB principal arrays) on one large-RAM node when the admitted budget fits. Validate systematic values, independent curves, hashes and34GiB packed output; separately measure code loading, encoding, packing, verification and peaks | Pending verified s17 code data and host budget |
| 6. Multi-GPU prototype and selection | Partition same-node arithmetic across GPUs; measure transfers, topology and synchronization. Compare fast-layout streamed constants (~256GiB aggregate/two-card capacity floor) with resident reuse (~398GiB/three-card floor), including all host/device lifetimes and headroom | Pending scalar/native interface; small prototypes may proceed before full preprocessing |
| 7. Full GPU comparison | Validate exact canonical/packed agreement with the CPU baseline and independently check correctness. Compare complete measured latency and deployment cost; keep CPU available without CUDA | Pending measured multi-GPU design and admitted budgets |
| 8. Progress toward s18 | Recompute preprocessing/encoding/communication capacities using measurements, then select same-node versus multi-node or memory-saving designs. Retain s19–20 exploratory rows | Pending s17 evidence |

Steps1–3 are implementation prerequisites, not authorization for larger allocations.
Any validation expected to exceed30seconds is submitted and handed off. Mode
combinations remain candidates where a real memory constraint justifies them;
they do not block a faster CPU baseline if it fits. Host-backed GPU and disk-backed
CPU remain optional fallback routes with explicit transfer/I/O/runtime penalties.
Unknown preprocessing peak and multi-GPU communication are separate gates. Update
[the maintained reference](encoding-status.md) after every relevant measured result.

Initial audit: the common component-field implementation already supports s2–20;
`orbitField`, `OrbitBasis` coordinates, transform/preprocessing arrays, native
C/Go bindings and CUDA descriptions still use32-bit packed extension values.
Count headers alone were widened earlier and do not solve34-bit arithmetic.
The [scalar-port experiment](../experiments/s17-arithmetic/report.md) starts step1
with a private64-bit packed adapter. Job74290 passes independent scalar and
legacy-ordering checks through s20; source/binary/evidence hashes reconcile.
This clears the scalar gate for native CPU work, without raising encoder limits.
The [native kernel prototype](../experiments/s17-native/report.md) shares recursion
across32/64-bit types. Job74295 passes120 optimized and120 sanitizer cases,
independent fixtures/oracles and shared CPU regressions;410 source files and all
recorded cache/evidence hashes reconcile. Next are wide interfaces/ownership and
parallel matrix execution; the [interface/matrix prototype](../experiments/s17-matrix/report.md)
now implements both with bounded ownership/range/oracle/sanitizer/regression checks
pending. Job74312 passes wide API/matrix/race checks but fails the shared test
harness working directory; a corrected fresh retry will complete the gate. No
full encoder or preprocessing admission is claimed.

## Five-step plan

| Step | Work | Required output | Status |
|---|---|---|---|
| 1. Resource accounting | Audit arrays, lifetimes, widths, accesses and preprocessing/output costs at s=13..20 | Source-linked inventory and reproducible analytical model | Complete: source audit/model |
| 2. Execution designs | Compare blocked in-memory and external-memory transforms/products, selective preprocessing and spectrum choices | Concrete candidates, cost models and explicit unresolved dependencies | Complete: candidates/gates |
| 3. Focused prototypes | Implement uncertain operations at manageable sizes under constrained RAM; compare with the existing code | Verified decompositions and measured memory/I/O, not just timings | Prototype round complete; performance gates open |
| 4. Local calibration | Measure useful-core scaling, sustained storage and compute/I/O overlap; assess GPU suitability | Calibrated resource model and reasons for any larger hardware test | Bounded local calibration complete; sustained/integrated gates open |
| 5. Design selection | Select from measured evidence; integrate and validate progressively at s=13,14,15 when budgets permit | Canonical outputs, measured resource use and next hardware recommendation | Local s=13/14/15 CPU integration complete; larger-memory design gates open |

The first local feasibility target, s=15, is demonstrated on the 48 GiB host.
The design must still account for s=16..20.
No runtime/RAM/disk budget is increased by this plan. Agree numerical budgets
before larger runs. Reuse comparable baseline results, durable preprocessing and
profiles. Keep setup, cache loading, database encoding, output and validation
separate. Measure temporary storage and bytes transferred as well as peak memory.

The initial local study justified its prototype designs. The current scaling
objective extends performance work through s18; s19–20 remain exploratory rather
than required implementation milestones. See the maintained reference above.

## Current step: larger-memory preparation efficiency

The verified s15 checkpoint and Linux cached baseline remain the starting point.
The first full s16 attempt completed discovery but was stopped because preparation
progress exceeded its budget. Support is retained; see the
[stop report](../experiments/hpc-s16/report.md). Private base-field lookup now
improves complete s10 preparation by 1.56× and the measured s16 leaf kernel by
3.55×; see the [matched comparison](../experiments/preparation-arithmetic/report.md).
Full s16 preparation now completes in the actual GPU-assisted run below; s16
database encoding remains unverified. Expand the current
design-selection/scaling step with these sub-steps:

1. **Targeted profile.** Separate root generation, small polynomial products and
   larger native convolution/transform work. Use bounded s16 kernels and existing
   complete smaller-run profiles; sample saved s16 support without repeating
   discovery. Retain exact sources, checksums, CPU profiles, stage clocks and RSS.
   A partial kernel profile cannot establish the full s16 stage proportions.
2. **Specialized CPU kernels.** Compare reduced Go dispatch overhead and native
   C++ small-product loops with the lookup baseline. Preserve the field basis,
   exact coefficients and full identity/output checks.
3. **Bounded parallel batches.** Evaluate independent root/factor batches on
   multiple CPU cores, accounting for per-worker buffers and coordinated merges.
   Do not share the current single-caller convolution workspace across workers.
4. **Root-generation reuse.** Test cheaper powers/inverses and reuse between
   sorted exponents, verifying every generated factor against the current method.
5. **Matched GPU trials.** Verify an available GPU allocation/toolchain, then
   compare batched root/product and transform kernels against CPU alternatives.
   Include device setup, allocation, transfers, synchronization, reusable state
   and complete pipeline effects; record device memory and deployment requirements.
   GPU benefit is unmeasured and is not assumed from hardware specifications.
   Retain a practical CPU path; no backend or ordering is settled in advance.
6. **Select and admit.** Choose from kernel and complete smaller-code evidence,
   then measure a bounded integrated s16 preparation interval using reviewed saved
   support provenance. Reassess time/RAM/I/O budgets before another full run;
   preserve preprocessing/profiles and publish no full-feasibility claim early.

Sub-step 1 now passes in the [targeted profile](../experiments/preparation-profile/report.md):
sampled Go leaf/merge work and native transforms remain distinct opportunities.
Sub-step 2 now has [CPU kernel and crossover results](../experiments/preparation-kernels/report.md):
direct private Go leaf loops are integrated and full smaller-code checks pass.
Global cutoff remains unchanged; optimal choices depend on field/shape/workers.
Root generation was smaller than product work in this
sample, so defer sub-step 4 until product improvements are checked unless later
integrated evidence changes the ranking. The GPU tile results below justify a
complete-transform comparison; benefit for that full operation remains unmeasured.
Later sub-steps can be
reordered or omitted with recorded measurement-based reasons.
Initial profiling budget: one CPU-only Slurm job,
four cores, 16 GiB requested, 12 GiB sampled process-group RSS guard, 30-minute
allocation and 25-minute controller deadline. Cap representative convolution
buffers at 2^22 points and sampled support at 65,536 exponents. No full s16
executor allocation, discovery, or preparation is part of that job. GPU trials
need a separate explicit device/workload budget recorded before submission.

### Next comparisons: CPU products and GPU transform tiles

The targeted profile supports these concrete comparisons within sub-steps 2/5:

1. Measure lookup-specialized Go versus native C++ polynomial loops with the
   same deterministic coefficients and exact output comparison. Sweep the
   existing Go/native-transform cutoff on representative s16 product sizes;
   include allocation, plan setup and complete smaller-code checks before choosing
   a production setting. Keep generic arithmetic as the correctness reference.
2. Validate one available GPU and its compiler/driver. Prototype batched affine
   transform tiles over the existing packed extension field, matched exactly to
   native CPU tiles. Sweep tile sizes and total buffer sizes up to 2^22 values;
   separately measure initialization, reusable tables/buffers, host/device copies,
   forward/inverse kernels, synchronization and complete transfer-inclusive calls.
   These tiles are components of the larger transforms, not a full GPU matrix/
   Taylor backend. Report that scope even if the kernel is faster.
3. Use both results to choose the next integration. GPU tile gains alone do not
   justify moving the full pipeline or predict full s16 runtime. If the GPU
   prototype loses, report its algorithm/layout limitations rather than treating
   the result as a limit on other GPU implementations.

Trial budget per job: four CPU cores, 16 GiB requested RAM, 12 GiB sampled process
RSS guard, 30-minute Slurm allocation and 25-minute controller deadline. GPU trial
requests one device and caps explicit device allocations at 512 MiB. Keep all
large-code caches intact and retain generated fixtures/profiles/binaries outside
Git with checksums. No full s16 run is included. CPU and GPU experiments can run
independently under these per-job bounds; log every attempt and any correction.

### Completed comparison checkpoint and next integration

The CPU trials pass specialized Go/C++ kernels and three cutoff settings in
complete s10 codes. Direct private Go leaf loops are selected; they avoid adding
a native-call API. The unchanged cutoff still needs wider field/shape calibration.
The [GPU tile trials](../experiments/gpu-transform-tiles/report.md) also pass all
exact forward/inverse comparisons. On one H200 NVL, warmed transfer-inclusive
tiles at 2^22 values are 19–25× faster than the four-core CPU reference using the
same lookup arithmetic. Context initialization is separate; no whole-encoder
speedup is inferred. This completes the initial component trial in sub-step 5.

The composition checkpoint covers matrix/Taylor conversion, transposes and
persistent buffers. It measures complete transforms and smaller-code preparation
against the improved CPU path, including setup, transfers and CPU core scaling.
Checks cover row and arbitrary column bases, canonical outputs, memory ownership
and repeated calls. Its acceptance work is:

1. Compose canonical row tiles and arbitrary-basis column tiles with exact
   h-adic Taylor conversion and transposes. Verify independent small Horner
   outputs, complete native CPU outputs and forward/inverse round trips.
2. Own tables/plans and device workspace once per executor. Keep buffers resident
   across forward/product/inverse sequences where possible, with explicit
   ownership and no retained Go input pointers. Repeated calls must not reuse
   stale values. Reject shapes outside the admitted bound.
3. Measure complete transforms and resident convolution against 4/16-core CPU
   references, including initialization, plan preparation, transfers, staging,
   synchronization and memory. Preserve per-stage diagnostic timings/profiles.
4. Connect the experimental executor to complete s10 preprocessing/encoding with
   the existing exact identity, systematic, curve, output-hash and readback checks.
   Compare CPU and GPU-assisted runs and fresh cache use; leave ordinary CPU
   constructors available. Record CPU/GPU boundaries and all remaining transfers.
5. Select the next integration from complete pipeline evidence. A full s16 attempt
   requires a separate reviewed budget/provenance decision after these checks;
   tile or transform gains alone do not establish its runtime or feasibility.

Initial composition budget per job: one GPU, 16 CPU cores for 4/16-core comparisons,
16 GiB host RAM requested, 12 GiB sampled process RSS guard, 512 MiB explicitly
allocated device buffers/tables, 30-minute allocation/25-minute controller deadline,
and at most 2^22 transform values. CUDA driver/context memory is separate from
explicit buffers and must be reported. Preserve caches, source/binary hashes and
all attempts; no full s16 discovery/preparation is included.
The [composition experiment](../experiments/gpu-matrix/report.md) now passes items
1–4 and supplies the selection evidence for item 5: complete forward/inverse
transforms, arbitrary column bases, persistent buffers, resident convolution,
independent small Horner/Taylor checks, small CUDA memcheck and exact full s10
preparation/cached outputs. At 2^22 values, complete forward transform time is
0.137 s versus 0.760 s on 16 CPU cores; resident convolution is 0.379 s versus
2.737 s. GPU-assisted s10 preparation is 2.719 s versus 6.806 s, while complete
worker observations including setup/checks/I/O are 6.790 s versus 10.743 s.
These instrumented measurements apply to one H200 NVL, not consumer GPUs or full
s16. The default CPU constructors and complete CPU online encoding remain intact.

The [resident integration](../experiments/gpu-products/report.md) now passes
arbitrary/unequal operands, low prefixes, padding, overlap-add, old-library
compatibility, CUDA memcheck and complete smaller pipelines. At the same cutoff,
s10 preparation falls from2.735 to2.400 s; cutoff1024 measures1.528 s versus
6.647 s CPU. Default CPU constructors and the global cutoff are unchanged.

The [bounded s16 measurement](../experiments/s16-resident-intervals/report.md)
also passes exact partial products. Two consecutive2^19-position root-product
intervals total36.691 s on the current CPU path,20.205 s resident at cutoff1024,
or17.267 s at4096. The latter is2.12× faster for this early work; artificial
partial finalization is separate. The census captures exact root-batch degrees
and a logical later-operation inventory without rediscovery or large allocations.

Next estimation gate: later reciprocal/identity/spectrum need transforms up to
2^32 points, outside the measured/admitted2^22 CUDA prototype cap. Extend and test
the private domain count/17-level row metadata or select a validated blocked
alternative, record its memory/workload budget, then measure representative later
sizes and test extrapolation across scales. The existing C ABI call lengths are
already size_t and field values remain uint32. This limit is not evidence that
the hardware cannot fit s16; no whole-run speedup or deadline is inferred yet.
Keep CPU batch/root alternatives eligible as profiles change. Full s16 readmission
still needs a separate budget/provenance decision; do not start it automatically.

Proceed in this order:

1. Extend private GPU plan counts through 2^32 and row descriptions through
   17 levels. Keep field values at 32 bits, reject unsupported shapes, and retain
   old-library compatibility. Test the full-domain description without allocating
   a full-domain workspace.
2. Review workspace growth and peak allocation, including temporary replacement
   buffers, tables/plans and separate driver overhead. Compare contiguous device
   execution with a bounded-block alternative before requesting larger budgets.
3. Validate and measure progressively larger forward/inverse transforms and
   resident products against CPU/reference outputs. Initially retain the existing
   512 MiB explicit device, 12 GiB RSS and 25-minute controller limits; any shape
   increase must fit those bounds or have a new recorded budget before submission.
4. Combine measured scaling with the exact s16 operation inventory, CPU root work,
   reusable setup, memory and I/O. Report an estimate with its unmeasured portions
   and uncertainty; a full s16 run remains a separate admission decision.

Gate1 and the initial bounded portion of gates2/3 now pass in the
[GPU scaling experiment](../experiments/gpu-scaling/report.md): versioned2^32
counts,17 row levels, allocation-free metadata validation and exact CPU/GPU
operations through2^25. Per-engine limits preserve the512 MiB cap. Releasing old
workspace before growth admits384 MiB array payload at2^25 without replacement
buffer overlap. The full-domain contiguous payload model is48 GiB for three
arrays; this excludes tables/plans/context and host arrays. The remaining choice
is a larger explicitly budgeted contiguous test or a validated blocked schedule.
Neither full-domain execution nor the whole-run estimate is complete.
User approved the next bounded increment:2^26..2^28 points,4 GiB explicit
device cap,16 GiB host request/12 GiB sampled RSS,one GPU/16 CPUs,30-minute
allocation/25-minute controller. This admits only that operation sweep, retaining
all output checks/profiles and resource guards; full s16 remains separate.
The approved sweep now passes: at2^28, GPU forward2.651 s vs16-core native
CPU65.114 s, inverse2.533 s and resident product6.669 s. Exact full-array checks
pass. Explicit device peak3.001 GiB and sampled process RSS4.141 GiB remain under
the approved bounds. See the scaling report for CPU-product methodology and
provenance. Four doublings remain to2^32; review the next device/host/time budget
and validation schedule before submitting another increment. Full preparation
speedup remains unmeasured.

User selected the [actual guarded preparation](../experiments/gpu-preparation/report.md)
instead of further synthetic sweeps. Reuse saved s16 support, integrate full-domain
resident products/spectrum, validate complete smaller preparation plus normalized
root checkpoint replay, then launch preparation-only under one GPU/32 CPUs,
256 GiB host request/224 GiB sampled RSS,128 GiB Go soft target,64 GiB explicit
device cap,64 GiB new cache/128 GiB scratch and a24-hour allocation (23.5-hour
controller/23-hour worker). These limits are for the actual attempt; they are not
a runtime forecast. Preserve profiles and completed stages; stop after constants
are saved. Full s16 completion and database encoding remain separate outcomes.

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

## Next step: complete C++ online encoding

**Implemented and validated** in the [native-online experiment](../experiments/native-online/report.md).
At that checkpoint, all seven cached s=14 words matched and peak RSS remained
about 4.59 GiB. Its paired runs gave a descriptive 6.3% time reduction, qualified
by clock discrepancies and host variance. The timing verification below now
provides a consistent comparison; the original measurements remain unchanged.

The recorded step before s=15 admission was to move the remaining database-dependent work into C++:
buffer preparation, pointwise field products and canonical output assembly.
Reuse the existing parallel Taylor transforms, fixed-constant maps, blocked
transposes and explicit workers. One coarse native call borrows pointer-free
input, preprocessing and workspace arrays; no database pointers are retained.
Discovery and reusable code preprocessing remain Go.

Validate identical full codewords against the Go path and small independent
references. Profile cached s=14 with the retained preprocessing, reporting setup,
interface/copy costs, encoding stages and fresh-process peak memory. Use this
complete native online baseline to choose further optimizations. Mathematical
encoding remains separate from cryptographic/application encoding.

## Timing verification before further tuning

The [timing experiment](../experiments/encoding-timing/report.md) brackets complete
encoding with the native stage clock and independently records Go/POSIX/wall
intervals, process CPU time and peak RSS. Short busy/sleep controls and small
encoding fixtures must pass before the cached s=14 comparison. Native stage sums
must fit their same-clock total. Three ordered pairs alternate native/hybrid
execution with identical inputs, workers and profiling, reusing one plan and
workspace; retain the first cold-workspace result and all later measurements.
The completed study has agreeing large-run clocks (within 0.376 ms) and matching
hashes. The first pair precedes/overlaps an observed battery-to-AC transition;
the two subsequent AC pairs give 84.47 s full C++ versus 94.55 s hybrid, about
10.7% less elapsed time with similar CPU cost and 4.59 GiB process peak RSS.
These are profiled, warm-workspace observations, not a universal speedup or a
historical root-cause diagnosis. Reuse this consistent baseline for further
measurements. Code discovery/preparation was not repeated; the next scaling
work remains the s=15 preparation-memory audit below.

## s=15 preparation-memory audit checkpoint

The [audit](../experiments/preparation-memory/report.md) is complete; no s=15
preparation or encoding was run and its guards remain unchanged. Source-based
simultaneous coefficient arrays are about 21.62 GiB, or 22.02 GiB with a support
capacity scenario. Cached online arrays are about 18.42 GiB. These are analytical
payloads using estimated K, not RSS bounds. Scaling s=14 RSS by four does not
establish a requirement for a larger machine.

Small layout measurements identify a projected 4.54–6.02 GiB validation map and a
4.65 GiB retained-factor scenario. Those objects belong to earlier phases and
must not all be added to the final FFT peak. A proposed membership bitset needs
128 MiB. The existing s=14 endpoint/CPU profiles do not measure in-stage live
heap, so they cannot establish that unchanged s=15 preparation fits the 32 GiB
RSS guard.

Compact support validation, bounded factor batches, private reusable FFT buffers
and fresh Newton-result reuse are now implemented. Small reference comparisons
pass. The [targeted s=14 checkpoint](../experiments/preparation-checkpoint/report.md)
passes complete P/G/spectrum digests from its retained basis. Diagnostic
preparation took 34m 32.5s, with 9.01 GiB process peak RSS and 5.42 GiB peak
reachable Go heap at checkpoints; reusable convolution buffers were released
before spectrum generation. Forced GC/profiling and loading the retained basis
change the measurement boundary from the historical run, so this is not a
controlled latency comparison. The live heap is consistent with the roughly
22 GiB s=15 principal-array model, but does not prove a 32 GiB RSS bound. Review
s=15 guards before launching; the 48 GiB host remains a plausible target.
HPC/GPU or external-memory work is not yet required by the evidence.

## Subsequent admission: s=15

1. Use the completed audit and diagnostic s=14 live-heap/RSS measurements to
   review s=15 time/RAM/storage limits. Online encoding suggests roughly 6–9
   minutes and 18–22 GiB RAM on this host; these remain planning estimates.
   Preparation has a roughly 22 GiB principal-array model plus runtime/native
   overhead. Neither multiplying historical RSS nor extrapolating live heap
   alone establishes its process RSS.
2. Retain the implemented compact membership, bounded factor batches and
   preparation-only scratch reuse. Full s=14 constants and small independent
   references pass; use the profiles to identify any additional memory change
   needed before admission. Make no change solely to fit proportional RSS.
3. Extend and test the experimental native/loader guards for s=15 (30-bit packed
   extension elements and N=2^30 still fit the current element representation).
   Review numerical time/RAM/storage limits before launching; retain checkpoint
   resume, profiling and sampled RSS guards.
4. Run exact discovery/preparation and three encodings, then one fresh cached
   process for online RSS. Persist constants/profiles and update evidence only
   after correctness checks. If live preparation exceeds the local budget, use
   the measured shortfall to choose bounded external memory or a larger commodity
   RAM machine. s=16+ architecture decisions follow those results.


### Local s=15 admission experiment

The [experiment](../experiments/parallel-s15/report.md) implements the next local
attempt: experimental preparation/native/loader guards allow s=15, while public
DiscoverOrbit remains at s≤13. Small-buffer 30-bit arithmetic, loader shape caps
and full suites must pass before launch. The run separates exact discovery,
reusable preparation, artifact persistence, three complete native encodings,
packed I/O and fresh-process cached execution. Preparation retains internal
heap checkpoints and diagnostic GC; encoding does not add forced internal GC.

Use a 24 GiB Go soft target, 32 GiB process RSS guard, 20 GiB new cache budget
and initially four hours of active compute. The current initial run has a
[user-authorized 5.5-hour override](../experiments/parallel-s15/report.md#live-time-guard-extension),
without restarting the worker or changing RAM/cache limits. The s=14 measurement suggests s=15 preparation
alone may exceed two hours; this extends time, not the memory limit. Low battery
pauses preserve running memory; completed stages persist across process restarts.
A guard stop is evidence of the encountered resource limit, not success or proof
that a larger commodity machine is required. Full results are now recorded in the completed checkpoint below.

The completed s=15 attempt [resumed after its manual sleep break](../experiments/parallel-s15/report.md#resumed-after-manual-break)
on the same worker, under a new active-time monitor. Break time is excluded from
the 5.5-hour budget. Do not rerun its resume command or resume the old fixed
monitors directly. Preparation latency must be labeled interrupted; later online
measurements remain a separate boundary.


## Completed local s=15 checkpoint and next step

The [first-use and cached runs](../experiments/parallel-s15/report.md) pass for
s=15,m=2,t=5, with exact K=216,914,047 and N=2^30. Three complete native encodings
take 391.35/393.19/396.48 s; fresh cached encoding takes 393.69 s and reproduces
the full output hash. Whole-process peaks are 20.55 GiB first use and 18.46 GiB
cached. All systematic positions, sampled independent curves and packed readback
pass. Reusable code files total about 8.81 GiB, with a 4.81 GiB minimal online
file. Profiles, exact sources and cache hashes remain retained outside Git.

This demonstrates this standalone configuration locally, completing the s=15
admission gate described above. Earlier resource models/forecasts remain historical
scenarios. Preparation spans a manual sleep pause; its raw elapsed time is not an
uninterrupted latency baseline. Public discovery defaults still stop at s≤13;
application-plan/cryptographic execution, TRL 5 acceptance and s≥16 remain open.

Next, design the s=16 memory schedule before another full run. Audit the retained
s=15 profiles and simultaneous buffers, then specify which arrays can be reused,
fused or processed in blocks. Address 2^32 cardinalities/indexing separately from
field-element width. Compare a bounded-RAM design with a larger commodity-RAM
baseline, including I/O and preprocessing costs; do not assume more kernel tuning
alone solves the memory gap. Validate any selected design against retained s=15
words/constants and small references before requesting a new large-run budget.


## Development moves to a larger-memory host

Local laptop development stops at s=15. The [development handoff](development-handoff.md)
is the continuation entry point: begin on a single shared-memory commodity CPU
node, provisionally 256 GiB RAM, validate the host/toolchain and transferred s=15
cache, then perform the s=16 audit above. The [destination receipt](../message-board/2026-11-09-1316-codex-orient-001-completed.md)
now verifies transferred reference data; [HPC setup](hpc-setup.md) records the
toolchain and scheduler inventory. A [cached CPU baseline](../experiments/hpc-cached-s15/report.md)
now passes full suites and s=15 output checks on a compute node. Larger s remains
unverified; the move does not select distributed
HPC or a GPU, change protocol scope or require every field to use specialized hardware.
Keep the generic local path and references intact, and retain all source-machine
caches/profiles until verified transfer. New host/job budgets replace laptop values
only after resource review; no additional large laptop run is planned.


## Actual GPU-assisted s16 preparation checkpoint

The [actual preparation report](../experiments/gpu-preparation/report.md) records
job72875 COMPLETED/0:0:2h20m15s preparation,46.0 s final persistence and2h24m15s
allocation elapsed. Saved support was reused; no rediscovery. Full P*G identity
passes, normalized P and final G/spectrum are retained. Sampled process RSS72.575
GiB and new cache35.430 GiB stay below the admitted guards. Large artifact hashes
were computed by the controller; source/log/metadata reconciliation passes.
No full-domain synthetic sweep was needed. Next is separately admitted cached
s16 database encoding and its output checks; do not restart preparation or launch
encoding automatically. A matched full CPU baseline/whole-run speedup, application
integration and TRL5 acceptance remain unverified.


## CPU encoding profile after preparation

The [cached s14 profile](../experiments/cpu-encoding-profile/report.md), job72955,
passes six full canonical hashes. Median encoding248.7/140.7/91.2 s at8/16/32
workers; Taylor49.9/40.9/38.9 s. At32 threads Taylor consumes42.7%, while rows and
columns/transposes scale much better. Prioritize within-block CPU Taylor
parallelism, using the validated gather algebra as a candidate and the current
recurrence as reference. Do not first optimize the small buffer/pointwise/assembly
stages from this evidence. No implementation selected or new job submitted;
s16 memory-pressure/encoding validation and GPU-online integration remain open.


### Current CPU step: parallelize inside Taylor blocks

1. Implement the two-pass gather decomposition in native C++: correct the low
   quotient coefficients from the untouched upper quotient, then compute the
   remainder. Reverse the pass order for inverse conversion. Preserve the
   existing recurrence as a one-worker/small-work reference, constant-product
   maps, field coordinates and caller-owned in-place storage.
2. Split coefficient ranges inside large blocks when independent block count
   underuses requested workers. Each destination must have exactly one writer;
   join workers between dependent passes. Avoid a full-size extra buffer and
   retain coarse block scheduling where it already offers enough parallelism.
3. Validate against independent sequential polynomial division, single-worker
   transforms, forward/inverse round trips and complete canonical codewords.
   Cover varied splits, sparse/dense lower terms, zero/one coefficients and
   multiple worker counts. Run sanitizer checks plus existing native suites.
4. Compare published recurrence and candidate in one CPU allocation at16/32
   workers with two complete cached s14 encodings per case. Retain code/data
   hashes, setup, CPU profiles, stage clocks, RSS and full output checks. Select
   from whole-encoding and Taylor timings, including any regressions or extra
   scheduling/memory cost; do not assume gather is faster on CPUs.
5. If accepted, update the implementation/docs and reassess one cached s16
   encoding under a separate reviewed memory/time budget. No GPU-online claim
   or s16 speedup follows from the smaller run.

Initial validation/comparison budget:one CPU-only32-core job,16 GiB requested
host RAM,12 GiB sampled process RSS,30-minute allocation/25-minute controller,
8 GiB new-cache cap. Existing preprocessing/profiles remain intact. Build old
and candidate from recorded exact sources; no discovery or preprocessing.
Submit once prepared, then hand off under rule8 and wait for the user's completion
ping before analysis or dependent work. The accepted comparison below closes
the smaller-code implementation gate; larger sizes retain their own validation gate.


CPU-023 implementation and correctness/comparison gates1–4 now pass in
[cpu-taylor](../experiments/cpu-taylor/report.md), job73099/taylor-01. Full suites,
independent Taylor/matrix tests, ASan/UBSan/TSAN and eight complete s14 hashes pass.
At16 workers encoding142.2 to118.5 s (1.20x), Taylor39.4 to15.4 s; at32 encoding
87.9 to61.0 s (1.44x), Taylor35.7 to8.8 s (4.07x). Other transform stages remain
close. Select the hybrid within-block schedule, preserving the recurrence for
one-worker/small/coarse-parallel cases. Gate5 remains: review budget and validate
one cached s16 encoding; do not assume this ratio at full size or launch it
automatically. No new job was submitted during result review.


User authorized [one cached s16 CPU encoding](../experiments/cached-s16-encoding/report.md)
after publishing the selected Taylor implementation. One32-worker/no-GPU job,
256 GiB host request/224 GiB sampled RSS,2-hour allocation/110-minute controller/
90-minute worker,16 GiB new-cache/scratch guards. Reuse verified actual-01 online
constants; focused native checks and tiny cached smoke precede full systematic/
curve/packed-readback checks. No code preparation or automatic second encoding.
The attempt now passes as recorded below; any further job must be handed off
under rule8 after submission.


## Completed cached s16 CPU encoding

[Encoding-01](../experiments/cached-s16-encoding/report.md), job73131, passes:
32 CPU workers/no GPU,20m52.503s encoding and34m53s job elapsed. All918,672,421
systematic positions,16 sampled independent curves and1027 packed readbacks pass.
Peak sampled RSS74.370 GiB, packed output8 GiB retained with hashes/profiles.
The accepted Taylor schedule now executes full s16; no matched old-code s16
speedup ratio or independent complete-reference word was measured. Stage costs
are rows44.1%, columns/transposes36.0%, Taylor16.0%; profiling alone does not pick
a new optimization. Source/log/metadata reconciliation passes, with the packed
digest taken from completed controller hashing rather than a fresh bulk reread.
One cached mathematical run closes this admission gate, not fresh-repeat equality,
application/cryptographic integration or TRL5. No further job is auto-launched.


## Completed step: optional primary CUDA client encoding

Retain CPU (Go/C++) and CUDA as peer client paths; preprocessing origin does not
select the client's encoding hardware. Promote reusable CUDA code to
encoding/liftrs-cuda, leaving historical experiment shims/provenance intact.
Share field/matrix/code data across backends without linking CUDA into ordinary
CPU builds or requiring liftrs_cpp for CUDA-only builds. Parameterize by s within
the current2..16 packed arithmetic bounds, with s16 as the full-size target.

1. Add WithCUDA to verified online plans, explicit library/device cap, owned
   resident code/workspace, serialization/Close and CPU-only unavailable behavior.
2. Keep all five transforms, products and canonical assembly resident between
   message upload and final word download; reuse code constants and buffers.
3. Validate across small s against independent Go and optional native CPU paths,
   repeated/zero/invalid inputs, error recovery and resource release. Preserve
   CPU-only/CGO-disabled builds, old GPU library compatibility and existing
   transform/preprocessing regressions. Record variable-time arithmetic limits.
4. Submit one bounded validation job (one GPU/16 CPUs,16 GiB host/12 GiB RSS,
   512 MiB explicit device,30-minute allocation/25-minute controller). Hand off
   under rule8; no dependent full run until the user's completion ping.
5. After validation, separately admit cached s16 GPU encoding against the retained
   CPU canonical/packed hashes. Account for persistent19.42 GiB code plus48 GiB
   arrays (~67.42 GiB total payload); provisionally96 GiB device cap and explicit
   host/time limits. Include setup/transfers/checks/reuse, and distinguish any
   historical CPU timing comparison from a matched allocation. No full run or
   speedup claim before admission and verified completion.

This is standalone mathematical client encoding, not full application/cryptographic
encoding or server-side query computation. CPU remains a maintained main path.

Validation job73243 passed (see the [report](../experiments/cuda-encoding/report.md)):
small independent Go matches, complete s10 repeats, build compatibility, race/vet,
historical matrix checks and zero CUDA memcheck errors/leaks. Steps1–4 are complete.
Step5 is admitted in the [full cached s16 runner](../experiments/cuda-s16-encoding/report.md):
one GPU/16 CPUs,96 GiB device cap,128 GiB requested host RAM/96 GiB RSS guard,
64 GiB Go soft limit,2-hour allocation/110-minute controller/90-minute worker.
It performs two same-message encodings through one session; no preprocessing or
CPU re-encoding. Full-size performance and output agreement are recorded below.


Step5 completed in job73257: [evidence](../experiments/cuda-s16-encoding/report.md)
passes full CPU canonical/packed hash agreement across both GPU calls. Encoding
85.954/60.216 s, reusable setup7.193 s, owned device67.423 GiB, host RSS39.024 GiB,
30m02s whole validation allocation. No guard tripped. The earlier32-thread CPU
encode was1252.503 s; ratios14.6–20.8x are historical, not matched benchmarks.
No next experiment is auto-launched. Runtime opportunities remain in matrix
row/column kernels, output-buffer/transfer reuse and packed writing. The next step
below prioritizes memory feasibility before standalone runtime tuning.
First-download overhead needs diagnosis before attributing its cause.
Application/cryptographic execution and TRL5 remain separate.


## Current step: memory optimization for s17

User priority: reduce memory requirements before further standalone runtime tuning.
Retain the verified s16 CPU/GPU encoders as baselines and maintained main paths.
Use the [capacity reference](encoding-status.md#machine-capacity-and-feasibility)
when evaluating larger host memory and multiple GPUs/nodes. Their observed
availability does not authorize larger jobs; distributed execution needs its own
implementation, communication measurements and budget review.
Code parameters and output representation remain independent of backend choice.

The [catalogue](../pirconfig/configs_catalogue.json) estimates K3,870,933,765 for
s17,m2,t5; N=2^34 is exact. Relative to s16, N grows fourfold and internal field
values grow from32 to34 bits. A straightforward64-bit internal representation
therefore makes major arrays about eight times larger; base values still fit
uint32 and final packed output is34 GiB. The existing GPU layout would need
roughly530–550 GiB device memory plus220–260 GiB host RAM; the CPU path roughly
500–600 GiB host RAM including allowance for overhead. These are rough port
models, not measured peaks, memory bounds or admitted budgets. In particular,
current arithmetic does not support s17 and this layout exceeds the measured
H200's roughly140 GiB device capacity. Preprocessing requires its own analysis.

1. **Audit memory by phase.** Count sizes and overlapping lifetimes for discovery,
   preparation, save/load, reusable constants, encoding workspace, canonical
   assembly, packing and validation. Separate peak host RAM, pinned host buffers,
   owned GPU allocations/context, persistent storage and temporary disk. Compare
   modeled s16 lifetimes with measured baselines before projecting s17. Identify
   which34-bit quantities require wider arithmetic and which base values can
   retain32-bit slots; do not widen all arrays indiscriminately.
2. **Compare bounded-memory designs.** Evaluate buffer reuse and earlier release,
   in-place transforms, streamed or recomputed code constants, tiled CPU/GPU
   transforms and streamed packed output. Consider compact34-bit or split-word
   internal layouts where useful, accounting for unpacking and arithmetic costs.
   Streaming the final word alone cannot resolve the transform-workspace gap.
   Whole-domain operations must retain correct padding, ordering and dependencies;
   tiled kernels alone do not establish a bounded-memory encoder.
3. **Present the memory–time choices before selecting one.** For each viable
   option report peak host/device memory, disk requirements, extra CPU/GPU work,
   transfer/I/O bytes and passes, setup cost, first-message and reused-session
   encoding time, and implementation/deployment costs. Distinguish measurements,
   estimates and unknowns. State explicitly when lower GPU memory moves data to
   host RAM/disk or requires recomputation. Show the alternatives and recommendation
   to the user before adopting a material memory–time trade-off; analysis may
   continue without selecting that trade-off. No silent slower default, larger
   resource budget or replacement of a maintained backend.
   Preserve the faster validated mode as default whenever it fits the admitted
   memory budgets. Slower memory-saving modes address insufficient memory; when
   memory is sufficient they require explicit command-line opt-in. See the next
   part below for shared CPU/GPU work and the selection rule.
4. **Validate the chosen schedule on s16 first.** After the choice, implement the
   smallest representative prototype with explicit memory caps, profiles and
   retained reusable data. Check CPU/GPU canonical and packed hashes against the
   saved baseline; measure setup, transfers, encoding and I/O separately at matched
   budgets where practical. Test the claimed peak-memory reduction, not only kernel
   latency. Each experiment needs its own runner/report/results and any job over30
   seconds must be submitted and handed off under rule8.
5. **Gate s17 implementation and admission on the evidence.** Review actual host
   and device capacity, the selected memory schedule and its measured time penalty
   or benefit. Then extend field arithmetic, indices/metadata and tests for34-bit
   values; audit preprocessing and storage as well as client encoding. Agree on
   full-run resources/deadlines before submission. Update runtime estimates using
   representative wider-arithmetic/streaming measurements; the earlier resident
   GPU projection cannot predict a streamed encoder's runtime.

Packing/writing and GPU row/column speed work follow memory feasibility unless
needed by the chosen memory schedule (for example, streamed packed output).
Maintain shared canonical field/record conventions and independent CPU builds.
This step does not establish s17 feasibility, change PIR/security contracts or
complete application/cryptographic integration or TRL5 acceptance.


### Memory audit checkpoint and choice

The [source/measurement audit](../experiments/encoding-memory/report.md) and its
[allocation-free model](../experiments/encoding-memory/run.py) reconcile large s16
CPU/GPU arrays with saved measurements and inventory discovery/preparation/output
lifetimes. This completes the initial inventory and alternatives comparison;
transient preparation peaks, pinned staging and proposed-schedule timings remain
unmeasured. No design selection, new budget or s17 support follows from the model.

The s17 current GPU principal model is526.420 GiB if base arrays stay32-bit;
sharing two work arrays lowers it to398.420 GiB. Buffer reuse alone cannot fit.
An in-place work array plus streamed code models130 GiB before overhead (needs
a new algorithm and separate cap/headroom review). Host/disk-backed tiles can
use less GPU memory, with potentially large transfer/I/O costs. Compact34-bit
storage alone still models286.420 GiB in the current layout.

Recommendation: validate two-work-array reuse and host-plan release first on s16,
with no added transforms/recomputation/transfer passes; keep resident baselines.
Alternatively prioritize the tiled s17 route if the user prefers a direct capacity
solution despite extra traffic. Present the in-place/tiled time-memory choices
before selecting one. The report quantifies modeled traffic sensitivity and disk
inventory, distinguishing all projections from measurements. Prototype and
long-job admission await the user's choice; no experiment is auto-launched.


User selected buffer reuse first. [GPU-029 prototype](../experiments/cuda-buffer-reuse/report.md)
implements the two-array schedule as an opt-in online-only mode; public CUDA
layout and CPU paths are retained. Validate small oracle/memcheck cases, matched
s10 default/reuse timing and two full s16 reuse calls under56 GiB device cap,
within previous host/time budgets. No added transforms or host transfers, but
address/allocation effects on latency must be measured. Fulls16 comparison with
job73257 is historical, not matched. No larger s17 memory–time route is selected.


GPU-029 completed: [job73576](../experiments/cuda-buffer-reuse/report.md) passes
323 source identities, all oracle/lifecycle/build checks,80 matrix records, zero
CUDA memcheck errors/leaks and full CPU canonical/packed hash equality across both
s16 calls. Owned device memory is51.427 GiB versus67.423 GiB (23.725% saved), exactly
one full array minus parity/error storage, under56 GiB cap. Encoding57.418/56.975 s;
matched s10 default/reuse differs by less than0.1%, with no observed slowdown.
The prior fulls16 timing is historical on another node. Host RSS39.031 GiB;
host-plan release remains unimplemented in this harness. Full job19m05s, no guards.
This closes validation of the selected buffer-reuse prototype, which remains
opt-in; no default switch is made by results review. The remaining s17 gate is
still a material memory–time choice (new in-place/streamed-code schedule or
host/disk-backed tiles), not more runtime tuning. No next job or budget increase
is auto-authorized by successful completion.


User authorized CPU memory optimization after publishing GPU checkpoint999f4a5.
[CPU-030](../experiments/cpu-buffer-reuse/report.md) transfers dead native transpose
scratch as owning output, detaching it from the workspace; future calls allocate
fresh scratch so old results remain valid. Exact scratch/output alias only is
admitted; other overlaps stay rejected. Same five transforms, no added copies or
recomputation. s16 principal CPU target74.267→58.267 GiB when old outputs are
released; retained caller results cost extra. Go reference/hybrid schedules stay
as before. Validate ownership/errors/concurrency, ASAN/UBSAN, matched s10 buffer
schedules and two s16 CPU calls under68 GiB RSS guard,128 GiB requested RAM,
32cores/noGPU,64 GiB Go soft and prior2h time/cache limits. Full-size validation
now passes as recorded below; no s17 route or runtime equivalence is selected by this step.


## Next part: shared CPU/GPU memory schedules

This planning track proceeds alongside the submitted
[CPU-030 validation](../experiments/cpu-buffer-reuse/report.md), job73586. Its
results remain pending until the user's completion ping; do not inspect the job,
logs or results while doing independent documentation/design work. Use completed
GPU-029 evidence and the source-based memory audit for current facts. New CPU
measurements must wait for CPU-030 before evaluating its actual savings/latency.

### Selection rule for memory modes

Prefer the faster validated implementation whenever it fits the admitted host
RAM and GPU-memory budgets, including runtime/driver/staging headroom. A memory
optimization that slows execution is a capacity solution, not a general default:
use it when the faster compatible path cannot fit. If memory is sufficient, that
slower mode must remain optional and require explicit command-line selection.
Keep the faster path available; saving memory alone does not justify a slowdown
for a run that already fits. Do not infer an improvement from smaller allocations.

Before adopting a trade-off, state which memory limit prevents the faster mode
from fitting and quantify saved host/device bytes, disk requirements, added
computation/transfers/I/O, setup and first/reused encoding latency. Distinguish
measurements, projections and unknowns; an unmeasured penalty is not zero.
Explain the selected mode/reason to the user and keep CPU/GPU backend selection
explicit. CLI option names and automatic capacity selection are not implemented
by this plan; future CLI exposure must obey this opt-in rule. Larger budgets
and material trade-offs still need the existing user decision/admission step.

### Ordered work for both backends

1. **Remove full-size transform scratch while transforms run.** Investigate
   in-place transposes, alternate layouts and bounded scratch tiles, starting
   with the decomposition shared by native CPU and GPU. Current reuse only
   reclaims scratch after transforms finish. Prove prefix/full and rectangular
   layout dependencies, ordering and alias lifetimes before changing kernels.
   CPU/GPU need separate execution kernels and locality measurements. Compare
   RAM/VRAM saved with extra passes, synchronization and slower accesses; retain
   the full-scratch fast mode if the smaller mode is slower.
2. **Stream reusable preprocessing constants.** Load spectrum/P chunks when
   needed instead of retaining full arrays. GPU host-to-device streaming saves
   VRAM but retains host backing and repeats uploads; CPU streaming saves RAM
   through disk reads, unless a different backing source is explicitly counted.
   Assess mapped-file residency, pinned staging, disk/page-cache accounting and
   cold versus reused code. Report lost setup amortization and recomputation
   alternatives. This step complements scratch reduction; neither alone has
   demonstrated s17 feasibility.
3. **Stream canonical/packed output.** Evaluate a bounded output sink that avoids
   retaining a full canonical word, accounting for scattered coordinate order,
   routing buffers, packed writes, checksums and validation/readback. Preserve
   existing owning-array APIs and their memory needs; introduce streaming as an
   explicit mode/API rather than silently changing returned-output lifetime.
   Measure extra routing/storage work and end-to-end latency, not just packing.
4. **Evaluate host/disk-backed transforms if resident schedules still do not fit.**
   Keep only working tiles on GPU or in CPU RAM while host RAM/disk backs large
   arrays. Count every transform/Taylor pass, staging buffer, temporary file and
   transfer, including page-cache/cgroup memory. Choose a bounded dependency
   schedule before estimating runtime. Larger host/disk traffic is an explicit
   penalty, not free GPU-memory savings. Preserve cache/checkpoint provenance.
5. **Combine and validate before s17 admission.** Compare the viable schedules
   under the same resource limits on small cases and s16, using exact baseline
   words, owning/streamed-output checks, profiles and total peak memory. Review
   the memory–time alternatives with the user before selecting a slower route
   or changing budgets. Then extend34-bit arithmetic/serialization/indexing and
   audit preprocessing separately. The roughly130 GiB one-array/streamed-code
   s17 scenario is a model with tight headroom, not an implemented capacity or
   runtime result. Slower routes remain capacity-driven or explicit CLI opt-ins.

This records the next work; no new implementation, CLI flag, job or resource
increase is introduced. Jobs over30 seconds keep the submit-and-handoff rule.
CPU and GPU remain maintained, explicitly selected paths; mathematical encoding,
application/cryptographic integration and TRL5 acceptance stay distinct.


GPU-032 starts the first shared-plan item on GPU: [transform scratch study](../experiments/gpu-transform-memory/report.md).
Direct rows operate in the original layout, removing full scratch and physical
transposes but potentially losing coalescing. A fair transform-only fast baseline
owns data+scratch; direct mode owns data only. CLI defaults fast and requires
`--layout direct`/`compare` for the potentially slower option. Validate square/
rectangular forward/inverse/native/Go matches and memcheck through2^25 values,
512 MiB device cap,16 GiB host/12 GiB RSS,16 CPU cores/one GPU,30m allocation/25m
controller. No online-default switch or fulls16 encoding: safe canonical output
still needs a separate memory schedule. CPU job73586 is left untouched and its
results awaited the completion ping. Both completed reviews are recorded below.


### Completed CPU-030 and GPU-032 checkpoints

User confirmed both jobs finished. [CPU73586](../experiments/cpu-buffer-reuse/report.md)
passes328 source identities, full repeated CPU/GPU canonical/packed hash agreement,
retained-output/overlap/error/concurrency tests, native/Go suites, race/vet and
ASAN/UBSAN. Sampled host peak58.351 GiB versus prior74.370 GiB,16.019 GiB lower,
matching the one-array saving. Encoding1281.328/1247.173 s, full two-call validation
allocation1h04m36s. Matched s10 times differ in both directions (+2.34% first,
-0.73% second); historical full-size comparisons also have both signs. No stable
runtime penalty/benefit or universal equality is established. Profiles/output kept.

[GPU73669](../experiments/gpu-transform-memory/report.md) passes333 source identities,
all square/rectangular forward/inverse/native/small-Go comparisons and zero CUDA
memcheck errors/leaks. Direct rows save exactly4N bytes and perform no transpose.
At2^25 values working arrays256→128 MiB, but forward/inverse medians are2.10%/2.82%
slower. Larger tested cases show roughly2–8% penalties. Keep fast as default when
memory fits; direct remains explicit CLI opt-in or a capacity solution. This
proves only the transform component, not fulls16 online execution. Job2m14s,
512 MiB cap, no guard tripped. CPU and GPU ran on separate allocated nodes.

Next GPU gate: design safe canonical assembly/output without recreating a full
scratch array after the transforms. Evaluate bounded routing/streaming or a proven
in-place schedule, then combine with code/spectrum chunking. One candidate is
narrowing the final base-field values before assembly, using bounded staging and
disjoint compact source/destination regions; it needs alias/dependency proof,
correctness tests and explicit accounting of compaction, conversion and I/O time.
It is an unimplemented design, not another memory saving or an approved default.
The extra runtime of direct rows and any output/code-streaming costs must be shown
before selecting an end-to-end slower mode. No new job or budget is admitted by
completion of these checkpoints. s17/preprocessing/cryptographic/TRL5 gates remain.


### Completed GPU-033 and CPU-034 compact-output gates

[GPU-033](../experiments/gpu-output-memory/report.md) passes job73815: two full
s16 compact calls match retained CPU/GPU words, all systematic positions, sampled
curves and packed readback. Oracle/staging-boundary/ownership checks,80 matrix
regressions, build compatibility, race/vet and memchecks pass with0 errors/leaks.
Device ownership is35.429 GiB versus51.427 GiB prior reuse (31.11% less), exactly
one16 GiB array removed plus2 MiB staging. Host peak remains39.028 GiB. Full
encoding takes101.526/93.332 s versus prior57.418/56.975 s on a separate allocation;
this historical comparison cannot isolate the penalty. Matched s10 compact is
9.60%/10.67% slower. Direct rows take about55 s versus prior24 s; compaction only
0.043 s, host expansion4.6–9.6 s. Setup/load/checks/I/O are separately reported.
Allocation completes28m59s. Retain resident reuse as the experiment default;
compact requires `--memory-mode compact`. Public WithCUDA defaults are unchanged.

[CPU-034](../experiments/cpu-output-memory/report.md) passes job73818. Bounded
columns replace full transpose scratch; the H prefix stays in place and checked
two-byte assembly uses disjoint arena halves before backwards expansion and
ownership transfer. Old words remain valid. Go reference/fast native APIs stay
available. Go/native suites, ownership/alias/error checks, ASAN/UBSAN and TSAN
pass; matched s14 and one full s16 word match retained hashes with complete
systematic/sampled-curve/packed-readback checks. Peak RAM is39.409 GiB versus
58.351 GiB fast (32.46% less), under48 GiB guard; modeled principal arrays39.345
GiB plus overhead. Full encode1289.759 s (21m29.8s), compaction/move23.586 s and
expansion6.064 s. Prior fast calls1247.173–1281.328 s are historical comparisons.
Matched s14 compact costs3.86–6.13% more encoding time. Independent full validation
806.8 s and setup25.5 s are separate costs; allocation completes49m22s. Normal
fast methods remain default when memory fits; compact is an explicit API/CLI mode.

Both jobs ran on gnode17 within admitted host/device/time budgets, with no guard
tripped. Exact source/evidence reconciliation passes (GPU338 / CPU344 source
identities); profiles, reusable preprocessing and packed outputs remain retained.
The completed controller hashes, metadata and file lengths verify full8 GiB
outputs without another bulk rehash during recorded-cache review. Accounting and
analysis are linked in the reports. No new validation job or default switch follows.

### Next decision: resident constants and s17 representation

Transform/output memory gates now have measured s16 implementations on both
backends. Continue the ordered plan with an allocation/representation audit and
streamed reusable spectrum/P design. GPU streaming saves device memory through
extra uploads and host backing; CPU streaming trades resident RAM for disk/page-
cache reads. Count staging, pinned memory, retained host plans, code load/setup,
repeated transfers and packing/I/O before proposing a concrete mode or job budget.
Preserve faster resident paths and require explicit opt-in when memory fits.

The compact uint16 schedule requires s<=16 and2L<=N. s17 has17-bit base values and
34-bit extension values, so it needs wider arithmetic/storage and a reviewed
assembly layout; current byte counts do not establish feasibility. Code/spectrum
residency, preprocessing support and complete peak accounting remain open. Present
those memory–time alternatives before selecting a material trade-off. GPU row
locality is a measured runtime opportunity, but does not replace the memory-first
s17 gate. Application/cryptographic/TRL5 acceptance remains separate.


### MEMORY-035: constants/representation audit and selected route

[Allocation-free audit and design](../experiments/streamed-constants/report.md)
with [reproducible results](../experiments/streamed-constants/results.json) closes
the source-width/peak-inventory part of this gate. The recommended next prototype
is fast-layout GPU streaming at s16: P uploads straight to the prefix arena and
spectrum chunks borrow dead transpose scratch before inverse. Model32.005 GiB
owned VRAM versus compact35.429 GiB, retaining fast rows and the host verified
plan (~39 GiB host RAM). Extra19.422 GiB code upload per encoding replaces
amortized resident setup; kernel/copy/synchronization latency remains unknown.
No pinned staging or new GPU tile allocation is assumed. Keep resident fast default.

Alternatives: one-arena/direct-row GPU streaming16.25 GiB arrays with the known
compact access penalty; file-backed CPU streaming20.172 GiB principal RAM plus
page-cache/runtime overhead, reading19.422 GiB per encode and initial integrity
verification. Both require explicit selection and measured end-to-end costs.
The user subsequently selected fast-layout streaming, implemented/validated by
GPU-036 below. The audit itself introduced no mode/default or budget increase.

At s17 (catalogue K still estimated), two fast GPU arrays alone256 GiB exceed
available140.40 GiB. A uint64 one-arena layout with uint32 disjoint halves can
potentially hold17-bit base output in128 GiB plus256 MiB staging, but requires
wider coordinate arithmetic/ABI/plans and correctness proofs. Retained host plan
+message+output projects220.841 GiB; file-backed code reduces principal host
storage to78.670 GiB plus cache/runtime overhead and142.420 GiB file backing.
Bounded CPU file-backed arrays project144.670 GiB, before additional ownership
copies if required. These exceed existing host/device admission or require new
streaming/width code; the numerical projection is not feasibility evidence.


### GPU-036: selected fast-layout streamed constants

The user selected fast-layout GPU streaming after reviewing trade-offs.
[Prototype/validation report](../experiments/gpu-streamed-constants/report.md)
implements the MEMORY-035 recommendation without changing ordinary CPU/GPU or
resident defaults. Optional streamed constructor installs verified shape; every
call synchronously borrows host P/spectrum. P uploads directly to prefix arena;
spectrum chunks borrow dead transpose scratch before inverse, retaining fast rows.
No device code array or new GPU tile is allocated; no C/CUDA host pointer persists.
Private Go adapter retains immutable host code until Close, unlike ordinary WithCUDA.

Job73940 now passes on gnode03/H200 NVL, exit0:0,16m22s allocation, within
admitted budgets and no guard tripped. Reconciliation resolves350 source identities
and sealed evidence/cache digests. All systematic positions, sampled independent
curves, retained CPU/GPU hashes, packed readbacks, small/boundary/ownership/range/
old-ABI tests,80 matrix regressions, build combinations, race/vet and memchecks
pass (0 errors/leaks). Existing profiles, verified caches and outputs are retained.

Measured s16 owned34,364,742,932 bytes (32.005 GiB), saving19.422 GiB /37.77%
versus resident reuse. Host RSS39.029 GiB is unchanged. Two full encodes63.487/
58.851 s include19.422 GiB code uploads in2.234/2.239 s through64 scratch tiles
of256 MiB. Rows remain fast at23.65–23.71 s, columns20.57–20.58 s; native totals
56.94/56.99 s differ from outer Go latency. Code-copy time is already in H2D.
Load/setup/independent checks/packing/I/O are separately reported; do not infer
encoding speed from the much shorter complete validation allocation.

Matched s14 resident5.1790/5.1759 s versus streamed5.5273/5.4666 s measures6.72%/
5.62% extra encoding time, two fixed-order samples. Full s16 comparison against
prior57.418/56.975 s on another allocation is historical, not a matched penalty.
First-use setup and repeated code uploads remain separate costs. Explicit
`--memory-mode streamed` is required; resident experiment and public defaults stay.

This completes the selected s16 streamed-constants gate. No new job starts from
this review. s17 fast two-array storage would still be256 GiB, beyond this GPU.
Next capacity work must combine streaming with a reviewed one-arena layout and
34-bit arithmetic/coordinates/artifacts, while accounting for host backing and
preprocessing. Keep material one-arena/file-backed trade-offs explicit before
selection; prior128.25 GiB one-arena s17 projection is not run admission or proven
feasibility. s17/application/cryptographic/TRL5 gates remain open.


### CPU-037: selected file-backed fast CPU constants

After publishing GPU-036 as ce29557, the user requests the CPU counterpart.
[File streaming prototype](../experiments/cpu-streamed-constants/report.md) verifies
existing LRSprobe1/2 bytes without retaining P/spectrum arrays. P reads directly
into prefix data; preserve H prefix and clear only the tail; read spectrum through
dead full transpose scratch between forward/inverse. Keep five fast transforms
and detach scratch as owning output. No extra tile or compact passes; resident/
compact CPU and GPU paths remain. Explicit `--memory-mode file`, default resident.

Model s16 principal8N+4K =35.422 GiB versus58.267 GiB fast resident; savings include
19.422 GiB constants and3.422 GiB saved H. Adds19.422 GiB logical reads per encode,
initial checksum/range pass and potentially cache/disk pressure. No timing benefit
assumed. Kernel file cache is additional memory, not hidden savings. Measure input/
output page residency without faulting it into process RSS or dropping caches;
record job-scoped cgroup memory if identifiable. Guard RSS48 GiB, RSS+tracked cache
96 GiB and job cgroup96 GiB when available, within existing128 GiB host request.
These are sampled observations and exclude untracked shared/node cache; do not
claim a total35 GiB cap. Code backing must remain immutable; stat/descriptor checks
are not protection against hostile concurrent writes. Existing formats stay intact.

Job73976 now passes on gnode17,32 CPUs/no GPU, exit0:0 in36m45s, with no guard
tripped. Reconciliation resolves358 exact sources/sealed evidence/cache digests.
Go/reference/ownership/error/header/file mutation/chunk tests, ASAN/UBSAN/TSAN,
normal/compact regression suites, Go-only build, race/vet, all systematic values,
16 sampled curves,1027 packed readbacks and full retained CPU/GPU hashes pass.
Profiles and backing caches remain retained. No kernel correction/default change.

Measured full s16 encoding1249.616 s (20m49.6s), read/range-check8.963 s for exact
19.422 GiB /78 logical chunks. Process RSS35.484 GiB versus prior58.351 GiB is
39.19% lower, but tracked file cache adds RAM. Sampled RSS+tracked resident files
peaks61.820 GiB; current charged batch-task subtree peaks47.917 GiB (recorded
memory.peak47.919 GiB). Shared cache/charging differ, residency sampling30 s can
miss peaks, and untracked node memory is excluded. No matched full-size resident
cache/cgroup comparison exists; do not claim a39% total-memory saving or35 GiB
total cap. Initial verification warms cache; cold-disk latency remains unmeasured.

Matched s14 resident58.853/58.667 s versus file59.377/59.357 s costs0.89%/1.18%
encoding time, two fixed-order cached samples. RSS3.595→2.214 GiB, inclusive tracked
RSS+cache20.385→19.009 GiB. Full s16 timing versus prior resident is historical.
Open/verify/native metadata37.181 s, independent checks178.754 s and pack55.768 s
remain visible separately; first encoding includes work-array allocation.

This closes the selected CPU file-streaming gate. Both CPU/GPU constant streaming
now have s16 evidence and remain explicit options alongside resident fast defaults.
Current fast CPU s17 two-array/message model~270.42 GiB before output-conversion
copies/cache/overhead still exceeds
256 GiB. Next capacity design must combine bounded/one-arena workspaces with
streamed code, wider arithmetic/coordinates/artifacts and cache-aware admission.
Prior144.67 GiB bounded CPU /128.25 GiB one-arena GPU projections are not total-
memory guarantees, implemented s17 support or admitted runs. No new job follows
this review; s17/application/cryptographic/TRL5 gates remain open.


### Remaining output/backed-transform gates: GPU then CPU

User authorizes finishing the two remaining options in this order:
1. GPU packed output to a bounded sink, without a full host canonical word.
2. GPU host-backed transforms with bounded VRAM and measured host/transfer costs.
3. CPU packed output to a bounded sink, preserving existing owning results.
4. CPU disk-backed transforms with bounded RAM and page-cache/I/O accounting.

GPU-038 starts gate1: direct packed canonical scatter into dead GPU scratch and
bounded packed download/writer. Retain fast workspace32.005 GiB in streamed-code
mode; target~16 GiB less host encoding memory at s16, plus writer buffer. Extra
atomic packing work and writer/hash costs unknown; transfer16→8 GiB. Output file
8 GiB and OS cache remain counted. Benchmark verifies mapped packed artifact in
a separate phase; no full uint32 verification array. Preserve existing APIs and
resident/owning defaults; use explicit stream-output CLI. No s17 claim or budget
increase. Gate2 needs dependency/transfer design and bounded checks before full
admission; do not confuse streamed constants with backed transforms. Dependent
work waits for the user's job-completion ping; no monitoring >30s jobs.

GPU-038 packed-output validation submitted as job74091 (one explicit H200 under
unchanged budgets). Gate1 remains pending; await completion ping without
monitoring, then continue GPU backed transforms and both CPU gates.


GPU-038 job74091 stopped at the s8 packed-output oracle before matched/full runs.
The new coordinate-zero reduction wrote a local chunk sum instead of the joined
shared sum; corrected to the established reduced value, with an independent
16-chunk constant-word regression in both layouts. Failed evidence/profiles are
retained. Retry uses a fresh run under unchanged budgets; GPU host-backed and
CPU routes remain dependent on a passed output gate. No measured memory/time
claim or completion of either remaining option follows from this failed run.

Corrected GPU-038 retry submitted as job74094 /packed-02 under the same budgets.
Await completion ping without monitoring before continuing dependent gates.


GPU-038 retry74094 passes gate1: two full packed s16 outputs match CPU/GPU hashes,
all systematic/curve/readback and parity/memcheck checks pass. Sink66.227/65.485 s
includes packing/write/hash, plus Sync; encoding RSS~23.01 GiB, mapped verifier
~31.01 GiB, GPU32.005 GiB unchanged. Matched s14 output pipelines27–29% shorter;
not an encoding-only speedup. Source/evidence reconciled, caches/profiles retained.
Proceed gate2 GPU host-backed transforms before CPU output/backed work. No default
or s17/budget change.


GPU-039 starts gate2 with an explicit host-backed **transform component**, before
full online integration. Two device tiles replace two full arrays, while the
full mutable uint32 array stays in host RAM plus one column staging tile.
GPU arithmetic is unchanged; Taylor corrections and row/column batches add
H2D/D2H transfers and launches. Runtime cost is unknown, never treated as free.
[Component experiment](../experiments/gpu-host-backed/report.md) checks Go/resident
agreement, inverse, batch tails, admission and exact allocation through2^22 values;
full s16 metadata only. First job budgets: oneH20016CPU,16 GiB host request,
12 GiB RSS,512 MiB device cap,20-minute controller/30-minute Slurm deadline.
These are lower than previous admitted limits. Existing resident paths remain
default; any slower backed encoding later needs explicit CLI selection and
memory/time admission. Gate2 remains incomplete until online integration and
validation; CPU packed output and CPU disk-backed transforms follow GPU work.

GPU-039 component job74121 /backed-01 submitted under those reduced budgets.
No monitoring: await the user completion ping; full online integration and both
CPU gates remain dependent on measured validation and subsequent admission.

GPU-039 job74121 passed/reconciled370 sources: full progressive correctness and
memcheck/regressions pass. At2^22 values, GPU32.76→1.26MiB, host data16MiB+256KiB
staging, but forward0.1278→7.7341s (60.5x slower), H2D1.105GiB/D2H0.160GiB.
GPU-040 parallelizes arithmetic within tiles under the same memory cap before
online integration; rows/columns dominate, not copies. Resident defaults remain
unchanged. Full encoding integration and CPU output/disk-backed gates remain.

GPU-040 implementation now uses one thread per butterfly pair within each
bounded tile. Joined quotient stages read the unchanged high half; inverse
corrections read immutable joined scratch. Row origins and column layout are
preserved; explicit row transposes disappear. Native sm90 compilation and an
independent sequential-vs-parallel GF(256) stage/layout check pass.
[Fixed-input GPU comparison](../experiments/gpu-backed-parallel/report.md) repeats
job74121 cases within identical budgets, requiring identical fixtures, tracked
memory, host staging and transfer counts; runtime validation remains pending.
Full online GPU integration and CPU gates remain subsequent work.

GPU-040 same-memory comparison submitted as job74125 /parallel-01. Await the
user completion ping without monitoring; GPU online integration and both CPU
gates remain subsequent, dependent work.

GPU-040 job74125 passes374-source reconciliation and full component validation.
AtN2^22,T65536, forward7.734→1.338s, row/column GPU6.496→0.03943s; memory and
transfers unchanged. Still10.47x the prior resident sample; tiny s4 cases20–43%
slower. GPU-041 now integrates the complete five-transform online schedule as
an explicit host-backed mode, retaining faster resident defaults. Host code,
N arena plus N canonical output and bounded staging count toward RAM; repeated
transfers/latency remain an explicit penalty. Small/cached s10 validation first,
not a fulls16/s17 admission. CPU packed/backed gates follow GPU work.

GPU-041 implements `WithCUDAHostBacked(options,tileElements)` plus optional
native online ABI, five transforms, tiled products and GPU orbit generation;
native CPU scatters canonical values/XOR parity. One hostN arena plus owning
N output, retained code/message and one native tile are accounted. s16 principal
host payload projects54.84GiB plus staging/runtime/cache; full runtime unknown.
This explicitly trades VRAM for RAM/PCIe/latency and leaves faster resident
defaults intact. Current backed output is owning; its packed sink still needs
integration. [Validation](../experiments/gpu-host-encoding/report.md) covers
s2/s3/s4 constant subcodes, retained s6/s8 and full s10 CPU/GPU workflows under
512MiB device/16GiB requested host budgets before any large-run choice.
Native/Go builds and controller self-check pass; runtime validation pending.

GPU-041 complete small/cached-s10 validation submitted as job74129 /online-01,
with explicit `--memory-mode host-backed` under the same reduced budgets.
Await completion ping without monitoring; full backed packed output, large-run
admission and both CPU gates remain subsequent work.

GPU-041 job74129 passes382-source/evidence reconciliation and complete s10
correctness. Backed GPU1.606–1.617s vs native CPU0.313–0.319s, device560400bytes,
host4MiB arena+4MiB owning output+256KiB staging+code/message, H2D456069120/
D2H128913404bytes. Keep explicit slower memory mode; no fulls16/s17 claim.
Proceed CPU-042 direct packed sink on fast file-streaming path, retaining twoN
transform arrays and reusing dead scratch for packed scatter and bounded writes.
No reduction in two-array transform peak claimed; output detachment/extra pack
pass avoided. CPU disk-backed transforms remain the next authorized gate.

CPU-042 adds direct canonical s-bit packing in dead scratch through native
atomic byte OR and joined coordinate-zero parity. File sink holds its lock
through bounded copied writes/hash, preserves earlier owning results and reuses
scratch. TwoN transform arrays remain counted; no peak transform-RAM reduction
is claimed. [Validation](../experiments/cpu-packed-output/report.md) uses
constant s2..8 and retained s6/s8 codes,1/16workers, writer/input/error recovery,
ASan and matched cached s10 owning+pack versus direct-sink pipelines. CPP/combined
local file-encoder/format checks and byte-boundary self-check pass; new runtime
checks pending. CPU disk-backed transforms follow, with I/O trade-offs explicit.

CPU-042 validation submitted as job74131 /packed-01 with explicit
`--output-mode packed`,16CPU/16GiB requested RAM, no GPU,30min Slurm and
20min controller. Await user completion ping without monitoring; CPU disk-backed
transforms follow this output gate.

CPU-042 job74131 passes389-source/cache/evidence reconciliation, small Go
oracles/ASan and matched s10 packed hashes. Direct pipeline0.32457/0.32736s vs
owning+pack0.33997/0.32778s, fixed-order two samples; no robust speedup claim.
TwoN arrays remain; atomic assembly costs6.06/7.11ms vs2.16ms owning.
CPU-043 starts remaining disk-backed transform component: two N-word files,
two T-word RAM tiles, repeated Taylor/file transpose/row I/O, filesystem cache
extra. s16 requires32GiB backing storage, beyond current16GiB new-cache budget;
not admitted. Small/progressive N<=2^22 prototype fits current budgets and
precedes full online integration. Runtime/physical-disk cost unknown.

CPU-043 native component now uses exclusive pre-sized little-endian data/scratch
files, joined bounded Taylor passes, blocked file transposes and native affine
vector batches. No full array/mmap is allocated by execution; twoT private
vectors plus64KiB serialization, with filesystem cache/metadata/runtime extra.
[Component validation](../experiments/cpu-disk-transform/report.md) compares
retained Go/GPU oracles and current resident CPU through2^22, offsets/canaries,
FD/length/tile/range/error checks and ASan. Record logical I/O/time, filesystem,
end-state retained-file cache and job-scoped cgroup when available. CPP/combined
local builds and bounded transpose/dependency self-check pass; native runtime
and full disk-backed online encoding remain pending. Fulls16 storage32GiB not
admitted within current16GiB new-cache cap. No default path changes.

CPU-043 component validation submitted as job74134 /disk-01, explicit
`--memory-mode disk`,16CPU/noGPU/16GiB RAM request/16GiB new-cache guard,
30min Slurm. Await user completion ping without monitoring; full disk-backed
CPU online integration remains subsequent work.

CPU-04374134 passes395-source/evidence reconciliation and full component/ASan
checks. N2^22,T65536:589824bytes private buffers vs32MiB resident arrays,
4.721s forward/2.600s inverse vsresident0.631s. Backing32MiB; logicalI/O
1.137GiB read/0.191GiB write per transform. NFS/cache/writeback effects vary;
end mincore0pages does not prove no cache during execution. Full job cgroup
peak1.419GiB includes build/validation. CPU-044 integrates five-transform disk
online encoding with one owning N-word host output plus bounded tiles; code
stays streamed. s16 host payload projects19.42GiB+tiles/metadata/runtime,
32GiB backing+code/output storage/cache extra. Fulls16 not admitted. Small/full
s10 validation first, faster resident/file-fast defaults unchanged.

CPU-044 now implements `OpenScalingCPPDisk` and one native online call containing
five file transforms, bounded file products/code copies/zeroing and canonical
assembly into a new owning N-word host result. Code stays streamed, mutable
work files are created exclusively and retained after Close. Actual nested
buffer liveness is counted conservatively as4T words+128KiB, not component2T.
[Validation](../experiments/cpu-disk-encoding/report.md) covers CPU042 small
Go fixtures/constant words,25/33-byte code headers, errors/recovery/ownership,
ASan and complete cached s10 file-fast/disk workflows. Existing CPP/combined
local checks and resource/offset self-check pass; runtime pending. Fulls16
backing32GiB not admitted. Full output still4N private bytes; a disk packed
sink is a separate combination, with no hidden large allocation/fallback.

CPU-044 full small/cached-s10 validation submitted as job74141 /online-01,
explicit `--memory-mode disk`, same16CPU/noGPU/16GiB requested RAM and16GiB
new-cache cap,30min Slurm. Await user completion ping without monitoring;
fulls16 disk storage remains unadmitted and optional output-mode combinations
remain separate from completing the current disk online route.


### Current memory-route status

A [bounded table-gap batch](../experiments/encoding-table/report.md) covers all22
originally unmeasured, implemented encoding cells at s10/14/15/16 before the next
design selection. It reuses code caches; no s≥17 support is implied.

The performance objective now extends through s18, with s19–20 exploratory.
Use the [maintained results and capacity tables](encoding-status.md) for measured
results and multi-GPU projections; the immediate work below remains the gate to
s17 before progressing to s18.

CPU-044 job74141 passes401 source identities and sealed cache/evidence checks,
small Go oracles/ASan and two full s10 encodings per mode with systematic,
independent-curve, canonical/packed hash and readback agreement. Disk encoder
first/reused2.860/1.136s versus file-fast0.332/0.311s, owning output4MiB plus
1.125MiB conservative native buffers and8MiB backing files. All comparisons are
profiled fixed-order samples on NFS; file cache/runtime remain extra.

The first four implementation investigations are now done for both backends.
Validation sizes differ; this is not completion of combined s17 admission.

| Approach | GPU evidence | CPU evidence |
|---|---|---|
| Bounded transform scratch/compact layout | [Compact path](../experiments/gpu-output-memory/report.md), full s16 | [Compact path](../experiments/cpu-output-memory/report.md), full s16 |
| Stream preprocessing constants | [Fast streamed code](../experiments/gpu-streamed-constants/report.md), full s16 | [File-streamed code](../experiments/cpu-streamed-constants/report.md), full s16 |
| Direct packed output without owning canonical word | [Bounded GPU sink](../experiments/gpu-packed-output/report.md), full s16 | [File CPU sink](../experiments/cpu-packed-output/report.md), small/ASan and matched s10 |
| Back transforms with host RAM/disk | [Host-backed GPU](../experiments/gpu-host-encoding/report.md), small/ASan and full s10; components through2^22 | [Disk-backed CPU](../experiments/cpu-disk-encoding/report.md), small/ASan and full s10; components through2^22 |

Owning API defaults, verified CPU/GPU paths and canonical representation remain.
Slower modes stay explicit API/CLI choices; fitting faster paths remain default.
Backed encoders currently return owning arrays: their packed sinks are not
implemented. Compact transforms/streamed constants/output do not all combine in
one path yet. Existing stream/sink modes must not be mistaken for full backing
or a total-RAM bound; cache/driver/runtime/retained outputs are additional.

**Next: execute the [s17 CPU-first/multi-GPU plan](#current-execution-plan-for-s17).**
Start width/scalar/native work independently of the running table-gap jobs.
Combinations and full lifetimes remain part of mode selection, with host/device/
storage/runtime costs explicit. Fulls16 CPU disk is now admitted under the user's
64GiB retained-storage guard and submitted with extended deadlines; see the
[retry report](../experiments/encoding-table/report.md#extended-cpu-retries).
Current full execution still supports s≤16. Wider scalar reference work alone
is not s17 encoding support, preprocessing admission or TRL5 acceptance.
