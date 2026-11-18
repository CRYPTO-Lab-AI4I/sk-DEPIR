#!/usr/bin/env python3
"""Reconcile worker setup/partition/roots/merge timings after completion ping."""
from pathlib import Path
import json
import runpy
HERE=Path(__file__).resolve().parent


def review(out,record,result):
    verify=runpy.run_path(str(HERE/'run.py'))['review_output']
    cache=Path(record['cache'])
    rows=verify(out/'workers.log',cache/'normal',4)
    sanitized=verify(out/'workers-memcheck.log',cache/'sanitized',1)
    package=json.loads((out/'workers.json').read_text())
    assert package==dict(measured=rows,sanitized=sanitized)
    assert {r['s'] for r in rows}=={8,12,17}
    assert all(r['genuine_support']==(r['case']==0) for r in rows)
    text=(out/'workers-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors'in text and 'LEAK SUMMARY: 0 bytes'in text
    assert (out/'workers.pprof').stat().st_size>0
    result.update(workers=package,scope='same-node independent root-worker scheduling component; not cross-node execution, fulls18 runtime or a default change')


if __name__=='__main__':
    overrides=runpy.run_path(str(HERE.parent/'preprocessing-root-boundary/analyze.py'))['SOURCE_OVERRIDES']
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/analyze.py'))['main'](here=HERE,extra_review=review,source_overrides=overrides)
