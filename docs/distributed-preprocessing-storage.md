# Distributed preprocessing: GPU-resident working storage

This design began after [small distributed products passed78321](../experiments/distributed-gpu-product/report.md).
Resident storage,complete FFT/products and the genuine distributed pipeline now
have bounded correctness and scaling evidence through s8/curve5 (81286).
See the [maintained status reference](encoding-status.md#current-distributed-preprocessing-step)
for current gates. Fulls18 remains unrun;the current allocation boundary is
[bounded discovery-result consumption](../experiments/distributed-discovery-results/report.md).

The existing host-shard implementation is a correctness reference. Its conservative
product model includes six full host shards, while GPU arithmetic uses bounded
tiles. At s18 this host model is3072 GiB aggregate before overhead. Increasing
that reference workload does not resolve the storage design needed for fast large
preprocessing. Prototype resident storage first, retaining the same arithmetic,
canonical values, ownership schedule and independent references.

## Original implementation sequence (bounded validation completed)

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

| Resource/cost | Validated host reference | Resident backend (bounded validated) |
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

The [assembled genuine pilot](../experiments/distributed-preprocessing-pilot/report.md)
now connects the reusable resident executor to root/Newton/identity/spectrum work,
with host coefficient shards and bounded sparse squaring transfers. GPU validation
is pending. Its discovery and artifact publication use a coordinator, and prefix
reload stays within one allocation; those are not a fulls18 ownership/recovery
layout. New host-state lifetimes and communication need measurement before any
larger resource admission, while the validated three-array GPU layout is retained.

The integrated genuine pipeline passed78898, including prefix reload and canonical
artifacts. Next partition discovery grids/union state in that pipeline. The new
mode uses the same three resident GPU arrays for tensor evaluation/interpolation,
local monomial lists/union bits and bounded index/sample traffic. Its deterministic
orbit assignment can be imbalanced; aggregate host/device memory and runtime must
be measured before a larger budget is admitted. Diagnostic archive and final
publication gathers remain small-pilot boundaries, not a fulls18 storage layout.

Partitioned discovery passed78909 in the genuine pipeline. Next remove complete
publication receive arrays and serialized buffers through explicit streamed output,
and resume a persisted prefix in fresh MPI processes. This changes host I/O/state
ownership, not GPU arithmetic; keep CPU corruption/stream/race checks and one
integrated cold/resume comparison. Diagnostic CPU/support arrays still remain,
and resumed rank count is fixed. No fulls18 memory/node-failure claim follows.


## Bounded host spectrum publication

The complete genuine s8/curve5 pipeline passed81286 with resident spectrum output.
The [next explicit prototype](../experiments/distributed-spectrum-publication/report.md)
keeps the same GPU transform in slot0 while the shared canonical publisher reads
4096-word tiles. One download feeds diagnostic and canonical output;no additional
FFT or genuine output/transfer bytes. Wide host spectrum allocation falls by8N/r
minus bounded buffers (s18:512GiB aggregate,32GiB/worker at16workers). This is an
isolated allocation projection,not total-RSS measurement or topology admission.
Fullsize throughput may change from interleaving file/network/GPU reads;resident
remains default and the experimental source option stays explicit. Complete
discovery/root/reciprocal working sets and production persistence remain separate
fullsize admission boundaries. GPU composition validation passed **81365**,including canonical wide/high-bit checks
and a matched resident arm. Source cold was3.545%slower in one bounded sample;
fullsize latency remains unknown and resident stays default. Next consume discovery
results directly from GPU tiles into the local union bitmap,avoiding the same
8N/r host allocation in that different phase. Do not sum those phase-specific
savings into a simultaneous peak-memory claim.


Optional discovery-result composition validation passed **81454** on
the existing two-node/two-H200 budget. It holds spectrum mode constant while
comparing discovery tiles with resident results. [Scope and evidence](../experiments/distributed-discovery-results/report.md).
Fullsize ownership/liveness and production persistence/logging
remain separate gates before fulls18 admission.


The [fullsize audit](../experiments/distributed-fullsize-audit/report.md) now
records actual code allocation/traffic/storage/logging limits. Next validate
[explicit Newton prefix inputs](../experiments/distributed-prefix-inputs/report.md),
removing the4N/r root clone/mask and uploading only selected input intersections
into the already zero-reset GPU slots. Dense default/correctness checks remain.
At s18/16workers the clone is16 GiB/worker;this is an avoided allocation,not a
measured total-RSS saving. Production bounded persistence/proofs and parameter/rank
ownership admission remain;small-case success does not admit fulls18.


Prefix-input validation passed **81729**,within the existing two-node/
two-H200 caps. [Exact scope and evidence](../experiments/distributed-prefix-inputs/report.md).
Dense default unchanged. The source-based audit is now the
reference for remaining production lifecycle/rank,persistence/proof/logging and
per-phase memory admission work;fulls18 remains unrun.


Next production recording step replaces optional per-proof output with fixed
per-kind ordered digests/counts and bounded completed-work progress. Every live
verification and per-proof RSS guard remains. Native Taylor windows also need
bounded totals with parent pass/target progress and errors retained. Full traces
remain the default and small reference for paired validation;this changes stored
evidence,not mathematical checks. Production parameters/ranks/non-oracle/persistence
remain separate work before fulls18 admission.

Bounded proof/native recording comparison **81738** passed under the existing two-node/two-H200 budget; [receipt](../experiments/distributed-proof-recording/submission.json). All six arms passed.


The next [bounded diagnostic serialization validation](../experiments/distributed-diagnostic-writes/report.md)
keeps diagnostic files byte-identical and retained storage unchanged, but explicitly
limits the extra serialization buffer to 64 KiB. Bulk writes remain default; extra
write calls and large-file throughput remain a measured/unknown trade-off. This
addresses temporary buffers, not production retention or coordinated restart.

Diagnostic serialization validation **81744** failed final controller review after passing bounded cold/restart and bulk control; wide inherited bulk mode. [receipt](../experiments/distributed-diagnostic-writes/submission.json). Retain partial passes; retry only missing bounded wide. Full s18 remains unrun/unadmitted.

SCALING-182: explicitly reset the diagnostic mode after the bulk control, test actual suite configuration before allocation, and rerun only bounded wide using [the targeted runner](../experiments/distributed-diagnostic-writes/run-wide.py). No arithmetic change, repeated cold/restart or resource increase.

Targeted missing bounded-wide validation **81746 passed** under unchanged caps; [receipt](../experiments/distributed-diagnostic-writes/submission-wide-02.json). All72bounded files/native/canonical checks passed;passing81744cold/restart/bulk phases were not repeated.


## Coordinated restart support binding (SCALING-183)

Bounded diagnostic composition is complete through81744retained partial passes
and81746passed targeted wide;original81744overall failure remains visible.
Next use one full coordinator support checksum plus fresh agreement on all workers'
complete root/prefix metadata. Reuse the same recovery/shard validator, maintain
trusted immutable input assumptions and full checksum coverage, retain local mode
as default, and measure coordinated versus local fresh restart from the same cold
cache. Capture both fixed control-frame rounds for independent metadata/digest
review;do not introduce unbounded packet logs or remove checks. Projected support
verification reads at s18/16workers:3.795TiB->121.446GiB aggregate,with two collective
rounds (about30KiB total fixed payload);latency remains unknown. No extra support
storage or GPU transforms;bounded hash/control buffers. [Exact design/resources](../experiments/distributed-restart-binding/report.md).
Same admitted two-node/two-H200 comparison;no unchanged wide/sanitizer repeat.
Production ownership/lifecycle/non-oracle entry,bounded retention and actual phase
count/domain/resource admission remain;full s18 is unrun/unadmitted.

Coordinated restart binding validation **81750 passed** under unchanged caps; [receipt](../experiments/distributed-restart-binding/submission.json). All three cold/coordinated/local arms passed;no fulls18 admission.


## Internal runtime extraction (SCALING-184)

81750validated one coordinator support hash versus eight local hashes,with all
metadata/native/shard/canonical checks. Small cached restart6.792011s versus6.661931s
local;keep local default and distinguish logical traffic from physical I/O/fullsize
latency. Update the main reference with measured resources/counters.

Next move existing operand/transpose/Taylor MPI transports and GPU session wiring
from test-only code into internal runtime,sharing constructor/progress/operation
identity and product/spectrum logic across narrow/wide cases. Keep test error
adapters and independent oracles separate;runtime returns errors,owns GPU close
and rejects failed/closed reuse. Preserve exact protocol checks and storage/kernel
bodies. Explicit build variants retain CPU-only and optional GPU client support.
The current experiment-scoped MPI bridge and bounded harness remain;this does not
create a public production API or admit larger ranks/full s18. [Exact scope](../experiments/distributed-runtime/report.md).

Validate local build matrix/admission/lifecycle errors,unchanged native references
and one existing-budget cold/coordinated-restart/wide composition. No repeated
local-vs-coordinator or unchanged sanitizer arm. Then continue production parameter/
ownership/phase admission,non-oracle entry and bounded retention/publication.

Shared MPI/GPU runtime validation **81763 passed** under unchanged caps; [receipt](../experiments/distributed-runtime/submission.json). All cold/coordinated-restart/bounded field18 arms passed;no public production/full s18 admission.


## Shared reciprocal and full-identity runtime (SCALING-185)

81763 passed the extracted MPI/GPU runtime on genuine s8cold/restart and bounded
field18. Next deduplicate the remaining main/wide Newton loop into one backend-
independent function with parameterized ownership/counts and error-returning
square/product/checkpoint callbacks. Keep all arithmetic, counts, dense/prefix
selection, checkpoint/reload order and native/full-identity checks. Move the
identical full code identity scan to runtime; do not treat inverse refinement alone
as verified code preparation. No oracle input or test framework enters the runtime.

No extra intended arrays/transfers/transforms/storage; dense default and explicit
prefix savings stay unchanged. Single-shard admission is not a whole-process memory
bound, and full-size GC/runtime effects remain unmeasured. [Exact plan](../experiments/distributed-reciprocal/report.md).
Use CPU cold/warm/failure references and one existing-budget cold/restart/default-
dense-preservation/wide integration; no unchanged sanitizer. Then continue production
discovery/root/phase ownership, parameter admission, non-oracle end-to-end entry and
bounded final publication. Full s18 remains unrun/unadmitted.

Shared reciprocal runtime validation **81777 passed** under unchanged caps; [receipt](../experiments/distributed-reciprocal/submission.json). All cold/restart/dense-preservation/wide arms passed;no public production/full s18 admission.


## Root runtime and actual-count admission (SCALING-186)

81777passed shared reciprocal cold/restart/dense/wide integration. Next extract
local root and subgroup lifecycle from test-only helpers, preserving factor/product/
reduction kernels and abort/close ordering. Diagnostic saves become optional runtime
observers while bounded tests retain every native reference comparison. Admit actual
leaf and rounded stage shard lengths before leaf GPU work;this single-array cap is
not aggregate RSS/device/storage admission. [Scope/resources](../experiments/distributed-root-runtime/report.md).
Validate CPU guards/failures, native compatibility and existing-budget genuine cold/
restart plus owned-wide root merges. No unchanged reciprocal/dense/wide-tail/sanitizer
repeat. Then continue production discovery/phase ownership, non-oracle end-to-end
entry, complete resource admission and bounded retention/publication. Full s18 unrun.

Root-runtime validation **82085 passed** under unchanged caps; [receipt](../experiments/distributed-root-runtime/submission.json). All cold/restart/owned-wide arms passed;no full s18 admission.


## Discovery GPU session integration (SCALING-187)

82085passed root runtime/admission/native/canonical checks. Post-run analyzer's
source path corrected with producing-revision/hash binding and a regression;
raw results unchanged, no rerun. Next replace discovery's duplicate session setup
with shared allocation/transport/progress/error/close, retaining sample randomness,
order,union/stopping and native/proof checks. Both stream/resident branches use
one method;resident default remains. No extra intentional arrays/transforms/I/O.
[Plan](../experiments/distributed-discovery-runtime/report.md): same-budget streamed
cold/coordinated restart/resident preservation; no unchanged wide/dense/sanitizer.
Production discovery sampling/ownership, full phase admission/non-oracle entry and
bounded retention/publication remain;fulls18 unrun/unadmitted.

Shared discovery-session validation **82096 passed** under unchanged caps; [receipt](../experiments/distributed-discovery-runtime/submission.json). All streamed cold/restart/resident-preservation arms passed;no full s18 admission.


## Explicit-policy discovery loop runtime (SCALING-188)

82096passed shared session integration. Next move sampling/union/closed-orbit routing/
archive loop from tests into error-returning runtime with explicit randomness,
CUDA/result options and logical list/shard caps. Preserve original good enumeration,
entropy/sample/masking/transform/count/proof/order semantics;seed109stays in test
adapter,not runtime default. Caller aborts peers on errors;no partial publication.
[Plan](../experiments/distributed-discovery-loop/report.md). Native GoodMonomialRange
is still CPU server work;no GPU-only/fastest fullsize claim. No new intentional
arrays/transforms/transfers/files;single-list bounds are not aggregate memory.
Validate local policy/count/wire errors plus same-budget cold/restart/resident
composition,no unchanged wide/dense/sanitizer. Then assemble production parameter/
phase ownership/non-oracle entry,bounded persistence/final publication and agree
larger topology/whole-job resources. Full s18 remains unrun/unadmitted.

Explicit-policy discovery loop validation **82097 passed** under unchanged caps; [receipt](../experiments/distributed-discovery-loop/submission.json). All cold/restart/resident-preservation arms passed;no full s18 admission.


## Reference-independent runtime workflow (SCALING-189)

82097passed discovery runtime. Next compose existing discovery/local-root/subgroup-
root/reciprocal/full-identity/spectrum/canonical sections into one error-returning
workflow without native-reference inputs/testing policy. Caller supplies explicit
random/resource/output/archive/checkpoint policy;test observers save intermediate
arrays for independent comparisons. Preserve operation/kernel/transport identities
and defaults;no intended extra coefficient arrays/transforms/transfers. Fullsize
GC/latency/RSS remain unmeasured. [Plan](../experiments/distributed-workflow/report.md).

Validate same-budget reviewed pilot cold/restart plus assembled cold, comparing
all compact checkpoints,full proof digests/native roots/identity/final bytes. Observer
writes are not production coordinated checkpoint commits;timing boundaries differ,
so no direct speedup claim. No unchanged wide/dense/sanitizer repeat. Afterward
integrate durable restart/atomic publication and parameterized command,then larger
representative topology/complete resource admission. Fulls18 unrun/unadmitted.

Reference-independent assembled workflow validation **82099 controller failed; all executed arms independently verified** under unchanged caps; [receipt](../experiments/distributed-workflow/submission.json). Pilot cold/restart/assembled outputs pass corrected native-metadata review;original failure retained,no production restart/full s18 admission.


## Final workflow publication commit (SCALING-190)

82099executed all three arms successfully;legacy reviewer then incorrectly required
an in-process reference event. Keep raw controller failure,verify native metadata
externally with strict checks/regression,and recover saved outputs/profiles without
GPU rerun. All76comparison files/proof summaries/native/final bytes match.

Next explicit final publication policy stages payloads,hashes/counts during writing,
verifies runtime/header/extents,and commits a synced manifest last. Avoid fullsize
reread/copy;retain incomplete files/reservations on failure. Extra SHA CPU and
manifest/directory sync cost remain unmeasured;direct observer option retained.
[Plan](../experiments/distributed-publication-commit/report.md). Same-budget cold/
restart+committed assembled validation,no unchanged wide/dense/sanitizer. Then
coordinated durable prefix restart/parameterized command/full phase admission and
representative topology. Fulls18 unrun/unadmitted.

Final publication-commit validation **82100 passed** under unchanged caps;
[verified evidence](../experiments/distributed-publication-commit/runs/publication-01/analysis.json).
Controller **2 min38.073 s**; cold **8.907 s**, restart **6.374 s**, committed assembled
**8.510 s**; assembled peak host RSS **378.824 MiB/worker**. All76comparison files,
proof/native checks and three manifest-bound payloads match. Timing boundaries differ;
no direct speedup claim. Fulls18 remains unrun/unadmitted.


## Durable assembled prefix commits (SCALING-191)

82100passed final publication. Connect source-bound compact root/Newton commits
using existing codec and checksum loader, without another support payload copy.
At discovery completion sync/seal the streamed archive and hard-link it under the
checkpoint loader's support name. Bind actual code header/curve/field/ranks/sample
count/raw archive format and digest in a synced source record; bind its digest in
each prefix record. Agree stage metadata before writes; sync all rank shards,
collect digests, publish commit marker last and barrier before advertising success.
Existing legacy checkpoint records remain loadable without the optional source digest.

[Implementation and validation plan](../experiments/distributed-workflow-checkpoints/report.md).
Explicit job option; default observer path unchanged. No added numerical arrays,
transforms, GPU transfers or payload copy/reread on the cold path. Added metadata,
control exchanges and sync latency are unmeasured; same-filesystem hard links and
POSIX sync semantics required. Preserve incomplete data. Source hashes detect
corruption, not malicious rewriting. Logical file inventories count the support
alias again although its physical payload blocks are shared.

Same-budget pilot cold/restart plus assembled durable commits, independent checks
of all17commit manifests/68shards and source alias, then root/prefix8 loader in a
fresh CPU process with profile. This validates persisted recovery inputs; arithmetic
resume in the assembled runtime remains next, followed by parameterized command,
phase admission and representative topology/fulljob approval. No unchanged wide/
dense/sanitizer repeat. Fulls18 remains unrun/unadmitted.

Durable assembled checkpoint validation **82101 passed** under unchanged caps;
[verified evidence](../experiments/distributed-workflow-checkpoints/runs/checkpoints-01/analysis.json).
Controller **2 min41.494 s**; cold **10.400 s**, pilot restart **7.532 s**, assembled
**9.893 s**; assembled peak host RSS **380.570 MiB/worker**. All17commit manifests,
68shards, source binding/alias, native/proof comparisons and fresh-process file
reload pass. Arithmetic restart integration remains next; fulls18 unrun/unadmitted.


## Assembled arithmetic resume (SCALING-192)

82101passed durable source-bound prefix commits and fresh-process file reload.
Now restore root and a completed Newton prefix before GPU allocation, with the
existing coordinated support verifier: one full support hash on the coordinator,
complete metadata agreement for each record, unchanged shard/hash/EOF/range/padding/
normalization checks. Require caller-supplied trusted source digest, matching full
code spec/rank count and checkpoint source binding. No discovery/roots are rerun.
Cold and restarted entries share one Newton/identity/spectrum/publication tail.

[Plan and implementation](../experiments/distributed-workflow-resume/report.md).
Reuse the immutable verified support archive through same-filesystem hard links
rather than copying a fullsize payload into the new publication. New checkpoints
still commit per-rank shards and source metadata; final manifest binds the reused
archive and newly computed code/generator. Source cache must remain immutable.
No new numerical arrays/transforms/transfers; restart needs two resident shards
and their disk reads plus one support hash. Fullsize reload/verification/runtime
remain unknown. This explicit restart policy leaves existing defaults unchanged.

Validate reviewed pilot cold/restart plus assembled cold and fresh-process prefix8
restart under unchanged caps, with full remaining arithmetic proof/native-work,
checkpoint and final-byte equality. Do not repeat unchanged wide/dense/sanitizer
comparisons. Then parameterized command/full phase and working-set admission,
representative larger topology, and complete fulls18 resource approval. Fulls18
remains unrun/unadmitted; failure during discovery/root construction is not yet
covered by this completed-Newton-prefix restart boundary.

Assembled arithmetic-resume validation **82102 passed**;
[verified evidence](../experiments/distributed-workflow-resume/runs/resume-01/analysis.json).
Controller **3 min11.398 s**, assembled cold **10.691 s** /restart **8.297 s**;
restart peak host RSS **381.387 MiB/worker**, owned GPU about **0.244 MiB/worker**
(context/driver extra). Remaining13steps/proof/native work/checkpoints/final bytes
match the pilot, with one support hash/no-copy archive. Fulls18 unrun/unadmitted.


## Parameterized distributed server command (SCALING-193)

82102passed assembled cold/fresh-prefix8 arithmetic restart. Add the actual
[command](../cmd/pir-preprocess-mpi/README.md) around shared runtime, independent of
test oracles, with explicit JSON parameters/modes/limits, all-rank policy agreement,
secure default randomness, mandatory profiles and periodic local resource guards.
At exact K, admit canonical/compact checkpoint/metadata retention before GPU
discovery allocations; retain local list/shard/rounded root/CUDA guards.
[Validation plan](../experiments/distributed-command/report.md): actual binary cold/
restart compared with reviewed pilot under unchanged two-node/two-H200 caps.

Omit test-only support/root-stage/identity copies while preserving live checks and
full proof/native/output comparisons. New admission/watchdog/metadata overhead is
unmeasured fullsize; no numerical array/transform/transfer change intended. Memory/
log policies stay explicit; command does not authorize resource increases.
Complete fullsize host/GC/native working-set/headroom admission still remains,
then representative larger rank/topology and fulls18 resource agreement. Current
samplers and individual array caps are not a complete phase memory model.
