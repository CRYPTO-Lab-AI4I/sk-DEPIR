# Parallel root-worker measurement prepared

- Timestamp: 2026-11-18T11:04:11.637916+00:00
- Agent: Codex
- Work: SCALING-109
- Status: Prepared

77126 reviewed after user's explicit go-ahead: COMPLETED0:0,4min9s,HPCgnode13,
4H200/4CPU/8GiBhost,247.575s controller. Native/root/export/recovery/race and GPU
small/export/memcheck pass. Both memchecks0errors/0leakedbytes;589source/
45evidence/75original-cache seals reconcile. An ignored unrelated script was
captured in producer source but absent from Git. Its exact matching snapshot
is retained outside Git at.cache/liftrs/preprocessing-root-boundary/root-01/
review-source-inspect.py, and explicitly named by the reviewer; all other sources
stay Git-bound. Original ignored material is not published/modified. No broad
review exception. Profiles/binaries/cache retained. Sampled maximum command RSS
587.914MiB includes build/test work, not a memory-saving measurement.

[New experiment](../experiments/parallel-root-workers/report.md): shared normalized
root helpers and existing factor/product arithmetic, deterministic complete-orbit
partition helper with immutable sorted copies and degree loads within2. Reserve
partition buffers once, retaining K+O(workers) copied indices; fullsize support
sharding remains absent. Product adapter now depends on Product only; no dispatch
or API/default change. Existing caller scanned. New opt-in GPU driver compares
whole-root/four-card schedule with1/2/4 one-GPU workers and balanced four-GPU merge.
Genuine s8 plus synthetic s12/s17 subsets65536/262144 indices, including>2^32.
Complete native CPU references, degree/normalization, independent scalar Horner
witnesses, full coefficient equality, saved byte hashes and small genuine CUDA
memcheck. Setup/partition/worker/merge and verification/persistence clocks explicit.
No distributed/fullsize speed claim; Go CPU profiles include opaque cgo waits.

Local root partition/equivalence/admission tests pass0.041s; CPP+CUDA optional test
compiles/skips without opt-in env. Reviewer test rejects missing/imbalanced/corrupt
results; Python AST, sbatch bash -n,runner --check and git diff --check pass.
GPU experiment correctness/performance not yet run. Existing bounded controller
adds complete CLI/race/ASan/vet/cache/root restart gates and source/cache seals.
Prior ignored-source snapshot applies only if its exact producer digest matches.

README checked/updated, as are actual scaling reference, preprocessing internals,
experiment index, reviewed prerequisite report and explicit plan ordering:
root-workers+merge first, GPU communication next, buffered/concurrency-controlled
persistence separate, then select GPU count from complete elapsed time. No link
to the investigation report in the plan. No fullsize measurement cells changed.
Use existing one-node4H200/4CPU/8GiBhost/6GiBRSS/4GiBGo/4GiBowned-device/
2GiBnewcache/4GiBscratch,20min scheduler budget. New executors cap128MiB/card,
merge+workers at most1GiB summed owned caps; metadata/context additional. Submit
then hand off without monitoring. No new multi-node GPU or storage budget.
