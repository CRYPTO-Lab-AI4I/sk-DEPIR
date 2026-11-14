# Prioritize wide GPU preprocessing and clarify scaling tables

- Timestamp: 2026-11-14T08:44:29.855564+00:00
- Agent: Codex
- Work: SCALING-069
- Status: started

User corrects server preprocessing priority: implement the GPU s17 route, reusing
shared flow and keeping CPU as reference; do not prioritize another full CPU
preparation run. First widen reusable CUDA arithmetic/transform operations and
support a same-node multi-GPU resident preparation engine, then bounded
correctness/resource/communication validation before full-scale admission.
Preserve existing CPU/CUDA client paths and ordinary CUDA-free builds. No silent
slower memory mode or increased numeric experiment budget. Read README, board,
preprocessing and scaling docs, CUDA/native transforms and contracts.

Also rewrite every reference table cell with explicit status, runtime units and
named memory measures; split wide tables where useful. Preserve all recorded
numbers, laptop/HPC provenance and measured/projected/unknown distinctions.
The s16 disk cell must identify19.469GiB sampled process RSS,32GiB work files,
additional file cache and10h12m40s encoding, without interpreting sums as physical
RAM. README check and relevant docs accompany source changes. Jobs74447/74448
and their generated analysis remain uninspected until the user's completion ping.
