#!/usr/bin/env python3
"""Observe default MPI component selection without forcing a data transport."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
HOST=runpy.run_path(str(HERE.parent/'mpi-host-buffers/run.py'))


def suite(context):
    for key in ('OMPI_MCA_pml','OMPI_MCA_btl'):context['env'].pop(key,None)
    return HOST['suite'](context,transport={'OMPI_MCA_pml_base_verbose':'20','OMPI_MCA_btl_base_verbose':'40','UCX_LOG_LEVEL':'info'})


if __name__=='__main__':
    if '--check'in sys.argv:print('default component selection with PML/BTL/UCX logs; same host replay/limits; 2 HPC nodes / 4 CPUs / 8 GiB host / zero GPUs; no forced transport')
    else:HOST['BASE']['main'](here=HERE,extra_suite=suite,require_gpu=False)
