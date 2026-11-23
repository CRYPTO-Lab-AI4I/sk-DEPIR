#!/usr/bin/env python3
"""Verify eight-worker source/evidence/cache seals, logical outputs and native work."""
from pathlib import Path
import argparse,json,runpy
HERE=Path(__file__).resolve().parent;H=runpy.run_path(str(HERE/'run.py'))
p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');a=p.parse_args();out=HERE/'runs'/a.name
raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache']);check=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))
result=check['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'));arms={};native={}
for mode in ('cold','resume'):
 if not any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):continue
 rows=H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=8)
 arms[mode]=H['review'](rows,cache/mode,8 if mode=='resume'else 1,cache/'cold'if mode=='resume'else None)
 native[mode]=H['PROOF']['native_review'](out,mode,rows,workers=8)
if raw['status']=='passed':assert raw['composed_vectors']==dict(status='passed',arms=arms)and raw['native_taylor_totals']==json.loads(json.dumps(native))
result.update(arms=arms,native_taylor_totals=native,scope='Eight-worker bounded admission validation;not fullsize/eight GPUs');(out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
