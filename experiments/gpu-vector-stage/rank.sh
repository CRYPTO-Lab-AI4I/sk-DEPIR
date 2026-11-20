#!/bin/bash
set -euo pipefail
ulimit -f 32768
export GMON_OUT_PREFIX="${1}/profile-${SLURM_PROCID:?}"
root="$1"
shift
exec "$@" "${root}/rank-${SLURM_PROCID}"
