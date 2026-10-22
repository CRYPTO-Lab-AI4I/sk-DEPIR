# Agent instructions

## Project goal

Turn this academic implementation of the secret-key PIR protocols in
[`sk_DEPIR.pdf`](sk_DEPIR.pdf) into an open-source library that developers can
understand and use. Prioritize clear APIs, readable code, thorough documentation,
and changes that are easy to review and audit. Preserve the distinction between
the paper's constructions, the implemented behavior, and experimental shortcuts.

The current target is **TRL 5**, as defined by the ordered milestones and
acceptance criteria in [ROADMAP.md](ROADMAP.md). Implement the client and
preprocessing phases. Storage only returns requested records through an I/O
interface; do not build server-side protocol computation. Defer TRL 6 scaling
and advanced optimization work until the user chooses to pursue it. Document
current limitations and do not describe the target as already achieved.

## Rules for every contributor and agent

1. **Track every change.** Read the shared
   [message board](message-board/README.md) for every work item. Create one
   timestamped Markdown file per message in `message-board/`: a `started`
   message before editing and a separate `completed` or `handoff` message
   afterward. Add each message to the board index. Record changed paths,
   reasons, verification, limitations, and links to relevant files, including
   any experiment report. Do not leave code, documentation, configuration,
   generated files, or results untracked. Preserve previous messages; correct
   an older message with a new file.
2. **Document code as you write it.** Every code change needs documentation
   appropriate to its effect: public API comments and usage examples for API
   changes, explanation of non-obvious protocol or arithmetic behavior beside the
   code, and updates to relevant design or user documentation. Keep comments
   accurate and explain invariants and security assumptions rather than restating
   syntax. Do not submit undocumented code.
3. **Make experiments reproducible.** Put each new experiment in its own
   `experiments/<descriptive-name>/` directory. Include a runnable script,
   a `report.md`, and results or a durable link to them. The report must link to
   the exact script and results, state the revision, parameters, environment,
   commands, methodology, units, and limitations, and distinguish measured data
   from estimates or paper values. Do not add results without their report and
   script link. See [experiment conventions](experiments/README.md).
4. **Check the README on every change.** Read the relevant sections of
   [README.md](README.md) and update them whenever the public API, supported
   parameters, setup, tests, performance claims, or project layout changes.
   Record the README check on the message board even when no edit is needed.
5. **Keep security and correctness claims narrow.** The current code is an
   academic prototype. Do not claim production security or full implementation
   of a paper construction without evidence. Keep the Lift-RS RM-encoding
   placeholder, field compatibility, randomness source, and benchmark exclusions
   visible until each is resolved and verified.
6. **Leave reviewable evidence.** Record commands and outcomes, including skipped
   checks and why. Link tests and benchmarks to their implementations. Avoid
   changing benchmark methodology silently; explain any change to measurement
   boundaries in the report and README.

## Repository map

- `pir/`: Go client API, Lift-RS protocol, and RM encoder testing placeholder.
- `utils/`: finite-field arithmetic, curves, permutation, masking, and I/O.
- `encoding/`: external RM encoders in C++ and Julia.
- `server/`: Julia HTTP query handler.
- `input/` and `output/`: example CSV data and generated matrix locations.
- `message-board/`: one file per agent message, with `README.md` as its index.
- `experiments/`: reproducible experiment scripts, reports, and results.

The README describes the current implementation and its limits. Read the paper
and relevant code before changing protocol behavior or making performance or
security claims.
