# Preprocessing components76756 verified

- Timestamp: 2026-11-17T15:19:41.334027+00:00
- Agent: Codex
- Work: SCALING-104
- Status: Completed

Job76756 COMPLETED0:0 on HPCgnode04,4H200/4CPU/8GiBhost,11min44s scheduler,
702.609s controller. Producer10b5f9a. **578source/45evidence/377cache identities**
reconcile without overrides, including actual CPU/reference/canonical-output
hashes. Eight transform fixtures, three product fixtures, one >32bit-coordinate
witness;171normal/76sanitized records. Both CUDA memchecks report0errors and
0leakedbytes. CPU/native/race/ASan/vet, root/cache restarts, CUDA-free readback,
input/cap rejection and subsequent reuse pass. No execution/arithmetic changes.

[Report](../experiments/preprocessing-tiles/report.md#completed-component-validation-scaling-104)
links original raw data, retained CPU/native profiles, seals/review and accounting.
Normal command110.255s; memcheck347.750s; maximum sampled process RSS676.516MiB;
592.426MiB cache retained outside Git. Native host gprof is not GPU/wall profiling.
Sanitizer's single wide-row witness takes316.264s with only start/end work
counters: documented logging gap, controller heartbeats not counted as progress.
Instrument/split such vectors before future long-vector jobs. Original logs remain
byte-identical, including raw profiler formatting/whitespace.

Two warm profiled four-H200 samples at2^22points: resident forward235.820–236.090ms
versus host-backed8MiB/card1.303–1.305s (about5.53times latency), owned GPU
allocations70.276→38.276MiB aggregate, plus32MiB host staging. Each tiled direction
uploads3230.969MiB/downloads327.914MiB for a32MiB array. Product539.435–539.450ms
versus fourteen split products3.019–3.020s (about5.60times latency), owned GPU
allocations102.276→30.414MiB, unchanged16MiB host output plus4MiB-minus4bytes
temporary. Ratios use sample means; fixed order/profile overhead and short vector
lengths prevent extrapolation to fulls18 or whole preprocessing.

Updated actual main scaling reference (including s18 preprocessing table),
active plan steps3/4/8 and CUDA/experiment docs. README checked and updated.
Fulls17 passed cells remain unchanged. Fulls18 split final-product projection
1536→384GiB arrays and64GiB-minus4bytes extra host temporary remain projections,
not workflow support. Next is step5 four-H200 hybrid vs sixteen-H200 distributed
workflow comparison; current tiled Taylor gathers/launch count need improvement.
No new job, slower default, mode adoption or budget increase.

Verification: analyzer passes complete seals/coverage/numerical/admission checks;
2 Python review tests pass; changed-document whitespace check passes. Raw sealed
logs are preserved rather than normalized. Scheduler accounting read once after
completion ping, no monitoring. All small raw evidence and review outputs are
tracked; cache fixtures/profiles remain referenced outside Git.
