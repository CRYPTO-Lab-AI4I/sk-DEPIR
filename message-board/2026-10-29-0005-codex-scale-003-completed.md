# Step-4 bounded local calibration completed

- Timestamp: 2026-10-29 00:05 UTC
- Work item: SCALE-003
- Status: completed (local component calibration; integrated/sustained gates open)

[Report/scripts/results](../experiments/local-calibration/report.md) now cover
native worker counts, explicit-cache block transfers, overlap and fresh-process
resource peaks. The [driver](../experiments/local-calibration/main.go) uses existing
immutable C++ plans and distinct buffers; no production encoder algorithm/API or
worker default changed. Test-only [resource probes](../encoding/liftrs-shared/calibration_resource_test.go)
run modes in separate fresh processes and retain exact code/online data.

Three-repeat native tile medians: sixteen 4 MiB tiles improve from 8.305 s at one
worker to 1.208 s at eight and 0.669 s at sixteen (6.87x/12.42x). A follow-up with
2048 x 32 KiB tiles (balanced s13 matrix cardinality, same aggregate 64 MiB input)
improves from 5.796 s to 0.829/0.454 s (6.99x/12.77x). All complete hashes and
inverse checks pass. Current canonical-prefix plans are not generalized image-
basis/affine matrix tiles; this is not a full-encoder parallel speedup claim.

Requested macOS F_NOCACHE via documented SDK constant 48. MiB-sized transfers
substantially outperform small calls over the 256 MiB file trials. Raw timings
include first-write/fsync outliers. The report separates short filesystem-visible
transfer rates from sustained/durable device bandwidth; fsync is not F_FULLFSYNC.
Sequential one-compute-stage pipeline takes 8.341 s versus 8.320/8.332 s with two/
four overlap slots: no useful overlap gain observed. This does not measure a
pipeline combining all CPU workers with I/O. Larger integrated data movement and
longer storage qualification remain gates, rather than assumed pass/bandwidth bounds.

Fresh s10 baseline preparation: 12.000 s, 78.38 MiB process peak RSS. Paired/root-
only prototype: 8.795 s, 50.62 MiB; exact serialized artifact SHA-256 matches.
Prototype cumulative allocation still grows (275M to 522M bytes), motivating scratch/
GC work. Fresh minimal s13 load: 1.010 s and 623.08 MiB peak RSS, ~311 MiB final
heap for ~302 MiB plan arrays. Per-element decoding temporary allocations are a
concrete next improvement. Source JSON export is a separate one-time 5.325 s /
2312.55 MiB operation. No new s13 full encode or expensive preprocessing was rerun.

Filtered GPU inventory records Apple M5 Pro/Metal4 support without identifiers.
Metal compiler lookup exits 72; no installation or GPU benchmark was performed.
CPU-first is the recommended candidate, not a hardware-performance verdict about
GPUs. No larger hardware, changed resource budget or supported s is selected.

Verification: smoke/full/small-tile resource cases pass, including full file content,
complete output hashes, inverse recovery and prepared-artifact equality. Tagged
shared suite/local package checks and final tagged vet pass. Race-instrumented quick
driver passes (Go scheduling/buffers only; C++ kernels unchanged). Analyzers verify
source/cache/binary/profile/evidence hashes; Python syntax, changed-doc links and
whitespace pass. A sandbox module-cache-stat warning did not prevent successful
builds. Historical pre-followup sources are content-addressed snapshots. See
[verification](../experiments/local-calibration/verification.json). No unrelated
full-suite or full-encoder baseline rerun was needed.

All profiles, text views, provenance and reasonably sized caches are retained:
~1.25 GiB initial files, small-tile plan/binary, and ~302 MiB s13 online data plus
s10 artifacts in local-resources, all under ignored .cache/liftrs/. Original caches
are untouched. README checked/updated; roadmap, study, algorithm notes, contract
and experiment index reflect measured boundaries. Next step 5 should integrate
native generalized tiles and parallel scheduling with blocked Taylor/transpose
transfers, improve minimal block decoding, verify the combined s13 path, then
review explicit budgets before s14/15. GPU/HPC or full-size feasibility claims are
premature. Changes remain uncommitted; no publish request was made for this item.
