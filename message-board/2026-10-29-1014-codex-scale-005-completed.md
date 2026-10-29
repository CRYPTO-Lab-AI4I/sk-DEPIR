# Profiled s=14 first use and cached encoding completed

Timestamp: 2026-10-29 10:14 UTC. Agent: Codex. Work: SCALE-005. Status: completed.

Published the previous parallel integration as `9ac6ae8`, then completed the
requested s=14,m=2,t=5 run with retained Go/native profiles and preprocessing.
This new experiment, catalogue evidence and code changes remain uncommitted for
review. The [report](../experiments/parallel-s14/report.md) links all scripts,
raw phases, verification, profiles, native samples and source/binary/cache hashes.

## Results

Exact K=50,976,325 at N=268,435,456; rate 0.189901609. Discovery takes 254.898 s
(two samples), reusable preparation 2,017.985 s including the complete P·G check,
and three profiled encodings 90.521/90.533/91.987 s (median 90.533 s).
A fresh cached process takes 92.118 s and peaks at 4.589 GiB RSS, compared with
13.726 GiB first-use peak. Setup/discovery/preparation plus one median encoding
is 39 min 24 s, excluding artifact I/O and verification. All four canonical hashes
match, with every systematic symbol, 16 independently evaluated degree-5 curves,
packed write/fsync and 1,027 readback positions checked.

Public code files under `.cache/liftrs/parallel-s14/s14-m2-t5/` total 2.19 GiB:
support.bin, online.bin (1.19 GiB P/spectrum), g.bin. Packed synthetic output is
448 MiB. Checksums, metadata, measured binaries and source provenance are kept;
original s=13 caches remain retained. A first-use-process RSS sampler motivated
the fresh cached check because old runtime pages kept RSS high after logical
preparation release. No discovery/preparation was repeated for that check.

## Changes and verification

- [Native width guard](../encoding/liftrs-cpp/c_api.cpp) and
  [independent arithmetic tests](../encoding/liftrs-cpp/test.cpp) extend through
  s=14. [Sparse preparation executor](../encoding/liftrs-shared/scaling_prepare_cpp.go)
  reuses code-only native matrix plans and transpose scratch under Go orchestration.
- [Exact discovery core](../encoding/liftrs-shared/orbit.go) is shared without
  lifting default public discovery/JSON basis guards. The
  [paired preparation core](../encoding/liftrs-shared/scaling_experimental.go)
  accepts the same field executor. The [minimal reader](../encoding/liftrs-shared/scaling_online.go)
  admits s=14 explicitly bounded artifacts; mathematical/information-set semantics
  are preserved.
- [First-use/cached probe](../encoding/liftrs-shared/parallel_firstuse_probe_test.go),
  [small preparation/field checks](../encoding/liftrs-shared/scaling_prepare_cpp_test.go)
  and [experiment scripts](../experiments/parallel-s14/report.md) retain checkpoints,
  atomic binary files, profiling, resource guards and output verification.
- [Catalogue](../pirconfig/configs_catalogue.json), [guide](../pirconfig/README.md),
  [catalogue test](../pirconfig/catalogue_test.go) and
  [exact-evidence test](../pirconfig/s14_evidence_test.go) distinguish verified s=14
  standalone encoding from application integration, preserving Monte Carlo evidence.
- Updated README, roadmap, encoding guides, algorithm/contract/scaling documents
  and experiment index. README checked and updated with current measurements,
  resource boundaries, retained artifacts and public/default API distinctions.

Both full repository suites pass in an isolated copy, as do focused race checks,
tagged vet, native arithmetic/table ASan/UBSan through s=14, catalogue/planner tests,
source/cache/evidence analyzers, cross-process hash equality, documentation links,
Python syntax and whitespace checks. Failed early RSS-access and cached-harness
build attempts are preserved and explained in the report. No full s=14 Go baseline,
application/cryptographic integration or s=15 qualification is claimed.

Next: assess s=15 first-use lifetimes/RAM and time budgets using these retained
profiles and caches before admission; larger commodity hardware remains an option.
