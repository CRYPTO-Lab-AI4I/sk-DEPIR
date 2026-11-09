# Development handoff: larger-memory encoding

Development on the local 48 GiB laptop stops at the verified s=15 checkpoint.
Larger-scale work moves to a larger-memory environment. Start from code/results
checkpoint **`d721182`** and the current documentation commit; retain the local
implementation and fixtures as correctness references. No benchmark is left running
or paused, and the old PID-specific resume commands are historical records.

## Decision and scope

Target **s≥16 on larger-memory commodity systems; feasibility remains unverified**.
Begin with a **single shared-memory CPU node**, provisionally **256 GiB RAM**, a
64-bit OS, sufficient job wall time and local scratch storage. This is a planning
target, not a measured preprocessing requirement or a guarantee that s=16 fits.
A cluster allocation can supply that node; MPI/distributed execution and a GPU are
not prerequisites. Select architecture-specific acceleration from measurements,
retaining a practical generic path. s=17–20 are future design targets, not supported
or demonstrated HPC configurations.

The goal remains fast, reviewable encoding on commodity hardware and **TRL 5**,
not a change to protocol functionality or TRL 6 acceptance. Client/preprocessing
only; storage returns requested bytes through ordinary I/O. Server linear
computation remains future research. See [AGENTS.md](../AGENTS.md),
[roadmap](../ROADMAP.md), [encoder contract](encoder-contract.md),
[data representation](data-representation.md) and [protocol I/O](protocol-io.md).

## Verified starting point

All rows below use m=2,t=5 on the Apple M5 Pro, 48 GiB, macOS arm64 host.
These are standalone mathematical encoders, not an integrated application pipeline.

| Configuration | Evidence | Encoding | Measured process RAM |
| --- | --- | ---: | ---: |
| s=14, K=50,976,325 | [Timing study](../experiments/encoding-timing/report.md) | 84.47 s median full native | about 4.59 GiB cached |
| s=15, K=216,914,047 | [First-use and cached runs](../experiments/parallel-s15/report.md) | 393.19 s median; 393.69 s fresh cached | 20.55 GiB first use; 18.46 GiB cached |

Smaller supported configurations and their evidence are in the
[configuration catalogue](../pirconfig/README.md). s=15 passes exact monomial
enumeration/support discovery, the full P·G identity,
all systematic positions, sampled independently evaluated degree-t curves,
repeated canonical hashes and packed-output readback. Fresh cached output matches.
Preparation includes diagnostic GC/profiling and a manual sleep pause: its raw
elapsed time is **not an uninterrupted setup latency baseline**. Historical
forecasts and planner experiment snapshots retain their original inputs.

Implementation boundaries:

- [liftrs-shared](../encoding/liftrs-shared/README.md): common API, field/code
  metadata, reusable preprocessing, serialization and orchestration, principally Go.
- [liftrs-cpp](../encoding/liftrs-cpp/README.md): complete database-dependent native
  encoding and parallel transforms; `ScalingCPPEncoder.EncodeWithStats` is the
  complete path. `EncodeHybridWithStats` is the older comparison path.
- [liftrs-go](../encoding/liftrs-go/README.md): Go transform reference/fallback.
- Experimental full preparation remains bounded to s≤15. Native/minimal-file
  count handling now accepts s=16 with small-buffer/header validation only. Public
  `DiscoverOrbit` and ordinary basis/JSON limits remain s≤13. Supported catalogue
  entries identify standalone encoding, not application or cryptographic integration.
- [appref](../encoding/appref) and [cryptoref](../encoding/cryptoref) are small
  correctness references. The production file-based PIR API still has an RM
  testing placeholder; [rme_legacy](../encoding/rme_legacy) is archived reference
  material, not the scalable Lift-RS backend.

## First work on the new machine

1. **Establish the environment.** Record revision, CPU model/ISA, physical cores,
   NUMA topology, RAM and actual job/cgroup allocation, OS, compiler flags, storage
   filesystem/capacity and profiling permissions. Rebuild binaries for this host.
   Do not compare different machines as an isolated algorithm speedup.
2. **Validate portability and small cases.** Use generic C++17 `-O3` first, with
   CGO enabled and the native build tag. Keep Go/dense references. Record native
   profiler availability; unavailable privileged tools must not silently disable
   all CPU profiling. Validate byte-identical small outputs before any large run.
3. **Transfer and verify retained data, then reproduce cached s=15.** Separate
   loading, native initialization, encoding, validation and output I/O. Reuse the
   existing artifacts rather than spending hours rediscovering them by default.
4. **Audit s=16 before admission.** Start with the retained s=15 heap/CPU profiles
   and array lifetimes. The current schedule projects about **74.3 GiB of online
   arrays** at estimated K=919,440,829,N=2^32 (`16N+12K` bytes for four-byte values).
   This excludes allocator/runtime/OS overhead and is not a preprocessing/RSS
   bound. Compare buffer reuse/blocking with a larger-RAM baseline; do not assume
   the old proportional RSS forecasts are requirements.
5. **Validate the count/header extension and review budgets.** At
   s=16, N=2^32 cannot fit the original uint32 cardinalities/headers, although individual
   extension values still fit 32 bits. Audit domain sizes, lengths, indexes,
   shifts, native descriptions and file versions separately. Preserve the existing
   `LRSprobe1` fixtures/loaders; do not silently reinterpret their headers or widen
   every value. s≥17 also needs wider extension values. Add small-buffer boundary
   and reference tests before approving a profiled, guarded s=16 first-use run.

The [s=16 representation status](encoding-s16.md) now records wide count/header
handling, stored widths selected from s, and small-buffer boundary checks.
Full preprocessing/encoding feasibility and measured admission remain open.

Set new numerical RAM, scratch-storage, worker and time budgets from that audit
and the actual allocation. The laptop's 24 GiB Go/32 GiB RSS/20 GiB cache/5.5h
limits are historical, not new-node defaults. Retain preprocessing/profiles and
separate setup from database encoding. Do not start an unbounded full s=16 run.

## Toolchain and test entry points

Measured baseline: Go 1.27.3, Apple Clang 21, Python 3.9, C++17 with generic `-O3`,
16 native workers. `go.mod` declares 1.21.6; that is not evidence that every older
compiler/runtime matches the measured behavior. Use an installed compatible Go,
Python 3 and C++17 compiler (`clang++` or `g++`) on the new host; record versions.
The laptop's `/tmp/sk-depir-go-toolchain/...` path is not a repository dependency.

From the repository root, after setting local Go cache/scratch locations:

```sh
CGO_ENABLED=1 CXX=g++ go test -tags liftrs_cpp \
  ./encoding/liftrs-shared ./encoding/liftrs-cpp ./pirconfig ./planner
python3 experiments/parallel-s15/check.py --go go --run-name new-host-validation
```

The second command also runs race/vet and both full suites in an isolated copy,
protecting CSV fixtures from legacy tests. Choose a new result name; never
replace historical results. Build dependencies/commands are in the
[native README](../encoding/liftrs-cpp/README.md).

**Do not run the old large runner unchanged on Linux/HPC.**
[parallel-s15/run.py](../experiments/parallel-s15/run.py) contains macOS `pmset`,
`caffeinate` and `/usr/bin/sample` controls and fixed laptop budgets; its Go harness
also fixes s=15 and 16 workers. Adapt host/job monitoring, memory units, clocks,
profiling, paths and budgets in a new documented experiment. The live guard/pause
scripts are historical recovery utilities, not an HPC job launcher. Preserve
finalizing-before-terminal-status ordering so analysis waits for sealed evidence.

## Data to transfer separately from Git

A Git clone contains source, reports, profiles and manifests, **not large caches**.
The source machine repository is `/Users/tmour/Code/sk-DEPIR-AIF/sk-DEPIR`.
Preserve it and its caches until transfer/checksum verification succeeds.

Transfer these files from `.cache/liftrs/parallel-s15/s15-m2-t5/`, retaining their
relative paths on the destination:

- `support.bin`, `basis-metadata.json`: reusable discovered support.
- `online.bin`, `online-metadata.json`: minimal P/spectrum plan, about **4.81 GiB**.
- `g.bin`: generator for identity/reference work. All code files total **8.81 GiB**.
- `word-packed.bin`, `word-metadata.json`: optional retained s=15 output fixture,
  another **1.875 GiB**, useful for exact output comparisons.

The s=14 equivalents are under `.cache/liftrs/parallel-s14/s14-m2-t5/`.
Ignore cached executables as runnable assets on another OS/ISA; preserve matching
old binaries separately if needed to symbolize historical profiles. Native raw
sample locations are linked from the [archive manifest](../experiments/parallel-s15/monitor-sample-archive.json).
The [destination receipt](../message-board/2026-11-09-1316-codex-orient-001-completed.md)
now verifies transferred s=13/14/15 data. See [HPC setup](hpc-setup.md) for the
shared toolchain and Slurm GPU inventory. The [cached CPU baseline](../experiments/hpc-cached-s15/report.md)
now passes full suites and s=15 output checks on a compute node; GPU execution
and s≥16 remain unverified.

Verify the transferred s=15 files against the committed first-use manifest:

```sh
python3 - <<'PY'
import hashlib, json
from pathlib import Path
r = json.loads(Path('experiments/parallel-s15/initial/results.json').read_text())
cache = Path(r['cache_directory'])
for name, expected in r['cache_sha256'].items():
    h = hashlib.sha256()
    with (cache / name).open('rb') as f:
        for block in iter(lambda: f.read(1 << 20), b''):
            h.update(block)
    assert h.hexdigest() == expected, name
print('Transferred s=15 cache matches the recorded manifest')
PY
```

These checks establish byte identity with verified retained data, not independent
mathematical validity of arbitrary public files. The canonical s=15 word hash is
`5c09282e5cd2a77bd9541b5023a487f8deaf794a516e7e6a77007e7f9d08b265`.

## Agent coordination

Read this handoff, AGENTS.md, the latest message-board entries and the linked
contracts before editing. Record a started message on the new host, including its
environment and artifact-transfer status. Keep one file per message; document
all changes and experiments, check README, and retain reproducible source/binary
and artifact hashes. No environment-specific decision, new hardware backend or
larger-field feasibility claim is settled by this migration.
