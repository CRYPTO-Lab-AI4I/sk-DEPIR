#!/usr/bin/env python3
"""Review local Taylor guards, remote routes and repeated product timings."""
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
        if spec['memcheck']:arm['sanitizer']=[H['VECTORS']['STAGE']['sanitizer_summary']((out/f'{mode}-{r}.stdout.log').read_text(),(out/f'{mode}-{r}.stderr.log').read_text())for r in (0,1)]
        if spec['kind']=='benchmark':
            for r in arm['measurements']:
                if r['rank']==0:groups.setdefault((r['dimension'],r['shape'],spec['backend']),[]).append(dict(seconds=r['seconds'],work_seconds=r['work_seconds'],capture_seconds=r['validation_capture_seconds']))
        checked[mode]=arm
    if raw['status']=='passed':assert raw['composed_vectors']==dict(status='passed',arms=checked)
    result.update(arms=checked,samples=[dict(values=1<<k[0],dimension=k[0],shape=k[1],backend=k[2],count=len(v),samples=v,median_seconds=statistics.median(r['seconds']for r in v),median_work_seconds=statistics.median(r['work_seconds']for r in v),minimum_work_seconds=min(r['work_seconds']for r in v),maximum_work_seconds=max(r['work_seconds']for r in v))for k,v in sorted(groups.items())],error=raw.get('error'),scope='Local Taylor32/64 small sanitizer plus synthetic N4096/16384 product comparison only; no fullsize/s18 result.')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
