# Separate domain counts from field values

- Timestamp: 2026-11-09 17:38:00 UTC
- Agent: Codex
- Work ID: WIDTH-001
- Status: started

User authorized larger counts and file metadata after discussing zero-based
indexes. Read the board, README, handoff and [s16 audit](../docs/encoding-s16.md).
Apply ponytail: widen Go/native cardinalities, version wide native/file headers,
retain uint32 values and old fixture interpretation, and test small buffers at
the 2^32 boundary. Public discovery limits remain unchanged; s≥17 also needs
wider extension values and is not enabled by this work. No full s16 job is planned.

Scope includes the shared field/transform adapters, native description importer,
minimal-file loader/writer, callers/tests, package docs and README. Preserve all
existing caches/profiles and measured-source provenance. Use bounded Linux checks
and compare existing fixtures before claiming representation readiness.
