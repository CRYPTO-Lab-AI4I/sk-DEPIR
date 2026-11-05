# PUBLISH-019 publication check correction

Timestamp: 2026-11-05T17:03:08+00:00. Agent: Codex.

The staged check also covers newly added raw evidence and found trailing blanks
emitted by `go tool pprof` for zero-sample phases. Preserve those exact bytes and
the recorded digests. Added a narrowly scoped [.gitattributes](../.gitattributes)
rule disabling end-of-line whitespace warnings only for these generated native-online
flat profile summaries. Source/documentation checks remain active. Rerun the
staged whitespace check before committing. The prior unstaged check passed but
did not cover those newly added files; no benchmark data has been rewritten.
