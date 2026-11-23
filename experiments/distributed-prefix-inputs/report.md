# Newton prefix input views

Status: **passed81729**,prefix-01. Implementation **2f63e9d**;parent83b828d.
[Runner](run.py),[Slurm entry](run.sbatch),[analyzer](analyze.py),
[reviewer tests](test_review.py),[readiness](readiness/). Planned results:
`runs/prefix-01/`;reusable immutable source,binaries,native references,profiles and
artifacts remain outside Git at `.cache/liftrs/distributed-prefix-inputs/prefix-01/`.
Producing revision/source identities will be sealed by the controller.

## Scope and semantics

The [fullsize allocation audit](../distributed-fullsize-audit/report.md) found
that every Newton step copies the full root shard,then zero-masks coefficients
beyond that step's prefix. The new explicit `PIR_PIPELINE_NEWTON_INPUT=prefix`
uses an immutable logical prefix of the existing root instead. Existing product
reset zeroes both GPU input slots;only selected source intersections are uploaded.
Unused source tails are logically omitted and may contain the later root
coefficients. This is distinct from dense input,whose entire padding must be zero.

The [pure range helper](../../encoding/liftrs-shared/distributed_prefix_input.go)
validates intervals without allocating data. The
[executor](../../encoding/liftrs-shared/distributed_executor_cuda.go) shares one
product/reset/FFT/inverse implementation for both modes. Active values retain
base-field checks;zero-reset tails remain zero. Source arrays are not modified.
There is no second product recurrence and no changed field representation,
nonaliasing bound,packet identity or canonical artifact format.

The current **dense default remains unchanged**. Prefix views are explicit and
limited here to Newton inputs;root merges,full identity,spectrum/discovery retain
their original interfaces. Other optional memory modes remain independently
selected. Complete client CPU/GPU encoding is unaffected by this server prototype.

## Memory and time trade-off

Avoid one4N/r-byte host root clone per Newton step. At s18 with16illustrative
workers that is16 GiB/worker. Before upload,the named root/old-generator/squared/
clone inventory is64 GiB/worker;prefix input removes the clone,leaving48 GiB in
that inventory. This is **not measured peak RSS**:result allocation occurs later,
compiler/GC lifetime may differ and serialization/native/runtime/other arrays
remain. Sixteen workers are a capacity example,not supported/admitted topology.

Prefix input also skips host→GPU uploads of the logically zero tail. For s18's
known K/schedule,the [model](../distributed-fullsize-audit/results/projections.json)
projects36 TiB dense Newton input uploads versus1,394.893 GiB prefix uploads
across workers. This26.43-fold input-traffic reduction is **not** an end-to-end
speedup:full-domain FFTs,resets,Taylor/communication,readback and storage stay.
No extra transform,network payload or retained output bytes are intended.
View checks and changed transfer ordering have unmeasured fullsize latency;
do not adopt a new default from a small single timing sample.

For genuine s8,N65,536/K7,381,16updates:prefix mode predicts **0 clone words** and
**161,392 uploaded extension words aggregate**,versus a16,384-word root clone
per worker per step and **2,097,152 aggregate uploaded words** dense. At4-byte
slots that is645,568bytes versus8 MiB of Newton input uploads. Actual counters
increment only after successful GPU transfer and are checked per rank using
independent interval formulas;no unrelated identity/spectrum uploads are counted.

## Validation and resources

Four arms reuse the maintained native/support/root/prefix/identity/canonical/MPI
reviewers and fixed seed109:

1. Genuine s8/curve5 prefix-input cold run.
2. Fresh-process prefix8 restart with prefix inputs.
3. Matched dense-input cold run;discovery-result/spectrum streaming and compact
   checkpoint flags stay identical in both matched cold arms.
4. Existing bounded field18 tail at dimensions5/12,using prefix inputs for Newton,
   dense full identity,independent native/reciprocal references and high-bit
   GPU source publication. This validates actual64-bit input lifting/upload.

Require exact outputs and source immutability,correct step/clone/upload counts,
existing dense zero-padding rejection,full polynomial/nonaliasing identity checks,
all packet proofs and verified restart. Pure tests cover partial/empty/uneven
ownership,metadata-only wide offsets,invalid extent/mode and reconstruction of
an explicit masked array without changing its source. CPU/race/native checks,
13Python reviewer suites,audit tests and81454saved-evidence replay pass. A local
reviewer fixture initially expected132rather than195rank-local uploaded words;
manual interval enumeration caught it before allocation. Its initial log is
retained separately;the final fixture verifies both contributing square prefixes.
No CUDA arithmetic changed and no sanitizer-only arm is repeated.

Same admitted HPC allocation:**two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**,four workers/two per card,**64 MiB owned GPU/worker=128 MiB/card**,
**1 GiB RSS guard/worker**,zero CUDA-pinned buffers;**128 MiB retained cache/2 GiB
scratch/32 MiB per file**,**720-second controller/15-minute Slurm deadline**,
eligible nodes excludinggnode03. Context/driver/MPI internals are additional to
owned bytes. No increased budgets or fixed GPU-count design ceiling.

Estimate **3–6 minutes excluding queue**,based on81454's2min50.037s three-arm
controller and81365's2min48.928s four-arm controller. Added bounded wide work
previously took5.793s;prefix/dense change does not add transforms. Existing
220-second test/240-second launcher limits have headroom. Cache projected below
48 MiB,rank logs below12 MiB/file,inside caps based on retained prior evidence.
These are projections until measured;no fullsize runtime estimate. Completed
arm evidence/profiles persist before later work;no useful earlier cache is deleted.

Timestamped/flushed group/sample/prefix/pass/tile/resource logs remain. Arm timing
includes setup/transforms,transfer/communication,clone/mask or view handling,
hash/sync/checkpoint/reload,canonical/diagnostic publication and proof logging.
Controller adds native export,builds,CPU/race checks,independent review and profile
processing. Actual input-word counters are common instrumentation in both modes;
isolated instrumentation overhead is unmeasured. Fixed order/one sample is not
statistical speedup evidence. Use complete-arm costs when comparing modes.

## Reproduction and remaining gates

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200. Verify actual committed source,
required helper/wrapper/manifests and native inputs before submission. Every
participating worker checks executable/library/input paths and cwd before GPU
phases. Native reference preflight initializes neither CUDA nor MPI.

```sh
python3 -B experiments/distributed-prefix-inputs/test_review.py
python3 -B experiments/distributed-prefix-inputs/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-prefix-inputs/slurm-%j.log \
 experiments/distributed-prefix-inputs/run.sbatch prefix-01
# After registered completion notice:
python3 -B experiments/distributed-prefix-inputs/analyze.py prefix-01
```

After validation,update the maintained reference with actual clone/upload counts,
memory,latency and exact outputs. The audit identifies separate production gates:
parameterized rank/lifecycle/non-oracle execution,real local-count/subtree memory
admission,bounded diagnostic/proof/progress recording and complete persistence/
restart policy. Current small-case harness and proof logs cannot simply be scaled
to fulls18. Fulls18 remains unrun and needs separate topology/resource/storage/
headroom/deadline agreement.


[Committed-source preflight](readiness/snapshot-checks.json) passed on2f63e9d:
**845 build inputs /3,856,765 bytes**,required prefix/executor/pilot/wide/
wrapper/manifest paths and13reviewers plus audit tests checked from frozen source.
All383code hashes match the successfully compiled tagged readiness tree;binary
hash is recorded. Frozen preflight verified108014historical native bytes and
actual s8/curve5 input/root geometry. Fresh frozen-source build is inside allocation
before GPU phases;second potentially>30s local rebuild omitted transparently.
[Native preflight](readiness/snapshot-native.log),[CPU](readiness/cpu.log),
[race](readiness/race.log),[native tests](readiness/native-tests.log).
No local CUDA/MPI initialization.


Submitted **81729** after readiness75524c2. [Receipt](submission.json),
[scheduler log](slurm-81729.log),results `runs/prefix-01/`. Four arms:prefix cold,
fresh-process prefix8 restart,matched dense cold,and bounded field18 prefix/native
checks. Existing two-node/two-H200 caps retained;estimate3–6min excluding queue.
Receipt-bound local completion notification registered outside Git;no model polling.


## Verified result:81729

Independent reconciliation passed **845 source/96 evidence/496 cache identities**,
producing75524c2. All four arms passed exact native root/prefix/identity/spectrum/
canonical/restart checks and actual clone/upload counter comparisons.

| Arm | Profiled wall time | Peak RSS per worker | Owned GPU per worker |
|---|---|---|---|
| Prefix-input cold | 11.328589 seconds | 377.789 MiB | 0.244 MiB |
| Prefix-input fresh-prefix8 restart | 8.994542 seconds | 379.551 MiB | 0.244 MiB |
| Matched dense-input cold | 11.652888 seconds | 380.027 MiB | 0.244 MiB |
| Bounded field18 prefix/native/source | 6.687571 seconds | 379.930 MiB | 3.125 MiB |

Prefix cold uploaded **161,392 extension words** aggregate versus **2,097,152**
dense,a12.994-fold reduction in Newton input words. Prefix removed the
**16,384-word root clone per worker per step**;at s8 this is64 KiB. Restart
prefix uploads161,367words;the two wide shapes uploaded233words together,with
no root clone. Dense default remains unchanged. In one fixed-order sample,prefix
cold was **0.324299 seconds (2.783%) faster**;this is not a statistical/fullsize
speedup claim. Small RSS differences do not measure the projected large saving.

HPC gnode16/17,two H200s,four workers/four CPU cores,8 GiB requested host RAM;
all prior caps retained,context/driver additional to owned bytes. Controller
**3 min11.051 s**,cache **37.509 MiB**.
[Analysis](runs/prefix-01/analysis.json),[raw results](runs/prefix-01/results.json),
[scheduler log](slurm-81729.log). Next audit production integration begins with
bounded proof/native progress recording,preserving every runtime verification
and resource guard while avoiding billions of stored per-packet/window rows.
Fulls18 remains unrun/unadmitted.
