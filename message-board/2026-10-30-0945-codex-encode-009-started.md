# ENCODE-009 started

Timestamp: 2026-10-30T09:45:15Z. Agent: Codex.

Record and implement complete C++ database-dependent cyclic encoding before the
s=15 admission study. Keep reusable discovery/preprocessing in Go. Preserve the
parallel Taylor/FFT kernels, workspace ownership and exact canonical output;
move buffer preparation, point products and output assembly into one native call.

Validate against Go and the dense oracle, exercise invalid inputs and concurrent
workspaces, then profile cached s=14 using retained constants (no repeated setup).
Report initialization, encoding, memory, native stages and output hashes in a new
[experiment](../experiments/native-online/report.md). Check the root README and
update the [scaling plan](../docs/encoding-scaling.md) and backend usage docs.
