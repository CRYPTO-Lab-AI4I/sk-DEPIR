#!/usr/bin/env python3
"""Wide assembled adapters plus existing genuine four-worker cold/restart."""
from pathlib import Path
import hashlib
import runpy
import sys
HERE=Path(__file__).resolve().parent
FOUR=runpy.run_path(str(HERE.parent/'distributed-pipeline-four/run.py'))
PILOT=FOUR['PILOT'];VECTORS=FOUR['VECTORS'];FFT=PILOT['FFT']


def multiply(a,b):
    """Independent shift/XOR base-field oracle;canonical polynomial0x40081."""
    value=0
    while b:
        if b&1:value^=a
        b>>=1;a<<=1
        if a&(1<<18):a^=0x40081
    return value


def reciprocal(count,target):
    p=[1]+[((j+1)*0x1337)&((1<<18)-1)for j in range(1,count)]
    g=[1]+[0]*(target-1)
    for k in range(1,target):
        for j in range(1,min(count,k+1)):g[k]^=multiply(p[j],g[k-j])
    return p,g


def review_streams(directory,label,n,target,generator,spectrum,high):
    """Require byte-exact widths and high-bit retention at the shared stream boundary."""
    assert len(generator)==4*n and len(spectrum)==len(high)==8*n
    pattern=b''.join((((1<<35)^((i+1)*0x9e3779b9))&((1<<36)-1)).to_bytes(8,'little')for i in range(n))
    assert high==pattern and any(int.from_bytes(high[i:i+8],'little')>0xffffffff for i in range(0,len(high),8))
    for suffix,data in [('generator',generator[:4*target]),('spectrum',spectrum),('high',high)]:
        assert (directory/f'{label}-stream-{suffix}.bin').read_bytes()==data


def wide_review(rows,directory):
    hosts=[r for r in rows if r.get('kind')=='host'];assert len(hosts)==4 and {r['rank']for r in hosts}==set(range(4))and len({r['host']for r in hosts})==2
    for host in {r['host']for r in hosts}:
        local=[r for r in hosts if r['host']==host];assert len(local)==2 and len({r['visible_devices']for r in local})==1
    values=[r for r in rows if r.get('kind')=='wide-tail'];assert len(values)==8
    specs=[];transpose={};calls={};packets={};payload={}
    def wire(label,width,op,n,offset,count,source,target,data=None):
        p=(op,n,0,offset,count,source,target)
        for rank,direction in ((source,'send'),(target,'receive')):
            key=(label,rank,direction,width,p);assert key not in packets;packets[key]=True
            if data is not None:payload[key]=hashlib.sha256(data).digest()
    def words(data,width):assert len(data)%width==0;return[int.from_bytes(data[i:i+width],'little')for i in range(0,len(data),width)]
    for dim in (5,12):
        label=f's18-d{dim}';n=1<<dim;cols=1<<(dim//2);span=2+2*(dim-dim//2);capacity=min(n,3*max(cols,n//cols));target=16 if dim==5 else 64;count=9 if dim==5 else 17
        chosen=[r for r in values if r['dimension']==dim];assert len(chosen)==4 and {r['rank']for r in chosen}==set(range(4))
        for r in chosen:
            assert (r['s'],r['n'],r['target'],r['root_count'],r['first'],r['words'],r['slot_bytes'],r['coefficient_bytes'])==(18,n,target,count,r['rank']*n//4,n//4,8,4)
            assert r['matches']is True and 0<r['owned_device_bytes']<=64<<20 and r['peak_rss_bytes']<1<<30
        root,coefficients=reciprocal(count,target)
        op=1;known=1;operations=[]
        for size in PILOT['prefixes'](target):
            for source in range(4):
                for dest in range(4):
                    if source==dest:continue
                    lo=max(source*n//4,(dest*n//4+1)//2);hi=min((source+1)*n//4,((dest+1)*n//4+1)//2,(size+1)//2)
                    for offset in range(lo,hi,4096):
                        take=min(4096,hi-offset);data=b''.join(multiply(coefficients[i],coefficients[i]).to_bytes(4,'little')for i in range(offset,offset+take));wire(label,4,op,n,offset,take,source,dest,data)
            op+=1;operations.append((label+f'/newton-{size}','product',op,min(count,size),min(2*known-1,size),size));op+=3*span;known=size
            full=b''.join((directory/f'{label}-rank-{rank}-prefix-{size}.bin').read_bytes()for rank in range(4));assert full==b''.join(v.to_bytes(4,'little')for v in coefficients[:size])+b'\0'*(4*(n-size))
        operations.append((label+'/identity','product',op,count,target,count+target-1));op+=3*span
        convolution=[0]*(count+target-1)
        for i,a in enumerate(root):
            for j,b in enumerate(coefficients):convolution[i+j]^=multiply(a,b)
        assert convolution[:target]==[1]+[0]*(target-1)
        identity=b''.join((directory/f'{label}-rank-{rank}-identity.bin').read_bytes()for rank in range(4))
        assert identity==b''.join(v.to_bytes(4,'little')for v in convolution)+b'\0'*(4*(n-len(convolution)))
        operations.append((label+'/spectrum','spectrum',op,None,None,n));op+=span
        for name,mode,start,ac,bc,keep in operations:
            for rank in range(4):calls[(name,rank)]=(18,dim,start,mode,ac,bc,keep)
            for child in range(3 if mode=='product'else 1):
                start_op=start+child*span;inverse=child==2;specs.append((name,child,n,cols,start_op,inverse,capacity));matrix=start_op+(0 if inverse else 2*(dim-dim//2))
                for rank in range(4):
                    for operation in (matrix,matrix+1):
                        for key,count_words in VECTORS['transpose_packets'](rank,name,child,operation,n,4096,4):transpose[key]=count_words
        generator=b''.join((directory/f'{label}-rank-{rank}-generator.bin').read_bytes()for rank in range(4));assert len(generator)==4*n
        final=b''.join((directory/f'{label}-rank-{rank}-prefix-{target}.bin').read_bytes()for rank in range(4));assert final==generator
        spectrum=b''.join((directory/f'{label}-rank-{rank}-spectrum.bin').read_bytes()for rank in range(4));expected=b''.join((directory/f'{label}-rank-{rank}-expected-spectrum.bin').read_bytes()for rank in range(4));assert spectrum==expected and len(spectrum)==8*n and all(v<1<<36 for v in words(spectrum,8))
        high=b''.join((directory/f'{label}-rank-{rank}-high.bin').read_bytes()for rank in range(4))
        review_streams(directory,label,n,target,generator,spectrum,high)
        for width,count_words,data in ((4,target,generator),(8,n,spectrum),(8,n,high)):
            for source in range(1,4):
                for offset in range(source*n//4,min((source+1)*n//4,count_words),4096):
                    take=min(4096,min((source+1)*n//4,count_words)-offset);wire(label,width,op,n,offset,take,source,0,data[width*offset:width*(offset+take)])
            op+=1
    observed={}
    for r in rows:
        if r.get('kind')=='tail-call':
            key=(r['case'],r['rank']);assert key not in observed;observed[key]=(r['s'],r['dimension'],r['operation_start'],r['mode'],r.get('a_count'),r.get('b_count'),r['keep'])
    assert observed==calls
    VECTORS['review_transpose_routes'](rows,transpose,4);FFT['review_taylor_routes'](rows,specs,4)
    actual={}
    for r in rows:
        if r.get('kind')!='operand-proof':continue
        key=(r['case'],r['rank'],r['direction'],r['word_bytes'],PILOT['packet_key'](r['packet']));assert key not in actual
        own,remote=bytes.fromhex(r['own_proof']),bytes.fromhex(r['remote_proof']);assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]
        if key in payload:assert own[:32]==payload[key]
        actual[key]=r
    assert set(actual)==set(packets)
    for key,r in actual.items():
        label,rank,direction,width,p=key;peer=p[6]if direction=='send'else p[5];other=actual[label,peer,'receive'if direction=='send'else'send',width,p]
        assert r['own_proof']==other['remote_proof']and r['remote_proof']==other['own_proof']
    assert sum(r.get('kind')=='completed'and r['cases']==2 for r in rows)==4
    return dict(measurements=values,executor_calls=len(observed),operand_proofs=len(actual),scope='Bounded canonical field18 reciprocal/product/spectrum and high-bit raw publication;not a fulls18 grid/code artifact.')


def suite(c):
    c['run']('wide-pipeline-review-tests',['python3','-B','experiments/distributed-pipeline-wide/test_review.py'],30)
    extra=[dict(name='wide',test='TestWidePipelineTail',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM=''),review=wide_review)]
    result=FOUR['suite'](c,extra_arms=extra)
    c['resources']['scope']='Genuine s4/s6 four-worker cold/restart plus bounded canonical field18 adapter tail;not fulls18'
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min;genuine cold/resume plus field18 bounded32/4096 tail;no CUDA/default/resource increase;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
