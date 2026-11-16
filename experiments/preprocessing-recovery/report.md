# Incremental preprocessing recovery

Scope: cheap, optional safe-boundary progress saving for the longer full s17 GPU
trial. This is correctness/recovery validation, not a runtime benchmark. The
[check script](check.py) records revision plus exact modified source SHA-256
identities because checks precede publication. The [full controller](../gpu-s17-full-preprocessing/run.py)
performs a matching native recovery preflight before starting its GPU workload.

## Method and reproduction

On the HPC login host, using the retained Go1.27.3/cgo native toolchain and cached
build objects, run each bounded check separately:

```sh
python3 -B experiments/preprocessing-recovery/check.py unit
python3 -B experiments/preprocessing-recovery/check.py race
python3 -B experiments/preprocessing-recovery/check.py asan
python3 -B experiments/preprocessing-recovery/check.py vet
python3 -B experiments/preprocessing-recovery/check.py controller
```

Every invocation has25s per-command timeout, four Go workers/build concurrency,
no CUDA dependency and no GPU allocation. JSON results contain exact commands,
source identities, host and outputs: [unit](unit-checks.json), [race](race-checks.json),
[address sanitizer](asan-checks.json), [vet](vet-checks.json),
[controller](controller-checks.json). These timings are test execution only, not
full preprocessing or checkpoint-I/O performance.

`TestPreprocessingRecoveryBoundaries` interrupts native s4,m2,t3 after a completed
sample, root batch, intermediate reciprocal prefix and final reciprocal prefix;
reads the streamed checkpoint into a fresh state; resumes and compares canonical
bytes against independent Go arithmetic. It also exercises multiple occupied and
empty product-tree carry levels plus pending factors, immutable loaded state,
checksum/trailing-byte/size damage and a valid-format wrong parent digest.
`TestCLIRecoveryRestart` uses signal-equivalent stop requests at real phase events,
then a fresh CLI invocation on the cache, ending with completed-file readback.
Retention tests save three generations, require exactly two remain, damage the
newest and recover the previous, then reject two damaged generations.
A deadline-writer test verifies overhead allowance aborts before writing bytes.
Existing preprocessing/reference tests and CLI completed-root restart run too.

Race and address-sanitizer tests cover the changed recovery/CLI paths. LeakSanitizer
cannot run under this sandbox's ptrace: the first unrestricted attempt printed
PASS then a LeakSanitizer fatal diagnostic. Address-sanitizer results explicitly
use `ASAN_OPTIONS=detect_leaks=0`, not a leak-check claim. Earlier broader race/ASan
attempts including all `TestPreprocessing` cases exceeded25s; the targeted reruns
are recorded separately and do not imply a complete sanitizer regression run.

## Costs and limits

No new transform, forced root merge or GPU transfer is introduced. Saves stream
64KiB buffers from borrowed arrays and preserve two atomic complete generations
per phase. CLI write cost is included in profiled workflow latency; the periodic
1% overhead target is adaptive and can be exceeded by blocking filesystem calls.
Full-size write latency remains unknown. `SIGUSR1` requests a forced save/exit75
at the next safe boundary, one hour before the full worker/controller deadline.
An unfinished FFT/native operation cannot be recovered or interrupted safely at
arbitrary instruction boundaries. Full identity/readback checks still run.

At s17, projected extra snapshot payload is about132GiB, plus small leaf/metadata
and filesystem overhead, within the unchanged512GiB retained guard alongside
235.324GiB canonical files. Read/validation and an occasional restored-root copy
add restart costs only; see the [contract](../../docs/preprocessing.md#incremental-recovery).
No full s17 measurement or CUDA-kernel/recovery claim follows from these small
native tests; existing CUDA arithmetic is unchanged. The replacement will use
97h worker/98h controller/100h Slurm, not a promised runtime.
