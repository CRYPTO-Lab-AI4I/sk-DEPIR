# Download-objective rerun comparison

## Purpose and reproduction

Change optimization from storage versus upload+download to storage versus download,
then quantify changes using identical catalogue inputs. Upload and total-traffic
accounting remain the original repeated-address model. Grouped storage is still a
planned backend; these are logical cost calculations, not network measurements.

Run all four grid scripts first, from the repository root:

```sh
python3 experiments/greedy-costs/run.py
python3 experiments/application-costs/run.py
python3 experiments/tradeoff-costs/run.py
python3 experiments/dp-costs/run.py
python3 experiments/download-objective-comparison/run.py
```

Each cost script accepts `--go /path/to/go`. The exact comparison script is
[run.py](run.py), its [results](results.json) record all input SHA256 hashes and
its [table](table.md) summarizes changes. Comparison requires only Python 3 and git.
Reruns require the Go toolchain and dependencies described in each linked report.
Checkpoint: revision `2367b92` plus PLAN-014 working changes; exact search source
hashes, environment and commands are recorded in each regenerated results file.
The comparison itself is deterministic except timestamp/provenance fields.
The run used macOS 26.6.2 ARM64, Python 3.9.6 and Go 1.27.3; no timings are compared.

## Inputs and methodology

Current scripts/reports: [greedy](../greedy-costs/report.md),
[application width search](../application-costs/report.md),
[bounded tradeoffs](../tradeoff-costs/report.md), [cost DP](../dp-costs/report.md).
Old results, summaries and option tables are retained in each experiment directory
with `total-objective-` prefixes and linked from those reports. Their generating
source is available at revision `2367b92`; current scripts generate the download
objective. This comparison never regenerates or modifies historical archives.

Use all 21 fixed workloads with at least 5,000 application records, decimal units,
and the same historical 13-configuration catalogue snapshots. In particular, keep
estimated s=8 K=3117 fixed in both sides; the current catalogue's independently
verified K=3121 is separate evidence and not part of this controlled comparison.

Compare exact stored and downloaded bits at the endpoint, and sets of retained
(storage,download) frontier points. For display comparisons use the exact generated
rows, which include both representative options and their rounded costs. Intermediate
frontiers are approximate outputs; changed points do not imply a different feasible
layout family or global optimality. DP coordinate rounding and bounded-search bucket
reduction now use download and can therefore retain different alternatives even if
the endpoints remain identical.

## Results

All 21 endpoint storage/download costs match their historical counterpart in all
four grids. Each search table changes one displayed row: **20 GB / 10 KB**, whose
lower-storage representative changes from **8.652x storage, 65.522 MB download**
to **8.684x storage, 58.842 MB download**. This is about **0.37% more storage** and
**10.2% less download**; neither point dominates the other. The minimum-download
option in that row remains unchanged.

All 21 retained intermediate frontier sets change in both search experiments.
See results.json for exact added/removed counts. Thus endpoint stability does not
mean the search change had no effect; it changes approximation choices and some
returned alternatives. The search still protects greedy baselines using storage
and download. Upload or total-communication dominance is no longer promised.

## Implementation and verification

The search now compares download in ranking, Pareto pruning, frontier buckets,
DP seed selection, windows and cells. Greedy storage ties also use download.
Explicit DP options are `DownloadCells` and `DownloadWindowBits`; metadata uses
`download_bits` and `download_unit_bits`. DP supports storage/download budgets;
only an additional upload budget requires a third dimension and is rejected.
Bounded search preserves upload as a constraint dimension when that limit is used.

Tests include cases where upload makes total-traffic ranking oppose download
ranking, independent exhaustive small searches, exact-unit DP, and budget checks.
Historical accounting formulas and `CommunicationRate=B/(U+D)` are unchanged;
reported inverse download rate is `D/B`. Neither encoder arithmetic nor physical
storage layout changed as part of the objective rerun.
