# Two-node memory redistribution gate

Status: **two-node validation77127 passed** on HPC gnode13/gnode14, zero GPUs.
Complete scheduler job **5 s**. [Receipt](submission.json),
[raw results](runs/exchange-01/results.json), [verified analysis](runs/exchange-01/analysis.json),
[accounting](runs/exchange-01/accounting.txt), [Slurm log](slurm-77127.log).
Producer revision `e06a33d`; three source identities, four evidence logs and the
retained compiled binary reconcile against the producer's seals without overrides.
Work SCALING-106/107/108. This measures host-memory exchange, not distributed GPU/root performance.

[Probe](probe.cpp), [guarded runner](run.py), [review tests](test_review.py), [completion analyzer](analyze.py), [Slurm entry](run.sbatch).
This optional server experiment uses installed Open MPI4.1.8 with Slurm resource
support, selected after module/tool inspection. It adds no MPI/CUDA dependency
to ordinary library or CPU-client builds. Local module queries emit a sandbox
socket-permission diagnostic; compilation succeeds. Compute-node MPI initialization and host transfers now pass; the selected
fabric/protocol and GPU-direct support were not instrumented. Installed
components do not prove RDMA or GPU-direct support.

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

## Completed host-memory exchange (SCALING-108)

All three input sizes, one warm-up plus three measured exchanges each, pass on
both distinct nodes. Every received uint64 value and reverse-exchange value is
checked, including bit33 and rank/position ownership. Both ranks complete;
all four commands exit0. No shared-filesystem array or GPU workload was used.
The operation has an MPI barrier before timing; forward call time includes local
and remote all-to-all work, and is reduced to the slowest rank outside the clock.
Validation and the reverse exchange remain outside the forward clock. Logging,
allocation, startup and binary compilation are included in the5-second job.

| Input per rank | Remote bytes sent per node per forward pass | Three forward latencies | Median | Remote payload throughput per node |
|---|---|---|---|---|
| 1 MiB |0.5 MiB |98.488–101.499 microseconds |99.122 microseconds |4.926 GiB/s |
| 8 MiB |4 MiB |0.585–0.691 ms |0.663 ms |5.894 GiB/s |
| 32 MiB |16 MiB |2.263–2.574 ms |2.447 ms |6.385 GiB/s |

Throughput is **remote bytes per node divided by median slowest-rank latency**.
Each node also receives the same amount. Two rank records share the same reduced
time and are not independent samples: there are three samples/size, not six.
For32MiB inputs, aggregate bidirectional sent payload is12.770 GiB/s; that is not
per-node throughput or measured physical link bandwidth. No confidence bound,
large-message saturation study or compute/transfer overlap is established.

Explicit data buffers peak at64 MiB/rank. OS-reported peak process RSS is
**297.734 MiB on rank0 /287.867 MiB on rank1**, including MPI/runtime overhead,
within the1GiB guards. Scheduler step MaxRSS reports only about5MiB and does not
represent these rank processes; it is retained as accounting, not substituted
for the rank measurements. The job requested4GiB per node (8GiB aggregate),
2CPU per node (4total), no GPU. Binary remains in
`.cache/liftrs/preprocessing-transport/exchange-01/probe`; logs and MPI configuration
are retained with source/evidence/binary hashes. No cache was deleted.

Reproduce the completion review without any scheduler/process execution:

```sh
python3 -B experiments/preprocessing-transport/test_review.py
python3 -B experiments/preprocessing-transport/analyze.py exchange-01
```

Three review tests pass, including missing/same-node/RSS cases and the factor-of-two
remote-byte calculation. No producer execution code was changed during review.
Actual selected transport/adapter, GPU staging, four-node contention, distributed
transform schedules, root workers and large buffered persistence remain separate
validation gates. This result supports proceeding with direct messages; it does
not validate complete s18 preprocessing or justify applying this bandwidth to
four-node/full-size forecasts. No new job launched, and root job77126 was not
inspected as part of this completion ping.
