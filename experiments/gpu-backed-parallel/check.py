#!/usr/bin/env python3
"""Independent stage formulas vs sequential sparse division over GF(256)."""
import random
import runpy
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
runpy.run_path(str(ROOT/'experiments/gpu-host-backed/check.py'))
rng=random.Random(4001)
def mul(a,b):
    result=0
    while b:
        if b&1:result^=a
        a<<=1
        if a&256:a^=0x11b
        b>>=1
    return result
def power(a,n):
    result=1
    while n:
        if n&1:result=mul(result,a)
        a=mul(a,a);n>>=1
    return result
def xor_sum(xs):
    result=0
    for x in xs:result^=x
    return result
def serial(data,terms,c,delta,inverse):
    a=data[:];half=len(a)//2
    if not inverse:
        for i in range(2*half-1,half-1,-1):
            for d,coef in terms:a[i-half+d]^=mul(coef,a[i])
    for j in range(half):
        if not inverse:
            left=a[j]^mul(c,a[half+j]);a[j],a[half+j]=left,left^mul(delta,a[half+j])
        else:
            q=mul(power(delta,254),a[j]^a[half+j]);a[j]^=mul(c,q);a[half+j]=q
    if inverse:
        for j in range(half):
            q=a[half+j]
            for d,coef in terms:a[j+d]^=mul(coef,q)
    return a

def parallel(data,terms,c,delta,inverse):
    half=len(data)//2;a=data[:];out=[0]*len(a)
    if not inverse:
        # Joined quotient kernel only reads unchanged upper-half sources.
        for j in range(half):a[half+j]=data[half+j]^xor_sum(mul(coef,data[2*half+j-d]) for d,coef in terms if d>j)
        for j in range(half):
            remainder=a[j]^xor_sum(mul(coef,a[half+j-d]) for d,coef in terms if d<=j)
            left=remainder^mul(c,a[half+j]);out[j],out[half+j]=left,left^mul(delta,a[half+j])
    else:
        for j in range(half):
            a[half+j]=mul(power(delta,254),data[j]^data[half+j]);a[j]=data[j]^mul(c,a[half+j])
        for j in range(half):
            out[j]=a[j]^xor_sum(mul(coef,a[half+j-d]) for d,coef in terms if d<=j)
            out[half+j]=a[half+j]^xor_sum(mul(coef,a[2*half+j-d]) for d,coef in terms if d>j)
    return out
for half in (1,2,4,8,32):
    for repeat in range(12):
        data=[rng.randrange(256) for _ in range(2*half)];terms=[(1<<k,rng.randrange(256)) for k in range(half.bit_length()-1)]
        c,delta=rng.randrange(256),rng.randrange(1,256)
        want=serial(data,terms,c,delta,False);assert parallel(data,terms,c,delta,False)==want
        assert parallel(want,terms,c,delta,True)==serial(want,terms,c,delta,True)==data
# Pair mapping covers every element once in both layouts, including batch tails.
for length in (2,8,32):
    for vectors in (1,3,32):
        for k in range(length.bit_length()-1):
            half=1<<k
            for contiguous in (False,True):
                seen=[]
                for i in range(length*vectors//2):
                    vector,pair=(i//(length//2),i%(length//2)) if contiguous else (i%vectors,i//vectors)
                    group,j=(pair//half)*2*half,pair%half
                    for coeff in (group+j,group+half+j):seen.append(vector*length+coeff if contiguous else coeff*vectors+vector)
                assert sorted(seen)==list(range(length*vectors))
h=runpy.run_path(str(Path(__file__).with_name('run.py')))
from copy import deepcopy
row={'s':4,'dimension':8,'n':256,'tile_elements':48,'input_sha256':'input','forward_sha256':'output','resident_owned_bytes':4096,'backed_owned_bytes':2048,'forward_seconds':1.0,'forward_stages':[0.0]*12,'inverse_stages':[0.0]*12}
rows=[deepcopy(row) for _ in range(8)];previous=deepcopy(rows)
assert h['compare'](rows,previous)[0]['new_over_old']==1
for key in ('input_sha256','backed_owned_bytes'):
    bad=deepcopy(rows);bad[0][key]='changed'
    try:h['compare'](bad,previous)
    except AssertionError:pass
    else:raise AssertionError('accepted changed '+key)
bad=deepcopy(rows);bad[0]['forward_stages'][8]=1
try:h['compare'](bad,previous)
except AssertionError:pass
else:raise AssertionError('accepted changed transfers')
print('Parallel sparse stage forward/inverse, pair-layout and comparison checks pass')
