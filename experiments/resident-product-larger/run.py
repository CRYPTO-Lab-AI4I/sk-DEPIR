#!/usr/bin/env python3
"""Larger GPU products with an exact streaming convolution oracle."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
PRODUCT=runpy.run_path(str(HERE.parent/'distributed-gpu-product/run.py'))
RESIDENT=runpy.run_path(str(HERE.parent/'distributed-resident-product/run.py'))
VECTORS=PRODUCT['VECTORS']


def specs():
    out=[]
    for backend in ('host','resident-local'):
        out.append(dict(name=f'random-reference-{backend}',backend=backend,dim=14,pattern='random',shapes=[0,1],sample=None,memcheck=False))
        out.append(dict(name=f'ones-memcheck-{backend}',backend=backend,dim=12,pattern='ones',shapes=[0,1],sample=None,memcheck=True))
    for sample in range(3):
        for dim in (14,16,18):
            for backend in (('host','resident-local')if sample%2==0 else('resident-local','host')):
                out.append(dict(name=f'sample-{sample}-d{dim}-{backend}',backend=backend,dim=dim,pattern='ones',shapes=[0],sample=sample,memcheck=False))
    return out


def prefix_oracle(n,ac,bc,rank):
    # Independent of the producer's streaming XOR recurrence: build a prefix
    # sequence and take two clipped endpoints for every owned coefficient.
    prefix=[0]*(ac+1);mask=(1<<17)-1
    for j in range(ac):
        a=1 if j in (0,ac-1)else(((j+1)*0x9e3779b9)^0x1337)&mask
        prefix[j+1]=prefix[j]^a
    first=rank*n//2
    return b''.join((prefix[min(k+1,ac)]^prefix[min(ac,max(0,k-bc+1))]).to_bytes(4,'little')for k in range(first,first+n//2))


def review(rows,directory,spec):
    resident=spec['backend']=='resident-local';n=1<<spec['dim']
    result=PRODUCT['review'](rows,directory,cases=[(17,spec['dim'])],tile_words=4096,shapes=spec['shapes'],pointwise_words=lambda n,capacity:min(n//2,capacity,17)if resident else min(capacity,17))
    if resident:RESIDENT['resident_claims'](rows,result)
    for r in result['measurements']:
        assert r['oracle_pattern']==spec['pattern']and r['local_taylor']is resident and r['resident']is resident
        assert r['tile_words']==4096 and r['setup_seconds']>0 and r['inverse_reused']is True
        assert r['geometry_reuse_checked']is(len(spec['shapes'])>1)
        pair=[q for q in result['measurements']if q['case']==r['case']]
        assert r['work_seconds']==max(q['local_work_seconds']for q in pair)>0
        assert r['validation_capture_seconds']==max(q['local_capture_seconds']for q in pair)>=0
        assert abs(r['local_seconds']-r['local_capture_seconds']-r['local_work_seconds'])<1e-8
        if spec['pattern']=='ones':
            actual=(directory/f"rank-{r['rank']}-{r['case']}-output.bin").read_bytes()
            assert actual==prefix_oracle(n,r['a_count'],r['b_count'],r['rank'])
    result['sample_spec']=spec;return result


def suite(c):
    c['run']('larger-review-tests',['python3','-B','experiments/resident-product-larger/test_review.py'],30)
    plan=specs();c['resources']['larger_plan']=plan;arms=[]
    for spec in plan:
        def check(rows,directory,spec=spec):return review(rows,directory,spec)
        arms.append(dict(name=spec['name'],test='TestScaledGPUProducts',review=check,memcheck=spec['memcheck'],profile=spec['sample']==0,
            env=dict(PIR_PRODUCT_BACKEND=spec['backend'],PIR_SCALED_DIM=str(spec['dim']),PIR_SCALED_PATTERN=spec['pattern'],PIR_SCALED_SHAPES=str(len(spec['shapes'])),PIR_TAYLOR_LOCAL_GUARD='0')))
    return VECTORS['suite'](c,full_fft=True,arms=arms,scope='structured dense s17-field products N16384/65536/262144,3samples/backend,exact linear oracle;small arbitrary/sanitizer references;no fulls17/s18 preprocessing result')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch/32MiBfile/720s/15min;N16K/64K/256K structured dense,3samples host/localGPU;one large geometry perarm to bound proof logs;no default/admission change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
