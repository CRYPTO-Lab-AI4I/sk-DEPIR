#!/usr/bin/env python3
"""Owned root execution with independently retained per-stage native validation."""
from pathlib import Path
import hashlib
import json
import runpy
import shutil
import sys
HERE=Path(__file__).resolve().parent
WIDE=runpy.run_path(str(HERE.parent/'distributed-pipeline-wide/run.py'))
FOUR=WIDE['FOUR'];TREE=FOUR['TREE'];PILOT=FOUR['PILOT'];DISCOVERY=FOUR['DISCOVERY'];VECTORS=FOUR['VECTORS']


def reference_map(manifest):
    expected={}
    for case in manifest['cases']:
        label=case['label'];counts=case['counts'];stages=TREE['stages'](counts)
        assert len(counts)==4 and sum(counts)==case['count']+4
        for rank in range(4):
            expected[f'{label}-rank-{rank}-leaf.bin']=4*counts[rank]
            expected[f'{label}-rank-{rank}-final.bin']=4*stages[-1]['Domain']//4
        for stage in stages:
            for rank in range(stage['Ranks']):expected[f'{label}-level-{stage["Level"]}-group-{stage["First"]}-rank-{rank}-output.bin']=4*stage['Domain']//stage['Ranks']
    assert set(expected)==set(manifest['files'])
    for name,size in expected.items():
        record=manifest['files'][name];source=Path(record['source'])
        assert Path(name).name==name and not source.is_absolute()and '..'not in source.parts and source.parts[:2]==('.cache','liftrs')
        assert record['bytes']==size and 0<size<=32<<20 and len(record['sha256'])==64
        int(record['sha256'],16)
    return expected


def copy_references(root,manifest,target):
    expected=reference_map(manifest);target.mkdir();total=0
    for name,size in expected.items():
        record=manifest['files'][name];source=root/record['source'];assert source.stat().st_size==size
        digest=hashlib.sha256()
        with source.open('rb')as src,(target/name).open('xb')as dst:
            for block in iter(lambda:src.read(16<<10),b''):digest.update(block);dst.write(block)
        assert digest.hexdigest()==record['sha256'],name
        (target/name).chmod(0o400);total+=size
    return total


def review_roots(rows,directory,manifest,references,wide=False):
    cases=[c for c in manifest['cases']if(c['s']==18)==wide]
    values=[]
    for case in cases:
        label=case['label'];counts=case['counts'];stages=TREE['stages'](counts)
        trees=[r for r in rows if r.get('kind')=='root-tree'and r['case']==label];assert len(trees)==4 and {r['rank']for r in trees}==set(range(4))
        assert all(r['leaf_counts']==counts and r['levels']==2 and r['support_count']==case['count']and r['matches']is True for r in trees)
        sources={}
        def checked(name):
            actual=(directory/name).read_bytes();record=manifest['files'][name]
            assert len(actual)==record['bytes']and hashlib.sha256(actual).hexdigest()==record['sha256']
            assert actual==(references/name).read_bytes()
            return actual
        for rank in range(4):sources[rank,1]=checked(f'{label}-rank-{rank}-leaf.bin')
        for stage in stages:
            first,size,n=stage['First'],stage['Ranks'],stage['Domain'];dim=n.bit_length()-1;cols=1<<(dim//2);levels=dim-dim//2;span=2+2*levels;capacity=min(n,3*max(cols,n//cols))
            original=[r for r in rows if r.get('tree_case')==label and r['tree_level']==stage['Level']and r['group_first']==first]
            assert all(r['group_size']==size and 0<=r['group_rank']<size and r['rank']==first+r['group_rank']for r in original)
            selected=[dict(r,rank=r['group_rank'])for r in original]
            starts=[r for r in selected if r.get('kind')=='root-stage'and r['event']=='started'];ends=[r for r in selected if r.get('kind')=='root-stage'and r['event']=='completed']
            assert len(starts)==len(ends)==size
            assert all(r['stage']==stage and r['dimension']==dim for r in starts+ends)
            assert {r['rank']for r in starts}=={r['rank']for r in ends}==set(range(size))
            assert all(r['matches']is True and 0<r['owned_device_bytes']<=64<<20 and r['peak_rss_bytes']<1<<30 for r in ends);values.extend(ends)
            name=f'{label}/level-{stage["Level"]}/group-{first}/root';calls=[r for r in selected if r.get('kind')=='executor-call']
            assert len(calls)==size and {r['rank']for r in calls}==set(range(size))
            assert all((r['case'],r['s'],r['dimension'],r['operation_start'],r['mode'],r['a_count'],r['b_count'],r['keep'])==(name,case['s'],dim,1,'product',stage['Sources'][0]['Count'],stage['Sources'][1]['Count'],stage['Count'])and r['matches']is True for r in calls)
            specs=[];transpose={}
            for child in range(3):
                start=1+child*span;inverse=child==2;specs.append((name,child,n,cols,start,inverse,capacity));matrix=start+(0 if inverse else 2*levels)
                for rank in range(size):
                    for op in (matrix,matrix+1):
                        for key,words in VECTORS['transpose_packets'](rank,name,child,op,n,4096,size):transpose[key]=words
            VECTORS['review_transpose_routes'](selected,transpose,size);PILOT['FFT']['review_taylor_routes'](selected,specs,size)
            children=[sources[first+layout['First'],layout['Ranks']]for layout in stage['Sources']]
            # Same interval/payload oracle as the native hierarchy;only tile size
            # differs from its small127-word diagnostic transport setting.
            TREE['review_operands'](selected,stage,children,tile_words=4096)
            output=b''.join(checked(f'{label}-level-{stage["Level"]}-group-{first}-rank-{rank}-output.bin')for rank in range(size))
            sources[first,size]=output
            closed=[r for r in rows if r.get('kind')=='root-group-closed'and r['case']==label and r['level']==stage['Level']and r['first']==first]
            assert len(closed)==size and {r['rank']for r in closed}==set(range(first,first+size))and all(r['matches']is True for r in closed)
        final=b''.join(checked(f'{label}-rank-{rank}-final.bin')for rank in range(4));assert final==sources[0,4]
        finals=[r for r in rows if r.get('kind')=='root-final'and r['case']==label];assert len(finals)==4 and {r['rank']for r in finals}==set(range(4))
        assert all(r['domain']==stages[-1]['Domain']and r['count']==case['count']+1 and r['matches']is True for r in finals)
    assert sum(r.get('kind')=='root-stage'and r['event']=='started'for r in rows)==8*len(cases)
    assert not list(directory.glob('*-leaf-reference.bin'))and not list(directory.glob('*-rank-*-reference.bin'))
    return dict(measurements=values,cases=[c['label']for c in cases],reference_mode='retained controller files;no complete root diagnostic arrays in executor')


def review(rows,directory,resume=False,source=None,cold=None,cases=(4,6),canonical=None,curve=2):
    manifest=json.loads((directory.parent/'root-references.json').read_text());references=directory.parent/'root-reference-files'
    def cold_review(rows,directory):
        core=[r for r in rows if r.get('discovery')is not True and 'tree_case'not in r]
        result=PILOT['review'](core,directory,streamed=True,ranks=4,hierarchy=True,cases=cases,curve=curve)
        result.update(DISCOVERY['review_discovery']([r for r in rows if r.get('discovery')is True],directory,cases,ranks=4))
        result['hierarchy']=review_roots(rows,directory,manifest,references)
        return result
    result=FOUR['RESTART']['review'](rows,directory,resume,source,cold,ranks=4,hierarchy=True,core_review=cold_review,cases=cases,canonical=canonical,curve=curve)
    assert all(r['root_mode']=='owned'and r['reference_mode']=='retained'and r['support_streamed']is True and r['discovery_layout']=='distributed'for r in result['measurements'])
    if resume:assert not any('tree_case'in r or r.get('kind')in('root-tree','root-reshape')for r in rows)
    return result


def suite(c):
    c['run']('owned-root-review-tests',['python3','-B','experiments/distributed-owned-roots/test_review.py'],30)
    manifest=json.loads((c['source']/'experiments/distributed-owned-roots/root-references.json').read_text())
    total=copy_references(VECTORS['BASE']['ROOT'],manifest,c['cache']/'root-reference-files')
    shutil.copyfile(c['source']/'experiments/distributed-owned-roots/root-references.json',c['cache']/'root-references.json')
    c['resources'].update(root_reference_bytes=total,root_reference_validation='controller byte comparisons after arm,not GPU arithmetic inputs');c['save']()
    c['env']['PIR_PIPELINE_ROOT_MODE']='owned'
    extra=[dict(name='owned-wide',test='TestOwnedWideRoots',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM=''),review=lambda rows,directory:review_roots(rows,directory,manifest,c['cache']/'root-reference-files',wide=True))]
    result=WIDE['suite'](c,review_fn=review,extra_arms=extra)
    c['resources']['scope']='Owned-root genuine s4/s6 cold/restart plus bounded field18 tail/owned roots;not fulls18'
    c['resources']['reference_strategy']='Retained canonical/native files;root-stage oracle files checked by controller;bounded synthetic wide fixture remains diagnostic'
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min;owned genuine cold/resume plus bounded field18 tail/root merge;retained native controller checks;no resource/default increase;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
