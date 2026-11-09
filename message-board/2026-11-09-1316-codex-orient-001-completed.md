# HPC repository orientation complete

- Timestamp: 2026-11-09 13:16:00 UTC
- Agent: Codex
- Work ID: ORIENT-001
- Status: completed
- Starting revision: `0f26303` (implementation/results checkpoint `d721182`)

## Scope and understanding

Reviewed the board index and historical completion summaries, recent scaling and
handoff messages, root and encoding READMEs, catalogue/planner/reference READMEs,
[roadmap](../ROADMAP.md), [handoff](../docs/development-handoff.md), representation,
encoder and protocol-I/O contracts, algorithm/scaling documentation and relevant
experiment evidence. Inspected the shared minimal loader, native matrix adapter,
preparation code, field adapter, native import guards and legacy PIR encoder.
This was orientation, not a complete code/security audit or a new experiment.

Current verified scope is standalone mathematical encoding through s=15,m=2,t=5:
K=216,914,047 and N=2^30. Public ordinary discovery remains s≤13. Go handles
discovery/reusable preprocessing/loading; the opt-in matrix backend performs all
database-dependent encoding in C++. The cyclic formulation preserves the paper's
omega-power information set and canonical coordinates. Small dense, cryptographic
and application references remain correctness oracles. Large application execution,
cryptographic integration, grouped storage and TRL 5 acceptance remain open; the
legacy file-based API still uses its RM testing placeholder. Storage remains ordinary
I/O, with no server-side protocol computation in scope.

At s=16, N=2^32 requires cardinality/header/native-description changes even though
individual packed extension elements still fit uint32. In particular,
`orbitField.q`, `q*q` and the minimal file's uint32 N header cannot represent that
cardinality. This observation confirms the existing handoff gate, not a complete
width audit. The handoff's approximately 74.3 GiB online-array scenario is an
estimate, not a measured preprocessing/RSS bound. Preserve old `LRSprobe1` data.

## Transferred data receipt

The ignored `.cache/liftrs/` contains s=13 JSON preprocessing; s=14 and s=15
support, P/spectrum online plans, generators, metadata and packed output fixtures;
small s=6 fixtures; and earlier calibration/scaling assets and native samples.
No cache or profile was changed, removed, loaded into an encoder or executed.

Read each file in 8 MiB blocks using Python `hashlib.sha256`, comparing:

- s=13 `cyclic.json` with `.cache/liftrs/s13-m2-t5/metadata.json`: match
  (`a2a3bcd5913d821c99d08ceca05d6c5d73ce2abe6e3987910c0dcf5dfe97eb22`).
- Every present entry of `cache_sha256` from
  [s=14 initial results](../experiments/parallel-s14/initial/results.json), resolved
  beneath `.cache/liftrs/parallel-s14/s14-m2-t5/`: all seven data/metadata files match.
- Every entry of `cache_sha256` from
  [s=15 initial results](../experiments/parallel-s15/initial/results.json), resolved
  beneath `.cache/liftrs/parallel-s15/s15-m2-t5/`: all seven files match.

s=14's recorded `builds/ca3cbb8174e7704b/probe` and its `provenance.json` are absent
from that cache. These are historical build assets, not missing preprocessing;
new-host executables must be rebuilt. Other transferred fixtures/profiles were
inventoried, not exhaustively checksum-verified. Matching hashes establishes
transfer byte identity against retained evidence, not new independent mathematical
validation. Packed-file digests are distinct from canonical uint32-word digests.

## Host observations and verification limits

Read-only commands: `git status --short`, `git log -4 --oneline`, documentation/source
reads with `rg`, `cat` and `sed`, cache inventory with `find`, Python checksum loop,
`hostname`, `lscpu`, `free -h`, `df -h . /tmp`, compiler/path/module inspection and
`/proc`/cgroup reads. Initial working tree was clean.

- Session host `lnode02`: Ubuntu 24.04.4 LTS, x86_64, two Xeon Silver 4516Y+
  sockets, 48 physical cores / 96 logical CPUs, two NUMA nodes; `free -h` reports
  503 GiB total RAM. This is visible host capacity, not a reserved job allocation.
- Session CPU affinity includes 0–95 and memory nodes 0–1. Inspected session and
  ancestor cgroups show no finite CPU quota or memory.max/high limit. This is not
  scheduler authorization or a resource reservation.
- Repository is on NFS; `/tmp` is on a local NVMe filesystem with about 404 GiB
  available at inspection. Do not assume home I/O has local-scratch performance.
- GCC/g++ 14.2.0 is on PATH. Python 3, module tooling, perf and numactl are present.
  Slurm srun/sbatch binaries exist under `/cm/local/apps/slurm/25.05/bin/`, but
  are not on this session's PATH. No job was submitted.
- `go` was absent from PATH; the visible module listing had no Go entry. No
  toolchain was installed or downloaded, and no build/test was attempted.
- `perf_event_paranoid=4`; profiler permission still needs checking. No native
  profiler test or GPU availability validation was performed.

README check: read the root README and relevant package READMEs. No API, support,
setup command or performance claim was changed. Existing pre-transfer statements
in README/handoff/scaling documentation are superseded for this destination by
the verified receipt above; several older status paragraphs also precede s=15.
Documentation reconciliation can accompany the next setup work item, retaining
historical experiment provenance. Only this work item's two messages and the
board index changed. `git diff --check` and new-message relative-link checks pass.

Next: establish a compatible Go toolchain and actual compute allocation/budgets;
validate small Go/native outputs and profiling on Linux; adapt the macOS-specific
runner into a documented HPC experiment; reproduce cached s=15; then finish the
s=16 representation/memory audit before admitting a large run. No large discovery,
preparation or encoding run was started, and no s≥16 feasibility is claimed.
