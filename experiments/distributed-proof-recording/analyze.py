#!/usr/bin/env python3
"""Reconcile bounded summaries against independently reviewed full trace peers."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent;H=runpy.run_path(str(HERE/'run.py'))
p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');a=p.parse_args()
out=HERE/'runs'/a.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
arms={};logs={};native={}
for mode in ('cold','resume','summary-cold','summary-resume','wide','summary-wide'):
    if not any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):continue
    rows=H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=4);logs[mode]=rows
    if mode.startswith('summary-'):
        reference=mode.removeprefix('summary-')
        arms[mode]=H['summary_review'](rows,cache/mode,logs[reference],reference=='resume',cache/'cold'if reference=='resume'else None,arms.get('cold')if reference=='resume'else None,reference=='wide')
    elif mode=='wide':arms[mode]=H['trace_wide_review'](rows,cache/mode)
    else:arms[mode]=H['review'](rows,cache/mode,mode=='resume',cache/'cold'if mode=='resume'else None,arms.get('cold'))
    native[mode]=H['native_review'](out,mode,rows)
if raw['status']=='passed':
    assert raw['composed_vectors']==dict(status='passed',arms=arms)
    assert raw['native_taylor_totals']==json.loads(json.dumps(native))
sizing={mode:H['SCALE']['COMPACT']['storage'](cache/mode,cases=(8,))for mode in ('cold','resume')if mode in arms}
if raw['status']=='passed':assert sizing==raw['checkpoint_storage']
for base in ('cold','resume','wide'):
    if base in arms and 'summary-'+base in arms:assert arms[base]['artifacts']==arms['summary-'+base]['artifacts']if base!='wide'else arms[base]['executor_calls']==arms['summary-'+base]['executor_calls']
result.update(arms=arms,native_taylor_totals=native,checkpoint_storage=sizing,error=raw.get('error'),scope='Bounded recording vs full trace;runtime checks retained;not fulls18')
(out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
