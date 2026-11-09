#!/bin/bash
#SBATCH --job-name=depir-width
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem=64G
#SBATCH --time=01:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
export GOTOOLCHAIN=local
export GOPATH="$PWD/.cache/liftrs/toolchains/gopath"
export GOCACHE="/tmp/sk-depir-width-${SLURM_JOB_ID}-go-cache"
export GOPROXY=off GOSUMDB=off GOMAXPROCS=4 CGO_ENABLED=1 CXX=g++
GO="$PWD/.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go"
# Small full-width buffers only. -asan instruments Go/cgo; UBSan catches shifts
# and arithmetic in C++, with immediate failure rather than recovered diagnostics.
CGO_CXXFLAGS='-fsanitize=undefined -fno-sanitize-recover=all' \
 CGO_LDFLAGS=-fsanitize=undefined "$GO" test -p 4 -asan -tags liftrs_cpp \
 ./encoding/liftrs-shared -run 'TestOrbitCountBoundary|TestScalingOnlineCountWidths|TestCPPCountBoundary|TestScalingCPPLargeFields|ExampleWriteScalingOnline' \
 -count=1 -timeout=5m
# Reuse the existing bounded baseline runner for both full suites, race/vet and
# a cached s15 encoding. It stages inputs and owns an explicit sampled RSS guard.
python3 experiments/hpc-cached-s15/run.py "$1"
