# CPU-034 / GPU-033 — Compact memory validation complete

Timestamp: 2026-11-12T08:10:24.911828+00:00
Agent: Codex
Status: completed

After the user's completion ping, reviewed jobs73818 CPU /73815 GPU. Both
completed exit0:0 on gnode17,49m22s /28m59s, within admitted budgets and no guard
tripped. Final sacct accounting saved with each run; no new job submitted.

[CPU report](../experiments/cpu-output-memory/report.md) and
[GPU report](../experiments/gpu-output-memory/report.md) now link manifests,
analyses, phases, sanitizers and accounting. Both recorded-cache analyzers pass:
344 CPU /338 GPU exact source identities, sealed evidence/cache digests, complete
canonical/packed CPU/GPU hash agreement, all systematic positions,16 sampled
curves and1027 packed readbacks. The controller's8 GiB payload hash is preserved;
analysis rechecks metadata/physical length without a fresh full bulk rehash.
Existing preprocessing, profiles, binaries and output caches remain intact.

CPU peak sampled RAM39.409 GiB vs prior fast58.351 (32.46% less); full encoding
1289.759 s (21m29.8s), with23.586 s compaction/move and6.064 s expansion.
Matched s14 compact is6.13%/3.86% slower. Full timing comparison against prior
1247.173–1281.328 s is historical, with different allocations/runtime memory
limits. Verification806.8 s and setup25.5 s remain visible as separate costs.
Native/Go-only suites, race/vet, owning/alias/error tests and ASAN/UBSAN/TSAN pass.

GPU owned device35.429 GiB vs reuse51.427 (31.11% less), exact16 GiB removal plus
2 MiB staging. Host39.028 GiB effectively unchanged; download16→8 GiB and final
packed disk output8 GiB unchanged. Encodes101.526/93.332 s vs historical reuse
57.418/56.975 s on another allocation. Matched s10 compact9.60%/10.67% slower.
Rows55 s dominate versus prior24 s; compaction0.043 s, host expansion4.6–9.6 s.
Oracle/boundary/range-before-narrowing/ownership checks,80 matrix regressions,
build combinations, race/vet and memchecks pass (0 errors,0 leaked bytes).

Updated README, native/shared/CUDA API docs and scaling plan to replace pending
claims with measured results and comparison limits. README check completed.
Corrected analyzer summary wording from default/reuse to resident/compact; no
arithmetic/default/measurement boundaries changed. Existing CPU implementation
from the previous item remains staged with this evidence, uncommitted.

Recorded next gate: resident spectrum/P streaming and s17 width/layout audit,
with host/device savings versus uploads/disk/page-cache/setup/runtime explicit.
Both compact modes remain optional; faster validated paths stay default when
memory fits. s17 needs17-bit base and34-bit extension handling, and cannot use
this uint16 layout unchanged. No s17, cryptographic/application or TRL5 claim.

Verification: both `analyze.py compact-01 --recorded-cache` commands pass,
self-checks for both experiments pass, Python AST/local report links and diff
whitespace checks pass. Full tests/sanitizers are the completed jobs' recorded
results; no redundant long tests or further experiments were launched.
