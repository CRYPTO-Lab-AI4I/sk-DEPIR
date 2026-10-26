# Application width-search cost grid

Default application workloads require **at least 5,000 records**. This report and
its regenerated results now contain 21 qualifying cases; scenarios below that threshold
were removed from the grid. This workload filter does not restrict the planner API or
tiny correctness fixtures. Costs for retained scenarios are unchanged.

## Purpose and reproduction

Evaluate the same 21 application specifications as the [unsliced baseline](../greedy-costs/report.md),
now using `GreedyApplication` with default logical-width search and full-download
residuals enabled at ratio 1. These are **logical cost calculations using estimated
capacities**, not measured storage, network traffic, or encoded databases.

- Runnable script: [run.py](run.py); driver: [main.go](main.go).
- Catalogue snapshot: [pir-configs.json](pir-configs.json).
- Complete plans, evidence, environment and source checksums: [results.json](results.json).
- Exact-bit totals and normalized costs: [summary.csv](summary.csv).
- Implementations: [application planner](../../planner/application.go),
  [greedy kernel](../../planner/greedy.go), [evaluator](../../planner/evaluate.go),
  and [cost formulas](../../planner/costs.go).
- Verification: [application tests](../../planner/application_test.go). The driver
  independently re-evaluates every chosen plan and checks mixed-plan ratio bounds.

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/application-costs/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

Run from any directory; the script resolves repository paths, builds the driver,
and overwrites only this experiment's results and CSV. Use `--go go` when available.
No remote service, external encoder, or plaintext database is required.

Base revision: `8b322b83ddd566a2becd2b38d59793b26d91678d`, with the uncommitted PLAN-007
implementation, PLAN-008 correction and EXP-003 driver identified by SHA-256 in results metadata.
The base revision alone does not contain the new planner. Environment:
go1.27.3, darwin/arm64, Python 3.9.6,
macOS-26.6.2-arm64-arm-64bit. No hardware timing or memory performance is measured.
There is no randomness in this cost calculation; catalogue capacities originate from
recorded Monte Carlo estimates. No seed is required here.

## Parameters and methodology

Decimal GB=`10^9` bytes, TB=`10^12` bytes, KB=`10^3` bytes, MB=`10^6` bytes.
Sizes: 1, 20, 50, 100, 500 GB and 1 TB; record sizes: 1 KB, 10 KB, 1 MB, 10 MB.
Every application size divides exactly by record size. The snapshot includes all
13 `s=8..20,m=2` configurations. Require 100 conjectured security bits at a
`2^50` protocol-query sample limit. Each original lookup consumes one query per
fragment per indexed group, reducing the allowed number of application lookups.

Default width search includes the original record width, field widths, and powers
of two starting at 8 bits. Storage ranks first, total upload plus download second.
`AllowFullDownload=true`, `FullDownloadRatio=1`, no absolute storage/upload/download
limits and no absolute full-piece cap. At most one full-download suffix is considered.
Each piece is read on every original lookup; indexed traffic repeats for all fragments.
The script also records an independently selected plan with full downloads disabled,
using the same width search, as a diagnostic alternative.

Storage is normalized by original application bits. Communication is upload plus
download per original record lookup divided by original application bits. Percentages
below multiply that communication fraction by 100. Actual framing, fixed range
requests, byte rounding, manifests, and metadata are excluded. Logical zero upload
for a whole download does not mean an actual remote request has zero bytes.

## Results

**Methodology correction:** an initial run exposed a flaw in the per-width full-only
comparison: narrow logical records inflate indexed traffic and could qualify even a
whole-terabyte download. PLAN-008 corrected the implementation before producing the
results below. A shared cap now uses the lowest indexed download among no-full greedy
plans over all searched widths, including the original width, with security requirements
but without cost budgets. The selected encoded parts' cap and optional absolute limits
also apply. The driver verifies that full-download bits never exceed original-width
download for this factor-1 grid. The initial erroneous table was replaced, not mixed
with corrected results. The original unsliced experiment remains separate.

| Application | Record | Storage / application | Communication / application per lookup | Mode |
|---|---|---:|---:|---|
| 1 GB | 1 KB | 5.436× | 36.0427% | encoded |
| 1 GB | 10 KB | 5.436× | 360.1119% | encoded |
| 20 GB | 1 KB | 4.796× | 4.5039% | encoded + residual |
| 20 GB | 10 KB | 4.793× | 43.8575% | encoded + residual |
| 20 GB | 1 MB | 4.552× | 3080.1216% | encoded + residual |
| 50 GB | 1 KB | 4.553× | 3.2434% | encoded |
| 50 GB | 10 KB | 4.553× | 32.4879% | encoded |
| 50 GB | 1 MB | 4.553× | 3248.7102% | encoded |
| 50 GB | 10 MB | 3.965× | 24548.1440% | encoded + residual |
| 100 GB | 1 KB | 4.679× | 2.6758% | encoded |
| 100 GB | 10 KB | 4.679× | 26.7351% | encoded |
| 100 GB | 1 MB | 4.671× | 1876.3931% | encoded + residual |
| 100 GB | 10 MB | 4.241× | 5921.4037% | encoded + residual |
| 500 GB | 1 KB | 4.142× | 1.2786% | encoded + residual |
| 500 GB | 10 KB | 4.124× | 12.6861% | encoded + residual |
| 500 GB | 1 MB | 4.071× | 1085.6027% | encoded + residual |
| 500 GB | 10 MB | 3.943× | 9442.0986% | encoded + residual |
| 1 TB | 1 KB | 4.127× | 0.7179% | encoded |
| 1 TB | 10 KB | 4.127× | 7.1633% | encoded |
| 1 TB | 1 MB | 4.127× | 716.2521% | encoded |
| 1 TB | 10 MB | 4.127× | 7162.5163% | encoded |

## Interpretation and limitations

With full downloads disabled, storage is roughly 4.1–5.4 times application size
across this grid, with much less dependence on original record width. That supports
the purpose of split/stack rearrangement. The full-download policy no longer selects whole-terabyte downloads. However, split/stack
choices can incur indexed communication larger than the entire database when records
are large. Full-download and fragment choices trade storage for communication; this objective deliberately prioritizes storage. Detailed upload,
download, fragment counts, logical widths, residual sizes, and indexed-only alternatives
are retained in the linked CSV and JSON.

All encoded target plans remain provisional: empirical capacities do not establish
executable encoder capacity or production security. Full-only plans have no indexed
configuration and therefore are not labelled provisional by the evaluator; this does
not mean a new full-download backend or cryptographic suite exists. The heuristic
search is not optimal. Costs are deterministic for this source/snapshot; timestamps,
revision metadata and environment values may differ on reproduction. The historical
unsliced experiment is preserved unchanged apart from its previously recorded context note.
