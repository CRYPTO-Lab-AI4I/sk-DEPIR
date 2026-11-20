#!/usr/bin/env python3
"""Reconcile raw identities and every saved GPU vector output after completion."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(HERE/'run.py'))


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');args=p.parse_args()
    out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    checked={}
    for mode in ('normal','memcheck'):
        if raw['status']=='passed' or any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):
            checked[mode]=[HELP['review'](HELP['BASE']['parse'](out/f'{mode}-{rank}.stdout.log'),cache/mode/f'rank-{rank}',rank if mode=='memcheck'and raw.get('resources',{}).get('vector_memcheck_split')else None)for rank in (0,1)]
    summaries=[]
    if 'memcheck'in checked:
        for rank in (0,1):
            summaries.append(HELP['sanitizer_summary']((out/f'memcheck-{rank}.stdout.log').read_text(),
                (out/f'memcheck-{rank}.stderr.log').read_text()))
    if raw['status']=='passed':assert checked==raw['vector_stages']
    component_pass=set(checked)=={'normal','memcheck'} and len(summaries)==2
    if component_pass:assert all(c['exit_code']==0 for c in raw['commands'])
    result.update(controller_status=raw['status'],component_validation='passed'if component_pass else'incomplete',
        sanitizer_summaries=summaries,
        interpretation='Component evidence reviewed independently; raw controller status is preserved. Native profile reports may require separate postprocessing.')
    result.update(tests=['source/evidence/cache identity reconciliation','completed vector-stage arms and output bytes'],vector_stages=checked,error=raw.get('error'),scope='GPU row/column stages only; no distributed FFT, Taylor or fulls18 claim')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
