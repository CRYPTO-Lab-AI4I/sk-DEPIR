#!/usr/bin/env python3
"""Reconcile matched validator measurements after completion ping."""
import json
from pathlib import Path
import runpy
HERE=Path(__file__).resolve().parent


def review(out,record,result):
    package=json.loads((out/'comparison.json').read_text())
    local=runpy.run_path(str(HERE/'run.py'));local['validate_pair'](package['records'])
    verify=runpy.run_path(str(HERE.parent/'parallel-root-workers/run.py'))['review_output']
    cache=Path(record['cache'])
    for entry in package['records']:
        label=entry['mode']+'-'+str(entry['repeat'])
        assert verify(out/(label+'.log'),cache/label,4)==entry['rows']
        assert (out/(label+'.pprof')).stat().st_size>0
    assert {r['mode'] for r in package['sanitized']}=={'bitmap','sorted'}
    for entry in package['sanitized']:
        label=entry['mode']+'-memcheck'
        assert verify(out/(label+'.log'),cache/label,1)==entry['rows']
        text=(out/(label+'.log')).read_text()
        assert 'ERROR SUMMARY: 0 errors'in text and 'LEAK SUMMARY: 0 bytes'in text
    tests=(out/'membership-tests.log').read_text()
    for name in ('TestSupportMembershipEquivalence','TestSortedSupportWideMemoryAndProgress','TestPreparationRootPartitions','TestRootPartitionAdmission'):
        assert '--- PASS: '+name in tests
    result.update(comparison=package,scope='explicit subset-validator/worker comparison; no fullsupport default change, fulls18 speed or distributed execution')


if __name__=='__main__':
    overrides=runpy.run_path(str(HERE.parent/'preprocessing-root-boundary/analyze.py'))['SOURCE_OVERRIDES']
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/analyze.py'))['main'](here=HERE,extra_review=review,source_overrides=overrides)
