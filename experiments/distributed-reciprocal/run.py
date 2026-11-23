#!/usr/bin/env python3
"""Shared oracle-free Newton orchestration with preserved checkpoint/native checks."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
RUNTIME=runpy.run_path(str(HERE.parent/'distributed-runtime/run.py'))
BIND=RUNTIME['BIND'];SCALE=RUNTIME['SCALE'];VECTORS=RUNTIME['VECTORS'];PROOF=RUNTIME['PROOF'];DIAG=RUNTIME['DIAG']


def dense_review(rows,directory):
    result=PROOF['PREFIX']['dense_review'](rows,directory)
    result['proof_recording']=PROOF['check_summaries'](rows,rows,'trace')
    result['support_verification']=BIND['support_claims'](rows,directory,None)
    return result


def suite(c):
    c['run']('reciprocal-review-tests',['python3','-B','experiments/distributed-reciprocal/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_RESTART_SUPPORT='coordinator',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('reciprocal-cpu',[]),('reciprocal-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestRankReciprocal|^TestRankFullCodeIdentity|^TestReciprocal','-count=1','-timeout=20s'],60)
    extra=[dict(name='dense',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_NEWTON_INPUT='dense',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded'),review=dense_review),dict(name='wide',test='TestWidePipelineTail',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_SPECTRUM='stream'),review=PROOF['trace_wide_review'])]
    result=SCALE['suite'](c,review_fn=BIND['review'],extra_arms=extra)
    arms=result['composed_vectors']['arms'];assert arms['cold']['artifacts']==arms['dense']['artifacts']
    result['diagnostic_writes']={mode:DIAG['diagnostic_review'](c['out'],mode,c['cache']/mode)for mode in ('cold','resume','dense','wide')}
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4))for mode in ('cold','resume','dense','wide')}
    c['resources'].update(reciprocal='Shared backend-independent Newton loop and full code identity checker;same square/product/checkpoint order',scope='Genuine s8 prefix cold/restart,default dense cold preservation,and bounded field18;not public production entry/fulls18')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;shared reciprocal cold/restart/dense/wide;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
