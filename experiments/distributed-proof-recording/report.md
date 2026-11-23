# Bounded protocol proof and native window recording

Status: **passed job 81738**, all six recording-01 arms independently verified. Implementation revision **08f1100** (parent 8a9b915).
[Runner](run.py),[Slurm entry](run.sbatch),[analyzer](analyze.py),
[reviewer tests](test_review.py),[readiness](readiness/). Planned results:
`runs/recording-01/`;immutable source,binaries,native references,profiles and
reusable data remain outside Git under
`.cache/liftrs/distributed-proof-recording/recording-01/`.
Exact implementation/producing revisions are recorded before submission.

## Production boundary and preserved checks

[81729 passed](../distributed-prefix-inputs/report.md#verified-result81729) all four
prefix/dense/native/restart arms. Prefix cold uploaded161392words versus2097152
dense and avoided the16384-word root clone per worker/step. Its single matched
11.328589s cold sample versus11.652888s dense is not fullsize speedup evidence.
The [fullsize audit](../distributed-fullsize-audit/report.md) identified another
hard boundary:current s18 main transposes alone could emit440401920proof rows per
worker,and native Taylor launch logging also scales per window.

The [shared recorder](../../encoding/liftrs-shared/distributed_proof_log.go) changes
stored evidence only. Every MPI descriptor,payload pre/post stability and peer
fingerprint check,field check,full identity and canonical comparison remains.
The existing per-record RSS guard still executes before recording or hiding
any proof row. No random bytes,arithmetic,transport order or representation changes.
Digests describe public evidence consistency,not authentication or protocol proofs.

Explicit `PIR_PIPELINE_PROOF_LOG` choices:

- `full` (default/unset):existing full rows;no recorder hash work or footer.
- `trace`:full rows plus counts/digests,for independently reviewed small baselines.
- `summary`:no stored per-packet rows for the fixed five recognized proof kinds;
  one count/semantic-byte/SHA256 record per kind/rank,plus approximately30-second
  completed-record progress. Unknown event kinds and normal stage/error/resource
  events remain visible. Totals are marked unknown while work is running.

State has at most five SHA256/count entries,independent of packet count. Each
current proof row is serialized transiently,not accumulated. Source rows exclude
only rank/time/resource annotations;rank is bound by its separate summary. Go
JSON semantic bytes are length-framed with an8-byte little-endian size before
hashing,in order. Numeric values remain exact beyond2^53. The independent Python
codec preserves nested packet field order and Go string escaping. Reordering,
truncation,changed descriptors/data hashes,counts or byte lengths rejects comparison.
A summary after failure is incomplete and cannot establish success.

## Native progress

Explicit `LIFTRS_GPU_TAYLOR_PROGRESS=bounded` suppresses tiny per-window start/
completion rows and emits observed completed window/coefficient totals at engine
close. Parent Go logs retain transform/pass/target completion and approximately
30-second advancing updates;native errors remain visible. It does not disable
`LIFTRS_CPU_PROGRESS`. Ordinary/unset native progress behavior stays unchanged.
Other native phases retain their existing progress policy;this does not claim all
remaining production logging/persistence/rank admission is complete.

Native window/coefficient totals are independently derived from actual Taylor
plans and target ownership intervals,including subgroup domains and quotient/
remainder windows. Small native spectra and high-bit publication still use the
previous independently checked field18 path. No CUDA arithmetic changed.

## Validation design and evidence boundaries

One allocation contains three full-trace/summary pairs (six short arms):

1. Genuine s8/curve5 cold trace and summary execution.
2. Fresh-process prefix8 restart trace and summary,both from the verified cold cache.
3. Existing bounded field18 reciprocal/product/spectrum/high-bit trace and summary.

Both modes use the same prefix inputs,streamed discovery/spectrum,compact
checkpoints and bounded native Taylor progress. Trace mode is independently
reviewed using every original per-packet route/payload oracle. Summary mode must
match every per-kind ordered digest/count/byte total and every canonical artifact.
For this **bounded validation only**,the controller reconstructs the already
verified baseline proof rows to feed the existing full reviewers;production
summary recording neither stores nor requires that baseline list. Actual
non-proof metrics,root stages,process IDs,artifacts and runtime checks come from
the summary run. Do not describe summary files as standalone full traces.

Pure CPU/race tests cover default behavior,ordering,volatile metadata exclusion,
serialization failures,incomplete status and fixed state size. Go/Python golden
fixtures preserve uint64 integers above2^53. Reviewer regressions reject count,
digest,completion,mode and leaked full rows. Actual retained81729/81454cold/restart
proof streams have identical normalized ordered digests,confirming replay
stability before this allocation. Existing native/canonical/source counters and
14Python reviewer suites remain. Native window totals are separately checked in
stderr;all native failure records reject the arm. No new sanitizer-only arm is
needed for this recording-only CUDA change;previous clean arithmetic coverage is
retained,not claimed as new coverage for the recording composition.

## Runtime,memory,I/O and resources

Runtime impact is **unmeasured**. Summary mode removes proof-row writes but still
serializes and hashes each proof and preserves the per-proof RSS syscall. Trace
mode hashes and writes;thus the paired comparison isolates writes,not the cost
against original full/no-hash mode. Fixed arm order and one sample are not a
statistical/fullsize speedup study. Bounded native counters are common to both
arms;their latency relative to old native logging is not isolated. Default/full
behavior stays available. No slower memory strategy is silently adopted.

Recorder state is bounded;GPU workspace and protocol payloads are unchanged.
Only stored diagnostics shrink. Useful existing full traces,profiles and caches
are retained;there is no deletion or verification sampling. Normal progress/
plan records are not suppressed and their fullsize bounds still need review.

Same admitted HPC allocation:**two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**,four workers/two per card,**64 MiB owned GPU/worker=128 MiB/card**,
**1 GiB RSS guard/worker**,zero CUDA-pinned buffers;**128 MiB retained cache/2 GiB
scratch/32 MiB per file**,**720-second controller/15-minute Slurm deadline**,
eligible nodes excludinggnode03. Driver/context/MPI internals are additional to
owned bytes. No increased resources or fixed GPU-count design ceiling.

Estimate **4–7 minutes excluding queue**,based on81729's3min11.051s for four arms.
Six arms add roughly20s prior measured arm latency plus profile review and hashing
cost. Existing220-second test/240-second launcher cutoffs and720-second controller
retain headroom. Cache projected below64 MiB;full-trace rank logs below12 MiB/file
based on prior37.509MiB cache/≤7.77MB logs. These are projections until measured.
Summary output is additionally bounded per proof kind;normal plan/stage/native
rows remain. No fullsize runtime estimate follows from this allocation.

All phases flush timestamps,work/pass/group IDs and resource observations. CPU
profiles/completed arm results persist before later work. Arm timing includes
recording/setup/transforms,communication,checkpoint/hash/sync/reload,canonical/
diagnostic publication and correctness checks. Controller adds native export,
builds,CPU/race tests,independent digest/native/packet review and profile processing.
Isolated logging overhead remains unmeasured until these comparisons are inspected.

## Reproduction and next work

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200. Go and CUDA builds pass without
local GPU execution. Required source/wrapper/native manifests and executable/
library/input/cwd paths must be checked on every participating node before GPU
phases. Exact committed source/native compatibility preflight precedes submission.

```sh
python3 -B experiments/distributed-proof-recording/test_review.py
python3 -B experiments/distributed-proof-recording/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-proof-recording/slurm-%j.log \
 experiments/distributed-proof-recording/run.sbatch recording-01
# After registered completion notice:
python3 -B experiments/distributed-proof-recording/analyze.py recording-01
```

After completion,record actual log sizes,state/counts,native totals,latency,memory
and artifact comparisons in the maintained reference. Continue production
parameter/rank/lifecycle and non-oracle execution,canonical persistence/restart
ownership and per-phase real count/domain/resource admission. This does not admit
or implement the full s18 workflow;fulls18 remains unrun.

### Committed-source readiness

Implementation **08f1100** was frozen into 851 build inputs; all 383 Go/C/C++/CUDA
source hashes match the locally compiled readiness tree. All 14 frozen Python
reviewer suites passed. The tagged native binary reproduced **108014 historical
reference bytes** and exported/validated the actual s8/curve5 inputs and owned-root
references. See [source checks](readiness/source-checks.json),
[frozen reviewers](readiness/frozen-reviewers.json) and
[native preflight](readiness/native-preflight.log). CUDA and tagged Go builds passed;
no local GPU/MPI work was executed. The admitted allocation rebuilds committed
source and checks paths/environment on every participating node before GPU phases.
Raw retained MPI/tool outputs are preserved verbatim, including whitespace;
authored source/documentation whitespace checks passed.

### Submission

Job **81738** submitted from **cd681de** with unchanged resources above.
[Receipt](submission.json); results will be `runs/recording-01/`, Slurm log
`slurm-81738.log`. No scheduler polling or result claim follows submission.

### Verified result81738

All six arms passed independent source/evidence/cache reconciliation: **851 source
files, 130 evidence files and 692 cache files**. Producing revision **cd681de**;
[receipt](submission.json), [analysis](runs/recording-01/analysis.json),
[raw results](runs/recording-01/results.json), [Slurm log](slurm-81738.log).
HPC gnode13/gnode14, two H200s, four MPI workers/four CPU cores and 8 GiB requested
host RAM total, with the unchanged admitted guards described above.

| Arm | Result/runtime | Peak host RSS | Aggregate rank stdout | Proof records checked |
|---|---|---|---|---|
| cold | Passed; 9.496726 seconds | 378.316 MiB/worker | 22.693 MiB | 31,574 |
| resume | Passed; 6.765242 seconds | 378.070 MiB/worker | 12.224 MiB | 17,090 |
| summary-cold | Passed; 9.415148 seconds | 378.484 MiB/worker | 2.792 MiB | 31,574 |
| summary-resume | Passed; 6.813566 seconds | 378.051 MiB/worker | 1.927 MiB | 17,090 |
| wide | Passed; 4.874992 seconds | 379.629 MiB/worker | 25.383 MiB | 40,012 |
| summary-wide | Passed; 4.944900 seconds | 380.543 MiB/worker | 1.061 MiB | 40,012 |

Cold summary replaces 31,574 per-packet rows with 16 summaries; restart replaces
17,090 with 12; wide replaces 40,012 with 12. Ordered digests, counts, semantic byte
lengths, all native Taylor window/coefficient totals and every output comparison
match. Runtime MPI/field/full-identity/canonical checks remained enabled.
Owned GPU peaks are 0.244 MiB/worker for s8 and 3.125 MiB/worker for bounded field18;
CUDA/MPI/context memory is additional. Controller runtime **3 min 19.204 seconds**;
retained cache **40.750 MiB** (exact inventory is sealed in raw results).

One paired cold summary was faster, while restart and wide summary were slower.
These single, fixed-order samples do not establish a runtime benefit or penalty
at full size; trace includes hashing that default full mode omits. Summary remains
explicit. Stdout totals include other unsuppressed stage/plan/metric events; native
stderr also remains, so these are not total retained-storage figures. No full s18
feasibility or runtime follows from this small validation.

Next: bound diagnostic serialization without changing bytes, durability or
verification; then continue production ownership/lifecycle/non-oracle entry,
bounded persistence/restart and real-count resource admission.
