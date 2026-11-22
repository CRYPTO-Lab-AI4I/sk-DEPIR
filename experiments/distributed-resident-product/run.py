#!/usr/bin/env python3
"""Resident products using the shared three-FFT schedule and pointwise kernel."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
PRODUCT=runpy.run_path(str(HERE.parent/'distributed-gpu-product/run.py'))
VECTORS=PRODUCT['VECTORS']


def review(rows,directory):
    result=PRODUCT['review'](rows,directory,pointwise_words=lambda n,capacity:min(n//2,capacity,17))
    return resident_claims(rows,result)


def resident_claims(rows,result,ranks=2):
    """Review resident preservation/reuse independently of admitted sizes/tiling."""
    retained={}
    for r in result['measurements']:
        assert all(r.get(k)is True for k in ('resident','right_spectrum_preserved','reused'))
        key=(r['rank'],r['s'],r['dimension']);retained.setdefault(key,r['owned_device_bytes']);assert retained[key]==r['owned_device_bytes']
    for r in rows:
        if r.get('kind')=='pointwise-check':
            assert all(r.get(k)is True for k in ('resident','alias_rejected','layout_rejected','invalid_rejected','third_preserved'))
            assert r['checked_words']==(1<<r['dimension'])//ranks
    return result


def suite(c):
    c['run']('resident-product-review-tests',['python3','-B','experiments/distributed-resident-product/test_review.py'],30)
    return VECTORS['suite'](c,full_fft=True,review_fn=review,test_name='TestDistributedResidentProducts',reference_test='TestDistributedGPUProducts',reference_review=PRODUCT['review'],scope='bounded resident polynomial products; no fullsize/s18 preprocessing result')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch/720scontroller/15minscheduler;host-shard product regression+resident products normal/memcheck;no fulls18/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
