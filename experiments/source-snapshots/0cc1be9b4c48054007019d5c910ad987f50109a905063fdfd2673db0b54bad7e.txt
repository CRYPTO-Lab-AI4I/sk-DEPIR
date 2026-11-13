#!/usr/bin/env python3
import runpy
from pathlib import Path
h=runpy.run_path(str(Path(__file__).with_name('run.py')))
rows=[]
for i in range(8):
 n,tile=4096,320;st=[0.0]*12;st[8]=st[9]=8*n;st[10]=4*tile;st[11]=26
 rows.append(dict(n=n,dimension=12,tile_elements=tile,matches=3,closed=True,recovery=True,resident_owned_bytes=8*n+100,backed_owned_bytes=8*tile+100,forward_stages=st[:],inverse_stages=st[:]))
h['verify'](rows)
from copy import deepcopy
for key,value in [('matches',0),('closed',False),('backed_owned_bytes',1)]:
 bad=deepcopy(rows);bad[0][key]=value
 try:h['verify'](bad)
 except AssertionError:pass
 else:raise AssertionError('accepted '+key)
# Dependency proof used by tiled quotient: sources stay in unchanged high half.
for half in (2,8,64,1024):
 for d in (1,half//2):
  for j in range(d):assert 3*half//2<=2*half+j-d<2*half
print('Host-backed metadata/allocation/traffic and quotient-dependency checks pass')
