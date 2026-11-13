# Width-aware code-data validation submitted

- Timestamp: 2026-11-13T20:22:29.252845+00:00
- Agent: Codex
- Work: SCALING-056
- Status: handoff

Published corrected matrix checkpointd283f8a, then implemented f24682c. New ScalingCodeData/header/limits APIs load resident arrays once at s-dependent widths, verify streamed payloads without domain arrays, write canonical formats and compute64-bit systematic coordinates. Existing canonical LRSprobe1/2 bytes remain unchanged; smaller-s legacy wide-count input stays readable and writes canonical counts. LRSprobe3 stores explicit field/tower/generator and8/4/8 slot metadata at s17..20; offsets/counts remain64-bit, base P stays32-bit. Mandatory trusted SHA/count/byte caps, field/range/P-constant/EOF validation and narrowing protection. No actual s17 preprocessing or encoder admission changed.

[Format/API spec](../docs/scaling-code-data.md), [controller/tests/report](../experiments/s17-code-data/report.md), [submission](../experiments/s17-code-data/submission.json). Public comments/example and root/shared/experiment READMEs, scaling reference/plan updated. Tests cover small real legacy round trips, unchanged bytes/RAM widths, alternate old headers, all new header widths/truncations/overflows, allocation caps and digest/range/error checks, high-value8-byte slot chunks. Whole genuine s17 payload validation remains unavailable; header and slot-codec checks are separate.

Static gofmt/controller --check/Python/shell syntax, local documentation links and git diff --check pass. Full Go-only and CPU regressions are in job74396. Same4CPU/noGPU/8GiB request/6GiB RSS,2GiB cache/4GiB scratch,10m worker/15m controller/20m Slurm. Results/profiles experiments/s17-code-data/runs/data-01/, log experiments/s17-code-data/slurm-74396.log, matched executables .cache/liftrs/s17-code-data/data-01/. Only deterministic tiny unit preparation is regenerated; existing expensive caches untouched. No post-submit monitoring or other-job inspection; await user completion ping. No full s17 encoding/runtime/table measurement claimed.
