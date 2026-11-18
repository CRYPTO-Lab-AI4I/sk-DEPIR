# s18 preprocessing: four-GPU hybrid versus sixteen-GPU distributed

**Decision analysis, not a run or resource admission.** Both complete routes
still require implementation. Recommend the **sixteen-GPU/four-node distributed
route for the speed goal**, distributing host/root work as well as device arrays.
Do not carry the current host-backed transform into a full run unchanged.
Four GPUs remain a possible lower-resource route after host-lifetime and tiled
kernel improvements, with an explicit runtime trade-off.

## Reproduce and inspect

[Model and self-check](model.py), [tests](test_model.py), [generated results](model.json).
Run from the repository root:

```sh
python3 -B experiments/s18-preprocessing-routes/test_model.py
python3 -B experiments/s18-preprocessing-routes/model.py > /tmp/s18-routes.json
cmp experiments/s18-preprocessing-routes/model.json /tmp/s18-routes.json
GOTOOLCHAIN=local GOCACHE="$PWD/.cache/liftrs/preprocessing-workflow-build-cache" \
  .cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go tool pprof \
  -top -nodecount=12 \
  .cache/liftrs/gpu-s17-full-preprocessing/preprocess-75832 \
  .cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5/cpu-4145752060.pprof
```

Prepared from revision `0aac5a9`; exact model/input/arithmetic source SHA256
identities are in model.json and the final analysis source is tracked by Git.
The model imports existing admission and split-product models and reads the
reconciled full-s17 phases and bounded-component evidence; no giant cache reread
or new GPU experiment is needed. [Saved profile summary](s17-resume-cpu-top.txt)
uses the retained matching resumed binary/profile. All caches remain intact.
The first pprof command attempt lacked a writable default Go cache; the explicit
repository GOCACHE command succeeded. No profiling job was rerun.

Measured evidence comes from HPC, not laptop:

- [Full s17 preprocessing75832](../gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088):
  4 H200 GPUs, 32 active CPU workers, one exclusive node, 864 GiB requested RAM;
  resumed worker9 h58 min38 s, peak sampled process RSS348.393 GiB. Earlier
  discovery21 min38 s used two samples. Root construction8 h33 min57 s dominates.
- [Bounded components76756](../preprocessing-tiles/report.md#completed-component-validation-scaling-104):
  4 H200 GPUs, 4 CPU cores, 8 GiB requested RAM. At2^22 points, host-backed forward
  latency is about5.53 times resident and split-product latency about5.60 times
  resident on four cards. These profiled samples **do not establish full-size
  penalties** or end-to-end scaling.
- [Exact s18 count76647](../s18-preprocessing-admission/report.md):
  K=16300262095, N=68719476736, m=2,t=5. This is measured dimension, not prepared
  support or code. [Streamed complete export76710](../preprocessing-export/report.md)
  passes bounded correctness; full-size host/RSS/I/O savings remain unmeasured.

## Complete workflow comparison

All s18 numbers below are **source-derived principal payload projections**,
not measured RAM peaks. Assume four H200 GPUs/node, about907 GiB scheduler RAM
capacity/node, a hypothetical112 GiB owned cap/card, with metadata/context and
runtime headroom additional. Neither16 GPUs nor these full-size memory/storage
budgets are admitted by the earlier bounded experiment.

| Stage | Four GPUs, one HPC node: hybrid design | Sixteen GPUs, four HPC nodes: distributed design |
|---|---|---|
| Exact dimension / monomial expansion | Count already measured; full good-index expansion still required. Discovery model includes both good and found index lists. | Same dimension; distribute expansion/index ownership rather than replicate full lists. Count speed is not workflow speed. |
| Discovery: sampling, base-grid, interpolation, union and support verification | About770.893 GiB principal host arrays. Current resident transform/grid needs1024 GiB aggregate GPU arrays and cannot fit four112 GiB caps. Need tiled **base-grid** as well as tiled extension interpolation; the existing HostTiled prototype alone is insufficient. | Balanced host sharding projects192.723 GiB/node; transform arrays64 GiB/card. Need sharded random coefficients/grid/bitmap/support, distributed transforms, union/Frobenius validation and matching canonical support order. Two samples is an s17 observation, not a guaranteed s18 count. |
| Root factors / P product tree / normalization | Largest padded root-product domain is N/4;384 GiB aggregate product arrays,96 GiB/card. Representative support+root merge input/output payload242.893 GiB. Retained frontier, GC and recovery arrays are additional. GPU capacity fits, but the existing serial factor/carry schedule limits speed. | Root subtrees can be independent across nodes and merged exactly. Need deterministic partition/normalization, actual load balancing and checkpoint identity. Sixteen GPUs alone do not parallelize the host factor builder. Small products may benefit more from independent batches than one16-card collective. |
| Newton reciprocal G | Last three products pad to N/2,N,N. Need prefix-aware splitting, not just the validated full final-identity split. Last-step principal arrays633.446 GiB; split temporary lifts the stage model to697.446 GiB. If the dead old prefix is not collected, the retention envelope is825.446 GiB, before other overhead. | Balanced stage-aware square/output/P/support sharding projects158.362 GiB/node; old-prefix retention can add32 GiB/node. Full products96 GiB/card at N. Must shard Newton state and validate prefix multiplication; gathering to a coordinator recreates host pressure. |
| Independent full P·G identity | Fourteen N/8-block pairs; largest padded product N/4,96 GiB/card. Host principal arrays633.446 GiB plus up to64 GiB minus4 bytes temporary. Repeated GPU transforms/transfers and XOR assembly; full-size cost unknown. | One full product,96 GiB/card. Host principal arrays158.362 GiB/node if evenly sharded; compare every output coefficient to X^(N−1)+1 using disjoint ranges, joined status and complete digest. |
| Final G spectrum and canonical publication | Tiled full transform still needs integration with verified base upload/streamed export. P+G plus1 MiB stream tile projects256.001 GiB after command-owned support/recovery release. Tiling requires the512 GiB extension array in host RAM while it operates unless directly backed by a different validated design. | Resident transform64 GiB/card, stream output in canonical rank/range order; sharded P+G plus bounded tiles about64 GiB/node. Need distributed publication/digest and ordered client format, not16 unrelated code files. |
| Checkpoints, I/O and final verification | Same canonical950.170 GiB total. Recovery/scans/cache pressure charge the single host and filesystem. Canonical output size does not shrink with more GPUs. | Same950.170 GiB aggregate, not per node. Parallel I/O and sharded state can help only if storage supports it. Need coordinated atomic completion, source/shape checks, deterministic ordering and restart after rank failure. |
| Setup and communication | One node; no cross-node traffic. Extra CPU packing, host/device transfers, launch overhead and serial host work remain. | More contexts/plans, larger scheduler allocation and cross-node redistribution. Must measure effective transport, synchronization, staging and filesystem contention. Network implementation is absent. |

The final tiled-spectrum row deliberately separates the **post-transform export**
model from transform workspace. A host-backed512 GiB array plus live P/G alone
would be about768 GiB, even after support release; dead Newton arrays must be
collected/reused safely. The256.001 GiB export number is not the peak of a tiled
spectrum workflow.

### Newton audit: a capacity gap beyond the final identity

[Shared inversion](../../encoding/liftrs-shared/orbit_polynomial.go) doubles the
old reciprocal, allocates its square, then requests a prefix product.
[Product padding](../../encoding/liftrs-shared/orbit_product.go) uses the **full
operand degree**, even when only a prefix is returned, to prevent aliasing.
A truncated square view retains its full backing array. At the final s18 step:

| Host data | Payload and lifetime |
|---|---|
| Support |121.446 GiB through preparation in the current flow |
| P / root alias, counted once |60.723 GiB, retained |
| Old reciprocal prefix (2^35 coefficients) |128 GiB **while building the square**; dead before the new product output is needed |
| Square backing (2^36−1 coefficients) |256 GiB minus4 bytes, retained through convolution |
| New reciprocal G (N−K coefficients) |195.277 GiB, allocated for convolution output |
| Square-building principal stage |**566.169 GiB** (no new G yet) |
| Product principal stage |**633.446 GiB** (old prefix can be reclaimed) |
| Product plus split temporary |**697.446 GiB** |
| Conservative retention envelope if old prefix is not collected |**825.446 GiB**, not simultaneous necessary live data |

The model separates square construction from convolution. Adding old prefix and
new output as necessarily simultaneous would overstate required live memory by
128 GiB. Go liveness/GC can reclaim that old prefix; actual RSS may retain it and
other prior temporaries, so the825.446 GiB envelope remains relevant to admission
without being a lower bound. Neither number is measured RSS.

The three oversized Newton steps need3,7,13 contributing block pairs with an
N/8 block size. The full identity needs14; blindly replaying the identity driver
would not validate prefix truncation. All local padding must still cover each
full subproduct degree before its returned prefix is clipped.

A useful **shared host-lifetime change** is dropping command-owned support and
recovery references immediately after normalized root publication, before Newton,
while retaining certified metadata/checksum and immutable external caches. This
would reduce the stage-aware split-Newton principal model to about**576 GiB** (saving121.446
GiB), with no new polynomial operations or canonical output bytes. Keeping an
uncollected old prefix would instead leave about704 GiB of retained payload. It is not
implemented: the current shared function still uses support lengths and holds
support through preparation. Refactor ownership boundaries rather than mutate a
caller's support; actual GC/cache/RSS and restart behavior must be validated.
No zero runtime cost or whole-job memory admission is claimed. Discovery's
770.893 GiB model remains an independent one-node headroom concern.

## Runtime: what can and cannot be estimated

The complete latency is setup + discovery + roots + reciprocal + identity +
spectrum + persistence/readback + non-overlapped communication/checkpoint cost.
Checkpoint time already nested in a measured phase must not be counted twice.
**Neither complete route currently has a defensible measured runtime estimate.**

| Part | Measured s17 basis | s18 interpretation |
|---|---|---|
| Discovery |21 min38 s, two samples |Four times as many grid values; random sampling, host scans and tiling/network costs change. Count-only37.4 ms cannot predict this phase. |
| Root construction |8 h33 min57 s |K grows4.205-fold. Holding effective throughput fixed gives about36.0 h for linear K work, or40.9 h for K(log2K)^2 work. These are **work-scaling scenarios**, not lower/upper confidence bounds or route forecasts. Parallel subtrees can change them. |
| Reciprocal |24 min22 s |Three oversized products require splitting or distributed transforms; fresh host square/output and validation work remain. Unknown full-size latency. |
| Full identity |9 min05 s |Four-GPU split adds fourteen products;16-card distributed route adds communication. The small5.60-times ratio is not a full-size multiplier. |
| Final spectrum |4 min03.751 s |Four-GPU host tiling adds traffic/packing; distributed resident adds network. Unknown full-size latency. |
| Final cache verification |24 min44 s |About950 GiB canonical data at s18; hash/readback and cache behavior depend on storage and host distribution. This excludes earlier duplicate publication/verification passes. |

The retained resumed Go CPU profile attributes about32.65% cumulative samples to
`karatsubaLow`, with substantial base-field addProduct/mul work. About59.27% of
flat samples are in `runtime.cgocall`, which does not separate native CPU work,
GPU work or waiting. These are CPU samples over a mixed workflow, not wall-time
fractions or proof that the dominant stage is a GPU kernel. They support parallel
root/host investigation; they do **not** justify a fourfold16-GPU speedup claim.

### Network sensitivity, not a bandwidth measurement

Assume an even four-node ownership exchange where3/4 of a512 GiB array crosses
node boundaries per redistribution. A proposed two-redistribution transform
moves768 GiB across nodes; a six-redistribution analogue of the current local
schedule moves2304 GiB. A full product has three transforms, giving2.25 or6.75
TiB of aggregate cross-node payload under these respective assumptions.
These are different hypothetical schedules, not measured traffic or universal
bounds. Taylor source exchanges, host transfers, roots and protocol overhead
are excluded. Current same-node peer pointers cannot access another node's GPU.

| Effective egress per node (hypothetical) | Product communication at2 redistributions/transform | At6 redistributions/transform |
|---|---:|---:|
| 3 GiB/s |192 s |576 s |
| 12 GiB/s |48 s |144 s |
| 25 GiB/s |23.04 s |69.12 s |

These are ideal payload/throughput quotients assuming all four nodes send
concurrently, with no startup, synchronization, contention or extra exchanges;
not additional measured wall time and not a quoted HPC interconnect specification.
Compute/communication overlap is unproven. A bounded multi-node test must measure
usable bandwidth, transport support and actual pass counts before a full forecast.

## Persistence and resource admission

Canonical reusable files project**950.170 GiB**. The existing conservative
upper envelope for two complete snapshots per phase is another**528.001 GiB**:
**1478.170 GiB (about1.444 TiB)** combined, before logs, profiles, filesystem
headroom, failed attempts or simultaneous replacement of an old complete stage.
This is an upper envelope of retained phase generations, not an instruction to
write/reload all snapshots. Model storage is aggregate for either route; four
nodes do not multiply identical canonical artifacts. A fresh full run needs a
separate explicit storage plan and agreement; current approval does not cover it.

Existing s17 snapshots cost78.525 s,0.219% of the resumed worker, but fulls18
checkpoint overhead and distributed atomicity are unmeasured. Preserve completed
sampling, normalized roots and complete Newton prefixes; a failed transform
workspace is not a completed prefix. For distributed roots, retain global range,
source, normalized polynomial identity and a joined completion record so a rank
failure cannot publish a partial code as complete. Keep profiler and useful
progress/checkpoint output bounded without discarding prior raw logs.

## Recommendation and next implementation gates

Choose the **distributed16-GPU route for speed**, with four independent root
workers/nodes and balanced host ownership, then collective operations only where
needed. It avoids the demonstrated repeated host traffic and one-node
GC/headroom and tiled-transfer pressure. This is a design recommendation with unknown network/runtime
costs, not proof that16 GPUs are fastest. Compare smaller node/card counts in
bounded measurements when practical; do not force all small root products onto
sixteen cards. A root-worker prototype may deliver value before a full distributed
transform; the complete pipeline still needs both.

User selected distributed prototyping in SCALING-106; fullsize resources and
material slower modes still require separate agreement. Intermediate exchanges
use memory/direct messages. Persistence is separately validated with large
buffered writes and controlled file ownership/writer concurrency. The next work is:

1. Refactor command-owned support/root lifetimes at the normalized-root boundary,
   sharing immutable parameters and existing arithmetic; test restart/identity
   and memory ownership. This benefits either route and adds no deliberate
   recomputation or canonical storage, while actual latency/RSS remain to measure.
2. Prototype deterministic independent root subtrees with the same verified root
   factors and product kernels; compare complete polynomial bytes with the shared
   reference. Record host/native/GPU work separately; do not infer speed from
   worker count. Keep checkpoint ownership and global progress explicit.
3. Prepare a small two-node transport/redistribution comparison with bounded
   staging and real field/order/hash checks. Request a concrete multi-node budget
   before submission. Check installed transport support first; do not select an
   MPI/NCCL dependency from theory alone. Run jobs over30 s only by submission
   and handoff, with phase/work counters; fix the known single-vector logging gap
   before any long-vector component run.
4. Use actual pass/traffic/profile measurements to choose distributed transform
   ownership, Newton/product dispatch and checkpoint/publication integration.
   Keep ordinary CPU client builds free of transport/CUDA dependencies, and
   retain the validated resident server default whenever its admitted budget fits.

The comparison itself launched no job and selected no full-size budget, slower
default or mode. Subsequent bounded work is recorded in the active plan. s19–20 remain exploratory rows in the main reference.
