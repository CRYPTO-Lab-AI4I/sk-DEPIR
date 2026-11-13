#!/usr/bin/env python3
"""Check bounded file-transpose tiling, offsets, and disk evidence admission."""
import runpy
from copy import deepcopy
from pathlib import Path
h=runpy.run_path(str(Path(__file__).with_name('run.py')))
rows=[]
for _ in range(8):
 n,tile=4096,320;st=dict(buffer_bytes=8*tile+65536,transposes=2,read_bytes=16*n,write_bytes=16*n,read_calls=8,write_calls=8)
 rows.append(dict(n=n,dimension=12,tile_elements=tile,matches=3,errors_recovery_canaries=True,backing_bytes=8*n+64,forward=st.copy(),inverse=st.copy()))
h['verify'](rows)
for key,value in [('buffer_bytes',1<<30),('transposes',1),('read_bytes',0)]:
 bad=deepcopy(rows);bad[0]['forward'][key]=value
 try:h['verify'](bad)
 except AssertionError:pass
 else:raise AssertionError('accepted '+key)
for height,width,tile in ((16,8,48),(64,32,320),(256,128,1536),(32,64,320)):
 src=list(range(height*width));dst=[None]*len(src);br=1
 while br*2<=height and br*2<=tile//(br*2):br*=2
 bc=min(width,tile//br)
 for row in range(0,height,br):
  for col in range(0,width,bc):
   nr,nc=min(br,height-row),min(bc,width-col);assert nr*nc<=tile
   for i in range(nr):
    for j in range(nc):
     index=(col+j)*height+row+i;assert dst[index] is None;dst[index]=src[(row+i)*width+col+j]
 assert dst==[src[i*width+j] for j in range(width) for i in range(height)]
# Taylor quotient sources stay in its untouched high half.
for half in (2,8,64):
 for degree in (1,half//2):
  for j in range(degree):assert 3*half//2<=2*half+j-degree<2*half
print('Disk blocked-transpose, dependency and resource-record checks pass')
