#!/usr/bin/env python3
"""Four MPI workers sharing two GPUs: bounded s18 collective validation."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
FIELD=runpy.run_path(str(HERE.parent/'s18-field-admission/run.py'))
VECTORS=FIELD['VECTORS']


def review(rows,directory,spec):
    cases=[(18,d)for d in ((5,)if spec['small']else(5,12))]
    if spec['kind']=='fft':result=FIELD['FFT']['review'](rows,directory,cases=cases,ranks=4)
    else:
        result=FIELD['PRODUCT']['review'](rows,directory,cases=cases,ranks=4,pointwise_words=lambda n,capacity:min(n//4,capacity,17))
        FIELD['RESIDENT']['resident_claims'](rows,result,ranks=4)
    hosts=[r for r in rows if r.get('kind')=='host']
    for host in {r['host']for r in hosts}:
        selected=[r for r in hosts if r['host']==host]
        assert len(selected)==2 and len({r['visible_devices']for r in selected})==1
    assert all(r['s']==18 and r['local_taylor']is True and 0<r['owned_device_bytes']<=64<<20 for r in result['measurements'])
    return result


def suite(c):
    c['run']('four-worker-review-tests',['python3','-B','experiments/distributed-four-worker/test_review.py'],30)
    arms=[]
    for spec in FIELD['specs']():
        arms.append(dict(name=spec['name'],test='TestS18GPUFFT'if spec['kind']=='fft'else'TestS18GPUProducts',
            memcheck=spec['memcheck'],profile=not spec['small'],env=dict(PIR_S18_FIELD_RANKS='4',PIR_S18_FIELD_SMALL='1'if spec['small']else'0'),
            review=lambda rows,directory,spec=spec:review(rows,directory,spec)))
    return VECTORS['suite'](c,full_fft=True,arms=arms,ranks=4,scope='four MPI workers sharing2H200s,bounded canonical s18 FFT/products,small memchecks;not four-GPU throughput/full root tree/s18 preprocessing')


if __name__=='__main__':
    if '--check'in sys.argv:print('2 HPC nodes/2H200/4CPU/8GiB RAM;4workers/1core each,2workers/card,64MiBowned/worker=128MiB/card aggregate,1GiBRSS/worker;128MiBcache/2GiBscratch/32MiBfile,720s/15min;no hardware increase/default/fullsize claim;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
