# Step-3 scaling prototypes

- Timestamp: 2026-10-28 23:14 UTC
- Work item: SCALE-002
- Status: started

Implement opt-in correctness/resource probes for sparse transform setup, cyclic
root-only/Frobenius-paired preparation, lean online workspaces, and exact bounded
external transform/output execution. Keep defaults unchanged. Compare constants,
polynomials and words against independent existing paths; retain profiles and
reasonably sized reusable probe data. Separate live workspace accounting from
Go heap/RSS and logical I/O from actual storage-device traffic. Start with small
instances, then bounded setup measurements; avoid rebuilding the s=13 cache or
rerunning comparable full performance baselines. Document any unimplemented
matrix decomposition or unresolved scale gate explicitly rather than claiming
a prototype proves target-size practicality. Update study, report and README.
