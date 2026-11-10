# Bounded preparation arithmetic comparison

The first full s16 run was stopped after its measured progress exceeded the job
budget; see the [run and cancellation evidence](../hpc-s16/report.md). This
experiment tests whether bounded base-field multiplication tables improve its
small polynomial products without changing code coordinates or constants.

Starting revision: `34e1514734abf17cd57d3d094985d45b0c474955` plus recorded working
changes. [run.py](run.py), [run.sbatch](run.sbatch),
[baseline.json](baseline.json), and [analysis script](analyze.py) define reproduction.
The [test harness](../../encoding/liftrs-shared/preparation_benchmark_test.go)
uses the existing first-use validation. Matching binaries and reusable small
constants are retained outside Git under `.cache/liftrs/preparation-arithmetic/`.
Every measured Go file has a recorded hash and retained source snapshot. Native
source/compiler configuration is unchanged between variants; binary hashes,
compiler version, CPU/Slurm topology and job allocation are recorded per job.

## Method and budgets

Run from the repository root:

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/preparation-arithmetic/slurm-%j.log \
 experiments/preparation-arithmetic/run.sbatch profile-01 --profile-only
sbatch --output=experiments/preparation-arithmetic/slurm-%j.log \
 experiments/preparation-arithmetic/run.sbatch compare-01
python3 -B experiments/preparation-arithmetic/analyze.py compare-01
```

Use fresh names. Jobs request four CPU cores, 16 GiB, 30 minutes, no GPU; the
controller guards sampled process-group RSS at 12 GiB and runtime at 25 minutes.
All code runs from isolated copies on node-local scratch. The full s16 cache is
never opened. Local scratch and small shared-cache files are retained. No large
preprocessing is repeated.

The initial profile uses 32,768 monic quadratic factors over GF(2^16), fixed seed
1301 and the current polynomial `0x1002d`. It measures the existing Go leaf tree
below the 32,768-operand native-transform cutoff. Factor allocation is outside
the timed region; polynomial arithmetic/allocation and CPU profiling are inside.
This is a representative kernel, not the full root loop or a valid whole code.

The matched comparison measures the same leaf product with and without lookup,
including separate table setup time. Baseline then optimized each execute three
complete first-use s10,m2,t5 runs, four native workers, seed 1603 for discovery,
4 GiB Go soft memory target, one encoding per run. Each run independently discovers,
saves/loads constants, prepares the code, verifies the full P·G identity, encodes,
checks every systematic position and 16 independently evaluated curves, hashes,
packs/fsyncs and reads back 1,027 positions. Every run has its own cache. Stage
times include CPU profiling and diagnostic GC checkpoints. Shared NFS persistence
and cached OS I/O are included; no unprofiled or cold-storage claim is made.

Baseline restores only the original `scaling_experimental.go` from its exact
snapshot. The common field-dispatch branch exists in both comparison builds,
with a nil table in baseline. The independent initial profile predates that
branch. Thus this measures the lookup choice with otherwise matched code.
Variant order is fixed, not randomized; CPU/thermal/load effects remain possible.

## Results

[Profile job 72810](runs/profile-01/results.json),
[CPU attribution](runs/profile-01/baseline-pprof.log), and
[kernel measurement](runs/profile-01/baseline-leaves/leaves.json) record an
8.278-second leaf product: 84.1% of sampled CPU time is in table-free
`BinaryField.Mul`, 95.6% cumulatively in Karatsuba. This identifies a kernel
bottleneck; the interrupted full s16 preparation profile is unavailable.

The matched comparison is in [job 72811](runs/compare-01/results.json), with
derived correctness/timing evidence in [analysis.json](runs/compare-01/analysis.json).
Both variants passed. Hardware was gnode17, dual Intel Xeon 6710E (128 physical
cores visible), with four allocated CPUs, GCC 14.2.0, Go 1.27.3 linux/amd64,
generic native C++17 `-O3`, and Linux. See recorded
[CPU](runs/compare-01/cpu.log), [compiler](runs/compare-01/compiler.log),
[Go](runs/compare-01/go.log) and [allocation](runs/compare-01/job.log).

| Measurement | Table-free baseline | Lookup | Baseline / lookup |
|---|---:|---:|---:|
| s16 leaf kernel, one profiled sample, seconds | 8.095 | 2.279 | 3.55× |
| s10 preparation median, three samples, seconds | 14.022 | 8.970 | 1.56× |
| s10 sum of measured phase times, median, seconds | 17.120 | 11.962 | 1.43× |
| s10 complete worker process, median, seconds | 17.883 | 13.160 | 1.36× |

The leaf table setup took 0.000745 seconds, separate from its kernel time. Complete
preparation timings include table setup. Complete-worker clocks include profile
shutdown, GC outside individual phases and process exit, but also up to roughly
one second of controller polling delay. Builds are separate. Sampled worker peak
RSS was 45.59 MiB baseline and 43.98 MiB lookup; this small difference is not a
memory-saving claim. The small constants and matching binaries total about
72.5 MiB in the retained comparison cache.

All six complete runs found K=146,641, N=1,048,576. Support, P/spectrum, G,
canonical output and packed output match exactly across variants/repeats. The
analyzer rehashes every constant and verifies the recorded full-code checks.
Both shared/native package suites passed in each variant. Full repository suites,
race/vet and fresh small-cache admission are recorded separately in
[validation-02](../hpc-s16/runs/validation-02/results.json).
Job 72812 passed both complete repository suites, focused native race checks,
vet, legacy fixture build, fresh small first-use/cached comparison and evidence
verification. Local independent-oracle/table-lifetime tests also pass with CGO
disabled. Shell/Python syntax, documentation links and whitespace checks pass.
No full s16 speedup or completion forecast follows from these smaller-code timings.

## Implementation and limits

[preparation_products.go](../../encoding/liftrs-shared/preparation_products.go)
builds logarithm and duplicated power arrays using the existing field polynomial.
A primitive generator is found explicitly; element 2 need not be primitive.
Zero multiplication is handled separately. Arrays use less than 768 KiB at s16,
are immutable for the private preparation session, and are cleared from the
prepared basis before online use. No artifact or field coordinates change.
General public field arithmetic remains table-free. This input-indexed arithmetic
is not constant-time and is restricted to public code preprocessing values.

The focused arithmetic test compares all pairs through s8, every input through
s16 against boundary/random partners, and a polynomial with nonprimitive element
2 against an independent bit-arithmetic oracle. Existing scaling tests compare
P, G, spectrum and complete words to the reference and check that tables do not
escape into online state. The analyzer requires identical public constants and
canonical/packed output hashes across all six complete runs.

This portable change adds no dependencies. Native leaf kernels, more parallel
preparation and GPU variants remain possible follow-ups if measurements justify
their implementation/deployment cost. They have not been measured here. Discovery,
root exponentiation, large transforms and output I/O have separate costs; faster
leaf multiplication cannot eliminate them. Full s16 feasibility remains open.
