#!/usr/bin/env python3
"""Bounded complete streamed publication checks using the shared GPU controller."""
from pathlib import Path
import hashlib
import json
import runpy
import shutil
HERE=Path(__file__).resolve().parent

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def suite(context):
    run=context['run'];source=context['source'];cache=context['cache'];out=context['out']
    env=context['env'];go=context['go'];library=context['library']
    binary=cache/'export.test';cli=cache/'preprocess-cuda';plain=cache/'preprocess-go'
    run('export-build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','-c','-o',binary,'./encoding/liftrs-shared'],source)
    run('polynomial-stage',[binary,'-test.run=^TestPreprocessing(PolynomialStage|Native|RecoveryBoundaries)$','-test.v',
                           '-test.cpuprofile='+str(out/'polynomial-stage.pprof')],source)
    env.update(PIR_GPU_PREPARATION_LIBRARY=str(library),PIR_GPU_PREPARATION_CAP=str(128<<20))
    for layout in ('peer','local'):
        env['LIFTRS_GPU_PREPARATION_VECTOR_LAYOUT']=layout
        for count in (1,2,4):
            env['PIR_GPU_PREPARATION_DEVICES']=','.join(map(str,range(count)))
            run('export-api-'+layout+'-'+str(count),[binary,'-test.run=^TestGPUPreprocessingExport$','-test.v',
                '-test.cpuprofile='+str(out/('export-api-'+layout+'-'+str(count)+'.pprof'))],source)
    reference=cache/'peer-1'/'code'
    expected={name:sha(reference/name) for name in ('online.bin','generator.bin')}
    cases=[]
    for layout in ('peer','local'):
        for count in (1,4):
            name='export-'+layout+'-'+str(count);dest=cache/name
            command=[str(cli),'--s','8','--t','5','--cache',str(dest),'--backend','cuda',
                     '--gpu-library',str(library),'--devices',','.join(map(str,range(count))),
                     '--gpu-bytes',str(1<<30),'--max-array-bytes',str(8<<20),
                     '--native-bytes',str(64<<20),'--max-code-bytes',str(8<<20),
                     '--spectrum-tile-bytes',str(1<<20),'--vector-layout',layout]
            run(name+'-fresh',command,source)
            for f,digest in expected.items():assert sha(dest/'code'/f)==digest
            manifests={stage:sha(dest/stage/'manifest.json') for stage in ('support','root','code')}
            events=[json.loads(s) for s in (out/(name+'-fresh.log')).read_text().splitlines() if s.startswith('{')]
            assert any(r.get('phase')=='spectrum-tile-bytes' and r.get('completed')==1<<20 for r in events)
            run(name+'-cached',command+['--resume'],source)
            assert all(sha(dest/stage/'manifest.json')==digest for stage,digest in manifests.items())
            (dest/'code').rename(dest/'code-first')
            run(name+'-root-restart',command+['--resume'],source)
            assert all(sha(dest/'code'/f)==digest for f,digest in expected.items())
            assert all(sha(dest/stage/'manifest.json')==manifests[stage] for stage in ('support','root'))
            run(name+'-cuda-free-readback',[str(plain),'--s','8','--t','5','--backend','cpp',
                '--cache',str(dest),'--max-array-bytes',str(8<<20),'--native-bytes',str(64<<20),
                '--max-code-bytes',str(8<<20),'--resume'],source)
            cases.append(dict(layout=layout,gpus=count,cache=str(dest),sha256=expected,
                              tile_bytes=1<<20,first_manifests=manifests))
    # Keep the damaged copy for reproducibility; no useful valid cache is altered.
    corrupt=cache/'export-corrupt';shutil.copytree(cache/'export-peer-1',corrupt)
    payload=corrupt/'code/generator.bin'
    with payload.open('r+b') as stream:
        first=stream.read(1);stream.seek(0);stream.write(bytes([first[0]^1]))
    run('export-corruption',[str(plain),'--s','8','--t','5','--backend','cpp','--cache',str(corrupt),
        '--max-array-bytes',str(8<<20),'--native-bytes',str(64<<20),
        '--max-code-bytes',str(8<<20),'--resume'],source,expect=1)
    # A rejected output cap must leave reusable support/root but no published code.
    failed=cache/'export-admission';cmd=[str(cli),'--s','4','--t','3','--cache',str(failed),
        '--backend','cuda','--gpu-library',str(library),'--devices','0','--gpu-bytes',str(1<<30),
        '--max-array-bytes',str(1<<20),'--native-bytes',str(8<<20),'--max-code-bytes','1',
        '--spectrum-tile-bytes','24']
    run('export-admission-failure',cmd,source,expect=1)
    assert not (failed/'code').exists()
    assert (failed/'support/manifest.json').exists() and (failed/'root/manifest.json').exists()
    run('export-admission-restart',cmd+['--max-code-bytes',str(1<<20),'--resume'],source)
    env['PIR_GPU_PREPARATION_DEVICES']='0,1,2,3'
    run('export-api-memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full',
        '--error-exitcode','99',binary,'-test.run=^TestGPUPreprocessingExport$','-test.v'],source)
    cli_memcheck=cache/'export-memcheck'
    run('export-cli-memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full',
        '--error-exitcode','99',str(cli),'--s','4','--t','3','--cache',str(cli_memcheck),
        '--backend','cuda','--gpu-library',str(library),'--devices','0,1,2,3',
        '--gpu-bytes',str(1<<30),'--max-array-bytes',str(1<<20),
        '--native-bytes',str(8<<20),'--max-code-bytes',str(1<<20),'--spectrum-tile-bytes','24'],source)
    for label in ('export-api-memcheck','export-cli-memcheck'):
        text=(out/(label+'.log')).read_text()
        assert 'ERROR SUMMARY: 0 errors' in text and 'LEAK SUMMARY: 0 bytes' in text
    (out/'export-cases.json').write_text(json.dumps(cases,indent=2)+'\n')

if __name__=='__main__':
    workflow=runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))
    workflow['main'](here=HERE,extra_suite=suite)
