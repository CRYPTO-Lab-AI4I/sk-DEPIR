#!/usr/bin/env python3
"""Compose resident complete-vector kernels with the shared rank transpose."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
VECTORS=runpy.run_path(str(HERE.parent/'distributed-gpu-vectors/run.py'))


def review(rows,directory):
    result=VECTORS['review'](rows,directory)
    retained={}
    for r in result['measurements']:
        assert all(r.get(k)is True for k in ('resident','third_preserved','reused','range_rejection_preserved'))
        key=(r['rank'],r['case']);retained.setdefault(key,r['owned_device_bytes']);assert retained[key]==r['owned_device_bytes']
    return result


def suite(c):
    c['run']('resident-review-tests',['python3','-B','experiments/distributed-resident-vectors/test_review.py'],30)
    return VECTORS['suite'](c,review_fn=review,test_name='TestDistributedResidentVectors',baseline=True,scope='resident separable GPU vectors and shared rank transpose; Taylor/full resident FFT/product absent')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch/720scontroller/15minscheduler;old vector regression+resident normal/memcheck;no fulls18/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
