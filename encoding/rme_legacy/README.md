# Legacy Reed–Muller encoding

This directory archives the RM-only C++/Julia implementations and their helpers,
build files, Julia dependency environment, benchmark and `encoding_time.png`.
That plot describes the old RM benchmark, **not Lift-RS performance**. The unused
experimental `LRSE2D` helper remains in the historical C++ source; it is not the
current binary-field Lift-RS encoder.

The [process driver](process/process.go) and thin
[PIR compatibility bridge](../../pir/legacy_encoder_bridge.go) preserve the old
file-based `LiftedRS.Encode` testing placeholder. It still produces RM codewords;
it does not use the standalone Lift-RS encoders. Keeping the bridge avoids
silently changing or disabling the existing test API. No RM algorithm or process
handling remains outside this directory; the bridge only converts matrices and
invokes the common permutation/masking stage.

```sh
# Build the legacy C++ process used by file-based tests.
make -C encoding/rme_legacy rme
# Optional historical standalone benchmark.
cd encoding/rme_legacy
c++ -O3 -std=c++17 -o bench_encode bench_encode.cpp
./bench_encode
# Julia fallback environment (Julia itself is optional if C++ is built).
julia --project=. -e 'using Pkg; Pkg.instantiate()'
```

The Julia driver is `GoRMEInterface.jl`; includes are relative to its source
location. The process driver selects the archived C++ binary first and otherwise
starts Julia with this directory's project. Both accept the historical stdin
line protocol. Run PIR tests from the repository root or `pir/`, as before.
Generated binaries remain ignored; the existing local binary was retained here.
The old plot contains historical embedded measurements without the modern
experiment provenance; do not use it as current benchmark evidence.
