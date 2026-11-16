# First guarded full s17 GPU preprocessing

Status: **approved for first local-layout full trial; submission follows**. Exact count and
complete small workflow validation pass job75810; no full s17 support/cache or
latency exists. [Budget](budget.json) must record explicit user numeric agreement
before the controller launches. This experiment constructs public preprocessing,
not database encoding or storage-side protocol computation.

## Scripts and validation

[Controller](run.py), [Slurm script](run.sbatch), [small controller tests](test_controller.py),
[completion analyzer](analyze.py), [actual preprocessing CLI](../../cmd/pir-preprocess/README.md).
[Exact sizes](../s17-preprocessing-admission/s17-sizes.json), [size calculator](../s17-preprocessing-admission/sizes.py)
and [count validation](../s17-preprocessing-admission/runs/counts-01/analysis.json)
bind K=3876566203,N=17179869184,m2,t5. Small fresh/root/final/CUDA-free/race/ASan/
memcheck workflows are independently verified, not full-size validation.
[Local check record](local-checks.json) links exact commands/source hashes and
[controller tests](controller-tests.log). Full-runtime behavior remains unmeasured.

```sh
python3 -B experiments/gpu-s17-full-preprocessing/run.py --check
python3 -B experiments/gpu-s17-full-preprocessing/test_controller.py
# ONLY after exact budget agreement is recorded in budget.json:
sbatch --output=experiments/gpu-s17-full-preprocessing/slurm-%j.log \
  experiments/gpu-s17-full-preprocessing/run.sbatch first-01
# A later explicit --resume reuses certified completed stages, never overwrites.
# Only after a completion ping:
python3 -B experiments/gpu-s17-full-preprocessing/analyze.py first-01
```

Results/logs/profiles/source hashes/budget identity go to `runs/<name>/` and
`.cache/liftrs/gpu-s17-full-preprocessing/`. The shared cache's `s17-m2-t5/` holds
atomic support/root/code stages. Earlier profiles, binaries and failed partial
stages remain. CLI cache locking rejects concurrent cooperating writers. No agent
monitoring/polling follows submission. The controller itself emits resource
heartbeats and preserves the latest actual worker phase/counter; unknown remaining
work is never an ETA. Host support ordering/Frobenius, normalization and reused-
root scans now emit completed counts per65536-element block with30s throttling,
without a new full pass or large array. GPU progress remains at existing kernel/
wave boundaries; a long or blocked kernel cannot report subkernel completion.
Instrumentation/GC overhead is unquantified and included in profiled timings.

## Full trial resource budget

| Resource | Proposal |
|---|---|
| Placement | One eligible exclusive node; exclude gnode03, no fixed node |
| CPU | All128 node cores allocated/billed;32 active host workers |
| Host RAM | 864 GiB request;800 GiB process-group RSS cap;720 GiB Go soft limit |
| Device | 4 H200s;112 GiB owned cap/card,448 GiB aggregate |
| New retained files | 512 GiB total cache + result/log guard |
| Local scratch | 16 GiB cap; initial32 GiB free required |
| Shared filesystem | Initial576 GiB free; live64 GiB free headroom guard |
| Wall cutoffs | Worker45 h,controller46 h,Slurm48 h; not runtime estimates |
| Profiles/checkpoints | Enabled and retained; immutable completed stage reuse |

An exclusive node prevents competing batch jobs consuming unaccounted host RAM;
it costs128 allocated cores even though only32 workers are selected. Node scheduler
RAM is about907 GiB, so request864 GiB leaves about43 GiB outside the request.
Controller requires832 GiB initially available and stops below32 GiB node available
headroom. Device H200 memory measured140.4 GiB/card;112 GiB cap leaves about28 GiB
outside it. Per-array host limit128 GiB/native128 MiB/code payload161 GiB are
separate from aggregate RSS. Memory-limit/GC cost if reached is unknown. This is
an admitted-budget first trial, not a proven bound on driver/GC/filesystem usage.
48 h is within the configured100 h partition limit; QoS/association/quota still
matter. Available shared filesystem space is not authorization or quota.

The GPU array model is3·8N=384 GiB (96/card) before metadata/staging. The initial
full command explicitly selects `--vector-layout local`, the faster measured
four-card large-component candidate; it adds eight transpose passes/FFT (~2 TiB
aggregate array traffic at full domain). Small whole-workflow cases do not settle
full speedup. Default library peer layout is unchanged and either explicit choice
is supported by the budget checker. No slower memory-saving path is forced.
Eight/sixteen GPUs are authorized options to evaluate, but current implementation
cannot span these four-GPU nodes; first trial uses the implemented four-card route.

## Exact file sizes and host lifetime audit

| Public artifact | Calculated size |
|---|---|
| Support | 28.883 GiB |
| Normalized root checkpoint | 14.441 GiB |
| Canonical P/spectrum cache | 142.441 GiB |
| Generator G | 49.559 GiB |
| Total successful payload | 235.324 GiB |

These are exact byte-count formulas rounded toGiB, not measured disk consumption.
A successful fresh run writes/hashes and reads back each once; logs, profiles,
metadata, allocator/filesystem overhead and retained failed stages add storage.
The512 GiB guard allows headroom/partial stages but does not allow unlimited
failed retries. Exact byte values/source identity are in s17-sizes.json.

[Discovery](../../encoding/liftrs-shared/preprocessing_core.go) uses a128 GiB
uint64 grid,~28.883 GiB good-index list,~28.883 GiB output support and two2 GiB
bitmaps that may coexist during validation:189.765 GiB principal model. The
serialized root/P has14.441 GiB and G49.559 GiB. Final support+P+G+spectrum are
220.883 GiB, with saved normalized root sharing P backing, not another copy.
The GPU product arrays are384 GiB independently of those host totals.

This is **not peak RSS**. [Newton products](../../encoding/liftrs-shared/orbit_polynomial.go)
retain the previous G and its squared operand while receiving a new result;
[overlap-add](../../encoding/liftrs-shared/orbit_product.go) can hold the accumulator
and a local convolution output together. Root product-tree intermediates, full
64 GiB identity output, earlier dead-but-uncollected arrays, runtime/native metadata,
CUDA staging and filesystem cache add memory. The proposed large request/RSS/GC
limits leave headroom around current models; no measured guarantee is claimed.
No large duplicate payload is introduced by this controller or streaming verifier.

Controller checks the published support K against the verified count during
execution, guards RSS/storage/free headroom/deadlines and uses previously tested
bounded TERM/KILL waits. It does not reset GPUs. Startup peer-copy/transpose probe
aborts promptly on device failure before large allocations. SIGTERM retains a
profile and completed checkpoints; a failed root tree before normalization is
not a reusable root checkpoint. Sampling itself is restarted if interrupted.

## Evidence boundaries

CLI checks exact dimension, support ordering/Frobenius closure, normalized root,
full polynomial identity and canonical write/readback. Large payload digests are
recorded by the CLI and immutable stage manifests. Exit does **not** silently
read235 GiB again: it hashes source/binaries/plugin, profiles, metadata and logs.
The default completion analyzer verifies those seals, stage binding/header/size
and successful worker verification events; independent post-completion large-file
rehashing must be separately submitted as a CPU-only job if required, labeled
outside preprocessing latency. Failed/unreaped workers are not successful runs.
Full output cache and timings must not be published as measurements until review.

## Full trial approval (SCALING-083)

After the explicit numeric full resource proposal and local/peer trade-off
comparison, user directs "ok let's do local then". Budget approval is recorded
in budget.json for exactly the proposed4H200112GiB/card, exclusive128cores/
32workers,864GiBhost/800GiBRSS/720GiBGo soft,512GiBretained/16GiBscratch and
45/46/48-hour worker/controller/scheduler cutoffs. Local is explicit, peer
library default remains unchanged. Existing historical preparation checks remain
intact; no-allocation checks are rerun against the approved manifest. Approval
and submission do not establish a completed s17 cache or measured peak/latency.
