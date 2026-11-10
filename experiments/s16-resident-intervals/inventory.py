#!/usr/bin/env python3
"""Logical large-operation sizes from the saved-support census, not timings."""
import collections,hashlib,json,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent

def product_blocks(a,b,limit,domain):
    limit=min(limit,a+b-1);a=min(a,limit);b=min(b,limit)
    chunk=domain//2;out=[]
    for i in range(0,min(a,limit),chunk):
        ai=min(chunk,a-i)
        for j in range(0,min(b,limit-i),chunk):
            bj=min(chunk,b-j);full=ai+bj-1;n=1<<(full-1).bit_length()
            out.append({'a':ai,'b':bj,'limit':min(limit-i-j,full),'transform_points':n})
    return out

def inventory(census,cutoff=4096):
    k,n=census['k'],census['n'];counts=collections.Counter();details=[];levels=[]
    def add(a,b,limit,stage):
        # Below the FFT cutoff the original implementation uses CPU arithmetic.
        smallest=min(a,b,limit)
        if smallest<max(24,cutoff):return
        blocks=product_blocks(a,b,limit,n)
        for block in blocks:counts[(stage,block['transform_points'])]+=1
        if stage!='root-carry':details.append({'stage':stage,'a':a,'b':b,'limit':limit,'blocks':blocks})
    for degree in census['batch_degrees']:
        level=0;root=degree
        while level<len(levels) and levels[level] is not None:
            add(levels[level]+1,root+1,levels[level]+root+1,'root-carry');root+=levels[level];levels[level]=None;level+=1
        if level==len(levels):levels.append(root)
        else:levels[level]=root
    root=None
    for degree in levels:
        if degree is None:continue
        if root is None:root=degree
        else:add(root+1,degree+1,root+degree+1,'root-finish');root+=degree
    assert root==k
    precision=1;square_muls=0
    while precision<n-k:
        size=min(2*precision,n-k);square_muls+=precision
        add(min(k+1,size),2*precision-1,size,'reciprocal');precision=size
    add(k+1,n-k,n,'identity')
    sizes=[{'stage':stage,'transform_points':size,'convolutions':count,
            'fft_calls':3*count,'within_current_gpu_cap':size<=1<<22}
           for (stage,size),count in sorted(counts.items())]
    sources={name:hashlib.sha256((ROOT/name).read_bytes()).hexdigest() for name in
        ['encoding/liftrs-shared/orbit_product.go','encoding/liftrs-shared/orbit_polynomial.go','encoding/liftrs-shared/factor_product.go','encoding/liftrs-shared/scaling_experimental.go']}
    return {'k':k,'n':n,'product_cutoff':cutoff,'large_operations':sizes,'noncarry_details':details,
            'base_normalization_muls':k+1,'reciprocal_square_muls':square_muls,
            'final_spectrum_forward_points':n,'source_sha256':sources,
            'estimation_status':'Incomplete: early root intervals and <=2^22 kernels only. No whole-run time/speedup inferred.',
            'excluded':'Internal batch leaf trees, root exponentiation, validation, CPU overhead, memory/GC, persistence and timing extrapolation. Counts are logical operations, not measured cost.',
            'gpu_boundary_work':'Full 2^32 transforms need wider private shape counts, 17 row-basis levels and a newly admitted workspace; current experimental API rejects those sizes.'}

def self_test():
    assert product_blocks(7,7,3,16)==[{'a':3,'b':3,'limit':3,'transform_points':8}]
    assert len(product_blocks(8,10,17,16))==2
    tiny={'k':12,'n':64,'batch_degrees':[4,4,4]}
    r=inventory(tiny,24);assert r['k']==12 and r['final_spectrum_forward_points']==64

if __name__=='__main__':
    self_test()
    if len(sys.argv)!=2 or Path(sys.argv[1]).name!=sys.argv[1]:raise SystemExit('Usage: inventory.py RUN_NAME')
    out=HERE/'runs'/sys.argv[1];census=json.loads((out/'census.json').read_text());result=inventory(census)
    (out/'inventory.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k!='noncarry_details'},indent=2))
