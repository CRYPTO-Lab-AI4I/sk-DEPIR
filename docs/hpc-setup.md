# HPC setup

The initial Linux setup uses Go 1.27.3 linux/amd64 and GCC 14.2.0, preserving
the Go version of the verified laptop checkpoint. This is environment setup,
not a larger-instance performance result. See the
[development handoff](development-handoff.md) for validation and admission gates.

## Shared and local Go toolchains

The toolchain is installed on shared storage at
`/hpc/home/crypto/tmour/sk-DEPIR/.cache/liftrs/toolchains/go1.27.3-linux-amd64/go`.
The initial `/tmp/sk-depir-go-toolchain/go` copy is also retained. Its archive came
from [the official Go downloads](https://go.dev/dl/), with filename
`go1.27.3.linux-amd64.tar.gz` and SHA-256
`65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40`,
verified against `https://go.dev/dl/?mode=json` before extraction.
The downloaded archive remains at `/tmp/sk-depir-go1.27.3.linux-amd64.tar.gz`.

From the repository root, for interactive use or compute jobs:

```sh
export PATH="$PWD/.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin:$PATH"
export GOTOOLCHAIN=local
export GOCACHE=/tmp/sk-depir-go-cache
export GOPATH="$PWD/.cache/liftrs/toolchains/gopath"
go version
```

Shared toolchain/dependency paths are accessible through the repository's NFS
mount; the `/tmp` build cache remains local and temporary. Jobs should choose
their own scratch/cache directories. No shell startup file or system package was changed.
`GOTOOLCHAIN=local` prevents implicit toolchain downloads; dependency downloads
are separate. The standard compiler/backend requirements remain in the root README.
Pinned module dependencies have been downloaded into the shared GOPATH; the
[cached s15 experiment](../experiments/hpc-cached-s15/report.md) runs offline with
`GOPROXY=off` and `GOSUMDB=off` against that prepared cache.

## Slurm and GPUs

Load the module in each shell or batch script that uses Slurm:

```sh
module load slurm/slurm/25.05
sinfo -N -o '%N %P %t %c %m %G'
squeue -u "$USER"
```

At the 2026-11-09 inventory, the controller reported:

| Partition | Nodes | Slurm CPUs per node | Configured memory per node (MiB) | GPU resources |
|---|---:|---:|---:|---|
| `aiq` | 10 | 144 | 2,088,816 | B200, usually 8 per node; one drained node lists 7 |
| `gpuq` | 17 | 128 | 928,610 | H200, usually 4 per node; one idle node lists 3 |

Both partitions report a maximum wall time of 4 days 4 hours and allow all
groups/accounts/QoS at the partition level. Node states and availability change;
query Slurm again before admission. Account associations, QoS limits and actual
GPU job admission were initially untested. The CPU-only baseline was admitted under
account `crypto_login`, QoS `quota_crypto_gpuq`. `defq` has no nodes in this inventory.

The interactive session is on `lnode02`; its CPU/RAM inventory is separate from
the compute nodes above. The [completed CPU baseline](../experiments/hpc-cached-s15/report.md)
runs on `gnode02` (Xeon 6710E), with 16 allocated cores. A 64 GiB memory request
did not configure a finite memory cgroup cap; the runner adds an explicit sampled
RSS guard. Both full suites and cached s=15 validation pass. Perf profiling is
denied; Go profiles and native stage counters remain retained.
At that initial setup no GPU was allocated. Subsequent
[tile trials](../experiments/gpu-transform-tiles/report.md) now validate one H200
NVL on gnode15, with 150,109,880,320 bytes reported device memory, driver
580.126.20, CUDA 12.9.86 and GCC 14.2.0. Forward/inverse tile outputs match the
native CPU exactly. The jobs request `--gres=gpu:1` and four CPU cores; the toolkit
module is `cuda12.9/toolkit/12.9.1`. CUDA remains an optional experiment dependency.
These are warmed component measurements including separate transfer/setup clocks,
not a working complete GPU encoder or consumer-GPU benchmark. The library
encoding implementation remains CPU-based; full s16 feasibility remains open.

## Retained reference data

The [transfer receipt](../message-board/2026-11-09-1316-codex-orient-001-completed.md)
records matching SHA-256 checks for s=13 preprocessing and all seven s=14/s=15
data/metadata files, including packed output fixtures. Keep these reusable caches
and historical profiles intact. Shared home storage is NFS; account for staging
and I/O separately from encoding in future experiments.
