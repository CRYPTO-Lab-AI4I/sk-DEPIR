# Experiment conventions

Place each new experiment in `experiments/<descriptive-name>/`. Keep unrelated
measurements in separate directories. Commit small, useful raw results; for large
artifacts, link to a durable location and record checksums in the report.

Each experiment must contain:

- A runnable script such as `run.sh` or `run.py` with all parameters explicit.
- `report.md` linking to that exact script and its results (local relative links
  or durable external links).
- Results in the same directory or a durable link recorded in `report.md`.

The report must state the purpose, repository revision, prerequisites, commands,
input data and parameters, random seeds if applicable, hardware/software
environment, timing boundaries, units, results, and known limitations. Explain
how to reproduce and interpret each figure or table. Separate actual measurements
from extrapolations and values copied from the paper. When an experiment changes,
update its script, results, and report together, then post an update on the
[agent message board](../message-board/README.md) and check the README.

Do not put unreported measurements into the README. Link each new performance
claim there to its experiment report.

## Recorded experiments

- [Bounded-sample parameters](bounded-sample-parameters/report.md): analytical
  security-profile selection, not rate measurements.
- [Lift-RS rates](liftrs-rates/report.md): uniform monomial sampling for all 13 PIR
  configurations, with reproducible counts and confidence intervals.

- [Greedy planner costs](greedy-costs/report.md): logical storage and per-record
  traffic for a 21-case application/record-size grid, with complete compact plans.

- [Application width-search costs](application-costs/report.md): updated 21-case grid,
  split/stack plans, full-download policy results, and indexed-only alternatives.

- [Bounded tradeoff costs](tradeoff-costs/report.md): nondominated alternatives for
  the 21-case grid, with communication rate and explicit approximation flags.

Default application-cost grids require at least **5,000 records**; the retained grid
has 21 cases. Code-parameter experiments and correctness fixtures are separate.

- [Cost-grid DP](dp-costs/report.md): separate maximum-coverage DP on rounded
  storage/communication coordinates, with explicit cost windows and the 21-case table.

- [Paper/cyclic encoder comparison](encoder-comparison/report.md): exact bounded
  encodings with reusable public preprocessing artifacts; separate setup, load,
  encoding, process memory and output I/O measurements.

- [Download-objective comparison](download-objective-comparison/report.md): rerun
  all four cost grids with storage/download ranking; preserve prior results and
  compare endpoints, displayed options and approximate intermediate frontiers.

- [Table-free encoding checkpoint](tablefree-encoding/report.md): portable
  arithmetic through s=20, preserved old codewords, and exact s=9 encoding with
  separate discovery/preparation processes, prepared artifacts and packed output.

- [Encoder CPU profiles](encoder-profile/report.md): separate s=9 paper/cyclic
  preparation and database-encoding profiles, with checkpoint output verification.

- [Fast polynomial products](fast-polynomial/report.md): independent convolution
  checks, kernel timings and same-machine old/new preprocessing and encoder costs.

- [CPU encoding at s=10](encoder-scaling/report.md): exact capacity, reusable JSON
  artifacts, paper/cyclic agreement, independent curve checks and separate heap/output measurements.

- [Current s=10 cyclic CPU profiles](encoder-profile-s10/report.md): separate reusable
  preparation and database encoding, with stage/source attribution and field-kernel priorities.

- [Progressive cyclic scaling](cyclic-scaling/report.md): measured s=11/12 with
  an explicit first-use/RAM policy and labelled forecasts for larger catalogue fields.

- [Portable field multiplication](field-multiplication/report.md): general-kernel
  comparison and bounded reusable FFT constant tables, with full cyclic scaling.
