# Four-worker distributed GPU collective validation

Status: **submitted as79239**, `workers-01`, implementation `139d75e`.
[Receipt](submission.json), [scheduler log](slurm-79239.log). GPU results pending;
actual producing revision/source identities are recorded by the controller.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/checks.json).
Expected results `runs/workers-01/`, reusable snapshots/shards/profiles in
`.cache/liftrs/distributed-four-worker/workers-01/` outside Git.

The genuine two-rank coordinator-memory/restart pipeline passed79213, and bounded
canonical s18 arithmetic passed78917. Next validate collective communication with
more than one peer per worker, before larger rank-owned root/discovery integration.
This experiment uses **four MPI workers sharing two H200s**, not four GPUs.

## Shared implementation and checks

Use the same s18 full-extension FFT and exact base-field product drivers, with
`PIR_S18_FIELD_RANKS=4`. Empty/2 preserves two workers; other values reject.
The shared test lifecycle admits only2/4 workers. Resident allocations receive
64 MiB/worker for this run; ordinary two-worker limits remain128 MiB/card.
Go/C++ arithmetic recurrences and CUDA kernels are unchanged. Client builds/API
and defaults remain unchanged.

The existing runner now accepts an explicit worker count, reusing the validated
four-worker launch flags from the root-tree experiment:2 tasks/node,1CPU/task,
one GPU/node shared with `--gpu-bind=single:2 --gres-flags=allow-task-sharing`.
Every worker checks executable/wrapper/library/output paths before arithmetic.

Existing output/proof reviewers accept an explicit rank count. Every remote
transpose tile includes the peer in its key for four-worker runs, and each
send/receive round and reciprocal proof is checked. Taylor windows independently
intersect target/source ownership across all four workers. No nonlocal peer can
be omitted or aliased with another worker's identical offset. Native/scalar full
FFT and schoolbook references check all output bytes and base-field padding.
Direct pointwise checks retain high36-bit field components, invalid/alias/layout
rejection and source/third-slot preservation. Equal partitions are explicitly
admitted for these bounded fixtures; this reviewer does not claim uneven-rank
coverage. Generic CPU orchestration separately covers uneven/empty ownership.

One allocation, four arms:

| Arm | Values | Checks |
|---|---|---|
| FFT normal | 32 and4096 | All four forward/inverse shards; independent native/scalar oracle |
| Product normal | 32 and4096 | Two shapes,three FFTs/product,full-extension pointwise,exact schoolbook |
| FFT memcheck | 32 | Same checks plus CUDA errors/leaks |
| Product memcheck | 32 | Same checks plus CUDA errors/leaks |

Small sanitizers are warranted by a new rank-local GPU layout; unchanged two-worker
micro-gates are not rerun as separate allocations. Keep existing source/runtime
seals,CPU profiles and timestamped phase/pass/peer progress. Completed arms persist
before later phases. Analyze only after the user's completion ping.

## Unchanged hardware budget and trade-offs

**Two HPC nodes /two H200s /four CPU cores /8 GiB requested host RAM**, with two
workers sharing one H200 and4 GiB requested RAM per node. **64 MiB owned GPU/worker
=128 MiB aggregate/card**, **1 GiB RSS/worker**,zero pinned buffers,
**128 MiB retained cache /2 GiB scratch /32 MiB file**, **720-second controller /
15-minute Slurm**. Exclude gnode03; other eligible nodes allowed. Same physical
resources as the preceding two-worker job and historical four-worker78261.

Per-worker main arrays halve at four workers; aggregate payload stays constant.
At4096 values,three8-byte arrays occupy24 KiB/worker versus48 KiB at two workers.
Lookup tables duplicate:about3 MiB/worker,6 MiB/card with two workers,plus plans,
transfer tiles and CUDA contexts. Owned caps include all explicit engine payload;
CUDA context/driver memory is additional. Four1 GiB RSS guards sum to4 GiB across
two nodes,below the8 GiB requested RAM,with headroom for controller/builds.

Additional CUDA contexts,CPU/native reference work and extra peer routes can slow
this sharing topology; runtime effect is **unknown**. The test explicitly selects
it for correctness. It is not a default,capacity optimization or four-GPU speedup
claim. Cache/log size projects below64 MiB within existing128 MiB cap. Prior suites
complete around2–4 minutes; estimate **3–6 minutes excluding queue**, with existing
240-second launcher/220-second test cutoffs and generous sanitizer headroom for
only32 values. Whole-controller time includes builds,profiling,setup,oracles and
I/O; diagnostic kernel timings retain prior capture/proof boundaries. No fullsize
or stable throughput extrapolation is justified.

## Preflight and reproduction

Focused tests cover four-worker exact output/peer/round coverage and missing
transpose/Taylor endpoints; old two-worker reviewers still pass. Launcher tests
assert4tasks/2pernode/1CPU each/sharedGPU flags and unchanged aggregate caps.
Native-only preflight checks all two-/four-worker oracle shards without GPU/MPI
initialization. Tagged driver compiles locally; CUDA source is unchanged. Before
submission create the real committed-source snapshot and run the new reviewer
suite from it,checking required script/wrapper paths. No new runtime JSON input
is required. Ordinary builds remain CUDA/MPI-free.

```sh
python3 -B experiments/distributed-four-worker/test_review.py
python3 -B experiments/distributed-four-worker/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-four-worker/slurm-%j.log \
 experiments/distributed-four-worker/run.sbatch workers-01
# After completion ping:
python3 -B experiments/distributed-four-worker/analyze.py workers-01
```

Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200. Preserve local compiler socket
warnings separately from successful builds. Full multi-level distributed roots,
four-worker genuine discovery/generator/restart integration,more physical GPUs/
nodes and fulls18 remain separate remaining work. No new experiment resource
admission is inferred from capacity estimates.


[Committed snapshot check](readiness/snapshot-checks.json) passed:790 files /
3,633,152 bytes,required driver/wrapper/runner files present,and new reviewer
suite passed from frozen source. No local GPU/MPI initialization. Submission
revision372cf19 adds only this evidence to implementation139d75e.
