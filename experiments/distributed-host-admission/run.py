#!/usr/bin/env python3
"""Eight-worker command correctness and phase admission;requires new resource approval."""
from pathlib import Path
import hashlib
import json
import os
import runpy
import shutil
import sys
HERE=Path(__file__).resolve().parent
COMMAND=runpy.run_path(str(HERE.parent/'distributed-command/run.py'))
VECTORS=COMMAND['VECTORS'];PROOF=COMMAND['PROOF'];REFERENCE=runpy.run_path(str(HERE/'reference.py'))['REFERENCE']


def review(rows,directory,known=1,source=None):
    values=[r for r in rows if r.get('kind')=='workflow-result']
    assert len(values)==8 and {r['rank']for r in values}==set(range(8))
    for r in values:
        assert (r['case'],r['n'],r['k'],r['samples'])==('s8',65536,7381,2)
        assert r['entry']=='command'and r['identity_verified']is True and r['reference_input']is False
        assert r['resumed']==(known>1)and r['refinement']['Steps']==(16 if known==1 else 13)
        assert r['root_counts']==(REFERENCE['root_counts']if known==1 else None)
        assert r['refinement']['RootCloneWordsPerStep']==0
    assert sum(r['refinement']['UploadedWords']for r in values)==(161392 if known==1 else 161367)
    assert len([r for r in rows if r.get('kind')=='job-complete'])==8
    recording=PROOF['check_summaries'](rows,rows,'trace')
    publication=COMMAND['WORKFLOW']['publication_review'](rows,directory)
    for item in publication['files']:
        data=(directory/item['name']).read_bytes();assert item==REFERENCE['files'][item['name']]
        assert len(data)==item['bytes']and hashlib.sha256(data).hexdigest()==item['sha256']
    source_bytes=(directory/'s8-checkpoint-source.json').read_bytes();source_sha=hashlib.sha256(source_bytes).hexdigest();metadata=json.loads(source_bytes)
    assert metadata['version']==1 and metadata['SupportFormat']=='uint64-exponents-le'
    assert metadata['header']==publication['header']==REFERENCE['header']and metadata['Ranks']==8 and metadata['Samples']==2
    assert metadata['SupportSHA']==REFERENCE['files']['s8-archive.raw']['sha256']and metadata['SupportBytes']==8*7381
    assert os.path.samefile(directory/'s8-support.bin',directory/'s8-archive.raw')
    commits={};events=[r for r in rows if r.get('kind')=='workflow-checkpoint']
    for stage,reference in REFERENCE['stages'].items():
        if stage!='root'and reference['prefix']<known:continue
        record=json.loads((directory/f's8-{stage}-commit.json').read_text())
        assert record['ranks']==8 and record['prefix']==reference['prefix']and record['stage']==stage
        assert (record['s'],record['k'],record['n'],record['layout'])==(8,7381,65536,'prefix')
        assert record['source_sha256']==source_sha and record['support_sha256']==metadata['SupportSHA']and len(record['sha256'])==8
        digest=hashlib.sha256();size=0
        for rank,expected in enumerate(record['sha256']):
            data=(directory/f's8-rank-{rank}-{stage}.bin').read_bytes()
            assert len(data)==4*min(8192,max(0,record['prefix']-rank*8192))and hashlib.sha256(data).hexdigest()==expected
            digest.update(data);size+=len(data)
        assert size==reference['bytes']and digest.hexdigest()==reference['sha256']
        matched=[r for r in events if r['record']['stage']==stage]
        assert len(matched)==8 and {r['rank']for r in matched}==set(range(8))and all(r['record']==record for r in matched)
        commits[stage]=record
    assert len(events)==8*len(commits)and len(commits)==(17 if known==1 else 15)
    phases=[r for r in rows if r.get('kind')=='host-phase-admission']
    wanted={'checkpoint-reload','Newton','identity','spectrum-and-publication'}if known>1 else{'good-enumeration','discovery','local-roots','root-group-1','root-group-2','root-group-3','root-reshape','Newton','identity','spectrum-and-publication'}
    assert {(r['rank'],r['inventory']['phase'])for r in phases}=={(rank,phase)for rank in range(8)for phase in wanted}
    assert len(phases)==8*len(wanted)
    for r in phases:
        p=r['inventory'];assert sum(p['components_bytes'].values())==p['named_bytes']
        assert r['reserve_bytes']==768<<20 and r['rss_cap_bytes']==1<<30
        assert p['named_bytes']+r['reserve_bytes']<=r['rss_cap_bytes']
        if p['phase']=='Newton':assert p['components_bytes']['base_coefficients']==4*4*8192
        if p['phase']=='local-roots':assert p['components_bytes']['base_coefficients']==6*REFERENCE['root_counts'][r['rank']]*4
        if p['phase'].startswith('root-group-'):
            level=int(p['phase'].rsplit('-',1)[1]);stage=[v['stage']for v in rows if v.get('kind')=='root-stage'and v['event']=='started'and v['rank']==r['rank']and v['stage']['Level']==level]
            assert len(stage)==1 and p['components_bytes']['base_coefficients']==6*(stage[0]['Domain']//stage[0]['Ranks'])*4
    if source:
        assert os.path.samefile(source/'s8-support.bin',directory/'s8-archive.raw')
        claims=[r for r in rows if r.get('kind')=='restart-support'];assert len(claims)==8 and {r['rank']for r in claims}==set(range(8))
        for r in claims:assert r['verified']is True and r['mode']=='coordinator'and r['hash_passes']==(1 if r['rank']==0 else 0)and r['hashed_bytes']==(8*7381 if r['rank']==0 else 0)and r['agreements']==2
        agreements=[r for r in rows if r.get('kind')=='restart-support-agreement'];assert len(agreements)==16 and {(r['rank'],r['checkpoint'])for r in agreements}=={(rank,stage)for rank in range(8)for stage in ('root','newton-8')}
        for r in agreements:
            record=json.loads((source/f's8-{r["checkpoint"]}-commit.json').read_text());digest=COMMAND['BIND']['binding_frame'](record,source,'s8')
            assert r['frames']==[digest+(record['support_sha256']if rank==0 else'00'*32)for rank in range(8)]
    assert not list(directory.glob('*.partial'))
    return dict(measurements=values,proof_recording=recording,publication=publication,commits=commits,host_phases=phases,scope='Eight workers on two GPUs;not eight GPUs/fullsize host feasibility')


def suite(c):
    if os.environ.get('SLURM_NTASKS')!='8':raise RuntimeError('requires admitted eight-task allocation')
    c['resources'].update(nodes=2,ranks=8,cpus_per_node=4,host_request_GiB_per_node=8,rss_cap_GiB_per_rank=1)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',LIFTRS_GPU_TAYLOR_PROGRESS='bounded')
    c['run']('host-review-tests',['python3','-B','experiments/distributed-host-admission/test_review.py'],30)
    c['run']('host-cpu-tests',[c['go'],'test','./encoding/liftrs-shared','./cmd/pir-preprocess-mpi','-run','^TestHostPhase|^TestDistributedJob|^TestCommandResources','-count=1','-timeout=20s'],60)
    binary=c['cache']/'pir-preprocess-mpi';c['run']('command-build',[c['go'],'build','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','-o',binary,'./cmd/pir-preprocess-mpi'],120)
    wrapper=c['cache']/'preprocess-rank.sh';shutil.copyfile(c['source']/'cmd/pir-preprocess-mpi/rank.sh',wrapper)
    config=dict(spec=dict(FieldDegree=8,Dimension=2,CurveDegree=5),ranks=8,library=str(c['cache']/'libliftrs-cuda.so'),device=0,device_bytes=32<<20,max_local_words=8192,max_good_words=8192,max_root_words=8192,max_archive_words=0,rss_bytes=1<<30,host_reserve_bytes=768<<20,storage_bytes=32<<20,file_bytes=32<<20,deadline_seconds=220,newton_input='prefix',spectrum='stream',discovery_result='stream',proof_log='trace')
    cold=c['cache']/'cold';cold_config=c['cache']/'cold.json';resume_config=c['cache']/'resume.json'
    cold_config.write_text(json.dumps(dict(config,directory=str(cold)))+'\n');c['run']('config-preflight',[binary,'--config',cold_config,'--check-config'],20)
    def cold_review(rows,directory):
        result=review(rows,directory)
        resume_config.write_text(json.dumps(dict(config,directory=str(c['cache']/'resume'),resume_directory=str(cold),source_sha256=hashlib.sha256((cold/'s8-checkpoint-source.json').read_bytes()).hexdigest(),known_prefix=8))+'\n')
        return result
    arms=[dict(name='cold',command=[binary,'--config',cold_config,'--test-seed=109'],wrapper=wrapper,memcheck=False,profile=True,review=cold_review),dict(name='resume',command=[binary,'--config',resume_config],wrapper=wrapper,memcheck=False,profile=True,review=lambda rows,d:review(rows,d,8,cold))]
    result=VECTORS['suite'](c,arms=arms,ranks=8,cpu_cores=8,scope='Eight-worker command with host phase admission;not eight-GPU bandwidth/fullsize')
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=8),workers=8)for mode in ('cold','resume')}
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('NOT SUBMITTED:requires approval for8CPU/16GiBhost total;2nodes/2H200/8workers,32MiBownedGPU/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min;no fulls18')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
