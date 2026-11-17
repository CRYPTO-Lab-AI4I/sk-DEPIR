# Streamed canonical export ready for bounded validation

- Timestamp: 2026-11-17T09:14:59.088690+00:00
- Agent: Codex
- Work: SCALING-100
- Status: Prepared

Implemented shared verified polynomial stage and explicit export in
[preprocessing.go](../encoding/liftrs-shared/preprocessing.go) and
[preprocessing_export.go](../encoding/liftrs-shared/preprocessing_export.go).
DefaultPrepare and PreparePolynomials share root/reciprocal/full identity logic;
WriteCode calls the passed GPU SpectrumTo primitive and canonical header/P codec.
Public API preserves immutable caller support/root, session serialization, error
recovery and independent data lifetime. No new polynomial/transform algorithm.

CLI `--spectrum-tile-bytes` opts in; zero remains resident. Only command-owned
support/recovery references are dropped after identity; persisted support/root/
recovery and callback stay. Common atomic publication/hash/fsync/range/readback
logic handles both modes; no new strict manifest fields or incompatible bytes.
Both CLI modes release final prepared references before bounded readback. No
forced GC, extra transform or output bytes. Runtime/peak changes unmeasured.
Source-derived s18 final arrays256.001GiB with1MiB tile are not RSS or feasibility;
discovery/product/device pressure remains.

Verification: focused native polynomial/resident equality, existing native
independent-byte/restart tests, recovery-boundary tests and option tests pass
(~3.7s); Go-only no-CUDA option/unavailable-backend tests pass; CPU/CUDA test
packages compile; Python AST/controller policy, shell syntax, relative docs and
whitespace checks pass. No GPU runtime or full peak claim yet. New GPU API
writer/range/lifecycle/immutability/reuse tests and complete CLI stream/resident/
Go byte checks, restart/corruption/admission recovery, CUDA-free readback and
four-GPU memchecks are queued in the [new recipe](../experiments/preprocessing-export/report.md).
Existing small resident CLI/race/ASan/vet suites run as regressions in that job.

README, API/CLI docs, active reference/plan and experiment index checked/updated.
Existing budget4H200/4GiB-owned,4CPU/8GiBhost/6GiBRSS,2GiBretained/4GiBscratch,
20min Slurm; no resource increase or fulls18 admission. Freeze producer, submit
fresh export-01 and hand off without monitoring. GPU76706 remains uninspected.
