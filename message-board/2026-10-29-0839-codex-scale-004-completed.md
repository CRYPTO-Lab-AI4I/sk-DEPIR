# Parallel s=13 CPU integration completed

Timestamp: 2026-10-29 08:39 UTC. Agent: Codex. Work: SCALE-004. Status: completed.

Documented and completed expanded integration steps 1–3 in the
[scaling plan](../docs/encoding-scaling.md). The calibration checkpoint was first
committed/pushed as `6a9f26f`. This new integration remains a separate working-tree
change for review.

## Implementation and documentation

- [Native ABI](../encoding/liftrs-cpp/c_api.h), [execution](../encoding/liftrs-cpp/c_api.cpp),
  [kernel](../encoding/liftrs-cpp/transform.hpp) and [binding](../encoding/liftrs-cpp/plan.go):
  general ordered bases and affine offsets; parallel matrix rows/columns,
  blocked in-memory transposes; reusable code-only Taylor tables; elapsed stage
  accounting, ownership checks and explicit pthread flags.
- [Shared adapter](../encoding/liftrs-shared/scaling_matrix_cpp.go),
  [minimal loader](../encoding/liftrs-shared/scaling_online.go),
  [runnable example](../encoding/liftrs-shared/scaling_matrix_cpp_example_test.go),
  [matrix tests](../encoding/liftrs-shared/scaling_matrix_cpp_test.go),
  [loader validation](../encoding/liftrs-shared/scaling_online_validation_test.go)
  and updated old file-test helper: experimental opt-in API, independent
  workspaces and checksum/shape/field/prefix-capacity checks. Codewords retain
  the original information set and canonical ordering.
- Updated [README](../README.md), [roadmap](../ROADMAP.md),
  [encoder contract](../docs/encoder-contract.md),
  [algorithm status](../docs/encoding-algorithms.md), both backend guides and
  [experiment index](../experiments/README.md). README checked and updated to
  distinguish current parallel evidence from historical serial forecasts.

## Measured evidence and checks

[Experiment report](../experiments/parallel-encoding/report.md) links runner,
analysis, manifests, raw phases, profiles, native samples and verification scripts.
Final three-repeat medians are 29.795 s at eight workers and 20.152 s at sixteen;
all six full s=13 canonical hashes match. Peak whole-process RSS is 1.148 GiB.
The sixteen-worker result is ~7.61× faster than the historical 153.409 s native
median, not a fresh matched trial. Minimal loading is 0.167 s; setup and workspace
allocation are timed separately. The first integration also passed, but built
small Taylor maps online; preserve its snapshots/results separately from the final
reusable-preparation measurements. Native sampling overlaps one final repeat.

Default and native full suites pass in an isolated copy. Tagged vet, focused
race tests, dense/Horner/inverse checks, workspace reuse/concurrency, malformed
file tests, ABI/arithmetic ASan/UBSan and a new matrix ASan/UBSan fixture pass.
The final prefix guard, runnable example and explicit pthread linkage received
focused follow-up checks. Documentation links, Python syntax, evidence/source
checksums and whitespace checks pass. Sanitizers are not a race proof; Linux
platform qualification and sustained storage measurements were not performed.

Preprocessing/profiles/binaries are retained under the experiment manifest's
`.cache/liftrs/parallel-encoding/` paths; original s=13 caches remain read-only.
Content-addressed source snapshots preserve intermediate measured versions.
No large preparation was repeated, public default was switched, or new s admitted.
The next gate is s=14 exact discovery/preparation and complete encoding under
reviewed budgets, followed by s=15. Disk-backed execution, larger commodity hosts,
application integration and GPU offload remain separate scope.
