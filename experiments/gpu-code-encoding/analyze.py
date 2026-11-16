#!/usr/bin/env python3
"""Shared evidence reconciliation, followed by client-specific requirements."""
from pathlib import Path
import json
import runpy

HERE = Path(__file__).resolve().parent


def review(out, record, result):
    log = (out/'code-small.log').read_text()
    assert '--- PASS: TestCUDACodeSmall ' in log
    assert '--- PASS: TestCUDACodeResources ' in log
    assert '--- PASS: TestCUDACodeNativeFixtures ' in (out/'code-fixtures.log').read_text()
    wide = (out/'code-wide-kernels.log').read_text()
    for test in ('TestGPUPreparationWide', 'TestGPUPreparationPeerLifetime'):
        assert '--- PASS: '+test+' ' in wide
    for label, count in [('code-check', 60), ('code-memcheck', 12)]:
        events = [json.loads(line) for line in (out/(label+'.log')).read_text().splitlines()
                  if line.startswith('{')]
        assert any(row.get('resident_code_cases') == count and row.get('wide_positions')
                   is True for row in events), label
        assert any(row.get('work') == 'encoding-operations' and
                   row.get('completed') == row.get('total') == 10 and
                   row.get('event') == 'completed' for row in events), label
    memcheck = (out/'code-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors' in memcheck and 'LEAK SUMMARY: 0 bytes' in memcheck
    result['tests'].extend([
        'resident GPU client vs independent Go and wide CPU on genuine smaller codes',
        '1/2/4GPU peer/local, ownership, zero/invalid/retry/cap/state/coexisting clients',
        'forced32/64-bit client schedules, true34-bit positions and s17 kernel/lease regressions',
        'client CUDA memcheck'])
    result['scope'] = ('bounded server regressions and complete small resident GPU client; '
                       'no full s17 encoding/preprocessing performance or admission')
    result['client_native_cases'] = 60
    result['client_sanitizer_cases'] = 12


if __name__ == '__main__':
    workflow = runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/analyze.py'))
    workflow['main'](here=HERE, extra_review=review)
