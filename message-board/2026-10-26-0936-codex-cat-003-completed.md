# CAT-003 — Table 3 estimates incorporated

- Timestamp: 2026-10-26 09:35:42 UTC
- Agent: Codex
- Status: completed

Read the complete [Table 3](../paper.pdf#page=66) and its description in Section
4.2.3. Previous catalogue transcription used only Table 1. Corrected the
[target catalogue](../pirconfig/configs_catalogue.json) to use empirical rates `0.2024` for
`s=15,m=2,t=5` and `0.2367` for `s=18,m=2,t=5`. Retained both Table 1 and Table 3
references, documenting which supplies each point estimate. Estimated integer
capacities use the already documented rounding rule; encoding support is unchanged.

Table 3 has no `m=2,s=12,t=6` or `m=2,s=20,t=5` row, so those entries retain their
Table 1 estimates. Its other rows do not match the current target security profile.
The table does not supply sample counts or confidence intervals; no such metadata
was invented and no new experiment results were added.

Updated [transcription tests](../pirconfig/catalogue_test.go), [catalogue README](../pirconfig/README.md),
and [planner documentation](../docs/parameter-planning.md). Checked and updated the
[root README](../README.md) to refer to both tables.

Verification: Go 1.27.3, `GOCACHE=/tmp/sk-depir-go-cache`,
`go test ./catalogue -count=1` passed; the temporary schema-test environment's
`python catalogue/schema_test.py` passed all three tests. `gofmt` and
`git diff --check` passed. No loader/encoder implementation changed, so the full
suite was not repeated for this metadata correction.
