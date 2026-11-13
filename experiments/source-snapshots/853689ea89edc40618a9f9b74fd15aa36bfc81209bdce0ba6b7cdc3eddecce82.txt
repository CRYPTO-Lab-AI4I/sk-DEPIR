#!/usr/bin/env python3
"""Short check that input admission rejects mismatched preparation metadata."""
from copy import deepcopy
from pathlib import Path
import runpy
h=runpy.run_path(str(Path(__file__).with_name('run.py')))
m=dict(s=16,t=5,k=918672421,n=1<<32,identity_verified=True,sha256=h['EXPECTED'],bytes=20854558905)
h['validate_metadata'](m,m['bytes'])
for key,value in [('s',15),('n',0),('k',1),('identity_verified',False),('sha256','0'*64),('bytes',1)]:
    bad=deepcopy(m);bad[key]=value
    try:h['validate_metadata'](bad,m['bytes'])
    except AssertionError:pass
    else:raise AssertionError('accepted invalid '+key)
print('Cached-s16 metadata admission self-check passes')

# Recorded-cache mode still checks physical length and metadata/digest agreement.
check=runpy.run_path(str(Path(__file__).with_name('analyze.py')))['check_packed_metadata']
summary={'s':16,'n':1<<32,'packed_sha256':'packed-reference','canonical_sha256':'canonical-reference'}
metadata={'packed_bytes':1<<33,'sha256':'packed-reference','canonical_sha256':'canonical-reference','seed':71}
check(summary,metadata,1<<33)
for key,value in [('packed_bytes',1),('sha256','wrong'),('canonical_sha256','wrong'),('seed',72)]:
    bad=deepcopy(metadata);bad[key]=value
    try:check(summary,bad,1<<33)
    except AssertionError:pass
    else:raise AssertionError('accepted invalid packed '+key)
print('Packed metadata reconciliation self-check passes')


model=h['memory_model'](918672421,1<<32)
assert model['file_principal_bytes']==8*(1<<32)+4*918672421
assert model['resident_principal_bytes']-model['file_principal_bytes']==4*((1<<32)+2*918672421+1)
assert model['logical_code_read_bytes']==20854558905-33
print('CPU file memory/read model passes')

import json,tempfile
phases=json.loads((Path(__file__).resolve().parents[1]/'cached-s16-encoding/runs/encoding-01/encoding/phases.json').read_text())
by={row['phase']:row for row in phases};n,k=by['load-online']['n'],by['load-online']['k'];by['load-online']['plan_bytes']=0
row=by['encode-0'];tile=1<<26
row.update(file_streamed=True,read_ns=0,read_bytes=4*(n+k+1),read_chunks=(k+1+tile-1)//tile+(n+tile-1)//tile,workspace_before_bytes=0,workspace_after_bytes=4*n)
second=deepcopy(row);second.update(phase='encode-1',workspace_before_bytes=4*n)
check=deepcopy(by['verify-0']);check['phase']='verify-1';phases.extend([second,check])
with tempfile.TemporaryDirectory() as d:
    dest=Path(d)
    def verify(rows):
        (dest/'phases.json').write_text(json.dumps(rows));return h['verify_phases'](dest,16,True,2)
    verify(phases)
    for phase,key,value in [('load-online','plan_bytes',1),('encode-0','read_bytes',0),('encode-0','read_chunks',1),('encode-1','workspace_after_bytes',0),('verify-1','canonical_sha256','wrong')]:
        bad=deepcopy(phases);next(row for row in bad if row['phase']==phase)[key]=value
        try:verify(bad)
        except AssertionError:pass
        else:raise AssertionError('accepted invalid '+phase+'/'+key)
print('CPU file ownership/read phase checks pass')

cache=runpy.run_path(str(Path(__file__).with_name('cache_memory.py')))['FileCache']()
with tempfile.TemporaryDirectory() as d:
    file=Path(d)/'cache';file.write_bytes(b'x'*8192)
    first=cache.sample([file]);assert 0<=sum(first.values())<=8192
    file.write_bytes(b'x'*12288);second=cache.sample([file]);assert 0<=sum(second.values())<=12288
cache.close();assert cache.maps=={}
print('Non-faulting file residency observer passes')
