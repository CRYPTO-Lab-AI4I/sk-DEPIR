#!/usr/bin/env python3
"""Bounded GPU spectrum publication with complete cold/restart and wide source."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
SCALE=runpy.run_path(str(HERE.parent/'distributed-pipeline-scaling/run.py'))
VECTORS=SCALE['VECTORS'];WIDE=SCALE['OWN']['WIDE']


def source_claims(result,narrow=False):
    for r in result['measurements']:
        assert r['spectrum_mode']=='stream'and r['spectrum_resident_host_words']==0
        assert r['spectrum_downloaded_words']==r['words']
        if narrow:assert r['spectrum_host_words']==min(4096,r['words'])
    return result


def review(rows,directory,resume=False,source=None,cold=None):
    return source_claims(SCALE['review'](rows,directory,resume,source,cold),narrow=True)


def wide_review(rows,directory):
    result=source_claims(WIDE['wide_review'](rows,directory))
    # Low-dimensional field18 spectra can remain in Fq. The separate high-bit
    # fixture must actually pass through the GPU-backed source after the spectrum.
    for r in result['measurements']:
        assert r['high_source_downloaded_words']==r['words']
        label,rank=r['case'],r['rank']
        assert(directory/f'{label}-rank-{rank}-high-source.bin').read_bytes()==(directory/f'{label}-rank-{rank}-high.bin').read_bytes()
    data=(directory/'s18-d12-stream-high.bin').read_bytes()
    assert any(int.from_bytes(data[i:i+8],'little')>=1<<32 for i in range(0,len(data),8))
    return result


def suite(c):
    c['run']('spectrum-source-review-tests',['python3','-B','experiments/distributed-spectrum-publication/test_review.py'],30)
    c['env']['PIR_PIPELINE_SPECTRUM']='stream'
    extra=[dict(name='wide',test='TestWidePipelineTail',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_SPECTRUM='stream'),review=wide_review)]
    result=SCALE['suite'](c,review_fn=review,extra_arms=extra)
    c['resources'].update(spectrum_output='explicit GPU tile source,one download/word,no actual resident host spectrum',scope='Genuine s8/curve5 source-backed cold/fresh-prefix8 restart and bounded canonical field18 source/native/high-bit regression;not fulls18')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min,0CUDApinned;explicit streamed spectrum,genuine s8 cold/restart+bounded wide;resident default retained;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
