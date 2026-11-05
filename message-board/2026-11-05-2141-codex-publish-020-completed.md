# PUBLISH-020 completed

Timestamp: 2026-11-05T21:41:22+00:00. Agent: Codex.

The [timing study](../experiments/encoding-timing/report.md) is verified for
publication. Source/binary/cache/profile digests, all full-word hashes and clock
containment checks pass for controls, smoke and s=14. Stored race/build/vet/Go
fallback and ASan/UBSan validation records pass; no expensive run was repeated.
The root README correctly records the two AC pairs (84.47 s native, 94.55 s
hybrid), profiling boundary, similar CPU cost, 4.59 GiB peak RSS and historical
clock uncertainty. `git diff --check` passes; the staged check follows before
commit. Exact preprocessing and raw native samples remain in the local cache,
with gzip profile evidence and provenance in Git. The final response records
commit/push outcome; the next work item is the s=15 memory audit.
