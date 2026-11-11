# Native CPU scratch/output reuse

Status: implemented; validation submitted as Slurm73586, full results pending. The GPU checkpoint was published as
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
an analytical target, not measured RSS. During complete encoding there is one
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
remain pending. s17 still requires wider arithmetic and a different memory
schedule; this is not application/cryptographic integration or TRL5 acceptance.
