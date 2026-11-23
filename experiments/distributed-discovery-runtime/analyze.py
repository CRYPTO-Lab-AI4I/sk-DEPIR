#!/usr/bin/env python3
"""Reconcile coordinated/local support verification without relaxing shard checks."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent;H=runpy.run_path(str(HERE/'run.py'))
p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');a=p.parse_args()
out=HERE/'runs'/a.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
arms={};counts={};native={}
for mode in ('cold','resume','resident-discovery'):
    if not any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):continue
    rows=H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=4)
    if mode=='resident-discovery':arms[mode]=H['resident_review'](rows,cache/mode)
    else:arms[mode]=H['BIND']['review'](rows,cache/mode,mode=='resume',cache/'cold'if mode=='resume'else None,arms.get('cold')if mode=='resume'else None)
    counts[mode]=H['DIAG']['diagnostic_review'](out,mode,cache/mode)
    native[mode]=H['PROOF']['native_review'](out,mode,rows)
if raw['status']=='passed':
    assert raw['composed_vectors']==dict(status='passed',arms=arms)
    assert raw['diagnostic_writes']==counts and raw['native_taylor_totals']==json.loads(json.dumps(native))
    assert arms['cold']['artifacts']==arms['resident-discovery']['artifacts']
sizing={mode:H['SCALE']['COMPACT']['storage'](cache/mode,cases=(8,))for mode in ('cold','resume')if mode in arms}
if raw['status']=='passed':assert sizing==raw['checkpoint_storage']
result.update(arms=arms,diagnostic_writes=counts,native_taylor_totals=native,checkpoint_storage=sizing,error=raw.get('error'),scope='Shared discovery GPU session;sampling and runtime checks retained;not production discovery/full s18')
(out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
