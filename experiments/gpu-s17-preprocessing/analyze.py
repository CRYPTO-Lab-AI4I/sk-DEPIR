#!/usr/bin/env python3
"""Reconcile completed GPU preparation evidence without checking live jobs."""
import argparse
import json
from pathlib import Path
import runpy
HERE=Path(__file__).resolve().parent

def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
    out=HERE/'runs'/args.name
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out)
    record=json.loads((out/'results.json').read_text())
    for label in ('gpu-preprocessing','memcheck'):
        log=(out/(label+'.log')).read_text()
        assert '--- PASS: TestGPUPreparationSmall ' in log and '--- PASS: TestGPUPreparationWide ' in log,label
        if record.get('peer_lifetime_check'):assert '--- PASS: TestGPUPreparationPeerLifetime ' in log,label
    assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
    window_rows=[json.loads(line) for line in (out/'wide-monomial-windows.log').read_text().splitlines() if line.startswith('{')]
    assert any(r.get('kind')=='address-map' and r['match'] for r in window_rows)
    windows=[r for r in window_rows if 'first_index' in r]
    assert len(windows)==3 and all(r['match'] and r['owned_device_bytes']<=min(256<<20,record['policy']['per_device_cap_bytes']) for r in windows)
    assert any(r['first_index']>=1<<32 for r in windows)
    tiles=json.loads((out/'tiles.json').read_text())
    expected={1<<power for power in range(record['policy']['gpus'].bit_length())}
    assert {r['gpus'] for r in tiles}==expected
    if record.get('vector_layout_comparison'):
        assert len(tiles)==2*len(expected)
        assert {(r['gpus'],r['vector_layout'])for r in tiles}=={(n,layout)for n in expected for layout in ('peer','local')}
        for n in expected:
            pair=[r for r in tiles if r['gpus']==n]
            assert len({r['owned_device_bytes']for r in pair})==1,'layout changes allocation payload'
            assert all(r['owned_device_bytes']<=n*record['policy']['per_device_cap_bytes']for r in pair)
    assert len({r['forward_sha256'] for r in tiles})==1
    for mode in ('optimized','sanitized'):
        rows=[json.loads(line) for line in (out/(mode+'.log')).read_text().splitlines() if line.startswith('{')]
        cases=[r for r in rows if 'match' in r]
        assert any(r.get('phase')=='logging-self-check' and r.get('unit')=='cases' for r in rows)
        assert any(r.get('phase')=='legacy-log-self-check' and 'unit' not in r for r in rows)
        assert len(cases)==(18 if record.get('vector_layout_comparison')else 9) and all(r['match'] for r in cases)
        if record.get('vector_layout_comparison'):assert {r['local_vectors']for r in cases}=={False,True}
    log=(out/'legacy-online.log').read_text();assert '--- PASS: TestCUDAOnlineEncoding ' in log
    progress=[json.loads(line) for line in (out/'gpu-preprocessing.log').read_text().splitlines() if line.startswith('{')]
    assert any(r.get('work')=='gpu-preparation' and r.get('event')=='completed' and r.get('unit') for r in progress)
    result.update(tests=['GPU exact small-code discovery/P/G/identity/spectrum/checkpoint bytes','s17 high-bit Horner/product/root kernels','complete resident tile hash/inverse and peer layout','legacy single-GPU online regression','compute-sanitizer memcheck/leaks; Go/native regressions; shared recurrence sanitizers'],
                  tiles=tiles,wide_monomial_windows=windows,progress_records=len(progress),scope='bounded resident GPU prototype; no full s17 preprocessing, full-size admission or runtime forecast')
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))

if __name__=='__main__':main()
