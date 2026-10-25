# DOC-035: Replacement paper reconciled

- Timestamp: 2026-10-25T20:04:20.991981+00:00
- Agent: Codex
- Status: completed

The user replaced the old PDFs with [paper.pdf](../paper.pdf). Its Table 1
includes the bounded-sample setting missing from the former version. Updated
[parameter planning](../docs/parameter-planning.md) to use Conjecture 6.1 with
`m=2`, `s=8..20`, `2^50` samples and a 100-bit conjectured computational target.
Minimum degrees are 8, 7, 6, and 5 for `s=8`, `s=9`, `s=10..12`, and
`s=13..20`, respectively. Lower degrees remain future lower-security alternatives.
The equality case at `s=10` is explicit.

Linked the reproducible [script](../experiments/bounded-sample-parameters/run.py),
[results](../experiments/bounded-sample-parameters/results.csv), and
[report](../experiments/bounded-sample-parameters/report.md), including the paper
hash, base revision, environment, formulas, and limitations. Table 1's 712 GB
capacity now motivates two blocks exceeding 1 TB; paper estimates remain distinct
from actual encoding evidence. This supersedes DOC-034's obsolete-paper discussion.

Updated [encoder requirements](../docs/encoder-contract.md) to reference the initial
parameter targets; checked and updated [README](../README.md) and
[agent instructions](../AGENTS.md) for the new paper path. Historical board files
retain their original links and statements.

Verification: extracted and visually reviewed pages 38 and 64 (PDFKit rendering).
Ran the script successfully; checked all 13 output rows and the 100-bit bound;
checked relative file links in current docs, README, AGENTS, and the new report.
`git diff --check` passed. No runtime implementation changed and no encoding
benchmark was run. The new paper and prior user edits were preserved.
