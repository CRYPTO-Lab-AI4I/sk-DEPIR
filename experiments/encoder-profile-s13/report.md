# CPU profiles and retained preprocessing at s=13

## Purpose and reproduction

Profile discovery, reusable preparation, cache save/load and one encoding with
[main.go](main.go), then compare the complete output with the
[unprofiled baseline](../cyclic-s13/report.md). CPU sampling is enabled by default;
forced GC and profile start/stop are outside phase timers, while sampling during
the stage is included. Preserve unprofiled latency results separately.

From the repository root (Go, Python 3/POSIX wait4 and local disk space):

```sh
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoder-profile-s13/verify.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoder-profile-s13/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name setup-new
# Subsequent encoding profiles reuse the saved cache without code preparation:
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoder-profile-s13/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --reuse-cache --run-name reuse-new
```

The preparation command requires the default cache to be absent; use it on a
fresh checkout/machine. On this machine the cache already exists, so use the
reuse command with an unused run name. For a separate regeneration while retaining
the existing cache, build `./experiments/encoder-profile-s13` and pass a different
`-cache` directory to [the driver](main.go), retaining the recorded source provenance.
[run.py](run.py) preserves each named run directory instead of overwriting it.
It records exact source hashes, revision, policy, environment, CPU profiles,
phase/memory observations and verification. [policy.json](policy.json) retains
the two-hour whole-case timeout, 12 GiB Go-memory guard and one encoding call.
The host is the same Apple M5 Pro/48 GB machine as the unprofiled run; runtime
metadata is recorded again. No tests or other benchmarks run beside timing.

## Cache retention and boundaries

Keep the complete trusted JSON cache under ignored
`.cache/liftrs/s13-m2-t5/`, with code specifications, SHA-256 and build-source
provenance. The retained profiling binary is also stored there for later pprof
analysis. The experiment copies cache metadata/provenance into its small results;
large cache/binary files stay outside Git. There is no automatic deletion or
silent replacement of an existing cache. Source changes need not invalidate
mathematically compatible constants: cache-build provenance and profile-build
provenance are recorded separately and the complete output must match baseline.

A 2 GiB cache-file budget is separate from the 12 GiB process Go-memory guard and
proposed public distribution budget. The opt-in
[`LoadCyclicEncoderWithLimit`](../../encoding/liftrs/orbit_artifact.go) applies
its byte limit to both complete JSON and embedded basis. Default standalone
2/64 MiB caps remain unchanged. Input reads are capped independently and decoded
memory is guarded; byte caps alone are not RAM bounds. Structural loader checks
assume trusted provenance and do not prove mathematical authenticity.

The driver prepares once, serializes/syncs/checksums the cache, drops its original
encoder, then actually reloads that file before profiling encoding. Save/load
and profile start/stop are separate from encoding. The first-use arithmetic sum
(discovery + preparation + encoding) remains comparable in boundaries to the
baseline but is instrumented. `workflow_s` additionally includes cache save/load;
whole-case wall time includes verification and instrumentation setup/finalization.
Later reuse runs omit discovery/preparation. Caches contain no database data or
cryptographic keys. The final output uses the same seed-71 message, all systematic
symbols, 16 independent actual curves and canonical/packed hashes as baseline.
Heap sampling and Go-runtime `Sys` are observations, not exact workspace bounds;
whole-case RSS is separate. This is not an application-plan or GPU benchmark.

## Preflight verification

[verify.py](verify.py), [verification.json](verification.json) and
[verification.log](verification.log) record passing loader/driver tests, vet and
an isolated full Go suite. [smoke.py](smoke.py) and
[smoke-results.json](smoke-results.json) exercise fresh preparation, actual
cache reload, CPU-profile creation and a subsequent reuse-only run over a
s=4,m=2,t=2 correctness fixture. Disposable tiny fixture files are not optimization
targets; only that smoke test uses temporary caches/profiles.

[Loader tests](../../encoding/liftrs/orbit_cache_test.go) check exact byte limits,
negative/zero limits, default complete/basis caps, larger opt-in nested basis,
malformed inputs and byte-identical round trips/encoding.
[The public example](../../encoding/liftrs/orbit_cache_example_test.go) bounds
reads before loading. [Race checks](race.log) passed for the changed loader and
existing mathematical package. No field/transform/encoding algorithm changes.

## Results and recovery

The setup run is recorded in [setup/results.json](setup/results.json),
[raw events](setup/events.jsonl) and [cache metadata](setup/cache-metadata.json).
Discovery, preparation, cache save and cache load completed under CPU sampling.
Then the monitor expected a discovery `samples` field in a post-load code event
and raised `KeyError`. Its cleanup killed the child before a complete encoding
profile. The [original metadata](setup/initial-results.json) and
[exact runner snapshot](setup/run-snapshot.py) are preserved; the corrected status
is `monitor_failed_after_cache_load`. The zero-byte `setup/encode_0.pprof` is
incomplete and excluded. Setup process peak RSS was not captured and is not
reconstructed from heap samples.

The fixed monitor accepts the actual post-load metadata and saves partial results
on unexpected exceptions. [test_run.py](test_run.py) checks the logger against
real fresh/reuse smoke events; [monitor-regression.log](monitor-regression.log)
records its passing result. This change affects the Python monitor only; the Go
encoder, cache and profiled mathematical code are unchanged.

A [reuse-only run](reuse-01/results.json) reloaded the retained cache and completed
one profiled encoding. Its [events](reuse-01/events.jsonl),
[profile](reuse-01/encode_0.pprof) and [cache metadata](reuse-01/cache-metadata.json)
are preserved. Discovery/preparation were not repeated. Source hashes identify
the original `0d94d4c` setup and the monitor fix in the reuse working tree;
cache-build and profile-build manifests are recorded separately.

| Stage | Source run | Profiled time |
|---|---|---:|
| Discovery | setup | 85.109 s |
| Reusable preparation | setup | 1,784.863 s (29.75 min) |
| Cache serialize/hash/write/sync | setup | 1.498 s |
| Actual cache load | setup | 4.864 s |
| Actual cache load | reuse-01 | 4.834 s |
| Database encoding | reuse-01 | 186.352 s |

The cache is **453,243,425 bytes (432.25 MiB)**, retained at
`.cache/liftrs/s13-m2-t5/cyclic.json`, with `metadata.json`, `provenance.json` and
matching `profile-bin-*` binaries in the same ignored directory. SHA-256:
`a2a3bcd5913d821c99d08ceca05d6c5d73ce2abe6e3987910c0dcf5dfe97eb22`.
This is below the 2 GiB local file budget; nothing was evicted or published as a
large Git artifact. Cache-load time includes capped reads, checksum, JSON decoding,
structural/Frobenius checks and runtime-table reconstruction, not polynomial setup.

The reuse process took **194.28 s** wall time, with **2.257 GiB** peak RSS.
Sampled peak heap during its encoding was 1.236 GiB; retained cache/message/output
are included. Setup's sampled preparation peak heap was 4.601 GiB; peak process Go-runtime
`Sys` was 7.260 GiB during preparation and remained at that level through cache
save/load, below the 12 GiB guard.
These are observations, not OS-enforced or exact workspace bounds. No resource
guard fired. Setup and encoding come from **separate processes**: do not present
one combined wall time or peak RSS for them.

The unprofiled preparation baseline is 1,765.374 s and median encoding 185.153 s.
These profiled observations are about 1.10% / 0.65% higher, respectively, but this
is not a controlled measurement of profiling overhead: run conditions, one versus
three encoding calls, and cache loading differ. Preserve the unprofiled baseline.
No optimization or speedup was implemented by this experiment.

## CPU attribution and next candidate

[analyze.py](analyze.py) regenerates [table.md](table.md), [analysis.json](analysis.json)
and flat/cumulative/call-tree/source views from the retained profiles/binaries,
without rerunning setup or encoding:

```sh
python3 experiments/encoder-profile-s13/analyze.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go
PYTHONDONTWRITEBYTECODE=1 python3 experiments/encoder-profile-s13/test_run.py
```

| Self CPU sample share | Preparation | Encoding |
|---|---:|---:|
| Fixed-constant lookup kernel (`constantProduct.mul`) | 38.46% | 44.20% |
| Base-field multiplication (`BinaryField.Mul`) | 27.14% | 26.48% |
| Transform recursion's own work | 17.56% | 17.51% |

These flat shares exclude callees. Cumulative shares include callees and overlap;
they must not be added to flat shares or interpreted as fractions of all wall
time. For example, preprocessing polynomial division is 67.54% cumulative,
including the arithmetic above; encoding transforms are 97.67% cumulative and the
prefix low product is 40.46% cumulative within that work. Collected samples cover
90.89% / 90.26% of profile duration; the remaining duration is not classified as
I/O or any other bottleneck by this measurement.

Primary views: [preparation flat](setup/preparation-flat.txt),
[cumulative](setup/preparation-cum.txt), [source](setup/preparation-source.txt),
[encoding flat](reuse-01/encode_0-flat.txt),
[cumulative](reuse-01/encode_0-cum.txt) and [source](reuse-01/encode_0-source.txt).
Raw profiles for discovery, preparation, cache save/load and encoding remain in
the corresponding directories. The next candidate is reducing fixed-constant
lookup-kernel overhead: compare unrolled lookups and alternative row widths under
an explicit table-memory budget, then test complete encoding using the retained
cache. The profile does not prove whether loop overhead, cache behavior or another
factor dominates inside that kernel, nor that a wider table will be faster.

## Verification and limitations

The reuse output passes all 11,926,435 systematic symbols, sampled coordinate
mappings, 16 independently evaluated actual curves and packed sync/readback.
Complete canonical and packed hashes equal the unprofiled s=13 baseline.
[analyze.py](analyze.py) verifies runtime source snapshots, raw events, profile
files, retained binary/cache hashes and baseline agreement. The
[evidence manifest](evidence-manifest.json) covers retained scripts/results/views.
Preflight Go tests, smoke checks, race checks and the isolated full suite passed;
the logger regression and completed reuse run validate the subsequent monitor fix.
No Go source changed after those checks.

The large cache and matching binaries are local and ignored by Git; copying the
small results alone does not copy that local state. Their specs/provenance and
regeneration commands permit rebuilding it. New run names preserve existing
profiles; never delete a useful cache merely to satisfy the default run name.
The process releases RAM on exit, while the reusable data remain on disk.
There is no new production-security, application-plan execution, GPU,
external-memory or s>=14 support claim.
