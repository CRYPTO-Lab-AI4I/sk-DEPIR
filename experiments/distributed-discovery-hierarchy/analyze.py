#!/usr/bin/env python3
"""Reconcile sealed source,cache and all discovery/hierarchy proofs after completion."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))
p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');args=p.parse_args()
out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text())
result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
arms={}
if any(c['label']=='normal'and c['exit_code']==0 for c in raw['commands']):
    arms['normal']=H['review'](H['VECTORS']['BASE']['parse_rank_logs'](out/'normal.log',ranks=4),Path(raw['cache'])/'normal')
if raw['status']=='passed':assert raw['composed_vectors']==dict(status='passed',arms=arms)
result.update(arms=arms,error=raw.get('error'),scope='Genuine s4/s6 four-worker discovery and hierarchical roots;not full pipeline/fulls18.')
(out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
