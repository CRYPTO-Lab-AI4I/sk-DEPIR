#!/usr/bin/env python3
"""Compare whole-peer exchange with bounded coefficient ownership routing."""
from pathlib import Path
import hashlib
import runpy
import sys
HERE=Path(__file__).resolve().parent
ROOT=runpy.run_path(str(HERE.parent/'distributed-root-merge/run.py'))
VECTORS=ROOT['VECTORS']


def specs():
    return [dict(name='peer-reference',dim=16,layout='peer',memcheck=False),
            dict(name='routed-memcheck',dim=12,layout='routed',memcheck=True)]+[
                dict(name=f'routed-d{dim}',dim=dim,layout='routed',memcheck=False)for dim in (12,14,16)]


def review(rows,directory,spec):
    dim=spec['dim'];n=1<<dim;count=n//2-1;layout=spec['layout']
    result=ROOT['review'](rows,directory,dict(reference=False,dim=dim,layout=layout,verify_inputs=True))
    proofs=[r for r in rows if r.get('kind')=='operand-proof'];actual={}
    if layout=='routed':
        source=(directory/f'rank-1-s17-d{dim}-shape-0-local-root.bin').read_bytes()
        for r in proofs:
            p=r['packet'];direction=r['direction'];offset=p['Offset'];words=min(4096,count-offset)
            assert p==dict(Operation=1,Domain=n,Operand=1,Offset=offset,Words=words,Source=1,Target=0)
            assert offset in range(0,count,4096)and direction in ('send','receive')
            assert r['rank']==(1 if direction=='send'else 0)
            own,remote=bytes.fromhex(r['own_proof']),bytes.fromhex(r['remote_proof'])
            digest=hashlib.sha256(source[4*offset:4*(offset+words)]).digest()
            assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]==digest
            key=(direction,offset);assert key not in actual;actual[key]=r
        assert set(actual)=={(d,o)for d in ('send','receive')for o in range(0,count,4096)}
        for offset in range(0,count,4096):
            a,b=actual['send',offset],actual['receive',offset]
            assert a['own_proof']==b['remote_proof']and a['remote_proof']==b['own_proof']
    else:assert not proofs
    for r in result['root_inputs']:
        expected=(count,count)if layout=='peer'else((0,count)if r['rank']==0 else(count,0))
        assert (r['sent_words'],r['received_words'])==expected and r['route_tile_words']==4096
    result['operand_proof_events']=len(proofs)
    return result


def suite(c):
    c['run']('root-routing-review-tests',['python3','-B','experiments/distributed-root-routing/test_review.py'],30)
    # The shared arithmetic suite already covers other rank planners; include
    # the new CUDA-free router and its races before any GPU work.
    c['env'].update(CC='mpicc',CXX='mpicxx')
    for label,race in (('operand-routing-cpu',[]),('operand-routing-race',['-race'])):
        c['env']['CGO_ENABLED']='1'if race else'0'
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestRankOperand','-count=1','-timeout=20s'],60)
    plan=specs();c['resources']['root_routing_plan']=plan
    arms=[dict(name=s['name'],test='TestDistributedRootMerge',memcheck=s['memcheck'],profile=not s['memcheck'],
        env=dict(PIR_ROOT_PRODUCT_DIM=str(s['dim']),PIR_ROOT_INPUT_LAYOUT=s['layout'],PIR_TAYLOR_LOCAL_GUARD='0'),
        review=lambda rows,directory,s=s:review(rows,directory,s))for s in plan]
    return VECTORS['suite'](c,full_fft=True,arms=arms,scope='bounded root operand ownership routing versus whole-peer reference; native/scalar checks;no fulls18 discovery/Newton/production ownership claim')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2 HPC nodes/2 H200/4 CPU/8 GiB host;128 MiB owned/card,1 GiB RSS/rank,128 MiB cache/2 GiB scratch/32 MiB file;720s/15min;5arms peer/routed;no admission/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
