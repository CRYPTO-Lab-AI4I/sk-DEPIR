#!/bin/bash
set -euo pipefail
ulimit -f 32768
profile="${1}-${SLURM_PROCID:?Slurm rank required}.pprof"
shift
exec "$@" "-test.cpuprofile=${profile}"
