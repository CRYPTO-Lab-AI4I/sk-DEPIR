#!/usr/bin/env python3
"""Short check that input admission rejects mismatched preparation metadata."""
import json
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

# Validate repeated GPU phase admission using a tiny existing passing workflow.
import tempfile
phases=json.loads((Path(__file__).resolve().parents[1]/'cuda-encoding/runs/validation-01/workflow/cuda/phases.json').read_text())
with tempfile.TemporaryDirectory() as directory:
    out=Path(directory)
    def verify(rows):
        (out/'phases.json').write_text(json.dumps(rows))
        return h['verify_phases'](out,10)
    verify(phases)
    for phase,key,value in [('verify-1','canonical_sha256','wrong'),('verify-0','systematic_checked',0),('pack-output','packed_bytes',1),('encode-0','cuda_stages',{'native_total_ns':0,'elapsed_ns':0,'owned_device_bytes':1<<40})]:
        bad=deepcopy(phases)
        next(row for row in bad if row['phase']==phase)[key]=value
        try:verify(bad)
        except AssertionError:pass
        else:raise AssertionError('accepted invalid '+phase+'/'+key)
print('CUDA repeated-word admission self-check passes')

n,k=1<<32,918672421
baseline={'encoding':{'encodings':[{'cuda_stages':{'owned_device_bytes':55219301804}}]}}
expected=55219301804-4*(n+k+1)
r={'encoding':{'n':n,'k':k,'encodings':[{'cuda_stages':{'owned_device_bytes':expected}}]}}
h['verify_memory'](r,baseline,True)
r['encoding']['encodings'][0]['cuda_stages']['owned_device_bytes']+=4
try:h['verify_memory'](r,baseline,True)
except AssertionError:pass
else:raise AssertionError('accepted wrong streamed memory count')
print('Streamed memory count check passes')

phases=json.loads((Path(__file__).resolve().parents[1]/'cuda-encoding/runs/validation-01/workflow/cuda/phases.json').read_text())
load=next(row for row in phases if row['phase']=='load-online');n,k=load['n'],load['k']
for row in phases:
    if row['phase'].startswith('encode-'):
        row.update(code_upload_bytes=4*(n+k),code_upload_seconds=0,spectrum_chunks=1,stream_tile_bytes=4*n)
with tempfile.TemporaryDirectory() as directory:
    out=Path(directory)
    def verify(rows):
        (out/'phases.json').write_text(json.dumps(rows));return h['verify_phases'](out,10,True)
    verify(phases)
    for key,value in [('code_upload_bytes',4*n),('code_upload_seconds',-1),('code_upload_seconds',100000),('spectrum_chunks',2),('stream_tile_bytes',0)]:
        bad=deepcopy(phases);next(row for row in bad if row['phase']=='encode-0')[key]=value
        try:verify(bad)
        except AssertionError:pass
        else:raise AssertionError('accepted invalid '+key)
print('Streamed transfer/chunk/stage checks pass')
