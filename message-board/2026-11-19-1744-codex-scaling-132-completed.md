# SCALING-132 — completed

UTC: 2026-11-19T17:44:31.791164+00:00

Verified78001 seals654 source/15 evidence/3 cache. First staged same-node payload blocked in send_values/receive_values, zero completed handoffs/proofs/captures, 8waiting heartbeats,90 s test cutoff. Scheduler1min55s, controller115.021s, largest logged RSS44.938MiB/rank below1GiB guard. No corruption/fix/speed inference; later arms unrun. Four workers requested ob1/self,tcp. [Report](../experiments/mpi-transport-control/report.md). Parser now rejects real malformed events but ignores Go stack structs; regression test passes. Raw evidence/status unchanged, README/main reference/plan/index updated.
