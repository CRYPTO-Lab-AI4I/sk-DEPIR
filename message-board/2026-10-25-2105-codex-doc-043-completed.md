# DOC-043: Critical review findings

- Timestamp: 2026-10-25T21:05:28.033805+00:00
- Agent: Codex
- Status: completed

Reviewed all four docs; specification files unchanged. Findings:

1. [Encoder contract](../docs/encoder-contract.md): the mathematical encoder still
   handles bit conversion and padding, blurring its boundary with the application
   encoder. Prefer prepared field-element input and code artifact at this boundary.
2. [Parameter planning](../docs/parameter-planning.md): “supported configurations”
   includes provisional empirical entries. Distinguish candidates from encodable
   entries without designing a new catalogue API. Rate evidence wording also
   conflates exact calculated storage ratios with measured execution results.
3. [Data representation](../docs/data-representation.md): the one-field-element
   PIR record model is introduced after grouping and encoding. Introduce that
   basic fact in the opening database definitions; keep detailed arithmetic later.
4. Data representation's Query section repeats [protocol I/O](../docs/protocol-io.md).
   Keep layout/index/reassembly semantics in the former and execution rules in
   the latter, with a link. Keep mask separation in data representation.
5. Encoder open decisions list monomial algorithms and information-set construction
   as unspecified despite citing the paper's constructions. Distinguish selecting
   and implementing a method from missing mathematical specification.
6. Planner runtime/resource accounting does not explain reuse of code preprocessing
   across databases with the same code; separate per-code setup from per-database
   encoding when using measurements to check constraints.

Minor issues: long/fragmented lines, missing blank line after the grouping list,
repeated explanations of mask separation, and implementation notes interrupting
some conceptual sections. Prefer targeted edits over a wholesale reorganization.

README checked and consistent with current planned/implemented distinction.
Review was read-only except board messages/index; no runtime tests or experiments.
No new security or optimality claims were evaluated or introduced.
