#!/usr/bin/env python3
"""Check complete disk-encoding resource/evidence admission and probe offsets."""
import runpy
from copy import deepcopy
from pathlib import Path
h=runpy.run_path(str(Path(__file__).with_name('run.py')))
rows=[]
for i in range(9):
 s=i+2 if i<7 else (6 if i==7 else 8);n=1<<(2*s)
 rows.append({'case':i,'s':s,'n':n,'matches':3,'errors_recovery_owning':True,'stats':{'calls':5,'boundary_calls':1,'BufferBytes':16*min(n,1536)+2*65536,'BackingBytes':8*n,'OutputBytes':4*n,'ReadBytes':1,'WriteBytes':1}})
h['verify_small'](rows)
for key,value in [('calls',4),('BufferBytes',1<<30),('BackingBytes',0),('OutputBytes',0)]:
 bad=deepcopy(rows);bad[0]['stats'][key]=value
 try:h['verify_small'](bad)
 except AssertionError:pass
 else:raise AssertionError('accepted '+key)
# Code headers are not aligned to uint32; only mutable work-file offsets are.
for header in (25,33):
 k,n=589,4096;p_offset=header;spectrum_offset=header+4*(k+1)
 assert p_offset%4==spectrum_offset%4==1
 assert spectrum_offset+4*n==header+4*(k+1+n)
print('Disk online resource and unaligned artifact-offset checks pass')
