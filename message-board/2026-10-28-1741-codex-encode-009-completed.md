# s=13 cyclic benchmark completed and documented

- Timestamp: 2026-10-28 17:41 UTC
- Work item: ENCODE-009
- Status: completed

## Policy and results

The user clarified that one hour is a ballpark, not a hard requirement, and
allowed a two-hour case timeout. The [policy](../experiments/cyclic-s13/policy.json)
retains the 12 GiB observed Go-memory guard and 9 GiB soft GC target. Historical
one-hour policies and reports are preserved.

The [runner](../experiments/cyclic-s13/run.py) uses the unchanged mathematical
backend and scaling driver from `7eff0ff`, m=2,s=13,t=5. The
[report](../experiments/cyclic-s13/report.md),
[results](../experiments/cyclic-s13/results.json),
[raw events](../experiments/cyclic-s13/events.jsonl),
[progress](../experiments/cyclic-s13/progress.jsonl),
[table](../experiments/cyclic-s13/table.md),
[summary script](../experiments/cyclic-s13/summarize.py) and
[evidence manifest](../experiments/cyclic-s13/evidence-manifest.json)
record source/host, measurement boundaries, checks and limitations.

- Exact K=11926435,N=67108864, polynomial 0x201b; rate 0.17771773,
  storage overhead N/K=5.6269.
- Discovery 85.95 s (one support sample), reusable preparation 1765.37 s
  (29.42 min), encoding calls 185.15/185.01/185.20 s; median 185.15 s.
- First use 33.94 min; whole case including three encodings and verification/I/O
  40.15 min. Peak process RSS 5.337 GiB; no memory/timeout guard fired.
- All 11926435 systematic symbols, sampled public coordinate mappings and
  16 independently evaluated actual curves passed for the final output.
  Canonical/packed SHA-256 and sync/readback checks passed; packed output 104 MiB.
- The previous 78-minute conservative first-use forecast included a 1.5 factor;
  even its 52-minute point estimate exceeded the observed first use. Preparation
  was faster and encoding slower than their separate point forecasts. Do not
  treat that model as a validated extrapolation to s>=14.

An [initial optional monitor](../experiments/cyclic-s13/initial-monitor-failure/README.md)
failed because sandbox permissions denied `ps`, before any phase completed.
Its raw files and exact failing harness source are archived. The retained runner
uses elapsed-time progress and final wait4 RSS; it also cleans up its child on
monitor failure. The failed attempt is excluded from timing comparisons.

## Catalogue, documentation and verification

[record-catalogue.py](../experiments/cyclic-s13/record-catalogue.py) promotes only
completed evidence. The [catalogue](../pirconfig/configs_catalogue.json) now records
exact standalone in-memory support for s=13, retaining historical Monte Carlo
K=11955914. Updated [catalogue tests](../pirconfig/catalogue_test.go) and added
[the evidence check](../pirconfig/s13_evidence_test.go). Supported measured catalogue
configurations now span s=8..13; this is not application-plan execution or production
security. No large-cache loading, larger paper comparison or s>=14 measurement
was added, and JSON loader caps remain unchanged.

README was checked and updated. Also updated the [roadmap](../ROADMAP.md),
[encoder contract](../docs/encoder-contract.md),
[algorithm plan](../docs/encoding-algorithms.md),
[parameter planning](../docs/parameter-planning.md),
[package README](../encoding/liftrs-go/README.md),
[catalogue README](../pirconfig/README.md) and
[experiment index](../experiments/README.md). Earlier cost-grid snapshots keep
their original inputs and are not rerun by this work.

[Verification script](../experiments/cyclic-s13/verify.py),
[commands/outcomes](../experiments/cyclic-s13/verification.json) and
[raw log](../experiments/cyclic-s13/verification.log): catalogue/planner tests,
vet and isolated-copy full `go test ./...` all passed, after timing completed.
No race rerun was needed: mathematical/concurrent code is unchanged from the
previous passing checkpoint. Measured source/event hashes, archived failing
harness hash, summary reproduction, documentation targets and `git diff --check`
pass. Changes are uncommitted; no commit/push was requested for this work item.
