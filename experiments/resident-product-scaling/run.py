#!/usr/bin/env python3
"""Matched product samples, changing only backend and declared transfer tile."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
PRODUCT=runpy.run_path(str(HERE.parent/'distributed-gpu-product/run.py'))
RESIDENT=runpy.run_path(str(HERE.parent/'distributed-resident-product/run.py'))
VECTORS=PRODUCT['VECTORS']


def specs():
    out=[]
    for sample in range(3):
        for tile in (127,4096):
            for backend in (('host','resident')if sample%2==0 else('resident','host')):
                out.append(dict(name=f'sample-{sample}-{backend}-tile-{tile}',sample=sample,backend=backend,tile=tile,dims=[12,14],memcheck=False))
    for backend in ('host','resident'):
        out.append(dict(name=f'memcheck-{backend}-tile-4096',sample=None,backend=backend,tile=4096,dims=[12],memcheck=True))
    return out


def review(rows,directory,spec):
    count=lambda n,capacity:min(n//2,capacity,17)if spec['backend']=='resident'else min(capacity,17)
    result=PRODUCT['review'](rows,directory,cases=[(17,d)for d in spec['dims']],tile_words=spec['tile'],pointwise_words=count)
    resident=spec['backend']=='resident';retained={}
    if resident:RESIDENT['resident_claims'](rows,result)
    for r in result['measurements']:
        assert r['resident']is resident and r['tile_words']==spec['tile']and r['setup_seconds']>0
        assert r['reused']is True
        assert r['work_seconds']>0 and r['validation_capture_seconds']>=0
        assert abs(r['local_seconds']-r['local_capture_seconds']-r['local_work_seconds'])<1e-8
        assert r['local_work_seconds']>0 and r['local_capture_seconds']>=0
        if not resident:assert r['validation_capture_seconds']==0
        pair=[q for q in result['measurements']if q['case']==r['case']]
        assert len(pair)==2 and all(q['work_seconds']==r['work_seconds']and q['validation_capture_seconds']==r['validation_capture_seconds']for q in pair)
        assert r['work_seconds']==max(q['local_work_seconds']for q in pair)
        assert r['validation_capture_seconds']==max(q['local_capture_seconds']for q in pair)
        if resident:assert r['right_spectrum_preserved']is True
        key=(r['rank'],r['dimension']);retained.setdefault(key,r['owned_device_bytes']);assert retained[key]==r['owned_device_bytes']
    # Keep arm annotations outside producer rows so composition seals reconcile
    # with a fresh independent review, rather than changing logged evidence.
    result['sample_spec']=spec
    return result


def suite(c):
    c['run']('scaling-review-tests',['python3','-B','experiments/resident-product-scaling/test_review.py'],30)
    plan=specs();c['resources']['matched_plan']=plan
    arms=[]
    for spec in plan:
        def check(rows,directory,spec=spec):return review(rows,directory,spec)
        arms.append(dict(name=spec['name'],test='TestMatchedGPUProducts',review=check,memcheck=spec['memcheck'],profile=spec['sample']==0,
            env=dict(PIR_PRODUCT_TILE_WORDS=str(spec['tile']),PIR_PRODUCT_BACKEND=spec['backend'],PIR_PRODUCT_DIMS=','.join(map(str,spec['dims'])))))
    return VECTORS['suite'](c,full_fft=True,arms=arms,scope='matched synthetic s17-field products N4096/16384,3 samples/backend/tile; small sanitizer subset only; no fulls17/s18 preprocessing result')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch/720s/15min;N4096/16384,s17field,host/resident,127/4096-word tiles,3 samples and N4096 sanitizer subset;no admission/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
