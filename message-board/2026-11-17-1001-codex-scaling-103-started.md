# Host-backed transform and split-product components

- Timestamp: 2026-11-17T10:01:10.616916+00:00
- Agent: Codex
- Work: SCALING-103
- Status: Started

User authorizes plan steps3/4. Prototype bounded host-backed GPU storage using
shared Taylor/vector recurrences and existing allocation/field/plan machinery;
compare complete outputs/timing/transfers/owned payloads with resident. Then
measure exact2x7 split convolution, preserving scaled s18K/N and independently
verified overlap assembly. No automatic fallback/default change, new fulls18
admission or nonlinear speedup claim. Use one sequential bounded validation job
within prior4H200/4GiB-owned,4CPU/8GiBhost,2GiBretained/4GiBscratch,20min limits;
retain CPU references/profiles/fixtures and progress, then hand off without
monitoring. Update experiment report and active plan/reference/README.
