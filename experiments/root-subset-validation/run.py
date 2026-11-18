#!/usr/bin/env python3
"""Explicit bitmap/search root-worker comparison; ordinary defaults unchanged."""
from pathlib import Path
import json
import runpy
HERE=Path(__file__).resolve().parent


def validate_pair(records):
    assert [(r['mode'],r['repeat']) for r in records]==[('bitmap',0),('sorted',0),('sorted',1),('bitmap',1)]
    digests={}
    for record in records:
        for row in record['rows']:
            assert row['validation']==record['mode'] and row['reference_validation']=='sorted'
            key=row['case']
            if key in digests:assert row['sha256']==digests[key]
            else:digests[key]=row['sha256']
    assert set(digests)==set(range(4))


def suite(context):
    run=context['run'];source=context['source'];cache=context['cache'];out=context['out'];go=context['go'];env=context['env']
    verify=runpy.run_path(str(HERE.parent/'parallel-root-workers/run.py'))['review_output']
    binary=cache/'subset.test'
    tests='TestSupportMembershipEquivalence|TestSortedSupportWideMemoryAndProgress|TestPreparationRootPartitions|TestRootPartitionAdmission'
    run('membership-tests',[go,'test','./encoding/liftrs-shared','-run',tests,'-v','-count=1'],source)
    run('membership-race',[go,'test','-race','./encoding/liftrs-shared','-run',tests,'-count=1'],source)
    run('subset-build',[go,'test','-c','-tags','liftrs_cpp,liftrs_cuda','-o',binary,'./encoding/liftrs-shared'],source)
    env['PIR_GPU_PREPARATION_LIBRARY']=str(context['library'])
    records=[]
    for mode,repeat in [('bitmap',0),('sorted',0),('sorted',1),('bitmap',1)]:
        label=mode+'-'+str(repeat)
        destination=cache/label;env['PIR_ROOT_WORKER_OUTPUT']=str(destination)
        run(label,[binary,'-test.run=^TestIndependentRootWorkers$','-test.v','-test.timeout=10m',
            '-root-validation='+mode,'-root-reference-validation=sorted','-test.cpuprofile='+str(out/(label+'.pprof'))],source)
        rows=verify(out/(label+'.log'),destination,4)
        records.append(dict(mode=mode,repeat=repeat,rows=rows))
    validate_pair(records)
    sanitized=[]
    env['PIR_ROOT_WORKER_SMALL']='1'
    for mode in ('bitmap','sorted'):
        label=mode+'-memcheck';env['PIR_ROOT_WORKER_OUTPUT']=str(cache/label)
        run(label,['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',binary,
            '-test.run=^TestIndependentRootWorkers$','-test.v','-test.timeout=10m','-root-validation='+mode,'-root-reference-validation=sorted'],source)
        rows=verify(out/(label+'.log'),cache/label,1)
        assert all(r['validation']==mode and r['reference_validation']=='sorted' for r in rows)
        text=(out/(label+'.log')).read_text()
        assert 'ERROR SUMMARY: 0 errors'in text and 'LEAK SUMMARY: 0 bytes'in text
        sanitized.append(dict(mode=mode,rows=rows))
    (out/'comparison.json').write_text(json.dumps(dict(records=records,sanitized=sanitized),indent=2)+'\n')


if __name__=='__main__':
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))['main'](here=HERE,extra_suite=suite)
