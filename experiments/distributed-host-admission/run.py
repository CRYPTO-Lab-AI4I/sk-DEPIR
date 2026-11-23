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


def recording_review(rows,mode,ranks=8):
    if mode=='trace':return PROOF['check_summaries'](rows,rows,'trace')
    assert mode=='summary'and not any(r.get('kind')in PROOF['KINDS']for r in rows)
    footers=[r for r in rows if r.get('kind')=='proof-summary'];seen=set()
    for r in footers:
        key=r['rank'],r['proof_kind'];assert key not in seen and 0<=r['rank']<ranks and r['proof_kind']in PROOF['KINDS'];seen.add(key)
        assert r['complete']is True and r['mode']=='summary'and r['codec']=='go-json-v1-length64le'
        assert r['records']>0 and r['semantic_bytes']>0 and len(bytes.fromhex(r['sha256']))==32
    assert {(rank,kind)for rank in range(ranks)for kind in ('tile-proof','taylor-proof','operand-proof')}<=seen
    return dict(records=sum(r['records']for r in footers),kinds=len(seen),stored_proof_rows=0,summaries=len(footers),scope='Runtime checks and recorded ordered digests;no independent full-row replay for this shape')


def review(rows,directory,known=1,source=None,reference=REFERENCE,proof_mode='trace'):
    h=reference['header'];s=h['Spec']['FieldDegree'];n,k=h['N'],h['K'];label=f's{s}';words=n//8
    selected={name:v for name,v in reference['stages'].items()if name=='root'or v['prefix']>=max(1,known)}
    expected_steps=sum(name!='root'and v['prefix']>max(1,known)for name,v in reference['stages'].items())
    uploads=0;previous=1
    for prefix in sorted(v['prefix']for name,v in reference['stages'].items()if name!='root'):
        if prefix>max(1,known):uploads+=min(k+1,prefix)+min(2*previous-1,prefix)
        previous=prefix
    values=[r for r in rows if r.get('kind')=='workflow-result']
    assert len(values)==8 and {r['rank']for r in values}==set(range(8))
    samples={r['samples']for r in values};assert len(samples)==1 and 1<=next(iter(samples))<=2+(128+s-1)//s
    if reference.get('samples',2)is not None:assert samples=={reference.get('samples',2)}
    for r in values:
        assert (r['case'],r['n'],r['k'])==(label,n,k)
        assert r['entry']=='command'and r['identity_verified']is True and r['reference_input']is False
        assert r['resumed']==(source is not None)and r['refinement']['Steps']==expected_steps
        assert r['root_counts']==(reference['root_counts']if known<=1 and not(source and known==1)else None)
        assert r['refinement']['RootCloneWordsPerStep']==0
    assert sum(r['refinement']['UploadedWords']for r in values)==uploads
    assert len([r for r in rows if r.get('kind')=='job-complete'])==8
    recording=recording_review(rows,proof_mode)
    publication=json.loads((directory/f'{label}-publication.json').read_text())
    events=[r for r in rows if r.get('kind')=='workflow-publication'];assert len(events)==1 and events[0]['rank']==0 and events[0]['record']==publication
    assert publication['header']==h and publication['version']==1 and publication['identity_verified']is True and publication['samples']in samples and publication['support_format']=='uint64-exponents-le'
    assert [v['name']for v in publication['files']]==[f'{label}-code.bin',f'{label}-generator.bin',f'{label}-archive.raw']
    for item in publication['files']:
        data=(directory/item['name']).read_bytes();assert item==reference['files'][item['name']]
        assert len(data)==item['bytes']and hashlib.sha256(data).hexdigest()==item['sha256']
    source_bytes=(directory/f'{label}-checkpoint-source.json').read_bytes();source_sha=hashlib.sha256(source_bytes).hexdigest();metadata=json.loads(source_bytes)
    assert metadata['version']==1 and metadata['SupportFormat']=='uint64-exponents-le'
    assert metadata['header']==h and metadata['Ranks']==8 and metadata['Samples']in samples
    assert metadata['SupportSHA']==reference['files'][f'{label}-archive.raw']['sha256']and metadata['SupportBytes']==8*k
    assert os.path.samefile(directory/f'{label}-support.bin',directory/f'{label}-archive.raw')
    commits={};events=[r for r in rows if r.get('kind')=='workflow-checkpoint']
    for stage,pinned in selected.items():
        record=json.loads((directory/f'{label}-{stage}-commit.json').read_text())
        assert record['ranks']==8 and record['prefix']==pinned['prefix']and record['stage']==stage
        assert (record['s'],record['k'],record['n'],record['layout'])==(s,k,n,'prefix')
        assert record['source_sha256']==source_sha and record['support_sha256']==metadata['SupportSHA']and len(record['sha256'])==8
        digest=hashlib.sha256();size=0
        for rank,expected in enumerate(record['sha256']):
            data=(directory/f'{label}-rank-{rank}-{stage}.bin').read_bytes()
            assert len(data)==4*min(words,max(0,record['prefix']-rank*words))and hashlib.sha256(data).hexdigest()==expected
            digest.update(data);size+=len(data)
        assert size==pinned['bytes']and digest.hexdigest()==pinned['sha256']
        matched=[r for r in events if r['record']['stage']==stage]
        assert len(matched)==8 and {r['rank']for r in matched}==set(range(8))and all(r['record']==record for r in matched)
        commits[stage]=record
    assert len(events)==8*len(commits)
    phases=[r for r in rows if r.get('kind')=='host-phase-admission']
    wanted={'checkpoint-reload','Newton','identity','spectrum-and-publication'}if source and known>=1 else{'local-roots','root-group-1','root-group-2','root-group-3','root-reshape','Newton','identity','spectrum-and-publication'}|({'good-enumeration','discovery'}if known==1 else{'cached-discovery'})
    assert {(r['rank'],r['inventory']['phase'])for r in phases}=={(rank,phase)for rank in range(8)for phase in wanted}
    assert len(phases)==8*len(wanted)
    for r in phases:
        p=r['inventory'];assert sum(p['components_bytes'].values())==p['named_bytes']
        assert r['reserve_bytes']==768<<20 and r['rss_cap_bytes']==1<<30
        assert p['named_bytes']+r['reserve_bytes']<=r['rss_cap_bytes']
        if p['phase']=='Newton':assert p['components_bytes']['base_coefficients']==4*4*words
        if p['phase']=='local-roots':assert p['components_bytes']['base_coefficients']==6*reference['root_counts'][r['rank']]*4
        if p['phase'].startswith('root-group-'):
            level=int(p['phase'].rsplit('-',1)[1]);stage=[v['stage']for v in rows if v.get('kind')=='root-stage'and v['event']=='started'and v['rank']==r['rank']and v['stage']['Level']==level]
            assert len(stage)==1 and p['components_bytes']['base_coefficients']==6*(stage[0]['Domain']//stage[0]['Ranks'])*4
    if source:
        assert os.path.samefile(source/f'{label}-support.bin',directory/f'{label}-archive.raw')
        if known==0:
            claims=[r for r in rows if r.get('kind')=='cached-discovery'];assert len(claims)==8 and {r['rank']for r in claims}==set(range(8))
            for r in claims:
                assert r['verified']is True and r['hash_passes']==(1 if r['rank']==0 else 0)and r['hashed_bytes']==(8*k if r['rank']==0 else 0)
                assert r['source_sha256']==hashlib.sha256((source/f'{label}-checkpoint-source.json').read_bytes()).hexdigest()
            assert not any(r.get('discovery')is True for r in rows)
        else:
            claims=[r for r in rows if r.get('kind')=='restart-support'];assert len(claims)==8 and {r['rank']for r in claims}==set(range(8))
            for r in claims:assert r['verified']is True and r['mode']=='coordinator'and r['hash_passes']==(1 if r['rank']==0 else 0)and r['hashed_bytes']==(8*k if r['rank']==0 else 0)and r['agreements']==(1 if known==1 else 2)
            agreements=[r for r in rows if r.get('kind')=='restart-support-agreement'];assert len(agreements)==8*(1 if known==1 else 2)and {(r['rank'],r['checkpoint'])for r in agreements}=={(rank,stage)for rank in range(8)for stage in (('root',)if known==1 else('root','newton-8'))}
            for r in agreements:
                record=json.loads((source/f'{label}-{r["checkpoint"]}-commit.json').read_text());digest=COMMAND['BIND']['binding_frame'](record,source,label)
                assert r['frames']==[digest+(record['support_sha256']if rank==0 else'00'*32)for rank in range(8)]
    assert not list(directory.glob('*.partial'))
    scope='Eight workers on two GPUs;not eight GPUs/fullsize host feasibility'if s==8 else'Representative s10 recovery;not fullsize host-memory evidence'
    return dict(measurements=values,proof_recording=recording,publication=publication,commits=commits,host_phases=phases,scope=scope)


def suite(c,reference=REFERENCE,proof_mode='trace',recoveries=(8,),seed=109):
    if os.environ.get('SLURM_NTASKS')!='8':raise RuntimeError('requires admitted eight-task allocation')
    h=reference['header'];s=h['Spec']['FieldDegree'];words=h['N']//8;label=f's{s}'
    c['resources'].update(nodes=2,ranks=8,cpus_per_node=4,host_request_GiB_per_node=8,rss_cap_GiB_per_rank=1)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',LIFTRS_GPU_TAYLOR_PROGRESS='bounded')
    c['run']('host-review-tests',['python3','-B','experiments/distributed-host-admission/test_review.py'],30)
    c['run']('host-cpu-tests',[c['go'],'test','./encoding/liftrs-shared','./cmd/pir-preprocess-mpi','-run','^TestHostPhase|^TestDistributedJob|^TestCommandResources|^TestArchive','-count=1','-timeout=20s'],60)
    binary=c['cache']/'pir-preprocess-mpi';c['run']('command-build',[c['go'],'build','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','-o',binary,'./cmd/pir-preprocess-mpi'],120)
    wrapper=c['cache']/'preprocess-rank.sh';shutil.copyfile(c['source']/'cmd/pir-preprocess-mpi/rank.sh',wrapper)
    config=dict(spec=h['Spec'],ranks=8,library=str(c['cache']/'libliftrs-cuda.so'),device=0,device_bytes=32<<20,max_local_words=words,max_good_words=words,max_root_words=words,max_archive_words=0,rss_bytes=1<<30,host_reserve_bytes=768<<20,storage_bytes=32<<20,file_bytes=32<<20,deadline_seconds=220,newton_input='prefix',spectrum='stream',discovery_result='stream',proof_log=proof_mode)
    cold=c['cache']/'cold';cold_config=c['cache']/'cold.json'
    cold_config.write_text(json.dumps(dict(config,directory=str(cold)))+'\n');c['run']('config-preflight',[binary,'--config',cold_config,'--check-config'],20)
    recovery_names={keep:('resume'if recoveries==(8,)else('discovery-resume'if keep==0 else('root-resume'if keep==1 else'prefix-resume')))for keep in recoveries}
    def cold_review(rows,directory):
        result=review(rows,directory,reference=reference,proof_mode=proof_mode)
        sha=hashlib.sha256((cold/f'{label}-checkpoint-source.json').read_bytes()).hexdigest()
        for keep,name in recovery_names.items():
            (c['cache']/(name+'.json')).write_text(json.dumps(dict(config,directory=str(c['cache']/name),resume_directory=str(cold),source_sha256=sha,known_prefix=keep))+'\n')
        return result
    arms=[dict(name='cold',command=[binary,'--config',cold_config,'--test-seed='+str(seed)],wrapper=wrapper,memcheck=False,profile=True,review=cold_review)]
    for keep,name in recovery_names.items():
        arms.append(dict(name=name,command=[binary,'--config',c['cache']/(name+'.json')],wrapper=wrapper,memcheck=False,profile=True,review=lambda rows,d,keep=keep:review(rows,d,keep,cold,reference,proof_mode)))
    result=VECTORS['suite'](c,arms=arms,ranks=8,cpu_cores=8,scope='Eight-worker command recovery/host admission;not eight-GPU bandwidth/fullsize')
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=8),workers=8)for mode in ['cold',*recovery_names.values()]}
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('NOT SUBMITTED:requires approval for8CPU/16GiBhost total;2nodes/2H200/8workers,32MiBownedGPU/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min;no fulls18')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
