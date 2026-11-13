# GPU-039: Component evidence reconciled

Timestamp: 2026-11-12T19:34:55.704980+00:00
Agent: Codex
Status: completed

Job74121 passed,370 source identities/cache and evidence hashes verified;8 Go/resident/inverse cases,6 memchecks,80 matrix regressions,race/vet passed. On gnode17 H200 completed3m59s. Dimension22 s16: resident0.127805s vs backed7.734124s, GPU34352772→1322628bytes, host domain16MiB+256KiB stage, H2D1186942976/D2H171921408bytes. This is a substantial measured runtime penalty, not a full encoder estimate. Preserve all evidence/profiles. README checked and updated; full integration and CPU gates remain incomplete.

See [component report](../experiments/gpu-host-backed/report.md), [implementation](../encoding/liftrs-cuda/backed.hpp), [test](../encoding/liftrs-shared/gpu_backed_test.go).
