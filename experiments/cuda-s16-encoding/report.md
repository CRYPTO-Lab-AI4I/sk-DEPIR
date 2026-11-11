# Full cached s16 CUDA client encoding

Status: job73257 completed and evidence reconciliation passed. The primary optional CUDA backend
passed [modular validation job73243](../cuda-encoding/report.md). CPU (Go/C++)
remains a maintained main client path; code preprocessing can be obtained from
another machine independently of client hardware.

## Reproduce and preserve evidence

[run.py](run.py) executes the guarded workflow;
[run.sbatch](run.sbatch) requests its allocation; [analyze.py](analyze.py)
reconciles completed evidence; [check.py](check.py) tests admission rejection.
The worker is [TestCUDACachedEncoding](../../encoding/liftrs-shared/cuda_cached_test.go),
using the same profiled load/setup/encode/verify/pack/readback harness as the
[CPU reference](../cached-s16-encoding/report.md). Run from the repository root:

```bash
module load slurm/slurm/25.05
python3 -B experiments/cuda-s16-encoding/check.py
sbatch --output=experiments/cuda-s16-encoding/slurm-%j.log experiments/cuda-s16-encoding/run.sbatch encoding-01
# Only after the completion ping:
python3 -B experiments/cuda-s16-encoding/analyze.py encoding-01 --recorded-cache
```

The controller records revision, exact source SHA256 identities (including dirty
files as durable source snapshots), build commands, binary/cache/evidence hashes,
CPU affinity, host/GPU/driver/compiler/Go versions and sampled process-group RSS.
Production encoding files must match the passed validation manifest; the new
full-size test adapter is recorded separately. Build uses GCC14.2.0, CUDA12.9.1,
Go1.27.3, cgo and the independent `liftrs_cpp,liftrs_cuda` build features. The
shared library is rebuilt for the allocated GPU with `nvcc -arch=native`; its
runtime name is supplied explicitly. The public CUDA-only API does not require
C++; this benchmark uses native clocks from the shared CPU comparison harness.

Results are written to `runs/encoding-01/results.json`, phase JSON, profiles and
logs. Allocation output is `slurm-JOB_ID.log`; the retained executable/library,
packed output and metadata are under `.cache/liftrs/cuda-s16-encoding/encoding-01/`.
These large local artifacts remain outside Git. Source snapshots and reasonably
sized profiles are versioned. No process/scheduler monitoring occurs after
submission until the user pings. Automatic guards execute inside the job.

## Workload and boundaries

Use s16,m2,t5, K918,672,421 and N4,294,967,296, deterministic message seed71.
Load the existing verified `online.bin` from
`.cache/liftrs/gpu-preparation/s16-m2-t5/` (20,854,558,905 bytes, SHA256
`f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874`).
The shared loader checks its digest; no discovery, preparation or cache rewrite
is included. A small s6 cached smoke test runs first, checked against its
retained CPU word. The full test creates one GPU session, uploads reusable code
constants once, and performs two same-message encodings using the same device
buffers. All five transforms/products/assembly stay on device between message
upload and final canonical output download. Public encoding latency includes
allocation and both transfers; per-stage CUDA event/wall clocks separate costs.
Device setup, cache loading, independent correctness checks, hashing, packing
and packed readback are separate phases. Profiling is enabled for both encodes;
these are instrumented measurements, not unprofiled latency baselines.

Every repeat checks all K systematic values,16 independent sampled curves and
repeat digest equality. Both complete canonical digests and the final packed
digest must equal [CPU job73131](../cached-s16-encoding/runs/encoding-01/results.json):

- Canonical: `c538a1bc9b408dc4d283aba887f808825404a1c2b8dce693aa7f3015d1152a48`.
- Packed: `1c68bfe60f3914b06e31995afc67d5bb3d0a6776c4750af836f8c9794ee18bb9`.

The final word is packed into8,589,934,592 bytes (16 bits/value) and1027 positions
are read back. Packing stays on the host through the shared format. This creates
one8 GiB output; it does not repeat the CPU encoding or duplicate preprocessing.
The CPU encode reference is1252.502871 seconds at32 threads on a previous
allocation. Any ratio to that number is a historical comparison, not a matched
CPU/GPU benchmark. Seconds, bytes and binary GiB are used throughout.

## Admission and limitations

One GPU,16 CPU cores,128 GiB requested host RAM,96 GiB sampled RSS guard,64 GiB
Go soft limit,96 GiB explicit device-allocation cap. Require at least98 GiB free
allocated-device memory and128 GiB initially available host RAM; maintain32 GiB
node headroom. This targets the available H200-class GPU, with actual hardware
recorded in the manifest. This implementation requires NVIDIA CUDA; the generic
CPU implementation has no GPU/toolkit dependency.

The device payload model is48 GiB for three full arrays plus19.42 GiB reusable
code, about67.42 GiB before small tables/plans and driver overhead. The cap covers
owned device allocations, not CUDA driver/context memory. Principal host input,
code and output arrays total about38.84 GiB; Go/runtime/verification overhead is
additional and measured. These are admission models, not measured full-run peaks.
Allocation deadline2 hours; controller110 minutes; full worker90 minutes; smoke
5 minutes. New cache and scratch are independently capped at16 GiB and require
20 GiB free initially. Existing caches are read-only links, excluded from new
storage accounting. Guards terminate the process group and preserve available
profiles/checkpoints; failure stops dependent work.

`analyze.py --recorded-cache` verifies all sources, evidence, metadata and sizes,
but uses the completed controller's sealed packed-payload digest rather than
rereading8 GiB. Omit that option for a bulk digest audit (hand off if over30 seconds).
The completed result below applies to this configuration and hardware. Variable-time arithmetic,
trusted code metadata, prototype randomness and cryptographic/application
integration limits remain unchanged; this is mathematical encoding, not TRL5
acceptance or a production side-channel security claim.


## Completed result

[Results](runs/encoding-01/results.json), [reconciled analysis](runs/encoding-01/analysis.json),
[phases](runs/encoding-01/encoding/phases.json) and
[final Slurm accounting](runs/encoding-01/accounting.log) record successful
execution, exit0:0,30m02s allocation elapsed. All317 source identities, recorded
cache/evidence hashes, both full CPU canonical digests, packed digest, all K
systematic checks,16 sampled curves per repeat and1027 packed readbacks pass.
This reconciliation uses the sealed completed-controller digest for the8 GiB
packed file, checking its physical size/metadata without a fresh bulk reread.
The revision is ae27de56679ae031154cc3f290a04c753ecaed19 plus exact source snapshots
in the manifest. No production encoding correction was needed.

Hardware: gnode17, NVIDIA H200 NVL143,771 MiB, driver580.126.20, Xeon6710E,
16 allocated CPU cores,128 GiB requested RAM, GCC14.2.0, CUDA12.9.1, Go1.27.3.

| Phase | Seconds |
|---|---:|
| Load/check existing online cache | 164.070 |
| Reusable GPU session setup/code upload | 7.193 |
| First encoding, including transfers and allocation | 85.954 |
| First independent verification | 543.395 |
| Second encoding, same session/message | 60.216 |
| Second independent verification | 763.995 |
| Pack/write final8 GiB output | 56.689 |
| Packed readback check | 1.629 |

The full worker took1694.873 s (28m14.9s). Its independent verification/readback
phases total1309.019 s (21m49s), so the30-minute validation allocation is not the
encoding latency. CPU reference job73131 took1252.503 s per encoding. These GPU
calls are14.57x and20.80x faster than that historical32-thread CPU run. Host/setup
and allocation conditions differ; this is not a matched speedup trial or a broad
hardware guarantee. The first message including cache load, session setup,
encoding and pack/write sums to313.906 s (5m13.9s), excluding correctness auditing,
profiling boundaries, builds and other controller work. A later call with packing
is about116.905 s in this run; only two same-message calls were measured.

Device-owned allocations:72,394,976,680 bytes (67.423 GiB), below96 GiB cap;
this excludes driver/context memory. Peak sampled process-group RSS39.024 GiB;
Slurm reports39.05G. No guard tripped. Prepared constants and both CPU/GPU packed
outputs remain retained; GPU output remains tightly packed at16 bits/value.

### Where the time went

| Instrumented encoding component | First call, seconds | Second call, seconds |
|---|---:|---:|
| Message upload | 0.438 | 0.442 |
| Taylor transforms | 6.309 | 6.310 |
| Matrix columns | 20.998 | 21.009 |
| Matrix rows | 24.377 | 24.384 |
| Layout | 0.643 | 0.643 |
| Products | 0.205 | 0.204 |
| Canonical assembly | 0.363 | 0.360 |
| Output download | 31.837 | 1.152 |
| Native total wall clock | 85.941 | 55.368 |
| Public API total wall clock | 85.954 | 60.216 |

These instrumented components use different clock boundaries and are not an
exhaustive additive wall-time decomposition. The first download is much slower;
page commitment/staging is a possible explanation, not a verified cause. The
second Go profile attributes about4.85 s to output-buffer clearing, accounting
for most of the gap between API and native wall time. Go profiles otherwise
attribute native time to cgo and do not identify GPU instruction hotspots.
Matrix rows/columns dominate warmed device work; focus subsequent optimization
on those kernels and output-buffer/transfer reuse. Packing is also comparable
to warmed encoding time when persistent output is required. No optimization or
extra benchmark is automatically launched from this review.

Full-size mathematical GPU encoding now passes with repeat/CPU output agreement.
This does not establish application/cryptographic encoding, constant-time security,
s17 support, smaller-GPU feasibility or TRL5 acceptance. CPU remains a main path.
