# Committed final workflow output

Status: **passed job82100**, publication-01; all three arms independently verified. Implementation **1d6d89d** (parent a7eebbe).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/).
[Results](runs/publication-01/analysis.json), [raw results](runs/publication-01/results.json), [Slurm log](slurm-82100.log); immutable source/binaries/reusable data
outside Git: `.cache/liftrs/distributed-publication-commit/publication-01/`.
Implementation/producing revisions are recorded before submission.

## Recovered workflow validation

[82099's executed arms passed corrected independent review](../distributed-workflow/report.md#recovered-validation82099).
The original job remains failed because its controller expected a test-only native
reference event from a runtime that intentionally has no reference input. The
reviewer now uses separately verified native metadata explicitly;default pilot
behavior still requires the original event. Every entropy/count/route check remains.
No arithmetic or GPU rerun was needed. Assembled retained profiles were processed
post-completion, separately from the original sealed evidence.

## Final file commit boundary

[Publication helper](../../encoding/liftrs-shared/distributed_publication_files.go)
wraps the existing code/generator/raw-support writers. Exclusive reservation and
`.partial` names prevent accidental overwrite/retry under an exclusively owned
directory/label. Successful writes update exact byte counters and SHA256 states.
No full payload reread or new full-size serialization buffer is introduced.
Commit requires a successful runtime identity result, matching canonical code header,
exact expected file sizes and complete writer counts, then fsyncs/closes payloads,
renames them, syncs the directory, and writes/syncs/renames a manifest last followed
by another directory sync. Consumers accept only a complete manifest with matching
payloads,not just visible filenames. The files are not individually published as
an atomic group;the manifest is the logical commit marker.

The manifest binds canonical code header, sample count, identity-result status,
file names/sizes/hashes and explicit `uint64-exponents-le` raw archive format. That
raw support archive is **not** the library's canonical support-cache format. The
outer experiment receipt/source seals bind producing code provenance;manifest
checksums alone do not authenticate an untrusted producer or prove an identity.
Later readers must verify actual file hashes;streaming hashes assume immutable
caller-owned storage and are not a second readback corruption test.

Failed/short writes poison the writer. Partial files/reservations remain for
investigation;no useful output is deleted. Retry requires another directory/label,
not silent cleanup or overwrite. Payload rename/manifest failures may leave some
final payload names without a marker. If the last directory sync fails after
manifest rename,the API still reports failure even if a complete marker is visible;
a caller must not advertise success without resolving that failure. No automatic
resume/recovery guarantee follows from this helper.

The bounded workflow test explicitly selects `PIR_WORKFLOW_PUBLICATION=commit`.
Unset/`direct` preserves its prior direct-file observer policy. This is a test/job
option,not a new public CLI. Runtime math still sees only caller-owned writers;
no checkpoint/atomic file policy is forced into client encoding or numerical kernels.
Durable distributed prefix restart and production command/resource admission remain.
Full s18 preprocessing is unrun/unadmitted.

## Memory, I/O and runtime cost

Three small streaming SHA256 states and counters replace no numerical arrays;the
code header is bounded metadata. No extra transforms/device transfers or payload
copies/retained payload bytes;renames reuse existing staged files. Extra storage is
a small JSON manifest and empty exclusive reservation. A full s18-sized extra
payload reread would be about889.446 GiB with the raw archive;the streaming design
avoids that additional I/O, but **hash CPU cost remains**.

Compared with the prior assembled observer,the same three payload fsyncs remain,
plus one manifest fsync,two directory fsyncs and four renames. Commit overhead and
large-file SHA throughput are unmeasured. This is an explicit publication-integrity
boundary,not a free speed/memory optimization. Direct observer policy remains
available;do not infer a fullsize peak RSS or runtime from this small case. The
experiment controller's final evidence/cache seal adds its own read/hash work and
is included in controller timing,not treated as production-free persistence.

## Validation

Pure CPU tests cover exact bytes/hashes/metadata,hidden staged output,failed identity,
short/incomplete extents,poisoned writer,repeat close/commit and no overwrite. Tests
exercise persistence semantics only;they do not claim their dummy payload has a
valid polynomial identity. The integration computes full identity in runtime and
compares native bytes independently. Reviewer tests reject altered committed bytes
and require explicit commit selection in the real suite.

One same-budget allocation retains the fully reviewed pilot cold/restart baseline
and runs the assembled runtime with committed output. Corrected independent
review checks all76checkpoint/support/identity comparison files,ordered proof
counts/digests/bytes,native roots,Taylor counts,entropy/sample/support ownership
and final native/pilot code/generator/archive bytes. It additionally checks the
manifest against actual files and the saved commit event and requires no staged
payload remains. No unchanged wide/dense/sanitizer comparison is added.

Pilot rows precede deferred GPU close. Assembled runtime rows include runtime close,
observer publication commit and its syncs;launcher includes shutdown. Manifest
policy differs from pilot prefix manifests/native observers,so timings are not a
direct speedup comparison. Controller includes builds/native export,tests,review,
profile processing and evidence sealing. Saved logs retain timestamped phases,
actual completed work and resource observations. Completed data/profiles persist.

## Resources and reproduction

Same admitted HPC allocation: **two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**,four workers/two per card; **64 MiB owned GPU/worker=128 MiB/card**,
**1 GiB RSS/worker**,zero CUDA-pinned; **128 MiB retained cache/2 GiB scratch/32 MiB
per file**, **720-second controller/15-minute Slurm limit**,eligible nodes excluding
gnode03. Context/driver/MPI memory extra. Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200;
ordinary CPU client builds remain CUDA/MPI-free.

Estimate **3–6 minutes excluding queue**,based on82099's completed commands and
prior three-arm controllers;cache below64 MiB/logs below12 MiB/rank projected. No
resource increase. Exact frozen source/native preflight precedes submission;fresh
builds and required executable/library/wrapper/input/cwd checks on every node
precede GPU phases. No rerun solely for82099's Python correction.

```sh
python3 -B experiments/distributed-publication-commit/test_review.py
python3 -B experiments/distributed-publication-commit/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-publication-commit/slurm-%j.log \
 experiments/distributed-publication-commit/run.sbatch publication-01
# After terminal notice:
python3 -B experiments/distributed-publication-commit/analyze.py publication-01
```

After validation integrate production coordinated prefix checkpoint/restart and
parameterized command/full phase admission,then representative larger topology and
complete fulls18 resource agreement. Fulls18 remains unrun/unadmitted.

### Immutable-source readiness

Implementation **1d6d89d** is frozen into **908 build inputs**;
all **402** Go/C/C++/CUDA hashes match compiled readiness source.
All **23 frozen reviewer suites** pass. CPU-only native preflight reproduces
**108014 historical reference bytes** plus actual s8/curve5 inputs/root geometry.
[Source checks](readiness/source-checks.json), [reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). No local CUDA/MPI initialization.
Fresh committed-source build and each-node path/environment preflight remain in the job.

### Submission82100

Job **82100** submitted from **d53213f**, implementation **1d6d89d**.
[Receipt](submission.json);results `runs/publication-01/`,Slurm log `slurm-82100.log`.
Unchanged resources above;estimate3–6minutes excluding queue. No polling or GPU result claim.

### Verified outcome82100

Producing revision **d53213f**; receipt and all **908 source /81 evidence /371 cache**
seals reconcile. Controller **2 min38.073 s**; retained logical cache **36.220 MiB**.
HPC allocation as above, four workers on two H200s.

| Arm | Launcher runtime | Peak host RSS per worker | Outcome |
|---|---|---|---|
| Pilot cold | 8.907 s | 380.117 MiB | Passed |
| Pilot fresh-process restart | 6.374 s | 379.695 MiB | Passed |
| Assembled committed publication | 8.510 s | 378.824 MiB | Passed |

Owned device allocation about **0.244 MiB/worker**, excluding context/driver/MPI.
All76intermediate comparison files, ordered proof hashes/counts, native roots and
Taylor work, discovery entropy/count/ownership and final canonical bytes match.
The saved manifest/event and three actual payload hashes/extents match, with no
staged payload left. The exclusive reservation is retained intentionally.
Timings have the different boundaries described above; this is correctness and
persistence evidence, not a speedup estimate or fullsize peak-memory measurement.
Next: [durable assembled prefix commits](../distributed-workflow-checkpoints/report.md),
then arithmetic resume/parameterized command and complete phase/resource admission.
