# Guarded Linux s16 first use

Status: audit/controller checks and admission validation job 72766 pass.
Full job **72768** was stopped at the user's request after **8h58m32s** on gnode02.
Discovery completed in **2h56m37.545s**, yielding K=918,672,421; its support file
is retained and its checksum reverified. Preparation reached 124,780,544 of
918,672,421 scanned positions (13.58%) in the root-product loop. Recent progress
suggested about 36 additional hours for that loop alone, exceeding the initial
21-hour worker budget. That is an extrapolation, not a completion forecast.
Sampled peak process-group RSS was 85,962,891,264 bytes (80.06 GiB).
See the [external-stop record](runs/first-01/interruption.json) and
[recording script](record-stop.py); the original interrupted `results.json` still
says `running` and is preserved as historical evidence. The analyzer checks the
stop record first. No complete s16 preparation or encoding result is claimed. This
is standalone mathematical encoding, separate from masks/application execution.

Slurm TERM was sent to the batch shell, which ended before the Python controller
could finalize or flush the active preparation profile (zero bytes). Completed
phase/heap profiles survive. The batch script now uses `exec python3` so future
batch signals reach the controller directly. The saved discovery checksum
is `021446aa0d2764f4020e989ee54f3a99c63c7d9a407f5babbcb333919a2854ad`.
No cache was deleted and no full retry was launched. The
[preparation arithmetic comparison](../preparation-arithmetic/report.md) follows
up with bounded kernel profiles and smaller complete runs.

Starting revision: `34e1514` plus the recorded experimental harness/guard changes.
Keep the [development handoff](../../docs/development-handoff.md) and
[s16 representation status](../../docs/encoding-s16.md) separate from this admission.

## Source-based audit and policy

Run [audit.py](audit.py) with `python3 -B experiments/hpc-s16/audit.py` to reproduce
[audit.json](audit.json). Inputs and exact source hashes are recorded there.
It uses catalogue-estimated K=919,440,829 at s=16,m=2,t=5,N=2^32 and also evaluates
the admitted capacity ceiling K=N/2. That ceiling is a resource scenario, not a
capacity prediction. Discovery must establish exact K before preparation; a
capacity above that envelope stops explicitly.

| Principal payload / files | Estimated K, GiB | Meaning |
|---|---:|---|
| Discovery with both index lists and illustrative capacity slack | 53.05 | Go int lists, random bytes, grid, transpose and membership bits |
| Root-product illustration | 41.13 | Bounded flat factors, retained roots and convolution arrays |
| Identity / late Newton arrays | 86.85 | Including shared transpose and FFT workspace |
| Newton with old reciprocal and support capacity slack | 96.56 | Conservative additional lifetime scenario |
| Prepared arrays plus executor | 56.56 | Includes retained support and transpose |
| Complete online arrays | 74.28 | Constants, input, aliased workspace, transpose and canonical output |
| Code files plus two packed output fixtures | 51.43 | Initial and fresh cached outputs retained separately |

These are array/file calculations, not measured RSS bounds. They exclude GC-
retained unreachable objects, allocator/runtime/native overhead and OS activity.
At the K=N/2 ceiling the conservative Newton scenario is about 108 GiB and online
arrays about 104 GiB; the cache remains below 64 GiB. Prefix power-of-two jumps
are included. Factor batching and scratch reuse are already implemented; the
historical map/leaf model is not silently reused as current measured evidence.

[policy.json](policy.json) requests **32 CPU cores, 256 GiB and 24 hours**, no GPU.
Use a **192 GiB Go soft target**, **224 GiB sampled process-group RSS guard**,
**64 GiB new-cache cap**, **128 GiB scratch cap** and a **23.5-hour runner deadline**.
The initial worker has at most 21 hours and the fresh-cache worker two hours;
the remaining interval covers builds, hashing/finalization and scheduler margin.
The process group receives SIGTERM on a guard stop (allowing profile flushing),
then SIGKILL after 60 seconds. Worker test timeouts are slightly longer than their
controller deadlines. Slurm's requested memory was not enforced on prior jobs,
so the controller does not rely on it. Samples can miss transient peaks; this is
not a hard allocation limit. No automatic retry or cache deletion is permitted.

Admission additionally needs at least 320 GiB host MemAvailable, 128 GiB free local
scratch and 80 GiB free shared cache space. Stop if node MemAvailable drops below
32 GiB. These are node-health checks, not a reservation of all visible host memory.
An exclusive cache lock prevents concurrent writers; source provenance is checked
before reusing completed constants. A changed source fingerprint stops for review.

Runtime is unmeasured. Domain/level extrapolation from the profiled s15 encoding
gives roughly 50 minutes for one s16 encoding, excluding loading/validation/output;
this is not a worker-scaling prediction. Preparation may take many hours. Historical
s15 preparation spans a manual pause and cannot supply a clean transferred latency
baseline. The 24-hour request is an exploratory cap, not a completion forecast.
The 32 workers parallelize native matrix execution; monomial enumeration, support
scans, root generation and Go traversal of the base-field grid remain serial.

## Reproduction and verification

[run.py](run.py) reuses the existing first-use mathematics through
[TestS16FirstUse](../../encoding/liftrs-shared/s16_firstuse_probe_test.go) and the
shared private first-use helper. Normal suites skip the opt-in case. Default
public discovery remains s≤13; only experimental preparation admits s16.

First run the controller checks and small validation:

```sh
python3 -B experiments/hpc-s16/test_run.py
module load slurm/slurm/25.05
sbatch --cpus-per-task=4 --mem=16G --time=00:15:00 \
 --output=experiments/hpc-s16/slurm-%j.log \
 experiments/hpc-s16/run.sbatch validation-01 --validate
```

Validation runs focused race/vet, both full suites in an isolated source copy and
fresh first-use/cached GF(64),m=2,t=4 cases. Support, P/spectrum, G and packed/canonical
digests must match the retained s6 reference. Guard tests cover memory/cache/
scratch/deadline decisions, process cleanup, symlink accounting and allocation-free
count/payload arithmetic. No large arrays are allocated in these checks.

After validation passes on the same sources:

```sh
sbatch --output=experiments/hpc-s16/slurm-%j.log \
 experiments/hpc-s16/run.sbatch first-01 --admission validation-01
```

Use fresh run names; output directories are never overwritten. The full job
refuses changed admission sources. Results/progress live in `runs/<name>/`;
exact source snapshots and the matching binary are retained. The controller
automatically runs a fresh cached process only after first use passes. It seals
evidence before reporting `passed`, writes a summary and appends a timestamped
result message to the repository board. A failure is retained as partial evidence.

[Admission results](runs/validation-01/results.json) and
[analysis](runs/validation-01/analysis.json) pass full suites, race/vet, complete
small first-use/cached checks and byte-identical retained reference constants.
Use `python3 -B experiments/hpc-s16/analyze.py first-01` for live status; after
completion, [analyze.py](analyze.py) verifies sealed evidence, sources, constants,
initial/fresh output digests and phase clocks without repeating preprocessing.

## Measurement and persistence boundaries

Separate executor setup, exact discovery, support persistence, reusable preparation,
prepared-data persistence/release, loading, online setup, one encoding, validation,
packing/fsync and readback. Preparation checkpoint GC/heap profiles are diagnostic
overhead, as in the retained s15 harness. Go CPU profiles stay enabled; denied perf
access is recorded separately. Native clocks and stage counters remain available.
Profiles and the RSS monitor interfere with timing; no unprofiled latency claim is made.

Completed code data is written **directly to shared NFS** under
`.cache/liftrs/hpc-s16/s16-m2-t5/` with streaming SHA-256, fsync and atomic names.
Support serialization no longer makes a K-element uint32 copy. This I/O boundary
differs from the cached s15 local-XFS experiments: cache loading and packed output
here include shared-filesystem effects. Packing includes fsync. NFS can cache data;
this is not a cold-storage throughput measurement. Node-local scratch retains the
isolated sources/build cache. The policy accounts for disk use separately from RAM.

An interrupted discovery must restart; completed support survives. Incomplete
preparation is not resumable, but completed support is reused. Completed online
constants are reused without regeneration. Source-machine and s13/s14/s15 caches
remain untouched. Fresh cached output has its own directory/symlinks and cannot
overwrite the initial output; both fixtures fit the retention policy.

Acceptance requires exact discovered support, Frobenius closure, the full P·G
identity, every systematic symbol, 16 independently evaluated degree-5 curves,
complete repeat/cached canonical and packed digests, and 1,027 packed readbacks.
These checks have stated independent/sampling limits and do not constitute
production cryptographic qualification. No catalogue/TRL claim changes automatically.
