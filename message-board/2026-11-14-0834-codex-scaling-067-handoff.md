# Bounded validation retries submitted

- Timestamp: 2026-11-14T08:34:36.893629+00:00
- Agent: Codex
- Work: SCALING-067
- Status: handoff

Source checkpoint `2a2095e0b287a7751c9399d464285ea7a32aa02a` records s16 disk success, both failed-validation
forensic reconciliations and the corrected all-mode/dense-prefix test coverage.
Submitted two independent CPU-only jobs in parallel, without dependencies:

-74447: native-02 preprocessing suites/sanitizers plus large s16 disk artifact
rehash; [submission](../experiments/s17-preprocessing/retry-submission.json),
[report](../experiments/s17-preprocessing/report.md), results
`experiments/s17-preprocessing/runs/native-02/`.
-74448: modules-02 CPU regression/sanitizer/baseline timing validation;
[submission](../experiments/cpu-encoding-modules/retry-submission.json),
[report](../experiments/cpu-encoding-modules/report.md), results
`experiments/cpu-encoding-modules/runs/modules-02/`.

Each retains4cores/8GiB request,6GiB sampled RSS guard,4GiB Go soft limit,
2GiB retained cache/4GiB scratch,600s command/900s controller/1200s Slurm limits.
Existing large disk files are read for identity verification, not recreated;
new retained payload remains small. No encoding rerun or full s17 allocation.
No monitoring or dependent result analysis; await user completion ping.

README checked and updated in checkpoint. Python analyzer correct/mismatched
streaming-digest self-check passes; Go formatting, Python/Bash syntax and authored
diff checks pass. Git's staged whitespace check reports pre-existing whitespace
in immutable saved scheduler/compiler logs; preserved verbatim for recorded
hash identity, with no authored-code whitespace issue. Full new tests remain
pending. No push requested/performed and no timing/support claim beyond reports.
