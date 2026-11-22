# Fresh-process restart and bounded canonical publication

Status: **submitted as78915**, run `restart-01`, implementation `98695bc`.
[Submission receipt](submission.json), [scheduler log](slurm-78915.log);
results will be retained under `runs/restart-01/`. No integrated outcome yet.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[focused reviewer tests](test_review.py), [readiness](readiness/checks.json).
This follows [passed partitioned discovery78909](../distributed-discovery-pilot/report.md).
Actual revision/source/cache identities are recorded by the allocation controller.

## What changes

Use the same genuine s4/s6 pipeline with explicit
`PIR_PIPELINE_PUBLICATION=stream`. The previous gather mode remains the default
when omitted. `PIR_PIPELINE_RESUME_FROM` optionally names a completed cold-run
cache; empty means cold execution. No public API or client dependency changes.

The cold arm performs partitioned discovery, local roots, distributed root/Newton/
identity/spectrum, prefix commits and streamed publication. The second arm starts
new MPI processes in the same allocation, loads root and completed Newton prefix8
from disk, skips discovery and root computation, and continues all remaining
Newton steps. The full independent P*G identity, final spectrum and canonical
artifact comparison still run. It does not reuse in-memory process state.

[Checkpoint loading](../../encoding/liftrs-shared/distributed_recovery.go) checks
bounded metadata, s/N/K/rank count, completed-prefix shape, trusted support hash,
per-shard hashes, canonical field values, padding and local array admission before
GPU allocation. Root/prefix records must agree and match the independent native
support. Missing, corrupt or incompatible state fails early. Checksums detect
corruption; trusted source/experiment provenance remains required and is not
replaced by a hash. The current pilot resumes the same rank count and distribution.
It does not validate arbitrary node loss, elastic resizing or a scheduler restart.

[Publication](../../encoding/liftrs-shared/distributed_publication.go) streams
rank-ordered coefficient intervals through one capped receiver tile on rank0.
Existing canonical codecs write the header, normalized P and spectrum directly
to the code file; G is written as a reusable generator file. Only non-padding
words are sent. No complete receive/publication polynomial or spectrum array is
allocated. The typed MPI adapter shares one descriptor/codec/proof implementation
for32/64-bit words; ordinary CPU tests cover both widths. The GPU pilot remains
two ranks with the original s4/s6 field/word representation, not s18 support.

The coordinator's independent CPU oracle is still retained for this small gate,
and discovery still gathers a diagnostic support archive. Removing publication
copies does not remove those diagnostic full arrays or establish a fulls18 RAM
layout. All existing GPU kernels and arithmetic recurrences are unchanged.

## Validation kept proportionate

One allocation, two arms:

1. Complete cold genuine s4/s6 run with streamed publication.
2. Fresh MPI-process resume for both instances from the cold arm's prefix8.

No repeated CUDA sanitizer:78909 already passed the maintained kernels with the
smallest integrated sanitizer; changes here concern host loading, typed transport
and I/O. Focused CPU/race tests cover source-preserving32/64-bit ordered publication
across1/2/3/5 contiguous ownership partitions, uneven/empty shards, and invalid
checkpoint hashes/bindings/prefixes/memory admission. Native genuine shape tests
run without MPI/GPU initialization before submission.

The independent reviewer retains all complete artifact/prefix/identity and exact
MPI/transpose/Taylor route checks. It admits only the resumed suffix and validates
that root/prefix8 bytes match the cold source; source process identity must differ
from the resumed process. Cold and resumed artifacts must match the retained78898
canonical hashes and independent native output, including the streamed generator.
Fresh-run metadata explicitly reports skipped discovery and prefix8; a resume
cannot pass by silently running cold preparation again. The old/default modes
retain their existing checks and behavior.

## Memory, storage and time trade-offs

Publication previously retained complete P/G/spectrum receive arrays totaling
12*N bytes on rank0 (padded arrays), plus two complete serialized code buffers
of about8*N+8*K bytes together. The streamed path removes those copies and uses
one tile of at most4096 words (16 KiB at the pilot width), plus two16 KiB file
comparison buffers. Input shards and the CPU diagnostic oracle remain. The generic
streamer selects file width from Word; callers must match canonical section/header
width and admit each owned source array.

A unified generator file adds4*(N−K) bytes of retained diagnostic/reusable output
beside the existing G shards. Canonical files remain byte-identical. Runtime may
change with extra file operations, chunking and reduced network/padding traffic;
no stable penalty or speedup is known. Stream mode stays explicit, and gather
remains the previous default. No additional FFT, CUDA array or host-backed matrix
strategy is introduced. Prefixes/profiles and prior caches are preserved.

At the largest N4096 instance, removed publication-array payload is48 KiB and
two old code buffers about47 KiB; fixed bounded buffers can dominate this tiny
case. This is an ownership/restart correctness gate, not a small-case memory
speedup claim. Fullsize memory/time need representative admitted measurements.

## Resources and measurement boundaries

Same admitted **two HPC nodes / two H200s / four CPU cores / 8 GiB requested host
RAM**, one GPU/two cores/4 GiB per node; **128 MiB owned GPU/card**, **1 GiB RSS/rank**,
no pinned buffers, **128 MiB retained cache / 2 GiB scratch / 32 MiB file**,
**720-second controller / 15-minute scheduler**. Exclude gnode03; other eligible
nodes allowed. No resource increase. New loading checks run on every rank before
GPU work and use the existing shared-path/working-directory/library preflight.

78909 completed two integrated arms/builds/profiles in141.458 seconds. Two cold/
resume arms reuse one frozen build; budget **3–6 minutes estimated**, excluding
queue, within the same240-second launcher/220-second test cutoff per arm and
controller/scheduler headroom. No new native build is needed locally because CUDA
sources are unchanged; the allocation still rebuilds once from the frozen source.
Cache remains projected below32 MiB and per-process proof files below20 MiB,
with completed cold output saved before resume starts.

All timings include independent native reference work, logging/profiling,
checkpoint commits and streamed verification; compare cold/resume phases as such,
not as unprofiled end-to-end performance or kernel speedup. Resume still recomputes
its small CPU oracle for validation, but skips GPU discovery/root work. Native
profiling cannot separate GPU execution and waits inside cgo. Flush timestamped
phase/prefix/pass events, packet proofs and resources using the existing logger.
Do not monitor after submission; wait for the user's completion ping.

```sh
python3 -B experiments/distributed-restart-pilot/test_review.py
python3 -B experiments/distributed-restart-pilot/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-restart-pilot/slurm-%j.log \
  experiments/distributed-restart-pilot/run.sbatch restart-01
# After the user's completion ping:
python3 -B experiments/distributed-restart-pilot/analyze.py restart-01
```

Raw results/profiles/logs: `runs/restart-01/`. Cold/resume prefix/code evidence,
source and binaries: `.cache/liftrs/distributed-restart-pilot/restart-01/`, outside
Git with checksum manifests. `cold/` remains intact when `resume/` starts. This
is a fixed-configuration fresh-process restart, not fullsize/node-failure recovery.
