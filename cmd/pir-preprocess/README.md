# Persistent server preprocessing

`pir-preprocess` is an experimental command around the shared `CodePreprocessor`
API. It targets the POSIX server toolchain (advisory file locks). It constructs public code data; it neither reads a plaintext database/key
nor performs storage-side PIR computation. Arithmetic and canonical formats stay
in `encoding/liftrs-shared` and its existing CPU/GPU backends. The command accepts
m=2, s=2..17; **full s=17 execution remains unvalidated and unadmitted**.

Build the server executable and CUDA plugin as described in
[preprocessing](../../docs/preprocessing.md). For a bounded s8 example:

```sh
go build -tags liftrs_cpp,liftrs_cuda -o pir-preprocess ./cmd/pir-preprocess
./pir-preprocess --s 8 --t 5 --backend cuda --devices 0 --workers 4 \
  --gpu-library /absolute/path/libliftrs-cuda.so --gpu-bytes 1073741824 \
  --max-array-bytes 8388608 --native-bytes 67108864 \
  --max-code-bytes 8388608 --cache /absolute/path/s8-code
# Repeat with identical specification/caps and --resume to reuse completed stages.
```

Use a scheduler/controller for jobs over30 seconds. These byte flags cap individual
principal host arrays, native owned bytes, each GPU's owned bytes and final code
payload respectively. They do **not** cap total RSS, GC retention, CUDA contexts,
filesystem cache or total storage. The experiment controller adds aggregate RSS,
storage, time and profiling guards; larger runs need an agreed complete budget.
`--vector-layout peer` stays default; `local` explicitly adds transpose/peer traffic.
`--backend cpp` is for native reference runs; unavailable requested backends fail.
Ordinary builds work without CUDA, including checking an already-complete cache.

The cache is trusted local state with an advisory exclusive lock. A SHA256
manifest supplied alongside arbitrary files does not prove mathematical code
correctness. Discovery certifies exact support and preparation checks the complete
P·G identity; cache reads check representation, source binding and stored identities.
Do not point `--resume` at untrusted third-party artifacts.

Each stage (`support/`, `root/`, `code/`) contains payloads and `manifest.json`.
Support and root use the existing codecs; code contains canonical `online.bin`
and base-field `generator.bin`. Manifests retain the code specification, support/
root bindings, exact sizes/digests, executable/plugin SHA256 and source revision
(injected by experiment builds; manual builds say `unknown`). Files and directories
are synced and a single rename publishes a stage. Failed hidden stage directories
are retained, never accepted as complete; account for them in storage budgets.
No completed stage is overwritten. Use `--resume` or a fresh cache explicitly.

A completed-cache restart checks all four payloads without creating a GPU engine
or another resident code array. With support/root but no code stage, it verifies
and loads those checkpoints, skips discovery/root construction, and recomputes G,
its complete identity and spectrum. A support-only restart skips discovery.
An interrupted discovery is restarted; partial sampled support is not certified.
Current root checkpoint granularity is the completed normalized root, not every
internal product-tree step.

Fresh execution streams/hashes payloads, reads each back once, and retains them.
This adds disk writes/readback, sync latency and retained storage; it is not a
cost-free memory optimization. Final cache validation streams without a second
resident P/spectrum copy. Root restart still owns loaded checkpoint arrays and
native/Go runtime overhead. No compression or slower memory mode is selected.

CPU profiles get unique `cpu-*.pprof` names in the cache (or an explicit fresh
`--profile` path); earlier profiles are retained. Stderr emits unbuffered JSON
phase events, advancing byte counts about every30 seconds, and existing library/
native work progress. SIGTERM stops profiling and exits with143, retaining partial
stages. This does not guarantee progress while a filesystem sync or kernel is
blocked; controller heartbeats label remaining work unknown. Measured workflow
latencies include logging/profiling, publication and readback; they are separate
from unprofiled preparation/kernel baselines.

[Workflow validation and reproducible scripts](../../experiments/gpu-preprocessing-workflow/report.md)
cover publication failure, corruption/binding rejection, Go-oracle equality,
fresh execution, complete-cache reuse and root-checkpoint reuse. The same-node
engine accepts1/2/4/8 device ordinals; measured HPC nodes have four H200 GPUs. Eight or
sixteen GPUs on the HPC system require cross-node orchestration/communication that
is not implemented. The command rejects16; this is an implementation boundary,
not a preprocessing design ceiling.

## Exact dimension without full host arrays

`--dimension-only` uses `CountCodeMonomials` and publishes only
`dimension/manifest.json` plus a unique profile. It needs no `--max-code-bytes`,
N-element host grid or K-element index list; ordinary preprocessing's array
admission is unchanged. Native/device extent and allocation caps still apply.
Use the [bounded admission experiment](../../experiments/s17-preprocessing-admission/report.md)
for s16/s17 GPU counts. A dimension record is not certified sampled support or
code data, and does not admit full preprocessing. Complete execution still
enumerates indices, samples support and checks all identities. `--resume` may
reuse only the trusted dimension record in this mode.

## Explicit streamed spectrum output

Append `--spectrum-tile-bytes 1048576` to an admitted CUDA preprocessing command
to opt into canonical export with a1 MiB host tile. Zero/omitted keeps resident
output. The option requires CUDA and8 bytes..64 MiB; it is invalid with the native
reference backend or `--dimension-only`. The selected size appears in progress
logs. It adds no transforms/output payload but may change transfer/write and GC
latency. Full-size costs are unmeasured; do not choose this for speed by default.

The shared verified polynomial stage lets this CLI release its support/recovery
references before exporting P/spectrum/G. Persisted checkpoints remain intact.
The existing atomic publication, digests, full range/EOF verification and cached
CUDA-free readback apply to both modes. No new manifest fields are required and
canonical files remain byte-compatible. Failed temporary stages stay unpublished
for diagnosis. The integration is implemented with bounded GPU validation pending;
see [experiment and source-linked tests](../../experiments/preprocessing-export/report.md).
