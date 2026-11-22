#!/usr/bin/env python3
"""Complete bounded GPU discovery with tile and resident result ownership."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
SPECTRUM=runpy.run_path(str(HERE.parent/'distributed-spectrum-publication/run.py'))
SCALE=SPECTRUM['SCALE'];VECTORS=SCALE['VECTORS']


def discovery_claims(rows,result,mode,resume=False):
    assert all(r['discovery_result_mode']==mode for r in result['measurements'])
    summaries=[r for r in rows if r.get('kind')=='discovery-result']
    samples=[r for r in rows if r.get('kind')=='discovery-sample']
    if resume:
        assert not summaries and not samples
        return result
    assert len(summaries)==4 and len(samples)==sum(r['samples']for r in summaries)
    for r in summaries+samples:
        words=r['grid_words'];buffer=min(4096,words)if mode=='stream'else words
        assert r['result_mode']==mode and r['result_buffer_words']==buffer
        assert r['result_resident_host_words']==(0 if mode=='stream'else words)
        assert r['result_downloaded_words']==words*(r['samples']if r['kind']=='discovery-result'else 1)
    return result


def review(rows,directory,resume=False,source=None,cold=None):
    result=SPECTRUM['review'](rows,directory,resume,source,cold)
    return discovery_claims(rows,result,'stream',resume)


def resident_review(rows,directory):
    # Spectrum stays streamed in both arms;only discovery result ownership differs.
    result=SPECTRUM['review'](rows,directory)
    return discovery_claims(rows,result,'resident')


def suite(c):
    c['run']('discovery-result-review-tests',['python3','-B','experiments/distributed-discovery-results/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('union-cpu',[]),('union-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestDiscoveryUnion|^TestDiscoveryResultMode$|^TestDiscoverySampling','-count=1','-timeout=20s'],60)
    extra=[dict(name='resident-discovery',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_DISCOVERY_RESULT='resident',PIR_PIPELINE_SPECTRUM='stream'),review=resident_review)]
    result=SCALE['suite'](c,review_fn=review,extra_arms=extra)
    c['resources'].update(discovery_result='Explicit bounded tile consumption into existing union bitmap;matched resident discovery;spectrum streamed in both',scope='Genuine s8/curve5 tile discovery cold/fresh-prefix8 restart plus matched resident discovery cold;not fulls18')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min,0CUDApinned;explicit discovery-result tiles,s8 cold/restart+matched resident discovery;spectrum streamed in both;resident default retained;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
