# PLAN-014 — Download search objective, exact capacity and documentation reconciled

- Timestamp: 2026-10-28T10:31:30Z
- Agent: Codex
- Status: completed

## Search implementation and evidence

Changed [cost tie-breaks](../planner/costs.go), [greedy descriptions](../planner/greedy.go),
[application selection](../planner/application.go), [bounded search](../planner/tradeoffs.go)
and [cost DP](../planner/dp.go) to optimize storage/download. Dominance, sorting,
frontier buckets, seed selection and numerical windows use download. Upload and
CommunicationBits=U+D accounting are unchanged; total CommunicationRate remains
B/(U+D). When bounded search has an upload budget it preserves that constraint
coordinate. DP now supports download limits in its two axes and rejects only
upload budgets; explicit options/metadata are renamed DownloadCells,
DownloadWindowBits, DownloadBits and DownloadUnitBits. Documented the experimental
API change in [planner README](../planner/README.md).

Reran all four 21-case grids, updating their scripts, outputs and reports:
[greedy](../experiments/greedy-costs/report.md),
[application](../experiments/application-costs/report.md),
[tradeoff](../experiments/tradeoff-costs/report.md),
[DP](../experiments/dp-costs/report.md). Prior JSON/CSV/tables are archived under
`total-objective-` filenames; their source is revision 2367b92. Historical catalogue
snapshots remain unchanged, intentionally isolating objective changes from the new
s=8 exact capacity. Current source hashes match regenerated results.

Added the reproducible [objective comparison](../experiments/download-objective-comparison/report.md)
with [script](../experiments/download-objective-comparison/run.py),
[results](../experiments/download-objective-comparison/results.json) and
[table](../experiments/download-objective-comparison/table.md). All endpoint
storage/download costs are unchanged in all 21 workloads for all four grids.
Each search display changes one row: 20 GB / 10 KB lower-storage representative
moves from 8.652x/65.522 MB to 8.684x/58.842 MB (about +0.37% storage, -10.2%
download). Intermediate frontier sets change in all 21 rows: tradeoff adds/removes
226/165 points, DP 262/219. Each retains storage/download witnesses for all original
greedy baselines (3 strict improvements, 18 equal). These results do not guarantee
upload/total dominance or global optimality.

## Catalogue and documentation

[Current catalogue](../pirconfig/configs_catalogue.json) records exact K=3121,
polynomial 0x11d, verified capacity and standalone orbit support for s=8,m=2,t=8.
Preserved Monte Carlo evidence separately, including its historical point K=3117.
The other 12 target capacities/statuses remain estimated/candidate. This does not
claim integrated application execution, grouped storage or production security.
[Catalogue tests](../pirconfig/catalogue_test.go) connect exact capacity and output
agreement to saved encoder evidence; [example](../pirconfig/example_test.go) updated.

Narrowed [PrepareCyclicFFT's comment](../encoding/liftrs-go/orbit.go): only G*H is
O(N log² N); H=(P*M) mod X^K retains Karatsuba. Updated the
[five-step plan](../docs/encoding-algorithms.md#6-five-step-implementation-plan)
with fast/truncated multiplication, preprocessing complexity checks, stage-specific
memory and packed/grouped I/O measurements, exact CPU/storage metadata, and
explicit later scope gates. Public artifact publication requires reproducible
construction, provenance and mathematical cross-checks beyond trusted loader
shape validation. No new artifact verifier is claimed implemented.

Reconciled [encoder contract](../docs/encoder-contract.md),
[planner design](../docs/parameter-planning.md), [roadmap](../ROADMAP.md),
[catalogue README](../pirconfig/README.md), [rate report](../experiments/liftrs-rates/report.md),
[encoder report](../experiments/encoder-comparison/report.md), [experiment index](../experiments/README.md),
and [README](../README.md). README checked and changed for objective, DP budgets,
exact evidence, and comparison links. Encoder measurements were not rerun: its only
source change is the complexity comment; report identifies the prior measured source.

## Verification and remaining decisions

- Planner/catalogue tests passed, including reverse upload/download ranking,
  exhaustive tiny search, exact-unit DP and download-budget checks.
- `go test -race ./planner ./pirconfig` and targeted `go vet`: passed.
- Full `go test ./...` in an isolated checkout: passed; legacy CSV side effects
  remain outside this workspace. Existing error fixtures were updated to the new
  metric/API instead of weakening expectations.
- Four complete grid reruns, old/new comparison, source/input hashes, edited
  relative file links and `git diff --check`: passed.

No immediate user decision is needed for CPU work. After CPU checkpoints, review
measured needs before starting deferred external-memory/GPU work; a GPU vendor,
native dependency and broader TRL 6 scope are not selected. The proposed publication
budget remains configurable. Working changes are not committed/pushed in this turn,
since the request is implementation and review correction, not publication.
