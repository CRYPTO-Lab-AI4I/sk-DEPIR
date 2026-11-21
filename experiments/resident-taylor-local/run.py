#!/usr/bin/env python3
"""Validate GPU-local Taylor access and compare unchanged product backends."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
FULL=runpy.run_path(str(HERE.parent/'distributed-resident-fft/run.py'))
PRODUCT=runpy.run_path(str(HERE.parent/'distributed-gpu-product/run.py'))
RESIDENT=runpy.run_path(str(HERE.parent/'distributed-resident-product/run.py'))
MATCHED=runpy.run_path(str(HERE.parent/'resident-product-scaling/run.py'))
VECTORS=PRODUCT['VECTORS']


def specs():
    out=[dict(name='legacy-check',kind='legacy',memcheck=False),dict(name='local-fft-check',kind='fft',memcheck=False),dict(name='local-fft-memcheck',kind='fft',memcheck=True),dict(name='local-check',kind='local',memcheck=False),dict(name='local-memcheck',kind='local',memcheck=True)]
    for sample in range(3):
        backends=('host','resident','resident-local')if sample%2==0 else('resident-local','resident','host')
        for backend in backends:
            out.append(dict(name=f'sample-{sample}-{backend}',kind='benchmark',sample=sample,backend=backend,tile=4096,dims=[12,14],memcheck=False))
    return out


def review(rows,directory,spec):
    if spec['kind']=='legacy':return PRODUCT['review'](rows,directory)
    if spec['kind']=='fft':
        result=FULL['review'](rows,directory)
        assert all(r.get('local_taylor')is True for r in result['measurements'])
        return result
    if spec['kind']=='benchmark':
        adapted=dict(spec,backend='host'if spec['backend']=='host'else'resident')
        result=MATCHED['review'](rows,directory,adapted);result['sample_spec']=spec
        assert all(r.get('local_taylor')is(spec['backend']=='resident-local')for r in result['measurements'])
        assert not any(r.get('kind')=='local-window-check'for r in rows),'guard must not prewarm benchmark plans'
        return result
    result=PRODUCT['review'](rows,directory,tile_words=4096,pointwise_words=lambda n,capacity:min(n//2,capacity,17))
    RESIDENT['resident_claims'](rows,result)
    assert all(r.get('local_taylor')is True for r in result['measurements'])
    checks=[r for r in rows if r.get('kind')=='local-window-check'];assert len(checks)==12
    assert {(r['rank'],r['s'],r['dimension'])for r in checks}=={(rank,s,dim)for rank in (0,1)for s,dim in VECTORS['CASES']}
    assert all(r['mutation_kinds']==['upload','scatter','localcopy']for r in checks)
    assert all(all(r.get(k)is True for k in ('canonical_proof_invalidated','rejection_preserved','both_layouts','host_snapshot_preserved'))for r in checks)
    result['local_guard_checks']=len(checks);return result


def suite(c):
    c['run']('local-review-tests',['python3','-B','experiments/resident-taylor-local/test_review.py'],30)
    plan=specs();c['resources']['local_taylor_plan']=plan;arms=[]
    for spec in plan:
        def check(rows,directory,spec=spec):return review(rows,directory,spec)
        benchmark=spec['kind']=='benchmark'
        env=dict(PIR_TAYLOR_LOCAL_GUARD='0')
        if benchmark:env.update(PIR_PRODUCT_TILE_WORDS='4096',PIR_PRODUCT_DIMS='12,14',PIR_PRODUCT_BACKEND=spec['backend'])
        test='TestMatchedGPUProducts'if benchmark else('TestDistributedGPUProducts'if spec['kind']=='legacy'else('TestLocalTaylorGPUFFT'if spec['kind']=='fft'else'TestLocalTaylorGPUProducts'))
        arms.append(dict(name=spec['name'],test=test,review=check,memcheck=spec['memcheck'],profile=benchmark and spec['sample']==0,env=env))
    return VECTORS['suite'](c,full_fft=True,arms=arms,scope='GPU-local Taylor correctness/cache invalidation and matched N4096/16384 products; no fulls17/s18 preprocessing result')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch/720s/15min;legacy and local32/64 checks+local small memcheck then3backend samples;no default/admission change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
