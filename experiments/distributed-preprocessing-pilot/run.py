#!/usr/bin/env python3
"""One genuine-code pipeline gate: discovery through canonical artifact."""
from pathlib import Path
import hashlib
import json
import runpy
import sys
HERE=Path(__file__).resolve().parent
FFT=runpy.run_path(str(HERE.parent/'distributed-gpu-fft/run.py'))
VECTORS=FFT['VECTORS']


def prefixes(target):
    known=1;out=[]
    while known<target:
        known=min(2*known,target);out.append(known)
    return out


def read_words(path,width=4):
    data=path.read_bytes();assert len(data)%width==0
    return data,[int.from_bytes(data[j:j+width],'little')for j in range(0,len(data),width)]


def artifacts(directory,s,k):
    n=1<<(2*s);label=f's{s}'
    def joined(stage):return b''.join((directory/f'{label}-rank-{r}-{stage}.bin').read_bytes()for r in (0,1))
    reference={stage:(directory/f'{label}-rank-0-reference-{stage}.bin').read_bytes()for stage in ('p','g','spectrum')}
    for stage,count in (('p',k+1),('g',n-k),('spectrum',n)):
        data=joined(stage);assert len(data)==4*n and data==reference[stage]+b'\0'*(4*(n-count))
    assert joined('identity')==(1).to_bytes(4,'little')+b'\0'*(4*(n-2))+(1).to_bytes(4,'little')
    support=[]
    for rank in (0,1):
        _,indices=read_words(directory/f'{label}-rank-{rank}-support.bin',8)
        assert indices==sorted(set(indices))and not set(support).intersection(indices);support.extend(indices)
    assert len(support)==k
    binding=hashlib.sha256((directory/f'{label}-support.bin').read_bytes()).hexdigest()
    for stage,count,source in [('root',k+1,reference['p'])]+[(f'newton-{size}',size,reference['g'][:4*size])for size in prefixes(n-k)]:
        commit=json.loads((directory/f'{label}-{stage}-commit.json').read_text())
        assert commit==dict(s=s,k=k,n=n,prefix=count,ranks=2,support_sha256=binding,stage=stage,sha256=[hashlib.sha256((directory/f'{label}-rank-{r}-{stage}.bin').read_bytes()).hexdigest()for r in (0,1)])
        assert joined(stage)==source+b'\0'*(4*(n-count))
    actual=(directory/f'{label}-code.bin').read_bytes();expected=(directory/f'{label}-reference-code.bin').read_bytes()
    assert actual==expected and actual[:9]==b'LRSprobe1'and len(actual)==25+4*(k+1+n)
    assert [int.from_bytes(actual[i:i+4],'little')for i in (9,13,17,21)]==[s,2,k,n]
    return dict(code_sha256=hashlib.sha256(actual).hexdigest(),checkpoint_prefixes=prefixes(n-k))


def packet_key(p):return tuple(p[k]for k in ('Operation','Domain','Operand','Offset','Words','Source','Target'))


def review(rows,directory,small=False):
    cases=[4]if small else[4,6];values=[r for r in rows if r.get('kind')=='pipeline'];assert len(values)==2*len(cases)
    hosts=[r for r in rows if r.get('kind')=='host'];assert len(hosts)==2 and len({r['host']for r in hosts})==2
    assert all(r['visible_devices']and ','not in r['visible_devices']for r in hosts)
    calls=[r for r in rows if r.get('kind')=='executor-call'];expected_calls={};specs=[];transpose={};operand={};result={}
    def operand_packet(label,op,n,which,offset,count,source,target):
        for rank,direction in ((source,'send'),(target,'receive')):operand[(label,rank,direction,(op,n,which,offset,count,source,target))]=True
    for s in cases:
        n=1<<(2*s);cols=1<<s;span=2+2*s;capacity=min(n,3*cols);label=f's{s}'
        selected=[r for r in values if r['s']==s];assert len(selected)==2 and {r['rank']for r in selected}=={0,1}
        k=selected[0]['k'];ac,bc=selected[0]['root_a_count'],selected[0]['root_b_count'];assert 1<=k<=n//2 and ac+bc==k+2
        for r in selected:
            assert (r['n'],r['k'],r['first'],r['words'],r['root_a_count'],r['root_b_count'])==(n,k,r['rank']*n//2,n//2,ac,bc)
            assert all(r[x]is True for x in ('matches','identity_verified','prefix_reloaded'))and 0<r['owned_device_bytes']<=128<<20 and r['peak_rss_bytes']<1<<30
        for rank,count in ((0,ac),(1,bc)):
            assert (directory/f'{label}-rank-{rank}-support.bin').stat().st_size==8*(count-1)
            target=1-rank;first=target*n//2;end=min(first+n//2,count)
            for offset in range(first,end,4096):operand_packet(label,1,n,rank,offset,min(4096,end-offset),rank,target)
        operations=[];op=1
        operations.append((label+'/root','product',op,ac,bc,k+1));op+=3*span;known=1
        for size in prefixes(n-k):
            for source in (0,1):
                for target in (0,1):
                    if source==target:continue
                    first=max(source*n//2,(target*n//2+1)//2);end=min((source+1)*n//2,((target+1)*n//2+1)//2,(size+1)//2)
                    for offset in range(first,end,4096):operand_packet(label,op,n,0,offset,min(4096,end-offset),source,target)
            op+=1;operations.append((label+f'/newton-{size}','product',op,min(k+1,size),min(2*known-1,size),size));op+=3*span;known=size
        operations.append((label+'/identity','product',op,k+1,n-k,n));op+=3*span
        operations.append((label+'/spectrum','spectrum',op,None,None,n));op+=span
        for _ in range(3):operand_packet(label,op,n,0,n//2,n//2,1,0);op+=1
        for name,mode,start,a,b,keep in operations:
            for rank in (0,1):expected_calls[name,rank]=(s,start,mode,a,b,keep)
            for child in range(3 if mode=='product'else 1):
                base=start+child*span;inverse=child==2;specs.append((name,child,n,cols,base,inverse,capacity));matrix=base+(0 if inverse else 2*s)
                for rank in (0,1):
                    for transpose_op in (matrix,matrix+1):
                        for direction in ('send','receive'):transpose[(rank,name,child,transpose_op,0,direction)]=n//4
        result[label]=artifacts(directory,s,k)
    actual_calls={}
    for r in calls:
        key=(r['case'],r['rank']);assert key not in actual_calls and r['matches']is True
        actual_calls[key]=(r['s'],r['operation_start'],r['mode'],r.get('a_count'),r.get('b_count'),r['keep'])
    assert actual_calls==expected_calls
    VECTORS['review_transpose_routes'](rows,transpose);FFT['review_taylor_routes'](rows,specs)
    proofs={}
    for r in rows:
        if r.get('kind')not in ('operand-proof','support-proof'):continue
        own,remote=bytes.fromhex(r['own_proof']),bytes.fromhex(r['remote_proof'])
        assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]
        if r['kind']=='operand-proof':
            key=(r['case'],r['rank'],r['direction'],packet_key(r['packet']));assert key not in proofs;proofs[key]=r
    assert set(proofs)==set(operand)
    for key,r in proofs.items():
        label,rank,direction,packet=key;peer=proofs[label,1-rank,'receive'if direction=='send'else'send',packet]
        assert r['own_proof']==peer['remote_proof']and r['remote_proof']==peer['own_proof']
    for s in cases:
        label=f's{s}';checks=[r for r in rows if r.get('kind')=='checkpoint'and r['case']==label]
        assert sum(r['reloaded']for r in checks)==2
    assert all(sum(r.get('kind')=='completed'and r['rank']==rank and r['cases']==len(cases)for r in rows)==1 for rank in (0,1))
    assert not list(directory.glob('capture-*'))
    return dict(measurements=values,artifacts=result,executor_calls=len(calls),operand_proof_events=len(proofs))


def suite(c):
    c['run']('pilot-review-tests',['python3','-B','experiments/distributed-preprocessing-pilot/test_review.py'],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for name,race in (('integrated-cpu',[]),('integrated-race',['-race'])):
        c['run'](name,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestRankSquare|^TestPolynomialProductsIndependent$|^TestFastPolynomialDivision$', '-count=1','-timeout=30s'],60)
    arms=[dict(name=name,test='TestDistributedPreprocessingPilot',memcheck=small,profile=not small,env=dict(PIR_PIPELINE_SMALL='1'if small else'0'),review=lambda rows,directory,small=small:review(rows,directory,small))for name,small in (('normal',False),('memcheck',True))]
    # No external tiled oracle build: this pilot uses the existing complete
    # native preprocessing implementation as its independent artifact oracle.
    return VECTORS['suite'](c,full_fft=False,arms=arms,scope='assembled genuine s4/s6 GPU preprocessing pilot; distributed root/Newton/identity/spectrum,coordinator discovery/artifact assembly;prefix save/reload,not fresh-process recovery or fulls18')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2 HPC nodes/2 H200/4 CPU/8 GiB host;128 MiB owned/card,1 GiB RSS/rank,128 MiB cache/2 GiB scratch/32 MiB file;720s/15min;2 integrated arms;no isolated kernel calibrations;no fulls18/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
