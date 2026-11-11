# Bounded GPU canonical output

Status: opt-in compact online prototype implemented; validation submitted as Slurm73815, runtime results pending.
The user authorized this next gate after publishing CPU/GPU validation1800e0d.
Public CPU/GPU defaults remain unchanged. The previous direct-row study established
only transform scratch savings; this prototype addresses the later output peak.

## Memory schedule and correctness invariants

[CUDA engine](../../encoding/liftrs-cuda/matrix.cu),
[optional ABI/Go binding](../../encoding/liftrs-cuda/bridge.go),
[private shared constructor](../../encoding/liftrs-shared/cuda_available.go) and
[worker](../../encoding/liftrs-shared/cuda_cached_test.go) select a new explicit
compact mode. It supports current s2..16 arithmetic and2L<=N. Prefix operands
share one N-sized uint32 arena; direct rows avoid full transpose scratch. Code
P/spectrum stay resident and immutable. No new transform, code recomputation or
message upload is added, but extra work below is real and must be measured.

After the final inverse, all used coefficients must lie in the base field.
For s<=16 they fit uint16. Compact each ascending chunk into an independent
staging allocation, then copy it to the arena's lower half. At most2 MiB staging
is allocated (2^20 values). All source reads for a chunk complete before its copy;
the compressed destination cannot overwrite any future chunk's uint32 input.
Default-stream ordering enforces this dependency. Tail coefficient N-1 is unused
by canonical assembly and is stored as0, consistent with the prior schedule.

The lower half then holds N uint16 source values; the disjoint upper half holds
N uint16 canonical values. Primitive-generator scatter preserves coordinate
order, parity and unique destinations; a small separate parity/error allocation
prevents scratch/output overlap. Range checks occur **before narrowing**; otherwise
invalid extension bits could disappear. Both halves fit the original4N-byte arena.

Download the canonical upper half into the first2N bytes of the caller-owned
N-element uint32 array. Expand backwards, decoding bytes rather than violating
host strict-aliasing rules. Stores lie above all unread packed input; no second
host array is needed. The returned numeric uint32 word, ownership and packed-file
format are unchanged. This download is uint16 storage, not the final s-bit file
format for s<16. Final shared packing remains s bits/value.

At exact s16 N2^32,K918,672,421, expected owned device memory is the validated
reuse count minus4N plus2 MiB staging:
**38,041,529,772 bytes (~35.429 GiB)** versus55,219,301,804 (~51.427 GiB).
This includes identical tables/plans and separate parity, excludes driver/context,
and is a model until measured. The new40 GiB cap is below the old56 GiB cap.
Host plan/input/owning output remain approximately39 GiB in the validation harness;
no host-plan release or streaming API is introduced.

## Explicit mode and costs

The [runner](run.py) defaults to `--memory-mode resident` (validated reuse path).
Testing this potentially slower mode requires **`--memory-mode compact`**.
When the resident path fits admitted budgets it remains the default. Compact is
an opt-in experiment/capacity solution, not an unconditional speed improvement.
Main `WithCUDA` has no new automatic selection or production memory option.
Low-level `NewBoundedCompact` requires a library exposing its optional ABI;
older libraries fail clearly rather than silently falling back.

Costs to report separately:

- Direct row access: prior component measurements showed roughly2–8% larger-case
  penalties; full-domain behavior remains unmeasured.
- GPU compaction: an extra read/write/staging pass,4096 ascending chunks at s16.
  Stage storage is bounded; approximate data traffic is10N bytes (40 GiB at s16)
  before canonical scatter, not another full GPU allocation.
- Bulk download:2N bytes instead of4N (8 versus16 GiB at s16), plus the existing
  4-byte range flag. Message upload is unchanged.
- Host expansion: an extra backwards CPU pass reading2N/writing4N bytes,
  with actual time measured. It may outweigh some transfer savings.

`CompactStats` records GPU compaction seconds, host expansion seconds and bulk
download bytes for the just-observed successful call. Buffer-stage totals already
include compaction; do not add overlapping metrics to infer total latency.
Native total includes expansion, and outer Encode includes Go allocation and
telemetry overhead. CPU/GPU profiles/stage clocks are retained. No cost is treated
as zero, and no default will switch based solely on saved bytes.

## Reproduce and preserve evidence

[Runner](run.py), [launcher](run.sbatch), [analyzer](analyze.py),
[self-check](check.py), [small oracle tests](../../encoding/liftrs-shared/cuda_online_test.go)
and [boundary fixtures](../../encoding/liftrs-shared/cuda_compact_test.go).
Run from repository root:

```bash
python3 -B experiments/gpu-output-memory/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-output-memory/slurm-%j.log experiments/gpu-output-memory/run.sbatch compact-01 --memory-mode compact
# After the user's completion ping:
python3 -B experiments/gpu-output-memory/analyze.py compact-01 --recorded-cache
```

Revision before edits1800e0d. The runner records exact actual revision/dirty source
snapshots, input/binary/evidence/cache digests, commands, CPU affinity, actual
host/GPU/driver/compiler/Go versions. It builds the library for allocated hardware
with CUDA12.9.1 `nvcc -O3 -lineinfo -arch=native`; GCC14.2.0/Go1.27.3/cgo and
cpp,cuda tags build the comparison worker. The primary CUDA-only API stays
independent of C++; this benchmark uses the shared native-clock/reference harness.

Results: `runs/compact-01/results.json`, phase JSON, profiles and logs; allocation
log `slurm-JOB_ID.log`. Retain library/executables, boundary-code spectra,
matched small fixtures and one8 GiB packed output under
`.cache/liftrs/gpu-output-memory/compact-01/` outside Git. Existing real s6/s10/s16
inputs are read-only links; no discovery/root/reciprocal preparation rerun.
s16 online input20,854,558,905 bytes, trusted SHA256
`f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874`.
Boundary P/G fixtures are deterministic and their code files/specifications and
checksums are retained, including memcheck fixtures. Profiling is enabled and
signal handlers flush active profiles before exit where available.

## Checks and comparison boundaries

First run build compatibility, existing default matrix regression, small resident/
compact oracle comparisons, repeated-zero-error recovery/owning-output/ABI/close
checks and memcheck. Prefix-heavy small shapes explicitly reject2L>N.

Additional K1 constant-word fixtures at s6/10/11 exercise chunks smaller than,
equal to and larger than staging capacity. P=1+X and G=sum X^j to degreeN-2
satisfy P*G=1+X^(N-1); every canonical value independently equals the message.
Test0,1,maximum base value and run these fixtures under memcheck as well. A legal extension bit16
inserted into a spectrum at s11 must trigger the range flag before uint16 could
hide it. These are arithmetic/assembly fixtures, not discovered catalogue codes
or security evidence. Real s10/full s16 use verified catalogue preprocessing.

Matched s10 resident/candidate workflows run two calls each in fixed order on one
allocation, checking full canonical/packed agreement. Then s6 smoke and two full
s16 compact calls reuse one engine. Every full call checks all systematic values,
16 sampled independent curves and repeat/canonical hashes, with1027 packed
readbacks and final tight8 GiB packed digest matching CPU/GPU references.
Device count must exactly equal the prior reuse count minus4N plus staging.
Compare full timing to job73576 only as historical; no full resident rerun is
silently added. Loading/setup/checks/packing/readback remain separate from Encode.

## Admission and pending evidence

One GPU16 CPU cores,128 GiB host request/96 GiB sampled RSS,64 GiB Go soft,
40 GiB explicit device cap for compact (56 GiB resident),512 MiB small cases.
Initial free device memory must exceed selected cap by2 GiB (42 GiB compact);
this is a guard, not a claim of measured feasibility on any GPU SKU. Initial
node RAM128 GiB, minimum headroom32 GiB; new cache/scratch guards16 GiB each,
20 GiB free initially. Allocation2h/controller110m/full worker90m;
small/memcheck/boundary cases have individual5–10m limits. Guard failure stops
later work and preserves available evidence. No agent monitoring after submission.

Before submission: CUDA sm90 compile-only and combined Go admission/lifecycle/
matrix-description tests pass (0.011 s test execution; builds separate).
Compile-only source/header/library/hash provenance retained in
`.cache/liftrs/gpu-output-memory/build-check/`, no GPU execution there. Python
checks validate metadata/repeats/caps, exact compact counts, half-download bytes
and nonnegative additional stages. Runtime correctness, memcheck, actual memory
and time costs remain pending. Recorded-cache analysis uses controller hashing
for the8 GiB file, while rechecking metadata/length and other sealed evidence.

s17 requires wider extension values and at least17-bit base storage; this uint16
prototype does not support it. Code/spectrum residency remains a separate memory
bottleneck. Application/cryptographic encoding, side-channel security and TRL5
acceptance remain unproven; current variable-time arithmetic/contracts are unchanged.
