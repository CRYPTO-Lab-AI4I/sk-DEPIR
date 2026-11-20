#!/usr/bin/env python3
"""Same-build GPU tree crossover with the bounded host cache control."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
CONTROL=runpy.run_path(str(HERE.parent/'ucx-registration-cache/run.py'))
COMPARE=runpy.run_path(str(HERE.parent/'mpi-buffer-comparison/run.py'))


def suite(context):
    CONTROL['configure'](context)
    return COMPARE['suite'](context)


if __name__=='__main__':
    if '--check'in sys.argv:print('explicit UCX_RCACHE_ENABLE=n; same-build staged/direct/direct/staged GPU tree; existing 2 HPC nodes / 2 H200 / 4 ranks / 4 CPUs / 8 GiB host / 15 min scheduler')
    else:COMPARE['TREE']['BASE']['main'](here=HERE,extra_suite=suite)
