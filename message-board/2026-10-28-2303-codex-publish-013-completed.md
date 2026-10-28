# Scaling study verified for publication

- Timestamp: 2026-10-28 23:03 UTC
- Work item: PUBLISH-013
- Status: completed (publication checks)

Publishing [five-step plan and completed steps 1–2](../docs/encoding-scaling.md),
[analytical model/report](../experiments/encoding-resources/report.md), source/input
manifest, tables, type probe and verification record. README checked and current;
roadmap, contract, algorithm notes and experiment index agree with the study.
Regenerated the model successfully; all source/input hashes and changed-document
file links pass. Go probe test/build and vet pass, with no workload tests in this
probe package. `git diff --check` passes. No production encoder code changed and
no repeated full encoder benchmark or suite was needed for this publication.
Existing caches/profiles remain intact.

The findings sharpen rather than replace the five-step plan. Step 3 first tests
avoidable local setup/loading/workspace costs, then exact tiled transform and
canonical output routing. This can establish whether s=15 has a practical local
path before choosing larger hardware. External-memory I/O, huge preprocessing
artifacts, exact larger capacities and width changes remain design/prototype gates;
steps 4–5 must calibrate and select a design before a feasibility claim. No HPC
architecture, higher memory budget or production backend is selected. Commit/push
outcome is reported to the user after these pre-publication checks.
