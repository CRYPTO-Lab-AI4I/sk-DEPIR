#!/usr/bin/env python3
"""Review bounded resident storage outputs, all transpose routes and sanitizer evidence."""
from pathlib import Path
import argparse
import json
import runpy
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');args=p.parse_args()
    out=HERE/'runs'/args.name;raw=json.loads((out/'results.json').read_text());cache=Path(raw['cache'])
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,require_success=raw['status']=='passed',command_fields=('label','seconds','exit_code'))
    checked={}
    for mode in ('separable','normal','memcheck'):
        if any(c['label']==mode and c['exit_code']==0 for c in raw['commands']):
            review=H['VECTORS']['review']if mode=='separable'else H['review']
            arm=review(H['VECTORS']['BASE']['parse_rank_logs'](out/(mode+'.log'),ranks=2),cache/mode)
            if mode=='memcheck':arm['sanitizer']=[H['VECTORS']['STAGE']['sanitizer_summary']((out/f'{mode}-{r}.stdout.log').read_text(),(out/f'{mode}-{r}.stderr.log').read_text())for r in (0,1)]
            checked[mode]=arm
    if raw['status']=='passed':assert raw['composed_vectors']==dict(status='passed',arms=checked)
    result.update(tests=['independent native/scalar separable oracle','full forward/inverse equality','third-slot preservation and stable retained allocation','out-of-field rejection without mutation','exact paired transpose proofs','source/evidence/cache seals','dual-stream CUDA summaries'],arms=checked,error=raw.get('error'),scope='Resident separable vectors only; Taylor/full resident FFT/product integration and fulls18 preprocessing remain unvalidated.')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
