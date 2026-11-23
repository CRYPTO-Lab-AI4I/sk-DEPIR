#!/usr/bin/env python3
"""One support hash plus full metadata agreement versus existing local recovery."""
from pathlib import Path
import hashlib
import json
import os
import runpy
import sys
HERE=Path(__file__).resolve().parent
DIAG=runpy.run_path(str(HERE.parent/'distributed-diagnostic-writes/run.py'))
SCALE=DIAG['SCALE'];VECTORS=DIAG['VECTORS'];PROOF=DIAG['PROOF']


def binding_frame(record,directory,label):
    # Commit JSON is a sorted Go map; frame JSON is a typed Go struct. Rebuild
    # its field order explicitly, retaining exact integers and Go HTML escaping.
    fields=('layout','s','k','n','prefix','ranks','support_sha256','stage','sha256')
    ordered={k:record[k]for k in fields if k!='layout'or record.get(k)}
    body=dict(Context='sk-DEPIR-rank-support-v1',Directory=os.path.abspath(directory),Label=label,Record=ordered)
    encoded=json.dumps(body,separators=(',',':'),ensure_ascii=False)
    for char,escape in (('<','\\u003c'),('>','\\u003e'),('&','\\u0026'),('\u2028','\\u2028'),('\u2029','\\u2029')):encoded=encoded.replace(char,escape)
    return hashlib.sha256(encoded.encode()).hexdigest()


def support_claims(rows,directory,mode):
    evidence=[r for r in rows if r.get('kind')=='restart-support']
    if mode is None:
        assert not evidence;return dict(mode='cold',hash_passes=0,hashed_bytes=0)
    assert len(evidence)==4 and {r['rank']for r in evidence}==set(range(4))
    size=(directory/'s8-support.bin').stat().st_size
    agreements=[r for r in rows if r.get('kind')=='restart-support-agreement']
    if mode=='local':assert not agreements
    else:
        assert len(agreements)==8 and {(r['rank'],r['checkpoint'])for r in agreements}=={(rank,stage)for rank in range(4)for stage in ('root','newton-8')}
        for r in agreements:
            assert Path(r['source']).resolve()==directory.resolve()and r['case']=='s8'
            record=json.loads((directory/('s8-'+r['checkpoint']+'-commit.json')).read_text())
            digest=binding_frame(record,directory,'s8')
            assert r['frames']==[digest+(record['support_sha256']if rank==0 else'00'*32)for rank in range(4)]
    for r in evidence:
        wanted=(1 if r['rank']==0 else 0)if mode=='coordinator'else 2
        assert r['mode']==mode and r['verified']is True and r['hash_passes']==wanted and r['hashed_bytes']==wanted*size
        assert r['agreements']==(2 if mode=='coordinator'else 0)
    return dict(mode=mode,hash_passes=sum(r['hash_passes']for r in evidence),hashed_bytes=sum(r['hashed_bytes']for r in evidence),support_bytes=size)


def review(rows,directory,resume=False,source=None,cold=None,mode='coordinator'):
    result=DIAG['review'](rows,directory,resume,source,cold)
    result['support_verification']=support_claims(rows,source if resume else directory,mode if resume else None)
    return result


def suite(c):
    c['run']('binding-review-tests',['python3','-B','experiments/distributed-restart-binding/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_RESTART_SUPPORT='coordinator',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('support-binding-cpu',[]),('support-binding-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestRankSupport|^TestRankPrefix','-count=1','-timeout=20s'],60)
    def local(rows,directory):
        cold=json.loads((c['out']/'composition.json').read_text())['arms']['cold']
        return review(rows,directory,True,c['cache']/'cold',cold,'local')
    extra=[dict(name='local-resume',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM=str(c['cache']/'cold'),PIR_PIPELINE_RESTART_SUPPORT='local',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded'),review=local)]
    result=SCALE['suite'](c,review_fn=review,extra_arms=extra)
    arms=result['composed_vectors']['arms'];assert arms['resume']['artifacts']==arms['local-resume']['artifacts']
    result['diagnostic_writes']={mode:DIAG['diagnostic_review'](c['out'],mode,c['cache']/mode)for mode in ('cold','resume','local-resume')}
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4))for mode in ('cold','resume','local-resume')}
    c['resources'].update(restart_support='Explicit coordinator full-file hash and full checkpoint-metadata agreement;matched default local hashes',scope='Genuine s8 cold,coordinated prefix8 restart and matched local restart;not fulls18')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;3cold/coordinated/local arms;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
