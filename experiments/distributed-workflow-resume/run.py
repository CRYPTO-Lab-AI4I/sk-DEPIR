#!/usr/bin/env python3
"""Fresh-process assembled arithmetic restart from durable source-bound checkpoints."""
from pathlib import Path
import os
import runpy
import sys
HERE=Path(__file__).resolve().parent
CHECKPOINT=runpy.run_path(str(HERE.parent/'distributed-workflow-checkpoints/run.py'))
WORKFLOW=CHECKPOINT['WORKFLOW'];BIND=CHECKPOINT['BIND'];SCALE=CHECKPOINT['SCALE'];VECTORS=CHECKPOINT['VECTORS'];PROOF=CHECKPOINT['PROOF'];DIAG=CHECKPOINT['DIAG']
assembled_review=CHECKPOINT['assembled_review'];checkpoint_review=CHECKPOINT['checkpoint_review']


def resume_review(rows,directory,baseline_rows,source):
    values=WORKFLOW['workflow_claims'](rows,known=8)
    proof=PROOF['check_summaries'](rows,baseline_rows,'trace')
    support=BIND['support_claims'](rows,source,'coordinator')
    checkpoints=checkpoint_review(rows,directory,known=8)
    publication=WORKFLOW['publication_review'](rows,directory)
    assert os.path.samefile(source/'s8-support.bin',directory/'s8-archive.raw')
    events=[r for r in rows if r.get('kind')=='workflow-resume']
    assert len(events)==4 and {r['rank']for r in events}==set(range(4))
    for r in events:
        assert r['case']=='s8'and r['prefix']==8 and r['discovery_skipped']is True and r['roots_skipped']is True
        assert Path(r['source']).resolve()==source.resolve()
        assert r['source_sha256']==checkpoints['source_sha256']
    assert not any(r.get('discovery')is True or r.get('kind')=='root-stage'for r in rows)
    files=[]
    prefixes=[v for v in WORKFLOW['OWN']['PILOT']['prefixes'](58155)if v>=8]
    for rank in range(4):
        names=[f's8-rank-{rank}-{suffix}.bin'for suffix in ('root','identity')]
        names.extend(f's8-rank-{rank}-newton-{v}.bin'for v in prefixes)
        for name in names:
            assert (directory/name).read_bytes()==(source/name).read_bytes(),name
            files.append(name)
    for name in ('s8-code.bin','s8-generator.bin','s8-archive.raw'):
        assert (directory/name).read_bytes()==(source/name).read_bytes(),name
    return dict(measurements=values,proof_recording=proof,support_verification=support,checkpoints=checkpoints,publication=publication,matched_checkpoint_files=files,scope='Fresh-process assembled prefix8 arithmetic restart;not fullsize/node-loss recovery')


def suite(c):
    c['run']('workflow-resume-review-tests',['python3','-B','experiments/distributed-workflow-resume/test_review.py'],30)
    c['env']['PIR_WORKFLOW_RESUME_FROM']=''
    def resumed(rows,directory):
        baseline=VECTORS['BASE']['parse_rank_logs'](c['out']/'resume.log',ranks=4)
        return resume_review(rows,directory,baseline,c['cache']/'assembled')
    extra=[dict(name='assembled-resume',test='TestAssembledRankWorkflow',memcheck=False,profile=True,env=dict(PIR_WORKFLOW_RESUME_FROM=str(c['cache']/'assembled'),PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded'),review=resumed)]
    result=CHECKPOINT['suite'](c,extra_arms=extra)
    c['resources']['workflow_resume']='Fresh-process assembled prefix8 restart from source-bound compact commits;one coordinator support hash;shared cold/resume numerical tail;hard-linked archive reuse'
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2nodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;pilot cold/restart+assembled cold/restart;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
