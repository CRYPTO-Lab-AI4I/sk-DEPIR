# SCALE-008 started

Timestamp: 2026-11-06T08:56:59.212229+00:00. Agent: Codex.

Published the bounded-preparation checkpoint as `3497bbd` (pushed to main).
Admit the user-requested local s=15,m=2,t=5 experiment after small-buffer tests
of 30-bit extension arithmetic, native import and allocation-free loader guards.
Extend only explicit experimental preparation/native/online-file limits; public
DiscoverOrbit remains s<=13. Reuse bounded preparation; use complete C++ online
encoding, retained constants/profiles, full systematic and sampled curve checks.

New [parallel-s15 experiment](../experiments/parallel-s15/report.md) will separate
setup, exact discovery, preparation, persistence, three encodings and fresh cached
execution. Keep 24 GiB Go soft target and 32 GiB RSS guard, with 20 GiB cache budget.
Allow four hours of active compute: scaling the measured s=14 preparation alone
suggests over two hours, so the earlier two-hour ballpark is too short for first
use. Start with the same low-battery pause/resume policy. Current host has 48 GiB
RAM, about 759 GiB free disk and confirmed AC at 100%. No s=15 feasibility claim
until measurements and checks pass; a guard stop preserves completed cache stages.
Check root README and report exact measured source provenance.
