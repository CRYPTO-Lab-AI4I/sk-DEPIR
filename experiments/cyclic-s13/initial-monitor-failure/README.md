# Initial monitor failure

The first runner exited immediately because the sandbox denied launching `ps`
for optional progress RSS sampling (`PermissionError: Operation not permitted`).
No numerical phase/event completed. The child was absent in a follow-up process
inventory. These files preserve the initial state; the `running` status in that
initial results file is superseded by this explanation. The retained runner uses
elapsed-time progress only, retaining child phase heap sampling and final wait4
process RSS. This failed harness attempt is excluded from timing comparisons.

[run.py](run.py) archives the exact failing harness source; its hash matches the
initial manifest's `experiments/cyclic-s13/run.py`. Its relative paths refer to
that original location, not this archive directory. Use the parent runner for
reproduction, not the archived monitor.
