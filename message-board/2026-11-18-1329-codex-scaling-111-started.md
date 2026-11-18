# Compare fully checked root-subset validators

- Timestamp: 2026-11-18T13:29:50.417885+00:00
- Agent: Codex
- Work: SCALING-111
- Status: Started

User approves measuring smaller subset validation and repeating root-worker
comparison. Add explicit sorted-membership option sharing cardinality/range/order/
Frobenius checks; retain bitmap validator as ordinary default. Sorted membership
uses existing indices with O(KlogK) search work and no N-sized bitmap; latency
trade-off unknown, no fullsupport/default adoption. Test equivalence/admission/
wide-index memory bounds. Repeat bitmap vs sorted root-worker benchmark with
matched CPU-oracle setup, same cases/1/2/4workers/merge/profiles and complete
CPU/scalar/hash checks. Use existing one-node4H200/4CPU/8GiBhost/6GiBRSS/
4GiBGo/4GiBowned/2GiBcache/4GiBscratch/20min scheduler budget. No new node/fullsize
or storage load. Document methodology changes and report pending results in main
reference, plan and README; commit source then submit and hand off, no monitoring.
