# Two-node memory redistribution gate

Status: **job77127 submitted**, following explicit two-node budget approval.
[Receipt](submission.json), [results](runs/exchange-01/), [Slurm log](slurm-77127.log).
Prepared probe/controller source `9a1bbf7`; the runner seals its actual revision
and source identities. No monitoring after submission.
Work SCALING-106. No distributed GPU/root performance or bandwidth established.

[Probe](probe.cpp), [guarded runner](run.py), [review tests](test_review.py), [Slurm entry](run.sbatch).
This optional server experiment uses installed Open MPI4.1.8 with Slurm resource
support, selected after module/tool inspection. It adds no MPI/CUDA dependency
to ordinary library or CPU-client builds. Local module queries emit a sandbox
socket-permission diagnostic; compilation succeeds. Compute-node transport
initialization and actual fabric availability remain untested. Installed
components do not prove RDMA or GPU-direct support or bandwidth.

Each of two ranks on distinct nodes keeps uint64 extension-field test patterns
in host RAM and exchanges them directly with MPI_Alltoall. Patterns include bit33
and different rank/range values; every received value and reverse-exchange value
is checked. Ranks on the same host are rejected. Buffer sizes1/8/32MiB each are
fixed; maximum explicit send+receive memory is64MiB/rank. No intermediate arrays,
communication queues or compute scratch are written to the shared filesystem.
Only executable/metadata/log artifacts use shared storage. This models host-staged
redistribution, not a field transform, GPU transport or parallel root computation.

For each size, one warm-up and three timed exchanges run; reverse exchange and
full comparisons are outside each forward clock but inside command duration.
Report slowest-rank forward latency, payload size and per-rank peak RSS. A two-node
all-to-all sends half of each rank's input remotely; use that byte count, not
entire input size, when deriving network throughput. No kernel benchmark or
production speedup may be inferred. A future four-node schedule and Taylor
traffic remain unmeasured. Timestamped pass start/completion events flush at each
bounded exchange; if a single exchange stalls, remaining work is unknown and the
controller deadline terminates it rather than reporting false progress.

Approved budget (explicit user agreement recorded in the submission receipt):

| Resource | Limit |
|---|---|
| HPC nodes |2 eligible nodes; gnode03 excluded |
| GPUs |0 |
| CPU allocation |2 cores/node,4 total;1 rank/node |
| Requested host RAM |4GiB/node,8GiB aggregate |
| Explicit data buffers |64MiB/rank |
| Per-rank RSS guard |1GiB, sampled before/after each pass |
| Retained artifacts/scratch |Binary and bounded metadata/logs only, expected less than16MiB;32MiB per artifact file enforced locally (fewer than128MiB total with this fixed command list); no array files |
| Deadline |300s MPI command, kill-after10s; bounded wrapper under480s;600s scheduler |

The runner records source hashes and exact Git revision, commands, rank hosts,
MPI build/runtime configuration, complete pass records and evidence/binary
SHA256 identities in `runs/<name>/results.json`. The binary remains reusable at
`.cache/liftrs/preprocessing-transport/<name>/`. No prior cache is modified.
CPU profiling is retained in the separate root-validation job; this initial MPI
latency gate intentionally uses no profiler to isolate transport startup/exchange
behavior, with validation/logging boundaries explicit. It is not a CPU/GPU
optimization profile. CPU/native transport profiling remains a later diagnostic.

```sh
module load gcc/14.2.0 openmpi/gcc/64/4.1.8
mpicxx -std=c++17 -O3 experiments/preprocessing-transport/probe.cpp -o /tmp/depir-transport-probe
python3 -B experiments/preprocessing-transport/run.py --check
# Only after explicit two-node budget approval:
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/preprocessing-transport/slurm-%j.log \
  experiments/preprocessing-transport/run.sbatch exchange-01
```

No large write workload is included. Persistence is a separate later gate with
large buffered sequential writes, one owner per file, controlled concurrency,
write-call/byte counts and write/fsync/readback timing. Stage persistence and
computation measurements separately before combining them into a full pipeline.
No fullsize workload, transport choice for production or resource increase is
implied by preparing this runner. Submit and hand off without monitoring after
approval; review only after completion ping.
