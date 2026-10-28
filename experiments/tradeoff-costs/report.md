# Capacity-ordered batched tradeoffs

## Purpose and reproduction

Calculate storage/download alternatives for 21 application workloads, each with
at least **5,000 records**. Correctness fixtures and the planner API are not restricted
by this workload filter. These are provisional logical estimates, not actual database
encodings, network measurements, or security validations.

- Script: [run.py](run.py); driver: [main.go](main.go).
- Catalogue snapshot: [pir-configs.json](pir-configs.json).
- Full plans, witnesses, environment and source hashes: [results.json](results.json).
- Per-alternative exact costs and flags: [summary.csv](summary.csv).
- Selected options with nearest catalogue overhead: [options.md](options.md), generated
  by the same script, preserving an original-greedy witness in each displayed row.
- Previous boundary-search results: [boundary-search-results.json](boundary-search-results.json).
- Implementation: [tradeoffs.go](../../planner/tradeoffs.go), [batches.go](../../planner/batches.go),
  [costs.go](../../planner/costs.go), [evaluate.go](../../planner/evaluate.go),
  [application.go](../../planner/application.go), [greedy.go](../../planner/greedy.go).
- Tests: [search tests](../../planner/tradeoffs_test.go), [batch tests](../../planner/batches_test.go),
  [compact/expanded tests](../../planner/greedy_test.go).

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/tradeoff-costs/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

Run from any directory; the script builds the driver and overwrites current results,
CSV and selected-options Markdown. It uses Python 3, Go and git, with no network service
or external encoder. Base revision is `8b322b83ddd566a2becd2b38d59793b26d91678d`, plus
the uncommitted implementation through PLAN-011. Exact source/input checksums and dirty
state are recorded; the base revision alone does not contain this search.
Environment: Go 1.27.3, darwin/arm64, Python 3.9.6, macOS 26.6.2. No hardware timing
or peak-memory claims are made. Reproduction metadata may vary across environments.

## Methodology and parameters

Use decimal GB=10^9 bytes, TB=10^12, KB=10^3, MB=10^6. Application sizes are 1, 20, 50,
100, 500 GB and 1 TB; record sizes are 1 KB, 10 KB, 1 MB and 10 MB, retaining only
pairs with at least 5,000 records. All retained sizes divide exactly. Use all 13
catalogue configurations, requiring 100 conjectured computational bits at a 2^50
protocol-query budget. Capacities remain Monte Carlo estimates; no new sampling
is performed here. Fragment queries consume the sample budget per protocol query.

**PLAN-011 methodology change:** replace ascending historical row-position admission
with descending-capacity configuration stages. Batch repeated full blocks of a code,
then pass remaining rows to smaller codes. Full blocks can be reordered without cost
changes because row ranges are disjoint and query sharing does not reduce traffic.
A final padded block can use any configuration; alternatively download the remaining
suffix once per original lookup. This removes permutations of equivalent full-block
choices. Small codes remain available for residuals.

Defaults: MaxFrontier=32, MaxStates=256 live partial plans per configuration stage,
MaxCounts=8 count choices per state/configuration, MaxBaselineBlocks=256. Enumerate all
counts if they fit; otherwise sample zero, one, filling counts and powers of two near
both ends. Within equal-stage/equal-remaining-row states, prune by dominance. Excess
cost frontiers use geometric download buckets; excess live states sample remaining
row keys and alternate their cost endpoints. These are approximations with no global
or multiplicative optimality guarantee. The previous 256-boundary cutoff is removed.
The API now uses state/count limits and flags, not boundary-admission options/flags.

Repeated runs use RowBlock.Repeat: divide the row range into equal consecutive blocks,
with a separate shorter padded final block. Zero repeat means one. Every repetition
still incurs separate storage, masks and uploaded addresses. Evaluation costs runs in
aggregate; no per-repetition allocation is required. No runtime encoder integration
or durable manifest implementation is implied by this planning representation.

Completed feasible greedy baselines and whole-application download remain protected
from approximation. Final exact dominance pruning retains an equal-or-better witness
for each protected baseline. Output/display targets can be exceeded to preserve them.
The script also checks against the actual [original greedy grid](../greedy-costs/results.json).

Storage overhead is stored bits / original application bits. Search optimizes
storage and download. The options table displays `D/B`, the inverse download
communication rate, and absolute download. Total communication rate `B/(U+D)`
remains a separate JSON/CSV accounting field. Indexed
traffic repeats for every fragment, including cover queries; full pieces are read
once. Logical costs exclude framing, metadata, identifiers and byte rounding. Whole-read
requests have nonzero physical headers even though logical index upload is zero.

The reference configuration minimizes absolute difference between application bits
and catalogue plaintext capacity K*s, breaking ties by configuration ID. Its reference
overhead is N/K, assuming full utilization, excluding application padding. It does not
assert that one configuration encodes the entire application at that overhead.

## Results and interpretation

See [the selected-options table](options.md) for two nontrivial choices per workload
where available, including catalogue overhead. Full data include every retained
alternative and approximation flags. The two displayed options include an original
greedy witness and the minimum-download endpoint; when identical, also show its
adjacent lower-storage point. This is not a balanced recommendation or a minimum-storage
claim. Whole-download-only options are omitted from that display, but remain candidates.

All **21/21** original greedy baselines retain equal-or-better witnesses in results
and representative selections: **3 strictly improve**, **18 match storage/download costs**.
Minimum-download endpoint costs match the archived total-objective run in all
21 cases. The displayed lower-storage representative changes for 20 GB / 10 KB;
all 21 intermediate frontier sets change. See the objective comparison below.

All 21 cases report state reduction and sampled count choices; no baseline construction
is interrupted. The search remains approximate. Do not interpret eliminating the old
boundary flag as eliminating truncation, or claim that every intermediate old option
is dominated by the new explored frontier. Protected baseline guarantees still hold.

## Verification and limitations

Two runs reproduced all 21 scenarios and generated option tables exactly. The driver
independently re-evaluates compact plans, checks costs and original input sizes, rejects
dominated/duplicate alternatives, and validates protected/display witnesses. Tests compare
untrimmed tiny searches with independent exhaustive enumeration, compare repeated runs
with expanded physical blocks, verify padded batches and billion-block compact arithmetic,
and exercise count/state bounds, deterministic ordering, budgets and overflow.

Directional budgets remain supported. Sampling can miss feasible or superior plans;
no automatic communication budget is selected. Target encoders, production security and
actual I/O feasibility remain unvalidated. The archived boundary-search JSON records its
original source hashes; current reproduction uses the revised sources, not that method.

## Download-objective rerun (2026-10-28)

The current implementation optimizes storage and **download**, throughout ranking,
dominance, frontier reduction and DP windows. Upload/total bits are still reported
using the existing repeated-address model; grouped storage remains a future backend.
No assumption of zero actual upload is introduced into accounting.

Rerun with the command above. `--render-only` only renders existing results and
must not be used to claim a new optimization run. The script/source hashes in
[results.json](results.json) identify this rerun (base revision `2367b92` plus
PLAN-014 changes). All configurations use the unchanged historical
[pir-configs.json](pir-configs.json) snapshot, including estimated K=3117 at s=8;
the current catalogue's separate exact K=3121 evidence is intentionally not mixed
into this objective-only comparison.

The previous total-objective [results](total-objective-results.json),
[CSV](total-objective-summary.csv) and [table](total-objective-options.md) are
archived. Their scripts and source are preserved in repository revision `2367b92`;
they were not generated by the current script. This report's older methodological
observations/comparisons remain historical where explicitly referring to that run.
The [objective comparison](../download-objective-comparison/report.md) provides a
reproducible comparison of all 21 old/new scenarios, endpoints, displayed options
and intermediate frontier points, using exact saved costs.

The reference configuration is still selected by closest plaintext capacity `K*s`.
Its table column displays `N/K` and inverse download rate `L` (including dummies),
assuming full use of that code. Stars protect storage/download costs against the
original greedy baseline, not upload or total communication.
