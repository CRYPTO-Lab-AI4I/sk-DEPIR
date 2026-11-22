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


def checkpoint_bytes(directory,label,stage,commit):
    """Expand a bounded diagnostic checkpoint,checking exact stored intervals/hashes."""
    n,ranks,prefix=commit['n'],commit['ranks'],commit['prefix'];layout=commit.get('layout','')
    assert layout in ('','prefix')and 0<prefix<=n and 1<=ranks<=64 and len(commit['sha256'])==ranks
    result=[]
    for rank in range(ranks):
        quotient,remainder=divmod(n,ranks);first=quotient*rank+min(rank,remainder);words=quotient+(rank<remainder)
        stored=min(words,max(0,prefix-first))if layout=='prefix'else words
        data=(directory/f'{label}-rank-{rank}-{stage}.bin').read_bytes()
        assert len(data)==4*stored and hashlib.sha256(data).hexdigest()==commit['sha256'][rank]
        result.append(data+b'\0'*(4*(words-stored)))
    return b''.join(result)


def artifacts(directory,s,k,initial=1,ranks=2,curve=2):
    n=1<<(2*s);label=f's{s}'
    def joined(stage):return b''.join((directory/f'{label}-rank-{r}-{stage}.bin').read_bytes()for r in range(ranks))
    reference={stage:(directory/f'{label}-rank-0-reference-{stage}.bin').read_bytes()for stage in ('p','g','spectrum')}
    for stage,count in (('p',k+1),('g',n-k),('spectrum',n)):
        data=joined(stage);assert len(data)==4*n and data==reference[stage]+b'\0'*(4*(n-count))
    assert joined('identity')==(1).to_bytes(4,'little')+b'\0'*(4*(n-2))+(1).to_bytes(4,'little')
    support=[]
    for rank in range(ranks):
        _,indices=read_words(directory/f'{label}-rank-{rank}-support.bin',8)
        assert indices==sorted(set(indices))and not set(support).intersection(indices);support.extend(indices)
    assert len(support)==k
    binding=hashlib.sha256((directory/f'{label}-support.bin').read_bytes()).hexdigest()
    for stage,count,source in [('root',k+1,reference['p'])]+[(f'newton-{size}',size,reference['g'][:4*size])for size in prefixes(n-k)if size>=initial]:
        commit=json.loads((directory/f'{label}-{stage}-commit.json').read_text())
        expected=dict(s=s,k=k,n=n,prefix=count,ranks=ranks,support_sha256=binding,stage=stage,sha256=[hashlib.sha256((directory/f'{label}-rank-{r}-{stage}.bin').read_bytes()).hexdigest()for r in range(ranks)])
        if 'layout'in commit:expected['layout']='prefix'
        assert commit==expected
        assert checkpoint_bytes(directory,label,stage,commit)==source+b'\0'*(4*(n-count))
    actual=(directory/f'{label}-code.bin').read_bytes();expected=(directory/f'{label}-reference-code.bin').read_bytes()
    assert actual==expected and actual[:9]==b'LRSprobe1'and len(actual)==25+4*(k+1+n)
    assert [int.from_bytes(actual[i:i+4],'little')for i in (9,13,17,21)]==[s,curve,k,n]
    return dict(code_sha256=hashlib.sha256(actual).hexdigest(),checkpoint_prefixes=[size for size in prefixes(n-k)if size>=initial])


def packet_key(p):return tuple(p[k]for k in ('Operation','Domain','Operand','Offset','Words','Source','Target'))


def review(rows,directory,small=False,initial=1,streamed=False,ranks=2,hierarchy=False,cases=None,curve=2):
    assert ranks in (2,4)and(ranks==2 or hierarchy)
    cases=([4]if small else[4,6])if cases is None else cases;values=[r for r in rows if r.get('kind')=='pipeline'];assert len(values)==ranks*len(cases)
    hosts=[r for r in rows if r.get('kind')=='host'];assert len(hosts)==ranks and len({r['host']for r in hosts})==2
    assert all(r['visible_devices']and ','not in r['visible_devices']for r in hosts)
    assert {r['rank']for r in hosts}==set(range(ranks))
    if ranks==4:
        for host in {r['host']for r in hosts}:
            local=[r for r in hosts if r['host']==host]
            assert len(local)==2 and len({r['visible_devices']for r in local})==1
    calls=[r for r in rows if r.get('kind')=='executor-call'];expected_calls={};specs=[];transpose={};operand={};result={}
    def operand_packet(label,op,n,which,offset,count,source,target):
        for rank,direction in ((source,'send'),(target,'receive')):operand[(label,rank,direction,(op,n,which,offset,count,source,target))]=True
    for s in cases:
        n=1<<(2*s);cols=1<<s;span=2+2*s;capacity=min(n,3*cols);label=f's{s}'
        selected=[r for r in values if r['s']==s];assert len(selected)==ranks and {r['rank']for r in selected}==set(range(ranks))
        k=selected[0]['k'];ac,bc=selected[0]['root_a_count'],selected[0]['root_b_count'];assert 1<=k<=n//2 and (hierarchy or ac+bc==k+2)
        for r in selected:
            assert (r['n'],r['k'],r['first'],r['words'],r['root_a_count'],r['root_b_count'])==(n,k,r['rank']*n//ranks,n//ranks,ac,bc)
            assert all(r[x]is True for x in ('matches','identity_verified','prefix_reloaded'))and 0<r['owned_device_bytes']<=128<<20 and r['peak_rss_bytes']<1<<30
        counts=selected[0].get('root_counts',[ac,bc])
        assert len(counts)==ranks and sum(counts)==k+ranks
        for rank,count in enumerate(counts):
            assert (directory/f'{label}-rank-{rank}-support.bin').stat().st_size==8*(count-1)
            assert all(r.get('root_counts',counts)==counts for r in selected)
            if initial==1 and not hierarchy:
                target=1-rank;first=target*n//ranks;end=min(first+n//ranks,count)
                for offset in range(first,end,4096):operand_packet(label,1,n,rank,offset,min(4096,end-offset),rank,target)
        operations=[];op=1
        if initial==1:
            if hierarchy:
                domain=1
                while domain<k+1 or domain<ranks*ranks:domain*=2
                for source in range(ranks):
                    for target in range(ranks):
                        if source==target:continue
                        lo=max(source*domain//ranks,target*n//ranks);hi=min((source+1)*domain//ranks,(target+1)*n//ranks,k+1)
                        for offset in range(lo,hi,4096):operand_packet(label,1,n,0,offset,min(4096,hi-offset),source,target)
                reshape=[r for r in rows if r.get('kind')=='root-reshape'and r['case']==label]
                assert len(reshape)==ranks and all(r['matches']is True and r['source']==dict(First=0,Ranks=ranks,Domain=domain,Count=k+1)and r['domain']==n and r['operation']==1 for r in reshape)
                op=2
            else:operations.append((label+'/root','product',op,ac,bc,k+1));op+=3*span
        known=initial
        for size in (size for size in prefixes(n-k)if size>initial):
            for source in range(ranks):
                for target in range(ranks):
                    if source==target:continue
                    first=max(source*n//ranks,(target*n//ranks+1)//2);end=min((source+1)*n//ranks,((target+1)*n//ranks+1)//2,(size+1)//2)
                    for offset in range(first,end,4096):operand_packet(label,op,n,0,offset,min(4096,end-offset),source,target)
            op+=1;operations.append((label+f'/newton-{size}','product',op,min(k+1,size),min(2*known-1,size),size));op+=3*span;known=size
        operations.append((label+'/identity','product',op,k+1,n-k,n));op+=3*span
        operations.append((label+'/spectrum','spectrum',op,None,None,n));op+=span
        if streamed:
            for count in (k+1,n-k,n):
                for source in range(1,ranks):
                    for offset in range(source*n//ranks,min((source+1)*n//ranks,count),4096):operand_packet(label,op,n,0,offset,min(4096,min((source+1)*n//ranks,count)-offset),source,0)
                op+=1
        else:
            assert ranks==2
            for _ in range(3):operand_packet(label,op,n,0,n//2,n//2,1,0);op+=1
        for name,mode,start,a,b,keep in operations:
            for rank in range(ranks):expected_calls[name,rank]=(s,start,mode,a,b,keep)
            for child in range(3 if mode=='product'else 1):
                base=start+child*span;inverse=child==2;specs.append((name,child,n,cols,base,inverse,capacity));matrix=base+(0 if inverse else 2*s)
                for rank in range(ranks):
                    for transpose_op in (matrix,matrix+1):
                        for key,words in VECTORS['transpose_packets'](rank,name,child,transpose_op,n,4096,ranks):transpose[key]=words
        result[label]=artifacts(directory,s,k,initial,ranks,curve)
    actual_calls={}
    for r in calls:
        key=(r['case'],r['rank']);assert key not in actual_calls and r['matches']is True
        actual_calls[key]=(r['s'],r['operation_start'],r['mode'],r.get('a_count'),r.get('b_count'),r['keep'])
    assert actual_calls==expected_calls
    VECTORS['review_transpose_routes'](rows,transpose,ranks);FFT['review_taylor_routes'](rows,specs,ranks)
    proofs={}
    for r in rows:
        if r.get('kind')not in ('operand-proof','support-proof'):continue
        own,remote=bytes.fromhex(r['own_proof']),bytes.fromhex(r['remote_proof'])
        assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]
        if r['kind']=='operand-proof':
            key=(r['case'],r['rank'],r['direction'],packet_key(r['packet']));assert key not in proofs;proofs[key]=r
    assert set(proofs)==set(operand)
    for key,r in proofs.items():
        label,rank,direction,packet=key;peer=proofs[label,packet[6]if direction=='send'else packet[5],'receive'if direction=='send'else'send',packet]
        assert r['own_proof']==peer['remote_proof']and r['remote_proof']==peer['own_proof']
    for s in cases:
        label=f's{s}';checks=[r for r in rows if r.get('kind')=='checkpoint'and r['case']==label]
        assert sum(r['reloaded']for r in checks)==(ranks if initial==1 else 0)
    assert all(sum(r.get('kind')=='completed'and r['rank']==rank and r['cases']==len(cases)for r in rows)==1 for rank in range(ranks))
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
