#!/usr/bin/env python3
"""Synthetic record checks for component memory, timing and layout gates."""
from copy import deepcopy
from pathlib import Path
import json,runpy,tempfile
h=runpy.run_path(str(Path(__file__).with_name('run.py')))
rows=[]
for dim in (8,9):
    for mode in ('fast','direct'):
        n=1<<dim;work=n*4*(2 if mode=='fast' else 1)
        rows.append(dict(dimension=dim,n=n,layout=mode,owned_bytes=work+1016,table_bytes=1000,working_array_bytes=work,matches=True,closed=True,
            forward_seconds=[.01],inverse_seconds=[.01],forward_stages=[[0,0,0,0,0,0,0,.009]],inverse_stages=[[0,0,0,0,0,0,0,.009]],input_sha256='input',reference_sha256='reference'))
with tempfile.TemporaryDirectory() as directory:
    out=Path(directory)
    def verify(value):
        (out/'transforms.json').write_text(json.dumps(value));return h['verify_transforms'](out,'compare',True)
    verify(rows)
    for key,value in [('owned_bytes',1<<40),('n',1),('matches',False),('working_array_bytes',1),('reference_sha256','wrong')]:
        bad=deepcopy(rows);bad[1][key]=value
        try:verify(bad)
        except AssertionError:pass
        else:raise AssertionError('accepted invalid '+key)
    bad=deepcopy(rows);bad[1]['forward_stages'][0][4]=.1
    try:verify(bad)
    except AssertionError:pass
    else:raise AssertionError('direct layout performed transpose')
print('Transform record, cap, scratch saving and transpose checks pass')
