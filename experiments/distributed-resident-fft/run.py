#!/usr/bin/env python3
"""Complete resident FFT through shared Taylor, vector and transpose schedules."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
FFT=runpy.run_path(str(HERE.parent/'distributed-gpu-fft/run.py'))
VECTORS=FFT['VECTORS']


def review(rows,directory,cases=None):
    result=FFT['review'](rows,directory,cases=cases)
    retained={}
    for r in result['measurements']:
        assert all(r.get(k)is True for k in ('resident','third_preserved','reused','range_rejection_preserved'))
        key=(r['rank'],r['case']);retained.setdefault(key,r['owned_device_bytes']);assert retained[key]==r['owned_device_bytes']
    return result


def suite(c):
    c['run']('resident-fft-review-tests',['python3','-B','experiments/distributed-resident-fft/test_review.py'],30)
    return VECTORS['suite'](c,full_fft=True,review_fn=review,test_name='TestDistributedResidentFFT',reference_review=FFT['review'],scope='complete resident bounded FFT with shared Taylor storage schedule; no resident product or fulls18 result')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch/720scontroller/15minscheduler;host-tile full FFT regression+complete resident normal/memcheck FFT;no fulls18/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
