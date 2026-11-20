#!/usr/bin/env python3
"""Taylor-complete distributed GPU FFT gate, reusing the admitted composition runner."""
from pathlib import Path
import json
import runpy
import sys
HERE=Path(__file__).resolve().parent
VECTORS=runpy.run_path(str(HERE.parent/'distributed-gpu-vectors/run.py'))


def packet_key(p):
    t=p['Target']
    return (p['Operation'],p['Domain'],p['Columns'],p['Level'],p['Quotient'],t['Block'],t['First'],t['Count'],p['Slot'],p['Source'],p['Words'])


def expected_windows(n,cols,degrees,level,quotient,capacity):
    """Enumerate global target intervals, independently of rank cursors/waves."""
    if not degrees:return []
    half=cols<<level;span=capacity//(len(degrees)+1);minimum,maximum=min(degrees),max(degrees)
    assert degrees==sorted(set(degrees))and minimum>0 and maximum<=half//2 and span>0
    events=[]
    for block in range(0,n,2*half):
        target_base=block+(half if quotient else 0)
        begin,end=(0,maximum)if quotient else(minimum,half)
        for target in (0,1):
            lo=max(target_base+begin,target*n//2);hi=min(target_base+end,(target+1)*n//2)
            for start in range(lo,hi,span):
                count=min(span,hi-start);first=start-target_base
                for slot,d in enumerate(degrees):
                    origin=block+(2*half if quotient else half)+first-d
                    # Operative remainder sources are in the completed upper
                    # half. Exclude inactive prefixes that may overlap targets.
                    active=origin+(0 if quotient else max(0,min(count,d-first)))
                    end_source=min(origin+count,block+2*half)
                    for source in (0,1):
                        if source==target:continue
                        a=max(active,source*n//2);b=min(end_source,(source+1)*n//2)
                        if a<b:events.append((target,source,block,first,count,slot,a,b-a))
    return events


def review(rows,directory):
    operations={};expected={};actual={};packets={};plans=[r for r in rows if r.get('kind')=='taylor-plan']
    measurements=[r for r in rows if r.get('kind')=='measurement'];operation=1;plan_count=0
    for index,(s,dim)in enumerate(VECTORS['CASES']):
        label=f's{s}-d{dim}';n=1<<dim;cols=1<<(dim//2);levels=dim-dim//2;capacity=min(n,3*max(cols,n//cols))
        for pass_id in (0,1):
            starts=[r['operation_start']for r in measurements if r['case']==label and r['pass']==pass_id];assert starts==[operation,operation]
            operations[(index,pass_id)]=operation+(2*levels if pass_id==0 else 0)
            base=operation+(2 if pass_id==1 else 0)
            for step in range(levels):
                level=levels-1-step if pass_id==0 else step
                selected=[r for r in plans if r['case']==label and r['pass']==pass_id and r['level']==level]
                assert len(selected)==2 and {r['rank']for r in selected}=={0,1};plan_count+=2
                assert selected[0]['degrees']==selected[1]['degrees']and all(r['capacity']==capacity and r['operation_start']==operation for r in selected)
                for stage in (0,1):
                    quotient=(stage==0)!=(pass_id==1);op=base+2*step+stage
                    for target,source,block,first,count,slot,a,words in expected_windows(n,cols,selected[0]['degrees'],level,quotient,capacity):
                        p=dict(Operation=op,Domain=n,Columns=cols,Level=level,Quotient=quotient,Target=dict(Block=block,First=first,Count=count),Slot=slot,Source=a,Words=words)
                        for rank,peer,direction in ((source,target,'send'),(target,source,'receive')):expected[(rank,label,pass_id,peer,direction,packet_key(p))]=words
            operation+=2+2*levels
    assert len(plans)==plan_count
    common=VECTORS['review'](rows,directory,operations=operations,full_fft=True)
    for r in rows:
        if r.get('kind')!='taylor-proof':continue
        key=(r['rank'],r['case'],r['pass'],r['peer'],r['direction'],packet_key(r['packet']));assert key not in actual
        own,remote=bytes.fromhex(r['own_proof']),bytes.fromhex(r['remote_proof'])
        assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]
        actual[key]=r['packet']['Words'];packets[key]=r
    assert actual==expected
    for key,r in packets.items():
        rank,label,pass_id,peer,direction,p=key;other=packets[(peer,label,pass_id,rank,'receive'if direction=='send'else'send',p)]
        assert r['own_proof']==other['remote_proof']and r['remote_proof']==other['own_proof']
    for r in measurements:
        assert r['taylor_proof_events']==sum(key[0]==r['rank']and key[1]==r['case']and key[2]==r['pass']for key in expected)
    return dict(**common,taylor_proof_events=len(actual),taylor_plans=len(plans))


def suite(c):
    c['run']('fft-review-tests',['python3','-B','experiments/distributed-gpu-fft/test_review.py'],30)
    return VECTORS['suite'](c,full_fft=True,review_fn=review)


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPC nodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch;720s/15min; baseline separable+full forward/inverse FFT+small CUDA memcheck; no fulls18 support/admission; no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
