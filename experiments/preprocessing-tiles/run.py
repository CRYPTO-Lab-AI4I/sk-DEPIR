#!/usr/bin/env python3
"""Bounded transform then split-product measurements, with shared job guards."""
from pathlib import Path
import hashlib
import json
import runpy
HERE=Path(__file__).resolve().parent


def verify_outputs(log, directory, manifest):
    """Full canonical hashes complement the tool's elementwise CPU comparisons."""
    expected={(x['kind'],x['index']):x for x in manifest}
    rows=[]
    completed=False
    for line in log.read_text().splitlines():
        if not line.startswith('{'):continue
        row=json.loads(line)
        if row.get('kind')=='completed':completed=True
        if row.get('kind') not in ('transform','product','origin'):continue
        assert row['matches'] is True
        reference=expected[(row['kind'],row['index'])]
        fields=[('output_file','output_sha256')] if row['kind']=='product' else [('forward_file','forward_sha256'),('inverse_file','inverse_sha256')]
        for name,digest in fields:
            if row[name]:
                path=directory/row[name]
                assert path.parent==directory
                hash=hashlib.sha256()
                with path.open('rb') as stream:
                    for block in iter(lambda:stream.read(1<<20),b''):hash.update(block)
                assert hash.hexdigest()==reference[digest],(row['kind'],row['index'],row[name])
        rows.append(row)
    assert completed,'component completion marker missing'
    return rows


def suite(context):
    run=context['run'];source=context['source'];cache=context['cache'];out=context['out'];env=context['env'];go=context['go']
    native=cache/'fixture.test';checker=cache/'component-check';fixtures=cache/'fixtures'
    run('fixture-build',[go,'test','-tags','liftrs_cpp','-c','-o',native,'./encoding/liftrs-shared'],source)
    env['PIR_PREPARATION_TILE_FIXTURES']=str(fixtures)
    run('cpu-fixtures',[native,'-test.run=^TestPreprocessingTileFixtures$','-test.v','-test.timeout=10m',
        '-test.cpuprofile='+str(out/'cpu-fixtures.pprof')],source)
    env.pop('PIR_PREPARATION_TILE_FIXTURES')
    run('component-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native',
        '-Xcompiler=-pthread','-pg','experiments/preprocessing-tiles/check.cu','-o',checker],source)
    manifest=json.loads((fixtures/'manifest.json').read_text())
    destinations=[cache/'outputs',cache/'memcheck-outputs']
    env['GMON_OUT_PREFIX']=str(cache/'cpu-profile-benchmark')
    run('components',[checker,fixtures/'transforms.bin',fixtures/'products.bin',destinations[0]],source)
    rows=verify_outputs(out/'components.log',destinations[0],manifest)
    assert any(r['kind']=='transform' and r['n']==1<<22 and r['backend']=='host-tiled' for r in rows)
    assert any(r['kind']=='product' and r['n']==1<<22 and r['backend']=='split' and r['subproducts']==14 for r in rows)
    profiles=sorted(cache.glob('cpu-profile-benchmark.*'));assert profiles,'native CPU profile not produced'
    for i,profile in enumerate(profiles):run('native-profile-'+str(i),['gprof','-b',checker,profile],source)
    env['GMON_OUT_PREFIX']=str(cache/'cpu-profile-memcheck')
    run('component-memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',
        checker,fixtures/'transforms.bin',fixtures/'products.bin',destinations[1],'--small'],source)
    sanitized=verify_outputs(out/'component-memcheck.log',destinations[1],manifest)
    text=(out/'component-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors' in text and 'LEAK SUMMARY: 0 bytes' in text
    (out/'components.json').write_text(json.dumps(dict(fixtures=manifest,measured=rows,sanitized=sanitized),indent=2)+'\n')
    env.pop('GMON_OUT_PREFIX')


if __name__=='__main__':
    workflow=runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))
    workflow['main'](here=HERE,extra_suite=suite)
