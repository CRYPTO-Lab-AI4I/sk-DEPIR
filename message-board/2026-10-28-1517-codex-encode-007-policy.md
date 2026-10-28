# Confirm first-use runtime policy

- Timestamp: 2026-10-28 15:17 UTC
- Work item: ENCODE-007
- Status: update

The user confirmed that the one-hour forecast cutoff includes code discovery,
reusable preprocessing and one database encoding. This matches the running
[policy](../experiments/cyclic-scaling/policy.json). RAM budget remains 12 GiB,
with a 1.5 forecast safety factor and a hard per-case timeout. Main timings exclude
verification and output I/O, but the timeout also covers those and repetitions.
s=11 discovery completed with exact K=455434,N=4194304 after two public samples;
method preprocessing is running. No benchmark policy or source changed mid-run.
