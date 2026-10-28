# Profile/cache harness verified before long run

- Timestamp: 2026-10-28 18:43 UTC
- Work item: ENCODE-010
- Status: update

Added explicit trusted-local JSON loading with
[LoadCyclicEncoderWithLimit](../encoding/liftrs-shared/orbit_artifact.go), including an
internal [basis limit](../encoding/liftrs-shared/orbit.go). Default 2/64 MiB caps and
structural checks remain unchanged. New [tests](../encoding/liftrs-shared/orbit_cache_test.go)
and [example](../encoding/liftrs-shared/orbit_cache_example_test.go) verify caps,
malformed data, embedded bases, byte/encoding round trips and bounded reads.

The [profile runner/report](../experiments/encoder-profile-s13/report.md) preserves
local complete caches, checksums/provenance and profiling binaries under ignored
`.cache/liftrs/`. It profiles discovery/preparation/cache save/load/one encoding,
actually reloads saved data before encoding and supports later reuse-only runs.
No mathematical algorithm changes. The 2 GiB local file cap is distinct from
12 GiB memory and public artifact budgets. Existing caches are not overwritten.

[Preflight commands](../experiments/encoder-profile-s13/verification.json),
[raw log](../experiments/encoder-profile-s13/verification.log),
[smoke fixture](../experiments/encoder-profile-s13/smoke-results.json) and
[race check](../experiments/encoder-profile-s13/race.log) pass: loader tests,
profile/cache save-load and reuse paths, vet, isolated full suite and liftrs race
checks. README checked/updated, plus package/API docs, encoder contract and
experiment index. `git diff --check` passes. Publish this harness checkpoint
before launching the expensive profiled run; final results remain pending.
