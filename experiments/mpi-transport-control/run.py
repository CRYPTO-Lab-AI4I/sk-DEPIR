#!/usr/bin/env python3
"""Explicit TCP-only MPI transport diagnostic, not a production memory mode."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
HOST=runpy.run_path(str(HERE.parent/'mpi-host-buffers/run.py'))


def suite(context):
    return HOST['suite'](context,transport={'OMPI_MCA_pml':'ob1','OMPI_MCA_btl':'self,tcp'})


if __name__=='__main__':
    if '--check'in sys.argv:print('explicit MPI ob1/self,tcp; same captured host replay and limits; 2 HPC nodes / 4 CPUs / 8 GiB host / zero GPUs; no default change')
    else:HOST['BASE']['main'](here=HERE,extra_suite=suite,require_gpu=False)
