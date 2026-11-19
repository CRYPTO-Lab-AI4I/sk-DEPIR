#!/usr/bin/env python3
"""Verify completed rank ownership/tree evidence without running MPI/GPU work."""
import argparse
import json
from pathlib import Path
import runpy
import statistics
HERE=Path(__file__).resolve().parent


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('name');a=p.parse_args()
    out=HERE/'runs'/a.name;raw=json.loads((out/'results.json').read_text());helper=runpy.run_path(str(HERE/'run.py'))
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out,command_fields=('label','seconds','exit_code'))
    normal=helper['BASE']['parse'](out/'normal.log');assert normal==raw['measurements']
    measured=helper['review'](normal,Path(raw['cache'])/'normal')
    small=helper['BASE']['parse'](out/'memcheck.log');assert small==raw['sanitized']
    helper['review'](small,Path(raw['cache'])/'sanitized',small=True)
    text=(out/'memcheck.log').read_text();assert text.count('ERROR SUMMARY: 0 errors')==4 and text.count('LEAK SUMMARY: 0 bytes')==4
    summary=[]
    for case in range(3):
        rows=[r for r in measured if r['case']==case and r['rank']==0 and r['sample']in(0,1)]
        summary.append(dict(case=case,s=rows[0]['s'],support_count=rows[0]['support_count'],
            median_seconds={f:statistics.median(r[f]for r in rows)for f in helper['FIELDS']}))
    result.update(summary=summary,tests=['shared routing/tree ownership and race','historical root/reference hashes',
        'coordinator-only full support/CPU oracle','disjoint complete Frobenius shards','all send/receive bytes and tree ownership chain',
        'unique final owner equals CPU root','four CUDA memchecks','source/evidence/cache seals'],
        scope='four logical ranks sharing two physical GPUs; no four-GPU speedup, distributed discovery/FFT, fulls18 support or default change',
        command_RSS='not sampled; per-rank high-water observations in measurements, not a total node/job peak')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))


if __name__=='__main__':main()
