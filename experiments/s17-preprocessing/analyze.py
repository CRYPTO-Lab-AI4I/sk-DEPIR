#!/usr/bin/env python3
"""Review a completed run and reconcile exact source/evidence/cache identities."""
import argparse
import json
from pathlib import Path
import runpy

HERE=Path(__file__).resolve().parent

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    args=parser.parse_args()
    out=HERE/'runs'/args.name
    result=runpy.run_path(str(HERE.parent/'s17-arithmetic/analyze.py'))['reconcile'](out)
    for name in ('preprocessing','asan'):
        log=(out/(name+'.log')).read_text()
        for test in ('TestPreprocessingNative','TestPreprocessingGoodMonomials','TestPreprocessingWideKernels','TestPreprocessingArtifacts','TestPreprocessingLimits'):
            assert '--- PASS: '+test+' ' in log,(name,test)
    for name in ('optimized','sanitized'):
        records=[json.loads(line) for line in (out/(name+'.log')).read_text().splitlines() if line.startswith('{')]
        assert len(records)==4 and {(r['word_bytes'],r['workers']) for r in records}=={(4,1),(4,4),(8,1),(8,4)}
        assert all(r['passed'] and r['complete_domain']==16 for r in records)
    progress=[json.loads(line) for line in (out/'preprocessing.log').read_text().splitlines() if line.startswith('{')]
    assert any(p.get('phase')=='generator-spectrum' and p.get('event')=='completed' for p in progress)
    assert any(p.get('phase')=='exact-good-monomials' and p.get('event')=='failed' for p in progress)
    result.update(tests=['exact small-code references and artifact/restart identities','high-bit s17 Horner/products/roots/monomials','forced32/64 small tensor/product schedules','Go/native regression suites; CUDA compile boundary','vet/race/ASan/UBSan'],scope='bounded validation only; no full s17 discovery/preparation, CUDA17 or admission',progress_records=len(progress))
    (out/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))

if __name__=='__main__':main()
