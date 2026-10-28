# s=13 cache retained; resume encoding profile after logger fix

- Timestamp: 2026-10-28 19:16 UTC
- Work item: ENCODE-010
- Status: update

Profiled discovery (85.11 s), preparation (1784.86 s), cache save (1.498 s) and
actual reload (4.864 s) completed. Retained complete cache is 453243425 bytes
(432.25 MiB) under ignored `.cache/liftrs/s13-m2-t5/`, with SHA-256 and build
provenance. Preparation/save/load CPU profiles are complete.

The Python monitor expected discovery `samples` in a post-load code event,
raising KeyError before a complete encoding. It killed its child safely.
[Setup evidence](../experiments/encoder-profile-s13/setup/results.json) records
that failure; original running metadata and exact runner source are archived.
The zero-byte encoding profile is incomplete and excluded. Whole-case RSS was
not captured by that failed monitor; do not infer it from partial data.

Fix the logger to use post-load metadata, add a regression against the real
[smoke fixture](../experiments/encoder-profile-s13/smoke-results.json), and save
partial results on unexpected monitor errors. Resume with `--reuse-cache` in a
new run directory; do not repeat expensive preparation or overwrite its data.
No Go encoder or cache contents changed. The final report must distinguish the
two runs and their phase measurements.
