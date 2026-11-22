# Owned root execution with retained native validation

Status: prepared; integrated GPU results pending. Parent revision0f57378.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [retained root manifest](root-references.json),
[readiness](readiness/). Producing revision/source hashes will be sealed by the
controller. Results use `runs/owned-01/`;reusable source,binaries,profiles and shards
stay in `.cache/liftrs/distributed-owned-roots/owned-01/` outside Git.

## What changes

[80962 passed](../distributed-pipeline-wide/report.md) the complete four-worker
s4/s6 cold/restart pipeline and bounded canonical field18 adapters. Its diagnostic
hierarchy still loaded complete native support/root references on every worker.
That validation scaffold cannot be the fullsize root executor.

The new [MergeRoots method](../../encoding/liftrs-shared/distributed_executor_cuda.go)
consumes only a validated merge stage,this worker's child coefficient shard and
checked operand transport. It shares the existing interval router and resident
three-FFT product,including canonical range,padding and nonaliasing checks.
The maintained tree schedule validates normalization/degree at each level.
No complete support array,native root oracle or gathered polynomial is required by
this method. Field degree and extension word type are independent of tree topology.

[One shared GPU adapter](../../encoding/liftrs-shared/pipeline_executor_gpu_test.go)
now configures workspace,operation identities,transpose/Taylor links and proof
logging for both subgroup roots and later full-domain Newton/identity/spectrum.
[Owned orchestration](../../encoding/liftrs-shared/pipeline_roots_gpu_test.go)
keeps independent sibling groups,closes their GPU/communicator before the next
level,and aborts the original allocation on failure. Actual roots come directly
from each worker's discovered closed-orbit support. Only local roots,child shards
and at most64 integer counts enter the reduction.

The complete pilot explicitly selects `PIR_PIPELINE_ROOT_MODE=owned`;its existing
`diagnostic` default remains available. This is an experimental validation option,
not a new production client memory mode or a changed default. Genuine retained
reference mode still streams global support to its writer without keeping a full
support archive on rank0. The new owned branch never calls the complete diagnostic
support loader. Standard client builds retain CPU support without CUDA/MPI.

## Validation keeps independent leaf/stage/final checks

The [manifest](root-references.json) retains64 native expected files,total112,460
bytes,with original job/revision/path/size/SHA-256 provenance. Genuine s4/s6 files
come from verified80962;synthetic field18 files come from verified79376. Those
producing jobs checked independent native leaves,subtrees,final polynomials and
scalar witnesses. Every selected file matches its producer's sealed cache digest.
The controller copies/checks these files before GPU work;workers do not load them
as arithmetic inputs. They are checked byte-for-byte against every newly saved
leaf,merge stage and final shard after each completed arm.

The reviewer reconstructs every expected group tree,stage dimensions,operation
identity and all three FFTs per merge. It checks transpose/Taylor proofs and
operand payload fingerprints using actual preceding-stage output bytes. Thus the
same independently verified native results cover the new executor without its
allocating complete native reference arrays. Missing stages,wrong source packets,
wrong group closure or altered outputs fail the allocation. Checksums detect
corruption;they are not authentication against malicious rewrites.

One admitted allocation combines four profiled arms:

| Arm | Workload and checks |
|---|---|
| cold | Genuine s4/s6 owned discovery/root hierarchy through Newton,full cyclic identity,spectrum,committed prefixes and canonical/generator publication. All native prefix/final bytes and root-stage files checked. |
| resume | Fresh four-worker MPI launch restores verified root/prefix8,skips discovery/root construction and reproduces remaining prefixes/artifacts. |
| wide | Existing bounded field18 reciprocal/product/spectrum and high-bit64-bit stream adapter regression. |
| owned-wide | The owned MergeRoots64-bit instantiation on synthetic16/4092 closed roots at s18;all leaves/stages/final shards match independently retained native files. |

The synthetic wide fixture is itself small,replicated test-input generation;this
is not fulls18 support discovery. Its executor receives only the owned local
polynomial. Genuine owned pipeline validation removes the complete diagnostic
support load;other fullsize restrictions remain. No fulls18 grid/artifact or
node-loss recovery claim follows from these cases.

## Resources and explicit memory–time boundaries

Same allocation: **two HPC nodes,two H200s,four CPU cores,8GiB total requested host
RAM**,four workers/two per card;**64MiB owned GPU/worker=128MiB/card aggregate**,
**1GiB RSS/worker**,zero CUDA-pinned buffers;**128MiB retained cache,2GiB scratch,
32MiB/file**;**720-second controller/15-minute scheduler**. Eligible nodes allowed,
gnode03 excluded. CUDA context/driver/MPI internals are additional. No resource
increase or fixed two-GPU design ceiling is implied.

The actual reduction no longer needs replicated O(K) support/native-root diagnostic
arrays;it retains owned support and current child/parent shards. Principal group
GPU workspace remains three extension arrays (12D/r bytes with32-bit slots,
24D/r with64-bit slots),plus field metadata,transfer/Taylor scratch. Host input,
two routed operand shards and result can overlap. Compact-to-full-domain expansion
still temporarily allocates its documented second identity-output shard;this
change does not solve every fullsize memory boundary.

Arithmetic remains one product/three FFTs per parent and the same bounded coefficient
redistribution;no extra production transform,gather or network phase is introduced.
Diagnostic native recomputation/readbacks move out of the worker workflow;controller
retained-file copy/verification is included in total job time. Native stage files
add112,460 retained bytes plus manifest,alongside69,746 existing native code inputs.
This is not a kernel speed benchmark:timing boundaries and diagnostic scratch/work
change,so do not label a shorter arm as an algorithmic speedup. Fullsize runtime
and validation persistence costs remain unmeasured. Faster validated production
paths/defaults are retained;no slower memory-saving strategy is silently adopted.

Estimate **3–6 minutes excluding queue**,based on80962's2min27.889s for three arms
and the already validated small field18 tree shapes. Per-launch220-second test /
240-second launcher and720-second controller bounds remain. All four arms use
normal profiling;no CUDA arithmetic changes or redundant sanitizer-only gate.
Prior field18/genuine/hierarchy clean memchecks remain. Cache projected below32MiB
under128MiB cap;proof logs below8MiB/rank/arm under32MiB/file,based on retained
same-size runs. Exact sizes and phase times are measured by the controller.

## Reproduction,preflight and limitations

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9,H200. Local all-tagged build initializes
neither MPI nor CUDA. Native shape preflight covers genuine supports and synthetic
field18 leaf counts/domains;focused race tests cover routing,root normalization,
checkpoint admission and stream widths. Reviewer regressions reject manifest path /
size / digest errors,corrupt retained bytes and omitted runtime-manifest packaging.
The new runtime manifest is explicitly admitted by the snapshot selector;other
historical JSON results remain excluded. Freeze exact committed source,verify its
required paths/reviewers and copy/check all retained native inputs before submission.
The admitted allocation repeats participating-node executable/library/input/output
path checks before expensive phases. Completed phases/profiles persist before later
work;existing timestamped phase,group,transform/pass/tile and resource logs remain.
Proof capture is included;isolated logging overhead is unmeasured.

```sh
python3 -B experiments/distributed-owned-roots/test_review.py
python3 -B experiments/distributed-owned-roots/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-owned-roots/slurm-%j.log \
 experiments/distributed-owned-roots/run.sbatch owned-01
# After completion notice:
python3 -B experiments/distributed-owned-roots/analyze.py owned-01
```

[Build](readiness/build.log), [native shapes](readiness/native.log),
[race](readiness/race.log), [reviewer](readiness/review.log).
GPU composition remains pending. Scalable complete field18 discovery/working-set
admission,full-domain generator/spectrum and final artifacts,larger representative
scaling/interruption/relaunch,measured headroom and separate fulls18 admission
remain. A successful small owned-root gate does not establish fulls18 readiness.
