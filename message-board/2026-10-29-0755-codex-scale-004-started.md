# Integrate parallel CPU encoding

Timestamp: 2026-10-29 07:55 UTC. Agent: Codex. Work: SCALE-004. Status: started.

The user authorized expanded scaling-plan steps 1–3. Implement general-basis/affine native transforms, parallel matrix/Taylor execution with blocked transfers, buffered minimal-plan loading and reusable workspaces. Validate small complete outputs and the retained s=13 hash, with CPU profiling and measured setup/load/encoding/RSS. Keep preprocessing in Go and old reference/default APIs intact. The 3–6× end-to-end gain is a target, not evidence.

Update [scaling design](../docs/encoding-scaling.md), [native guide](../encoding/liftrs-cpp/README.md), [shared guide](../encoding/liftrs-shared/README.md), [README](../README.md), and a new reproducible experiment. Preserve caches and historical measurements; no s=14/15 admission or GPU dependency.
