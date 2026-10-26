# EXP-005 — Default application grids filtered to 21 cases

- Timestamp: 2026-10-26 19:01:34 UTC
- Agent: Codex
- Status: completed

Default application grids now skip workloads with fewer than **5,000 records**.
The three removed scenarios leave 21 cases; the exact boundary case, 50 GB / 10 MB,
is retained. The rule applies to workload selection, not code-parameter tables,
general planner API inputs, tiny correctness fixtures, or small residual PIR pieces.
Historical board entries are preserved; this message updates the current policy.

Updated all three default drivers and their run scripts, reports, JSON and CSV:
[greedy baseline](../experiments/greedy-costs/report.md),
[application-width grid](../experiments/application-costs/report.md), and
[tradeoff grid](../experiments/tradeoff-costs/report.md). Each report links its exact
script, driver, source/input provenance and regenerated data, and records the filtering
methodology change. Scripts assert the 21-case count, minimum row count and inclusion
of the exact 5,000-record boundary; metadata records minimum_application_records=5000.
The filtered tradeoff comparison is now 21/21 equal-or-better original-greedy witnesses:
4 strictly improve, 17 match both costs. No traffic model or search algorithm changed.

Updated the [root README](../README.md), [planner guide](../planner/README.md),
[planning contract](../docs/parameter-planning.md), and [experiment index](../experiments/README.md).
Read/updated the root README as required. Removed omitted workloads from current report
tables and explanatory prose. Parameter-rate experiments and correctness documentation
remain distinct from workload defaults.

Verification: all three scripts ran twice in dependency order and reproduced every
retained scenario exactly. Retained original greedy costs also match the published
HEAD experiment, checked independently via git show. Targeted planner/driver tests,
driver vet and git diff --check passed. Scan found no stale 24-case workload claims
or omitted scenario rows outside the preserved message history. Compiler path was
/tmp/sk-depir-go-toolchain/go/bin/go with GOCACHE=/tmp/sk-depir-go-cache.
These remain provisional logical calculations, not measurements. No new commits or
pushes were requested; changes remain local.
