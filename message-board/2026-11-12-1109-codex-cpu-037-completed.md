# CPU-037 — File-backed CPU validation complete

Timestamp: 2026-11-12T11:09:48.880765+00:00
Agent: Codex
Status: completed

Reviewed job73976 only after user completion ping. COMPLETED/0:0, gnode17,
32 CPUs/no GPU,36m45s allocation (below conservative40–75-minute estimate and2h
limit), no guard tripped. Saved accounting. No new job/default or budget increase.

[Report](../experiments/cpu-streamed-constants/report.md),
[manifest](../experiments/cpu-streamed-constants/runs/file-01/results.json),
[analysis](../experiments/cpu-streamed-constants/runs/file-01/analysis.json),
[accounting](../experiments/cpu-streamed-constants/runs/file-01/accounting.log).
Recorded-cache analyzer passes358 exact source identities, sealed cache/evidence
hashes, full retained CPU/GPU canonical/packed digests, all systematic positions,
16 sampled independent curves and1027 packed readbacks. Controller8 GiB payload
hash retained; analysis checks metadata/size instead of another bulk reread.
Go/reference/file/header/checksum/truncation/path-replacement/chunk/alias/error/
retained-output tests, normal/compact regression suites, Go-only builds, race/vet,
ASAN/UBSAN/TSAN all pass. No kernel corrections required. Caches/profiles retained.

Measured s16 encode1249.616 s (20m49.6s), read/range-check8.963 s for19.422 GiB /
78 logical chunks, not physical disk traffic. RSS35.484 GiB versus prior58.351
GiB (39.19% less RSS). This is not39% total RAM reduction: tracked file pages
sample26.336 GiB max and RSS+tracked pages61.820 GiB max. Batch-task cgroup current
sample47.917 GiB /recorded peak47.919 GiB is charged subtree memory, not the same
as globally shared tracked file residency or whole-node memory. During encode/
verification combined tracked observation~56.104 GiB; packing raises it. Sampling
can miss short/final cache peaks. No matched full resident cache/cgroup baseline.
No claim of a total35 GiB deployment. Initial verification warms backing cache;
no cold-disk measurement or forced eviction.

Matched s14 resident58.8530/58.6672 s versus file59.3773/59.3573 s costs0.89%/1.18%
encoding time (two fixed-order cached samples). RSS3.595→2.214 GiB; inclusive
tracked RSS/cache20.385→19.009 GiB, including common s16 background residency.
Full s16 comparisons with earlier resident1247–1281 s are historical. Opening/
verification/native metadata37.181 s, independent checks178.754 s, pack55.768 s
stay separate; first encode includes workspace allocation. Read prefix overlaps
buffer timing and must not be added twice.

Updated experiment report, root/native/shared README and scaling plan to record
measured costs and cache/accounting limits; README check completed. No arithmetic,
API/default or methodology change in review. Analyzer/self-checks/local report
links/whitespace pass; no redundant long tests. Evidence/source changes staged,
uncommitted. Public file backing still requires immutable verified source, with
stat checks not a hostile-write guarantee. Resident stays default when it fits.

Both selected CPU/GPU s16 streaming gates close. Next capacity work needs bounded/
one-arena layouts plus wider s17 fields/coordinates/artifacts and cache-aware
host/device admission. Current s17 fast CPU two-array/message model~270.42 GiB is
before output copies/cache/overhead. Earlier bounded/one-arena projections are not
implemented s17 support or admitted runs. Application/cryptographic/TRL5 gates
remain open.
