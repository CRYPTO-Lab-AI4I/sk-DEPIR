#!/usr/bin/env python3
"""Rerun only the bounded wide arm missing from job81744; no cold/restart repeats."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent;H=runpy.run_path(str(HERE/'run.py'))

def suite(c):
    c['run']('diagnostic-review-tests',['python3','-B','experiments/distributed-diagnostic-writes/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream')
    arm=next(a for a in H['comparison_arms']()if a['name']=='wide')
    result=H['VECTORS']['suite'](c,full_fft=False,arms=[arm],ranks=4,scope='Targeted missing bounded diagnostic field18 arm;prior cold/restart preserved from81744;not fulls18')
    result['targeted_wide']=True
    result['diagnostic_writes']={'wide':H['diagnostic_review'](c['out'],'wide',c['cache']/'wide')}
    rows=H['VECTORS']['BASE']['parse_rank_logs'](c['out']/'wide.log',ranks=4)
    result['native_taylor_totals']={'wide':H['PROOF']['native_review'](c['out'],'wide',rows)}
    return result

if __name__=='__main__':
    if '--check'in sys.argv:print('Targeted bounded wide arm only;same2nodes/2H200/4CPU/8GiBhost/64MiBownedGPU/worker/1GiBRSS/worker/128MiBcache/2GiBscratch/720scontroller/15minSlurm;no submission')
    else:H['VECTORS']['BASE']['main'](here=HERE,extra_suite=suite)
