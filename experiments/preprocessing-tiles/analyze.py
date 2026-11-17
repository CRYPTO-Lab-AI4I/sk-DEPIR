#!/usr/bin/env python3
"""Reconcile bounded component correctness and quantitative memory/time counters."""
import json
import runpy
from pathlib import Path
HERE=Path(__file__).resolve().parent


def review(out,record,result):
    cache=Path(record['cache'])
    package=json.loads((out/'components.json').read_text())
    manifest=package['fixtures']
    verify=runpy.run_path(str(HERE/'run.py'))['verify_outputs']
    rows=verify(out/'components.log',cache/'outputs',manifest)
    sanitized=verify(out/'component-memcheck.log',cache/'memcheck-outputs',manifest)
    assert rows==package['measured'] and sanitized==package['sanitized']
    for r in rows:
        if r['kind']=='origin' or r.get('sample')==1:
            fields=('output_file',) if r['kind']=='product' else ('forward_file','inverse_file')
            assert all(r[name] for name in fields), 'final outputs must be retained'
    assert len(manifest)==12  # eight transforms, one affine witness, three products
    assert any(r['kind']=='origin' and r['first_coordinate']>1<<32 for r in rows)
    assert any(r['kind']=='origin' and r['first_coordinate']>1<<32 for r in sanitized)
    for index in range(8):
        subset=[r for r in rows if r['kind']=='transform' and r['index']==index]
        assert {r['backend'] for r in subset}=={'resident','host-tiled'}
        wanted={1,4} if index==7 else {1,2,4}
        assert {r['gpus'] for r in subset}==wanted
        for count in wanted:
            resident=[r for r in subset if r['gpus']==count and r['backend']=='resident']
            assert len(resident)==2 and {r['sample'] for r in resident}=={0,1}
            tiles={1<<20,8<<20} if index==7 else {1536,8192}
            for tile in tiles:
                backed=[r for r in subset if r['gpus']==count and r['backend']=='host-tiled' and r['tile_bytes']==tile]
                assert len(backed)==2 and {r['sample'] for r in backed}=={0,1}
        assert all(r['matches'] for r in subset)
        for r in subset:
            if r['backend']=='host-tiled':
                for direction in ('forward','inverse'):
                    stats=r[direction]
                    assert stats['seconds']>0 and stats['h2d_bytes']>0 and stats['d2h_bytes']>0
                    assert stats['host_stage_bytes']<=r['gpus']*r['tile_bytes']
                    assert stats['owned_device_bytes']<=r['gpus']*(128<<20)
    for index in range(3):
        subset=[r for r in rows if r['kind']=='product' and r['index']==index]
        assert {r['backend'] for r in subset}=={'resident','split'}
        wanted={1,4} if index==2 else {1,2,4}
        assert {r['gpus'] for r in subset}==wanted
        for count in wanted:
            for backend in ('resident','split'):
                samples=[r for r in subset if r['gpus']==count and r['backend']==backend]
                assert len(samples)==2 and {r['sample'] for r in samples}=={0,1}
        for r in subset:
            assert r['subproducts']==(14 if r['backend']=='split' else 1)
            if r['backend']=='split':assert r['largest_domain']==r['n']//4
    assert all(r.get('n',0)<=4096 or r['kind']=='origin' for r in sanitized)
    text=(out/'component-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors' in text and 'LEAK SUMMARY: 0 bytes' in text
    assert list(cache.glob('cpu-profile-benchmark.*'))
    result['components']=package
    result['tests'].extend(['complete independent CPU array and canonical SHA256 comparisons',
       'uint32/uint64, odd/even splits, tiny/tail batches and inverse recovery',
       '>32bit affine row origin vs independent Go Horner',
       'fourteen subproducts vs full GPU and CPU, small schoolbook references',
       'allocation/input admission and CUDA memcheck; native gprof retained'])
    result['scope']='bounded host-backed transform and split product; no fulls18 implementation/admission or default change'


if __name__=='__main__':
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/analyze.py'))['main'](here=HERE,extra_review=review)
