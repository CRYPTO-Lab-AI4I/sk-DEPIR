#!/usr/bin/env python3
"""Probe bounded MPI-owned staging with proof checks on both payload widths."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
TREE=runpy.run_path(str(HERE.parent/'root-merge-tree/run.py'))


def suite(context):
    context['env'].update(PIR_ROOT_TREE_DIAGNOSTICS='1',PIR_ROOT_TREE_REPEATS='12',PIR_ROOT_TREE_MPI_BUFFER='staged')
    context['resources'].update(mpi_buffer_mode='staged',staging_bytes_per_rank=512*1024,staging_bytes_aggregate=2*1024*1024)
    result=TREE['suite'](context)
    for field,expected in (('measurements',512*1024),('sanitized',8*1024)):
        hosts=[r for r in result[field]if r.get('kind')=='host']
        assert len(hosts)==4 and all(r['buffer_mode']=='staged'and r['staging_bytes']==expected for r in hosts)
        runpy.run_path(str(HERE.parent/'root-transfer-diagnostics/analyze.py'))['review_proofs'](result[field])
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('2 HPC nodes / 2 H200 / 4 logical ranks / 4 CPUs / 8 GiB host; 512 KiB MPI-owned staging per rank; explicit staged mode; runtime penalty unknown')
    else:TREE['BASE']['main'](here=HERE,extra_suite=suite)
