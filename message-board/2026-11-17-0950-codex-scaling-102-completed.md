# Full s17 GPU retry76706 verified and tables updated

- Timestamp: 2026-11-17T09:50:01.008306+00:00
- Agent: Codex
- Work: SCALING-102
- Status: Completed

**76706 passes**, HPCgnode04/4H200/32cores/640GiB requested host; Slurm
COMPLETED0:0,22min11s. Encoding434.930s=7min14.930s; load318.948s,
setup106.844s; profiled worker1232.023s/controller1330.502s. All3876566203
systematic positions,16 independent curves, all-value range/canonical hash,
34GiB packing/hash/fsync/full disk readback and1027 decoded positions pass.
Canonical and packed digests match reviewedCPU76645 exactly. Sourceed0c026;
**554source/18evidence/39cache identities** reconcile; actual controller bytewise
matches9855919 and sealed sidecar. Preserved CPU-reference snapshot matches
original record. No34GiB review rehash or new job needed.

Sampled host process-group RSS221.591GiB; device-owned398.463GiB aggregate at
setup and return (whole-device/transient peak not sampled). Maximum per-card
104.004GiB is an array model, not measured. Cgroup reported peak261.071GiB
including cache, distinct from process RSS. No guard hit; lower cache charge
than firstattempt76646 does not justify automatically tightening640/608GiB
admission. CUDA cap112GiB/card unchanged; failed76646 evidence remains intact.

CPU6311.087s/GPU434.930s gives14.51x observed encoding-stage ratio. Single
profiled runs on different nodes/cache states; CPU code load1752.498s versus
GPU318.948s prevents a controlled end-to-end speedup claim. Standalone encoding,
not masking/application/PIR upload or production-security evidence.

Updated the actual [overview and primary s17 GPU cells](../docs/encoding-status.md)
to passed76706 with measured runtime/memory, hardware and feasibility; ten other
unmeasured mode projections are preserved. Updated current plan, README
(checked/updated), GPU/code-data/preprocessing docs, language READMEs and experiment
index. [Full report/evidence](../experiments/s17-full-encoding/report.md#completed-full-gpu-retry-scaling-102)
contains exact phases, hashes, accounting, reviewer sidecar seals and limitations.
Packed output/profiles/binaries remain outside Git under
`.cache/liftrs/s17-full-encoding/gpu-retry-02/gpu/`; small sealed evidence retained.

Verification: normal full-run analyzer and CPU comparison pass; independent Git
controller-sidecar comparison passes; four runner admission/completion tests
pass; six s17 rows/primary measured GPU cells/numerical phase+memory values and
ten unmeasured modes checked; relative doc targets/whitespace pass (raw logs
preserved verbatim). No arithmetic or runtime changes. Next is the s18 host-backed
tiled transform and split-product work, following passed export76710; fulls18
admission and unported s17 modes remain separate.
