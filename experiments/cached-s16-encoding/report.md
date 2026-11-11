# Cached s16 CPU encoding

[Runner](run.py), [launcher](run.sbatch), [analyzer](analyze.py) and short
[metadata admission self-check](check.py). User authorized publication of the
validated CPU Taylor change, then one actual cached encoding. Parent runtime
revisionae27de5. Exact measured revision, all runtime source hashes/snapshots,
compiler/node/CPU/affinity, commands, cache/binary hashes and profiles are recorded
in each result manifest. [Encoding-01 results](runs/encoding-01/results.json),
job73131, now pass; the completed checkpoint is below.

```sh
python3 -B experiments/cached-s16-encoding/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/cached-s16-encoding/slurm-%j.log experiments/cached-s16-encoding/run.sbatch encoding-01
# After the user's completion ping:
python3 -B experiments/cached-s16-encoding/analyze.py encoding-01
```

## Input and computation

Reuse the unchanged [TestS16FirstUse worker](../../encoding/liftrs-shared/s16_firstuse_probe_test.go)
in online-only mode and its [phase harness](../../encoding/liftrs-shared/s15_firstuse_probe_test.go).
All database-dependent computation uses the C++ CPU matrix path with the selected
hybrid Taylor schedule. One32-worker encoding, message seed71, s16,m2,t5,
K918,672,421,N4,294,967,296. No GPU allocation, discovery, code preparation,
root-product reconstruction or automatic second full encoding.

Load the trusted19.42 GiB online.bin retained by GPU preparation job72875:
`.cache/liftrs/gpu-preparation/s16-m2-t5/online.bin`, SHA256
`f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874`.
Require that completed preparation manifest/digest, exact spec/K/N/byte count and
identity-verified metadata. Read-only symlinks in a fresh output cache point to
online.bin/metadata; no large input copy or original-file overwrite. The worker's
streaming loader verifies the actual input digest while loading, separately timed.

Encoding sources must match validated CPU Taylor job73099. Build one native-tag
test binary, run focused parallel-Taylor/full-width/count checks, then a tiny
cached s6 online-only smoke. Its canonical/packed hashes must match retained
reference data. Only then run s16. These checks do not repeat expensive code
preprocessing. Existing full suites/native sanitizers passed in the preceding
comparison; no encoder runtime code is changed for this experiment.

The full worker checks all918,672,421 systematic positions,16 independent sampled
curve restrictions using carryless base arithmetic, full canonical SHA256,
packed output SHA256 and1,027 packed readbacks. Retain an8 GiB packed word and
metadata. No independent full s16 reference word exists, and one sample does not
establish fresh-repeat equality, a robust latency distribution or production
security. This is standalone mathematical encoding, not application partitioning,
masking, storage upload or complete cryptographic encoding/TRL5 acceptance.

## Timing, profiles and resource review

Load/digest verification, native plan/workspace initialization, encoding, output
checks, packing/persistence and readback have separate profiled wall clocks.
Encoding includes allocation and five native transforms (three prefix2^31, two
full2^32), point products and canonical assembly. Native counters separate Taylor,
columns/transposes, rows, buffer, pointwise and assembly. Check their sum against
the enclosing native steady clock. Go profiles do not provide native instruction
attribution under the host's perf restrictions. Preserve all profiles, sources
and exact binary; label instrumented timings, excluding queue time/preparation.

Actual policy:one CPU-only32-core job in gpuq,256 GiB requested host RAM,
224 GiB sampled process-group RSS guard,existing worker192 GiB Go soft target,
2-hour allocation/110-minute controller/90-minute worker timeout. Require256 GiB
initially available node RAM,32 GiB during execution,20 GiB free shared storage
and scratch. New cache and scratch guards16 GiB each, excluding read-only input
symlinks. Slurm's request is not a demonstrated hard memory cgroup limit.

Principal in-memory payload model, GiB (not measured RSS):P/spectrum19.42,
overlaid convolution arena16, H3.42, transpose scratch16, input3.42, canonical
output16, totaling≈74.27. Thread stacks, metadata, runtime/GC and temporary
retention are additional. The established256/224 GiB envelope leaves headroom;
no inference that these limits are minimum hardware requirements. New durable
output≈8 GiB plus binary/smoke/metadata fits16 GiB cache cap. Input preparation
and all prior profiles remain untouched.

Controller samples process-group RSS, available node memory and cache/scratch
usage, forwards TERM to flush worker profiles, then KILL after its grace period
if needed. Incomplete .tmp files are retained but never accepted as successful
output. No automatic restart. The20–35-minute encoding/30–50-minute complete-run
conversation estimate remains an extrapolation from s14, not a guarantee.

## Retention and handoff

New data lives under `.cache/liftrs/cached-s16-encoding/RUN/`:compiled binary,
smoke and encoding subdirectories, read-only input links, packed word and word
metadata. Raw phase/results/profiles/source snapshots remain in the experiment
and source-snapshots. The final controller hashes new artifacts, and the analyzer
can recheck them after completion. Preserve failed attempts too.

Submit then hand off under AGENTS rule8. Do not poll scheduler/process state,
follow logs, wait for completion or analyze until the user's completion ping.
Automatic guards remain active; dependent work waits. No completion claim or
new performance number is recorded before final evidence review.


## Completed cached encoding

Job73131 at runtime revisionae27de5 completed successfully on gnode17,
Xeon6710E, GCC14.2.0 and Go1.27.3,32 CPU workers/no GPU. Exact environment,
affinity and commands are in the result manifest/logs.
[Slurm accounting](runs/encoding-01/accounting.log) records COMPLETED/0:0 and
**34m53s** allocation elapsed. This includes checks/build/profiling/finalization;
it excludes queue wait and all discovery/preparation.

| Phase | Profiled wall time |
|---|---:|
| Load and verify19.42 GiB online input | 4m22.332s |
| Native plan/workspace initialization | 10.213 s |
| One complete encoding | **20m52.503s** |
| Systematic/curve/canonical-hash checks | 6m42.171s |
| Pack/hash/persist output | 56.277 s |
| Packed readback checks | 1.506 s |

The worker process takes2004.025 s (33m24s); sum of phase clocks1985.002 s
excludes profiler start/stop, explicit GC and some wrapper work. Allocation
elapsed additionally includes the tiny smoke, targeted checks, build and artifact
hashing. The20–35-minute encoding and30–50-minute complete-run estimates were
consistent with this measurement. No matched old-code s16 encoding was run, so
this is not a measured s16 Taylor speedup ratio.

All918,672,421 systematic positions,16 independent sampled curves and1,027 packed
readbacks pass, including the full-field output range checks. Five transforms
execute within one native boundary call. The controller finishes every command
with exit0 and no guard stop. Sampled process-group RSS peaks at79,854,313,472
bytes (**74.370 GiB**), below224 GiB. The74.27 GiB principal-array estimate was
close, but remains a payload model rather than a proven memory limit or minimum
hardware requirement. Slurm batch MaxRSS has its own sampling boundary.

Native stage clocks, seconds:

| Stage | Seconds | Share of encoding |
|---|---:|---:|
| Taylor | 200.282 | 16.0% |
| Columns/transposes | 451.469 | 36.0% |
| Rows | 552.214 | 44.1% |
| Buffer work | 14.832 | 1.2% |
| Point products | 25.216 | 2.0% |
| Canonical assembly | 7.190 | 0.6% |

Counters sum within the enclosing native clock. Remaining elapsed includes
allocation/wrappers. Rows/columns now dominate; do not treat memory activity as
only the buffer counter, or these wall-stage fractions as instruction samples.
The load profile contains45.17 s sampled CPU work within262.33 s wall; that alone
does not distinguish storage/cache/page-fault/scheduling delays.

Packed output is8,589,934,592 bytes (**8 GiB**), retained with word-metadata.json
under `.cache/liftrs/cached-s16-encoding/encoding-01/encoding/`. Original
preprocessing is unchanged and referenced read-only. Canonical digest:
`c538a1bc9b408dc4d283aba887f808825404a1c2b8dce693aa7f3015d1152a48`.
Packed digest:
`1c68bfe60f3914b06e31995afc67d5bb3d0a6776c4750af836f8c9794ee18bb9`.
The canonical16 GiB uint32 array is not duplicated on disk; packed output and its
hashes are retained. The controller computed the full packed digest after the
worker's write-time hash; the worker separately checked packed readbacks.

The [analysis](runs/encoding-01/analysis.json) was generated with
`python3 -B experiments/cached-s16-encoding/analyze.py encoding-01 --recorded-cache`.
This bounded mode verifies296 exact source hashes, sealed logs/profiles, small
cache files, input provenance and metadata/digest/physical-size agreement. It
uses the completed controller's packed-payload digest rather than re-reading8 GiB;
the result records packed_payload_digest_recomputed=false. Default analysis still
rehashes all new cache artifacts; hand off that audit as a separate job if it is
expected to exceed30 seconds. New metadata-reconciliation self-checks reject
wrong size/hash/seed; original admission checks, syntax/links and whitespace pass.
Measured runtime code is unchanged during reconciliation, and all cache/profile
bytes remain retained. No new hardware or timing job was submitted.

This closes one experimental cached s16 mathematical encoding checkpoint on the
measured machine. It does not establish an independent full-reference match,
fresh-process repeat equality, a latency distribution, production security,
cryptographic/application execution or TRL5 acceptance. Any repeat or further
CPU/GPU optimization needs its own explicit experiment/handoff; none was started
automatically.
