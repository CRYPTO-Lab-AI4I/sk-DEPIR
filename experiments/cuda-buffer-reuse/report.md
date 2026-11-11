# CUDA two-work-array encoding prototype

Status: opt-in prototype passed full validation in Slurm73576. The user chose
buffer reuse after the [memory audit](../encoding-memory/report.md). The public
`WithCUDA` default retains the verified resident layout; CPU remains an independent
main path. No s17 support or general runtime equivalence is claimed.

## Implementation and expected memory change

[CUDA engine](../../encoding/liftrs-cuda/matrix.cu) adds `gm_create_reuse`, selected
by [NewBoundedReuse](../../encoding/liftrs-cuda/bridge.go). The shared constructor
mode is private to opt-in tests, not a new public selection option. Old libraries
without this optional symbol fail clearly for reuse; their default ABI is unchanged.
This mode is online-only; primitive Apply/Product requests are rejected.

Two padded prefix operands share an N-sized arena; require2L<=N and reject other
shapes explicitly. Their transforms borrow an independent N-sized transpose
buffer. After prefix interpolation, keep the first K values in the arena and
clear only the tail, avoiding the old saved-H copy. The full transform/product/
inverse reuses that arena. After it finishes, transpose scratch is dead and
becomes canonical output; parity/error words live in a separate small allocation.
The canonical scatter never overlaps its live input. Download remains to an
independently owned host array, so the existing caller contract is preserved.

Five transforms, code constants, field/coordinate rules and host transfers are
unchanged. One internal device copy is removed. No additional arithmetic,
recomputation or PCIe passes are introduced. Changed addresses/allocation behavior
may affect runtime; measure that effect rather than assuming equal speed.
Host plan retention is unchanged in the verification harness.

At exact s16 K918,672,421,N2^32,L2^31, predicted owned device memory is
`72,394,976,680 - 4N + 4(ceil((N-1)/4096)+1)` bytes, including the same tables/plans
as the saved baseline. This is about51.427 GiB versus67.423 GiB: nearly16 GiB
saved, with a roughly4 MiB parity/error buffer. A56 GiB cap tests the reduction;
the old layout cannot fit that cap. Measured results below confirm this prediction. A straightforward s17 extension would still need about398 GiB;
this prototype alone cannot make s17 fit this GPU.

## Reproduce

[Runner](run.py), [Slurm launcher](run.sbatch), [analyzer](analyze.py),
[admission self-check](check.py), [worker](../../encoding/liftrs-shared/cuda_cached_test.go)
and [small oracle checks](../../encoding/liftrs-shared/cuda_online_test.go).
Run from the repository root:

```bash
python3 -B experiments/cuda-buffer-reuse/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/cuda-buffer-reuse/slurm-%j.log experiments/cuda-buffer-reuse/run.sbatch reuse-01
# After the user's completion ping:
python3 -B experiments/cuda-buffer-reuse/analyze.py reuse-01 --recorded-cache
```

Results go to `runs/reuse-01/results.json`, per-phase JSON, profiles and logs;
allocation output is `slurm-JOB_ID.log`. Keep code fixtures, shared library, test
binary and one new8 GiB packed word in `.cache/liftrs/cuda-buffer-reuse/reuse-01/`.
No preparation/cache rebuild; small inputs link to retained s6/s10 files, and full
s16 input is the verified20,854,558,905-byte online file with SHA256
`f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874`.
The source revision before changes is24bbad9; the runner seals the actual revision,
exact dirty source snapshots, compile commands and all binary/evidence/cache
hashes. Only four production source paths are allowed to differ from GPU job73257:
CUDA source/header/Go bridge and the private shared constructor adapter. Changed
tests are recorded, and CPU production arithmetic remains identical.

## Validation and timing boundaries

CPU-only, CGO-disabled, CUDA-only, combined encoding tests and vet run inside the
allocation. Historical80-record matrix regression exercises ordinary primitives.
Small full-word tests compare the default path at s2/3/4/6/8 to independent Go,
and reuse at s6/8 (cover2L=N and2L<N). s2/3/4 intentionally exercise unsupported
prefix rejection. Nonzero/zero/nonzero calls, invalid length/field, recovery,
input immutability, primitive rejection, old-library rejection, double-close and
released memory are checked. CUDA memcheck/leak checks repeat the reuse cases.
Small-mode allocation savings must exactly equal4N minus parity/error bytes.

A matched s10,m2,t5 session runs default then reuse on the same GPU allocation,
two same-message encodings each. This is a fixed-order comparison, not randomized
or enough samples for broad claims. Setup, transfer-inclusive encoding, independent
verification, packing and readback are separate. Profiling is enabled for both
paths. Profiles and stage clocks are retained; Go cgo profiles do not identify
GPU instruction hotspots.

Only after these checks pass, run s6 smoke and two full s16 reuse encodings in one
session under56 GiB. Both full canonical hashes and final packed hash must equal
[CPU job73131](../cached-s16-encoding/report.md). Check every systematic value,
16 independent sampled curves per repeat and1027 packed readbacks. Exact device
allocations must match the saved default allocation count minus the expected
saving. Full-size default timing comes from [job73257](../cuda-s16-encoding/report.md)
and is historical across allocations; do not label it a matched s16 speedup.

## Budgets and pending evidence

One GPU/16 CPU cores,128 GiB requested host RAM,96 GiB sampled RSS guard,
64 GiB Go soft limit, full reuse56 GiB explicit device cap (below the prior96 GiB
budget),512 MiB cap for small cases. Initial GPU availability98 GiB and node
availability128 GiB checks retain prior headroom; no higher budget is requested.
Minimum32 GiB available node RAM while running; new cache and scratch capped at
16 GiB each with20 GiB free initially. Two-hour allocation,110-minute controller,
90-minute full worker,5-minute smoke/matched workers,10-minute small/memcheck
workers. Guard failures stop dependent work and preserve available evidence.
No scheduler/process monitoring after submission until the user pings.

Before submission, CUDA sm90 compile-only check and CUDA-only Go
admission/lifecycle/matrix-metadata checks pass. The compile-only library, exact
source/header snapshots and hashes are kept under
`.cache/liftrs/cuda-buffer-reuse/build-check/`; no GPU execution occurred there.
Python self-checks cover rejected metadata/repeat/digest/cap/device-saving errors.
Completed runtime, memory and latency results are recorded below. No timing result is
inferred from the analytical savings. Packed output remains16 bits/value and
prototype side-channel/security, application/cryptographic and TRL5 limits remain.
`--recorded-cache` analysis uses the completed controller's packed-file digest
without a fresh8 GiB reread; source/evidence digests, physical length and metadata
are still checked. All previous caches/profiles are retained.


## Completed validation

[Results](runs/reuse-01/results.json), [analysis](runs/reuse-01/analysis.json),
[full phases](runs/reuse-01/encoding/phases.json),
[memcheck](runs/reuse-01/memcheck.log) and
[accounting](runs/reuse-01/accounting.log) pass reconciliation:323 exact source
identities, sealed evidence/cache hashes, exact device allocation difference,
all small checks and both full CPU canonical/packed digest matches.
The allocation completed in19m05s, exit0:0; no guard tripped. The earlier35–50-minute
forecast was conservative: independent verification was much faster in this run.

The source revision was24bbad9d06775f16cfd512ef7f134219408873e9 plus exact source
snapshots in the manifest. Hardware was gnode10, H200 NVL143,771 MiB,
driver580.126.20, Xeon6710E,16 allocated cores,128 GiB requested RAM;
GCC14.2.0, CUDA12.9.1, Go1.27.3. The prior full default run used gnode17, so
cross-run latency differences cannot be attributed solely to this change.

| Device allocation | Bytes | GiB |
|---|---:|---:|
| Prior default | 72,394,976,680 | 67.423 |
| Reuse, both full calls | 55,219,301,804 | 51.427 |
| Saved | 17,175,674,876 | 15.996 |

This is **23.725% less owned device memory**, exactly one16 GiB array minus
4,194,308 bytes parity/error storage. It fits the56 GiB explicit cap. Driver/context
allocations are excluded from owned memory. Host sampled RSS39.031 GiB
(Slurm39.05G) remains about the same because the validation harness retains the
host plan/input/output. Host-plan release was not part of this implementation.

| Full s16 phase | Seconds |
|---|---:|
| Load/check prepared online file | 309.386 |
| Reusable device setup | 3.701 |
| First encoding | 57.418 |
| First independent verification | 157.852 |
| Second encoding | 56.975 |
| Second independent verification | 157.138 |
| Pack/write | 53.932 |
| Packed readback | 1.267 |

Worker wall805.8 s includes both encodes/checks and profiling boundaries.
Encoding includes host allocation and message/final-word transfers; load,
setup, correctness auditing and packing remain separate. The previous default
s16 calls were85.954/60.216 s. This is historical evidence, not a matched s16
speedup: in particular, first-download time changed from31.837 s to3.788 s,
and independent validation/load times also changed substantially. Second-call
native wall54.758 s versus prior55.368 s; Go output clearing accounted for
about2.21 s versus prior4.85 s. Do not infer that buffer reuse caused these changes.

The matched fixed-order s10 comparison on this allocation is more direct:

| Layout | Setup seconds | Encoding1 seconds | Encoding2 seconds |
|---|---:|---:|---:|
| Default | 0.537891 | 0.103897 | 0.103143 |
| Reuse | 0.532620 | 0.103885 | 0.103060 |

Encoding differs by less than0.1% here: **no observed runtime penalty** in this
small comparison. Two calls and fixed order are insufficient for a general
performance-equivalence claim. Both paths match full canonical/packed hashes.

CPU-only, CUDA-only, combined and CGO-disabled tests and vet pass;80 historical
matrix records pass. Default small Go comparisons pass at s2/3/4/6/8. Reuse s6/8
passes repeated/zero/error-recovery comparisons, primitive/old-ABI rejection and
close/release; s2/3/4 correctly reject unsupported prefix sizes. CUDA memcheck
reports0 errors and0 leaked bytes. Full reuse checks all918,672,421 systematic
positions and16 sampled curves on each call; final8 GiB packed output passes
1027 readbacks and the retained CPU hash. Profiles/preprocessing/output are kept.

The selected buffer-reuse validation step is complete. Keep it opt-in until a
separate integration decision; the public default and its admission estimate
remain unchanged. Further s17 feasibility still needs a substantially different
memory schedule (~398 GiB principal arrays under this schedule). In-place code
streaming versus host/disk-backed tiles remains a separate explicit memory–time
choice. No new job, larger budget or public-default switch follows this review.
