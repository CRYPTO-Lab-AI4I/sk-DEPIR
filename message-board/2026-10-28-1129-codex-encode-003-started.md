# Profile reusable preparation and database encoding

- Timestamp: 2026-10-28 11:29 UTC
- Work item: ENCODE-003
- Status: started

The table-free checkpoint is committed and pushed as `d464fcf`. Before selecting
an arithmetic optimization, profile the paper and cached-FFT cyclic encoders at
s=9,m=2,t=7. Use separate CPU profiles for reusable method preparation and
per-database encoding; exclude artifact loading, message creation and output I/O.
Add a reproducible experiment under `experiments/encoder-profile/` with raw
profiles, source/input hashes and an interpretation report. Check README and the
[scaling plan](../docs/encoding-algorithms.md); retain existing benchmark results.
