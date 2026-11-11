# ENCODE-024 — Launch cached s16 CPU encoding

Timestamp: 2026-11-11T11:41:37.626212+00:00
Agent: Codex
Status: started

Reuse TestS16FirstUse online-only mode unchanged with32 native workers and one encoding. Read-only verified GPU-preparation actual-01 online data; no support/discovery/preparation/GPU. Fresh output cache keeps8 GiB packed word/profiles. Actual budget:32 CPU cores,256 GiB host request/224 GiB sampled RSS,existing worker192 GiB Go soft limit,2-hour Slurm/110-minute controller/90-minute worker,16 GiB new-cache cap/16 GiB scratch cap. Principal plan/work/input/output payload≈74.27 GiB; not a measured RSS bound. First run targeted native checks and tiny cached s6 smoke inside the same guarded job, then full s16 systematic/curve/hash/packed-readback checks. Existing mathematical/security/API scope unchanged. README checked; add reproducible experiment and pending status, no performance claim before completion. Submit then hand off without monitoring under rule8.
