#!/usr/bin/env python3
"""Reconcile canonical discovery results,ownership bounds and matched outputs."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))
p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');a=p.parse_args()
out=HERE/'runs'/a.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
arms={}
for mode in ('cold','resume','resident-discovery'):
    if any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):
        rows=H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=4)
        arms[mode]=H['resident_review'](rows,cache/mode)if mode=='resident-discovery'else H['review'](rows,cache/mode,mode=='resume',cache/'cold'if mode=='resume'else None,arms.get('cold'))
if raw['status']=='passed':assert raw['composed_vectors']==dict(status='passed',arms=arms)
sizing={mode:H['SCALE']['COMPACT']['storage'](cache/mode,cases=(8,))for mode in ('cold','resume')if mode in arms}
if raw['status']=='passed':assert sizing==raw['checkpoint_storage']
comparison={mode:next(c['seconds']for c in raw['commands']if c['label']==mode)for mode in ('cold','resident-discovery')if mode in arms}
if 'resident-discovery'in arms and 'cold'in arms:assert arms['resident-discovery']['artifacts']==arms['cold']['artifacts']
result.update(arms=arms,source_resident_arm_seconds=comparison,checkpoint_storage=sizing,error=raw.get('error'),scope='Explicit discovery-result tiles,genuine s8/curve5 cold/restart and matched resident discovery;spectrum streamed in both;not fulls18')
(out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
