# Actual GPU-assisted s16 preparation

User chose an actual guarded run rather than additional synthetic size sweeps.
[Runner](run.py), [analyzer](analyze.py), [launcher](run.sbatch) and [worker](../../encoding/liftrs-shared/gpu_full_preparation_test.go)
reuse exact saved support and the verified resident transform/product backend.
Parent published checkpoint:412eb29. Each manifest records exact revision, dirty
source hashes/snapshots, hardware/toolchain, commands and artifact hashes.
Actual preparation now passes as recorded below; s16 database encoding remains unverified.

```sh
module load slurm/slurm/25.05
sbatch --cpus-per-task=4 --mem=16G --time=00:15:00 --output=experiments/gpu-preparation/slurm-%j.log experiments/gpu-preparation/run.sbatch validation-03 --validate
# After successful validation with identical runtime sources:
sbatch --output=experiments/gpu-preparation/slurm-%j.log experiments/gpu-preparation/run.sbatch actual-01 --admission validation-03
```

## Integration and correctness

The ordinary API/default CPU path remains intact. The private first-use harness
admits preparation-only execution and an optional normalized root checkpoint.
Full native setup remains available for transforms below4096 values. Resident
GPU convolutions handle eligible root carries, reciprocal and identity products;
full GPU transforms handle the generator spectrum. Host small products, root
arithmetic, overlap-add assembly and persistence remain Go. No device Go pointers
are retained. GPU state closes before native online work (validation only).

Validation uses s10,m2,t5: complete preparation plus CPU online systematic/curve/
packed-readback checks; then a fresh preparation uses the saved normalized root.
Support, G and online-file hashes must match the retained CPU reference from
products-02; replay G/online hashes must match the initial run. Full Go/native
suites, focused race/vet and direct checkpoint validation tests run first.
This is integration validation, not another large synthetic scaling experiment.

The actual job verifies and symlinks the original read-only support/basis metadata
from `.cache/liftrs/hpc-s16/s16-m2-t5`. Support SHA256:
`021446aa0d2764f4020e989ee54f3a99c63c7d9a407f5babbcb333919a2854ad`.
K918,672,421; N2^32. No rediscovery or original-file overwrite. A locked new
`.cache/liftrs/gpu-preparation/s16-m2-t5` owns outputs. Provenance rejects reuse
under changed runtime sources. It saves normalized P atomically as root.bin plus
metadata binding field/spec/K/N/support digest/content hash. Replay checks sizes,
hashes, field range and normalization. Every completed preparation still verifies
the full P*G=1+X^(N-1) identity before saving public constants. A trusted checkpoint
is not independently authenticated code data or a substitute for that check.

After preparation, stream online.bin (P/spectrum) and g.bin, retain hashes and
profiles, release CPU/GPU state and stop. No database encoding or automatic cached
encoding is launched. Existing reusable support/root/complete constants survive a
process restart. There is no mid-root-batch checkpoint: stopping before normalized
P is saved requires repeating the root stage. Reciprocal/spectrum work is repeated
if stopped after P but before complete constants. Atomic metadata marks completed
stages; partial .tmp files are never treated as reusable constants.

## Actual-run resource review

Per-job limits:one GPU/32 CPU cores,256 GiB host request,224 GiB sampled process
RSS guard,128 GiB Go soft target,64 GiB explicit device allocations,64 GiB new
cache guard,128 GiB scratch guard,24-hour allocation,23.5-hour controller and
23-hour worker timeout. Require320 GiB initially available node RAM,32 GiB while
running,80 GiB free shared cache storage,128 GiB scratch, and at least65 GiB free
on the allocated GPU before starting. These are admission limits, not forecasts.
An allocated device unable to meet them is rejected. Driver/context overhead is
outside explicit allocation accounting. CUDA remains an optional experimental
dependency; measured hardware was H200 NVL, not a consumer-GPU demonstration.

Full-domain GPU a/b/transpose buffers total48 GiB. Releasing old arrays before
growth avoids doubling that payload; tables/plans are small. Retained native
transpose scratch adds16 GiB of host memory. Support is6.85 GiB as Go ints;
P≈3.42 GiB, G≈12.58 GiB and spectrum16 GiB. During large products, immutable
inputs, the host result/prefix and assembly output coexist; Newton iterations and
GC may retain earlier arrays. The224 GiB guard provides room above these payloads;
no sum is claimed as measured process RSS. Preparation-only execution avoids the
additional online/database buffers. Root checkpoint≈3.42 GiB plus final online
file≈19.42 GiB and G≈12.58 GiB fit the64 GiB new-cache guard; original support is
symlinked and excluded from new-cache usage. Model sizes are not measurements.

The controller samples process-group RSS, node available RAM and cache/scratch
usage. Slurm's request alone is not a demonstrated hard cgroup memory bound.
TERM is forwarded to the worker group so CPU profiles can flush; KILL follows
only after the grace period. Native instruction profiling is unavailable under
host permissions; retain Go CPU/heap profiles and CUDA stage/wall timings.

The earlier roughly2.1x early-root improvement and larger transform gains justify
an attempt, not an end-to-end runtime promise. The24-hour limit retains the prior
large-job envelope. Final clocks will distinguish cached support loading,
reusable setup, root work, reciprocal/identity, spectrum and persistence; diagnostic
GC/heap profiling remains included. No full-run speedup is claimed against the
interrupted historical CPU attempt. All source/binary/cache/profile evidence is
retained; large constants stay outside Git. Do not monitor after handing off the
long run; wait for the user's completion ping.


Initial integration attempt [validation-01](runs/validation-01/results.json),
job72871, stopped in the Go repository suite before preparation: two PIR
placeholder tests lacked the legacy RM binary. Lift-RS packages passed. The
runner now builds that binary first, matching the established validator; fresh
validation-02 is the corrected attempt. Failed logs/source evidence are retained.


[Corrected validation-02](runs/validation-02/results.json), job72872, and its
[verified analysis](runs/validation-02/analysis.json) pass: complete CGO-disabled
Go and CUDA/native suites, legacy build, focused race/vet, complete s10 execution
and preparation-only root replay. Support/G/online files match retained CPU
reference hashes; replay G/online files match the first run. All283 source hashes
and sealed evidence verify. Full small preparation takes1.433 s and root replay
0.668 s, with GPU setup0.720/0.402 s separately; these are instrumented integration
checks, not full-s16 forecasts. GPU resources close to zero after each run.
Runtime sources must exactly match this validation before actual admission.


Final diff review caught a test filename collision after validation-02: the new
root-checkpoint helpers had replaced the existing diagnostic test file. Restore
that file exactly and keep the new tests in preparation_root_checkpoint_test.go.
Fresh validation-03 is required before actual admission. Historical source
snapshots preserve the exact validation-02 contents. Raw zero-sample pprof logs
retain their hashes under a targeted whitespace attribute; no log bytes changed.


[Final validation-03](runs/validation-03/results.json), job72873, and
[analysis](runs/validation-03/analysis.json) pass after restoring the original
diagnostic and separating the root-checkpoint tests.284 exact source hashes and
sealed evidence verify. Full Go/native suites, focused race/vet and s10 reference/
replay checks all pass. The actual admission uses validation-03 and exactly the
same runtime sources. Staged whitespace checking now passes with the targeted
raw-profile-log attribute; earlier measurement bytes remain unchanged.


## Actual submission and handoff

Job72875, actual-01, was submitted from published revisiond9a80f0 with
`--admission validation-03`. [Live manifest](runs/actual-01/results.json) and
[Slurm log](slurm-72875.log) record startup and eventual outcome. At the initial handoff, startup
was running and full preparation was not yet verified. The controller requires
identical validated runtime hashes and the frozen support digest before entering
the worker. No synthetic size sweep or database encoding is scheduled. Leave the
job unmonitored after handoff; when the user reports completion, run
`python3 -B experiments/gpu-preparation/analyze.py actual-01`, inspect any guard/
error outcome and update this report before making completion claims.


## Completed actual preparation

[Actual-01 manifest](runs/actual-01/results.json), job72875 at revisiond9a80f0,
passes with no guard stop. [Final accounting](runs/actual-01/accounting.log)
records COMPLETED/0:0 and **2h24m15s** allocation elapsed. Environment is recorded
in its hostname/CPU/GPU/compiler/toolkit logs; the run uses the H200 NVL/Xeon6710E
setup. The phase clocks are profiled wall timings, including diagnostic GC and
root checkpoint persistence, and exclude the original support discovery:

| Phase | Measured wall time |
|---|---:|
| Native/GPU executor setup | 0.849 s |
| Load and verify saved support (labeled discovery) | 25.702 s |
| Complete preparation, including identity verification | 2h20m15.469s |
| Stream final online/G files | 46.021 s |
| Release executors and host state | 1.147 s |

The worker process takes8491.095 s (2h21m31s); its phase clocks exclude profile
start/stop and some wrapper costs. Allocation elapsed additionally includes
controller startup, verified support input, build, profile summaries and cache
hashing. This is faster than the conversational3–5-hour rough estimate. It is
not a matched CPU/GPU whole-run speedup: the historical CPU attempt was stopped,
and its discovery/setup boundary differs.

Exact K918,672,421 and N4,294,967,296; full P*G=1+X^(N-1) identity verified.
The normalized root checkpoint and complete P/G/spectrum files were saved.
Sampled process-group RSS peaks at77,926,809,600 bytes (**72.575 GiB**), below the
224 GiB guard; Slurm accounting records a different sampling boundary. New cache
occupies38,043,010,613 bytes (**35.430 GiB**), excluding symlinked original support.
No runtime explicit-device peak was sampled;48 GiB remains the three-array payload
model, bounded by the64 GiB engine cap. Owned device state is zero after close.

There are223,322 resident product calls and one standalone spectrum transform.
Transfer-inclusive GPU calls total3745.148 s (62.419 minutes), including event and
call overhead. This is part of preparation wall time, not a separate latency or
GPU utilization percentage. Retained [Go profile summary](runs/actual-01/pprof-preparation.pprof.log)
shows substantial remaining root arithmetic/small-product work as well as Cgo
calls; it cannot identify GPU/native instruction hotspots.

Reusable files under `.cache/liftrs/gpu-preparation/s16-m2-t5`:

| File | Bytes | Recorded SHA256 |
|---|---:|---|
| root.bin | 3,674,689,688 | ec7c60c6e0f6349bf1a81d110012ec24ce330a33ac86b20991571941c11c6a79 |
| g.bin | 13,505,179,500 | 488d9d65d67b1b77a658071cab3349e00c0cff9f9c72802049d2bf1b20ff7cc0 |
| online.bin | 20,854,558,905 | f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874 |

The controller computed final cache digests before marking passed. The lightweight
[reconciliation script](reconcile.sh), run as
`bash experiments/gpu-preparation/reconcile.sh actual-01`, verifies284 exact source
hashes,42 sealed evidence files, small cached files, recorded metadata/digest
agreement and actual artifact sizes. Its [result](runs/actual-01/reconciliation.json)
explicitly does **not** rehash the three large payload files. The existing
[analyzer](analyze.py) can perform that bulk recheck; treat it as a nontrivial job
and hand it off if expected to exceed30 seconds. Existing evidence bytes remain
unchanged. No additional hardware job was submitted during reconciliation.

This demonstrates complete experimental s16 preparation on the measured machine.
It does not validate a fresh cached s16 database encoding, full application/
cryptographic execution or TRL5 acceptance. Next is one separately admitted cached
encoding with saved online.bin; it must check every systematic position, independent
curve samples and packed readback, retain profiling and obey the user's job handoff
rule. No encoding was launched automatically.


## Rough estimate for the next cached encoding

The current [online path](../../encoding/liftrs-shared/scaling_matrix_cpp.go)
executes the whole database-dependent encoding in C++; its transforms and scalar
work remain CPU. The optional GPU backend above accelerates preparation. A GPU
online encoding would require separate integration and correctness checks, so its
preparation gains are not applied to this estimate.

The same-host retained [width-02 s15 analysis](../hpc-cached-s15/runs/width-02/analysis.json)
measures705.864 s encoding on16 workers, versus655.937 s in the older baseline.
For s15, three prefix transforms have2^29 values and two full transforms2^30;
s16 needs2^31/2^32 respectively. All domains grow4x. A rough N*log2(N)^2 scaling
scenario gives705.864*4*(32/30)^2≈3212 s (53.5 minutes), excluding loading and
validation. This is a planning model, not a measured cost law. Field/basis changes,
allocation, memory bandwidth, larger scatter output and worker scaling introduce
uncertainty; do not assume doubling cores halves runtime.

Use **45–75 minutes for encoding alone**, roughly **1–1.5 hours for a fresh cached
run including loading, correctness checks and packed-output persistence**. Plan
around one hour for the kernel. The broader run estimate reflects width-02's
179.2 s s15 validation and its growth with message/output size plus I/O variability.
It reuses prepared code, excludes discovery/preparation and queue wait, and covers
one field-element database, not an entire application layout or cryptographic
encoding. No new timing job was run and no GPU-online speedup is claimed.
