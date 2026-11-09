# Linux compute-node cached s=15 validation

Subsequent `width-*` regression runs are documented in the separate
[count-width validation](../count-width-validation/report.md); their sources and
purpose differ from this original baseline.

Status: corrected job 72676 (`baseline-02`) **passed**. Its
[results](runs/baseline-02/results.json) and [batch log](slurm-72676.log) retain
completed evidence. This validates the unchanged
standalone mathematical encoder, without masks or application-plan integration.
Starting revision is `0f26303` plus the recorded setup/experiment changes.

## Reproduction

From the repository root, with shared Go 1.27.3 and downloaded pinned dependencies
under `.cache/liftrs/toolchains/` (see [setup](../../docs/hpc-setup.md)):

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/hpc-cached-s15/slurm-%j.log \
 experiments/hpc-cached-s15/run.sbatch baseline-01
```

[run.sbatch](run.sbatch) requests one gpuq node, 16 CPUs, 64 GiB RAM and one hour,
without requesting a GPU. [run.py](run.py) records job/environment information,
source hashes, binary digest, exact commands and outcomes under
`runs/<name>/results.json`, with phase JSON and profiles alongside it. Unique run
names prevent overwrites. Shared-home sources are copied to isolated node-local
scratch before tests; retained reference caches are read only.
Runner checks: `python3 -B experiments/hpc-cached-s15/test_run.py` verifies Linux
process-group RSS accounting and acceptance/rejection of retained/corrupted
canonical evidence; `bash -n experiments/hpc-cached-s15/run.sbatch` checks shell syntax.

## Method and boundaries

Use generic C++17 `-O3`, GCC 14.2.0 and Go 1.27.3, with 16 native workers.
Run focused race checks, vet, both full suites and the legacy fixture build before
the cached encoder. Adapt the existing [online-only probe](../../encoding/liftrs-shared/s15_firstuse_probe_test.go)
through its environment variables; no discovery/preparation is run. The original
macOS runner is not invoked. Stage only the verified online binary and metadata
into fresh job-local scratch. The probe writes new output there, never into the
reference cache. Constants retain their recorded regeneration provenance in the
[original s15 report](../parallel-s15/report.md).

Parameters: s=15,m=2,t=5,K=216,914,047,N=2^30, message seed 71. One complete
encoding is a portability baseline, not a repeated latency distribution. Separate
staging/checksums, loading, workspace/native initialization, encoding, validation,
packed output (including fsync), readback and output retention. Stage/output-copy
times use Python monotonic seconds; probe elapsed/native counters use nanoseconds.
Linux `/usr/bin/time -v` peak RSS uses KiB. Slurm enforces the one-hour wall limit;
the initial allocation exposed no finite memory cgroup limit. The runner therefore
samples each command's private process-group RSS every second and terminates it
above 64 GiB (a sampled guard, not a hard allocation cap; shared pages can be counted
more than once). Each command also has a timeout. The unchanged probe retains a 24 GiB Go soft target
and gets a 35-minute test timeout. Runtime and native memory are included in RSS.

Go CPU profiling remains enabled for each probe phase. Test perf access and retain
its outcome; a denied perf event is a native-profile limitation, not grounds to
disable Go profiles. GPU visibility does not prove allocation/access. Record CPU,
NUMA, affinity/cgroup limits, filesystem and actual allocated resources separately
from login-host inventory. Report setup/source-copy overhead separately from
encoding; this is neither cold-cache I/O nor an algorithm speedup against macOS.
The concurrent RSS monitor consumes CPU within the allocation; timings include
profiling and monitoring interference and are not uninstrumented latency baselines.

The allocation runs on `gnode02`, with an Intel Xeon 6710E host (two 64-core
sockets, one thread per core and two NUMA nodes). Its assigned CPU set is
`40-47,104-111`. Local scratch is XFS on the node's NVMe root filesystem; sources,
toolchain and retained results use NFS home storage. The [environment logs](runs/baseline-02/cpu.log)
and [recorded limits](runs/baseline-02/results.json) preserve exact observations.
Node state is shared with other allocations, so host contention is not controlled.

Acceptance requires the original complete canonical SHA-256, every systematic
symbol, 16 independent carryless-arithmetic curve checks, packed digest/length and
1,027 packed reads. Keep the new packed output and matching executable outside Git
under `.cache/liftrs/hpc-cached-s15/<name>/`; leave job scratch intact for diagnosis.
No external-memory or GPU encoder, s=16 support or TRL 5 completion is claimed.

## Completed baseline

[analyze.py](analyze.py), run with
`python3 -B experiments/hpc-cached-s15/analyze.py baseline-02`, verifies all
sealed evidence files, 227 recorded source files, the retained executable/output
and the original canonical/packed digests. Its [analysis](runs/baseline-02/analysis.json)
also checks native stage times against the enclosing native steady clock.
The three measured new runner/test sources have content-addressed copies in
[`source-snapshots`](../source-snapshots/README.md), preserving their exact bytes
independently of later edits. [Final verification](final-verification.json) records
script hashes, syntax/unit checks and documentation-link checks.
[Slurm accounting](accounting.log), collected with
`sacct -j 72675,72676 --format=JobID,State,Elapsed,AllocCPUS,ReqMem,MaxRSS,ExitCode -P`,
records successful job exit after 14m48s, including setup and correctness suites.

| Observation | Measured result |
|---|---:|
| NFS-to-local constant staging plus digest checks | 52.45 s |
| Online plan loading with digest validation | 5.872 s |
| Native plan/workspace setup | 0.016 s |
| Complete mathematical encoding, one repeat | 655.937 s |
| All systematic symbols, independent curves and canonical digest | 37.116 s |
| Packed output construction, hash and fsync | 12.622 s |
| 1,027 packed readbacks | 0.00127 s |
| New output copy to shared retention | 1.882 s |
| Probe process peak RSS (`time -v`) | 18.465 GiB |

All required validation passes: K=216,914,047 systematic symbols, 16 independently
evaluated degree-5 curves and canonical SHA-256
`5c09282e5cd2a77bd9541b5023a487f8deaf794a516e7e6a77007e7f9d08b265`.
The 2,013,265,920-byte packed output matches SHA-256
`bddc120cd9fc2d8c29743a2f9b83a17cb104f256d7e2fbc7e809533658373a2e`.
Both full Go/native suites, focused race checks and vet pass; the existing
optional network smoke test may skip when its configured server is unavailable.

Native stage totals sum to 655.712 s within the 655.937 s enclosing native
clock interval. Go and native elapsed clocks agree. Slurm's batch peak is
19,383,904 KiB, covering the batch rather than only the probe; its scope differs
from `time -v`. The runner's one-second sampled process-group peak is
19,830,255,616 bytes. None of these is a hard future RSS bound.

CPU profiles for all six probe phases remain beside the results. `go tool pprof`
can read the encoding profile with the retained binary, but 88.9% of its samples
appear as unsymbolized `[probe]`, with most remaining samples at `runtime.cgocall`.
The denied perf event prevents detailed native attribution on this host; stage
counters and Go profiles remain available. Sampled CPU totals are not calibrated
aggregate process CPU measurements. This single shared-node run is not a speedup
trial against the laptop or an optimization-selection experiment.

The shared Go toolchain, pinned modules, executable and new packed output remain
outside Git at the paths documented above. Source-node reference files were not
modified. [s=16 readiness](../../docs/encoding-s16.md) remains a representation and
preprocessing-memory gate; this run changes no encoder guards or arithmetic.

## Initial allocation correction

Job 72675 (`baseline-01`, [partial results](runs/baseline-01/results.json),
[supervisor receipt](runs/baseline-01/supervisor-status.json)) reached passing race/vet/legacy-build checks, but was
cancelled during the pure-Go suite before the large encoding. Its recorded
cgroup limits showed `memory.max=max` despite `--mem=64G`, so the requested memory
was not an enforced cap. Preserve this partial attempt and its logs separately;
the corrected runner adds explicit process-group memory/deadline monitoring.
No preprocessing or output reference was changed. `perf` was denied and the
CPU-only allocation exposed no GPU devices, as expected without a GPU request.
