# RESEARCH-001 — Systematic encoding algorithms reviewed

- Timestamp: 2026-10-27T08:32:04Z
- Agent: Codex
- Status: completed

## Findings and paths

Added [encoding-algorithms.md](../docs/encoding-algorithms.md), starting with paper
Section 4.2.4 and the Guo-Kopparty information set, then reviewing structured
Vandermonde algorithms, characteristic-two transforms, the August 2026 Dyadic
AFFT paper and public implementations. The note links primary sources and
separates kernel improvements from complete curve-lifted encoding.

Derived a cyclic polynomial formulation preserving the paper's information set:
base-field products can replace per-message extension-field solving/evaluation.
Also described generator recovery by gcds of uniformly random codewords, with
completion certified by the exact dimension. Both are proposals to validate,
not new implemented capabilities or benchmark results. Large vectors and artifacts
still require a memory/I/O plan. Explained coordinate-grid output to avoid a
retrieval-time discrete-log requirement.

Linked the review from [README.md](../README.md) and the
[encoder contract](../docs/encoder-contract.md). Earlier uncommitted work was retained.

## Evidence and verification

Inspected full relevant pages 25, 31-32 of the local paper with pypdf, reference
preprocessing in [reference.go](../encoding/liftrs/reference.go), and existing
[GF2n arithmetic](../utils/gf2n.go). Reviewed primary papers and implementation
sources through web browsing. Local paper SHA256:
`205571fe7afb594d82d461a900bf0a5e2e210b8c688260224dbc6a5d9f0b8076`.
Base revision: `42e9180`; working tree already contained documentation changes.

Relative links in the research note resolve; `git diff --check` passed. README
checked and updated. No encoder code, dependencies or experimental data changed;
no benchmarks or Go tests were run. The cyclic derivation has not yet been tested
against the oracle. The new paper's linked dyadic subdirectory could not be fetched;
its code availability and CPU timings are author reports, not independently verified
build, licensing or performance evidence. Shell DNS was unavailable, while web
browsing supplied the cited sources. No claim of globally optimal encoding is made.
