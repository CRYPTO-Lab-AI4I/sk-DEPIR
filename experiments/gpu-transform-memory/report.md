# GPU transform scratch reduction

Status: job73669 completed; component correctness, scratch reduction and timing reconciliation pass.
The user authorized GPU-first work after publishing4b08075. This study runs
independently of CPU job73586, whose state/logs/results are not inspected here.

## Choice and implementation

The fast layout transposes into a full-size scratch array, transforms the rows,
then transposes back. The direct layout transforms each row in its original
contiguous region, with one thread owning each row. Column and Taylor schedules,
affine row offsets, field values and coordinate order are unchanged. Direct mode
removes both transpose passes and the full scratch allocation, but adjacent threads
access distant addresses; loss of coalescing may outweigh those savings.

[CUDA kernel/engine](../../encoding/liftrs-cuda/matrix.cu) adds explicit
transform-only constructors, exposed as
[NewTransformBounded](../../encoding/liftrs-cuda/bridge.go). Fast mode owns only
data+scratch; direct mode owns data only. Neither allocates an unused second
convolution operand, so comparison isolates exactly one scratch array. Existing
ordinary/online/reuse constructors, public WithCUDA and CPU defaults are retained.
Transform-only engines reject Product/SetOnline, and older libraries missing the
optional ABI fail clearly for these constructors.

**This is not complete encoder memory reduction.** At a hypothetical full s16
transform it would remove16 GiB scratch, but canonical output assembly still
needs an independently safe destination. The existing reuse online encoder uses
dead scratch for that output. Eliminating scratch during transforms alone does
not eliminate its full online peak; streaming/routing or another output schedule
is a separate gate. No s17 implementation/capacity or online speedup is claimed.

## Reproduce and explicit slower-mode selection

[Runner](run.py), [launcher](run.sbatch), [analyzer](analyze.py),
[metadata self-check](check.py) and
[worker](../../encoding/liftrs-shared/gpu_direct_rows_test.go). From repo root:

```bash
python3 -B experiments/gpu-transform-memory/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-transform-memory/slurm-%j.log experiments/gpu-transform-memory/run.sbatch rows-01 --layout compare
# After the user's completion ping:
python3 -B experiments/gpu-transform-memory/analyze.py rows-01
```

The CLI defaults to `--layout fast`. `--layout direct` or `--layout compare`
explicitly opts into the potentially slower memory-saving experiment when both
layouts fit. The worker defaults to fast as well; its environment selector is
set by the CLI. This respects the documented policy: retain the faster validated
path whenever it fits; a slower mode is a capacity solution, or an explicit
command-line opt-in when memory is sufficient. No automatic online selection or
new production CLI memory option is introduced.

Revision before edits4b08075; exact runtime revision, dirty source snapshots,
source/binary hashes, command arguments, environment, CPU affinity/model, GPU,
driver, compiler and Go versions are saved in the manifest. Module versions are
Slurm25.05, GCC14.2.0 and CUDA12.9.1, with local Go1.27.3 and cgo. The library is
built for the allocated GPU with `nvcc -O3 -lineinfo -arch=native` and shared
PIC output. CPU reference uses the maintained native matrix implementation;
small cases additionally compare independent Go sparse-transform output.

Results go to `runs/rows-01/results.json`, transform records, CPU profiles and
logs; allocation output `slurm-JOB_ID.log`. Keep the executable/library, serialized
plans and largest input/reference arrays under
`.cache/liftrs/gpu-transform-memory/rows-01/`, outside Git. Records include seed,
shape via the source, input/reference digests and plan files. No polynomial/code
preprocessing is required: reusable matrix plans are retained and the input is
regenerated from deterministic seed3200+dimension. Largest fixture pair totals
256 MiB. Previous online preprocessing/profiles remain intact.

## Method and acceptance

Use s16 extension arithmetic, base polynomial0x1002d and the canonical tower field.
Domains2^8,2^9,2^16,2^17,2^20,2^21,2^24,2^25 cover even/odd dimensions and square/
rectangular decomposition. Each mode performs three forward/inverse pairs from
the same random input, fixed order fast then direct. Native reference preparation
and input allocation/copying/digests/checks are outside measured Apply calls.
Each forward word must equal the native result and each inverse equal the original
input. Small2^8/2^9 cases additionally compare the Go result.

Report reusable engine/plan setup separately. Apply wall time includes Go/cgo
boundary, message upload, kernels and final download; CUDA stages separately
record Taylor, columns, rows, transposes and transfers. Direct transpose time
must be zero. Setup/first-call versus reused behavior is retained rather than
hidden in a single ratio. Median of three repeats is descriptive, not a statistical
guarantee. Go CPU profiles include setup/reference/checking as well as cgo time;
they do not locate GPU instruction hotspots.

Owned allocation counts for matching shapes must differ by exactly4N bytes.
At the largest tested domain, working arrays are256 MiB fast versus128 MiB direct,
plus identical small tables/plan storage. Owned counts exclude CUDA context/driver.
This is a component bound, not total client memory. Separate retained input/
reference and temporary CPU-reference arrays contribute to host RSS.

CUDA memcheck/leak checks run small square/rectangular cases. Tests also reject
unsupported primitive/online calls, reject old-library selection, verify closure
and zero owned bytes. Existing default/reuse small online oracle checks run as
regressions. Combined and CUDA-only admission/lifecycle/metadata builds check
feature independence. Controller failure stops the workflow; no full encoding
job or follow-up is automatically submitted.

## Resource admission and pre-submission checks

One GPU,16 CPU cores,16 GiB requested host RAM,12 GiB RSS guard,4 GiB Go soft
limit and512 MiB explicit device cap. No second engine remains allocated across
comparisons. New cache2 GiB and scratch4 GiB guards,6 GiB initially free disk,
32 GiB initially available node RAM and16 GiB minimum node headroom. Allocation
30 minutes, controller25 minutes, worker20 minutes. Automatic guards/profile
flush remain active; no agent monitoring until the user's completion ping.

Before submission, CUDA sm90 compile-only and combined Go admission/lifecycle/
matrix-metadata checks pass (0.012 s test execution; builds separate). Compile-only
library, exact source/header snapshots and hashes are retained in
`.cache/liftrs/gpu-transform-memory/build-check/`; it has not executed on GPU.
Synthetic self-checks reject wrong shape/cap/scratch counts/digests and nonzero
direct transpose time. Completed GPU correctness/memory/timing are recorded below; the direct mode
has a measured runtime penalty at meaningful tested sizes. Any later slower online mode remains capacity-driven or
explicitly selected, with transfer/output-storage costs measured before adoption.


## Completed result and selection decision

[Results](runs/rows-01/results.json), [analysis](runs/rows-01/analysis.json),
[transform records](runs/rows-01/transforms/transforms.json),
[memcheck output](runs/rows-01/memcheck.log) and
[accounting](runs/rows-01/accounting.log) pass.333 exact source identities and all
saved cache/evidence digests reconcile. Square/rectangular forward/inverse
results agree with native references; small Go comparisons, primitive/old-ABI
rejection, closure and default/reuse online regressions pass. Memcheck reports
zero errors and0 leaked bytes. Direct transpose time is zero and memory saving
is exactly4N bytes in every comparison. No code correction needed from the run.

Revision4b08075e18b77b8316fe70690bac76adb6120b9a plus exact source snapshots.
Hardware gnode14,H200 NVL143,771 MiB, driver580.126.20, Xeon6710E,
16 allocated CPU cores,16 GiB requested RAM, GCC14.2.0/CUDA12.9.1/Go1.27.3.
Allocation2m14s, exit0:0, no guard. Controller sampled process-group RSS peaks
at0.691 GiB across commands; Slurm reports0.31G (different aggregation/sampling).
Full transform worker41.45 s; builds/setup/memcheck/profiles account for the rest.
CPU job73586 was not inspected during this experiment; it was independently
reviewed after the user confirmed both jobs finished.

Median of three transfer-inclusive calls, fixed fast-then-direct order:

| Values | Forward fast/direct, s | Direct forward penalty | Inverse fast/direct, s | Direct inverse penalty |
|---|---|---:|---|---:|
| 2^16 | 0.008715 / 0.009159 | 5.10% | 0.008554 / 0.009149 | 6.95% |
| 2^17 | 0.014667 / 0.015149 | 3.29% | 0.014444 / 0.015057 | 4.25% |
| 2^20 | 0.051555 / 0.054966 | 6.62% | 0.050958 / 0.054867 | 7.67% |
| 2^21 | 0.086742 / 0.090093 | 3.86% | 0.085752 / 0.089721 | 4.63% |
| 2^24 | 0.340776 / 0.354364 | 3.99% | 0.333155 / 0.351417 | 5.48% |
| 2^25 | 0.583994 / 0.596274 | 2.10% | 0.572089 / 0.588210 | 2.82% |

Tiny2^8/2^9 cases are faster in direct mode, where avoided launch/transpose costs
matter; do not extrapolate those ratios to large arrays. Setup is separate
(~0.31–0.32 s per layout at2^25), and includes CUDA/plan allocation. The measured
shape-specific penalty is broadly2–8% in the larger tested cases. Lost coalescing
is a plausible explanation, not an instruction-level profile result. Three
repeats and fixed order do not bound performance at untested sizes/hardware.

At2^25 values, fast working arrays256 MiB become128 MiB direct: **50% less working
storage**, exactly128 MiB saved. Owned totals269,233,796 versus135,016,068 bytes
include identical798,340 bytes tables/plan, excluding context/driver. Both fit the
512 MiB cap, so this experiment intentionally used explicit `--layout compare`.
**Keep fast as the default.** Direct remains a capacity-driven option or explicit
`--layout direct`/`compare` opt-in when sufficient memory exists; do not silently
adopt this slower mode for general encoding.

This demonstrates the transform component only, through2^25 values using s16
arithmetic. It does not establish2^32 execution or online peak-memory reduction.
The current online reuse encoder still needs safe canonical output storage;
using a full output array later can recreate the same peak. Next design gate is
bounded canonical assembly/output (or a validated in-place permutation), followed
by chunked code/spectrum residency. Extra host scatter/routing, transfers and disk
storage must be quantified and chosen before an end-to-end slower mode is adopted.
A fulls16 integration/benchmark and s17 wider arithmetic remain unverified.
Profiles/fixtures kept; no automatic next run or default change from this review.
