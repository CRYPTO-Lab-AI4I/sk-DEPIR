# SCALE-006 completed

Timestamp: 2026-11-05T21:57:48+00:00. Agent: Codex.

Completed the [s=15 memory audit](../experiments/preparation-memory/report.md).
The [model](../experiments/preparation-memory/model.py) traces discovery,
validation, factors, Newton inversion, full identity checking and cached online
arrays, using exact s=14 K and estimated s=15 K plus its Monte Carlo interval.
No production code, capacity status or s=15 guard changed; no s=15 preparation
or encoding was run.

The principal s=15 coefficient allocation scenario is 21.62 GiB, about 22.02 GiB
with support capacity slack. This is not an RSS bound. Cached online arrays are
18.42 GiB and code artifacts about 8.81 GiB. The 48 GiB host remains plausible;
naively scaling whole s=14 RSS to 54.9 GiB does not establish an unavoidable
memory requirement. Current endpoint/CPU evidence does not validate a 32 GiB
process guard or locate exact stage live-heap peaks.

Small profiled [replicas](../experiments/preparation-memory/replicas-final/results.json)
measure map/factor/bitset layouts, with maxima below 154 MiB process RSS.
Map density projects about 4.54–6.02 GiB at s=15; a proposed bitset needs 128 MiB.
Factor storage has a 4.65 GiB scenario and millions of objects. These are earlier
phases and not all additive to the final FFT peak. Keep both the pilot and final
replicas: the pilot baseline included CPU-profiler buffers, corrected in the
final run by sampling baseline after profiler initialization/GC.

Next: compact support validation and bounded factor construction, reusable
preparation buffers where needed, then targeted s=14 preparation from its retained
basis with live-heap/RSS checkpoints and exact P/G/spectrum comparisons. Set
s=15 resource limits from that evidence before guard extension/full admission.

[Verification](../experiments/preparation-memory/verification.json) passes source,
binary/profile/evidence digests, model assertions and Go vet. Python compilation,
documentation links and whitespace checks pass. Encoding suites were not repeated:
production code is unchanged. All CPU/heap profiles and reasonably sized replica
binaries are retained; no expensive preprocessing caches were regenerated/deleted.

README check: updated [root README](../README.md),
[scaling plan](../docs/encoding-scaling.md), experiment index and narrowly scoped
raw-profile whitespace attributes. The prior timing checkpoint was committed and
pushed as d638a9f before this audit began; this new audit is a separate local work
item, with its changed files and evidence linked above.
