#!/bin/bash
set -euo pipefail
log_base="$1-${SLURM_PROCID:?Slurm rank required}"
shift
exec "$@" --profile="${log_base}.pprof" >"${log_base}.stdout.log" 2>"${log_base}.stderr.log"
