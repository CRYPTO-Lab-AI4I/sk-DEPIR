# Targeted s16 preparation profile

Status: job **72816 passed**, including all bounded arithmetic checks and sealed
evidence verification. This executes sub-step 1 of the
[expanded preparation plan](../../docs/encoding-scaling.md#current-step-larger-memory-preparation-efficiency)
after publishing checkpoint `ea1ff1e`. The implementation is unchanged; the
opt-in [test harness](../../encoding/liftrs-shared/preparation_profile_test.go)
separates representative work to guide CPU/GPU experiments.

## Reproduction and scope

From the repository root:

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/preparation-profile/slurm-%j.log \
 experiments/preparation-profile/run.sbatch target-01
python3 -B experiments/preparation-profile/analyze.py target-01
```

[run.sbatch](run.sbatch) and [run.py](run.py) request four CPUs, 16 GiB and
30 minutes, with 12 GiB sampled process-group RSS and 25-minute controller guards.
Go soft limit is 4 GiB. No GPU, full s16 scratch allocation, discovery or whole-code
preparation is involved. Maximum transform size is 2^22 points, transpose scratch
16 MiB. Profiles and matching binaries are retained; no reference data is modified.

The source is checkpoint `ea1ff1e` plus recorded profiling-only files. All runtime
Go/C++/header/runner source hashes are recorded; changed source snapshots and the
binary are retained. The runner records CPU/Slurm/OS, GCC and Go versions. Generic
C++17 `-O3` remains unchanged. Compilation runs from an isolated local source copy
with offline dependencies. Per-command clocks include up to one second polling
delay; per-stage clocks exclude that delay and profile shutdown. Stage timings
are profiled, not unprofiled latency baselines. Forced GC before stages is excluded.

The support file in `.cache/liftrs/hpc-s16/s16-m2-t5/` is opened read-only, streamed
through SHA-256, checked against the saved first-run digest and size/metadata,
then sampled at indexes `floor(j*K/65536)` for j=0..65535. This spreads samples
across sorted exponents, avoiding a prefix-only bias. The 3.42 GiB file is read;
only 65,536 exponents are retained. This includes NFS/cache/hash costs separately.
The sample is not itself a full Frobenius-closed code or a complete preparation.

Stages are support read/verification, bounded executor/table setup, the existing
root powers/inverses/conjugates, leaf products in batches of 8,192 factors, merges,
factor verification, and representative convolutions at 2^18/2^20/2^22 points.
Root and polynomial coefficients use the existing GF(2^16) polynomial/tower.
Convolution inputs are deterministic base-field sequences; allocation/initialization
is outside the convolution clock. Convolution includes scratch clearing/allocation,
native forward/inverse transforms, Go pointwise products and output assembly.
Native transform call counts and wall time separate them from surrounding Go work.
Checks are timed separately: each factor vanishes at its root; products agree
at 0,1,omega; convolution coefficients stay in the base field, match 32 direct
low coefficients and agree at 0,1,omega with the product of input evaluations.

## Evidence and limits

Results: [raw run](runs/target-01/results.json), [stage clocks](runs/target-01/stages.json),
[analysis](runs/target-01/analysis.json). The analyzer verifies exact sources,
binary/evidence hashes, stage bounds and check counts. Binaries are outside Git at
`.cache/liftrs/preparation-profile/target-01/`; regenerate with a fresh run name.
No expensive new full-code cache is created; existing support/profiles are kept.

The run used gnode17, Intel Xeon 6710E, four allocated CPU cores, Go 1.27.3 and
GCC 14.2.0 on Linux; see [CPU](runs/target-01/cpu.log),
[allocation](runs/target-01/job.log), [Go](runs/target-01/go.log) and
[compiler](runs/target-01/compiler.log). Worker sampled peak RSS was
110,891,008 bytes (105.75 MiB); the profile worker completed in 49.70 seconds.
Support streaming/hash/sampling took 38.785 seconds and verified the original
digest, separately from compute. Setup took 0.030 seconds. These are single
profiled samples; short-stage CPU attribution has few samples and is noisy.

| Kernel | Work | Seconds | Native transforms, seconds |
|---|---|---:|---:|
| Root generation | 65,536 positions; 32,659 factors | 0.145 | — |
| Leaf products | Four batches, at most 8,192 factors each | 0.805 | 0 |
| Batch merges | 65,316 final coefficients | 0.879 | 0 |
| Convolution | 2^18 transform points | 0.309 | 0.288 |
| Convolution | 2^20 transform points | 1.475 | 1.377 |
| Convolution | 2^22 transform points | 6.589 | 6.193 |

The leaf and merge [profiles](runs/target-01/pprof-leaf-products.log) show
single-core Go multiplication/lookup and Karatsuba dominate their samples.
The sampled merge operands fall below the existing native cutoff of 32,768
coefficients, so all merges remain Go. Compare specialized small-product loops
and the Go/native crossover threshold before assuming a new backend is necessary.
Root generation is smaller than product work for this particular sample; optimize
it after validating product improvements, unless later integrated evidence changes
that ranking. This is a local workload comparison, not full-code stage proportions.

At 2^22 points, native transforms account for 94.0% of the convolution wall time;
the [CPU profile](runs/target-01/pprof-convolution-22.log) agrees at the native/cgo
boundary. That makes large transforms a sensible GPU comparison target. It does
not establish a GPU speedup or identify the dominant native subkernel: no device,
transfer, thread-scaling or native-symbol profile was measured. The inputs are
structured periodic base-field sequences, not a random database workload.

Verification: the analyzer resolved 247 recorded source files and checked binary,
all evidence hashes, support digest, factor/root evaluations, native call counts,
field ranges, convolution evaluations and direct prefixes. Focused unchanged
lookup/scaling tests pass with CGO disabled; the opt-in native profile compiles
and passes. Python/shell syntax and documentation links were checked. No complete
repository suite was repeated because this step changes only opt-in profiling
and documentation; the published full-suite/race/vet checkpoint remains the
production-code baseline.

These differently sized kernels must not be added into a full s16 runtime or used
as full-stage percentages. Native attribution in Go profiles is largely cgo;
the transform clocks complement it. No native hardware profiler is used (previous
allocations denied perf). Large-level merging, reciprocal, identity, full spectrum,
NUMA scaling and GPU transfers require further measurements. This profile cannot
demonstrate full s16 feasibility or GPU benefit.
