#!/usr/bin/env python3
"""Use the existing bounded GPU controller; add streamed-spectrum checks."""
from pathlib import Path
import runpy
HERE=Path(__file__).resolve().parent

def suite(context):
    run=context['run'];source=context['source'];cache=context['cache'];out=context['out']
    env=context['env'];go=context['go'];library=context['library']
    binary=cache/'spectrum.test'
    env.update(PIR_GPU_PREPARATION_LIBRARY=str(library),
               PIR_GPU_PREPARATION_DEVICES='0,1,2,3',
               PIR_GPU_PREPARATION_CAP=str(128<<20),
               PIR_GPU_SPECTRUM_MEASURE='1')
    run('spectrum-build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','-c','-o',binary,'./encoding/liftrs-shared'],source)
    run('spectrum-correctness',[binary,'-test.run=^TestGPUPreparationStreamedSpectrum$','-test.v',
                              '-test.cpuprofile='+str(out/'spectrum-correctness.pprof')],source)
    run('spectrum-transfer',[binary,'-test.run=^TestGPUPreparationSpectrumTransfer$','-test.v',
                          '-test.cpuprofile='+str(out/'spectrum-transfer.pprof')],source)
    env['PIR_GPU_PREPARATION_DEVICES']='0'
    run('spectrum-memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full',
        '--error-exitcode','99',binary,'-test.run=^TestGPUPreparationStreamedSpectrum$','-test.v'],source)
    log=(out/'spectrum-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors' in log and 'LEAK SUMMARY: 0 bytes' in log

if __name__=='__main__':
    workflow=runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))
    workflow['main'](here=HERE,extra_suite=suite)
