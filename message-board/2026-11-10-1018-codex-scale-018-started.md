# SCALE-018 — bounded s16 intervals and operation census

Timestamp: 2026-11-10T10:18:32.517226+00:00
Agent: Codex
Status: started

Resident integration products-02/job72839 passes exact products, CUDA memcheck and complete smaller pipelines. Proceed with the user-authorized bounded s16 work toward an estimate. Read the saved support file only, verify its original digest/spec/sorted order, count accepted root factors and batch degrees over the full support without rediscovery or a full support allocation. Compare the first two consecutive 524,288-position root-product intervals on current CPU (cutoff32768) and resident GPU (cutoff1024), keeping the same builder across intervals. Compare final partial products exactly; label them partial, not full code/preprocessing. Maximum prefix1,048,576 support entries; actual transforms≤2^22 values.

Policy unchanged: one GPU/16 CPUs (four used in these partial native products), 16 GiB host request/12 GiB sampled process RSS, 512 MiB explicit device bound, 30-minute allocation/25-minute deadline. Preserve reference caches and all new profiles/partial fixtures. Record operation census for later carries/Newton/identity/spectrum, identifying sizes outside the measured/admitted range. No unsupported full-domain GPU shape or full s16 run is admitted. A whole-run estimate must explicitly expose unmeasured later stages rather than scaling the s10 ratio. README/current-plan check and exact provenance required.
