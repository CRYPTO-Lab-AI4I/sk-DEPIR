#!/usr/bin/env python3
"""Concurrent subgroup root merges with sharded inter-level redistribution."""
from pathlib import Path
import hashlib
import runpy
import sys
HERE=Path(__file__).resolve().parent
FOUR=runpy.run_path(str(HERE.parent/'distributed-four-worker/run.py'))
PRODUCT=FOUR['FIELD']['PRODUCT'];VECTORS=FOUR['VECTORS']


def stages(counts):
    nodes=[dict(First=i,Ranks=1,Domain=c,Count=c)for i,c in enumerate(counts)];out=[];level=0
    while len(nodes)>1:
        level+=1;parents=[]
        for a,b in zip(nodes[::2],nodes[1::2]):
            width=a['Ranks']+b['Ranks'];count=a['Count']+b['Count']-1;domain=4
            while domain<count or domain<width*width:domain*=2
            sources=[dict(a,First=0),dict(b,First=a['Ranks'])]
            row=dict(Level=level,First=a['First'],Ranks=width,Domain=domain,Count=count,Sources=sources)
            out.append(row);parents.append({k:row[k]for k in ('First','Ranks','Domain','Count')})
        nodes=parents
    return out


def packet_key(p):return tuple(p[k]for k in ('Operation','Domain','Operand','Offset','Words','Source','Target'))


def review_operands(rows,stage,source_bytes):
    n=stage['Domain'];ranks=stage['Ranks'];expected={};packets={}
    for operand,layout in enumerate(stage['Sources']):
        child=source_bytes[operand];assert len(child)==4*layout['Domain']
        for source in range(layout['First'],layout['First']+layout['Ranks']):
            begin=(source-layout['First'])*layout['Domain']//layout['Ranks'];end=min(begin+layout['Domain']//layout['Ranks'],layout['Count'])
            for target in range(ranks):
                if source==target:continue
                lo=max(begin,target*n//ranks);hi=min(end,(target+1)*n//ranks)
                for first in range(lo,hi,127):
                    count=min(127,hi-first);p=(1,n,operand,first,count,source,target)
                    sha=hashlib.sha256(child[4*first:4*(first+count)]).digest()
                    for rank,direction in ((source,'send'),(target,'receive')):expected[(rank,direction,p)]=sha
    for r in rows:
        if r.get('kind')!='operand-proof':continue
        key=(r['rank'],r['direction'],packet_key(r['packet']));assert key not in packets
        own,remote=bytes.fromhex(r['own_proof']),bytes.fromhex(r['remote_proof'])
        assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]==expected[key]
        packets[key]=r
    assert set(packets)==set(expected)
    return len(packets)


def review(rows,directory,small=False):
    hosts=[r for r in rows if r.get('kind')=='host'];assert len(hosts)==4 and {r['rank']for r in hosts}==set(range(4))and len({r['host']for r in hosts})==2
    for host in {r['host']for r in hosts}:
        selected=[r for r in hosts if r['host']==host]
        assert len(selected)==2 and len({r['visible_devices']for r in selected})==1
    cases=[16]if small else[16,4092];results={};all_measurements=[]
    def words(data):assert len(data)%4==0;return[int.from_bytes(data[i:i+4],'little')for i in range(0,len(data),4)]
    for count in cases:
        label=f'roots-{count}';trees=[r for r in rows if r.get('kind')=='root-tree'and r['case']==label];assert len(trees)==4
        expected_counts=[5,5,5,5]if count==16 else[1025,1025,1023,1023]
        for r in trees:assert r['leaf_counts']==expected_counts and r['levels']==2 and r['support_count']==count and r['matches']is True
        expected_stages=stages(expected_counts);sources={}
        for rank in range(4):
            leaf=(directory/f'{label}-rank-{rank}-leaf.bin').read_bytes();ref=(directory/f'{label}-rank-{rank}-leaf-reference.bin').read_bytes()
            assert leaf==ref and len(leaf)==4*expected_counts[rank]and words(leaf)[0]==1 and words(leaf)[-1]!=0
            sources[(rank,1)]=leaf
        case_results=[]
        for stage in expected_stages:
            first,size=stage['First'],stage['Ranks'];dim=stage['Domain'].bit_length()-1
            selected=[dict(r,rank=r['group_rank'])for r in rows if r.get('tree_case')==label and r.get('tree_level')==stage['Level']and r.get('group_first')==first]
            assert all(r['group_size']==size and r['rank']in range(size)for r in selected)
            starts=[r for r in selected if r.get('kind')=='root-stage'];assert len(starts)==size and all(r['stage']==stage and r['dimension']==dim for r in starts)
            selected+= [dict(r,rank=r['rank']-first)for r in hosts if first<=r['rank']<first+size]
            # A two-member first-level group can share one host;the product
            # reviewer validates arithmetic/routes independently of host count.
            dest=directory/f'{label}-level-{stage["Level"]}-group-{first}'
            src=[sources[(first+layout['First'],layout['Ranks'])]for layout in stage['Sources']]
            outputs=b''.join((dest/f'rank-{r}-s18-d{dim}-shape-0-expected.bin').read_bytes()for r in range(size))
            leading=words(outputs)[stage['Count']-1]
            result=PRODUCT['review'](selected,dest,cases=[(18,dim)],shapes=(0,),ranks=size,host_count=len({r['host']for r in selected if r.get('kind')=='host'}),pointwise_words=lambda n,capacity:min(n//size,capacity,17),geometry=lambda s,d,shape:(stage['Sources'][0]['Count'],stage['Sources'][1]['Count'],leading))
            FOUR['FIELD']['RESIDENT']['resident_claims'](selected,result,ranks=size)
            assert all(r['owned_device_bytes']<=64<<20 and r['local_taylor']is True for r in result['measurements'])
            source_count=review_operands(selected,stage,src)
            actual=b''.join((dest/f'rank-{r}-s18-d{dim}-shape-0-output.bin').read_bytes()for r in range(size));assert actual==outputs
            coeff=words(actual);assert coeff[0]==1 and coeff[stage['Count']-1]!=0 and not any(coeff[stage['Count']:])and all(v<1<<18 for v in coeff)
            sources[(first,size)]=actual;all_measurements.extend(result['measurements']);case_results.append(dict(stage=stage,operand_proofs=source_count,**result))
        final=expected_stages[-1]
        actual=b''.join((directory/f'{label}-rank-{rank}-final.bin').read_bytes()for rank in range(4));ref=b''.join((directory/f'{label}-rank-{rank}-reference.bin').read_bytes()for rank in range(4))
        assert actual==ref==sources[(0,4)]and len(actual)==4*final['Domain']
        closed=[r for r in rows if r.get('kind')=='root-group-closed'and r['case']==label];assert len(closed)==8
        for rank in range(4):
            assert sorted(r['level']for r in closed if r['rank']==rank)==[1,2]
            witness=[r for r in rows if r.get('kind')=='root-reference'and r.get('case')==label and r['rank']==rank];assert len(witness)==1 and witness[0]['matches']is True and witness[0]['witnesses']==3 and witness[0]['wide_indices']is True
            final_rows=[r for r in rows if r.get('kind')=='root-final'and r['case']==label and r['rank']==rank];assert len(final_rows)==1 and final_rows[0]['matches']is True and final_rows[0]['domain']==final['Domain']and final_rows[0]['count']==count+1 and final_rows[0]['seconds']>0
        results[label]=case_results
    assert sum(r.get('kind')=='root-stage'for r in rows)==8*len(cases)
    assert sum(r.get('kind')=='hierarchy-completed'and r['cases']==len(cases)for r in rows)==4
    assert not list(directory.rglob('capture-*'))
    return dict(measurements=all_measurements,cases=results)


def suite(c):
    c['run']('hierarchy-review-tests',['python3','-B','experiments/distributed-root-hierarchy/test_review.py'],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for name,race in (('root-routing-cpu',[]),('root-routing-race',['-race'])):
        c['run'](name,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestShardedRoot|^TestRankOperand|^TestRootTree','-count=1','-timeout=30s'],60)
    c['run']('mpi-admission',[c['go'],'test','-tags','liftrs_mpi','./experiments/distributed-root/mpi','-count=1','-timeout=20s'],45)
    arms=[dict(name=name,test='TestHierarchicalRootMerge',memcheck=small,profile=not small,env=dict(PIR_ROOT_HIERARCHY_SMALL='1'if small else'0'),review=lambda rows,directory,small=small:review(rows,directory,small))for name,small in (('normal',False),('memcheck',True))]
    return VECTORS['suite'](c,full_fft=True,arms=arms,ranks=4,scope='hierarchical s18 root products,2+2 isolated subgroups then4-worker final merge;sharded redistribution;not fulls18 discovery/generator/restart')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2 HPC nodes/2H200/4CPU/8GiBhost,4workers,64MiBowned/worker=128MiB/card;1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min;normal16/4092 synthetic roots,small16-root sanitizer;group2 then4;no fulls18/default/resource increase;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
