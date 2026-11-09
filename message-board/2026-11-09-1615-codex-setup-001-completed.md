# Go installed; Slurm GPU inventory recorded

- Timestamp: 2026-11-09 16:15:42 UTC
- Agent: Codex
- Work ID: SETUP-001
- Status: completed

Installed official Go 1.27.3 linux/amd64 at
`/tmp/sk-depir-go-toolchain/go`, matching the historical Go version. Downloaded
`https://go.dev/dl/go1.27.3.linux-amd64.tar.gz` with curl and verified SHA-256
`65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40`
against the official `https://go.dev/dl/?mode=json` before extracting with tar.
Archive and release metadata remain under `/tmp/sk-depir-*`. No system package,
startup file or existing cache was changed. This local temporary installation
must be made available to compute nodes before Slurm execution.

Verification: `go version` reports `go1.27.3 linux/amd64`; `go env` reports cgo
enabled and GCC/g++ defaults. This focused command passes:

```sh
env GOTOOLCHAIN=local GOCACHE=/tmp/sk-depir-go-cache \
 GOPATH=/tmp/sk-depir-gopath GOMAXPROCS=2 CGO_ENABLED=1 CXX=g++ \
 /tmp/sk-depir-go-toolchain/go/bin/go test -p 2 -tags liftrs_cpp \
 ./encoding/liftrs-shared \
 -run '^(ExampleScalingOnlinePlan_WithCPPMatrix|TestCPPMatrixEncoding)$' \
 -count=1 -timeout=120s
```

The native matrix example and encoding tests compare complete small outputs
with the Go encoder and, for tiny configurations, the dense oracle. This is a
focused Linux smoke check, not the full portability suite or an s=15 rerun.

Corrected the orientation inventory's omission of scheduler-visible GPUs:
`module load slurm/slurm/25.05` succeeds. Queried `sinfo`, `squeue -u tmour`
and `scontrol show partition aiq/gpuq` after sandbox network escalation.
`aiq` reports 10 nodes with 144 Slurm CPUs and 2,088,816 MiB/node, usually
8 B200 GPUs (one drained node lists 7). `gpuq` reports 17 nodes with 128
CPUs and 928,610 MiB/node, usually 4 H200 GPUs (one idle node lists 3).
Both list a 4-day-4-hour maximum wall time and open partition-level access.
The user had no queued jobs in this snapshot. This does not establish account/QoS
admission, actual GPU memory, driver compatibility or compute-node performance.
No GPU allocation/job or benchmark was launched; the encoder remains CPU-based.

Added [HPC setup](../docs/hpc-setup.md) with toolchain environment commands,
temporary/node-local path limitations, Slurm commands and resource inventory.
Updated [root README](../README.md), [handoff](../docs/development-handoff.md)
and [scaling study](../docs/encoding-scaling.md) to link current setup and verified
transfer status, preserving historical measurement reports. README check complete:
no new supported parameters, GPU implementation or performance claim is added.
Both SETUP-001 messages and the board index track this work. `git diff --check`
and changed-document relative-link checks pass.

Next: choose persistent/shared toolchain placement, verify an actual compute
allocation and Linux profiling, run the remaining portability checks, then
reproduce cached s=15 under reviewed budgets. The s=16 audit remains required.
