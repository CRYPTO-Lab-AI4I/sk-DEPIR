# Explicit streamed canonical preprocessing export

Status: **implemented; bounded GPU validation76710 submitted**. This is step2 of the
[s18 memory plan](../../docs/encoding-scaling.md#s18-preprocessing-memory-design-and-comparison-scaling-096),
following the [passed primitive76704](../preprocessing-memory/report.md#completed-bounded-validation-scaling-099).
It does not admit fulls18 preprocessing or replace the faster resident default.

## Shared preparation and ownership

[Polynomial preparation](../../encoding/liftrs-shared/preprocessing.go) supplies
one root/reciprocal/full P*G identity procedure to resident `Prepare` and the new
`PreparePolynomials` entry point. The latter returns immutable verified P/G and
metadata, without a host spectrum or support/frontier reference. P and the root
checkpoint alias one allocation. Both entry points preserve caller-owned data.
[WriteCode](../../encoding/liftrs-shared/preprocessing_export.go) holds the session
lock and streams canonical header/P/spectrum through the existing GPU forward
transform and bounded tile. `WriteGenerator` preserves canonical reproducible G.
No second polynomial implementation or new transform schedule is introduced.

The [CLI](../../cmd/pir-preprocess/main.go) opts in with `--spectrum-tile-bytes`;
zero/omitted keeps resident behavior. It drops its own support, root/reuse and
loaded recovery references after identity verification. It preserves persisted
support/root/recovery files and the recovery-save callback. After code/G output,
both modes drop final prepared references before bounded verification. No forced
GC or array mutation is introduced; reclaimable objects do not guarantee RSS falls.

Code and generator files use the existing buffered hash/write/fsync, full
canonical range/digest/EOF checks, and atomic stage manifest/rename logic. Errors
leave diagnostic temporary directories; a completed code stage appears only after
verification. Manifest schema and canonical bytes remain unchanged: selected tile
size is recorded in the CLI progress event and controller command/case metadata,
not as an unknown field that would break older strict manifest readers. A cached
completed file can be verified without constructing a GPU engine. Root reuse
still performs the entire P*G identity before constructing a new spectrum.

## Memory and time tradeoff

| s18 final export-stage principal arrays | Estimate |
|---|---:|
| Previous resident support/P/G/spectrum |889.446 GiB |
| Stream spectrum, retain support |377.447 GiB with1 MiB tile |
| New CLI stream and release support references |**256.001 GiB with1 MiB tile** |

These are source-derived lifetimes, **not measured peak RSS**. Stream saving is
512 GiB less1 MiB, plus121.446 GiB support references removed by the CLI. Root/P
are aliases; releasing root does not save another60.723 GiB. P+G is4N+4 bytes.
Library callers keeping their support/checkpoints retain that memory. GC, recovery
aliases, runtime, write cache, buffers and metadata need additional headroom.
Discovery770.893 GiB and resident product arrays1536 GiB are unchanged, and final
resident GPU transforms still require2N extension arrays. This alone does not
make fulls18 runnable on four H200s or one host.

No extra transforms, discovery, polynomial multiplication or output payload is
added. Tile transfers/writes and changed garbage-collection timing may affect
latency. The primitive's32 MiB hash-sink timings cannot predict disk/full-size
latency. The explicit slower-memory-mode policy stays in effect: resident remains
default when its admitted memory fits; use the flag for this experiment, and
measure complete workflow costs before adopting a large-job strategy.

## Reproduction, budget and evidence

[Runner](run.py), [Slurm launcher](run.sbatch), [analyzer](analyze.py),
[GPU API tests](../../encoding/liftrs-shared/preprocessing_export_gpu_test.go),
[native stage tests](../../encoding/liftrs-shared/preprocessing_export_test.go),
[flag tests](../../cmd/pir-preprocess/export_test.go). The runner delegates
resource guards, profiles, source/evidence/cache seals and existing complete
small resident regression workflows to the [shared controller](../gpu-preprocessing-workflow/run.py).
Commit producer sources before submission; execution-start revision and all
copied source hashes are recorded. Binaries/profiles/reference caches remain at
`.cache/liftrs/preprocessing-export/export-01/`; raw results and progress logs at
`runs/export-01/`. Result paths are pending, not completed evidence yet.

```sh
python3 -B experiments/preprocessing-export/run.py --check
module load slurm/slurm/25.05
sbatch --output=experiments/preprocessing-export/slurm-%j.log \
  experiments/preprocessing-export/run.sbatch export-01
# After the completion ping:
python3 -B experiments/preprocessing-export/analyze.py export-01
```

Same previously agreed budget: one HPC node, **4 H200 /1 GiB owned per card**,4CPU,
8 GiB host request/6 GiB sampled RSS guard/4 GiB Go soft limit,2 GiB new retained
storage/4 GiB scratch.10-minute command/15-minute controller/20-minute Slurm
cutoffs; gnode03 excluded, no node pinned. New API cases cap each GPU at128 MiB;
complete small CLI cases keep1 GiB/card. Actual node/hardware/toolchain, exact
revision and producer hashes are recorded in results, not assumed from76704.
Native phase/pass and stream chunk progress is enabled; resource heartbeats are
not claimed as advancing work. No monitoring after submission.

The suite covers independent/native polynomial equality, six API layouts across
1/2/4GPUs peer/local, canonical export bytes vs independent Go, writer and parameter
failures, immutable support/root and restart, four complete s8 streamed CLI
workflows vs resident canonical/G bytes, cached/root restart, CUDA-free readback,
corruption and failed-output-admission recovery, plus four-GPU API/CLI memcheck.
Inherited native CLI regression/race/ASan/vet and resident complete workflows run
with the refactored code. CPU profiles cover phases; end-to-end command timings
include profiling, logging, hashes/fsync/readback, while native operation clocks
remain separately logged. The preparation phase includes spectrum in resident
mode but excludes it in streamed mode; compare complete workflows rather than
that phase alone. Small-case whole-command RSS cannot establish full-size memory
saving or large I/O penalty. Full s17 streamed and fulls18 remain unvalidated.

## First validation submission

**76710 submitted**, source implementation and recipe committed as`1deb2ef`.
[Receipt](submission.json); results `runs/export-01/`, scheduler log
`slurm-76710.log`, retained output/profiles/binaries
`.cache/liftrs/preprocessing-export/export-01/`. The controller records its
execution-start Git revision and exact copied code hashes; later receipt/board
changes do not change producer implementation. No job polling or log following;
await the user's completion ping before interpreting results. No inspection or
change to the separate GPU encoding retry76706.
