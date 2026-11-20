#!/usr/bin/env python3
"""Reconcile every matched arm and summarize samples without extrapolation."""
from pathlib import Path
import argparse
import json
import runpy
import statistics
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');args=p.parse_args()
    out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    checked={};groups={}
    for spec in H['specs']():
        mode=spec['name']
        if not any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):continue
        arm=H['review'](H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=2),cache/mode,spec)
        if spec['memcheck']:
            arm['sanitizer']=[H['VECTORS']['STAGE']['sanitizer_summary']((out/f'{mode}-{r}.stdout.log').read_text(),(out/f'{mode}-{r}.stderr.log').read_text())for r in (0,1)]
        else:
            for row in arm['measurements']:
                if row['rank']!=0:continue
                key=(row['dimension'],row['shape'],spec['tile'],spec['backend'])
                groups.setdefault(key,[]).append(dict(seconds=row['seconds'],work_seconds=row['work_seconds'],validation_capture_seconds=row['validation_capture_seconds']))
        checked[mode]=arm
    if raw['status']=='passed':assert raw['composed_vectors']==dict(status='passed',arms=checked)
    result.update(arms=checked,samples=[dict(dimension=k[0],values=1<<k[0],shape=k[1],tile_words=k[2],backend=k[3],count=len(v),samples=v,median_seconds=statistics.median(s['seconds']for s in v),median_work_seconds=statistics.median(s['work_seconds']for s in v),median_capture_seconds=statistics.median(s['validation_capture_seconds']for s in v),minimum_work_seconds=min(s['work_seconds']for s in v),maximum_work_seconds=max(s['work_seconds']for s in v))for k,v in sorted(groups.items())],error=raw.get('error'),scope='Synthetic s17-field products N4096/16384 only; no fulls17/s18 preprocessing or fullsize speedup estimate. Sanitizer coverage is N4096/tile4096 only.')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
