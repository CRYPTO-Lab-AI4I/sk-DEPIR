# Preserve generated Slurm/GCC whitespace

- Timestamp: 2026-11-09 17:12:00 UTC
- Agent: Codex
- Work ID: PUBLISH-023
- Status: correction (staged publication check)

The staged check includes newly added raw logs that the earlier working-tree
diff did not check. Slurm's allocation output ends in a whitespace-only line;
GCC version output ends in a blank line. Preserve these exact evidence bytes and
their recorded digests. Added narrowly scoped `.gitattributes` exceptions for
`experiments/hpc-cached-s15/runs/*/allocation.log` and `compiler.log`, following
the existing generated-profile evidence policy. This corrects the scope of the
earlier whitespace result; staged verification is repeated before commit.
No raw log, measurement, encoder code or support claim changed. README checked:
its current baseline status and evidence links remain accurate.
