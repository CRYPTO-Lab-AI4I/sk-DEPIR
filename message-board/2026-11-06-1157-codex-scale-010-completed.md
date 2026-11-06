# SCALE-010 completed: worker paused for sleep break

Timestamp: 2026-11-06T12:00:28.346766+00:00. Agent: Codex.

At the user's request, worker 84546 and active guard 89916 are confirmed stopped
(`T`, zero CPU); original parent 84443 remains stopped for final wait4 collection.
The old sleep-prevention child is terminated, awaiting parent reap. No computation
was restarted, resumed or discarded. [Pause metadata](../experiments/parallel-s15/manual-pause/status.json)
captures active time 10,579 s and the exact resume command. Existing caches/profiles
remain intact; unfinished preparation remains in RAM, not a restartable artifact.
Normal sleep can preserve this state; shutdown/restart or process termination cannot.

[extend_guard.py](../experiments/parallel-s15/extend_guard.py) now supports explicit
`--resume-from`. It verifies paused process identities, excludes the entire break
by starting from captured active duration, restores the same worker and a fresh
guard, and records the resumed monitor path. The old monitor chain remains stopped
until worker exit and then resumes in order for original statistics/follow-up.
The previous enforcing script is preserved by its content-addressed snapshot.
Do not directly SIGCONT the old guard, which would charge awake break time.
Do not resume until the user asks; this manual pause must not auto-resume on AC.

Three [policy tests](../experiments/parallel-s15/test_extend_guard.py) pass, including
manual-pause active time across different Python clock origins and long breaks.
Python compilation and diff whitespace pass. Resume itself was deliberately not
executed on the paused worker. Memory/cache limits remain unchanged; 9,221 seconds
of active budget remain. Profiled preparation spans the break and must be labeled
interrupted, rather than compared as uninterrupted latency.

README checked: no capability edit needed. Updated [experiment report](../experiments/parallel-s15/report.md#manual-sleep-break)
and [scaling plan](../docs/encoding-scaling.md). Next: await explicit resume, run
the recorded command under monitoring permissions, verify the worker runs, then
let the existing supervisor complete analysis/fresh cached measurement.
