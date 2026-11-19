# Diagnose the failed root-transfer boundary

Status: **diagnostic job77782 submitted**, results unreviewed (SCALING-123).
Implementation `8da6205`. [Submission receipt](submission.json), results/profiles
`runs/capture-01/`, progress `slurm-77782.log`.
[Frozen-input check](readiness/snapshot.json):640files/3,034,320bytes, all8 checked
input hashes match; no GPU/MPI execution or repair claim from local readiness.
The original [77704 failure](../root-merge-tree/report.md#failed-validation-review-scaling-122)
is unresolved: the sender and receiver logged different hashes for one131,073-word
polynomial before final CPU equality failed. No resource guard triggered. This
experiment captures that boundary; it is not a production fix or a staging workaround.

[Runner](run.py), [scheduler script](run.sbatch), [reviewer](analyze.py),
[shared tree runner](../root-merge-tree/run.py),
[opt-in test harness](../../encoding/liftrs-shared/root_tree_gpu_test.go),
[fingerprint tests](../../encoding/liftrs-shared/root_transfer_proof_test.go),
[diagnostic MPI control exchange](../distributed-root/mpi/world.go).

The normal tree path remains unchanged unless the explicit test environment
`PIR_ROOT_TREE_DIAGNOSTICS=1` is set. Each endpoint fingerprints coefficients via
the existing bounded little-endian codec. A sender fingerprints before and after
the synchronous send; the receiver fingerprints twice before GPU use. They exchange
both fingerprints in a fixed64-byte MPI control record (separate tag30) and check
local stability plus cross-peer equality. This diagnostic control is not an
adversarial authentication protocol. Proof records include case/sample/stride,
peer, hashes, current buffer address and current-data hash; the arrays contain
public synthetic preprocessing data, not client records.

On any disagreement, both endpoints save their current complete payload and proof
metadata. A second control exchange occurs only after each capture is flushed/
fsynced/closed, so the normal MPI abort cannot kill one side before it saves its
file. Source pre/post fingerprints distinguish observed source changes from
stable-but-different received bytes; they do not alone establish the cause.
Snapshot/file hashes and logged array hashes may differ if memory changes during
capture, which is itself evidence to inspect. Original CPU/full-shape/range and
final-root checks are retained. A clean transfer is passed to the GPU only after
the proof comparison. No C-owned/pinned staging buffer or changed transport
backend is adopted without diagnosis and an explicit memory/time comparison.

Use the same genuine s=8 and synthetic s=17 fixtures. Smaller cases keep their
original two timed samples; only the previously failing262,144-exponent case
repeats12 timed samples plus warm-up and final artifact pass. Normal chunk65,536
words and memcheck1,024 words cover boundaries in both index/coefficient widths.
The largest warm-up additionally retains its support/local-root arrays to bind
later failure captures. Existing artifact snapshots and all four profiles remain.
If normal execution fails, memcheck/profile summaries may again be unrun; the
primary goal is durable paired failure data, not a green complete-run status.

Extra instrumentation uses bounded64 KiB codec/hash buffers, additional CPU hash
passes, tiny control messages and failure-only snapshots. It adds latency and
changes allocation/timing conditions; an unreproduced fault is **not a repaired
fault**. No throughput baseline or production setting follows from diagnostic
medians. The same Go-owned send/receive arrays remain in use. No new full-array
staging copy, transform or production default is introduced. Diagnostic payload
snapshots are bounded by the admitted small cases (largest mismatch524,292bytes
per endpoint), not full-size checkpoint persistence.

Resource envelope stays unchanged:2HPC nodes/2H200 total,4 logical ranks sharing
2cards,1 CPU/rank,2CPU/4 GiB requested RAM per node;64 MiB owned/rank≤128 MiB/card,
1 GiB process RSS guard/rank,128 MiB retained cache/2 GiB scratch/32 MiB per-file,
15-minute scheduler/720-second controller cutoff. Tests220s/step240s/command260s
remain bounded. More repeats fit within the same deadline; no resource increase.
No GPU/node-count scaling claim is possible from shared cards. Driver/context
memory is additional to owned arrays; existing host/device guards remain active.

```sh
python3 -B experiments/root-merge-tree/test_review.py
python3 -B experiments/root-transfer-diagnostics/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/root-transfer-diagnostics/slurm-%j.log \
  experiments/root-transfer-diagnostics/run.sbatch capture-01
# After user completion ping only:
python3 -B experiments/root-transfer-diagnostics/analyze.py capture-01
```

The runner reuses exact frozen Git build inputs and Go1.27.3/GCC14.2/CUDA12.9/
OpenMPI4.1.8/Slurm25.05 setup. Results/profiles/logs go to `runs/capture-01/` and
`slurm-JOBID.log`; binary/library/fixture/shard/transfer/failure arrays remain
outside Git at `.cache/liftrs/root-transfer-diagnostics/capture-01/`, with source/
evidence/cache seals in results.json. Regenerate with the same revision and a
fresh name; no old preprocessing/cache/profile is removed. The reviewer labels
outcomes captured/not-reproduced/failed-elsewhere and keeps the original gate
unresolved even if all diagnostic checks pass.

Local readiness: tagged native/MPI/CUDA binary compiles;
[fingerprint symmetry/tail tests](readiness/proof-tests.log) pass, including the
131,073-word boundary and canonical codec equality. Snapshot-comparison test
identifies a changed/missing final word; existing3tree/4distributed/5staged/3shared
review tests pass. [Commands/source hashes](readiness/checks.json),
[build log](readiness/build.log). No MPI/GPU execution or diagnosis is claimed by
these local checks; compiler-wrapper sandbox socket warnings exit zero.
