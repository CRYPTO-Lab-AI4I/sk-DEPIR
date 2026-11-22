# Bounded discovery-result consumption

Status: **passed81454**,results-01. Implementation **1d8e000**;parent0ec3410.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/). Planned results:
`runs/results-01/`; immutable source, native references, binaries/profiles and
reusable artifacts remain outside Git at
`.cache/liftrs/distributed-discovery-results/results-01/`.
Exact implementation/producing revisions are recorded before submission.

## Why this allocation is next

[GPU spectrum publication81365 passed](../distributed-spectrum-publication/report.md#verified-result81365):
source cold10.713813s versus resident10.346982s in one matched sample (3.545%slower),
with identical canonical output and explicit bounded source ownership. Resident
remains default. Spectrum output no longer requires a full host result when
source mode is selected,but discovery still returns a complete Word[N/r] array
for every sample,then scans it into the local support union bitmap.

The new explicit `PIR_PIPELINE_DISCOVERY_RESULT=stream` leaves discovery output
in GPU slot0 and reads it once through a bounded4096-word tile. The existing
[executor](../../encoding/liftrs-shared/distributed_executor_cuda.go) shares the
same tensor/inverse transforms,field arithmetic,operation IDs and direct
resident-result downloader. Only result ownership differs. The optional tile
consumer updates the same bitmap,with ordered/exact extents and field-range
checks;each sample counts only newly nonzero positions. No sample clears prior
membership. Invalid tiles change neither bitmap nor counters. Exact completion
is required before accepting sample counts.

[Union implementation](../../encoding/liftrs-shared/discovery_union.go),
[ordinary CPU tests](../../encoding/liftrs-shared/discovery_union_test.go),
[assembled discovery](../../encoding/liftrs-shared/distributed_discovery_gpu_test.go).
Keep random byte generation,global-good order,sampling width,stopping rule,
root ownership/routing,full identity and canonical representation unchanged.
Resident discovery remains the default and keeps its original full result and
scan. Streamed results require the distributed discovery layout;an incompatible
selection fails before GPU work. Source mode is an explicit experimental
command-line environment selection,not a slower production default.

## Memory and time policy

Avoid one full discovery-result allocation of N/r extension words per sample,
replacing it with at most4096 words (16 KiB narrow/32 KiB wide). At s16 this is
16 GiB aggregate;at s17,128 GiB;at s18,512 GiB (32 GiB/worker for an illustrative
16 workers);s19/20 exploratory allocations are2/8 TiB aggregate. These are
allocation calculations,not measured total RSS savings or16-worker support/
admission. GPU workspace,base input grid,good-monomial list,bitmap,raw support
and root arrays remain additional. Repeated resident results may also extend
lifetimes through garbage collection;no numerical GC saving is claimed.

Discovery and spectrum occur in distinct phases. Their avoided allocations
**must not be summed** as a simultaneous peak-RSS saving. Complete fullsize
working-set/lifetime,root imbalance,storage and topology admission remain.

No extra transform,network message,device→host result bytes or storage is intended.
All result words are downloaded once per sample. Optional streaming adds callback,
order/range checking and interleaves bitmap work with transfer;fullsize latency
and overlap effects are unknown. Keep the faster validated resident default when
memory permits. Selecting this slower-or-unknown-cost option remains explicit.
The matched experiment isolates discovery mode by streaming spectrum output in
both cold arms;the earlier3.545%spectrum sample is not this option's penalty.

## Validation and resources

Three genuine s8/curve5 arms reuse the maintained complete pipeline/native
reference and root-stage reviewers:

1. Tile-discovery cold run through support,owned roots,Newton,full polynomial
   identity,spectrum and canonical artifacts.
2. Fresh-process prefix8 restart from that cold cache (no discovery repeated).
3. Matched resident-discovery cold run with the same streamed-spectrum mode.

N65,536,K7,381,seed109,two native-reference samples,known unequal root counts
[3442,1507,1541,895]. Independent native code/support/generator/spectrum and every
leaf/merge/final reference remain. Require matching random entropy,per-sample
counts,stopping sample,full support,canonical artifacts,completed-prefix/restart
checks and all MPI packet proofs. Record source/result mode,actual resident and
buffer words,per-sample download count and aggregate sample downloads. Tile mode
must have0resident result words and≤4096buffer words. The native canonical output
must match across cold modes;fixed arm order/one sample is not statistical
speedup evidence.

Ordinary CPU/race tests cover both widths (including nonzero bit35),persistent
membership across samples,new-versus-duplicate counts,partial/repeated/out-of-order/
empty/oversized/out-of-field tiles and bitmap admission. Existing sampling and
publication tests remain. Reuse81365's verified64-bit GPU source/transfer evidence;
no new CUDA arithmetic,wide GPU array shape or sanitizer-only arm. Fullfield18
GPU discovery remains unrun;bounded uint64 union tests do not certify it.
Twelve Python reviewer suites plus complete81365 saved-evidence replay must pass.
Freeze committed source and verify required helpers/native inputs before allocation.

Same admitted HPC resources:**two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**,four workers/two per card,**64 MiB owned GPU/worker=128 MiB/card**,
**1 GiB RSS guard/worker**,zero CUDA-pinned buffers;**128 MiB retained cache/2 GiB
scratch/32 MiB per file**,**720-second controller/15-minute scheduler deadline**,
eligible nodes excludinggnode03. Driver/context/MPI internals are additional to
owned bytes. No increased budget or fixed GPU-count design limit.

Estimate **3–6 minutes excluding queue**,based on81365's2min48.928s for four arms
and81286's2min22.513s for two arms. New work has three arms,unchanged GPU transforms,
one matched cold control and extra small union CPU/race tests. Per-arm220-second
test/240-second launcher deadlines retain headroom. Cache projected below48 MiB,
rank logs below12 MiB/file based on prior37.413MiB cache/≤7.77 MB logs. These are
projections,not measured costs for the new mode. Existing guards stay active;
completed profiles/results persist before later work. No useful cache is deleted.

Timestamped/flushed sample/group/prefix/pass/tile/resource logs remain. Tile
consumption emits completed/total result words;job-side observation throttles
long-phase updates to approximately30 seconds,with a final completion event.
Controller/native build logging remains. Arm time includes setup/transforms,
communication,union/reference work,checkpoint/hash/sync/reload,canonical/diagnostic
publication and proof logging;controller adds native export,builds,review and
profile processing. Isolated logging overhead/fullsize throughput are unmeasured.

## Reproduction and remaining work

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200. Shared executable/library/wrapper/
input/cwd checks run on every worker before GPU phases. Native CPU references
initialize neither MPI nor CUDA and stay retained with hashes.

```sh
python3 -B experiments/distributed-discovery-results/test_review.py
python3 -B experiments/distributed-discovery-results/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-discovery-results/slurm-%j.log \
 experiments/distributed-discovery-results/run.sbatch results-01
# After completion notice:
python3 -B experiments/distributed-discovery-results/analyze.py results-01
```

After completion,update the maintained reference with actual counters,latency,
RSS/device memory and exact outputs. Next audit remaining live discovery/root/
reciprocal arrays,ownership bounds and production persistence/logging rather than
extrapolating these small cases into fulls18 support. Fulls18 requires concrete
working-set,topology/storage/headroom/deadline agreement and remains unrun.


[Committed-source preflight](readiness/snapshot-checks.json) passed on1d8e000:
**837 build inputs /3,837,521 bytes**,required union/executor/pilot/
wrapper/native-manifest paths and12reviewer suites checked from frozen source.
All **381 code input hashes** match the successfully compiled tagged readiness
binary's build tree;binary hash is recorded. Frozen preflight helper checked
108014 historical native bytes and actual s8/curve5 support/root geometry.
Fresh snapshot build runs inside admitted allocation before GPU phases;second
potentially>30s local rebuild omitted. [Native preflight](readiness/snapshot-native.log),
[CPU](readiness/cpu.log),[race](readiness/race.log),[native tests](readiness/native-tests.log).
No local CUDA/MPI initialization.


Submitted **81454** after readiness77b9c8a. [Receipt](submission.json),
[scheduler log](slurm-81454.log),results `runs/results-01/`. Three arms:tile-discovery
cold,fresh-process prefix8 restart,and matched resident-discovery cold;spectrum
is streamed in both matched cold arms. Same resources/caps;estimate3–6min excluding
queue. Receipt-bound local completion notification registered outside Git;no model
polling. Analyze only after terminal notice.


## Verified result:81454

Independent reconciliation passed **837 source/79 evidence/402 cache identities**,
producing77b9c8a. All three genuine s8/curve5 arms passed exact entropy,sample
counts/stopping,support/root/identity/spectrum/canonical artifact and fresh-process
restart comparisons. Two discovery samples/worker each read16,384words exactly
once;tile mode retained0full result words and4096buffer words versus16,384resident
result words. No extra transforms,output or download words.

| Arm | Profiled wall time | Peak host RSS/worker | Owned GPU/worker |
|---|---|---|---|
| Tiled discovery cold | 11.256600 seconds | 377.820 MiB | 0.244 MiB |
| Fresh-process prefix8 restart | 7.952252 seconds | 378.926 MiB | 0.244 MiB |
| Resident discovery cold | 10.994534 seconds | 377.977 MiB | 0.244 MiB |

Spectrum output was streamed in both matched cold arms. Tiled discovery was
**0.262066 seconds (2.384%) slower** in this one fixed-order sample. This does
not establish a statistical or fullsize runtime penalty. Small RSS differences
do not measure the projected large allocation saving. Resident default remains;
tiled discovery is explicit. Controller runtime **2 min50.037 s**;cache
**37.154 MiB**. HPC gnode16/17,two H200s,four workers/four CPU cores,8 GiB requested
host RAM;all previous caps retained,CUDA context/driver extra to owned bytes.

[Analysis](runs/results-01/analysis.json),[raw results](runs/results-01/results.json),
[scheduler log](slurm-81454.log). Next source-based fullsize audit includes live
root/Newton arrays,root imbalance,serialization buffers,storage/logging and
production parameter/restart limits. One concrete candidate is the full Newton
root clone used only to zero-mask its prefix:4N/r bytes per step,16 GiB/worker
at s18/16workers. Prototype an explicit prefix view against the existing zeroed
GPU workspace,without another product implementation. Fulls18 remains unrun.
