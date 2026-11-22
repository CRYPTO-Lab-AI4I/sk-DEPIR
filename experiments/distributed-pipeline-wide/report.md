# Wide distributed preprocessing adapters

Status: prepared; integrated GPU results pending. Parent revision9e3dbc2.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/). Actual implementation and
producing revision will be sealed by the committed snapshot/controller. Results
use `runs/wide-01/`; source,binaries,profiles and reusable shards remain outside
Git in `.cache/liftrs/distributed-pipeline-wide/wide-01/`.

## Scope and implementation

The [complete four-worker pilot79983](../distributed-pipeline-four/report.md) passed
both genuine cold execution and fresh-process prefix8 restart. Its assembled
adapter used32-bit extension values. The [same pipeline implementation](../../encoding/liftrs-shared/distributed_pipeline_gpu_test.go)
now selects its extension word type explicitly while retaining32-bit base P/G
coefficients. Distributed discovery and executor/transpose/Taylor codecs use the
selected type; spectrum persistence/publication no longer assumes32 bits.
The existing two/four-worker defaults and canonical representation are unchanged.
Ordinary CPU client builds remain independent of CUDA/MPI.

[Stream publication](../../encoding/liftrs-shared/distributed_publication_gpu_test.go)
shares one typed section helper: base P/G use4-byte slots; spectra use4 bytes
through s16 and8 bytes for s17–20. It reuses existing checked MPI operand frames,
streaming writer,canonical limits and bounded4096-word tiles. No extra full-array
conversion,gather or transform is added. Full-domain publication still writes the
actual canonical code header; bounded raw-field diagnostics below deliberately
use separate raw outputs and never claim a small transform is a full s18 code.

This is a width-sensitive adapter gate,not genuine fulls18 discovery or generator
construction. The canonical s18 field has2^18 base values and2^36 extension values;
the actual code grid would have2^36 positions. Bounded transforms below have32 or
4096 positions. No claim of complete s18 preprocessing follows from their pass.

## Workload and checks

One allocation has three profiled normal arms:

1. **Cold:** existing genuine s4/s6 complete four-worker pipeline,including
   discovery,hierarchical roots,Newton,full cyclic polynomial identity,spectrum,
   coordinated checkpoint commits and streamed canonical/generator artifacts.
2. **Resume:** new MPI processes reload the verified root/prefix8,skip discovery
   and root work,and reproduce all remaining prefixes/final artifacts. These two
   arms reuse [the validated runner/reviewer](../distributed-pipeline-four/run.py)
   and trusted [native-input manifest](../distributed-coordinator-stream/reference-inputs.json).
3. **Wide:** [bounded field18 tail](../../encoding/liftrs-shared/pipeline_wide_gpu_test.go)
   reuses the maintained64-bit executor for Newton products,complete nonaliasing
   convolution,spectrum and raw stream publication. All four ranks participate.

| Wide case | Transform domain | Base polynomial coefficients | Reciprocal prefix | Full checked convolution length |
|---|---|---|---|---|
| dim5 | 32 values | 9 | 16 coefficients | 24 coefficients |
| dim12 | 4096 values | 17 | 64 coefficients | 80 coefficients |

Base polynomial is P[0]=1,P[j]=((j+1)*0x1337) modulo2^18. Each reciprocal prefix
is checked against an independent coefficient recurrence. Go uses direct base-field
multiplication; Python independently uses shift/XOR reduction with canonical
polynomial0x40081. Schoolbook convolution checks every output/tail coefficient and
its identity prefix. This is an inverse modulo X^prefix;it is not the full cyclic
code identity checked by the genuine arms.

Spectrum is checked against an independent native wide FFT. A separate deterministic
high-bit pattern explicitly exercises bits32–35 during MPI and little-endian8-byte
publication,because a bounded evaluation subspace can keep actual spectrum values
inside the base field. The pattern is ((1<<35) XOR ((i+1)*0x9e3779b9)) modulo2^36.
Raw generator/spectrum/high-pattern output must be byte-identical to concatenated
owned shards,with generator truncation to its actual prefix. Focused reviewer tests
reject high-bit truncation and wrong stream width.

The reviewer enumerates all expected executor operation IDs,three FFTs per product,
transpose/Taylor routes,sparse Frobenius-square packets and publication packets.
Before/after payload proofs bind both partners;Python also checks square payloads
against its independent recurrence and publication payloads against saved shard
bytes. Every stored prefix and complete convolution is independently checked.
No incorrect final artifact can hide a missing transfer. Checksums detect corruption,
not maliciously rewritten evidence. Cold/resume retain the prior canonical output
hashes and verify actual fourth-rank manifests/PID/host ownership.

## Resource,time and memory accounting

Unchanged admitted resources: **two HPC nodes,two H200s,four CPU cores,8 GiB total
requested host RAM**,four workers/two per card. **64 MiB owned GPU/worker =128 MiB
per card**,**1 GiB RSS/worker**,zero CUDA-pinned buffers;**128 MiB retained cache,
2 GiB scratch,32 MiB/file**;**720-second controller/15-minute Slurm cutoff**.
Other eligible HPC nodes allowed,gnode03 excluded. CUDA context/driver/MPI internals
are additional to owned-array accounting. No resource increase or production mode
selection is made. Ordinary faster validated paths/defaults stay available.

Estimated complete controller **3–5 minutes excluding queue**,based on79983 taking
2 min11.447 seconds for both genuine arms including builds/profiles. The extra
bounded tail has12 products and2 spectra;new complete latency is unmeasured.
Existing per-launch220-second test/240-second launcher and720-second controller
bounds leave headroom. No CUDA arithmetic changes or repeated sanitizer-only arm;
retain the prior clean field18/hierarchy/genuine memchecks. New width-sensitive
adapter behavior is validated together with the genuine regression.

Three extension work arrays take24N/r bytes at s18 (r=4),versus12N/r for32-bit
extension arrays. Canonical base coefficients stay4 bytes. Generic streaming adds
one4096-word tile (at most32 KiB at64 bits) on rank0;it does not materialize a new
whole spectrum. Diagnostic native FFT/reference/tail arrays replicate at most4096
values;this bounded independent oracle is not a fullsize memory strategy. The
canonical word-width increase is required representation,not an optional memory
optimization. Runtime penalty from wider transfers at fullsize is unknown.

Prior genuine proof logs are a few MiB/rank. The added bounded12-product tail is
projected below8 MiB/rank,under32 MiB/file;cache projected below32 MiB,under128 MiB
cap. Exact sizes are measured by the controller. Completed-arm results/profiles
persist before later phases. Complete case times include reference/setup,
communication,products/spectrum,logging and persistence;controller adds builds,
preflight,review and profile summaries. Existing timestamped phase/pass/transform/
tile logs and resource guards remain active;proof capture is included in timings,
with no isolated logging-overhead measurement.

## Reproduction and local evidence

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9,H200. Build all tagged interfaces without
initializing MPI/CUDA;native field18 transform oracles and actual genuine retained
support/hierarchy geometry pass. Focused race tests cover stream widths,checkpoint
binding,operand ownership and root scheduling. All affected Python suites and79983
cold/resume saved-evidence replay pass. Verify the actual committed-source snapshot,
runner/wrapper/library paths and retained-input manifest before submission. In-job
preflight checks shared input/executable/output paths on every participating node.

```sh
python3 -B experiments/distributed-pipeline-wide/test_review.py
python3 -B experiments/distributed-pipeline-wide/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-pipeline-wide/slurm-%j.log \
 experiments/distributed-pipeline-wide/run.sbatch wide-01
# After completion notice:
python3 -B experiments/distributed-pipeline-wide/analyze.py wide-01
```

[Build](readiness/build.log), [native](readiness/native.log), [race](readiness/race.log),
[reviewer](readiness/review.log), [legacy replay](readiness/legacy-replay.log).
GPU integration remains pending. Fulls18 still needs scalable discovery/root ownership
without replicated bounded reference state,complete wide full-domain workflow,
representative larger scaling/restart,measured capacity/headroom and separate
full-run admission. This gate does not authorize that allocation.


[Actual committed-source preflight](readiness/snapshot-checks.json) passed on
**ee36990**:811 build inputs/3,726,681 bytes. All required
wide/pipeline/MPI/wrapper/manifest paths exist;all seven reviewer suites passed
from frozen source. The exact snapshot input-copy helper verified69,746 retained
native bytes before allocation. The following readiness commit changes only
documentation/evidence.
