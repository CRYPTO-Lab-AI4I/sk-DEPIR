#!/usr/bin/env python3
"""Reconcile bounded diagnostic files, canonical outputs and retained evidence."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent;H=runpy.run_path(str(HERE/'run.py'))
p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');a=p.parse_args()
out=HERE/'runs'/a.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
CHECK=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))
result=CHECK['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
legacy_wide_bulk=str(raw['job_id'])=='81744'and raw['revision']=='c30ef0d73927bee7ad0f82982e841713976b04b7'
if legacy_wide_bulk:
    # Preserve scheduler/controller failure. Review saved arithmetic independently;
    # this historical wide arm used bulk mode and cannot validate bounded writes.
    assert raw['status']=='failed'
    CHECK['check_commands'](raw['commands'])
arms={};counts={};native={}
for mode in (('wide',)if raw.get('targeted_wide')else('cold','resume','bulk','wide')):
    if not any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):continue
    rows=H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=4)
    if mode=='wide':arms[mode]=H['PROOF']['trace_wide_review'](rows,cache/mode)
    else:arms[mode]=H['review'](rows,cache/mode,mode=='resume',cache/'cold'if mode=='resume'else None,arms.get('cold')if mode=='resume'else None)
    counts[mode]=H['diagnostic_review'](out,mode,cache/mode,mode!='bulk'and not(legacy_wide_bulk and mode=='wide'))
    native[mode]=H['PROOF']['native_review'](out,mode,rows)
if raw['status']=='passed':
    assert raw['composed_vectors']==dict(status='passed',arms=arms)
    assert raw['diagnostic_writes']==counts and raw['native_taylor_totals']==json.loads(json.dumps(native))
    if not raw.get('targeted_wide'):assert arms['cold']['artifacts']==arms['bulk']['artifacts']
sizing={mode:H['SCALE']['COMPACT']['storage'](cache/mode,cases=(8,))for mode in ('cold','resume')if mode in arms}
if raw['status']=='passed'and not raw.get('targeted_wide'):assert sizing==raw['checkpoint_storage']
if legacy_wide_bulk:
    assert json.loads((out/'composition.json').read_text())==dict(status='passed',arms=arms)
    assert arms['cold']['artifacts']==arms['bulk']['artifacts']
    result['partial_validation']={'bounded_cold':'passed','bounded_restart':'passed','bulk_control':'passed','wide':'arithmetic passed in bulk mode; bounded diagnostic validation missing','failure':'Final controller assertion: wide inherited bulk codec from preceding control'}
result.update(arms=arms,diagnostic_writes=counts,native_taylor_totals=native,checkpoint_storage=sizing,error=raw.get('error'),scope='Explicit bounded diagnostic serialization;unchanged bytes and default bulk mode;not fulls18')
(out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
