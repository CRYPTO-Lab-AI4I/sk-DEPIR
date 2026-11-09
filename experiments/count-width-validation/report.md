# Count-width compatibility validation

Status: corrected job 72699 (`width-02`) **passed**. The change widens
small in-memory domain counts while retaining uint32 coefficient arrays. Native
plans choose the stored count width from their domain; minimal files use 32-bit
K/N for s≤15 and 64-bit K/N at s=16. This adds eight file-header bytes only when
needed. Old files/descriptions retain their exact interpretation and bytes.

## Reproduction and scope

Starting revision: `39c6cf5` plus the recorded source changes. Go 1.27.3,
GCC 14.2.0, generic C++17 `-O3`, Linux x86_64. From the repository root:

```sh
module load slurm/slurm/25.05
sbatch --output=experiments/count-width-validation/slurm-%j.log \
 experiments/count-width-validation/run.sbatch width-01
```

Use a fresh name for each run. [run.sbatch](run.sbatch) first runs small-buffer
checks with AddressSanitizer and UndefinedBehaviorSanitizer, then invokes the
existing [bounded baseline runner](../hpc-cached-s15/run.py). Its results, phase
profiles and exact source hashes appear in `../hpc-cached-s15/runs/<name>/`.
The [baseline methodology](../hpc-cached-s15/report.md) defines staging, loading,
encoding, validation and packed-output boundaries. No preprocessing is regenerated;
original fixtures are never an output destination. Retained binaries/new outputs
remain outside Git in `.cache/liftrs/hpc-cached-s15/<name>/`.

The allocation requests 16 CPUs, 64 GiB and one hour, with no GPU. Sanitizer checks
use tiny arrays; the delegated baseline runner applies its explicit 64 GiB sampled
process-group RSS guard because the cluster did not enforce the earlier memory
request. Its large probe has a 35-minute test timeout and retained Go CPU profiles.
The new test suite covers the 2^32 count, high-bit field values and affine offsets,
small forward/inverse transforms, 32-level matrix metadata, malformed native/file
headers, caller allocation caps, old-format byte equality and writer error handling.

The cached s15 run must reproduce the original complete canonical and packed
digests after both full suites and race/vet pass. This is a correctness regression
run with profiling/monitoring, not a matched timing comparison. s16 preprocessing
and full encoding, s≥17 values, GPU execution and TRL 5 completion remain unverified.
Experimental full preparation remains guarded at s≤15 pending memory/admission review.

## Admission checks

The local count/header and native small-buffer tests pass, including the writer
example and failure propagation. Job 72698 (`width-01`) passed sanitizer and race
checks but vet rejected unkeyed `QuadraticElement` literals in the new test. The
runner stopped before encoding; [failed results](../hpc-cached-s15/runs/width-01/results.json)
and [batch log](slurm-72698.log) are retained. The test now names its A/B fields,
and local vet passes. Job 72699 repeats the corrected checks and baseline under
a fresh result name. Measured pre-fix/pre-format test sources remain in
content-addressed snapshots; no reference cache or historical result was changed.

## Completed verification

[Results](../hpc-cached-s15/runs/width-02/results.json),
[phase profiles/clocks](../hpc-cached-s15/runs/width-02/phases.json),
[analysis](../hpc-cached-s15/runs/width-02/analysis.json) and
[batch log](slurm-72699.log) record successful execution on gnode17, a shared
Xeon 6710E node, with CPU affinity `2-9,64-71`. Both full suites, focused race,
vet and the sanitizer checks pass. The additional public-probe width assertions
were also run locally afterward; no runtime implementation changed after launch.

Run the existing evidence analyzer with:

```sh
python3 -B experiments/hpc-cached-s15/analyze.py width-02
```

It verifies 231 measured sources, sealed evidence, retained binary/output and
same-clock native-stage containment. Exact dirty sources from both attempts are
retained in content-addressed `../source-snapshots/`, including pre-fix/pre-format
tests. The binary's build metadata confirms generic amd64/v1 and no sanitizer flags
in the large regression (sanitizers were used only in the preceding small checks).

| Measured phase | Seconds |
|---|---:|
| Constant staging and verification | 54.645 |
| Minimal online loading | 12.683 |
| Native/workspace initialization | 0.025 |
| Encoding, one profiled repeat | 705.864 |
| Complete systematic/curve/hash checks | 179.199 |
| Packing, hash and fsync | 14.834 |
| Packed readback | 0.00188 |
| New output retention | 2.454 |

The complete canonical word and 2,013,265,920-byte packed output reproduce the
original s15 digests. All 216,914,047 systematic symbols, 16 independent curves
and 1,027 packed reads pass. Probe peak RSS is **18.455 GiB**; its sampled
process-group peak is 19,826,192,384 bytes. Native stage counters fit within the
enclosing 705.864-second native clock. Go phase profiles remain retained.

The preceding baseline used gnode02, while this regression used gnode17. Timings
are not a controlled old/new comparison. Verification in particular took 179.2 s
versus the earlier 37.1 s; its profile attributes 95.9% of sampled CPU to the
verification closure's own work. Do not attribute this difference to the count
change without a matched host/memory-placement experiment. This work establishes
output compatibility, not a performance improvement or a measured overhead bound.

Independent native matrix and complete-online fixtures also pass ASan/UBSan:

```sh
g++ -std=c++17 -O2 -g -pthread -fsanitize=address,undefined \
 -fno-sanitize-recover=all encoding/liftrs-cpp/c_api.cpp \
 encoding/liftrs-cpp/matrix_test.cpp -o /tmp/sk-depir-count-width-matrix-test
/tmp/sk-depir-count-width-matrix-test
g++ -std=c++17 -O2 -g -pthread -fsanitize=address,undefined \
 -fno-sanitize-recover=all encoding/liftrs-cpp/c_api.cpp \
 encoding/liftrs-cpp/online_test.cpp -o /tmp/sk-depir-count-width-online-test
/tmp/sk-depir-count-width-online-test
```

These tiny native fixtures were run outside the tool sandbox after LeakSanitizer
reported its ptrace restriction inside it; both then exit zero. They do not run
a large codeword. Native perf remains unavailable under the node's permissions;
Go profiling and stage accounting are retained. Full s16 preparation/admission
and wider s≥17 values remain separate work.

[Final verification](verification.json) records current code hashes, formatting,
documentation-link checks and validation outcomes. Only test assertions/comments
and formatting were added after the measured source copy; the corresponding
local focused checks pass. Runtime implementation sources stayed fixed during
the completed regression.
