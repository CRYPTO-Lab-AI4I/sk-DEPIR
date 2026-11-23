#!/usr/bin/env python3
"""Genuine full pilot with local discovery grids, union bits and orbit routing."""
from pathlib import Path
import hashlib
import runpy
import sys
HERE=Path(__file__).resolve().parent
PILOT=runpy.run_path(str(HERE.parent/'distributed-preprocessing-pilot/run.py'))
VECTORS=PILOT['VECTORS'];FFT=PILOT['FFT']
# Canonical artifacts independently retained by integrated job78898.
CANONICAL={4:'69ebae873990075af7c9c558b5e9d76b2a4d237b89b4bb824205eb2de049c712',6:'a384ea4123bb5bcbfaa9cde64afbe6e6c9091373e6c22129ceda8abdec5709c4'}


def summary(rows,s,ranks=2,native_reference=None):
    results=[r for r in rows if r.get('kind')=='discovery-result'and r['s']==s]
    assert len(results)==ranks and {r['rank']for r in results}==set(range(ranks))
    results.sort(key=lambda r:r['rank']);n=1<<(2*s);k=results[0]['k'];samples=results[0]['samples']
    assert 0<samples<=2+(128+s-1)//s and all(r['samples']==samples for r in results)
    assert sum(r['local_good']for r in results)==sum(r['raw_count']for r in results)==sum(r['root_count']for r in results)==k
    for r in results:assert r['k']==k and r['grid_words']==n//ranks and r['first']==r['rank']*n//ranks and 0<r['owned_device_bytes']<=64<<20 and r['peak_rss_bytes']<1<<30
    reference=[r for r in rows if r.get('kind')=='discovery-reference'and r['s']==s]
    if native_reference is None:
        assert len(reference)==1 and reference[0]['rank']==0 and reference[0]['matches']is True and reference[0]['samples']==samples
        assert len(results[0]['entropy_sha256'])==64 and results[0]['entropy_sha256']==reference[0]['entropy_sha256']==reference[0]['reference_entropy_sha256']
    else:
        # A reference-independent runtime emits no in-process reference claim.
        # Bind every measured count/entropy to separately verified native data.
        assert not reference
        parameters=native_reference['parameters']
        assert parameters['spec']['FieldDegree']==s and parameters['spec']['Dimension']==2
        assert (parameters['n'],parameters['k'],parameters['samples'])==(n,k,samples)
        assert len(results[0]['entropy_sha256'])==64 and results[0]['entropy_sha256']==native_reference['entropy_sha256']
    count_rows=[r for r in rows if r.get('kind')=='discovery-sample'and r['s']==s];assert len(count_rows)==ranks*samples
    previous=[0]*ranks
    for sample in range(1,samples+1):
        pair=sorted([r for r in count_rows if r['sample']==sample],key=lambda r:r['rank']);assert len(pair)==ranks and [r['rank']for r in pair]==list(range(ranks))
        total=sum(r['local_count']for r in pair)
        for rank,r in enumerate(pair):
            assert r['k']==k and r['global_count']==total and previous[rank]<=r['local_count']<=n//ranks
            assert r['good_counts']==[q['local_good']for q in results]and r['operation_start']==1+(sample-1)*(4+2*s)
            previous[rank]=r['local_count']
        assert total<k if sample<samples else total==k
    return results,k,samples


def review(rows,directory,small=False,streamed=False):
    distributed=[r for r in rows if r.get('discovery')is True]
    core=[r for r in rows if r.get('discovery')is not True]
    result=PILOT['review'](core,directory,small,streamed=streamed)
    for r in result['measurements']:assert r['discovery_layout']=='distributed'
    cases=[4]if small else[4,6]
    for s in cases:assert result['artifacts'][f's{s}']['code_sha256']==CANONICAL[s]
    result.update(review_discovery(distributed,directory,cases))
    return result


def review_discovery(distributed,directory,cases,ranks=2,native_references=None):
    """Review the shared discovery component independently of later pipeline stages."""
    transpose={};specs=[];expected={};payload={};summaries={}
    def wire(label,kind,sequence,n,offset,count,source,target,width,data=None):
        for rank,direction in ((source,'send'),(target,'receive')):
            key=(label,rank,direction,kind,sequence,n,offset,count,source,target,width);expected[key]=True
            if data is not None:payload[key]=hashlib.sha256(data).digest()
    for s in cases:
        info,k,samples=summary(distributed,s,ranks,None if native_references is None else native_references[s]);summaries[f's{s}']=info;n=1<<(2*s);q=1<<s;label=f's{s}';capacity=min(n,3*q)
        indices=[]
        for rank in range(ranks):
            _,part=PILOT['read_words'](directory/f'{label}-rank-{rank}-support.bin',8)
            assert len(part)==info[rank]['root_count']
            assert all(min(i,i*q%(n-1))%ranks==rank for i in part)
            indices.extend(part)
        indices.sort()
        for sample in range(1,samples+1):
            start=1+(sample-1)*(4+2*s)
            for suffix,operations in (('base',(start,start+1)),('inverse',(start+2,start+3))):
                case=f'{label}/discovery-{sample}/{suffix}'
                for rank in range(ranks):
                    for op in operations:
                        for key,words in VECTORS['transpose_packets'](rank,case,0,op,n,4096,ranks):transpose[key]=words
            specs.append((f'{label}/discovery-{sample}/inverse',0,n,q,start+2,True,capacity))
            for owner in range(1,ranks):
                for offset in range(0,info[owner]['local_good'],4096):wire(label,1,sample,n,offset,min(4096,info[owner]['local_good']-offset),0,owner,4)
        for source in range(ranks):
            for target in range(ranks):
                if source==target:continue
                outgoing=[i for i in indices if source*n//ranks<=i<(source+1)*n//ranks and min(i,i*q%(n-1))%ranks==target]
                for offset in range(0,len(outgoing),4096):
                    part=outgoing[offset:offset+4096];wire(label,2,0,n,offset,len(part),source,target,8,b''.join(i.to_bytes(8,'little')for i in part))
        for source in range(1,ranks):
            archive=[i for i in indices if source*n//ranks<=i<(source+1)*n//ranks]
            for offset in range(0,len(archive),4096):
                part=archive[offset:offset+4096];wire(label,3,0,n,offset,len(part),source,0,8,b''.join(i.to_bytes(8,'little')for i in part))
    VECTORS['review_transpose_routes'](distributed,transpose,ranks);FFT['review_taylor_routes'](distributed,specs,ranks)
    actual={}
    for r in distributed:
        if r.get('kind')!='discovery-wire':continue
        key=(r['case'],r['rank'],r['direction'],r['phase_kind'],r['sequence'],r['domain'],r['offset'],r['words'],r['source'],r['target'],r['width'])
        assert key not in actual;actual[key]=r
        own,remote=bytes.fromhex(r['own_proof']),bytes.fromhex(r['remote_proof']);assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]
        if key in payload:assert own[:32]==payload[key]
    assert set(actual)==set(expected)
    for key,r in actual.items():
        peer=(key[0],key[9]if key[2]=='send'else key[8],'receive'if key[2]=='send'else'send',*key[3:]);other=actual[peer]
        assert r['own_proof']==other['remote_proof']and r['remote_proof']==other['own_proof']
    return dict(discovery=summaries,discovery_wire_events=len(actual))


def suite(c):
    c['run']('discovery-review-tests',['python3','-B','experiments/distributed-discovery-pilot/test_review.py'],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for name,race in (('discovery-cpu',[]),('discovery-race',['-race'])):
        c['run'](name,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestDiscoverySampling|^TestRankSquare|^TestPolynomialProductsIndependent$|^TestFastPolynomialDivision$','-count=1','-timeout=30s'],60)
    arms=[dict(name=name,test='TestDistributedPreprocessingPilot',memcheck=small,profile=not small,env=dict(PIR_PIPELINE_SMALL='1'if small else'0',PIR_PIPELINE_DISCOVERY='distributed'),review=lambda rows,directory,small=small:review(rows,directory,small))for name,small in (('normal',False),('memcheck',True))]
    return VECTORS['suite'](c,full_fft=False,arms=arms,scope='genuine s4/s6 assembled pilot with sharded monomial/grid/union and direct orbit routing;coordinator random stream/diagnostic support archive/publication;not fulls18/fresh-process recovery')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2 HPC nodes/2 H200/4 CPU/8 GiB RAM;128MiB owned/card,1GiB RSS/rank,128MiB cache/2GiB scratch/32MiB file;720s/15min;2 integrated arms;no fulls18/default/admission change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
