# Actual GPU-assisted s16 preparation

User chose an actual guarded run rather than additional synthetic size sweeps.
[Runner](run.py), [analyzer](analyze.py), [launcher](run.sbatch) and [worker](../../encoding/liftrs-shared/gpu_full_preparation_test.go)
reuse exact saved support and the verified resident transform/product backend.
Parent published checkpoint:412eb29. Each manifest records exact revision, dirty
source hashes/snapshots, hardware/toolchain, commands and artifact hashes.
Results are pending. Do not describe full s16 preprocessing as demonstrated.

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
