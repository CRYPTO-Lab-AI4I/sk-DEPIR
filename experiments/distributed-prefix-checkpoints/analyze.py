#!/usr/bin/env python3
"""Reconcile complete narrow cold/restart and bounded wide adapter evidence."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent
RUN=runpy.run_path(str(HERE/'run.py'));H=RUN['OWN']
p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');args=p.parse_args()
out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
checked={}
for mode in ('cold','resume','wide','owned-wide'):
    if not any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):continue
    rows=H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=4)
    if mode=='wide':checked[mode]=H['WIDE']['wide_review'](rows,cache/mode)
    elif mode=='owned-wide':checked[mode]=H['review_roots'](rows,cache/mode,json.loads((cache/'root-references.json').read_text()),cache/'root-reference-files',wide=True)
    else:checked[mode]=H['review'](rows,cache/mode,mode=='resume',cache/'cold'if mode=='resume'else None,checked.get('cold'))
if raw['status']=='passed':assert raw['composed_vectors']==dict(status='passed',arms=checked)
sizing={mode:RUN['storage'](cache/mode)for mode in ('cold','resume')if mode in checked}
if raw['status']=='passed':assert sizing==raw['checkpoint_storage']
result.update(checkpoint_storage=sizing,arms=checked,error=raw.get('error'),scope='Compact checkpoint genuine cold/restart and bounded wide regressions;not fulls18.')
(out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
