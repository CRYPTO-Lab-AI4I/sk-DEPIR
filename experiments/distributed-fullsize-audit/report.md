# Distributed fullsize allocation and execution audit

Status: source-based audit;no HPC job submitted by this experiment.
[Runnable model](audit.py),[checks](test_audit.py),[results](results/projections.json).
Results record revision/source hashes;regenerate after code changes. These are
allocation/traffic formulas and implementation findings,not measured peak memory,
fullsize runtime,support or resource admission. Historical full resident s17 runs
remain passed;the unrun target here is the new distributed fullsize path.

```sh
python3 -B experiments/distributed-fullsize-audit/test_audit.py
python3 -B experiments/distributed-fullsize-audit/audit.py \
 --output experiments/distributed-fullsize-audit/results/projections.json
```

HPC source audit using Go1.27.3;no GPU,matrix,data-domain allocation or cluster
query is needed. K at s16/17/18 is918672421/3876566203/16300262095 for m2/t5,
from [exact weighted count76647](../s18-preprocessing-admission/report.md).
K at s19/20 is unknown;do not fabricate Newton/storage figures there.

## What the device payload says

One worker/card is an illustrative mapping. Three extension arrays need3Nw/P,
where N=2^(2s),w4bytes at s16 and8bytes thereafter. Actual owned memory additionally
includes tables,plans,transport/windows and factor-engine state;context/driver
and MPI state are outside owned bytes. A112 GiB/card cap is a **scenario**,not
this experiment's approved cap. Existing validation still uses64 MiB/worker.

| s | Three-array payload at16 GPUs | Payload-only minimum power-of-two GPU count at112 GiB/card | Implementation/capacity limits |
|---|---|---|---|
| 16 | 3 GiB/card | 1 | Distributed complete path only measured on bounded cases |
| 17 | 24 GiB/card | 4 | Full resident path passed historically;new distributed fullsize path unrun |
| 18 | 96 GiB/card | 16 | Plausible array payload;complete memory/topology/rank/persistence admission unresolved |
| 19 | 384 GiB/card | 64 | Exploratory;current CUDA field-degree admission stops18 and current Taylor capacity3q exceeds its1-Miword cap |
| 20 | 1.5 TiB/card | 256 | Exploratory/hypothetical hardware (availability unverified);also exceeds current64-rank bound and CUDA/Taylor limits |

Eight cards would require192 GiB/card for s18's three arrays alone,so host streaming
cannot make that layout fit eight H200s. Sixteen cards reduce those arrays to96 GiB
before extras;this is neither proof of complete feasibility nor permission to
request them. More/fewer nodes and cards must be chosen for complete latency and
actual available hardware rather than a fixed two-GPU design ceiling.

Root stages use their own rounded polynomial domain and subgroup size. Exact
per-owner counts at fulls18 are unknown;the stable orbit-owner rule does not
promise balance. Local root factors and each subgroup's maximum rounded domain
must be admitted before allocation. The final root domain depends on K,not the
complete grid N;intermediate per-card peaks need not follow the final-stage mean.
Device factor workspace closes before each later rank workspace,as in the source.

## Host allocations and their actual boundaries

At s18/16workers,a full base array is16 GiB/worker;extension array32 GiB/worker.
The local union bitmap is0.5 GiB/worker. Discovery still retains a16 GiB base grid
and an8-byte index list of local good monomials,with append-capacity overhead;
raw support/orbit roots and routing buffers have different ownership. Their exact
sizes require local counts. Tiled discovery avoids the32 GiB result array in that
phase;streamed spectrum avoids it in a later phase. **Do not sum these as64 GiB
peak-RSS savings.** Go compiler liveness,capacity growth,retained references and GC
are not established by counting variable declarations.

Newton before upload can have named root,old generator,squared generator and
masked-root clone arrays. At s18/16workers these four base-array inventories total
64 GiB/worker,excluding everything else. Prefix input views remove the16 GiB
clone,leaving48 GiB in that inventory. The result allocation occurs later;whether
old arguments remain alive then is compiler/GC dependent. Neither inventory is
a measured peak or an upper bound;garbage/native/runtime/other phases can add more.
Do not declare s18 admitted by multiplying these inventories by node packing.

The strict dense loader already zero-resets GPU input slots. An explicit logical
prefix view can upload only each selected coefficient interval,leaving that reset
tail zero. This avoids clone/mask and redundant zero uploads without a second
product recurrence. [Prototype and validation](../distributed-prefix-inputs/report.md).
Dense padding validation and the existing default must remain unchanged.

The diagnostic `saveMatrixStageWords` uses `encoding/binary.Write`. The installed
Go1.27.3 implementation creates a full serialized byte buffer for uint32/uint64
slices (`GOROOT/src/encoding/binary/binary.go:Write` fast path). A base snapshot
can therefore add another16 GiB temporary per s18 worker;legacy spectrum dumps
could add32 GiB. Compact checkpoints and the validated streamed spectrum writer
have bounded codecs,but P/G/identity/tree diagnostics still need a bounded or
omitted production strategy. No old snapshots are deleted by this audit.

## Host/device input traffic is separate from FFT traffic

For exact s18 K,the current schedule has36Newton updates. Dense uploads of both
full operands total **36 TiB** across workers. Explicit prefix views project
**1,394.893 GiB (approximately1.362 TiB)**,a26.43-fold reduction in this input
traffic alone. GPU resets,FFT/Taylor work/inter-rank messages,output readback and
persistence are unchanged. This is **not** an end-to-end speedup estimate.
Full-domain3FFT products are still used even for small prefixes;growing transform
domains remain a potential runtime design,not implemented by these views.

## Persistence,logging and production entry remain hard gates

At s18,current diagnostic rank P/G/identity files add768 GiB;diagnostic spectrum
adds512 GiB. Cold compact root/Newton checkpoint payload is approximately512 GiB,
versus9.25 TiB padded. Canonical code payload is572.723 GiB,generator195.277 GiB,
support121.446 GiB,excluding headers/metadata/root-stage copies/profiles/temp data.
These overlap in contents but are separate files;do not pretend they are free
aliases. A production retention policy must avoid unnecessary new duplication
without deleting reusable verified caches. Filesystem capacity/throughput and
resource admission must cover the complete retained set and restart scratch.

Current fresh restart hashes the complete global support file in each rank for
both root and prefix loads. At s18 that is242.893 GiB per worker,about3.795 TiB
aggregate for16workers,before shard reads. A coordinated verified-manifest or
partitioned support binding needs design before calling this an efficient restart;
no checksum or correctness check is silently removed.

Current4096-word proof logging cannot be scaled unchanged. Main Newton/identity/
spectrum alone use112FFTs at s18. Their transposes would emit **440,401,920 proof
rows per worker** at16workers,excluding Taylor/discovery/root logs. Production
needs bounded summaries/progress while preserving every runtime descriptor,
payload/hash/field/identity check. The actual bytes and runtime of these fullsize
logs are unmeasured;small logs are not evidence that this is cheap.

Current complete runner is a bounded test harness:case selector stops at s8,
retained native reference admission at N≤2^20,owned-root pilot requires four ranks,
resources are64 MiB device/worker and32 MiB/file,reference/archive filenames and
full native comparisons remain mandatory. Library scheduling helpers cover larger
ranks mathematically,but that is not a production16-rank implementation. Fullsize
needs a parameterized non-oracle execution/verification/persistence entry,explicit
real resource guards and supported resume ownership. Do not just increase limits.
Also evaluate CPU GoodMonomialRange's O(N) monomial enumeration and coordinator
sampling/entropy/transfer costs;the GPU weighted count is count-only and cannot
replace actual discovery. No full runtime estimate is justified by this audit.

## Execution order after this audit

1. Validate explicit prefix-input views and their clone/upload counters against
   dense input,same canonical outputs and bounded wide arithmetic.
2. Assemble production parameter/rank/lifecycle ownership and bounded persistence/
   proof/progress recording using shared executors;retain runtime checks and
   canonical verified caches. Remove mandatory full native oracle dependence.
3. Admit each phase using real local good/support/root counts and rounded group
   domains;validate greater rank count/larger representative work only where the
   assembled production path needs it. Keep source/reference tests bounded.
4. Agree complete s18 topology,memory/storage/headroom/time resources and restart
   policy after those paths and evidence are concrete. Fulls18 remains unrun.
