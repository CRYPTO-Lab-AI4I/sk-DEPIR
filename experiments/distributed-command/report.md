# Parameterized distributed preprocessing entry

Status: **prepared; command GPU validation not yet run**. SCALING-193 follows
[passed assembled restart82102](../distributed-workflow-resume/report.md#verified-outcome82102).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/).
Results `runs/command-01/`; source/binaries/reusable caches outside Git:
`.cache/liftrs/distributed-command/command-01/`. Producing revisions are recorded
before submission. The exact command/configurations are retained with results/cache.

## Implementation and admission boundary

[Command/usage](../../cmd/pir-preprocess-mpi/README.md),
[public experimental policy](../../encoding/liftrs-shared/distributed_job.go) and
[CUDA/MPI runtime entry](../../encoding/liftrs-shared/distributed_job_cuda_mpi.go).
One explicit JSON configuration selects code/rank parameters, absolute shared
paths, GPU ordinal, word/byte budgets, deadline and all memory/log modes.
Parsing rejects unknown/trailing/oversized JSON and invalid ownership/modes/caps.
All ranks agree the complete typed policy before computation. The existing
reference-independent cold/restarted numerical paths are reused unchanged.
No native oracle or testing framework is required by the executable.

The command opens MPI on its owning OS thread; failures abort peers. Fresh per-rank
CPU profiles are mandatory. Local watchdogs sample RSS/storage/deadline every second;
SIGTERM/resource failure saves the profile and exits nonzero. A fail-fast launcher
is required. Profile/log storage is separately admitted by the controller. Sampling
is not an instantaneous hard bound; Slurm and per-file rlimits remain necessary.
Directory scans tolerate atomic rename races and reject symlinks in owned output.
Every node preflights the binary/wrapper/library/output paths before expensive work.

Once exact K is counted, a callback admits final payload sizes plus all retained
compact checkpoints/metadata and both logical archive names **before GPU discovery
allocation**, and repeats before root/Newton work. Existing local list/shard/rounded
root-stage/owned CUDA limits remain. This is a concrete storage/shape admission,
**not a complete fullsize host/GC/native/MPI memory model**. Complete phase headroom,
larger topology and fulls18 resources remain the next gates. Source admits s≤18;
small-case validation does not establish fullsize implementation feasibility.

The old single-node command and ordinary CPU library remain CUDA/MPI-free. Without
explicit native/CUDA/MPI build tags, the new command supports config checks and
returns an unavailable error for execution. Memory modes are explicit; no slower
memory policy replaces validated faster defaults. Cold command uses crypto/rand;
`--test-seed=109` is an explicit reproducible validation-only selection. Recovery
requires trusted source metadata hash and unchanged code/rank count. Same-filesystem
archive aliases avoid copying a verified support payload; source remains immutable.

## Runtime, storage and measurement changes

The command saves final code/generator/raw archive, source records and completed
compact checkpoints. It intentionally omits the test harness's per-rank discovery
support, root-leaf/stage and identity diagnostic copies; full live verification
and proof recording remain. This reduces diagnostic I/O/storage, not arithmetic.
No extra transforms, numerical arrays or device transfers are intended. New costs
are config agreement, scalar admission and one-second directory/RSS observations;
fullsize latency is unknown. Logical guards count hard-link names separately.
Profiles/logs remain outside payload storage; controller seals still read all
retained evidence and cache bytes. Do not present these as free production I/O.

## Validation

CPU/race tests cover strict JSON, scalar/device/shard/storage/file checks and
watchdog RSS/storage/file/deadline failures. Tagged builds produce the actual
command and a separate native-reference test binary without GPU/MPI initialization.
Python tests exercise real command-arm wiring; shared launcher supports an explicit
command/wrapper and symbolizes profiles with that executable, preserving its
existing test-binary default. New wrapper supplies `--profile`, never Go test flags.

One admitted job compares reviewed pilot cold/restart with the actual command
cold/fresh-prefix8 restart at s8/curve5,seed109,N65536,K7381. Independent reviewers
require identical ordered arithmetic proof digests/counts/bytes and native Taylor
work, all compact root/Newton payloads and final canonical code/generator/support.
All source/commit/shard/publication hashes and alias identity remain checked;
restart must hash support once on the coordinator. No unchanged wide/dense/sanitizer
repeat. Command paths use no diagnostic oracle files as inputs.

Command rows include runtime close/checkpoint/publication sync; pilot rows precede
deferred close. Launcher includes profile flush/shutdown; controller includes
builds, reference export, CPU tests, review/profile processing and all seals.
Persistence/logging boundaries differ, so no direct speedup claim. Timestamped
phases, actual completed work and bounded native progress logs remain enabled.

## Resources and reproduction

Unchanged admitted HPC allocation: **2 nodes /2 H200 /4 CPU cores /8 GiB requested
host RAM total**,4workers/two per card; **64 MiB owned GPU/worker =128 MiB/card**,
**1 GiB RSS/worker**,zero CUDA-pinned; **128 MiB total retained cache /2 GiB scratch /
32 MiB per file**, **720-second controller /15-minute Slurm cutoff**; eligible nodes
except gnode03. Command output storage cap32MiB/arm is within aggregate128MiB;
profile/log retention remains controller-owned. Context/driver/MPI extra.
Go1.27.3,GCC14.2,OpenMPI4.1.8,CUDA12.9/H200.

Estimated controller **3–7 minutes excluding queue**, based on82102's measured
191.398seconds plus command build/preflight/watchdog work; projected cache below
96MiB. No budget increase. Exact frozen-source/native preflight precedes submission,
then each-node executable/wrapper/library/config/output checks inside the allocation.

```sh
python3 -B experiments/distributed-command/test_review.py
python3 -B experiments/distributed-command/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-command/slurm-%j.log \
 experiments/distributed-command/run.sbatch command-01
# After terminal notice:
python3 -B experiments/distributed-command/analyze.py command-01
```

Next: complete phase host/GC/native working-set admission with real local counts,
representative larger rank/topology validation and fulls18 resource agreement.
**Full s18 preprocessing remains unrun/unadmitted.**
