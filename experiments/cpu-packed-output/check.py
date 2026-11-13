#!/usr/bin/env python3
"""Check atomic byte packing bounds and experiment evidence admission."""
import runpy
from copy import deepcopy
from pathlib import Path
h=runpy.run_path(str(Path(__file__).with_name('run.py')))
rows=[]
for case in range(9):
 s=case+2 if case<7 else (6 if case==7 else 8);n=1<<(2*s)
 for workers in (1,16):rows.append({'case':case,'s':s,'workers':workers,'matches':3,'writer_errors_recovery_owning':True,'stats':{'WorkspaceBytes':8*n,'WrittenBytes':(n*s+7)//8,'BufferBytes':257,'SHA256':'a'*64}})
h['verify_small'](rows)
for key,value in [('WorkspaceBytes',0),('WrittenBytes',0),('BufferBytes',258)]:
 bad=deepcopy(rows);bad[0]['stats'][key]=value
 try:h['verify_small'](bad)
 except AssertionError:pass
 else:raise AssertionError('accepted '+key)
# All bit fields stay inside the initialized packed prefix, including crossings.
for s in range(2,17):
 n=256;out=bytearray((n*s+7)//8)
 for point in range(n):
  value=(point*17+3)&((1<<s)-1);bit=point*s;first,shift=divmod(bit,8);bits=value<<shift
  for j in range((shift+s+7)//8):out[first+j]|=(bits>>(8*j))&255
 for point in range(n):
  bit=point*s;first,shift=divmod(bit,8);raw=sum(out[first+j]<<(8*j) for j in range((shift+s+7)//8));assert (raw>>shift)&((1<<s)-1)==(point*17+3)&((1<<s)-1)
print('CPU packed byte boundaries and resource-record checks pass')
