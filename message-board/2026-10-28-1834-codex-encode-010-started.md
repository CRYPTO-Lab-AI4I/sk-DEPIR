# Profile s=13 and retain reloadable preprocessing

- Timestamp: 2026-10-28 18:34 UTC
- Work item: ENCODE-010
- Status: started

After publishing `3385c0d`, implement an s=13 runner that profiles discovery,
reusable cyclic preparation and one encoding, retaining a local complete cache
across process exit. Existing JSON output is adequate for this checkpoint but
its default 2/64 MiB basis/encoder loading caps are not. Add an explicit opt-in
cyclic loader limit, preserving default caps and all structural/trust checks.
Use a 2 GiB cache-file cap within the 12 GiB project memory budget; keep caches
under ignored `.cache/liftrs/`, with provenance/checksums and no silent eviction.

Test cache limit handling and round trips before the long run. Profile CPU stages
separately, save and reload the cache outside encoding, and compare the resulting
complete output with the [unprofiled s=13 baseline](../experiments/cyclic-s13/report.md).
Keep source/profiling data, timing boundaries and local cache metadata in a new
`experiments/encoder-profile-s13/` report and scripts. Preserve unprofiled results;
profiled wall times include sampling overhead. Update README, API comments and
package/design documentation. No encoder arithmetic or code parameters change.
