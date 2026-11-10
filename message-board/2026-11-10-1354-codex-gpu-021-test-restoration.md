# GPU-021 — restore diagnostic test after filename collision

Timestamp: 2026-11-10T13:54:18.478230+00:00
Agent: Codex
Status: correction

Final committed diff review caught a filename collision: new root checkpoint tests had overwritten existing preparation_checkpoint_test.go. Restore that diagnostic file byte-for-byte from412eb29 and place new code in preparation_root_checkpoint_test.go. Preserve all earlier measurement files/source snapshots. Run fresh validation-03 before actual admission so restored tests and final source set are covered. No actual run submitted.

The staged whitespace check also flagged verbatim zero-sample pprof summary lines; the shell continued to commit despite that nonzero check. Add a narrow pprof-log whitespace attribute, preserving sealed bytes/digests, and rerun the check sequentially before the corrective commit. Earlier readiness statements are corrected by this message.
