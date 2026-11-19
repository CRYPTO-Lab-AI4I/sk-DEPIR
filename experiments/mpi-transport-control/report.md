# TCP-only MPI transport control

Status: **78001 timed out before completing any transfer** (SCALING-132).
[Receipt](submission.json), [raw results](runs/tcp-01/results.json),
[review](runs/tcp-01/analysis.json), [accounting](runs/tcp-01/scheduler-accounting.txt),
[progress](slurm-78001.log). Implementation `f937bcd`; raw failed state preserved.
[Runner](run.py), [scheduler script](run.sbatch),
[probe](../mpi-host-buffers/probe_test.go), [reviewer](../mpi-host-buffers/analyze.py).

## Review of 78001

Verified **654 source /15 original evidence /3 cache hashes**. All four workers
recorded ob1/self,tcp and512 KiB staged buffers on HPC gnode13/gnode14 (4 ranks,
4 CPU cores,8 GiB requested host RAM total,zero GPUs). The first staged-normal
arm hit its **90 s test deadline**. Scheduler elapsed **1 min55 s**, controller
**115.021 s**, failed command **90.578 s**. Largest logged process peak RSS/rank:
**44.938 MiB**, below the1 GiB limit. No memory guard failure.

**Zero completed handoffs, zero proofs, zero mismatch captures.** Eight heartbeats
at30/60 s kept completed counts at0; these were waiting observations, not progress.
Go timeout stacks place rank1 in SendRoot/send_values and rank0 in ReceiveRoot/
receive_values on same-node case0/cycle0. Later arms were not run. This does not
show a corruption fix, reproduced corruption or a TCP throughput penalty; the
forced transport was unusable for this test. The cause of the stall is unknown.
Do not extend a deadline based on these non-advancing counts.

Next: [observe automatic component selection](../mpi-transport-discovery/report.md)
under the original host replay, retaining the same resource budget and captures.
No further component is forced before observing which one the working configuration
selects. Default direct corruption and staged bounded results remain unchanged;
no production default or full s=18 claim. Source parser now ignores Go stack
struct arguments while still failing on malformed JSON events; raw logs unchanged.

The [host replay78000](../mpi-host-buffers/report.md#review-of-78000) reproduced
host payload corruption without CUDA or forced reclamation. Staged controls passed
both normal and forced reclamation arms. The direct failure was cross-node at
cycle0/case1. This warrants isolating the MPI transport from the same Go allocation,
payload, churn, proof and capture behavior, not changing arithmetic or GPU code.

Set `OMPI_MCA_pml=ob1` and `OMPI_MCA_btl=self,tcp` explicitly in the worker environment.
This requests TCP data transport, including same-node traffic, instead of the
MPI deployment's automatic component choice. The exact original active transport
was not established; do not label it RDMA or attribute a specific component bug.
Capture both requested settings in every worker host record; review rejects a
missing/different setting. Retain installed OpenMPI component metadata/warnings.
No other probe logic, chunk size, route, fixture, arm ordering or resource budget
changes. No MPI/library upgrade, GPU allocation or production default change.

Reuse all four existing arms, staged-normal / staged-reclaim / direct-normal /
direct-reclaim, with 64 cycles × four routes each. Replays the sealed sender
polynomial from77868 and synthetic uint64 high-bit patterns. Normal arms still
have artificial 16 MiB/rank churn; only reclaim arms additionally force Go GC/page
release. Expected payload hashes, exact progress/placement and both endpoint
captures/proofs remain mandatory; stop on failure. CPU profiles/phase/resource
logs, completed counts after barriers and 30 s waiting heartbeat remain enabled.

The TCP control adds no arrays/transforms/production output; owned staging stays
512 KiB/rank when selected. It may increase CPU work, copies and communication
latency. Memory/time cost and large-message throughput are unknown; this explicit
control must not silently replace the faster validated preprocessing path.
If direct passes, that supports transport-dependent behavior under this probe,
not a proven cause or permanent repair. If it fails, retain both captures and
continue isolation. Earlier failed gates remain failed even if this control passes.
Full s=18 preprocessing remains unrun.

A proposed `GODEBUG=madvdontneed=1` control was discarded **before submission**:
`runtime/runtime1.go` already defaults that setting to1 on Linux. Reading only
`mem_linux.go` missed initialization. Source hashes/excerpts are retained in
[readiness](readiness/runtime-inspection.json); no runtime flag is changed here.
Worker GODEBUG metadata is now logged for future evidence. This correction does
not establish how earlier MPI memory interactions occurred.

HPC resources unchanged: **2 nodes / 4 ranks / 4 CPU cores / zero GPUs**, 4 GiB
requested host RAM/node, 1 GiB process RSS guard/rank, 16 MiB churn/rank plus
sub-MiB payloads, at most 512 KiB staging/rank. Existing caps:128 MiB retained,
2 GiB build scratch,32 MiB/file,720 s controller/15 min scheduler. Each arm has
90 s test /100 s step /120 s command cutoff; profile report90 s. The original
host replay ended in29 s including an early failure; a complete control may run
longer. Estimated **1–3 min after scheduling**, not measured. No budget increase.

```sh
python3 -B experiments/mpi-transport-control/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/mpi-transport-control/slurm-%j.log \
  experiments/mpi-transport-control/run.sbatch tcp-01
# After user completion ping only:
python3 -B experiments/mpi-host-buffers/analyze.py tcp-01 \
  --experiment experiments/mpi-transport-control
```

Results/profiles/checkpoints: `runs/tcp-01/`; progress:`slurm-JOBID.log`; retained
replay/binary/captures: `.cache/liftrs/mpi-transport-control/tcp-01/`, outside Git
with seals in results.json. Original fixture hash/path and toolchain described in
[host replay](../mpi-host-buffers/report.md). Fresh run name plus recorded revision
regenerates the experiment; no useful cache/profile is deleted. No agent monitoring.

[Local verification](readiness/checks.json): host binary compiles and hash/proof
unit passes;28 Python reviewer tests pass, including rejection when the requested
transport does not reach every worker. Python/scheduler syntax checks pass.
No local MPI execution; actual transport control remains unvalidated until review.
