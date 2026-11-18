#!/usr/bin/env python3
"""Bounded shared root-boundary correctness and backend/recovery regressions."""
from pathlib import Path
import runpy
HERE=Path(__file__).resolve().parent


def suite(context):
    run=context['run'];source=context['source'];out=context['out'];cache=context['cache'];go=context['go'];env=context['env']
    pattern='TestPreparationRootPartitions|TestPreprocessingNative|TestPreprocessingPolynomialStage|TestPreprocessingRecoveryBoundaries'
    native=cache/'root-native.test'
    run('root-native-build',[go,'test','-c','-tags','liftrs_cpp','-o',native,'./encoding/liftrs-shared'],source)
    run('root-native',[native,'-test.run='+pattern,'-test.v','-test.count=1','-test.timeout=10m','-test.cpuprofile='+str(out/'root-native.pprof')],source)
    run('root-race',[go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared','-run',pattern,'-count=1','-timeout=10m'],source)
    binary=cache/'root-gpu.test'
    run('root-gpu-build',[go,'test','-c','-tags','liftrs_cpp,liftrs_cuda','-o',binary,'./encoding/liftrs-shared'],source)
    env.update(PIR_GPU_PREPARATION_LIBRARY=str(context['library']),PIR_GPU_PREPARATION_DEVICES='0,1,2,3',PIR_GPU_PREPARATION_CAP=str(128<<20))
    run('root-gpu',[binary,'-test.run=TestGPUPreparationSmall|TestGPUPreprocessingExport','-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'root-gpu.pprof')],source)
    env['PIR_GPU_PREPARATION_DEVICES']='0'
    run('root-memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',binary,'-test.run=TestGPUPreprocessingExport','-test.v','-test.timeout=10m'],source)
    text=(out/'root-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors'in text and 'LEAK SUMMARY: 0 bytes'in text


if __name__=='__main__':
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))['main'](here=HERE,extra_suite=suite)
