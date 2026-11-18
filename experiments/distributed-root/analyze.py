#!/usr/bin/env python3
"""Verify completed distributed-root evidence without running MPI/GPU work."""
from pathlib import Path
import argparse
import json
import runpy
import statistics
HERE=Path(__file__).resolve().parent


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');args=p.parse_args()
    out=HERE/'runs'/args.name
    shared=runpy.run_path(str(HERE/'run.py'))
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,command_fields=('label','seconds','exit_code'))
    raw=json.loads((out/'results.json').read_text())
    rows=shared['TRANSPORT']['parse'](out/'normal.log');assert rows==raw['measurements']
    sanitized=shared['TRANSPORT']['parse'](out/'memcheck.log');assert sanitized==raw['sanitized']
    measured=shared['review'](rows,Path(raw['cache'])/'normal')
    shared['review'](sanitized,Path(raw['cache'])/'sanitized',small=True)
    text=(out/'memcheck.log').read_text()
    assert text.count('ERROR SUMMARY: 0 errors')==2 and text.count('LEAK SUMMARY: 0 bytes')==2
    summary=[]
    for case in range(3):
        selected=[r for r in measured if r['case']==case and r['rank']==0 and r['sample']>=0]
        summary.append(dict(case=case,s=selected[0]['s'],support_count=selected[0]['support_count'],
            median_seconds={f:statistics.median(r[f]for r in selected)for f in shared['FIELDS']},
            remote_bytes_rank0=selected[0]['remote_bytes']))
    result.update(summary=summary,tests=['shared-fixture historical digests','complete normalized CPU root coefficients',
        'independent scalar root witnesses','cross-rank saved polynomial equality','two distinct ranks/nodes',
        'full receiving-GPU merge equality','both CUDA memchecks','source/evidence/cache seals'],
        scope='bounded distributed root products/host-polynomial exchange/replicated GPU merge; no generator, spectrum, full s18, scalable support discovery or multi-level merge tree',
        command_RSS='not sampled; rank high-water RSS in measurements is not aggregate job memory')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))


if __name__=='__main__':main()
