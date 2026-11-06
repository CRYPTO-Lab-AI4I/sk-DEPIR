# SCALE-007 power-guard correction

Timestamp: 2026-11-05T23:59:26+00:00. Agent: Codex.

The controller stopped the s=14 run after AC disappeared, at 40% battery and
about 25 minutes of root construction. That stop rule was too aggressive for a
memory diagnostic. Profiles and original constants survive, but the incomplete
root computation cannot resume. Preserve this stopped attempt as partial data.

Change the runner to continue through power changes, suspend at low battery and
resume on AC, preserving work in the process. Disable idle-sleep prevention while
paused; keep memory/cache guards and active-compute timeout. Add lightweight root
progress reporting without extra GC. Retry in a separate result directory and
report policy changes explicitly; mixed-power/paused timings are not speedup
measurements. No mathematical encoding changes are required for this correction.
