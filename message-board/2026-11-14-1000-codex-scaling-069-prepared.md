# Resident GPU preprocessing prototype and readable tables prepared

- Timestamp: 2026-11-14T10:00:02.012930+00:00
- Agent: Codex
- Work: SCALING-069
- Status: handoff (source checkpoint before bounded GPU validation)

Server preprocessing is now GPU-first, with CPU serving independent arithmetic/
client/reference roles. New [resident GPU engine](../encoding/liftrs-cuda/preparation.cu)
implements exact weight classification/expansion, root warp scans, base-grid and
extension transforms/products/spectrum, sharing the existing Go sampling/root/
reciprocal/identity/checkpoint procedure. Word width follows s;1/2/4/8 explicit
same-node peer shards use one generic implementation. Legacy CUDA32 empty-device
selection and ordinary CPU/client builds remain. No CPU/host-backed fallback.

Shared [arithmetic/vector/Taylor](../encoding/liftrs-cuda/transform.cuh),
[metadata](../encoding/liftrs-cuda/preparation_plan.hpp),
[kernels](../encoding/liftrs-cuda/preparation_kernels.cuh) and native membership
headers avoid cloned s17 algorithms. Coalesced32x32 peer transposes and power-of-
two slab shift/mask addressing avoid uncoalesced peer traffic/hot integer division.
Root scans need one initial power per warp; widening/narrowing reuse dead scratch.
Existing discovery data/scratch are reused when growing product operands. Three
principal product arrays still384GiB aggregate at s17,96GiB/card over four GPUs,
plus metadata/runtime/headroom. No full run, new large budget or measured speedup.

[Guide](../docs/preprocessing.md), README/native/shared/CUDA docs, contract and
[scaling plan](../docs/encoding-scaling.md) document explicit GPUDevices and
per-card caps, source-versus-validated capability, host staging and peer cost.
[Scaling reference](../docs/encoding-status.md) splits tables into preprocessing,
three CPU resident/file modes and three GPU resident/streamed modes, preserving
numbers/provenance. Every mode cell names runtime units and memory measures. The
s16 disk cell clearly identifies10h12min39.871s,19.469GiB peak RSS,32GiB work files
and59.271GiB additional tracked cache. Estimates/unknowns/feasibility stay explicit.

[Validation experiment](../experiments/gpu-s17-preprocessing/report.md) includes
runnable guarded controller/Slurm/analyzer, complete small Go oracles, s17 high-
value/index kernels,2^22-value dense tiles and legacy-client regression, memcheck
and shared-recurrence sanitizers. Window driver tests actual s17 classification
at indices above2^32 with only a123MiB class cache/tiny bitmap; tiny physical
GPU0 buffers probe production logical slab indexing, not peer hardware. Ordinary
CPU profiles, matching binaries and small code caches are retained.

[Static evidence](../experiments/gpu-s17-preprocessing/static-checks.json): nine
host recurrence cases across simulated1/2/4 slabs, both widths/s17 offsets, both
logging schemas, native/C syntax, table units, formatting/script checks pass.
Both CUDA source units compile to compute90 PTX; final combined sm90 library and
wide-window executable compile/link, and both ABI families export. CUDA device
execution, Go/cgo, peer visibility, memcheck and performance remain pending.
Optional native progress units preserve old JSON when absent; GPU logs name units.

Only one GPU is currently admitted; asynchronous request for two cards is pending.
Proposed2-GPU check keeps aggregate512MiB owned payload (256MiB/card),4CPU/8GiB,
6GiB RSS/4GiB Go soft,2GiB retained/4GiB scratch and600/900/1200s deadlines. If no
approval arrives, submit the independent1-GPU validation under existing caps;
peer testing waits. No monitoring or inspection of jobs74447/74448 or their
analysis. README checked/updated. Full s17 runner/admission/calibration and
cross-node GPU execution remain later gates after bounded evidence.
