# Fast-layout GPU encoding with streamed constants

Status: GPU-036 validation job73940 completed; evidence reconciliation passes. The user chose
this route after the [allocation/width audit](../streamed-constants/report.md).
Source checkpoint f9581d3 precedes these edits; the actual revision plus exact
source snapshots are recorded by the runner. This is explicit `--memory-mode
streamed`; default resident reuse and public WithCUDA remain unchanged.

## Implementation and expected trade-off

[CUDA engine](../../encoding/liftrs-cuda/matrix.cu) adds an optional streamed
constructor and synchronous per-call-code ABI, declared in
[gpu_api.h](../../encoding/liftrs-cuda/gpu_api.h) and loaded by
[Go binding](../../encoding/liftrs-cuda/bridge.go). Missing symbols fail clearly;
no fallback changes modes. SetOnline validates trusted immutable code and installs
shape/generator, allocating no device P/spectrum. Every Encode borrows P/spectrum
only until return. CUDA retains no host pointer. The private
[shared adapter](../../encoding/liftrs-shared/cuda_available.go) retains those
immutable arrays in Go and releases references on Close; ordinary WithCUDA does
not gain a host-plan dependency. Code identity/mathematical validity is the same
verified preprocessing contract, not established by low-level field-range checks.

Upload the first K P coefficients directly into the prefix arena after clearing
padding. The message occupies the other L-element region;2L<=N remains required.
Run the same three prefix and two full transforms using fast transposed rows.
Between full forward/inverse, borrow the first256 MiB of now-dead transpose
scratch for spectrum chunks. Synchronous copies and joined pointwise kernels
complete each read before the next upload overwrites the tile. The inverse then
reclaims full scratch. After inverse, ordinary canonical scatter writes the
scratch buffer, with separate parity/error storage. Download a full owning uint32
word. No direct-row access or base compaction/host expansion is added.

Both low-level code arrays are borrowed per call, not saved by C/CUDA. Output/code
alias is rejected in Go. Low-level calls validate code lengths/ranges; the shared
verified plan remains immutable. Failed calls return no public word, and a later
valid call can reuse the workspace. Closing the engine leaves old owning outputs
valid. Current s<=16/count/field limits are unchanged; this is not s17 support.

At s16,m2,t5,N2^32,K918,672,421,L2^31, the model predicts
**34,364,742,932 bytes (32.005 GiB)** owned device memory, exactly measured reuse
55,219,301,804 minus P/spectrum20,854,558,872 bytes. Tables/plans/parity stay the
same. Tile uses existing scratch and adds no GPU allocation. Compare measured
resident51.427 GiB and compact35.429 GiB, and the completed owned count matches it exactly.
Driver/context and pageable-copy staging remain outside OwnedBytes. Host plan,
message and owning output stay near the prior39 GiB RSS; no host/disk-backed code
or RAM reduction is claimed. Packed file remains8 GiB. No transform is added.

Per encode, code upload is20,854,558,868 bytes (19.422 GiB): first K P values and
full spectrum. P[K] is verified during setup but unused in the current prefix
algorithm. This adds code uploads on every reused encoding instead of amortizing
one setup upload. Existing message upload3.422 GiB and full16 GiB output download
remain. There are64 spectrum chunks of256 MiB at s16. Code-copy seconds are a
subset of H2D; do not double-count. Kernel/copy/launch/range-check costs and cold
page effects must be measured. Hypothetical8/16/32 GiB/s code-transfer bandwidth
would give2.43/1.21/0.61 s payload time, not an encoding-time prediction.

## Reproduce and retain evidence

[Runner](run.py), [Slurm launcher](run.sbatch), [analyzer](analyze.py),
[self-check](check.py), [small Go oracle](../../encoding/liftrs-shared/cuda_online_test.go),
[boundary tests](../../encoding/liftrs-shared/cuda_stream_test.go) and
[cached worker](../../encoding/liftrs-shared/cuda_cached_test.go).

```bash
python3 -B experiments/gpu-streamed-constants/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-streamed-constants/slurm-%j.log experiments/gpu-streamed-constants/run.sbatch streamed-01 --memory-mode streamed
# After the user's completion ping:
python3 -B experiments/gpu-streamed-constants/analyze.py streamed-01 --recorded-cache
```

Runner defaults to `--memory-mode resident`. It builds allocated-device-native
CUDA with nvcc12.9.1, GCC14.2.0, Go1.27.3 and cgo. CUDA-only and ordinary CPU builds
remain independent; cpp,cuda tags enable comparable native clocks in this harness.
Record CPU/GPU/OS/compiler/affinity, commands, revision, exact source/binary hashes,
stage clocks, profiles, process-group sampled RSS and owned device bytes.
Dirty source bytes are retained in source-snapshots. Results/logs/profiles live in
`runs/streamed-01/`; allocation log `slurm-JOB_ID.log`.

Retain reusable executables/library, fixtures and one8 GiB output under
`.cache/liftrs/gpu-streamed-constants/streamed-01/`, outside Git. Link read-only
saved s6/s14/s16 inputs; no discovery or preparation rerun and no deletion of prior
caches/profiles. The s16 input is the verified20,854,558,905-byte file from job72875,
SHA256 `f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874`.
Retained s14 input is `.cache/liftrs/parallel-s14/s14-m2-t5/online.bin`, SHA256
`88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8`.
The loader verifies expected whole-file identity; matching hashes alone do not
establish mathematical code validity, which comes from completed preprocessing.

## Validation and measurement boundaries

Small resident/streamed cases s2/3/4 reject oversized prefix shapes; s6/8 compare
full Go words, nonzero/zero/repeated calls, invalid input/recovery, retained words,
closure and unsupported primitive/old-library calls. New K1 s4 identity fixtures
independently yield constant words and exercise tiles1/64/255/256 over N256:
one-element, exact multi-tile, partial tail and one-full-tile scheduling. Check
exact upload bytes/chunks/tile and unchanged owned allocation/host source. Reject
zero/oversized tiles, missing/short code, output aliases and bad P/spectrum ranges;
verify recovery and owning outputs across Close. These parameterized boundaries
exercise the same loop; the full s16 checks exercise production256 MiB tiles.
Both oracle and boundary fixtures run under CUDA memcheck with leak checking.

Run80 default matrix regressions, ordinary CPU/CUDA-only/combined/CGO-disabled
suites, focused Go race and vet. Go race does not instrument GPU accesses. Then
matched retained s14 resident and streamed workflows run two encodes each, fixed
order on one allocation under4 GiB GPU cap, checking complete canonical/packed
agreement. This is a more meaningful transfer size than s10, not a silent change:
s14 uploads about1.19 GiB code and uses4 spectrum chunks per encode. Two fixed-
order samples do not establish a latency distribution. Setup, first/reused encode,
checks and I/O are separate. Profiles remain enabled.

Finally s6 smoke and two full s16 streamed encodings with one session check every
systematic position,16 sampled independent curves, repeat/full CPU digest matches
and1027 packed readbacks. Actual owned bytes must equal prior reuse minus exactly
P/spectrum; upload bytes/chunks/tile must match4(N+K)/64/256 MiB. Full timing versus
prior reuse/compact is historical; no full baseline rerun is silently added.
Controller seals the packed payload hash; recorded-cache analysis uses that hash
with metadata/size checks rather than another8 GiB rehash. No application masking,
partitioning, side-channel security or TRL5 acceptance is established here.

## Budget and current checks

One GPU16 CPUs,128 GiB requested RAM/96 GiB RSS guard,64 GiB Go soft limit,
**40 GiB device cap** streamed (resident default56 GiB). Require2 GiB additional
initial free GPU headroom,128 GiB initial available host RAM and32 GiB minimum
node headroom. New cache/scratch guards16 GiB each,20 GiB free initially. Two-hour
allocation,110-minute controller,90-minute full worker; automatic resource guards
and profiling remain active. Estimate30–45 minutes allocation, uncertain until
streaming is measured. No budget increase or s17 admission. Submit and yield;
no scheduler/log/process monitoring until the user completion ping.

Pre-submission combined Go lifecycle/admission/matrix-description tests pass
(0.012 s execution), CUDA sm90 compile-only passes. Retained compile artifact and
source hashes: `.cache/liftrs/gpu-streamed-constants/build-check/`. Python admission,
packed/repeat, exact memory-delta and transfer/chunk/stage self-checks pass. Python
AST, bash syntax and whitespace checks pass. GPU runtime/memcheck/matched/full memory and timing subsequently pass below. Resident stays default when admitted memory fits;
streaming remains an explicit optional capacity route.

Job73940 was reviewed after the completion ping; allocation log `slurm-73940.log`.
No new job is submitted by this results review.


## Completed result

[Manifest](runs/streamed-01/results.json), [analysis](runs/streamed-01/analysis.json),
[phases](runs/streamed-01/encoding/phases.json),
[oracle memcheck](runs/streamed-01/memcheck.log),
[boundary memcheck](runs/streamed-01/boundary-memcheck.log) and
[final accounting](runs/streamed-01/accounting.log) pass review. The analyzer
resolves350 exact source identities at f9581d3 plus recorded dirty-source
snapshots, checks sealed evidence/cache digests and validates retained CPU/GPU
canonical/packed hashes. Recorded-cache mode uses the completed controller's8 GiB
payload digest and rechecks metadata/physical length without a fresh bulk reread.
Accounting was saved after completion and is outside the earlier sealed map.

Both full encodes check all918,672,421 systematic positions and16 independent
sampled curves;1027 packed readbacks pass. Small Go oracle/retained-output/error/
old-ABI checks, K1 tile1/64/255/256 boundary checks, legal-extension/non-base product
rejection, source immutability, recovery and Close pass. Both CUDA memchecks
report0 errors and0 leaked bytes. All80 matrix regressions, ordinary CPU,
CUDA-only/combined/CGO-disabled builds, focused Go race and vet pass.

Hardware gnode03, H200 NVL143,771 MiB, driver580.126.20,16 CPU cores,
CUDA12.9.1/GCC14.2.0/Go1.27.3. Slurm completed exit0:0 in **16m22s**, well below the
30–45-minute conservative forecast and2h limit. No guard tripped. Full cached
worker took533.2 s; independent verification was substantially quicker on this
allocation than earlier runs. Cache/load/verification variation prevents using
total validation allocation time as encoding latency. Profiles/caches are retained.

| Metric | Previous resident reuse | Streamed |
|---|---:|---:|
| Owned GPU memory, GiB | 51.427 | 32.005 |
| Peak sampled host RSS, GiB | 39.031 | 39.029 |
| Full first encoding, s | 57.418 | 63.487 |
| Full second encoding, s | 56.975 | 58.851 |
| Code upload per encoding, GiB | 0 after setup | 19.422 |

Owned GPU count34,364,742,932 bytes matches the model exactly. Saving is
20,854,558,872 bytes (19.422 GiB, **37.77%**), with unchanged tables/plans/parity.
This fits40 GiB device cap; driver/context/staging remain additional. Host usage
is essentially unchanged. Each call records20,854,558,868 code bytes,64 spectrum
chunks and268,435,456-byte scratch tile. The4-byte difference from memory saving
is P[K], verified but not used by prefix encoding. Packed output stays8 GiB;
no repeated disk reads of code or extra persistent output are introduced.

Code-copy time is **2.234/2.239 s**, about8.7 GiB/s effective transfer, within the
audit's hypothetical bandwidth range. H2D including the existing message is
2.629/2.633 s. Rows23.651/23.710 s and columns20.581/20.572 s retain fast-layout
behavior; there is no55 s direct-row stage. Native totals56.939/56.991 s exclude
outer Go allocation/telemetry overhead. Whole encoding exceeds them by6.55/1.86 s;
these differences are not independently attributed to one allocator/runtime cause.
Do not add code-copy time again to H2D/native totals.

Full timing above compares separate allocations (prior gnode10, current gnode03),
so it is historical rather than a measured matched s16 penalty. The observed
whole-call differences are+10.57%/+3.29%; they do not define a general distribution.
Compared with the previous compact93.3–101.5 s, the new samples are much nearer
resident timing, but that comparison is historical too.

Matched s14, same allocation, fixed resident-then-streamed order:

| Mode | First encoding, s | Second encoding, s |
|---|---:|---:|
| Resident reuse | 5.1790 | 5.1759 |
| Streamed | 5.5273 | 5.4666 |
| Observed penalty | 6.72% | 5.62% |

Resident/streamed owned memory3.190/2.000 GiB respectively. Code-copy time0.1375 s
per encoding for1.190 GiB through4 chunks. Two fixed-order samples establish an
observed cost, not a general latency distribution. Load is cold/warm biased
(24.64 versus1.78 s); it is excluded from encoding comparisons and not a claimed
streaming speedup. Setup is1.390/1.249 s. Report setup amortization and first/reused
encoding separately; the streamed mode pays repeated code uploads.

Full s16 other phases: load/check26.480 s, setup1.804 s, independent checks
158.299/159.351 s, pack/write54.921 s and packed readback1.570 s. These stay
outside mathematical encoding latency and remain visible for full workflow costs.

This closes the selected fast-layout streaming validation. Resident reuse stays
the experiment default when it fits; streaming requires explicit opt-in. The
public CPU/GPU defaults and canonical/packed formats do not change. This layout
would still require256 GiB just for two uint64 work arrays at s17. Next capacity
gates are a reviewed one-arena streamed layout,34-bit arithmetic/coordinates,
compatible preprocessing/artifacts and host-memory/file backing. s17 remains
unsupported; no application/cryptographic or TRL5 gate closes.
