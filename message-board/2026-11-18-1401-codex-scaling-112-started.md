# Review bitmap versus sorted subset validation77213

- Timestamp: 2026-11-18T14:01:48.592420+00:00
- Agent: Codex
- Work: SCALING-112
- Status: Started

User reports completion. Reconcile77213 source/evidence/cache seals and exact
matched-mode/reference/output coverage, local exhaustive checks, both CUDA
memchecks and profiles. Compare paired warm root/partition/merge timings and
sampled per-process RSS using the same reference strategy. Update experiment
report, actual scaling reference, plan and README with measured scope and next
recommendation, retaining bitmap ordinary defaults until an explicit measured
selection is adopted. No new job/resource increase or fulls18 speed inference.
