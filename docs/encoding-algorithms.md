# Scalable Lift-RS encoding: algorithm review

Research checkpoint, 2026-10-27. This reviews exact systematic encoding for the
catalogue's curve-lifted codes (`m=2`, `s=8..20`), not ordinary line-lifted or
linearized Reed-Solomon codes. The [bounded implementation](../encoding/liftrs/README.md#paper-and-cyclic-encoders)
and [CPU comparison](../experiments/encoder-comparison/report.md) now validate the
main paper/cyclic paths; large-instance feasibility remains unverified. The [encoder contract](encoder-contract.md) remains authoritative.

## 1. The paper's algorithm

[Section 4.2.4, Lemma 4.9, Theorem 4.3 and Claim 4.2](../paper.pdf#page=31)
provide a quasi-linear systematic encoder. Write `q=2^s`, `N=q^m`, and let `K`
be the exact dimension. Choose an extension field `E=GF(q^m)`, a primitive
nonzero element `omega`, and an `GF(q)`-linear isomorphism `phi:E -> GF(q)^m`.
The paper gives an explicit trace construction of `phi`.

The information set is `phi(omega), ..., phi(omega^K)`. For zero-based message
index `j`, compute `phi(omega^(j+1))`; repeated squaring needs `O(log(j+1))`
extension-field multiplications, followed by a fixed linear map. No K-entry
coordinate table or discrete logarithm is needed. See also
[Guo-Kopparty, Appendix A, Theorem A.1](https://www.math.utoronto.ca/swastik/lifts-list-decoding.pdf).

**Reusable preprocessing:** enumerate good multivariate monomials and count K;
choose field representations and phi; discover the univariate support D by
interpolating random codewords and taking unions of supports. The paper bounds
missing-support probability by `q^(m-T)` after T independent samples. Under exact
arithmetic, every observed exponent belongs to D and `|D|=K`, so reaching K
exponents certifies completeness. A retry limit must fail explicitly if this
condition is not reached. Exact K, not a Monte Carlo capacity estimate, is needed.

**Per database:** solve

`sum_(i in D) a_i * omega^(i*j) = message[j-1], j=1..K`,

then evaluate `f(z)=sum_(i in D) a_i*z^i` on E and output
`c(x)=f(phi^-1(x))`. The paper gives encoding time `O~(m*N)` and preprocessing
`O~(N + m^(s+1)*q*t)`, with the sampling work described separately in Claim 4.2.
These are algebraic asymptotic bounds, not RAM, disk or GPU performance estimates.

The solve is more structured than a generic sparse interpolation problem:
`beta_i=omega^i` are distinct, and the matrix entries are `beta_i^j`. This is a
transposed ordinary Vandermonde matrix with a diagonal scaling. Fast transposed
multipoint techniques apply; ordinary interpolation of a degree-`K-1` polynomial
at the message coordinates would generally encode the wrong space.
[Bostan-Schost](https://cs.uwaterloo.ca/~eschost/publications/BostanSchost.pdf)
and the paper's BS04/BJS07 references supply relevant polynomial/structured-system
methods. General point sets require fast polynomial arithmetic and product trees;
D need not be consecutive or a geometric progression of exponents.

Preserve the subfield constraint: codeword values are in GF(q), although the
univariate polynomial has coefficients in E. Equivalently use the scalar extension
of the code when reasoning about individual univariate monomials. Validate the
Frobenius constraints rather than treating arbitrary E coefficients as messages.

## 2. The cyclic formulation

This is an implementation derivation from the paper's invariance and information
set, using classical cyclic-code structure; it is not a new literature result or a
claim of target-scale speedup. Extended cyclic and affine-invariant codes are related in
[Abdukhalikov (2005), Sections 2-3](https://doi.org/10.1016/j.jpaa.2004.02.016).

Put `n=N-1` and order nonzero coordinates as
`y_r=c(phi(omega^(r+1))), r=0..n-1`. Multiplication by omega is an invertible
GF(q)-linear coordinate transformation, so it preserves our curve constraints.
Consequently the punctured code is cyclic, still of dimension K. Puncturing is
injective because the information set is among these nonzero coordinates.

Let G be its cyclic generator polynomial, normalized by `G(0)=1`, of degree
`n-K`. It divides `1-X^n`. Put `P=(1-X^n)/G`, so `deg(P)=K` and `P(0)=1`.
Both polynomials are over GF(q). A systematic encoder with the paper's *same*
information set is:

```text
Encode(message[0:K], preprocessing):
    M = sum_j message[j] * X^j
    H = (P * M) mod X^K
    Y = G * H                       # degree < n; coefficients are y_0,...,y_(n-1)
    write y_r at coordinate phi(omega^(r+1)) for every r
    write sum_r y_r at coordinate phi(0)  # characteristic two
```

**Why it works:** `Y` is a multiple of G of degree below n, hence is a punctured
codeword. Since `G*P=1-X^n`, its first K coefficients equal M modulo `X^K`.
The extension symbol is the sum of the nonzero symbols: each degree-at-most-t
curve restriction has reduced degree at most `q-2`, so every affine line sums
to zero; partition the grid into parallel lines to obtain zero total sum.
Thus the result is exactly the unique full codeword with the prescribed values.

With fast polynomial multiplication, encoding costs `O(M_q(n)+M_q(K)+N)` base-field
operations, plus coordinate mapping, where `M_q(d)` is polynomial multiplication
cost over GF(q). This is quasi-linear with suitable multiplication. It avoids
per-message Vandermonde solving and full extension-field evaluation. It does not
prove better wall-clock performance or an asymptotic improvement over the paper.
The main implementation problem becomes fast multiplication with manageable I/O.

There are two preprocessing routes to compare:

- **Reuse the paper's D:** let `beta_i=omega^i`; form
  `P(X)=product_(i in D)(1-beta_i*X)`, then obtain G by exact division.
  Frobenius closure ensures P is over GF(q); check that explicitly. The reciprocal
  orientation of these factors matters.
- **Use random codewords directly:** sample uniform codewords using the good
  multivariate monomials, evaluate on the grid, reorder their nonzero values into
  coefficient polynomials Y, and accumulate `gcd(1-X^n, Y_1, ..., Y_T)`.
  Each gcd is a multiple of the true generator. Once its degree is `n-K`, it
  determines G exactly, up to normalization. This avoids discovering D and doing
  univariate extension-field interpolation. For each remaining irreducible factor,
  a uniform random codeword vanishes in that component with probability at most
  `1/q`; the union bound is at most `n*q^-T`. Cap retries and return an error
  rather than accepting a larger-degree gcd. This reasoning requires the exact
  code space and independent uniform samples, not estimated K or sparse samples.

The D-based route is now implemented for m=2, s=2..9 and checked against the
paper path and dense oracle. The random-codeword gcd route is not implemented.
Tests verify coordinate orientation, zero-coordinate extension and equality.
The [current dense oracle](../encoding/liftrs/reference.go) chooses different
systematic positions; compare using a message extracted at its information set,
not by feeding both encoders the same raw message and demanding identical words.

## 3. Algorithms and implementations to reuse or compare

| Source | Relevant contribution | Fit and limitation |
|---|---|---|
| [Coxon, fast transforms (2018 preprint; 2021 journal)](https://arxiv.org/abs/1807.07785) | Additive evaluation, interpolation and conversion among monomial, Newton and novel bases; general subspaces and subfield optimizations | General baseline covering our non-power-of-two extension degrees; include conversion cost |
| [Samanta-Badakhshan-Gong, August 2026](https://arxiv.org/html/2608.20855v1) | Matrix-style additive-FFT decomposition and cache-local recursive monomial evaluation | Promising evaluation kernel, not an entire systematic encoder; general-basis Algorithm 1 versus specialized Cantor Algorithms 3-4 |
| [Author additive-FFT repository](https://github.com/mtbadakhshan/additive-fft) | Paper reports C Dyadic AFFT code in `C/dyadic`; repository also describes C++/Sage implementations | Published benchmark uses GF(2^128), single-threaded CPUs. Not reproduced here; dyadic source subdirectory could not be fetched in this review, so no source/license compatibility claim |
| [klauspost/reedsolomon](https://github.com/klauspost/reedsolomon) and [Leopard](https://github.com/catid/leopard) | Optimized binary-field arithmetic and additive-FFT RS kernels | GF(2^8)/GF(2^16) kernels are useful references, not drop-in encoders for the curve-lifted code or all field degrees; Go dependency already listed in [go.mod](../go.mod) |
| [libfqfft geometric domain](https://github.com/scipr-lab/libfqfft/blob/master/libfqfft/evaluation_domain/domains/geometric_sequence_domain.tcc) | Actual C++ geometric interpolation/evaluation code | Useful implementation reference, but its contiguous-degree interface is not our transposed arbitrary-node solve; polynomial multiplication backend also needs binary-field compatibility checking |
| [NTL GF2EX](https://libntl.org/doc/GF2EX.cpp.html) and [Sage cyclic codes](https://doc.sagemath.org/html/en/reference/coding/sage/coding/cyclic_code.html) | Binary extension polynomial arithmetic and cyclic-code reference operations | Useful validation tools, not evidence of target-scale throughput or bounded memory |
| [Coxon, systematic multiplicity encoding](https://arxiv.org/abs/1704.07083) | Fast multivariate interpolation/evaluation for RM and multiplicity codes | Different polynomial spaces; their information sets cannot simply replace the lifted-code information set |

The 2026 paper reports beating LCH in 37/42 configurations, but only for its tested
GF(2^128) CPU evaluation kernels. This is not a speedup claim for our encoder.
A full Cantor special basis is not available for every required field: for `m=2`
our natural extension is GF(2^(2s)), and only s=8 and s=16 in this range give
power-of-two full extension degree. General-basis methods remain necessary.
Using GF(2^128) does not solve this uniformly: embedding GF(2^(2s)) requires
`2s` to divide 128. Polynomial-basis public field elements remain unchanged by
any explicitly converted internal transform basis.

No faster complete encoder specifically for these curve-lifted codes was identified
in the sources reviewed. This is a scoped search result, not a claim of optimality.
The strongest opportunities found concern cyclic reformulation, base-field
arithmetic, transform constants, batching, and memory/I/O scheduling.

## 4. Memory, storage order and CPU/GPU implications

For s=20, N=2^40. One packed base-field vector takes 2.5 TiB; one uint32-per-element
vector takes 4 TiB. An extension-field vector with uint64 entries takes 8 TiB.
These are arithmetic size calculations for one vector, not measured resource usage;
product trees, transforms and simultaneous databases need additional space.
The paper's time bound does not supply an out-of-core implementation.

Use block scheduling and external storage where needed; do not promise one-pass
streaming for transforms or polynomial multiplication. G/P preprocessing can also
be large and disk-backed. Reuse it across equal-code databases and batch only as
many columns as fit memory. Keep large-data artifact storage separate from the
small client metadata (fields, phi, omega, K and coordinate conventions).

Keep physical addressing in the fixed coordinate grid: after mapping by phi,
flatten the point using the documented coordinate order, then apply the secret
permutation. Storing data in omega-power order without another mapping would
require converting arbitrary query points to exponents, potentially a discrete-log
problem. Internal multiplicative ordering is useful for encoding but should not
impose that cost on retrieval. The layout must also support the planned
[grouped position reads](protocol-io.md#storage-layout-optimization).

CPU and GPU kernels can implement the same exact field operations and block
schedule. GPU acceleration is plausible for independent transforms and polynomial
multiplication, but GPU-resident capacity and transfer costs matter. The reviewed
2026 timings are CPU measurements, not evidence of GPU performance. Preserve a
portable implementation; select a native acceleration boundary only after profiling.

## 5. Checkpoint and remaining comparison

The [initial table-backed experiment](../experiments/encoder-comparison/report.md)
verified exact equality, preserved the paper's information set and measured separate
setup/load/encoding costs through N=65,536. Both the direct Karatsuba cyclic path
and the cached-FFT multiplication variant are available; the latter uses extension-
field arithmetic internally and serializes the fixed generator spectrum.

Next, expand the comparison and memory/I/O design before selecting a large-instance
backend. Reusable public artifacts already separate code-only constants from each
database. The current in-memory sizes, three-repeat CPU timings and small-file
output checks did not establish larger-code, external-storage or GPU feasibility.
The later [table-free s=9 checkpoint](../experiments/tablefree-encoding/report.md)
now extends exact encoding evidence through N=262,144; s>=10 and external-memory/
GPU feasibility remain unverified.

## 6. Five-step implementation plan

Prioritize the cached-FFT cyclic encoder for further development, retaining the
paper encoder as a correctness and performance baseline. Keep the same information
set, field conventions and codeword outputs. The following steps are planned,
not implemented support for the larger configurations. Complete the CPU and
moderate-instance checkpoints first; external-memory and GPU work are later stages,
subject to the [roadmap's scope and TRL gates](../ROADMAP.md).

### Step 1: Remove full-extension-field arithmetic tables

Implement portable polynomial-basis arithmetic for every `GF(2^s)`, `s=8..20`,
using uint32 base-field elements. Represent the quadratic extension by two
base-field elements instead of indexing log/exp tables of size `2^(2s)`.
Use carryless multiplication and reduction; profile before adding CPU-specific
kernels. Field arithmetic support alone does not enable encoding a full instance.

The current cached-FFT encoder accelerates only `G*H`; computing
`H=(P*M) mod X^K` still uses Karatsuba. Its full encoding bound is
`O(M(K)+N log² N)`, with `M(K)=O(K^log2(3))` today. The paper solver also uses
Karatsuba in its product-tree arithmetic. Implement and verify fast truncated
multiplication for the message-dependent prefix product, and fast general
multiplication/division for reusable preprocessing and the paper baseline.
Measure these operations separately before extrapolating whole-encoder runtime.
Changing field arithmetic alone does not make either complete pipeline quasi-linear.

**Acceptance:** exhaustive small-field and randomized larger-field checks,
unchanged encoder outputs, and arithmetic storage that does not grow with the
extension-field cardinality. The first encoding checkpoint is `s=9`, using its
catalogue curve degree, after arithmetic passes these checks.

### Step 2: Define compact, reusable public preprocessing artifacts

Separate the artifacts by responsibility:

| Layer | Contents | Distribution |
|---|---|---|
| Code description | Field polynomials, extension construction, coordinate map, primitive generator, exact dimension and algorithm version | Small public artifact |
| Reusable auxiliary data | Transform constants, compact support descriptions and selected polynomial data | Publish when reasonably sized |
| Large acceleration caches | Full polynomials, product trees or generator spectra | Optional local files, generated once and reused |

Use a configurable publication-size budget; **64 MiB per configuration is a
proposed initial default**, not an existing limit or requirement. Exceeding the
budget changes what is cached or distributed, rather than making the code
ineligible for encoding. Distinguish publication size from local RAM/disk budgets.

Exploit the paper's weight-based monomial classification for exact dimension and
support construction. Investigate compact support representations without assuming
that the full degree set compresses well. Catalogue capacity estimates cannot
replace exact dimensions in preprocessing. Keep all code-only work outside
per-database encoding and amortize it across databases sharing that code.

**Acceptance:** versioned artifacts with reproducible generation commands,
checksums, code specifications and verification evidence. Before publishing,
rebuild from the recorded source/specification, check exact dimension and field/
coordinate conventions, compare complete outputs against the independent dense
oracle on feasible fixtures, and cross-check paper/cyclic results and systematic
symbols on the publication configuration. Record hashes and these checks beside
the artifact. A loader's shape/range validation is not this mathematical verification;
large configurations must explicitly state any limits of independent checking. Report generation,
method preparation and loading costs separately from encoding; document which
auxiliary data are loaded, reconstructed or optionally cached.

### Step 3: Scale the in-memory CPU encoder through s=9..12

Advance one catalogue configuration at a time. Replace JSON numerical arrays
with a documented binary artifact format, reuse scratch buffers and limit
simultaneous allocations. Compare caching the generator spectrum with recomputing
it, recording the change in artifact size, setup and per-database cost explicitly.
Compare with the paper path where feasible, without conflating different
preprocessing boundaries.

**Acceptance:** exact agreement of codewords and systematic positions, verified
capacities, and reproducible measurements of discovery, preparation, artifact
loading, encoding, output I/O, peak memory and artifact size. These measurements
choose the next backend; do not infer target-scale feasibility from operation
counts alone. Measure each preparation method in its own process and report
encoding workspace separately from artifact loading, output buffering and other
process memory. Record exact CPU model, RAM, OS/toolchain, storage device/filesystem
and cache/sync policy. Use the intended packed output layout and, once available,
grouped-position storage; retain unpacked small-file checks only as labelled
correctness fixtures. The current experiment's process RSS and uint32 sink do not
satisfy these future resource/layout acceptance checks.

### Step 4: Implement encoding with bounded RAM

**Decision gate:** after Step 3, use its resource measurements to decide with the
user whether bounded-RAM work is necessary for the selected TRL 5 use case or
whether to pursue the larger TRL 6 scaling effort. This plan does not itself
activate the deferred work in the roadmap.

Introduce a configurable memory budget and disk-backed working arrays. Specify
block schedules for transforms, polynomial multiplication and coordinate
reordering, including preprocessing artifacts. Memory-mapping existing arrays
alone is not an external-memory algorithm; count temporary disk space and bytes
read/written. Preserve the separation from permutation/masking and the eventual
grouped storage layout, so later integration does not require a full in-memory
codeword.

**Acceptance:** a case exceeding the RAM budget completes correctly with measured
peak memory and disk traffic. First force external-memory paths using an
artificially small budget on manageable instances and compare their outputs with
the in-memory backend. Large-instance runs follow only after that check.

### Step 5: Add optional GPU acceleration where justified

**Decision gate:** after CPU profiling, present expected gains and deployment/
maintenance costs to the user before choosing a GPU backend and expanding the
current optimization scope. No GPU vendor or native dependency is selected here.

Keep code descriptions, artifacts and outputs compatible with the CPU path.
Accelerate arithmetic batches and transform blocks, transferring sufficiently
large blocks to amortize overhead. Preserve the portable CPU implementation;
select the GPU boundary based on profiling rather than assuming that every stage
benefits. GPU processing is confined to preprocessing/encoding, not required for
client retrieval or storage reads.

**Acceptance:** identical CPU/GPU outputs and an end-to-end improvement including
transfers, setup and memory requirements, with preprocessing reuse stated. Kernel
speed alone is insufficient. Do not commit to terabyte-scale runtime or publish
large acceleration caches before the earlier checkpoints establish their costs.

### Step 1 checkpoint: table-free arithmetic and s=9

[The implementation and experiment](../experiments/tablefree-encoding/report.md)
now provide portable base-field and two-component extension arithmetic for all
s=2..20 with fixed-size arithmetic state. The encoder adapter supports m=2,s<=9;
its transforms and polynomial scratch remain in memory. Existing s<=8 artifacts
and complete output hashes are preserved, and s=9,t=7 has exact K=17200 with
paper/cyclic agreement and sampled curve constraints. Table-free arithmetic is
slower than the former tables on the measured small instance. This completes the
arithmetic/s=9 part of Step 1, not fast truncated multiplication, compact binary
artifacts, bounded-RAM encoding or GPU acceleration.
