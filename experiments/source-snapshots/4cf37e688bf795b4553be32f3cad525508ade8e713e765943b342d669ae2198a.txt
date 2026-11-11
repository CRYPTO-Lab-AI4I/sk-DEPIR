#!/usr/bin/env python3
"""Short self-check for rejecting incomplete or incorrect profiling evidence."""
from copy import deepcopy
from pathlib import Path
import runpy
verify=runpy.run_path(str(Path(__file__).with_name('run.py')))['verify_phases']
phases=[]
for worker in (8,16,32):
    for repeat in range(2):
        phases.append({'phase':f'encode-{worker}-{repeat}','workers':worker,'native_stages':dict(calls=5,boundary_calls=1,buffer_ns=1,pointwise_ns=1,assembly_ns=1,taylor_ns=1,columns_transposes_ns=1,rows_ns=1),'native_before':{'steady_ns':0},'native_after':{'steady_ns':10}})
        phases.append({'phase':f'verify-{worker}-{repeat}','full_hash_match':True,'canonical_sha256':'b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c'})
assert len(verify(phases))==6
for kind in ('missing','wrong-hash','invalid-clock'):
    bad=deepcopy(phases)
    if kind=='missing':bad.pop()
    if kind=='wrong-hash':bad[1]['canonical_sha256']='0'*64
    if kind=='invalid-clock':bad[0]['native_stages']['rows_ns']=2_000_000
    try:verify(bad)
    except AssertionError:pass
    else:raise AssertionError('accepted '+kind)
print('Profile evidence self-check passes')
