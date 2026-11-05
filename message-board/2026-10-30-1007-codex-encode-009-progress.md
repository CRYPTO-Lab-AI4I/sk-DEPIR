# ENCODE-009 progress

Timestamp: 2026-10-30T10:07:00Z. Agent: Codex.

The complete C++ online path passes small oracle/hybrid comparisons, concurrent
workspaces, invalid-input checks, the independent native ABI sanitizer fixture,
and both isolated full repository suites. All three initial s=14 canonical hashes
match retained evidence. The first encoding took 86.6 seconds, but later repeats
have inconsistent clock accounting (native sequential stage sums exceed the
outer Go elapsed interval) and large timing spikes. These measurements are
preserved in [the experiment](../experiments/native-online/report.md), not silently
replaced or used for a speedup claim.

Repeat the cached comparison in fresh processes, preventing idle sleep only while
the benchmark runs (`caffeinate -i -w <benchmark-pid>`). Use two profiled repeats per
backend, and report timer consistency, all measured ranges and this methodology
change. Reusable s=14 constants remain unchanged; no discovery/preparation rerun.
