# Distributed preprocessing: GPU-resident working storage

Next implementation step after [small distributed products passed78321](../experiments/distributed-gpu-product/report.md).
The first [resident-storage/transpose prototype](../experiments/distributed-gpu-storage/report.md)
passed bounded validation78392 (208 normal +208 sanitized rank/pass results,
zero CUDA errors/leaks). Resident Taylor/FFT/product
integration, fullsize costs and resource admission remain separate. We remain
on step1 (distributed large products/FFTs).

The existing host-shard implementation is a correctness reference. Its conservative
product model includes six full host shards, while GPU arithmetic uses bounded
tiles. At s18 this host model is3072 GiB aggregate before overhead. Increasing
that reference workload does not resolve the storage design needed for fast large
preprocessing. Prototype resident storage first, retaining the same arithmetic,
canonical values, ownership schedule and independent references.

## Concrete implementation sequence

1. Separate rank transpose and Taylor source-window storage access from their
   shared communication schedule. Keep the current host adapter as a regression
   reference; add a GPU adapter. Both must use the same headers/proofs, global
   coordinates, pass order and barriers. Avoid a second FFT implementation.
2. Allocate three reusable owned GPU shards: left, right and scratch. Keep
   coefficients row-major and transformed vectors interleaved where required by
   the existing coalesced vector kernel. GPU pack/scatter must translate these
   views directly; avoid full host copies or an unmeasured strided-kernel fallback.
3. Forward left using scratch, then forward right using the same scratch while
   preserving left's spectrum. Multiply left in place, then inverse left using
   scratch. Taylor windows borrow bounded device/host work buffers; no complete
   host shard is required to fetch a source or update a target.
4. Upload operands and read final coefficients in bounded chunks. Intermediate
   FFT/Taylor communication stays in memory. Keep checkpoint/final writes buffered
   with controlled writer concurrency. No new intermediate disk matrix is proposed.
5. Validate both slot widths, row/interleaved layouts, odd dimensions, owner/tile
   tails, phase ordering, alias rejection and source preservation against the
   existing host path and independent schoolbook/native references. Repeat CUDA
   checks on small cases; reuse measured sanitizer costs. Compile and check remote
   paths before GPU execution. Retain phase progress, profiles and checksums.
6. Compare matched end-to-end products, including reusable allocation/setup,
   packing, communication, verification and output. Only then scale sizes/card
   counts or integrate a resident executor into roots/Newton/generator work.
   Genuine P/G identities, s18 field admission and the complete pipeline remain
   separate gates. Admission is required before increasing experiment resources.

## Memory and time trade-offs

N is the padded domain; P is the GPU/rank count; w is the internal field slot size
(8 bytes at s17/s18). These are payload models, not measured peaks.

| Resource/cost | Validated host reference | Proposed resident backend |
|---|---|---|
| Main GPU arrays | Bounded arithmetic tiles | Three shards, approximately3Nw/P per GPU |
| Main host arrays | Up to six shards, approximately6Nw/P per rank, including caller-retained inputs | Bounded transport/window buffers; caller inputs, final output and other preprocessing state are additional |
| Host/device traffic | Repeated vector/Taylor staging, including source snapshot readback | Initial operand upload, final result readback and required bounded inter-rank/window transfers |
| Work added | CPU packing and repeated transfers | GPU packing/scattering and larger reusable allocations |
| Runtime | Small products measured; fullsize unknown | Unknown until matched measurement; reduced staging is an expectation, not a measured speedup |
| Retained storage | Existing result/cache formats | Same formats/bytes; no new full intermediate disk matrix |

For s18, N=2^36 and w=8: three GPU arrays total1536 GiB. An illustrative16-H200
allocation gives96 GiB/card before workspace, metadata and runtime/context. This
is below approximately140.4 GiB/card hardware capacity with substantial payload
headroom; it is not proof of complete workflow feasibility or a fixed GPU count.
More/fewer cards should be compared for total latency when supported/admitted.

As a design target, two1-Miword host work buffers consume16 MiB/rank at w=8
(256 MiB across16 ranks), plus MPI/runtime and caller/I/O buffers. This excludes
input/output materialization, discovery/support data, root/Newton state, retained
spectra and filesystem cache. A complete base-field product contains up to256 GiB
of32-bit coefficients; streaming that result avoids requiring it all in RAM but
does not remove required I/O/storage. No full-workflow RAM or storage total follows
from these isolated models. Verify actual buffer lifetimes before a full run.

The immediate prototype should fit the existing two-node/two-H200 bounded budget
at small sizes. The projected16-GPU arrangement is not authorized by that budget.
Retain faster validated defaults when they fit; do not automatically fall back to
a slower host/disk route or assume new allocation/communication costs are free.

## First storage gate (SCALING-152)

The host and CUDA adapters now share the rank transpose schedule; descriptors,
proofs, ownership and round ordering are unchanged. Three reusable device shards
and a127-word device transfer tile are counted by the executor cap. GPU kernels
copy local intersections, pack outgoing rectangles and scatter incoming rectangles
directly between either physical layout. The control plane allocates only one
bounded host transport tile; diagnostic test inputs/oracles are additional.

The gate covers both slot widths, all four source/target layout combinations,
uneven/empty ownership, inverse transposes, alias/extent rejection, source/third-slot
preservation and stable retained allocation. It deliberately moves arbitrary slot
bits without arithmetic. This validates storage; it does not validate a resident
FFT, product or canonical field values. Resident vectors passed78394 and complete resident FFT passed78396. Pointwise/
products passed78399 with the same three arrays, operation schedule and
arithmetic. Tiny resident products are slower than host under the current
127-word/diagnostic boundary. Matched sizes/transfer batching passed78432:resident still slower atN16384.
Direct local Taylor GPU access passed78460 and wins matchedN16384 products
without increasing resident payload. Larger structured products with exact linear
oracles passed78469 through262144 values (resident1.554 s vs host-stored GPU4.911 s).
Local GPU roots plus a distributed final product passed78508 through65532 roots
(largest profiled net product0.603 s). Bounded operand routing passed78662 with no full peer polynomial and half
the largest-case input traffic. One-sample timings do not establish a latency
advantage. Complete root/Newton execution, full support ownership and distributed
discovery/checkpoint/publication integration remain. No large calibration or resource increase is implied.
