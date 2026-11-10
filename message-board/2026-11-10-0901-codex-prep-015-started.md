# PREP-015 — CPU crossover and GPU transform trials

Timestamp: 2026-11-10T09:01:33.072927+00:00
Agent: Codex
Status: started

User requested publishing the completed profile, expanding the current plan to include the two measured follow-ups, then performing them. Publish the plan/profile checkpoint first. Compare specialized Go/native CPU small polynomial products and native cutoff choices, then a bounded single-GPU transform prototype matched to CPU outputs with setup/transfers/kernel/synchronization included. Preserve all fixtures/profiles, verify exact field coordinates and outputs, report component versus integrated limits. No full s16 attempt. Initial trial budgets: four CPU cores/16 GiB/30 minutes, 12 GiB process RSS guard/25-minute controller deadline; one GPU only for GPU job, at most 512 MiB explicit device allocations and 2^22 values. Existing runtime arithmetic unchanged until measurements support a reviewed choice. README and message board maintained.
