#!/usr/bin/env python3
"""Bounded distributed polynomial product through two forward and one inverse FFT."""
from pathlib import Path
import hashlib
import runpy
import sys
HERE=Path(__file__).resolve().parent
FFT=runpy.run_path(str(HERE.parent/'distributed-gpu-fft/run.py'))
VECTORS=FFT['VECTORS']


def review(rows,directory,pointwise_words=None,cases=None,tile_words=127,shapes=(0,1),geometry=None,ranks=2,host_count=2):
    hosts=[r for r in rows if r.get('kind')=='host'];assert len(hosts)==ranks and {r['rank']for r in hosts}==set(range(ranks))and len({r['host']for r in hosts})==host_count
    assert all(r['visible_devices']and ','not in r['visible_devices']for r in hosts)
    cases=VECTORS['CASES']if cases is None else cases
    values=[r for r in rows if r.get('kind')=='measurement'];assert len(values)==ranks*len(shapes)*len(cases)
    checks=[r for r in rows if r.get('kind')=='pointwise-check'];assert len(checks)==ranks*len(cases)
    specs=[];expected_packets={};operation=1
    for s,dim in cases:
        n=1<<dim;cols=1<<(dim//2);levels=dim-dim//2;capacity=min(n,3*max(cols,n//cols));span=2+2*levels
        selected=[r for r in checks if r['s']==s and r['dimension']==dim];assert len(selected)==ranks and {r['rank']for r in selected}==set(range(ranks))and all(r['matches']is True and r['words']==(pointwise_words(n,capacity)if pointwise_words else min(capacity,17))for r in selected)
        for shape in shapes:
            label=f's{s}-d{dim}-shape-{shape}'
            ac,bc=(n//4+1,3*n//4)if shape==0 else(n//4,n//2)
            leading=1
            if geometry is not None:ac,bc,leading=geometry(s,dim,shape)
            assert ac>0 and bc>0 and ac+bc-1<=n and 0<leading<1<<s
            for child in (0,1,2):
                start=operation+child*span;inverse=child==2;specs.append((label,child,n,cols,start,inverse,capacity))
                first_op=start+(0 if inverse else 2*levels)
                for rank in range(ranks):
                    for op in (first_op,first_op+1):
                        for key,words in VECTORS['transpose_packets'](rank,label,child,op,n,tile_words,ranks):expected_packets[key]=words
            chosen=[r for r in values if r['case']==label];assert len(chosen)==ranks and {r['rank']for r in chosen}==set(range(ranks))and all(r['seconds']==chosen[0]['seconds']>0 for r in chosen)
            for r in chosen:
                rank=r['rank'];assert(r['s'],r['dimension'],r['shape'],r['a_count'],r['b_count'],r['first_coefficient'],r['words'],r['slot_bytes'],r['output_word_bytes'],r['operation_start'])==(s,dim,shape,ac,bc,rank*n//ranks,n//ranks,4 if s==4 else 8,4,operation)
                assert r['matches']is True and 0<r['owned_device_bytes']<=128<<20 and 0<r['peak_rss_bytes']<1<<30
                prefix=f'rank-{rank}-{label}';a=(directory/(prefix+'-output.bin')).read_bytes();b=(directory/(prefix+'-expected.bin')).read_bytes()
                assert a==b and len(a)==4*n//ranks and hashlib.sha256(a).hexdigest()==r['output_sha256']
                coeff=[int.from_bytes(a[i:i+4],'little')for i in range(0,len(a),4)]
                assert all(v<1<<s for v in coeff)
                for j,v in enumerate(coeff):
                    k=rank*n//ranks+j
                    if k==0:assert v==1
                    if k==ac+bc-2:assert v==leading
                    if k>=ac+bc-1:assert v==0
                assert r['proof_events']==sum(key[0]==rank and key[1]==label for key in expected_packets)
            operation+=3*span
    proof_count=VECTORS['review_transpose_routes'](rows,expected_packets,ranks)
    routes,counts=FFT['review_taylor_routes'](rows,specs,ranks)
    for r in values:assert r['taylor_proof_events']==sum(counts.get((r['rank'],r['case'],child),0)for child in (0,1,2))
    assert all(sum(r.get('kind')=='completed'and r['rank']==rank and r['cases']==len(shapes)*len(cases) for r in rows)==1 for rank in range(ranks))
    assert not list(directory.glob('capture-*'))
    return dict(measurements=values,proof_events=proof_count,pointwise_checks=len(checks),**routes)


def suite(c):
    c['run']('product-review-tests',['python3','-B','experiments/distributed-gpu-product/test_review.py'],30)
    c['run']('fft-review-tests',['python3','-B','experiments/distributed-gpu-fft/test_review.py'],30)
    result=VECTORS['suite'](c,full_fft=True,review_fn=review,test_name='TestDistributedGPUProducts')
    c['resources']['scope']='bounded distributed polynomial products;no fullsize preprocessing support'
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch;720s/15min;schoolbook products,full-extension pointwise,staged MPI and small CUDA memcheck;no fulls18/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
