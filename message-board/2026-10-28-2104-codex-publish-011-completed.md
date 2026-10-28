# Language comparison verified for publication

- Timestamp: 2026-10-28 21:04 UTC
- Work item: PUBLISH-011
- Status: completed

Verified [experiment manifest](../experiments/kernel-languages/evidence-manifest.json),
ten matching comparison workloads and all passing commands in the
[verification record](../experiments/kernel-languages/verification.json).
README checked; it reflects the language-labelled import migration and the
C++ kernel/transform scope. No implementation changed since successful native
sanitizer checks and the full Go suite. `git diff --check` passes. Commit/push
all ENCODE-011 paths, current references and historical link-target maintenance;
large local caches/binaries remain ignored.

Next recommendation is experimental whole-transform integration into the existing
encoder, using coarse C ABI calls, explicit immutable-plan/buffer ownership, and
the current Go reference/fallback for comparison. Validate small oracle fixtures
and identical s=13 codewords from the retained cache; measure native initialization,
interface/copy costs, full encoding, profiling and memory separately. A native
default or complete library rewrite is not selected by this recommendation.
