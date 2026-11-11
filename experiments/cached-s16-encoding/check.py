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
