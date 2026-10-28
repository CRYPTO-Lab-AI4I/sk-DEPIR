# Cyclic encoding at s=13

## Purpose and reproduction

Measure the next catalogue configuration, s=13,m=2,t=5, using the unchanged
fixed-constant cyclic backend from revision `7eff0ff`. The user clarified that
one hour was a ballpark and authorized a **two-hour whole-case timeout**; the
12 GiB observed Go-memory guard and 9 GiB soft GC target remain.
[policy.json](policy.json) records this allowance. Earlier forecast policies and
reports are preserved unchanged.

Run from the repository root with compatible Go, Python 3 and POSIX `wait4`:

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/cyclic-s13/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
python3 experiments/cyclic-s13/summarize.py
python3 experiments/cyclic-s13/record-catalogue.py
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/cyclic-s13/verify.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
```

[run.py](run.py) builds the existing [scaling driver](../cyclic-scaling/main.go).
[summarize.py](summarize.py) checks provenance and generates [table.md](table.md).
[record-catalogue.py](record-catalogue.py) records standalone encoder support only
when completed encoding and verification evidence exists; it retains earlier
sampling evidence. [verify.py](verify.py) checks catalogue/planner behavior and
runs the full Go suite in an isolated copy to protect CSV fixtures.

[results.json](results.json) records revision, exact source/policy hashes,
command, environment, phase observations, output hashes and process peak RSS.
[events.jsonl](events.jsonl), [stderr.txt](stderr.txt) and
[progress.jsonl](progress.jsonl) retain the raw execution data. Progress logs
record elapsed time each minute. The [evidence manifest](evidence-manifest.json)
also records hashes of retained scripts, controls, verification logs and the
updated catalogue/tests. An [initial monitor attempt](initial-monitor-failure/README.md)
failed immediately because the sandbox denied optional `ps` sampling; it produced
no phase measurements and is excluded. The retained runner does not require `ps`
and kills its child if monitoring exits unexpectedly.

## Environment and measurement boundaries

Apple M5 Pro, 18 CPU cores and 48 GB RAM; macOS 26.6.2 arm64, Go 1.27.3,
Python 3.9.6, default GOMAXPROCS, portable Go without native acceleration or GPU.
This is the same host as the [preceding comparison](../field-multiplication/report.md)
and its [hardware metadata](../encoder-profile-s10/results.json). Temporary packed
output uses local APFS storage; sync is included and readback is cached. It is
not a sustained storage throughput measurement. No tests or other benchmarks ran
alongside the retained timing run.

Code-only discovery performs exact monomial enumeration and support discovery,
using seed 1013 (one support sample). Reusable preparation constructs the cyclic
polynomials and generator spectrum. Each of three encoding calls uses the same
uniform field-element message, seed 71. GC before phases is outside their timers;
the driver's 100 ms Go-memory sampling/guard remains inside. Prepared constants
stay in memory. No large-cache JSON serialization/loading, application partition
execution, cryptographic masking or larger paper-encoder comparison is measured.

After the last encoding, the driver checks all systematic symbols, samples the
public coordinate mapper, and sums 16 actual degree-t curves using independent
carryless arithmetic. It hashes the complete canonical uint32 output and streams
packed output through bounded buffers, syncs it and verifies cached readback.
These checks do not exhaustively test every possible curve or message. Dense
oracle/cross-encoder checks remain covered by the existing smaller fixtures.

First use is discovery + reusable preparation + one median encoding, excluding
validation/I/O. Whole-case wall time additionally includes all three encoding
calls, validation/output and phase overhead. Sampled heap includes live cache,
message and output; it is not an exact scratch-memory bound. Go-runtime `Sys`
can retain allocations from preceding phases. Process peak RSS is a separate
whole-case observation, not isolated phase RSS. The memory guard observes Go
runtime memory, not an OS-enforced address-space limit.

## Measured results

Exact K=**11,926,435**, N=**67,108,864**, polynomial **0x201b**. Code rate K/N is
**0.17771773** and storage overhead N/K is **5.6269**. The historical Monte Carlo
capacity was K=11,955,914; the exact value is 0.25% lower and lies within the
stored sampling confidence interval. Its original evidence is retained.

| Stage | Time | Sampled peak heap | Sampled peak Go runtime |
|---|---:|---:|---:|
| Discovery | 85.95 s | 0.834 GiB | 0.984 GiB |
| Reusable preparation | 1,765.37 s (29.42 min) | 4.711 GiB | 5.342 GiB |
| Encoding 1 | 185.15 s | 1.181 GiB | 5.342 GiB |
| Encoding 2 | 185.01 s | 1.181 GiB | 5.342 GiB |
| Encoding 3 | 185.20 s | 1.181 GiB | 5.342 GiB |

Median encoding is **185.15 s**. First use is **33.94 min**; whole-case wall time
is **40.15 min**, with **5.337 GiB** peak process RSS. Neither timeout nor memory
guard fired. Packed output is **109,051,904 bytes (104 MiB)**. Packed write/sync
took 0.339 s; final systematic/curve verification took 0.576 s. Hashes and cumulative
allocations are in the raw results; cumulative allocated bytes are not peak RAM.

The [preceding forecast](../field-multiplication/s13-forecast.json) predicted about
52 minutes first use without the safety factor, or 78 minutes and 8.42 GiB with
its 1.5 factor. Actual first use was lower, mainly because preparation took
29.42 minutes rather than the point forecast's 48.04 minutes. Encoding was slower
than its 147.80 s point forecast. This is evidence that the heuristic's different
stage predictions can err in different directions, not a validated extrapolation
to s>=14. The completed run fits even the original one-hour ballpark; the explicit
two-hour allowance permitted measuring it without relying on forecast admission.

## Verification and current scope

All 11,926,435 systematic symbols, sampled public-coordinate checks, 16 independent
actual-curve sums and complete packed sync/readback checks passed for the final
output. Canonical SHA-256:
`6120f0388f6035dd480e391b26ef93c04ed368a77f2f93c02f6d4f56b99a7b79`.
Packed SHA-256:
`0b545a19ff25c1cdb26478ed2786c6ad76052f46c8a258bf11ad763c8f276980`.

The [catalogue](../../pirconfig/configs_catalogue.json) now records this standalone
in-memory cyclic support and exact capacity. The
[evidence test](../../pirconfig/s13_evidence_test.go) ties that promotion to a
completed run, dimensions and output checks. Catalogue/planner tests, vet and the
isolated full suite all passed; commands and outcomes are recorded in [verification.json](verification.json) and
[verification.log](verification.log). Source/event hashes and documentation links
are checked before completion. The mathematical backend is unchanged from the
previously tested checkpoint; no redundant race run is needed for result/catalogue
and documentation updates.

This reaches the current adapter's s<=13 bound. It does not establish large-cache
loading, application-plan execution, GPU/external-memory support, production
security or completed TRL 5 acceptance. Historical cost-grid snapshots keep their
original catalogue inputs; this experiment does not rerun their optimization.
