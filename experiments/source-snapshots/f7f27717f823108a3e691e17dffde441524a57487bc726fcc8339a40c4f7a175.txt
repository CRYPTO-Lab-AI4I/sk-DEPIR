#!/usr/bin/env python3
"""Self-check resource and evidence admission for complete host-backed runs."""
import runpy
from pathlib import Path
from copy import deepcopy
h=runpy.run_path(str(Path(__file__).with_name('run.py')))
rows=[{'s':s,'rejected_prefix':True} for s in (2,3,4)]
for s in (2,3,4,6,8):
 for tile in (256,1536):rows.append({'s':s,'tile':tile,'matches':3,'errors_recovery_close':True,'stats':{'transforms':5,'owned_device_bytes':1024,'host_arena_bytes':4*(1<<(2*s)),'host_stage_bytes':4*tile,'host_to_device_bytes':1,'device_to_host_bytes':1}})
h['verify_host'](rows)
for key,value in [('transforms',4),('host_arena_bytes',0),('host_stage_bytes',1<<30)]:
 bad=deepcopy(rows);bad[-1]['stats'][key]=value
 try:h['verify_host'](bad)
 except AssertionError:pass
 else:raise AssertionError('accepted '+key)
print('Host-backed full-encoding evidence/resource checks pass')
