# Portable field multiplication optimization

- Timestamp: 2026-10-28 16:30 UTC
- Work item: ENCODE-008
- Status: started

Compare branch-free bit multiplication and four-bit multiplication with the
[current kernel](../encoding/liftrs-go/binary_field.go), then select using kernel
and complete cyclic encoding measurements. Include table construction costs.
If general multiplication still leaves a justified hotspot, consider reusable
fixed-constant tables with an explicit memory budget (including tower delta).
Preserve field polynomials, coordinate mappings, prepared artifacts and codewords.
No CPU-specific instructions or new dependencies are planned.

Validate against independent arithmetic and existing encoder oracles; measure
code discovery, reusable preprocessing, per-database encoding and memory
separately through s=12. Recompute whether s=13 meets the agreed one-hour
first-use / 12 GiB forecast policy before admitting it. New scripts, reports and
raw results will live in `experiments/field-multiplication/`; preserve the
[previous scaling results](../experiments/cyclic-scaling/report.md).
Check README and relevant design/package documentation, record commands,
results, limitations and the final algorithm selection here when complete.
