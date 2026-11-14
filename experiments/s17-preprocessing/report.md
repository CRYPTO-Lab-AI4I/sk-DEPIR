# Shared width-aware preprocessing validation

Status: implementation committed as `ec12f1b89ee1db4c7eebdbcadf33a7b437948b80`; bounded validation
job74424 submitted ([submission](submission.json)), results reviewed below.
This does not establish full s17 preprocessing or CUDA17 support.

## Reproduction and evidence

- Controller: [run.py](run.py); Slurm entry: [run.sbatch](run.sbatch).
- Independent native checks: [check.cpp](check.cpp).
- Reconciliation after completion: [analyze.py](analyze.py).
- Results: `runs/native-01/results.json` and per-command logs/profiles alongside it
  (created by the submitted job; see `submission.json` for its ID/revision).
- Reusable small artifacts and matching test/native binaries:
  `.cache/liftrs/s17-preprocessing/native-01/`; manifests and source/evidence/cache
  SHA-256 maps are sealed in results. These caches are outside Git.

```sh
module load slurm/slurm/25.05 gcc/14.2.0
sbatch --output=experiments/s17-preprocessing/slurm-%j.log \
  experiments/s17-preprocessing/run.sbatch native-01
# After completion is reported:
python3 -B experiments/s17-preprocessing/analyze.py native-01
```

The job records exact Git revision, copied source hashes, hostname, CPU, compiler
and Go versions, Slurm allocation and commands. Source is snapshotted into fresh
scratch before testing. Go1.27.3 and GCC14.2 are the configured tools; no new
library dependency is introduced. Random oracle fixtures use seed170066; small
example discovery uses crypto/rand. No original laptop cache or existing running
job is read, modified or monitored.

## Method and limits

Four CPU cores,8 GiB requested host memory,6 GiB sampled process-group RSS guard,
4 GiB Go soft limit,2 GiB retained cache and4 GiB scratch caps. Worker commands
have600-second deadlines, the controller900 seconds and Slurm1200 seconds. The
existing guard helpers enforce those limits; sampled RSS is not a hard peak
measurement. There is no GPU request, full s17 domain allocation or new numerical
experiment budget. CUDA32 adapter compilation is checked without loading CUDA;
GPU execution comparison and CUDA17/multi-GPU implementation remain separate.

The suite compares exact monomial indices to independent Go `GoodMonomial` for
s2–8, and criterion witnesses above2^32 at s17. Complete discovery at s3/4/6/8,
m2, t2/2/4/5 uses one/four workers and deterministic readers. It compares the
support/samples, P/G, complete canonical serialized code and checkpoint replay
against the Go field/polynomial implementation. Sampling and union control flow
are shared by design; independent arithmetic and the membership criterion remain
separate. Support/root persistence checks digests, binding, malformed data, EOF,
normalization and caps. Metadata-only full s17 admission with a1 MiB native cap
must reject discovery before allocating a huge grid.

Short s17 transforms are compared directly to Go quadratic-field Horner
arithmetic, products to schoolbook arithmetic, and high-index roots to scalar
power/conjugate/norm calculations. A direct C++ driver forces BOTH uint32/uint64
specializations on the same complete GF(16) domain, checking tensor/Horner,
product, inversion, exact membership and release with one/four workers. This is
bounded width validation, not a genuine complete s17 preparation run.

Go/native regression suites, vet, race, Go ASan plus C++ UBSan, and direct
optimized/ASan/UBSan checks cover the changed shared paths. Long commands write
UTC/monotonic stage start/end/error records and30-second resource heartbeats to
`progress.jsonl`; native work counters and preprocessing events are captured in
per-command logs. The analyzer checks completion/failure progress events.

Each command reports wall seconds and sampled peak RSS bytes, including build
or profile overhead as labeled. Profiled preprocessing timings are diagnostic,
not latency baselines; no speedup or full s17 ETA is inferred. Reusable fixtures
include support+manifest, root+manifest, generator and canonical online data,
with checksums. Profiles and matching binaries remain for analysis. Full s17
artifacts are omitted because no full-size memory/storage budget was admitted.

## Result interpretation

[Local static/native checks](static-checks.json) pass: C++ syntax for all changed
ABIs, Go formatting, script syntax and four direct native small-domain cases
(both widths, one/four workers). The short native check was not profiled; it
establishes bounded primitive correctness only. Its binary is retained at
`.cache/liftrs/s17-preprocessing/local-check/check`; Slurm repeats the checks
with optimized and sanitized binaries and collects the preprocessing profiles.

Pending the completion ping and evidence reconciliation. Existing s16 timing
measurements remain historical; the native executor and weight-class enumeration
need calibration. Wider CUDA execution and a reviewed full-size aggregate memory,
storage and time budget are still required before fastest-server s17 admission.

## Job74424 review and corrected retry

The [sealed results](runs/native-01/results.json) are **failed**, with
[forensic reconciliation](runs/native-01/failure-analysis.json) verifying479
source,16 evidence and26 retained-cache identities. All targeted preprocessing
checks passed, including complete small-code byte/root/generator/restart oracles,
exact membership enumeration, high-bit s17 kernels, budget failures and artifact
checks. Go-only regressions and encoding/PIR/configuration suites also passed.
The broader native regression suite then stopped at `TestCPUEncodingModules`:
its s4,t2 fixture has K97,N256 and L256, exceeding the documented2L≤N admission
for file/disk/compact modes. No preprocessing oracle or mathematical identity
failed. This is incomplete validation, not a successful whole job.

The fixture is corrected to s4,t3 (K37,L128), retaining s6,t4 and s8,t5. A
separate dense-prefix test checks explicit compact/file/disk rejection while
comparing resident/wide outputs with Go. Production admission/layout is unchanged.
Native suites, CUDA compilation, vet/race and sanitizer checks must still run.
No full s17 or CUDA17 claim follows from the passing targeted checks.

Retry `native-02 --reconcile-disk` keeps the same4CPU/8GiB and600/900/1200s
limits. Before validation it uses the existing bounded command guard to invoke
`encoding-table/analyze.py fill-02 --case cpu-disk-s16`, independently rehashing
that completed job's retained large files and resolving historical source
identities. The exact analyzer source is checked against the source snapshot.
This is evidence analysis only, without another full encoding. Its wall time is
separate from preprocessing measurements. Results and profiles go to
`runs/native-02/`, with fresh retained cache outside Git; old evidence is kept.
