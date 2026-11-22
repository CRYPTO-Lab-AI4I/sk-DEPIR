#!/usr/bin/env python3
"""Complete four-worker cold/restart pilot using maintained shared algorithms."""
from pathlib import Path
import json
import runpy
import sys
HERE=Path(__file__).resolve().parent
COORD=runpy.run_path(str(HERE.parent/'distributed-coordinator-stream/run.py'))
RESTART=COORD['RESTART'];DISCOVERY=RESTART['DISCOVERY'];PILOT=RESTART['PILOT']
TREE=runpy.run_path(str(HERE.parent/'distributed-root-hierarchy/run.py'))
VECTORS=RESTART['VECTORS']


def cold_review(rows,directory):
    discovered=[r for r in rows if r.get('discovery')is True]
    core=[r for r in rows if r.get('discovery')is not True and 'tree_case'not in r]
    result=PILOT['review'](core,directory,streamed=True,ranks=4,hierarchy=True)
    result.update(DISCOVERY['review_discovery'](discovered,directory,[4,6],ranks=4))
    cases=[]
    for s in (4,6):
        values=[r for r in result['measurements']if r['s']==s]
        counts=values[0]['root_counts'];k=values[0]['k']
        assert k=={4:97,6:1910}[s]
        info=result['discovery'][f's{s}']
        assert counts==[r['root_count']+1 for r in info]
        cases.append((s,k,f's{s}',counts))
    kinds={'host','root-tree','root-reference','root-group-closed','root-final','hierarchy-completed'}
    tree_rows=[r for r in rows if 'tree_case'in r or r.get('kind')in kinds]
    result['hierarchy']=TREE['review'](tree_rows,directory,cases=cases)
    return result


def review(rows,directory,resume=False,source=None,cold=None):
    # The restart reviewer owns common PID/prefix/known-artifact bindings;only
    # cold composition differs from the legacy two-rank final-root merge.
    result=RESTART['review'](rows,directory,resume,source,cold,ranks=4,hierarchy=True,core_review=cold_review)
    assert all(r['discovery_layout']=='distributed'and r['reference_mode']=='retained'and r['support_streamed']is True for r in result['measurements'])
    assert all(0<r['owned_device_bytes']<=64<<20 for r in result['measurements'])
    if resume:assert not any('tree_case'in r or r.get('kind')in('root-tree','root-reshape')for r in rows)
    return result


def suite(c,extra_arms=None):
    c['run']('pipeline-four-review-tests',['python3','-B','experiments/distributed-pipeline-four/test_review.py'],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    c['run']('control-admission',[c['go'],'test','-tags','liftrs_mpi','./experiments/distributed-root/mpi','-count=1','-timeout=20s'],45)
    manifest=json.loads((c['source']/'experiments/distributed-coordinator-stream/reference-inputs.json').read_text())
    target=c['cache']/'reference-inputs'
    total=COORD['copy_inputs'](VECTORS['BASE']['ROOT'],manifest,target)
    c['resources'].update(reference_input_bytes=total,reference_strategy='retained native files and streamed publication;bounded replicated hierarchy support diagnostics')
    c['save']()
    c['run']('reference-preflight',['srun','--nodes=2','--ntasks=4','--ntasks-per-node=2','--cpus-per-task=1','--export=ALL,TMPDIR=/tmp','--chdir='+str(c['cache']),
        'bash','-c','set -eu; for p in "$1"/*; do test -r "$p"; done','reference-preflight',target],30)
    return RESTART['suite'](c,reference_from=target,review_fn=review,ranks=4,extra_arms=extra_arms)


if __name__=='__main__':
    if '--check'in sys.argv:print('same2 HPC nodes/2H200/4CPU/8GiB requested RAM;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker;128MiBcache/2GiBscratch/32MiBfile,720s/15min;genuine s4/s6 complete cold/fresh-prefix8 resume,shared hierarchy and streamed publication;no CUDA arithmetic/default/resource increase;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
