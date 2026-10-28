# s=13 profiles and retained cache completed

- Timestamp: 2026-10-28 20:02 UTC
- Work item: ENCODE-010
- Status: completed

## Publication and retention

Published the previous s=13 benchmark and standing profiling/retention policy as
`3385c0d`, then the tested local-cache API and profiling harness as `0d94d4c`,
before starting the expensive run. [AGENTS.md](../AGENTS.md) and
[experiment conventions](../experiments/README.md#profiling-and-preprocessing-retention)
retain the user's default: profile during optimization, preserve reasonably sized
preprocessing with provenance/checksums, document exceptions, and keep large local
files outside Git. Disable profiling only when optimization is complete.

The cache is retained at `.cache/liftrs/s13-m2-t5/cyclic.json`, 453243425 bytes
(432.25 MiB), under the 2 GiB local file budget. Metadata, source manifests and
matching profiling binaries remain alongside it. Explicit local JSON loading
preserves default 2/64 MiB limits and structural checks. This cache is actually
reloadable; no useful preprocessing was deleted or silently overwritten.

## Profiles and results

The [report](../experiments/encoder-profile-s13/report.md),
[runner](../experiments/encoder-profile-s13/run.py),
[analysis script](../experiments/encoder-profile-s13/analyze.py),
[CPU summary](../experiments/encoder-profile-s13/analysis.json),
[table](../experiments/encoder-profile-s13/table.md) and
[evidence manifest](../experiments/encoder-profile-s13/evidence-manifest.json)
link exact sources, raw profiles/text views, parameters, environment and checks.

- Profiled discovery 85.109 s, preparation 1784.863 s (29.75 min), save/sync/hash
  1.498 s, first actual reload 4.864 s.
- The setup monitor failed on an optional `samples` field after loading. Its
  [corrected result](../experiments/encoder-profile-s13/setup/results.json),
  initial metadata and exact runner snapshot preserve that failure. Completed
  discovery/preparation/save/load profiles are valid; its zero-byte encoding
  profile is incomplete and excluded. Setup process peak RSS is unknown.
- Fixed the Python logger and partial-result preservation. The
  [logger regression](../experiments/encoder-profile-s13/test_run.py) passes
  against real fresh/reuse fixture metadata. No Go mathematical code changed.
- [Reuse-only run](../experiments/encoder-profile-s13/reuse-01/results.json):
  reload 4.834 s, profiled encoding 186.352 s, whole case 194.28 s and 2.257 GiB
  peak RSS. Discovery/preparation were skipped. Complete canonical/packed hashes
  match baseline; all systematic symbols, 16 independent curves, sampled mapping
  and packed sync/readback pass. No resource guard fired.
- Setup and encoding are separate processes: no combined wall time or peak RSS
  is claimed. Profiled times do not replace the unprofiled latency baseline.

Flat self CPU samples for encoding: fixed-constant lookup kernel 44.20%, general
base multiplication 26.48%, transform recursion 17.51%. Preparation: 38.46%,
27.14%, 17.56%, respectively. Cumulative shares overlap and are not wall fractions.
Next candidate: compare lookup unrolling/row widths under a stated table-memory
budget, then validate full encoding from the retained cache. No new optimization
or speedup was implemented by profiling.

## Verification and documentation

Preflight loader/driver tests, fresh/reuse smoke fixture, race checks, vet and
isolated full Go suite passed; their records are linked by the report. The monitor
fix has a passing regression and the completed actual reuse run. No Go source
changed afterward. `analyze.py` verifies source snapshots, raw events, complete
profile/binary/cache hashes and baseline agreement; documentation targets,
evidence manifest and `git diff --check` pass. The cache is ignored by Git.

README checked and updated from pending to measured status. Also updated the
[encoder contract](../docs/encoder-contract.md),
[algorithm plan](../docs/encoding-algorithms.md) and
[package README](../encoding/liftrs-shared/README.md). Defaults, publication-size budget,
application integration and s>=14 limits remain distinct. New profile results,
monitor fix and final report are uncommitted; the requested pre-run checkpoints
were already pushed.
