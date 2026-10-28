# REVIEW-001 — Plan published; claim audit completed

- Timestamp: 2026-10-28T09:24:26Z
- Agent: Codex
- Status: completed

## Publication and audit scope

Published the five-step plan as `faf8085` to origin/main. Reviewed all five docs,
README/roadmap, orbit implementation and tests, catalogue status, and encoder,
rate, DP and tradeoff reports/results. No implementation or reported measurements
were changed during this audit. README reviewed; findings below include status
wording to reconcile later.

## Findings, in priority order

1. **The complete encoder is not yet quasi-linear.**
   [PrepareCyclicFFT's comment](../encoding/liftrs-go/orbit.go) says O(N log² N) work,
   but only the large G*H product uses that transform. H=(P*M) mod X^K still calls
   Karatsuba, as does the paper solver's product-tree arithmetic. The package
   README discloses Karatsuba, but the five-step plan does not explicitly schedule
   replacing this remaining multiplication bottleneck. Narrow the comment to the
   transformed product, give the full cost O(M(K)+N log² N), and add fast/truncated
   multiplication and discovery/setup complexity checkpoints to the scaling plan.
   This also invalidates treating prior N*s² chat projections as complete runtimes.

2. **Table metric descriptions conflict within reports.**
   [DP report](../experiments/dp-costs/report.md), Table conventions, and
   [tradeoff report](../experiments/tradeoff-costs/report.md), Methodology, still say
   options.md displays (U+D)/B, while current generators/tables display D/B. Later
   appended sections supersede them but leave conflicting active instructions.
   Revise the original descriptions in place; retain historical total-communication
   optimization/comparisons clearly labelled as such. Numerical tables themselves
   match the current display convention.

3. **Verified s=8 evidence has not reached catalogue/status descriptions.**
   [Comparison results](../experiments/encoder-comparison/results.json) contain exact
   K=3121 for s=8,m=2,t=8, while the catalogue retains estimated K=3117 and candidate
   status. The estimate is legitimate historical evidence; it should not silently
   change reproducible planning snapshots. Add exact capacity as separate current
   evidence and distinguish standalone bounded encoding from integrated execution.
   [Catalogue README](../pirconfig/README.md) attributes all unsupported statuses
   solely to unverified dimensions, which is stale for this one configuration.
   The difference is small (four symbols), not an invalidation of the cost grids.

4. **Older status paragraphs omit newer capabilities.**
   [Encoder contract](../docs/encoder-contract.md), Small reference encoder, ends
   with larger fields remaining unimplemented, and its final research paragraph
   says the options do not change supported reference sizes. Later orbit paragraphs
   and README correctly say s=2..8,N<=65536. [ROADMAP Current starting point](../ROADMAP.md)
   also treats automatic planning as separate future work despite implemented
   planners. Qualify those older statements by backend/integration layer and retain
   a single coherent current-state summary.

5. **Plan acceptance needs stage-specific resource evidence.**
   The report accurately calls memory a full-process RSS high-water mark. However,
   all three preparations share one child process; encoding RSS also includes
   load/output buffers, and file I/O only writes an unpacked uint32 result.
   Existing results do not provide separate peak memory per preparation method,
   encoding-only peak workspace, or packed/grouped pipeline I/O. These are not
   falsely claimed in the detailed report, but should be explicit prerequisites
   in Steps 3-4, rather than interpreting existing measurements as those checks.
   CPU metadata is only architecture/logical CPU count, not an exact chip model;
   filesystem/storage details also remain unspecified.

6. **Scope and publication validation gates need explicit ownership.**
   The plan defers external-memory/GPU stages to roadmap TRL gates, while the
   roadmap explicitly defers such optimization to TRL 6. This is not a claim they
   exist, but the decision point should be explicit before implementing Steps 4-5.
   Likewise the plan allows public artifact publication but loaders deliberately
   trust mathematical content; specify the independent artifact verifier and
   provenance check required before publication, rather than relying on structural
   loader checks. This is a planned acceptance gap, not a loader security claim.

## What matches and verification

- Encoder source and all public artifact SHA256 hashes match saved results.
- Recomputed medians match table.md, including 12.91 ms paper and 8.90 ms cached
  FFT at N=65536. Output hashes match across all methods for all five cases.
- Code genuinely separates discovery, method preparation, loading and message
  encoding; fixed-generator spectra and solver reciprocals are reusable artifacts.
- CPU-only, bounded in-memory support, absent streaming/GPU integration, unpacked
  I/O boundaries, three-repeat timing limitations and trusted artifacts are disclosed.
- Grouped reads remain explicitly a goal; evaluator still counts repeated upload.
- Relative file links in docs resolve. Historical DP/tradeoff source hashes differ
  only for run.py, consistent with the documented later render-only changes.

Read-only evidence audit and arithmetic/table checks performed; no fresh benchmark
campaign or full test rerun was needed. `git diff --check` passed for review records.
Only the message-board completion record/index are new after the published plan;
publish those as a separate review receipt to preserve a clean checkpoint.
