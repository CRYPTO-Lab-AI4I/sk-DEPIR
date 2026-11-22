#!/usr/bin/env python3
"""Admit canonical s18 arithmetic on bounded distributed GPU components."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
FFT=runpy.run_path(str(HERE.parent/'distributed-resident-fft/run.py'))
PRODUCT=runpy.run_path(str(HERE.parent/'distributed-gpu-product/run.py'))
RESIDENT=runpy.run_path(str(HERE.parent/'distributed-resident-product/run.py'))
VECTORS=PRODUCT['VECTORS']


def specs():
    return [dict(name=f'{kind}-{mode}',kind=kind,small=small,memcheck=small)
            for small,mode in ((False,'normal'),(True,'memcheck'))for kind in ('fft','product')]


def review(rows,directory,spec):
    cases=[(18,d)for d in ((5,)if spec['small']else(5,12))]
    if spec['kind']=='fft':result=FFT['review'](rows,directory,cases=cases)
    else:
        result=PRODUCT['review'](rows,directory,cases=cases,pointwise_words=lambda n,capacity:min(n//2,capacity,17))
        RESIDENT['resident_claims'](rows,result)
    assert all(r['s']==18 and r['local_taylor']is True for r in result['measurements'])
    return result


def suite(c):
    c['run']('s18-review-tests',['python3','-B','experiments/s18-field-admission/test_review.py'],30)
    c['run']('s18-tables',[c['go'],'test','./encoding/liftrs-shared','-run','^TestPreparationProductsS18Admission$|^TestPreprocessingArtifactsS18$','-count=1','-timeout=20s'],45)
    arms=[]
    for spec in specs():
        arms.append(dict(name=spec['name'],test='TestS18GPUFFT'if spec['kind']=='fft'else'TestS18GPUProducts',
            memcheck=spec['memcheck'],profile=not spec['small'],env=dict(PIR_S18_FIELD_SMALL='1'if spec['small']else'0'),
            review=lambda rows,directory,spec=spec:review(rows,directory,spec)))
    return VECTORS['suite'](c,full_fft=True,arms=arms,scope='canonical s18 field,bounded32/4096-element full-extension FFT and base-field products;small32-element sanitizer;not genuine/full s18 preprocessing')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2 HPC nodes/2H200/4CPU/8GiB RAM;128MiB owned/card,1GiB RSS/rank,128MiB cache/2GiB scratch/32MiB file;720s/15min;normal FFT/products then smallest sanitizer each;no fulls18/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
