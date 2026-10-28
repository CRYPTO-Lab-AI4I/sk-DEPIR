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
  --go /tmp/sk-depir-go-toolchain/go/bin/go --run-name setup
# Subsequent encoding profiles reuse the saved cache without code preparation:
GOCACHE=/tmp/sk-depir-go-cache python3 experiments/encoder-profile-s13/run.py \
  --go /tmp/sk-depir-go-toolchain/go/bin/go --reuse-cache --run-name reuse-01
```

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

## Results

The long profiled run is pending. Results will be recorded under `setup/`,
including CPU flat/cumulative/call-tree/source views, cache size/load costs,
profiled timings and baseline hash agreement. No new speedup is claimed.
