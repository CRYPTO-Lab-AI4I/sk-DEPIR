# Portable FFT constant multiplication validated

- Timestamp: 2026-10-28 16:48 UTC
- Work item: ENCODE-008
- Status: completed

## Selection and implementation

The [kernel comparison](../experiments/field-multiplication/main.go) found no
convincing s=9..12 improvement from a masked bit loop or general four-bit Horner
multiplication. Keep public general arithmetic unchanged. Instead,
[constantProduct](../encoding/liftrs-go/orbit_constant.go) prepares immutable
four-bit maps for [FFT level constants](../encoding/liftrs-go/orbit_fft.go).
These are code-only runtime metadata, constructed during discovery/loading and
reused across databases. They preserve field/tower coordinates, information sets,
version-1 artifact bytes and complete codewords.

Row entries are capped at 256 KiB per transform (512 KiB across base/extension),
with scalar fallback; descriptors/allocator overhead are additional. Measured
combined tables use 52.4/64.1/89.6/89.6 KiB at s=9/10/11/12. Changing offsets
and general tower-delta products remain scalar. No CPU-specific instructions,
new dependency or constant-time security guarantee is introduced.

## Measured results and limits

The [report](../experiments/field-multiplication/report.md),
[runner](../experiments/field-multiplication/run.py),
[control runner](../experiments/field-multiplication/run-baseline.py),
[results](../experiments/field-multiplication/scaling-results.json),
[table](../experiments/field-multiplication/table.md),
[summary/forecast script](../experiments/field-multiplication/summarize.py)
and [manifest](../experiments/field-multiplication/evidence-manifest.json)
record reproduction, source hashes, parameters, environment, timing/memory
boundaries and limitations. The initial control overlapped correctness checks and
is preserved but excluded; the fresh s=10 control ran without concurrent tests.
Historical s=11/12 runs use the same unchanged measurement driver.

- s=9,t=7: discovery 0.162 s, preparation 2.348 s, encoding 0.249 s.
- s=10,t=6: preparation 27.45 -> 12.70 s; encoding 2.77 -> 1.22 s (2.28x).
- s=11,t=6: preparation 145.26 -> 64.36 s; encoding 14.60 -> 6.15 s (2.38x).
- s=12,t=6: preparation 677.19 -> 307.59 s; encoding 67.76 -> 29.06 s (2.33x).
  First use is 5.90 min; whole-case peak RSS 0.923 GiB. RSS is an observation,
  not a guaranteed memory saving. Encoding medians use three calls.
- Exact capacities remain unchanged. All systematic symbols, 16 independent
  actual curves and canonical/packed output hashes passed at each measured field.
- The updated s=13,t=5 forecast is 78.0 min first use and 8.42 GiB with the
  agreed 1.5 factor. Time exceeds policy; s=13 was not run. The forecast is
  heuristic, not proof of infeasibility. Loader latency, large serialized caches,
  application integration and larger paper-encoder performance are not measured.

## Verification and documentation

[Tests](../encoding/liftrs-go/orbit_constant_test.go) cover independent arithmetic,
partial nibbles, zero/identity and budget fallback. Additional
[artifact/budget tests](../encoding/liftrs-go/orbit_constant_artifact_test.go) rebuild
byte-identical s=4/8/9 paper/cyclic artifacts and record table sizes through s=13.
The s=13 table-size test is not a full encoding measurement.

[Verification script](../experiments/field-multiplication/verify.py),
[commands/outcomes](../experiments/field-multiplication/verification.json)
and [raw log](../experiments/field-multiplication/verification.log): targeted tests,
race tests, vet, and isolated-copy `go test ./...` all passed. Validation ran after
retained measurements. Source/event/evidence hashes, prior complete outputs,
forecast reproduction, documentation targets and `git diff --check` pass.

README was checked and updated with the report and table-free/general versus
cached-transform distinction. Also updated the [encoder contract](../docs/encoder-contract.md),
[algorithm plan](../docs/encoding-algorithms.md),
[package README](../encoding/liftrs-go/README.md) and
[experiment index](../experiments/README.md). Previous reports/catalogue evidence
are preserved. Changes are uncommitted; commit/push was not requested for this item.
