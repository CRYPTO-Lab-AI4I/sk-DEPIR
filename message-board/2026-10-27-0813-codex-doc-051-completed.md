# DOC-051 — Upload optimization scoped; download tables regenerated

- Timestamp: 2026-10-27T08:13:07Z
- Agent: Codex
- Status: completed

## Changes

Added grouped storage layout and its streaming writer/client adapter to TRL 5 scope in
[AGENTS.md](../AGENTS.md), [ROADMAP.md](../ROADMAP.md), [README.md](../README.md), and the
[protocol I/O contract](../docs/protocol-io.md#storage-layout-optimization).
Ordinary reads of each group's same-position values remove repeated addresses;
server-side protocol computation remains outside scope. Linked this goal from
[data representation](../docs/data-representation.md) and documented the target
upload formula separately from the implemented evaluator in
[parameter planning](../docs/parameter-planning.md).

Updated the [tradeoff script](../experiments/tradeoff-costs/run.py) and
[DP script](../experiments/dp-costs/run.py) with `--render-only`, regenerating their
[tradeoff table](../experiments/tradeoff-costs/options.md) and
[DP table](../experiments/dp-costs/options.md) as storage overhead, D/B, and absolute
download. Reference columns show N/K and L (= reference D/B, including dummies).
Both [tradeoff report](../experiments/tradeoff-costs/report.md) and
[DP report](../experiments/dp-costs/report.md) document reproduction and the change.
Saved JSON/CSV, plans, search provenance and total-communication selection are preserved.

## Verification and limits

Ran `python3 experiments/tradeoff-costs/run.py --render-only` and the analogous
DP command. Verified all 21 rows per table against saved download costs and reference
query lengths; repeated rendering was byte-identical. An isolated asymmetric-cost
fixture verified D/B rather than (U+D)/B or C/3. `git diff --check` passed.
README checked and updated for scope and presentation. No Go behavior changed;
Go tests and expensive searches were not rerun. Grouped layout is not implemented
or measured; framing and byte alignment remain excluded. Existing selections have
not been reoptimized for download or the future upload model.
