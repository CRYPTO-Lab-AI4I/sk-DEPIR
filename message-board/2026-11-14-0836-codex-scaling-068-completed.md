# Publication checkpoint and s17 readiness

- Timestamp: 2026-11-14T08:36:28.623841+00:00
- Agent: Codex
- Work: SCALING-068
- Status: completed (local publication checkpoint; push verified separately)

Existing source, failed-validation evidence, corrected CPU fixtures, measured
s16 disk results, profiles and retry submission metadata are already committed
through be85df4. Add only this publication/readiness record and push main as
requested. Unreviewed job74447/74448 outputs remain local and uninspected; other
unreviewed historical table artifacts remain outside this publication scope.

Full s17 preprocessing is **not ready for launch**. Native shared source supports
s17 and targeted bounded tests passed, but complete corrected regression/
sanitizer validation is pending review of job74447 (CPU refactor timing validation
74448 also pending). No full-domain native calibration or measured peak exists.
The fast native product arrays alone384GiB plus Go arrays/GC/runtime make a host
peak around600GiB or higher plausible, not bounded. Capacity of an observed
~907GiB node does not establish headroom or authorize allocation. Need a guarded
full-size experiment runner, persistent artifacts/checkpoints, explicit aggregate
host/storage/deadline budgets and a defensible runtime estimate before submission.
Current artifacts project hundreds of GiB retained storage. No numeric budget
increase is inferred. CUDA preprocessing remains s≤16; wider/distributed GPU17
execution is unimplemented. Multiple GPUs provide capacity only, not support.

[Guide](../docs/preprocessing.md), [reference](../docs/encoding-status.md) and
[validation report](../experiments/s17-preprocessing/report.md) already state
these limits. README checked: no API, measurements or capability change in this
publication step, so no edit needed. Verification is read-only Git/document
review and authored-message whitespace check; no code or new performance test.
Next wait for validation completion ping, reconcile results, then calibrate/admit
a full native run or implement the wider GPU server route; no monitoring now.
