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
