# Greedy planner: application-size cost grid

## Sources and reproduction

This experiment calculates logical costs for the agreed 24 application/record-size
combinations. It does not allocate, encode, upload, or retrieve these databases.

- Script: [run.py](run.py); Go driver: [main.go](main.go).
- Exact catalogue snapshot: [pir-configs.json](pir-configs.json).
- Results with complete compact plans and evidence: [results.json](results.json).
- Flat exact-bit summary: [summary.csv](summary.csv).
- Implementation: [greedy partitioner](../../planner/greedy.go),
  [shared cost formulas](../../planner/costs.go), and [evaluator](../../planner/evaluate.go).
- Tests: [greedy/compact-plan tests](../../planner/greedy_test.go) and
  [evaluator/reference-read tests](../../planner/evaluate_test.go).

Base revision: `e559398b4544eb0b826289a83e92cbdf85f09443`, plus the uncommitted
PLAN-005 implementation saved with this report. Results record dirty-tree status
and SHA-256 hashes of all computational sources and the exact input snapshot.
The implementation is not claimed to exist in the base revision alone.

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/greedy-costs/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

Run from the repository root or another directory; the script resolves all paths.
Use `--go go` when the compiler is on PATH. The recorded environment is Go 1.27.3,
Python 3.9.6, macOS 26.6.2, Darwin arm64. The script builds a fresh driver before
running it. Git records revision metadata. No network service or external encoder
is needed. No hardware performance timings or memory measurements are reported;
these are deterministic model calculations. Repeated runs reproduce plans and costs;
timestamps, revision/dirty flags, and environment metadata may differ.

## Inputs and method

Use decimal units throughout: GB=`10^9` bytes, TB=`10^12` bytes, KB=`10^3` bytes,
MB=`10^6` bytes. Application sizes are 1, 20, 50, 100, 500 GB and 1 TB, with
records of 1 KB, 10 KB, 1 MB and 10 MB. Each size divides exactly by each record
size; no application-record rounding is needed. Rows equal application bytes divided
by record bytes, and record width equals record bytes times eight. No record packing
is performed.

Use all 13 `s=8..20,m=2` configurations, the catalogue's measured point capacities,
and requirements of at least 100 conjectured computational bits at a `2^50`-query
sample budget. These requirements use existing catalogue estimates, not a new
security analysis. All results remain provisional because capacities are estimated
and target encoders are not implemented.

At each row position, try every eligible next configuration and score its block plus
a greedy completion. The completion repeatedly fills the largest-fitting row capacity,
batching costs by division, then covers the final remainder using the cheapest
covering configuration. Choose minimum full-plan storage, break ties by total upload
plus download, then next-block height and configuration ID. Continue without
backtracking. No cost caps are imposed for this grid, and encoding time/memory do
not affect selection. This is a one-step-lookahead greedy heuristic, not exhaustive
search or a proof of optimality.

Within each row block, divide the entire record width into field-sized column
slices. Record these implicitly, without creating millions of rectangle objects.
A slice still represents its own PIR database with separate mask material. Databases
with the same configuration ID share client query generation across blocks; every
address list is transmitted separately and every database responds on every request.
The evaluator independently checks the resulting compact layout and its counts.

For `J=ceil(record_bits/s)` databases in a block, add:

- Storage: `J*N*s` bits.
- Upload per application-record retrieval: `J*L*log2(N)` bits.
- Download per application-record retrieval: `J*L*s` bits.

Here `N=2^(s*m)` and `L=2^s-1+128`; dummies and repetitions are included. Storage
includes zero-row/column padding. Counts exclude manifests, byte-range alignment,
backend request identifiers/headers, transport framing, and other adapter overheads.
They are not measurements of cloud-service traffic or physical storage consumption.

## Results

Storage is shown in decimal GB; upload/download are decimal MB **per record
retrieval**, not totals for downloading the database. Rates are application bits
divided by modeled storage bits. Display values are rounded; CSV/JSON retain exact
integer bit counts.

| Application | Record | Storage (GB) | Upload (MB) | Download (MB) | Effective rate |
|---|---|---:|---:|---:|---:|
| 1 GB | 1 KB | 9.446 | 11.011 | 5.505 | 10.587% |
| 1 GB | 10 KB | 10.486 | 23.020 | 11.510 | 9.537% |
| 1 GB | 1 MB | 65.536 | 766.000 | 383.000 | 1.526% |
| 1 GB | 10 MB | 655.360 | 7,660.000 | 3,830.000 | 0.153% |
| 20 GB | 1 KB | 134.352 | 33.309 | 16.655 | 14.886% |
| 20 GB | 10 KB | 173.679 | 117.685 | 58.842 | 11.516% |
| 20 GB | 1 MB | 327.680 | 2,044.000 | 1,022.000 | 6.104% |
| 20 GB | 10 MB | 655.360 | 7,660.000 | 3,830.000 | 3.052% |
| 50 GB | 1 KB | 268.704 | 33.055 | 16.528 | 18.608% |
| 50 GB | 10 KB | 439.767 | 340.272 | 170.136 | 11.370% |
| 50 GB | 1 MB | 786.432 | 3,834.000 | 1,917.000 | 6.358% |
| 50 GB | 10 MB | 1,310.720 | 15,320.000 | 7,660.000 | 3.815% |
| 100 GB | 1 KB | 537.408 | 66.110 | 33.055 | 18.608% |
| 100 GB | 10 KB | 671.105 | 166.384 | 83.192 | 14.901% |
| 100 GB | 1 MB | 1,048.576 | 2,302.000 | 1,151.000 | 9.537% |
| 100 GB | 10 MB | 2,621.440 | 30,640.000 | 15,320.000 | 3.815% |
| 500 GB | 1 KB | 2,512.607 | 202.959 | 101.480 | 19.900% |
| 500 GB | 10 KB | 2,684.690 | 330.261 | 165.131 | 18.624% |
| 500 GB | 1 MB | 4,980.738 | 8,184.002 | 4,092.001 | 10.039% |
| 500 GB | 10 MB | 7,864.320 | 38,340.000 | 19,170.000 | 6.358% |
| 1 TB | 1 KB | 4,746.025 | 235.576 | 117.788 | 21.070% |
| 1 TB | 10 KB | 5,369.380 | 660.523 | 330.261 | 18.624% |
| 1 TB | 1 MB | 9,437.187 | 11,002.003 | 5,501.002 | 10.596% |
| 1 TB | 10 MB | 10,485.760 | 23,020.000 | 11,510.000 | 9.537% |

## Interpretation and limitations

Large records create many parallel PIR databases. Compact metadata makes the
planning computation manageable but does not remove their communication. For
example, a 1 GB database of 10 MB records has only 100 rows; the smallest eligible
configuration has estimated capacity 3,117 rows. Its ten million column slices
therefore incur substantial zero-row padding: modeled storage is 655.36 GB, upload
7.66 GB, and download 3.83 GB per retrieval. The identical cost for 20 GB with
10 MB records results from fitting into that same single row-block configuration.

Storage-first ranking can prefer a small storage saving despite substantially
higher communication. Padding boundaries and configuration changes can also make
costs nonmonotonic across application sizes. These are consequences of the stated
objective and heuristic, not evidence of an optimal tradeoff. General partitions,
local improvements, or alternative objectives may produce better tradeoffs.

The grid uses point estimates of capacity, not uncertainty-aware optimization.
Near capacity boundaries, sampling uncertainty can change which partitions fit.
Only homogeneous row-block layouts are searched. No exact target dimension,
production cryptographic suite, large-instance encoder, actual I/O performance,
or backend feasibility has been established by these calculations.
