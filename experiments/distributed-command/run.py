#!/usr/bin/env python3
"""Parameterized MPI server command, compared with reviewed pilot cold/restart."""
from pathlib import Path
import json
import os
import runpy
import shutil
import sys
HERE=Path(__file__).resolve().parent
RESUME=runpy.run_path(str(HERE.parent/'distributed-workflow-resume/run.py'))
WORKFLOW=RESUME['WORKFLOW'];CHECKPOINT=RESUME['CHECKPOINT'];BIND=RESUME['BIND'];SCALE=RESUME['SCALE'];VECTORS=RESUME['VECTORS'];PROOF=RESUME['PROOF'];DIAG=RESUME['DIAG']


def command_review(rows,directory,baseline_rows,baseline_directory,known=1,source=None):
    values=WORKFLOW['workflow_claims'](rows,known=known)
    assert all(r['entry']=='command'and r['checkpoint_mode']=='commit'for r in values)
    assert len([r for r in rows if r.get('kind')=='job-complete'])==4
    proof=PROOF['check_summaries'](rows,baseline_rows,'trace')
    checkpoints=CHECKPOINT['checkpoint_review'](rows,directory,known=known)
    publication=WORKFLOW['publication_review'](rows,directory)
    support=BIND['support_claims'](rows,source,'coordinator')if source else None
    if source:assert os.path.samefile(source/'s8-support.bin',directory/'s8-archive.raw')
    files=[]
    stages=['root']+[f'newton-{v}'for v in WORKFLOW['OWN']['PILOT']['prefixes'](58155)if v>=known]
    for rank in range(4):
        for stage in stages:
            name=f's8-rank-{rank}-{stage}.bin';assert (directory/name).read_bytes()==(baseline_directory/name).read_bytes();files.append(name)
    for name in ('s8-code.bin','s8-generator.bin'):
        assert (directory/name).read_bytes()==(baseline_directory/name).read_bytes()
    archive=directory/'s8-archive.raw'
    if source:assert archive.read_bytes()==(source/'s8-archive.raw').read_bytes()
    else:
        support_words=[]
        for rank in range(4):
            data=(baseline_directory/f's8-rank-{rank}-support.bin').read_bytes();support_words.extend(int.from_bytes(data[i:i+8],'little')for i in range(0,len(data),8))
        assert archive.read_bytes()==b''.join(v.to_bytes(8,'little')for v in sorted(support_words))
    return dict(measurements=values,proof_recording=proof,checkpoints=checkpoints,publication=publication,support_verification=support,matched_checkpoint_files=files,scope='Parameterized command;fullsize complete host-phase/topology admission remains')


def suite(c):
    c['run']('command-review-tests',['python3','-B','experiments/distributed-command/test_review.py'],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',PIR_PIPELINE_RESTART_SUPPORT='coordinator',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream')
    c['run']('command-cpu',[c['go'],'test','./encoding/liftrs-shared','./cmd/pir-preprocess-mpi','-run','^TestDistributedJob|^TestCommandResources','-count=1','-timeout=20s'],60)
    binary=c['cache']/'pir-preprocess-mpi'
    c['run']('command-build',[c['go'],'build','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','-o',binary,'./cmd/pir-preprocess-mpi'],120)
    wrapper=c['cache']/'preprocess-rank.sh';shutil.copyfile(c['source']/'cmd/pir-preprocess-mpi/rank.sh',wrapper)
    config=dict(spec=dict(FieldDegree=8,Dimension=2,CurveDegree=5),ranks=4,library=str(c['cache']/'libliftrs-cuda.so'),device=0,device_bytes=64<<20,max_local_words=16384,max_good_words=65536,max_root_words=65536,max_archive_words=0,rss_bytes=1<<30,host_reserve_bytes=768<<20,storage_bytes=16<<20,file_bytes=32<<20,deadline_seconds=220,newton_input='prefix',spectrum='stream',discovery_result='stream',proof_log='trace')
    # File cap is per file and may not exceed aggregate retained output cap.
    config['storage_bytes']=32<<20
    cold=c['cache']/'command-cold';resume=c['cache']/'command-resume'
    cold_config=c['cache']/'command-cold.json';cold_config.write_text(json.dumps(dict(config,directory=str(cold)))+'\n')
    c['run']('command-config-preflight',[binary,'--config',cold_config,'--check-config'],20)
    # The resume config is written just before launch,once the cold source is sealed.
    resume_config=c['cache']/'command-resume.json'
    def cold_review(rows,directory):
        source=directory/'s8-checkpoint-source.json'
        import hashlib
        resume_config.write_text(json.dumps(dict(config,directory=str(resume),resume_directory=str(cold),source_sha256=hashlib.sha256(source.read_bytes()).hexdigest(),known_prefix=8))+'\n')
        return command_review(rows,directory,VECTORS['BASE']['parse_rank_logs'](c['out']/'cold.log',ranks=4),c['cache']/'cold')
    def resume_review(rows,directory):return command_review(rows,directory,VECTORS['BASE']['parse_rank_logs'](c['out']/'resume.log',ranks=4),c['cache']/'resume',8,cold)
    extra=[dict(name='command-cold',command=[binary,'--config',cold_config,'--test-seed=109'],wrapper=wrapper,memcheck=False,profile=True,review=cold_review),dict(name='command-resume',command=[binary,'--config',resume_config],wrapper=wrapper,memcheck=False,profile=True,review=resume_review)]
    result=SCALE['suite'](c,review_fn=BIND['review'],extra_arms=extra)
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4))for mode in ('cold','resume','command-cold','command-resume')}
    c['resources']['command']='Parameterized native MPI server binary;explicit JSON limits/modes,profiles,1-second local RSS/storage/deadline guards;not complete fullsize admission'
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2nodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;pilot cold/restart+command cold/restart;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
