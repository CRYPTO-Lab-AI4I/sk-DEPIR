#!/usr/bin/env python3
"""Discovery transforms through the shared GPU session;sample semantics unchanged."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
BIND=runpy.run_path(str(HERE.parent/'distributed-restart-binding/run.py'))
SCALE=BIND['SCALE'];VECTORS=BIND['VECTORS'];PROOF=BIND['PROOF'];DIAG=BIND['DIAG']


def resident_review(rows,directory):
    prefix=PROOF['PREFIX']
    result=prefix['input_claims'](prefix['DISCOVERY']['resident_review'](rows,directory),'prefix')
    result['proof_recording']=PROOF['check_summaries'](rows,rows,'trace')
    result['support_verification']=BIND['support_claims'](rows,directory,None)
    return result


def suite(c):
    c['run']('discovery-loop-review-tests',['python3','-B','experiments/distributed-discovery-loop/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_RESTART_SUPPORT='coordinator',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    c['run']('discovery-runtime-state-tests',[c['go'],'test','-tags','liftrs_cuda,liftrs_mpi','./encoding/liftrs-shared','-run','^TestRankDiscovery|^TestRankGPUDiscovery|^TestRankGPUSession','-count=1','-timeout=20s'],60)
    extra=[dict(name='resident-discovery',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_DISCOVERY_RESULT='resident',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded'),review=resident_review)]
    result=SCALE['suite'](c,review_fn=BIND['review'],extra_arms=extra)
    arms=result['composed_vectors']['arms'];assert arms['cold']['artifacts']==arms['resident-discovery']['artifacts']
    modes=('cold','resume','resident-discovery')
    result['diagnostic_writes']={mode:DIAG['diagnostic_review'](c['out'],mode,c['cache']/mode)for mode in modes}
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4))for mode in modes}
    c['resources'].update(discovery_runtime='Explicit-policy error-returning sampling/union/routing/archive runtime;seed remains only in test adapter',scope='Genuine s8 streamed cold/coordinated restart and resident discovery preservation;not fulls18/public production entry')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2nodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;cold/restart/resident-discovery;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
