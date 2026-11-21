#!/usr/bin/env python3
"""Local GPU roots followed by a genuinely distributed final polynomial product."""
from pathlib import Path
import hashlib
import runpy
import sys
HERE=Path(__file__).resolve().parent
PRODUCT=runpy.run_path(str(HERE.parent/'distributed-gpu-product/run.py'))
RESIDENT=runpy.run_path(str(HERE.parent/'distributed-resident-product/run.py'))
LARGER=runpy.run_path(str(HERE.parent/'resident-product-larger/run.py'))
VECTORS=PRODUCT['VECTORS']


def specs():
    return [dict(name='arbitrary-reference',dim=14,reference=True,memcheck=False),
            dict(name='root-memcheck',dim=12,reference=False,memcheck=True)]+[
                dict(name=f'root-d{dim}',dim=dim,reference=False,memcheck=False)for dim in (12,14,16)]


def words(path,width=4):
    data=path.read_bytes();assert len(data)%width==0
    return data,[int.from_bytes(data[j:j+width],'little')for j in range(0,len(data),width)]


def root_inputs(rows,directory,dim,layout="peer",verify_inputs=False):
    n=1<<dim;label=f's17-d{dim}-shape-0';count=n-4
    inputs=[r for r in rows if r.get('kind')=='root-input'];assert len(inputs)==2
    assert {r['rank']for r in inputs}=={0,1}
    saved={};support=set();references=[]
    for rank in (0,1):
        r=next(r for r in inputs if r['rank']==rank)
        assert (r['case'],r['s'],r['dimension'],r['support_count'],r['local_count'],r['a_count'],r['b_count'])==(label,17,dim,count,count//2,n//2-1,n//2-1)
        assert r['scalar_witnesses']==4 and r['local_preserved']is True
        assert 0<r['root_owned_device_bytes']<=64<<20 and 0<r['peak_rss_bytes']<1<<30
        assert all(r[k]>=0 for k in ('reference_seconds','local_root_seconds','exchange_seconds','input_seconds'))
        prefix=directory/f'rank-{rank}-{label}'
        artifacts=[('local-root','local_sha256'),('support','support_sha256')]
        if layout=='peer':artifacts.append(('peer-root','peer_sha256'))
        assert r.get('input_layout','peer')==layout
        assert r.get('peer_materialized',True)is(layout=='peer')
        for suffix,field in artifacts:
            data,values=words(Path(str(prefix)+f'-{suffix}.bin'),8 if suffix=='support'else 4)
            assert hashlib.sha256(data).hexdigest()==r[field]
            saved[rank,suffix]=data
            if suffix=='support':
                assert len(values)==count//2 and values==sorted(set(values))and not support.intersection(values)
                own=set(values);assert all(0<=v<(1<<34)-1 and (v*(1<<17))%((1<<34)-1)in own and (v*(1<<17))%((1<<34)-1)!=v for v in values)
                assert max(values)>1<<32;support.update(values)
            else:assert len(values)==n//2-1 and values[0]==1 and all(v<1<<17 for v in values)
        data,values=words(Path(str(prefix)+'-reference.bin'))
        assert len(values)==n//2 and all(v<1<<17 for v in values)
        references.append(data)
    assert len(support)==count
    if layout=='peer':assert saved[0,'local-root']==saved[1,'peer-root']and saved[1,'local-root']==saved[0,'peer-root']
    else:assert not list(directory.glob('*-peer-root.bin'))
    if verify_inputs:
        for rank in (0,1):
            for operand,suffix in enumerate(('left','right')):
                source=saved[operand,'local-root']
                expected=(source+b'\0'*(4*n-len(source)))[rank*2*n:(rank+1)*2*n]
                assert (directory/f'rank-{rank}-{label}-{suffix}-input.bin').read_bytes()==expected
    padded=b''.join(references);reference=padded[:4*(count+1)]
    assert not any(padded[4*(count+1):])and int.from_bytes(reference[:4],'little')==1
    leading=int.from_bytes(reference[-4:],'little');assert 0<leading<1<<17
    for r in inputs:assert r['leading_coefficient']==leading and r['reference_sha256']==hashlib.sha256(reference).hexdigest()
    return inputs,references,leading


def review(rows,directory,spec):
    if spec['reference']:
        return LARGER['review'](rows,directory,dict(backend='resident-local',dim=14,pattern='random',shapes=[0,1]))
    dim=spec['dim'];inputs,references,leading=root_inputs(rows,directory,dim,spec.get('layout','peer'),spec.get('verify_inputs',False))
    result=PRODUCT['review'](rows,directory,cases=[(17,dim)],shapes=[0],tile_words=4096,
        pointwise_words=lambda n,capacity:min(n//2,capacity,17),geometry=lambda s,d,shape:((1<<d)//2-1,(1<<d)//2-1,leading))
    RESIDENT['resident_claims'](rows,result)
    for r in result['measurements']:
        assert r['oracle_pattern']=='roots'and r['local_taylor']is True and r['inverse_reused']is True and r['geometry_reuse_checked']is False
        assert r['leading_coefficient']==leading
        actual=(directory/f"rank-{r['rank']}-{r['case']}-output.bin").read_bytes()
        assert actual==references[r['rank']]
    result['root_inputs']=inputs
    return result


def suite(c):
    c['run']('root-merge-review-tests',['python3','-B','experiments/distributed-root-merge/test_review.py'],30)
    plan=specs();c['resources']['root_merge_plan']=plan;arms=[]
    for spec in plan:
        arms.append(dict(name=spec['name'],test='TestScaledGPUProducts'if spec['reference']else'TestDistributedRootMerge',memcheck=spec['memcheck'],profile=not spec['memcheck'],
            env=dict(PIR_ROOT_PRODUCT_DIM=str(spec['dim']),PIR_PRODUCT_BACKEND='resident-local',PIR_SCALED_DIM='14',PIR_SCALED_PATTERN='random',PIR_SCALED_SHAPES='2',PIR_TAYLOR_LOCAL_GUARD='0'),
            review=lambda rows,directory,spec=spec:review(rows,directory,spec)))
    return VECTORS['suite'](c,full_fft=True,arms=arms,scope='bounded local GPU root construction + distributed final merge; replicated small support/native reference and whole local-operand exchange;not production partitioned root/Newton or fulls18 preprocessing')


if __name__=='__main__':
    if '--check'in sys.argv:print('same 2 HPC nodes/2 H200/4 CPU/8 GiB host;128 MiB owned/card aggregate at overlapping factor/product setup,1 GiB RSS/rank,128 MiB cache/2 GiB scratch/32 MiB file,720s/15min;5 arms;no resource increase/default/fullsize claim;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
