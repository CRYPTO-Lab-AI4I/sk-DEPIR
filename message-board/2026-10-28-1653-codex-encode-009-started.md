# Benchmark s=13 with a two-hour allowance

- Timestamp: 2026-10-28 16:53 UTC
- Work item: ENCODE-009
- Status: started

The user clarified that one hour was a ballpark, not a strict requirement, and
authorized the s=13 benchmark with a two-hour timeout. Keep the existing 12 GiB
observed Go-memory guard and 9 GiB soft GC target. Run s=13,m=2,t=5 using the
unchanged cyclic-scaling driver and fixed-constant backend from `7eff0ff`.
Separate exact discovery, reusable preparation, three database encodings and
validation/output. Preserve all earlier policies/reports as historical evidence.

New [experiment](../experiments/cyclic-s13/report.md) will include a runner,
explicit policy, raw events, process/progress metadata, source hashes and report.
Check all systematic symbols and 16 actual curves with independent arithmetic;
report exact capacity and canonical/packed hashes. No production integration,
large-cache serialization or paper-encoder comparison is implied. Review README
and design/catalogue status after results are known; record all outcomes here.
