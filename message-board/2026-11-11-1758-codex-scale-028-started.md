# SCALE-028 — Audit s16 lifetimes and compare s17 memory schedules

Timestamp: 2026-11-11T17:58:33.622159+00:00
Agent: Codex
Status: started

After plan commit24bbad9 was pushed, begin source-based phase memory accounting and reproducible alternatives comparison. Inspect discovery/preparation/load/CPU/GPU workspace/output/verification lifetimes, reconcile s16 against retained measurements and project s17 with explicit width assumptions. Produce a small allocation-free analytical experiment with script/report/results, source provenance and checks; no hardware timing claims. Compare lower-memory schedules and extra work/traffic before asking the user to choose a material trade-off. No full job, large allocation or chosen-schedule implementation is authorized by this audit alone.
