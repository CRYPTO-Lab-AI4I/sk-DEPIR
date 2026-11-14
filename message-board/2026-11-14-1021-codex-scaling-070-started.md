# Approved two-GPU preprocessing validation

- Timestamp: 2026-11-14T10:21:03.236578+00:00
- Agent: Codex
- Work: SCALING-070
- Status: started

User explicitly approves two GPUs for bounded peer validation:20 minutes,
512MiB aggregate owned GPU allocation (256MiB/card),4CPU and8GiB host RAM.
Submit existing runner as peer-01 with typed `gpu:h200:2` GRES, preserving
600/900/1200s worker/controller/Slurm guards,6GiB sampled RSS,4GiB Go soft,
2GiB retained and4GiB scratch caps. CUDA contexts remain additional as agreed.
This is small correctness/peer/tile validation, not full s17 admission. Job74453
and all other unreviewed jobs stay unmonitored; no results inspected. README
checked: capabilities and public API unchanged, so no edit needed. Record ID,
results/cache locations and source revision, then yield without monitoring.
