# Agent instructions

## Project goal

Turn this academic implementation of the secret-key PIR protocols in
[`paper.pdf`](paper.pdf) into an open-source library that developers can
understand and use. Prioritize clear APIs, readable code, thorough documentation,
and changes that are easy to review and audit. Preserve the distinction between
the paper's constructions, the implemented behavior, and experimental shortcuts.

The current target is **TRL 5**, as defined by the ordered milestones and
acceptance criteria in [ROADMAP.md](ROADMAP.md). Implement the client and
preprocessing phases. Storage only returns requested records through an I/O
interface; do not build server-side protocol computation. Optimizing upload
through a storage layout that places same-position values from a query-sharing
group together, and a client adapter that reads them with one address, is in scope.
Keep this goal distinct from the currently implemented repeated-upload model.
TRL 6 acceptance remains deferred. Encoding efficiency improvements may be
investigated within the current functional scope. Document current limitations
and do not describe the target as already achieved.

**Prioritize fast encoding.** No programming language, compiler, algorithm,
hardware backend, data layout, memory strategy or parallelization scheme is
settled. Evaluate alternatives whenever they could improve efficiency, including
native-code and CPU/GPU approaches; do not restrict optimization to the current
Go implementation or an earlier step ordering. Preserve correctness, the agreed
PIR functionality and security requirements. Choose using reproducible kernel
and end-to-end measurements, with reusable setup, memory and I/O accounted for.
Keep the library usable on standard setups and implementations reviewable;
report deployment costs rather than assuming the current architecture is best.

**Make memory–time trade-offs explicit.** The current next step is memory
optimization toward s17, as recorded in [the scaling plan](docs/encoding-scaling.md#current-step-memory-optimization-for-s17).
Before adopting a material trade-off, present host/device memory savings, extra
computation/transfers/I/O, storage requirements and setup/encoding latency;
distinguish measured costs, projections and unknowns so the user can choose.
Do not silently exchange lower memory for longer runtime or larger host/disk use.
Retain the faster validated path as the default whenever it fits the admitted
host/device memory budgets, with runtime and allocation headroom accounted for.
A slower memory-saving mode is justified by insufficient memory for that path.
When sufficient memory is available, selecting a slower mode must be optional
and require an explicit command-line option; do not enable it by default.
Document the selected mode, memory constraint and measured or estimated runtime
penalty. Unknown runtime cost is not evidence of a cost-free optimization.

**Retain CPU and GPU client encoding paths.** GPU encoding is optional and must
not replace the CPU path or make CUDA a dependency of ordinary builds. A client
may have no GPU even when code preprocessing is delegated elsewhere. Both
backends consume the same verified code data and canonical representation; keep
backend selection explicit and code parameters independent of backend choice.

**Target practical commodity hardware.** Benchmark on hardware actually available
on this machine and respect the agreed experiment budgets. Target commodity CPU
and GPU architectures; larger workloads may require more RAM, cores or GPU memory,
but keep hardware/software requirements as generic as practical. Avoid making
specialized hardware or unnecessary vendor-specific dependencies prerequisites.
Evaluate architecture-specific acceleration by its measured benefit and deployment
cost, retaining a practical generic path. Clearly label projections for unavailable
hardware; do not present them as measurements.

## Current development environment and handoff

Local laptop development stops at the verified standalone s=15 checkpoint.
Larger-scale encoding work moves to a larger-memory environment; read the
[development handoff](docs/development-handoff.md) before continuing. Start with
a single shared-memory commodity CPU node (provisionally 256 GiB RAM), retaining
the C++ online path and Go preprocessing/references. s≥16 is a larger-memory
target, **not demonstrated feasibility or automatic HPC/GPU support**. Audit
s=16 memory/cardinality changes, validate the new host and transferred caches,
and review job budgets before a full run. Keep all laptop caches/profiles intact;
Git alone does not transfer large preprocessing data. TRL 5 and protocol scope
remain unchanged. Do not continue large laptop benchmarking by default.

## Rules for every contributor and agent

Follow the [data representation conventions](docs/data-representation.md) for
field representation, record layout, and mask separation, the
[encoder requirements](docs/encoder-contract.md) for
codeword encoding and cryptographic encoding, and the
[protocol I/O contract](docs/protocol-io.md) for
queries and responses. Keep agreed contracts distinct from implemented capabilities.

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
   Write documentation that is precise and explanatory without being lengthy.
   Introduce context before details, define terms consistently, and avoid repeating
   requirements. Keep brief alternatives in parentheses when that is sufficient.
   Apply editing feedback to similar passages throughout the document, not only
   the selected sentence. Remove obvious qualifiers and unnecessary terminology
   while preserving technical distinctions and security requirements.
   Name files, directories, and APIs concisely by their contents or responsibility.
   Include enough context to identify what they contain (for example, PIR
   configurations). Keep maturity and implementation status in metadata or
   documentation; avoid vague names such as `candidates` without an identified subject. When the same logic has implementations in multiple languages,
   label both existing and new implementation directories with their languages
   (for example, `liftrs-go` and `liftrs-cpp`) and update active references.
   Preserve historical measurement/source provenance when moving code.
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

7. **Retain preprocessing and profiles during optimization.** Unless there is a
   strong documented reason otherwise, enable CPU profiling and keep reusable
   code-preprocessing data for later measurements. Persist reasonably sized local
   caches with code specifications, source provenance and checksums so process
   exit does not discard expensive setup. Keep large caches outside Git; link
   their local location and regeneration command from the experiment report.
   Respect this project's machine memory/storage budget and report any cache
   omitted for size/resource reasons. Separate discovery, reusable preparation,
   cache save/load and database encoding, and label profiled timings separately
   from unprofiled latency baselines. Disable profiling only when optimization is
   complete; document any earlier exception. Do not silently delete useful caches.

8. **Hand off jobs longer than 30 seconds.** For every job expected to take more
   than 30 seconds, submit it, report its identifier and result location, then
   yield to the user. Do not poll scheduler/process state, follow logs, wait for
   completion, or analyze results until the user pings after completion or
   explicitly asks for a check. This applies to validation and test jobs as well
   as experiments. Keep automatic resource guards, deadlines, profiling and
   checkpoint persistence active; those are part of job execution, not agent
   monitoring. Dependent work must wait for the user's completion ping.

## Repository map

- `pirconfig/`: JSON configuration catalogue, schema, and Go loading/validation.
- `planner/`: bounded tradeoff search, cost-grid DP, greedy partitioning, validation, and logical cost evaluation.
- `pir/`: Go client API, Lift-RS protocol, and RM encoder testing placeholder.
- `utils/`: finite-field arithmetic, curves, permutation, masking, and I/O.
- `encoding/`: `liftrs-shared/` common API/preprocessing/orchestration, Go-only
  kernels in `liftrs-go/`, native kernels/binding in `liftrs-cpp/`, archived RM in
  `rme_legacy/`, and `cryptoref/` / `appref/` correctness references.
- `server/`: Julia HTTP query handler.
- `input/` and `output/`: example CSV data and generated matrix locations.
- `message-board/`: one file per agent message, with `README.md` as its index.
- `cmd/pir-rate/`: standalone Lift-RS monomial-sampling rate estimator.
- `experiments/`: reproducible experiment scripts, reports, and results.

The README describes the current implementation and its limits. Read the paper
and relevant code before changing protocol behavior or making performance or
security claims.
