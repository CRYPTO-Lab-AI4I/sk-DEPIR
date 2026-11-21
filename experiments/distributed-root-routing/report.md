# Route root operands to their coefficient owners

Status: **submitted as78662**, run `routing-01`, producer `868a580`.
[Submission receipt](submission.json), [scheduler log](slurm-78662.log);
results will be retained under `runs/routing-01/`. No routed GPU result yet. This follows [passed root merge
78508](../distributed-root-merge/report.md), producer `0c16446`. The next allocation
records its frozen revision and complete source manifest.
[Readiness](readiness/checks.json), [runner](run.py), [Slurm entry](run.sbatch),
[analyzer](analyze.py), [review regressions](test_review.py).

## Shared implementation

The arithmetic remains unchanged: existing local GPU factor construction,
resident FFT/Taylor/pointwise product, independent native CPU root oracle and four
scalar root witnesses per rank. Only input ownership changes. The new private
[router](../../encoding/liftrs-shared/distributed_operands.go) assigns contiguous
coefficient intervals with the shared rank partitioner and routes bounded tiles
from each polynomial's owner directly to the target interval. It neither exchanges
nor materializes a complete peer polynomial, and it sends no zero padding.

The schedule is independent of CUDA, MPI and field degree. Two operand owners
may be different or equal; odd rank counts and empty target shards are supported
by the scheduler. Ordinary CPU/race tests cover one, two, three and five ranks;
this does not establish multi-rank GPU execution beyond the admitted two ranks.
Each output array is explicitly admitted with `MaxLocalWords`; callers must also
budget both arrays, owned source polynomials and transport/runtime memory.
Inputs are borrowed and preserved; a failure invalidates outputs and requires
collective abort rather than continuing peers with missing data.

The [MPI adapter](../../encoding/liftrs-shared/distributed_operands_mpi_test.go)
uses existing control exchange, coefficient codec and before/after payload proofs.
Descriptors bind operation, domain, operand, coefficient offset/count and both
owners. The reviewer checks every planned route and paired proof against the
actual saved source coefficients, then independently reconstructs both saved
input shards. Every final coefficient still matches the native/scalar reference.
A new arithmetic kernel or second root/product recurrence is unnecessary.

The [root driver](../../encoding/liftrs-shared/distributed_root_product_gpu_test.go)
selects `PIR_ROOT_INPUT_LAYOUT=peer|routed` explicitly. The prior peer entry stays
available and remains the default for that diagnostic test when the environment
variable is omitted. No production default, client dependency or public API changes.
Support and complete CPU reference are still replicated **only for this bounded
diagnostic gate**; routing polynomial inputs does not complete partitioned support,
the reduction-tree executor, Newton/generator work, discovery or fulls18 support.

## Memory, traffic and time trade-off

This step compares alternatives; it does not adopt a slower memory strategy as
a default. The routed path removes one full peer-polynomial host allocation per
rank; it retains the owned local polynomial, two host coefficient shards and the
existing GPU arrays. No additional large host array or device allocation is added.
For the largest admitted shape (N65536, each operand32767 coefficients):

| Cost | Peer reference | Routed prototype | Evidence |
|---|---|---|---|
| Complete peer polynomial in host RAM | 131,068 bytes/rank (about128 KiB) | None | Layout-derived payload, not measured RSS saving |
| Aggregate input network payload | 262,136 bytes | 131,068 bytes | Projected from ownership; source rank1 sends its right operand to rank0; no left-operand or padding transfer needed |
| Transfer granularity | Whole local polynomial exchange | At most4,096 coefficients (16 KiB) per tile;8 packets at largest size | Additional descriptor/proof handshakes may cost time |
| Principal GPU arrays | Existing resident arrays | Same | No kernel/layout change |
| Host output shards | Two arrays of128 KiB/rank each | Same | Explicit per-array admission |
| Input routing latency |78508 peer observed maximum13.462 ms at this shape | Unknown | Compare in the next allocation; includes skew/waits and bookkeeping |
| New retained storage | Saved local/peer roots, support/reference/output evidence | Replaces peer file with two padded input-shard files | A few MiB extra diagnostic evidence across suite; projected cache below32 MiB |

The receiver writes directly into its owning host shard; the sender borrows an
existing source slice. No separate complete receive polynomial or route staging
array is allocated. Existing runtime, MPI proof and serialization buffers remain.
The two input shards are subsequently uploaded by the unchanged shared product
driver; direct streaming into device storage is **not** claimed here. The tiny
fixture's128 KiB saving will not dominate process RSS; the purpose is correct
ownership composition before admitting larger data. Fullsize savings/runtime
cannot be inferred from this small workload.

## Exact validation and measurement

Five MPI arms:

1. Whole-peer root reference atN65536.
2. Routed small CUDA memcheck/leak case atN4096.
3. Routed normal roots atN4096.
4. Routed normal roots atN16384.
5. Routed normal roots atN65536.

All use the same s17 canonical field, deterministic synthetic Frobenius-closed
supports ofN−4 roots, existing GPU factor construction and distributed final
product. The final polynomial spans both output ranks. These are synthetic root
subsets, not full PIR supports. Larger cases are not sanitized. One normal sample
per geometry/layout establishes integration, not a statistically strong speedup.
The same-allocation peer reference retains a direct check of existing behavior.

The independent reviewer retains all product/Taylor/transpose route, preservation,
reuse and canonical-range checks. It also requires exact source-derived operand
proofs, no peer file in routed arms, byte-for-byte input-shard reconstruction and
expected per-rank network counts. Focused regressions reject missing/wrong routes,
paired but wrong payload proofs, incorrect saved inputs and an invalid race-run
CGO environment. CPU tests cover source preservation, memory admission, transport
errors and uneven/empty partitions with unbuffered in-memory links.

Final-product timing boundaries are unchanged from78508: upload, distributed
product/readback and proof logging, with diagnostic spectrum capture separate.
Input timing records local factor work, complete reference/witness work, routing
and evidence persistence separately from product work; constructor setup stays
separate. Routing includes control/payload proofs, waits and coefficient-shard allocation/
padding/copy. This intentionally extends78508's peer-only `exchange_seconds`
boundary to complete input assembly; use the new same-allocation peer reference
for a runtime comparison. Historical13.462 ms is context, not a matched baseline. Go CPU profiling
remains enabled; it cannot separate GPU execution and waits inside cgo. Extra
input-shard persistence is outside routing/product clocks, inside complete input
and controller time. Compare like boundaries and report measured costs, including
any runtime penalty, before choosing a future production path.

## Resources, preflight and reproduction

Same admitted **two HPC nodes / two H200s / four CPU cores / 8 GiB requested RAM**,
one GPU/two cores/4 GiB per node. Same **128 MiB aggregate owned GPU/card**,
**1 GiB RSS/rank**, no pinned buffers, **128 MiB cache / 2 GiB scratch /
32 MiB per file**, **720-second controller / 15-minute scheduler**. Product/factor
workspace overlap remains bounded as in78508; factor engine closes before FFT
work. No resource increase. Eligible nodes are allowed except gnode03.

78508 completed five arms, builds and profiles in148.668 seconds. This suite adds
router CPU/race checks and tiny tile/proof/input-file work; budget **3–6 minutes
estimated**, excluding queue, within the existing controller/scheduler cutoffs.
Retain the validated240-second launcher/220-second test guard per arm. Projected
cache below32 MiB and largest proof file below16 MiB leave headroom under admitted
128 MiB cache/32 MiB file limits. These new costs remain unmeasured.

Reuse node-by-node shared executable/wrapper/library/cwd/TMPDIR preflight and UCX
registration-cache controls. The race check explicitly enables CGO before invoking
`-race` (the base controller starts with CGO disabled); a focused test guards this
launcher requirement. Timestamped phase/progress/resource events and each
coefficient packet proof flush with logs; completed arms/profiles persist before
later phases. No agent monitoring after submission.

```sh
python3 -B experiments/distributed-root-routing/test_review.py
python3 -B experiments/distributed-root-routing/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-root-routing/slurm-%j.log \
  experiments/distributed-root-routing/run.sbatch routing-01
# After the user's completion ping:
python3 -B experiments/distributed-root-routing/analyze.py routing-01
```

Raw results/commands/profiles/logs: `runs/routing-01/`. Source snapshot, binaries
and coefficient evidence: `.cache/liftrs/distributed-root-routing/routing-01/`,
outside Git with retained checksum manifests. Earlier caches/profiles remain.
After this gate, integrate rank-owned reduction/repeated and truncated products
for Newton/generator identities through the same executor; admit s18 arithmetic
and partitioned support independently. Full distributed preprocessing is unrun.
