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
assert model['old_principal_bytes']-model['reuse_principal_bytes']==16*(1<<30)
assert model['reuse_principal_bytes']==12*(1<<32)+12*918672421+4
print('CPU memory model check passes')

# Synthetic phase metadata tests ownership transfer and repeat rejection without encoding.
import json,tempfile
phases=json.loads((Path(__file__).resolve().parents[1]/'cached-s16-encoding/runs/encoding-01/encoding/phases.json').read_text())
by={r['phase']:r for r in phases};n,k=by['load-online']['n'],by['load-online']['k']
workspace=(n+k)*4
enc=by['encode-0'];enc.update(workspace_before_bytes=workspace+4*n,workspace_after_bytes=workspace,scratch_transferred=True)
second=deepcopy(enc);second['phase']='encode-1';second['workspace_before_bytes']=workspace
verify=deepcopy(by['verify-0']);verify['phase']='verify-1'
phases.extend([second,verify])
with tempfile.TemporaryDirectory() as directory:
    dest=Path(directory)
    def check(rows):
        (dest/'phases.json').write_text(json.dumps(rows))
        return h['verify_phases'](dest,16,True)
    check(phases)
    for phase,key,value in [('encode-0','scratch_transferred',False),('encode-1','workspace_after_bytes',workspace+4*n),('verify-1','canonical_sha256','wrong')]:
        bad=deepcopy(phases);next(row for row in bad if row['phase']==phase)[key]=value
        try:check(bad)
        except AssertionError:pass
        else:raise AssertionError('accepted invalid '+phase+'/'+key)
print('CPU ownership/repeat metadata checks pass')
