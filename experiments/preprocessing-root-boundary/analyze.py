#!/usr/bin/env python3
"""Review sealed root-boundary/backend evidence after a completion ping."""
from pathlib import Path
import runpy
HERE=Path(__file__).resolve().parent


def review(out,record,result):
    native=(out/'root-native.log').read_text()
    for name in ('TestPreparationRootPartitions','TestPreprocessingNative','TestPreprocessingPolynomialStage','TestPreprocessingRecoveryBoundaries'):
        assert '--- PASS: '+name in native,name
    gpu=(out/'root-gpu.log').read_text()
    for name in ('TestGPUPreparationSmall','TestGPUPreprocessingExport'):
        assert '--- PASS: '+name in gpu,name
    assert '--- SKIP:' not in gpu
    memory=(out/'root-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors'in memory and 'LEAK SUMMARY: 0 bytes'in memory
    cache=Path(record['cache'])
    for name in ('root-native.test','root-gpu.test'):assert (cache/name).is_file()
    for name in ('root-native.pprof','root-gpu.pprof'):assert (out/name).stat().st_size>0
    result['tests'].extend(['normalized closed-orbit subset merge and caller immutability',
        'root publication failure prevents generator work','native recovery/export and race',
        'GPU root/recovery/canonical export plus CUDA memcheck'])
    result['scope']='shared root/generator boundary correctness; no parallel-root speed, measured memory saving or distributed preprocessing'


if __name__=='__main__':
    runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/analyze.py'))['main'](here=HERE,extra_review=review)
