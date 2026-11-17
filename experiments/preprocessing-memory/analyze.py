#!/usr/bin/env python3
"""Review sealed bounded spectrum evidence only after the completion ping."""
import json
import runpy
from pathlib import Path
HERE=Path(__file__).resolve().parent

def review(out,record,result):
    assert '--- PASS: TestGPUPreparationStreamedSpectrum ' in (out/'spectrum-correctness.log').read_text()
    rows=[]
    for line in (out/'spectrum-transfer.log').read_text().splitlines():
        if line.startswith('SPECTRUM_TRANSFER '):rows=json.loads(line.split(' ',1)[1])
    assert len(rows)==4 and {r['streamed'] for r in rows}=={True,False}
    assert len({r['sha256'] for r in rows})==1
    assert all(r['seconds']>0 and r['n']==1<<22 for r in rows)
    log=(out/'spectrum-memcheck.log').read_text()
    assert '--- PASS: TestGPUPreparationStreamedSpectrum ' in log
    assert 'ERROR SUMMARY: 0 errors' in log and 'LEAK SUMMARY: 0 bytes' in log
    result['tests'].extend(['s8/s17 streamed slots vs independent scalar Horner; 1/2/4GPU peer/local',
                            'short/error writer, shape/tile/field, reuse after failure and closed state',
                            'balanced resident/streamed dense spectrum hashes; bounded CUDA memcheck'])
    result['spectrum_transfer']=rows
    result['scope']='bounded streamed-spectrum primitive and existing small complete workflows; not integrated streamed preprocessing or fulls18'

if __name__=='__main__':
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/analyze.py'))['main'](here=HERE,extra_review=review)
