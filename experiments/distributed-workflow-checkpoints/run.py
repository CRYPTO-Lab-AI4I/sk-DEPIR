#!/usr/bin/env python3
"""Durable assembled prefix commits with source binding and fresh-process reload."""
from pathlib import Path
import hashlib
import json
import os
import runpy
import sys
HERE=Path(__file__).resolve().parent
PUBLICATION=runpy.run_path(str(HERE.parent/'distributed-publication-commit/run.py'))
WORKFLOW=PUBLICATION['WORKFLOW'];BIND=PUBLICATION['BIND'];SCALE=PUBLICATION['SCALE'];VECTORS=PUBLICATION['VECTORS'];PROOF=PUBLICATION['PROOF'];DIAG=PUBLICATION['DIAG']
assembled_review=PUBLICATION['assembled_review']


def checkpoint_review(rows,directory):
    """Check actual persisted bytes, every rank's commit, and the no-copy alias."""
    assert {r.get('checkpoint_mode')for r in WORKFLOW['workflow_claims'](rows)}=={'commit'}
    source_bytes=(directory/'s8-checkpoint-source.json').read_bytes()
    source=json.loads(source_bytes);source_sha=hashlib.sha256(source_bytes).hexdigest()
    publication=json.loads((directory/'s8-publication.json').read_text())
    assert source['version']==1 and source['header']==publication['header']
    assert source['Ranks']==4 and source['Samples']==publication['samples']==2
    assert source['SupportFormat']==publication['support_format']=='uint64-exponents-le'
    archive=(directory/'s8-archive.raw').read_bytes()
    assert source['SupportBytes']==len(archive)==8*7381
    assert source['SupportSHA']==hashlib.sha256(archive).hexdigest()
    assert os.path.samefile(directory/'s8-support.bin',directory/'s8-archive.raw')
    stages=[('root',7382)]+[(f'newton-{v}',v)for v in WORKFLOW['OWN']['PILOT']['prefixes'](65536-7381)]
    events=[r for r in rows if r.get('kind')=='workflow-checkpoint']
    assert len(events)==4*len(stages)
    records=[]
    for stage,prefix in stages:
        record=json.loads((directory/f's8-{stage}-commit.json').read_text())
        expected=dict(layout='prefix',s=8,k=7381,n=65536,prefix=prefix,ranks=4,support_sha256=source['SupportSHA'],stage=stage,source_sha256=source_sha)
        assert {k:v for k,v in record.items()if k!='sha256'}==expected
        assert len(record['sha256'])==4
        matching=[r for r in events if r['record']['stage']==stage]
        assert len(matching)==4 and {r['rank']for r in matching}==set(range(4))
        assert all(r['case']=='s8'and r['record']==record for r in matching)
        for rank,digest in enumerate(record['sha256']):
            data=(directory/f's8-rank-{rank}-{stage}.bin').read_bytes()
            assert len(data)==4*min(16384,max(0,prefix-rank*16384))
            assert hashlib.sha256(data).hexdigest()==digest
        records.append(record)
    assert not list(directory.glob('s8-*.partial'))
    return dict(source=source,source_sha256=source_sha,commits=records,support_alias='same inode;no payload copy',scope='Durable files and fresh-process loader;not arithmetic resume')


def suite(c):
    c['run']('checkpoint-review-tests',['python3','-B','experiments/distributed-workflow-checkpoints/test_review.py'],30)
    # Exercise required shared-filesystem semantics before any GPU work.
    c['run']('checkpoint-link-preflight',['python3','-B','-c','import os,pathlib,tempfile;\nwith tempfile.TemporaryDirectory(dir="'+str(c['cache'])+'") as d:\n p=pathlib.Path(d);(p/"a").write_bytes(b"link");os.link(p/"a",p/"b");assert os.path.samefile(p/"a",p/"b");f=os.open(d,os.O_RDONLY);os.fsync(f);os.close(f)'],10)
    c['env']['PIR_WORKFLOW_CHECKPOINT']='commit'
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('checkpoint-cpu',[]),('checkpoint-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestWorkflowCheckpoint|^TestRankPrefix|^TestRankSupport','-count=1','-timeout=20s'],60)
    result=PUBLICATION['suite'](c)
    rows=VECTORS['BASE']['parse_rank_logs'](c['out']/'assembled.log',ranks=4)
    result['workflow_checkpoints']=checkpoint_review(rows,c['cache']/'assembled')
    c['env']['PIR_WORKFLOW_RELOAD_SOURCE']=str(c['cache']/'assembled')
    c['run']('checkpoint-fresh-reload',[c['cache']/'vectors.test','-test.run=^TestReloadWorkflowCheckpoints$','-test.v','-test.timeout=20s','-test.cpuprofile='+str(c['out']/'checkpoint-reload.pprof')],30)
    c['resources']['workflow_checkpoints']='Explicit durable source-bound compact commits;fresh-process root/prefix8 file validation;not arithmetic resume'
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2nodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;pilot cold/restart+durable assembled commits+fresh CPU reload;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
