#!/usr/bin/env python3
"""Reconcile complete streamed publication only after the completion ping."""
from pathlib import Path
import hashlib,json,runpy
HERE=Path(__file__).resolve().parent

def review(out,record,result):
    cases=json.loads((out/'export-cases.json').read_text());assert len(cases)==4
    assert {(r['layout'],r['gpus']) for r in cases}=={(l,n) for l in ('peer','local') for n in (1,4)}
    for case in cases:
        cache=Path(case['cache']);name='export-'+case['layout']+'-'+str(case['gpus'])
        for f,digest in case['sha256'].items():assert hashlib.sha256((cache/'code'/f).read_bytes()).hexdigest()==digest
        manifest=json.loads((cache/'code/manifest.json').read_text())
        assert manifest['producer']['revision']==record['revision']
        for suffix in ('fresh','cached','root-restart','cuda-free-readback'):
            rows=[json.loads(s) for s in (out/(name+'-'+suffix+'.log')).read_text().splitlines() if s.startswith('{')]
            assert any(r.get('phase')=='workflow' and r.get('event')=='completed' for r in rows)
            if suffix=='fresh':assert any(r.get('phase')=='spectrum-tile-bytes' and r.get('completed')==1<<20 for r in rows)
            if suffix=='root-restart':assert not any(r.get('phase')=='discovery' for r in rows)
            if suffix=='cached':assert not any(r.get('phase') in ('setup','discovery') for r in rows)
    for layout in ('peer','local'):
        for count in (1,2,4):assert '--- PASS: TestGPUPreprocessingExport ' in (out/f'export-api-{layout}-{count}.log').read_text()
    assert '--- PASS: TestPreprocessingPolynomialStage ' in (out/'polynomial-stage.log').read_text()
    for label in ('export-api-memcheck','export-cli-memcheck'):
        text=(out/(label+'.log')).read_text();assert 'ERROR SUMMARY: 0 errors' in text and 'LEAK SUMMARY: 0 bytes' in text
    for label,reason in [('export-corruption','generator.bin checksum mismatch'),
                         ('export-admission-failure','canonical payload exceeds --max-code-bytes')]:
        assert next(c for c in record['commands'] if c['label']==label)['exit_code']==1
        assert reason in (out/(label+'.log')).read_text()
    result['export_cases']=cases
    result['tests'].extend(['shared verified polynomial stage','complete canonical/G stream vs resident and independent Go','stream cache/root restarts and CUDA-free readback','writer/reuse/parameter errors; no partial stage publication; corruption rejection','four-GPU API/CLI CUDA memcheck'])
    result['scope']='bounded complete explicit stream publication; not fulls17 stream or fulls18 preprocessing, peak memory saving or large-size latency'

if __name__=='__main__':
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/analyze.py'))['main'](here=HERE,extra_review=review,expected_failures={"export-corruption":1,"export-admission-failure":1})
