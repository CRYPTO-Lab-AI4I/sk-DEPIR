#!/usr/bin/env python3
"""Explicit Newton prefix inputs using the same product,owned pipeline and oracles."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
DISCOVERY=runpy.run_path(str(HERE.parent/'distributed-discovery-results/run.py'))
SPECTRUM=DISCOVERY['SPECTRUM'];SCALE=DISCOVERY['SCALE'];VECTORS=SCALE['VECTORS']


def upload_words(first,words,count):return max(0,min(first+words,count)-first)


def input_claims(result,mode,wide=False,resume=False):
    for r in result['measurements']:
        n=r['n'];first,words=r['first'],r['words'];target=r['target']if wide else n-r['k'];root=r['root_count']if wide else r['k']+1
        known=8 if resume else 1;expected=steps=0
        for size in SCALE['OWN']['PILOT']['prefixes'](target):
            if size<=known:continue
            ac,bc=min(root,size),min(2*known-1,size)
            expected+=(upload_words(first,words,ac)+upload_words(first,words,bc))if mode=='prefix'else 2*words
            known=size;steps+=1
        assert r['newton_input_mode']==mode and r['newton_steps']==steps
        assert r['newton_root_clone_words_per_step']==(0 if mode=='prefix'else words)
        assert r['newton_input_upload_words']==expected
    return result


def review(rows,directory,resume=False,source=None,cold=None):
    return input_claims(DISCOVERY['review'](rows,directory,resume,source,cold),'prefix',resume=resume)


def dense_review(rows,directory):
    return input_claims(DISCOVERY['review'](rows,directory),'dense')


def wide_review(rows,directory):
    return input_claims(SPECTRUM['wide_review'](rows,directory),'prefix',wide=True)


def suite(c):
    c['run']('prefix-input-review-tests',['python3','-B','experiments/distributed-prefix-inputs/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('prefix-view-cpu',[]),('prefix-view-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestRankPrefixInput|^TestDiscoveryUnion|^TestRankStream','-count=1','-timeout=20s'],60)
    extra=[dict(name='dense',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_NEWTON_INPUT='dense'),review=dense_review),dict(name='wide',test='TestWidePipelineTail',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_SPECTRUM='stream'),review=wide_review)]
    result=SCALE['suite'](c,review_fn=review,extra_arms=extra)
    c['resources'].update(newton_input='Explicit immutable prefix views;zero-reset destinations,active-only uploads;matched dense input control',scope='Genuine s8/curve5 prefix-input cold/fresh-prefix8 restart,matched dense input,and bounded field18 prefix-input/source/native checks;not fulls18')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min,0CUDApinned;s8 prefix-input cold/restart+matched dense+bounded field18 prefix;dense default retained;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
