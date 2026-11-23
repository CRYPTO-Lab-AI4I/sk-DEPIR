#!/usr/bin/env python3
"""Reconcile the parameterized command's source, persisted bytes and native work."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent;H=runpy.run_path(str(HERE/'run.py'))
p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');a=p.parse_args()
out=HERE/'runs'/a.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
CHECK=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))
result=CHECK['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
arms={};native={}
for mode in ('cold','resume','command-cold','command-resume'):
    if not any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):continue
    rows=H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=4)
    if mode.startswith('command-'):
        known=8 if mode=='command-resume'else 1;baseline='resume'if known==8 else'cold'
        arms[mode]=H['command_review'](rows,cache/mode,H['VECTORS']['BASE']['parse_rank_logs'](out/(baseline+'.log'),ranks=4),cache/baseline,known,cache/'command-cold'if known==8 else None)
    else:arms[mode]=H['BIND']['review'](rows,cache/mode,mode=='resume',cache/'cold'if mode=='resume'else None,arms.get('cold')if mode=='resume'else None)
    native[mode]=H['PROOF']['native_review'](out,mode,rows)
if raw['status']=='passed':
    assert raw['composed_vectors']==dict(status='passed',arms=arms)
    assert raw['native_taylor_totals']==json.loads(json.dumps(native))
sizing={mode:H['SCALE']['COMPACT']['storage'](cache/mode,cases=(8,))for mode in ('cold','resume')if mode in arms}
if raw['status']=='passed':assert sizing==raw['checkpoint_storage']
result.update(arms=arms,native_taylor_totals=native,checkpoint_storage=sizing,error=raw.get('error'),scope='Parameterized command small cold/restart;not fullsize phase/topology admission')
(out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
