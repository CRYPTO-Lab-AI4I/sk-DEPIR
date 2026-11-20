#!/bin/bash
set -euo pipefail
ulimit -f 32768
profile="${1}-${SLURM_PROCID:?Slurm rank required}.pprof"
log_base="$1-${SLURM_PROCID}"
shift
# Keep worker streams separate from Slurm/launcher diagnostics. This is explicit
# so historical host-only runners retain their original output convention.
if [[ "${PIR_RANK_LOG_FILES:-0}" == 1 ]]; then
  exec "$@" "-test.cpuprofile=${profile}" >"${log_base}.stdout.log" 2>"${log_base}.stderr.log"
fi
exec "$@" "-test.cpuprofile=${profile}"
