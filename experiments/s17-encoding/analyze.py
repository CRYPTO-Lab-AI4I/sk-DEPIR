#!/usr/bin/env python3
"""Reconcile completed wide encoding evidence without scheduler/process calls."""
import argparse
import json
from pathlib import Path
import runpy

HERE = Path(__file__).resolve().parent
VERIFY = runpy.run_path(str(HERE.parent / 's17-arithmetic/analyze.py'))['reconcile']


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    args = parser.parse_args()
    out = HERE / 'runs' / args.name
    result = VERIFY(out)
    tests = ['TestScalingWideCPPEncoding', 'TestScalingWideCPPLifetime',
             'ExampleScalingCodeData_WithWideCPP']
    for label in ('encoding', 'asan'):
        log = (out / (label + '.log')).read_text()
        for test in tests:
            assert '--- PASS: ' + test + ' ' in log, (label, test)
    for label in ('go-regressions', 'cpp-regressions'):
        assert (out / (label + '.log')).read_text().strip() == 'PASS', label
    native = (out / 'native-api.log').read_text()
    for test in ('TestWideOnlineCounts', 'TestWideOnlineBorrowed'):
        assert '--- PASS: ' + test + ' ' in native, test
    assert (out / 'lifetime-race.log').read_text().startswith('ok'), 'lifetime race'
    for label in ('schedule', 'schedule-sanitizer'):
        check = json.loads((out / (label + '.log')).read_text())
        assert check == {'cases': 12, 'match': True}, label
    result.update(tests=tests + ['TestWideOnlineCounts', 'TestWideOnlineBorrowed'],
                  regressions=['shared Go-only', 'shared liftrs_cpp', 'native API'],
                  direct_cases={'optimized': 12, 'ASan/UBSan': 12},
                  scope='complete wide schedule on real s3/4/6/8 codes and synthetic small edges; metadata-only large counts; no full s17 preparation/encoding or latency estimate')
    (out / 'analysis.json').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps({k: result[k] for k in ('status', 'job_id', 'source_files_verified',
          'evidence_files_verified', 'cache_files_verified', 'direct_cases')}))


if __name__ == '__main__':
    main()
