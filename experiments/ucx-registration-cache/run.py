#!/usr/bin/env python3
"""Explicit UCX registration-cache diagnostic on the same host payload replay."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
HOST=runpy.run_path(str(HERE.parent/'mpi-host-buffers/run.py'))
UCX=Path('/cm/shared/apps/ucx/1.18.0/bin/ucx_info')
SETTINGS={'OMPI_MCA_pml':'ucx','OMPI_MCA_pml_base_verbose':'20','OMPI_MCA_btl_base_verbose':'40','UCX_LOG_LEVEL':'info','UCX_RCACHE_ENABLE':'n'}


def configure(context):
    context['env'].pop('OMPI_MCA_btl',None)
    context['env'].update(SETTINGS)
    context['run']('ucx-version',[UCX,'-v'])
    context['run']('ucx-config',[UCX,'-f'])
    assert 'UCX_RCACHE_ENABLE=n'in (context['out']/'ucx-config.log').read_text().splitlines()
    context['resources']['diagnostic_transport']=dict(SETTINGS)
    return SETTINGS


def suite(context):
    return HOST['suite'](context,transport=configure(context))


if __name__=='__main__':
    if '--check'in sys.argv:print('observed UCX PML, explicit UCX_RCACHE_ENABLE=n; same host replay/limits; 2 HPC nodes / 4 CPUs / 8 GiB host / zero GPUs; no default change')
    else:HOST['BASE']['main'](here=HERE,extra_suite=suite,require_gpu=False)
