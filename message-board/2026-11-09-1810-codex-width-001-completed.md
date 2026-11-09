# Counts and saved metadata use the required width

- Timestamp: 2026-11-09 18:10:04 UTC
- Agent: Codex
- Work ID: WIDTH-001
- Status: completed
- Base revision: `39c6cf5`; changes remain available for review

Implemented the user's count-width change and subsequent storage refinement.
Go/native domain cardinalities use uint64 for safe arithmetic; coefficient,
offset and index values remain uint32. Saved widths are selected from s/domain
size: s≤15 keeps exact `LRSprobe1` bytes with 32-bit K/N, while s=16 uses
`LRSprobe2` with 64-bit K/N (eight additional header bytes). Native canonical/
ordered descriptors retain versions 1/2 for narrow counts and use 3/4 with one
extra high count word when required. No large value array is widened.

Added documented `WriteScalingOnline`, including an example and error propagation.
The loader reads both file versions and checks wide counts, caller element caps
and host byte-offset limits before allocation. Updated private field arithmetic,
generator search, transform level/table sizing, native description import/ordered
basis insertion, range checks and existing fixtures/callers. Full experimental
preparation remains capped at s≤15, ordinary discovery at s≤13. Native/minimal
representation and sparse probes handle s16 on small buffers; s≥17 still exceeds
the packed value width and is rejected. No full s16 run was admitted.

New tests cover independent full-width products, high-bit canonical/affine
Horner/inverse transforms, 32-level matrix metadata, old/new description widths,
2^32 header counts, malformed/overflow headers, pre-allocation caps, old byte
equality, wide-header decoding, writer errors and public probe bounds.
Existing small reference/golden tests remain in both full suites.

[Count-width report](../experiments/count-width-validation/report.md),
[batch script](../experiments/count-width-validation/run.sbatch),
[results](../experiments/hpc-cached-s15/runs/width-02/results.json),
[analysis](../experiments/hpc-cached-s15/runs/width-02/analysis.json) and
[final verification](../experiments/count-width-validation/verification.json)
link reproducible commands and evidence. The wrapper reuses the bounded baseline
runner, preserving original caches and directing fresh output to new scratch.

Slurm job 72698 passed small sanitizer/race checks, then stopped at vet's unkeyed
struct-literal finding in a new test, before encoding. Fixed the test with A/B
field names. Job 72699 on gnode17 (16 CPUs, 64 GiB request, one hour, no GPU)
passes sanitizer checks, both full Go/native suites, race/vet and cached s15.
Independent matrix and online ABI fixtures also pass native ASan/UBSan outside
the tool sandbox; the initial sandbox attempts hit LeakSanitizer's ptrace restriction.
No sanitizer was disabled. The large regression binary has no sanitizer flags.

All 216,914,047 systematic positions, 16 independent curves, complete canonical/
packed digests and 1,027 packed reads match the original s15 reference. Six phase
profiles, source/binary/output hashes and native clocks remain retained. Analyzer
passes 231 measured sources and all sealed evidence. Twenty-seven dirty measured
source versions are retained by content hash, including failed/pre-format tests.
Only test assertions/comments and formatting changed after the measured copy;
local focused checks pass on that final test source. Runtime code did not change.

Measured encoding is 705.864 s and probe peak RSS 18.455 GiB. The report includes
loading, staging, validation and output separately, including verification's
179.199 s versus the earlier 37.1 s on another node. Profiles attribute most
verification samples to its own closure work; this is not a matched timing trial,
so no overhead bound or performance improvement is claimed. Detailed native perf
remains unavailable; Go profiles and stage counters remain enabled. Fresh binary
and packed output stay outside Git at `.cache/liftrs/hpc-cached-s15/width-02/`.

Changed paths: shared Go count/field/transform/serializer adapters and tests,
native field/import/header and two fixture casts under `encoding/liftrs-cpp/`,
new count-width validation script/report/verification and delegated result runs,
measured source snapshots, root/package/experiment READMEs, handoff and s16 status
documentation, and this board work item. README check complete: distinguishes
stored-width selection, small s16 validation and full-run/value-width limits.
Formatting, diff whitespace and documentation links pass. The next gate is
preprocessing-memory/lifetime and resource-budget review before full s16 work;
large s16 feasibility, wider values, GPU encoding and TRL 5 completion remain open.
