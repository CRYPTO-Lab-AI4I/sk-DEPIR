#!/usr/bin/env python3
"""Resident shard layout/transport gate, sharing the validated rank schedule."""
from pathlib import Path
import hashlib
import runpy
import sys
HERE=Path(__file__).resolve().parent
VECTORS=runpy.run_path(str(HERE.parent/'distributed-gpu-vectors/run.py'))
SHAPES=((1,1),(1,3),(3,1),(7,11),(3,17),(16,16))
CASES=[(s,r,c)for s in (4,17)for r,c in (*SHAPES,*(((64,64),)if s==17 else()))]


def band(size,rank):
    # Independent two-rank split, including uneven and empty bands.
    first=(size+1)//2
    return (0,first)if rank==0 else(first,size-first)


def pattern(index,width):
    return (((index+1)*0x9e3779b97f4a7c15)^0xa5a5a5a5a5a5a5a5)&((1<<(8*width))-1)


def review(rows,directory):
    hosts=[r for r in rows if r.get('kind')=='host']
    assert len(hosts)==2 and {r['rank']for r in hosts}=={0,1}and len({r['host']for r in hosts})==2
    assert all(r['visible_devices']and ','not in r['visible_devices']for r in hosts)
    values=[r for r in rows if r.get('kind')=='measurement'];assert len(values)==2*8*len(CASES)
    expected_packets={};operation=1;retained={};outputs=[]
    for s,nr,nc in CASES:
        label=f's{s}-r{nr}-c{nc}';width=4 if s==4 else 8
        for pass_id in range(8):
            a,b=(nr,nc)if pass_id%2==0 else(nc,nr)
            selected=[r for r in values if r['case']==label and r['pass']==pass_id]
            assert len(selected)==2 and {r['rank']for r in selected}=={0,1}and selected[0]['seconds']==selected[1]['seconds']>0
            for r in selected:
                rank=r['rank'];first,count=band(b,rank);_,in_count=band(a,rank)
                assert (r['s'],r['rows'],r['columns'],r['operation'],r['word_bytes'],r['words'])==(s,a,b,operation,width,count*a)
                assert r['row_interleaved']==bool((pass_id//2)&1)and r['column_interleaved']==bool((pass_id//2)&2)
                assert all(r[k]is True for k in ('matches','source_preserved','third_preserved','reused'))
                assert 0<r['owned_device_bytes']<=128<<20 and 0<r['peak_rss_bytes']<1<<30
                key=(rank,label);retained.setdefault(key,r['owned_device_bytes']);assert retained[key]==r['owned_device_bytes']
                prefix=f'rank-{rank}-{label}-pass-{pass_id}'
                actual=(directory/(prefix+'-output.bin')).read_bytes();expected=(directory/(prefix+'-expected.bin')).read_bytes()
                want=b''.join(pattern((column*nc+row if pass_id%2==0 else row*nc+column),width).to_bytes(width,'little')for row in range(first,first+count)for column in range(a))
                assert actual==expected==want
                outputs.append(dict(rank=rank,case=label,pass_id=pass_id,bytes=len(actual),sha256=hashlib.sha256(actual).hexdigest()))
                _,peer_cols=band(b,1-rank);_,peer_rows=band(a,1-rank)
                events=0
                for direction,words in (('send',in_count*peer_cols),('receive',count*peer_rows)):
                    for offset in range(0,words,127):expected_packets[(rank,label,pass_id,operation,offset,direction)]=min(127,words-offset);events+=1
                assert r['proof_events']==events
            operation+=1
    proofs=VECTORS['review_transpose_routes'](rows,expected_packets)
    assert all(sum(r.get('kind')=='completed'and r['rank']==rank and r['cases']==8*len(CASES)for r in rows)==1 for rank in (0,1))
    assert not list(directory.glob('capture-*'))
    return dict(measurements=values,proof_events=proofs,outputs=outputs)


def suite(c):
    c['run']('storage-review-tests',['python3','-B','experiments/distributed-gpu-storage/test_review.py'],30)
    result=VECTORS['suite'](c,review_fn=review,test_name='TestDistributedGPUStorage',scope='resident three-shard storage and shared rank transpose; no resident FFT/product or fulls18 result')
    c['resources'].update(scope='resident three-shard storage and shared rank transpose; no resident FFT/product or fulls18 result',host_transport_tiles_per_rank=1,device_transport_tiles_per_rank=1)
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('existing2HPCnodes/2H200/4CPU/8GiBhost;128MiBowned/card/1GiBRSS/rank/128MiBcache/2GiBscratch;720s/15min;three GPU shards,127-word transfer tile,normal+CUDA memcheck;no admission/default/fulls18 change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
