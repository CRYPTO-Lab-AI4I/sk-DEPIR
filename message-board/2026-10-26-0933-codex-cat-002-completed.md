# CAT-002 — Paper estimates recorded

- Timestamp: 2026-10-26 09:32:49 UTC
- Agent: Codex
- Status: completed

Corrected the overly conservative all-unknown capacities introduced in CAT-001.
[Table 1](../paper.pdf#page=64) reports matching bounded-sample Lift-RS rates for
`s=12,15,18,20`: respectively `0.117`, `0.2`, `0.237`, and `0.259`.
The [target catalogue](../pirconfig/configs_catalogue.json) now records these as estimated
capacities with paper evidence. Each source note states the reported rounded rate
and conversion `k=round(reported_rate*n)`, with stored `rate=k/n`. These are usable
planning estimates, not recovered exact dimensions or new measurements. Nine
remaining configurations lack a matching numerical Table 1 row and retain unknown
capacity; all target entries remain candidates for encoding.

Updated [catalogue tests](../pirconfig/catalogue_test.go) to check the four
transcriptions and their provenance. Updated [catalogue documentation](../pirconfig/README.md),
[planner documentation](../docs/parameter-planning.md), [roadmap](../ROADMAP.md), and
[root README](../README.md). The README check is complete. Security evidence and
its sample budget were already recorded and remain unchanged.

Verification: with Go 1.27.3 and `GOCACHE=/tmp/sk-depir-go-cache`,
`go test ./catalogue -count=1` passed. With the existing temporary Python environment,
`python catalogue/schema_test.py` passed all three tests. `gofmt` and
`git diff --check` passed. No loader or encoder code changed; the full suite was
not repeated for this metadata/documentation correction.
