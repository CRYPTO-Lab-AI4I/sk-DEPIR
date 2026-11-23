#!/usr/bin/env python3
"""Explicit committed final output around the reference-independent runtime."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
WORKFLOW=runpy.run_path(str(HERE.parent/'distributed-workflow/run.py'))
BIND=WORKFLOW['BIND'];SCALE=WORKFLOW['SCALE'];VECTORS=WORKFLOW['VECTORS'];PROOF=WORKFLOW['PROOF'];DIAG=WORKFLOW['DIAG']


def assembled_review(rows,directory,baseline_rows,baseline_directory):
    return WORKFLOW['assembled_review'](rows,directory,baseline_rows,baseline_directory,require_publication=True)


def suite(c,extra_arms=None):
    c['run']('publication-review-tests',['python3','-B','experiments/distributed-publication-commit/test_review.py'],30)
    c['env']['PIR_WORKFLOW_PUBLICATION']='commit'
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('publication-cpu',[]),('publication-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestWorkflowPublication','-count=1','-timeout=20s'],60)
    result=WORKFLOW['suite'](c)if extra_arms is None else WORKFLOW['suite'](c,extra_arms=extra_arms)
    arm=result['composed_vectors']['arms']['assembled']
    assert arm['publication']is not None
    c['resources']['publication']='Explicit streamed-hash staged payloads and last-written synced manifest;not production restart'
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2nodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;pilot cold/restart+committed assembled output;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
