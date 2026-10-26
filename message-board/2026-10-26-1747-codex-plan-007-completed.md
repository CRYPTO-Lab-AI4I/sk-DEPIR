# PLAN-007 — Logical width search and bounded full downloads

- Timestamp: 2026-10-26 17:47:40 UTC
- Agent: Codex
- Status: completed

## Changes and decisions

Added [application planning](../planner/application.go) around the existing greedy
kernel: split each application record into fixed-width fragments, stack them in
record-major order, search widths, and retain an arithmetic mapping back to source
bits. The original `Greedy` API retains its unsliced behavior. Default search covers
the original width, catalogue field widths, and powers of two starting at 8 bits;
callers may provide alternatives. Storage ranks first, total traffic second.

Updated [evaluation](../planner/evaluate.go), [cost helpers](../planner/costs.go), and
[greedy completion](../planner/greedy.go) for identity-storage/full-download pieces.
Indexed traffic repeats for every fragment; full pieces are downloaded once per
original application lookup. Every database participates independently of relevance.
The rate numerator excludes fragment padding. Samples count protocol queries, so
fragmentation consumes the existing query allowance faster.

The user clarified that full downloads should remain comparable with encoded-part
communication. `AllowFullDownload` opts in; `FullDownloadRatio` defaults to **1**, using
**download** as the comparison metric. Full pieces therefore add at most the encoded
parts' download, for at most twice that download overall. A full-only plan compares
with that width's indexed-only largest-fitting baseline instead. The optional
`MaxFullDownloadBits` adds an absolute cap; ordinary budgets always apply. Invalid
ratios are rejected. This replaces the initially discussed unrestricted policy.
The heuristic considers one terminal full-download suffix, not arbitrary packing.

## Documentation and evidence

Updated the four design contracts: [representation](../docs/data-representation.md),
[planning](../docs/parameter-planning.md), [encoding](../docs/encoder-contract.md), and
[I/O](../docs/protocol-io.md). Updated the [planner guide](../planner/README.md),
[examples](../planner/example_test.go), [catalogue guide](../pirconfig/README.md),
[roadmap](../ROADMAP.md), and [agent repository map](../AGENTS.md).
Read and updated the root [README](../README.md) for the new planning API and limits.
The [existing cost-grid report](../experiments/greedy-costs/report.md) now identifies
its unsliced baseline and published source commit `8b322b8`; historical results and
methodology were not changed. No new cost experiment was run or claimed.

## Verification and limitations

- `go test ./planner -count=1`: passed, including [new tests](../planner/application_test.go)
  for storage invariance under record reshaping, source mapping/padding, actual small
  reference recovery, mixed-piece accounting, ratio limits, budgets, and overflow.
- `go test -race ./planner -count=1`: passed.
- `go vet ./planner`: passed.
- `go test ./... -count=1`: all packages passed in an isolated temporary copy, because
  legacy RM tests write example CSVs. Compiler: `/tmp/sk-depir-go-toolchain/go/bin/go`;
  cache: `/tmp/sk-depir-go-cache`. Final focused tests passed after editorial cleanup.
- `git diff --check`: passed. Changed Markdown relative file targets resolve.

These APIs plan layouts; they do not add an application runtime executor, a new
cryptographic suite, or large-instance encoding support. General supplied partitions
remain valid, but heuristic selection is not optimal. Estimated capacities remain
provisional. Relative caps govern selection, not validity of externally supplied plans.
The earlier checkpoint was pushed as `8b322b8` before this work; these new changes
remain local for review.
