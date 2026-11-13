# Direct CPU packed output

CPU-042 follows the validated [GPU output](../gpu-packed-output/report.md) and
[host-backed GPU route](../gpu-host-encoding/report.md). It adds explicit
`ScalingCPPFileEncoder.EncodePackedTo` on the verified fast file-streaming CPU
schedule. Resident/owning defaults remain. CPU disk-backed transforms are next.

## Reproduction and evidence

[Controller](run.py), [Slurm script](run.sbatch), [analyzer](analyze.py),
[byte-boundary/resource self-check](check.py),
[API](../../encoding/liftrs-shared/cpu_file_packed.go),
[native implementation](../../encoding/liftrs-cpp/c_api.cpp),
[tests](../../encoding/liftrs-shared/cpu_file_packed_test.go).

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-packed-output/slurm-%j.log experiments/cpu-packed-output/run.sbatch packed-01 --output-mode packed
# Only after user reports completion:
python3 -B experiments/cpu-packed-output/analyze.py packed-01
```

[Local compile/self-check evidence](compile-check.json) passed; runtime validation
**passed job74131**, run `packed-01`, submitted2026-11-12.
[Results](runs/packed-01/) and [Slurm log](slurm-74131.log) are populated by the
job. After the completion ping,389 source identities and all cache/evidence
hashes reconciled; [analysis](runs/packed-01/analysis.json). HEAD ce29557 plus recorded working sources; each
run records exact revision, source hashes/snapshots, commands and environment,
results, profiles and cache checksums. New retained cache
`.cache/liftrs/cpu-packed-output/packed-01/` holds native/ASan binaries,
reasonably sized code fixtures and four matched packed files. Constant subcode
fixtures are generated once and retained; the s6/s8 fixtures come from sealed
GPU-038 job74094. Trusted cached s10 preprocessing is reused from
`.cache/liftrs/gpu-products/products-02/cpu-0/online.bin` with full SHA256 checking.
No large preprocessing/profiles are deleted or regenerated.

One shared-memory CPU allocation,16CPU,16GiB RAM request,12GiB sampled process-
group RSS guard,8GiB Go soft limit, no GPU request or CUDA dependency for this
path. GCC14.2/Go1.27.3, Slurm module25.05. New cache/scratch guards16GiB each;
initial node availability32GiB/runtime headroom16GiB. Controller20min, individual
checks5–10min, Slurm30min. File cache, caller-retained results and runtime remain
extra; RSS is not total host memory. Inputs/output artifacts here are small.
Job guards execute internally; no agent monitoring before completion ping.

## Change and invariants

After the final inverse, transform scratch is dead. Native code zeroes its
packed prefix and scatters s-bit canonical values directly into it. A primitive
orbit generator guarantees distinct destinations; neighboring fields can share
bytes, so relaxed atomic byte OR joins their disjoint bits. Each thread writes
at most3 bytes per value for s<=16. Byte operations make the representation
independent of host endianness. Joined per-chunk parity supplies coordinate0.
All coefficients must be base-field values before emission; invalid values fail
without writing a sink. Ordinary canonical uint32 assembly remains unchanged.

The Go API holds its encoder lock through native execution and bounded copied
writes/hash. It retains/reuses both workspace arrays on sink calls; ordinary
EncodeWithStats still detaches its output. Mixed calls cannot mutate earlier
owning results. Metadata is checked before/after encoding under the existing
immutable-file contract. Writers must consume bytes synchronously and not
reenter; failures can leave partial destinations and an empty full-result digest.
Sync/Close/publication remain caller responsibilities.

Small tests cover constant words for s2..8 (all canonical positions, including
cross-byte fields) and retained nonconstant s6/s8 codes against Go. Each runs
with1/16 workers across nonzero/zero/repeated messages, buffer1/7/257 bytes,
short/failed writers, invalid buffers/messages, recovery, owning-result isolation,
workspace/read counters and Close. AddressSanitizer repeats native cases. Ordinary
Go/CPP tests and combined API build/vet guard compatibility. The generic CPU
packing reference was moved unchanged from the GPU cached test into a shared
CPP test helper; its historical measurement boundary is retained.

Matched cached s10 runs owning→CPU pack/write/hash and direct packed/write/hash
in that fixed order, two repetitions each. Both generate one packed file per
call from the same seed71 message; all four output files are independently
hashed. Pipeline timing includes compute, writer and hash; a separate inclusive
clock adds Sync/Close. Setup checks/loads the trusted file separately. CPU profiles
are retained for small/ASan/matched phases. This is not a repeated randomized
latency study; no full s16 timing or lower-memory claim follows from s10 alone.

## Memory, time and storage trade-off

Fast transforms still require two full uint32 N arrays (8N bytes); there is **no
reduction in that peak**. The packed sink avoids a detached canonical uint32
result and separate CPU packing pass. It reuses dead scratch rather than
allocating a new N result on the next call. At s16 these two arrays remain32GiB,
message3.422GiB plus native tables/runtime and file cache; the active packed
scratch prefix is8GiB within an already counted16GiB array. Parity counters cost
roughly4ceil((N-1)/65536) bytes. A1..64MiB writer buffer is additional, while the
caller/file cache can allocate more. A caller retaining earlier owning words
still adds4N per word. No full canonical uint32 word is built in the sink.

Direct scatter adds atomic-byte writes and zero fill; their runtime cost is
unknown pending measurement. It stays explicit `--output-mode packed`; faster
owning/resident paths remain defaults when they fit. Packed output storage is
ceil(Ns/8),8GiB at s16, unchanged. File-code reads remain4(K+1+N) logical bytes
per encoding, potentially served from OS cache, not necessarily physical I/O.
No new production disk workspace or code format is introduced. CPU disk-backed
transforms, which reduce transform RAM itself by extra disk I/O, remain a separate
pending gate with their own memory/time admission.


## Measured result

Job74131 passes all18 small-code cases (Go comparison,1/16workers, writer/input/
recovery/owning checks) and the ASan repeat, ordinary builds and vet. All four
s10 packed files match SHA25612604f70b336d895f092799963d3c572d70267b92975040b01174c70479f6fda.
Retained workspace8MiB for direct calls plus1MiB bounded writer; transform peak
unchanged. Source/evidence/cache reconciliation passes389 source identities.

| Profiled s10 pipeline | Repeat1 seconds | Repeat2 seconds |
|---|---:|---:|
| Owning plus pack/write/hash |0.339971|0.327777|
| Direct packed/write/hash |0.324573|0.327359|
| Owning including Sync/Close |0.346018|0.332050|
| Direct including Sync/Close |0.328187|0.331913|

Two fixed-order samples do not establish a reliable speedup. Direct native
assembly6.06/7.11ms versus owning2.16ms shows the atomic-write cost; avoiding
the separate packing pass offsets it here. Both read4780872 logical code bytes
in2chunks per encoding. Peak controller-sampled group RSS1053720576bytes is
across the whole validation/build job, not isolated encoding memory. File cache
and caller-retained owning words remain extra. Profiles/fixtures retained.
CPU disk-backed transforms are the remaining authorized implementation gate;
no full s16 direct-sink timing or s17 feasibility follows from this result.

Final accounting: gnode17, COMPLETED0:0,2m30s, Slurm MaxRSS112008KiB;
[accounting](runs/packed-01/accounting.log). This per-task sampling view differs
from the controller process-group peak and is not an encoder-only memory bound.
