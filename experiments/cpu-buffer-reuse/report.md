# Native CPU scratch/output reuse

Status: job73586 completed; full validation and evidence reconciliation pass. The GPU checkpoint was published as
999f4a5 before this CPU change. The user authorized the same memory reduction on
the native CPU encoder and correctness/memory/time validation. The Go reference
and hybrid path keep their existing separate-output schedules; CUDA is not needed.

## Ownership and memory change

[Complete native adapter](../../encoding/liftrs-shared/scaling_matrix_cpp.go)
now uses its N-element transpose scratch as the final canonical output. The last
inverse completes before scatter begins; scatter reads the separate convolution
arena, never its output. After success the workspace drops its scratch reference
and transfers ownership of the returned slice to the caller. Its next call lazily
allocates fresh scratch. Earlier returned words remain valid through later
encodes, hybrid calls, invalid-call recovery and encoder closure. Failure retains
scratch privately and returns no word. Workspace instances remain single-caller.

[Go binding](../../encoding/liftrs-cpp/plan.go) admits exact scratch/output
identity only. All other overlaps, including partial views, remain rejected.
[C ABI contract](../../encoding/liftrs-cpp/c_api.h) documents this lifetime;
there is no arithmetic/transform change in native execution. The public methods
and owning-output semantics stay the same. `ScalingCPPWorkspace.Bytes()` now
reports less owned storage after transfer; the caller's word is excluded as before.
The hybrid path allocates scratch again if needed and still returns a separate word.

For s16,m2,t5,K918,672,421,N2^32,L2^31, principal arrays fall from
`16N+12K+4` to `12N+12K+4` bytes: **74.267 to58.267 GiB**, saving16 GiB. This is
an analytical target; completed RSS measurements are recorded below. During complete encoding there is one
shared scratch/output allocation, not separate ones. If the caller retains old
words, each still takes16 GiB; the saving is not a bound on arbitrary retained
results. The benchmark releases the previous word and performs diagnostic GC
before the next phase, as in the baseline harness. Other runtime/allocator/thread
memory is measured separately.

No transforms, field operations, copies or host/device transfers are added. A
replacement scratch allocation on later calls replaces the old separate-output
allocation; first call uses workspace storage already allocated during setup.
Allocation/GC/locality can still change latency, so it is measured explicitly.
Preprocessing, canonical coordinates, masks and security contracts are unchanged.

## Reproduce and preserve evidence

[Runner](run.py), [Slurm launcher](run.sbatch), [analyzer](analyze.py),
[self-check](check.py), [ownership/worker tests](../../encoding/liftrs-shared/cpu_memory_test.go),
[overlap tests](../../encoding/liftrs-shared/scaling_matrix_cpp_test.go) and
[native sanitizer fixture](../../encoding/liftrs-cpp/online_test.cpp).
From the repository root:

```bash
python3 -B experiments/cpu-buffer-reuse/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-buffer-reuse/slurm-%j.log experiments/cpu-buffer-reuse/run.sbatch reuse-01
# After the user's completion ping:
python3 -B experiments/cpu-buffer-reuse/analyze.py reuse-01 --recorded-cache
```

The runner records actual revision plus exact dirty-source snapshots, source and
binary hashes, commands, CPU affinity/model, OS/compiler/Go versions, stage clocks,
profiles, sampled RSS and sealed evidence/cache digests. Source before the edits
is999f4a5; the actual measured source identity is the manifest, not that commit
alone. The allowed differences from passed GPU job73576 are the native binding,
C ABI comments/contract, sanitizer fixture and shared matrix adapter; CPU arithmetic
and GPU production files must otherwise match. Build requires Go1.27.3, GCC14.2.0,
cgo and `liftrs_cpp`; no CUDA build tag/toolkit/GPU is requested.

Results: `runs/reuse-01/results.json`, phase JSON, profiles and logs; allocation
log `slurm-JOB_ID.log`. Retain test/sanitizer binaries, small fixtures and one new
8 GiB packed word under `.cache/liftrs/cpu-buffer-reuse/reuse-01/`, outside Git.
The runner uses read-only links to saved s6/s10/s16 online inputs. The s16 file is
20,854,558,905 bytes with SHA256
`f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874`.
No code discovery/preparation is rerun or old cache/profile removed. Only completed
controller hashing reads the final8 GiB payload; `--recorded-cache` analysis later
checks its saved digest, metadata and length without another bulk reread.

## Validation and measurement

Focused tests check complete words against existing native/Go/dense references,
retained owning outputs, input immutability, repeated/zero words, recovery after
invalid inputs, exact versus partial aliasing and separate concurrent workspaces.
Native suite, CGO-disabled reference tests, focused Go race and vet run inside the
job. ASAN/UBSAN runs the whole native online fixture with both separate and aliased
output at1/4 workers, preserving independent constant-word expectations. Go race
is not a detector for native data races; ASAN/UBSAN do not prove their absence.

Matched s10,m2,t5 baseline/reuse workflows use32 workers and two encodes each,
fixed order on the same allocation. The baseline helper retains the old distinct
output schedule using the same five-transform native kernel. Report setup,
allocation-inclusive encoding, validation, packing and readback separately;
small fixed-order samples do not establish general timing equivalence.

After those checks, run s6 smoke and two complete s16 reuse encodings with one
workspace. Each checks all K systematic values,16 independent sampled curves,
repeat canonical digest and exact workspace ownership transfer. The final word
is packed at16 bits/value and1027 positions are read back. Both canonical digests
and final packed digest must match [CPU job73131](../cached-s16-encoding/report.md)
and the already reconciled GPU reference. Exact16 GiB ownership transfer is
recorded in per-call metadata; the RSS guard also tests actual memory reduction.
Full-size latency against prior job73131 is historical, not matched; no default
full-size baseline rerun is silently added.

One CPU-only allocation:32 cores,128 GiB requested RAM, **68 GiB RSS guard**,
64 GiB Go soft limit,32 GiB minimum node headroom and128 GiB initially available.
Two-hour Slurm limit,110-minute controller,90-minute full worker;5-minute small
workers. Cache/scratch guards16 GiB each with20 GiB initially free. CPU profiles
stay enabled for setup/encoding/checks/I/O. No scheduler/process/log monitoring
after submission until the completion ping; automatic guards remain active.

Pre-submission focused ownership, overlap, full-word and concurrency tests pass
with `go test -tags liftrs_cpp ...` (0.138 s test execution; build separate).
Python checks reject wrong metadata/digests, ownership/retention/repeat records
and verify the16 GiB model. Sanitizers, broader checks and full RSS/time evidence
are recorded below. s17 still requires wider arithmetic and a different memory
schedule; this is not application/cryptographic integration or TRL5 acceptance.


## Completed result

[Results](runs/reuse-01/results.json), [analysis](runs/reuse-01/analysis.json),
[phases](runs/reuse-01/encoding/phases.json),
[sanitizer output](runs/reuse-01/sanitizer.log) and
[final accounting](runs/reuse-01/accounting.log) pass review:328 exact source
identities, sealed evidence/cache digests, ownership-transfer metadata, full CPU/GPU
canonical and packed digest matches. Both full encodes check every systematic
position and16 sampled independent curves, then1027 packed readbacks pass.
Native/Go-only suites, focused race/vet, alias/retained-output/error/concurrency
tests and ASAN/UBSAN pass. No source correction was needed from this validation.
Recorded-cache analysis checks packed metadata/length and the completed controller's
8 GiB payload digest without a fresh bulk reread. Other cache/evidence digests are
recomputed. Large caches, packed output and profiles remain retained.

Actual source revision999f4a5 plus exact source snapshots in the manifest.
Hardware gnode17, Xeon6710E,32 CPU workers, no GPU; GCC14.2.0, Go1.27.3.
Allocation completed in **1h04m36s**, exit0:0, no guard tripped, below2h limit.
This agrees with the50–75-minute forecast. Full worker3640.7 s includes both
encodes, verification and I/O. Controller also runs builds/tests/profiles.

| Memory | Prior CPU | Reuse CPU |
|---|---:|---:|
| Principal modeled arrays, GiB | 74.267 | 58.267 |
| Peak sampled process-group RSS, GiB | 74.370 | 58.351 |
| Slurm MaxRSS | — | 58.37G |

Sampled peak decreases16.019 GiB (21.54%), consistent with one16 GiB allocation
removed plus small runtime differences. The run stays below68 GiB RSS guard.
Each successful Encode detaches exactly4N bytes, and replacement scratch on the
second call does not retain the first output through the workspace. Output words
remain caller-owned; keeping extra results still adds their storage. This result
applies to the native complete path, not new memory savings in Go/hybrid execution.

| Phase | Seconds |
|---|---:|
| Load/check online cache | 93.362 |
| Native setup/workspace allocation | 0.086 |
| First encoding | 1281.328 (21m21.3s) |
| First independent verification | 470.721 |
| Second encoding | 1247.173 (20m47.2s) |
| Second independent verification | 469.931 |
| Pack/write8 GiB | 57.337 |
| Packed readback | 1.806 |

Setup allocation reserves largely untouched pages; its short wall time does not
mean subsequent memory commitment is free. Prior job73131 encoded once in1252.503 s
on an earlier allocation of the same node/model. Current calls are2.30% slower
and0.43% faster respectively; this historical comparison does not isolate a time
penalty or prove equivalence.

Matched s10 fixed-order32-thread schedule comparison, two calls each:

| Layout | First encode, seconds | Second encode, seconds |
|---|---:|---:|
| Separate output baseline | 0.255081 | 0.237360 |
| Scratch/output transfer | 0.261056 | 0.235616 |

Differences have both signs (+2.34% first,-0.73% second), with too few samples to
establish a stable runtime penalty/benefit. Same five transforms and no new
arithmetic/copy passes are retained. Report the mixed samples, not a cost-free
or faster-encoding claim. Native stages still dominate total latency.

This closes selected CPU scratch/output validation. Public owning-output semantics
and ordinary CPU-only builds are preserved. Further in-place/streamed CPU/GPU
routes remain separate memory–time choices; a known slower mode must stay
capacity-driven or explicitly selected on the command line when memory fits.
No further job, new budget or default selection for a different route follows
this results review. s17 and application/cryptographic/TRL5 gates remain open.
