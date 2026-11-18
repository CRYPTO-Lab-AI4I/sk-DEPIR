#!/usr/bin/env python3
"""Bounded independent GPU roots plus complete merge, with shared job guards."""
from pathlib import Path
import hashlib
import json
import runpy
HERE=Path(__file__).resolve().parent


def review_output(log,directory,cases):
    rows=[json.loads(line) for line in log.read_text().splitlines() if line.startswith('{')]
    values=[row for row in rows if row.get('kind')=='measurement']
    assert any(row.get('kind')=='completed' and row['cases']==cases for row in rows)
    for index in range(cases):
        reference=directory/('case-'+str(index)+'-reference.bin')
        digest=hashlib.sha256(reference.read_bytes()).hexdigest()
        for workers in (0,1,2,4):
            samples=[row for row in values if row['case']==index and row['independent_workers']==workers]
            assert len(samples)==3 and {r['sample'] for r in samples}=={-1,0,1}
            for row in samples:
                assert row['matches'] and row['sha256']==digest and row['seconds']>0
                assert row['owned_device_bytes']<=4<<30
                if workers:
                    assert sum(row['partition_sizes'])==row['support_count']
                    assert max(row['partition_sizes'])-min(row['partition_sizes'])<=2
                if row['sample']==1:
                    path=directory/row['file'];assert path.parent==directory
                    assert hashlib.sha256(path.read_bytes()).hexdigest()==digest
    return values


def suite(context):
    run=context['run'];source=context['source'];cache=context['cache'];out=context['out'];go=context['go'];env=context['env']
    binary=cache/'workers.test'
    run('worker-build',[go,'test','-c','-tags','liftrs_cpp,liftrs_cuda','-o',binary,'./encoding/liftrs-shared'],source)
    run('partition-race',[go,'test','-race','./encoding/liftrs-shared','-run','TestPreparationRootPartitions|TestRootPartitionAdmission','-count=1'],source)
    env.update(PIR_GPU_PREPARATION_LIBRARY=str(context['library']),PIR_ROOT_WORKER_OUTPUT=str(cache/'normal'))
    run('workers',[binary,'-test.run=^TestIndependentRootWorkers$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'workers.pprof')],source)
    measured=review_output(out/'workers.log',cache/'normal',4)
    env.update(PIR_ROOT_WORKER_SMALL='1',PIR_ROOT_WORKER_OUTPUT=str(cache/'sanitized'))
    run('workers-memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',binary,'-test.run=^TestIndependentRootWorkers$','-test.v','-test.timeout=10m'],source)
    sanitized=review_output(out/'workers-memcheck.log',cache/'sanitized',1)
    text=(out/'workers-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors'in text and 'LEAK SUMMARY: 0 bytes'in text
    (out/'workers.json').write_text(json.dumps(dict(measured=measured,sanitized=sanitized),indent=2)+'\n')


if __name__=='__main__':
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))['main'](here=HERE,extra_suite=suite)
