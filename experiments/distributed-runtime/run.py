#!/usr/bin/env python3
"""Validate the shared non-test MPI/GPU session with existing independent oracles."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
BIND=runpy.run_path(str(HERE.parent/'distributed-restart-binding/run.py'))
SCALE=BIND['SCALE'];VECTORS=BIND['VECTORS'];PROOF=BIND['PROOF'];DIAG=BIND['DIAG']


def suite(c):
    c['run']('runtime-review-tests',['python3','-B','experiments/distributed-runtime/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_RESTART_SUPPORT='coordinator',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    c['run']('ordinary-cpu-library-build',['env','CGO_ENABLED=0',c['go'],'build','./encoding/liftrs-shared'],60)
    c['run']('cuda-only-library-build',['env','CC=gcc','CXX=g++',c['go'],'build','-tags','liftrs_cuda','./encoding/liftrs-shared'],60)
    c['run']('mpi-library-build',[c['go'],'build','-tags','liftrs_mpi','./encoding/liftrs-shared'],60)
    c['run']('cuda-mpi-library-build',[c['go'],'build','-tags','liftrs_cuda,liftrs_mpi','./encoding/liftrs-shared'],60)
    # These tagged tests reject admission/state errors without opening MPI/CUDA.
    c['run']('session-state-tests',[c['go'],'test','-tags','liftrs_cuda,liftrs_mpi','./encoding/liftrs-shared','-run','^TestRankGPUSession','-count=1','-timeout=20s'],60)
    extra=[dict(name='wide',test='TestWidePipelineTail',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_SPECTRUM='stream',PIR_PIPELINE_NEWTON_INPUT='prefix'),review=PROOF['trace_wide_review'])]
    result=SCALE['suite'](c,review_fn=BIND['review'],extra_arms=extra)
    result['diagnostic_writes']={mode:DIAG['diagnostic_review'](c['out'],mode,c['cache']/mode)for mode in ('cold','resume','wide')}
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4))for mode in ('cold','resume','wide')}
    c['resources'].update(runtime='Shared MPI transport and error-returning GPU session outside testing;thin test adapters and independent native references',scope='Genuine s8 cold/coordinated restart and bounded field18 shared session;not production entry or fulls18')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;shared-runtime cold/restart/wide;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
