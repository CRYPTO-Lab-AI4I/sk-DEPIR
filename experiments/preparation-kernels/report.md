# Specialized CPU products and transform crossover

This implements the CPU comparison in the
[current plan](../../docs/encoding-scaling.md#next-comparisons-cpu-products-and-gpu-transform-tiles).
Starting revision `4cb6193` plus exact recorded experiment/integration sources.
No full s16 preparation is attempted. Public field coordinates and file formats
are unchanged. A lookup-only private polynomial leaf path is integrated after
the initial comparison; the global native cutoff remains unchanged.

## Reproduction and boundaries

[run.sbatch](run.sbatch), [shared CPU/GPU controller](run.py),
[Go harness](../../encoding/liftrs-shared/preparation_trials_test.go),
[native kernel](cpu.cpp), [private fixture](fixture.hpp), and
[analyzer](analyze.py) define the measurement. From the repository root:

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/preparation-kernels/slurm-%j.log \
 experiments/preparation-kernels/run.sbatch cpu-01
python3 -B experiments/preparation-kernels/analyze.py cpu cpu-01
# After reproducing the integrated variant under a fresh name:
python3 -B experiments/preparation-kernels/analyze.py compare cpu-01 cpu-02
```

Use fresh run names. `cpu-01` uses dispatch lookup; `cpu-02` includes the integrated
direct leaf loops. Reproduce each using the exact recorded sources, not by
assuming the current tree is the old baseline. Both retain their sources/binaries.
Four CPU cores, 16 GiB requested, 12 GiB sampled RSS guard, 30-minute allocation,
25-minute controller deadline. No GPU in these jobs. Go 1.27.3, GCC 14.2.0,
C++17 `-O3`, Linux; node and topology are recorded per run.

Seed 1501 supplies canonical GF(2^16) inputs at polynomial `0x1002d`. For operand
lengths 32/128/512/2048/8192/32768, compare three timed repetitions of current Go
lookup Karatsuba, experiment-only Go lookup specialization, native matrix
convolution, and standalone C++ lookup Karatsuba. Allocation is included; fixture
creation, correctness checks and table/plan setup are separate. Native convolution
uses four workers, Go/C++ Karatsuba one core. C++ timing excludes per-call cgo:
integrating one native call per tiny product could erase gains and is not tested.
No sanitizer claim is made for the experiment-only C++ binary; all its outputs
are compared exactly with the Go reference. Go kernels retain a CPU profile.

The fixture holds base-field tables, canonical sparse transform metadata and
full-width extension input for the subsequent GPU experiment. SHA-256 and matching
binaries/caches are retained outside Git under `.cache/liftrs/preparation-kernels/`.
The shared controller records all Go/C++/CUDA/runner hashes and snapshots dirty
sources. Cache and evidence hashes are independently verified by the analyzer.

Each run also executes two complete s10,m2,t5 preparations/encodings at each
cutoff 32768/1024/128, four native workers, 4 GiB Go soft memory target, seed 1603,
unique caches. The existing first-use harness verifies P·G, every systematic
position, 16 independent curves, complete hashes and 1,027 packed readbacks.
The analyzer requires identical support/P/spectrum/G and codewords for all six
runs. These pipeline times include profiles, internal diagnostic GC and shared
NFS I/O by phase. Variant order is fixed; no unprofiled or cold-I/O baseline.

## Evidence and selection

[Initial job 72820](runs/cpu-01/results.json) and [analysis](runs/cpu-01/analysis.json)
compare original lookup dispatch with specialization. At n=32768, medians were
0.35762 s dispatch, 0.16135 s specialized Go, 0.12166 s standalone C++ and
0.07634 s native transform. Specialized Go was 2.22× faster than dispatch.
At n=32 the native transform took 0.868 ms versus 0.00243 ms specialized Go:
always using transforms would be a regression. These are different algorithms
and thread counts, not a language-only comparison.

Complete s10 preparation medians before integration were 9.020 s at cutoff
32768, 6.976 s at 1024 and 8.768 s at 128. Moving the cutoff too early loses
some benefit. These samples do not justify a universal cutoff or an s16 setting.

[Integrated job 72824](runs/cpu-02/results.json) and
[analysis](runs/cpu-02/analysis.json) verify the direct loops inside the existing
private preparation lookup path. The generic field path remains the reference;
tests check zeros, uneven operands and truncated prefixes. CPU run nodes/load can
differ, so cross-run pipeline numbers are historical comparisons, not isolated
integration speedups. Full repository regression is in
[validation-03](../hpc-s16/runs/validation-03/results.json).

Both CPU jobs ran on gnode17 (Xeon 6710E). Integrated preparation medians were
7.012 s at cutoff 32768, 6.696 s at 1024 and 9.081 s at 128. The default's
9.020→7.012 s is a sequential same-node observation, not a randomized or load/
frequency-controlled speedup trial. The
[cross-run check](runs/cpu-02/comparison.json) verifies byte-identical fixtures and
complete codewords before/after integration. Both trials verified all six complete
codes; the native C++ outputs match all reference coefficients. Sampled process
group peaks among commands were 301.44/306.64 MiB and retained caches including
fixtures/binaries/constants about 83.18/83.19 MiB. Compiler/build memory is included
in those maxima. Go CPU profiles and native clocks are retained.

Job 72827 (validation-03) passes both full repository suites, focused race checks,
vet, legacy fixture build and fresh small first-use/cache comparisons; its evidence
analyzer passes. Focused generic-versus-private lookup prefix tests also pass with
CGO disabled. Python/shell syntax, gofmt, links and whitespace checks pass.

The selected change avoids repeated field dispatch in Go polynomial leaves and
needs no new native dependency/API. A native batch-product backend and parallel
batches remain possible later. Keep the current global cutoff until measurements
cover the actual field sizes, shapes, workers and full preparation. No catalogue,
protocol, security or full s16 feasibility claim changes.
