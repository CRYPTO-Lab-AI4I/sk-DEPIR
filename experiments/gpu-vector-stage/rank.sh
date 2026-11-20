#!/bin/bash
set -euo pipefail
ulimit -f 32768
export GMON_OUT_PREFIX="${1}/profile-${SLURM_PROCID:?}"
if [[ "${PIR_VECTOR_MEMCHECK:-0}" == 1 ]]; then
  export PIR_VECTOR_MEMCHECK_AXIS="${SLURM_PROCID}"
fi
root="$1"
shift
exec "$@" "${root}/rank-${SLURM_PROCID}"
