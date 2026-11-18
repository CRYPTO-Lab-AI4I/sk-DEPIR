#!/bin/bash
set -euo pipefail
# Bound rank-generated profile files on remote nodes too (at most32MiB).
ulimit -f 32768
# Rank-specific profile names cannot collide when nodes reuse local PIDs.
export GMON_OUT_PREFIX="${1}-${SLURM_PROCID:?Slurm rank required}"
shift
exec "$@"
