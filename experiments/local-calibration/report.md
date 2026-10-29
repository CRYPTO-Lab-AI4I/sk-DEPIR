# Step 4: local CPU, storage and resource calibration

The bounded local calibration and resource checks completed successfully. This experiment uses the
existing C++ executor and retained preprocessing. It measures independent tiles,
file transfers and a bounded read/compute/write pipeline; it does not implement
parallel full-domain encoding, blocked matrix conversion or a GPU backend.

## Scripts and evidence

- [Calibration runner](run.py), [native tile/I/O driver](main.go),
  [macOS cache-policy helper](cache_darwin.go), [other-platform policy](cache_other.go),
  [analyzer](analyze.py).
- [CPU/storage/overlap results](initial/results.json), [events](initial/events.jsonl),
  [summary](initial/table.md); [small-tile results](small-tiles/results.json),
  [small-tile table](small-tiles/table.md); [smoke results](smoke/results.json).
- [Fresh-process resource runner](resources.py),
  [private resource probes](../../encoding/liftrs-shared/calibration_resource_test.go),
  [resource results](resources/results.json), [summary](resources/table.md).

```sh
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  PYTHONPYCACHEPREFIX=/tmp/depir-pycache \
  python3 experiments/local-calibration/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name initial
# Balanced s=13 matrix-size tiles; same 64 MiB aggregate payload:
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  PYTHONPYCACHEPREFIX=/tmp/depir-pycache \
  python3 experiments/local-calibration/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name small-tiles --small-tiles
# Run separately after the CPU/storage workload finishes:
GOCACHE=/tmp/sk-depir-go-cache CXX=/usr/bin/clang++ \
  PYTHONPYCACHEPREFIX=/tmp/depir-pycache \
  python3 experiments/local-calibration/resources.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
python3 experiments/local-calibration/analyze.py initial
python3 experiments/local-calibration/analyze.py resources
```

The calibration refuses an existing run name; choose fresh names to repeat. The
resource runner uses fresh processes and a fresh resources directory. A smoke
run uses `--quick --run-name smoke` and is correctness evidence only.

Measured source: e4d693a plus uncommitted calibration sources, identified by
SHA-256 in each manifest. Environment: Go 1.27.3, Apple Clang 21, C++17 generic -O3,
macOS 26.6.2 arm64, 48 GiB Apple M5 Pro with 18 logical CPUs. Compiler/OS strings
and machine RAM are recorded. Workspace is local APFS, with about 770 GiB free
before the run. No affinity, exclusive machine reservation or frequency control.

## Methodology

### Independent native tiles

Generate sixteen dense seed-1001 input tiles, each 2^20 uint32 extension elements
(4 MiB) over the existing s=13 quadratic field. Copy input outside timing. Use the
same immutable C++ plan with distinct buffers at 1,2,4,8,16 workers; three repeats
per setting, fixed order. GOMAXPROCS follows worker count. The timed batch includes
worker dispatch and native calls, with no per-element language crossing. Hash every
output outside timing against the serial reference; finally check inverse recovery.

This is an available-parallelism calibration. These are independent prefix-subspace
transforms, **not** the arbitrary-basis/offset tiles required by every matrix stage,
and not a speedup measurement of the five full encoder transforms. Global Taylor
steps, transposes, output routing and preparation can limit end-to-end scaling.

### File blocks and cache policy

For each 4 KiB, 64 KiB, 1 MiB and 4 MiB block size, write a deterministic 256 MiB
file, time fsync separately, then read it. Three repeats truncate/rewrite the same
per-size file. Full content is checked outside timing. Timed buffers are allocated
before measurements. Logical bytes/calls are recorded explicitly.

On this host the helper requests `F_NOCACHE=48`, verified from the local SDK's
`sys/fcntl.h`. This asks macOS to bypass its file data cache for the descriptor;
it does not disable SSD/controller caches, guarantee physical device traffic or
make a small transfer representative of sustained multi-terabyte encoding. The
subsequent validation reads use ordinary cached I/O. Fsync is not F_FULLFSYNC and
is not a guarantee that the controller's volatile cache is flushed to durable media.
Write bandwidth excluding sync must not be advertised as durable-media bandwidth.
Other OSes explicitly report cached I/O rather than guessing a bypass flag.

The fixed-order 256 MiB transfers are a **bounded block-size calibration**, not a
long sustained-throughput/device qualification. Report medians and individual
observations, including outliers. Use no hard throughput guarantee in the resource
model without a larger/longer workload-specific validation.

### Computation/I/O overlap

The same sixteen tiles are written to an input file. Compare sequential
read→decode→native transform→encode→write with a bounded three-stage pipeline
using two or four slots. One compute stage remains serial to isolate overlap;
this is separate from the worker-count experiment. Each slot owns a raw byte
buffer plus uint32 buffer (8 MiB per 4 MiB tile). Native work is identical and output
hashes are checked outside timing. Fsync is separately timed. Count both file
traffic and buffer payload; no claim of overlap for the full matrix encoder.

### Profiles and memory

Every phase has a Go CPU profile; startup/stop and pre-phase GC are outside elapsed
time. Allocation/heap endpoints include profiler overhead. Native time appears
under cgo; detailed native kernel attribution reuses prior unchanged-kernel
[stack evidence](../native-encoding/s13/go-cpp/cpu-sample.txt). Short storage phases
may contain few samples. Files are warm at various layers; CPU profiles do not
measure asynchronous device work. Whole-process peak RSS is collected with wait4,
separate from phase heap endpoints.

Calibration retains under 2 GiB of local plan/input/output files under
`.cache/liftrs/local-calibration/<source-id>/<run-name>/`; Go's soft memory target
is 2 GiB, not an OS cap. Each case has a 15-minute timeout. Binaries/provenance,
checksums and regeneration commands are retained, as are profiles and text views.
Resource probes use a 9 GiB soft target and ten-minute timeout per fresh process.
They retain additional reusable s=10 artifacts and a roughly 302 MiB s=13 online
file under `.cache/liftrs/local-resources/<source-id>/`. The original large s=13
cyclic cache remains read-only; no discovery/preparation at s=13 is repeated.

## Preparation and minimal-loader process boundaries

Generate and retain the exact s=10 basis once, then run baseline and paired/root-only
preparation in separate fresh processes. Compare complete serialized P/G/spectrum
artifacts by SHA-256. Preparation timing excludes serialization, but whole-process
RSS includes loading, serialization and runtime. This is a controlled resource
comparison with identical code parameters, not a large-scale capacity forecast.

Export the minimal s=13 binary once from the existing verified JSON. Export is a
conversion/setup operation and remains separately labelled. In another fresh
process, incrementally load and checksum the minimal file, without loading the
original JSON, retaining G/support, allocating encoding workspaces or encoding.
That process measures loader/resident-plan memory only; an encoding RSS budget
must add input, workspaces, output and runtime overhead.


## Results

### Useful CPU concurrency

Medians of three batches, with identical complete hashes at every worker count:

| Workers | Sixteen 4 MiB tiles | Speedup | 2,048 × 32 KiB tiles | Speedup |
|---:|---:|---:|---:|---:|
| 1 | 8.305 s | 1.00× | 5.796 s | 1.00× |
| 2 | 4.247 s | 1.96× | 3.057 s | 1.90× |
| 4 | 2.220 s | 3.74× | 1.548 s | 3.74× |
| 8 | 1.208 s | 6.87× | 0.829 s | 6.99× |
| 16 | 0.669 s | 12.42× | 0.454 s | 12.77× |

Both cases process 64 MiB input, with another 64 MiB output workspace. The smaller
case matches a balanced s=13 matrix's tile cardinality; it still uses canonical
prefix-subspace constants, **not** image-basis/affine-offset constants for every
matrix stage. Different transform sizes have different arithmetic work; compare
speedups within a column, not as two algorithms for the same full transform.
The smaller case establishes that thousands of current native calls do not erase
worker scaling at this granularity. Generalized tile kernels still need validation.

No saturation was observed by 16 workers for this workload. That does not mean
18 workers, more threads or every full transform will scale similarly. No claim
is made that 153 s full encoding becomes 153/12 seconds. Taylor conversion,
transposes, permutations and synchronization remain unmeasured parallel stages.
Whole-process peak RSS for the initial CPU/storage/pipeline process was 179 MiB.

### Block transfers and overlap

The [generated table](initial/table.md) retains exact medians. Under requested
F_NOCACHE, 4 KiB reads achieve ~128 MiB/s; 1 MiB and 4 MiB reads achieve ~4,867
and ~8,089 MiB/s over these 256 MiB files. Report these as short transfer rates,
not sustained or guaranteed device bandwidth. Excluding fsync, write medians are
~1,917, 2,522, 7,875 and 11,989 MiB/s at increasing block sizes; these are
filesystem/controller-visible completion rates, **not durable media throughput**.
The first 4 KiB write was 0.698 s versus 0.125–0.134 s later; one 64 KiB fsync took
0.604 s while others took ~0.008–0.009 s. Those outliers remain in raw results.

The data strongly argues against the scalar calls in the step-3 matrix probe.
Start with MiB-sized buffered transfers, with exact alignment/transpose schedules
measured separately. The data does not justify multiplying the toy matrix's byte
counts by the best observed bandwidth and claiming target-scale runtime.

| Pipeline | Read/compute/write median | Slot buffer payload |
|---|---:|---:|
| Sequential, 1 slot | 8.341 s | 8 MiB |
| Overlapped, 2 slots | 8.320 s | 16 MiB |
| Overlapped, 4 slots | 8.332 s | 32 MiB |

There is no meaningful overlap gain with one compute stage. Native compute
already dominates this workload; postpone elaborate overlap queues until parallel
matrix work and its full movement schedule are measured. These results do not
measure a pipeline combining sixteen compute workers with I/O.

### Fresh-process preparation and loading

| Process | Main phase | Process peak RSS |
|---|---:|---:|
| s=10 exact basis generation | 0.737 s | 26.16 MiB |
| s=10 baseline preparation | 12.000 s | 78.38 MiB |
| s=10 paired/root-only prototype | 8.795 s | 50.62 MiB |
| s=13 original-JSON → minimal-file export | 5.325 s | 2,312.55 MiB |
| s=13 minimal binary load, fresh process | 1.010 s | 623.08 MiB |

Both s=10 preparation processes produce the exact same 5,397,696-byte artifact,
SHA-256 `345e600ecf5754baade484b5cefaecdaa2eb3878083efd58f7287857aef77fdc`.
The prototype lowers observed peak RSS ~35% and phase time ~27% in this one case,
but allocated bytes increase from 275 million to 522 million. Its GC/scratch
behavior still needs work before a large-setup bound. This fresh-process result
improves the earlier small-fixture evidence; it does not guarantee s=15 preparation
fits RAM.

The minimal s=13 plan contains 316,141,200 coefficient bytes. Fresh loading peaks
at 623 MiB RSS, with ~311 MiB final heap. Binary per-element decoding still
allocates temporary data, so total loading allocations are ~635 million bytes.
A buffered block decoder is a concrete next simplification to test; peak is not
just the resident payload. Original JSON export retains its larger memory cost,
but happens once and is not silently included in every online load. No complete
encoding runs were repeated in step 4. Add input/workspace/output and actual
runtime behavior before budgeting a complete s=14/15 encoding process.

### GPU and hardware assessment

[Filtered GPU inventory](hardware.json) reports Apple M5 Pro/Metal 4 support.
`xcrun --find metal` returns exit 72 on the current setup; no compiler installation
or GPU benchmark was attempted. This is a toolchain-availability observation, not
an assertion that GPU execution is impossible or slower. CPU concurrency already
provides a substantial measured avenue without choosing a vendor API. A GPU shares
this machine's RAM budget; it does not remove large preprocessing or file costs.
A discrete GPU or larger CPU server remains an option after an integrated design
has measured capacity/throughput requirements.

## Implication for step 5

The preferred first implementation candidate is the lean CPU design with bounded
workspace and parallel native matrix tiles. Retain canonical ordering and the
reference paths. Next implement native arbitrary-basis/affine tile semantics and
blocked Taylor/transpose transfers, improve minimal-file decoding, and validate
the combined schedule against the s=13 reference. Use up to 8/16 workers as
**test settings**, not a universal default selected by independent-tile results.
Only then admit an s=14 or s=15 full workload under explicit measured budgets.
No HPC move, GPU dependency, higher public cache budget or larger supported s is
selected by these results. Longer sustained storage tests and integrated parallel
I/O remain acceptance gates if the chosen design relies on them.

## Verification and retention

[Verification](verification.json) links smoke/full correctness, native inverse
checks, exact prepared-artifact equality, tagged tests/vet, Go race-instrumented
quick driver, reproducible analyzers and source/cache/binary/profile hashes.
Race instrumentation covers Go scheduling/buffers, not C++ memory accesses; native
plans are immutable and kernels are unchanged. Current new flags were exercised
by the small-tile run; the original pipeline source for the race check is retained
as a content-addressed snapshot. Driver and private test probes do not change
production algorithms or APIs. No repeated full suite is needed for unaffected
packages; the tagged shared encoder suite and new driver compile/check pass.

Retained initial files total ~1.25 GiB; small-tile files add only a plan/binary.
Resource caches retain ~302 MiB online data plus s=10 code artifacts/binary.
All remain under `.cache/liftrs/`, with provenance/checksums; profiles/events and
text views are versioned alongside reports. No useful original caches were deleted.
Some profile/toolchain commands emitted a sandbox module-cache stat warning;
builds and checks completed successfully. Hardware reads used read-only metadata
access; identifiers/serials were omitted from the recorded GPU inventory.
