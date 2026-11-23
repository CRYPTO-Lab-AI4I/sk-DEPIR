#!/usr/bin/env python3
"""Assembled runtime compared with independently reviewed pilot/native outputs."""
from pathlib import Path
import hashlib
import json
import runpy
import sys
HERE=Path(__file__).resolve().parent
BIND=runpy.run_path(str(HERE.parent/'distributed-restart-binding/run.py'))
SCALE=BIND['SCALE'];VECTORS=BIND['VECTORS'];PROOF=BIND['PROOF'];DIAG=BIND['DIAG'];OWN=SCALE['OWN']


def workflow_claims(rows,known=1):
    assert known in (1,8)
    values=[r for r in rows if r.get('kind')=='workflow-result']
    assert len(values)==4 and {r['rank']for r in values}==set(range(4))
    for r in values:
        assert (r['case'],r['n'],r['k'],r['samples'])==('s8',65536,7381,2)
        assert r['identity_verified']is True and r['reference_input']is False
        assert r['root_counts']==([3442,1507,1541,895]if known==1 else None)
        assert r.get('resumed',False)==(known>1)
        if known>1:assert r['initial_prefix']==known
        assert r['refinement']['Steps']==(16 if known==1 else 13) and r['refinement']['RootCloneWordsPerStep']==0
    assert sum(r['refinement']['UploadedWords']for r in values)==(161392 if known==1 else 161367)
    return values


def publication_review(rows,directory):
    path=directory/'s8-publication.json';record=json.loads(path.read_text())
    events=[r for r in rows if r.get('kind')=='workflow-publication']
    assert len(events)==1 and events[0]['rank']==0 and events[0]['record']==record
    assert record['version']==1 and record['identity_verified']is True and record['samples']==2 and record['support_format']=='uint64-exponents-le'
    assert record['header']['K']==7381 and record['header']['N']==65536
    assert [f['name']for f in record['files']]==['s8-code.bin','s8-generator.bin','s8-archive.raw']
    assert not list(directory.glob('s8-*.partial'))
    for item in record['files']:
        data=(directory/item['name']).read_bytes()
        assert len(data)==item['bytes']and hashlib.sha256(data).hexdigest()==item['sha256']
    return record


def assembled_review(rows,directory,baseline_rows,baseline_directory,require_publication=False):
    values=workflow_claims(rows)
    proof=PROOF['check_summaries'](rows,baseline_rows,'trace')
    native_metadata=json.loads((directory.parent/'reference-inputs/s8-reference.json').read_text())
    discovery=OWN['DISCOVERY']['review_discovery']([r for r in rows if r.get('discovery')is True],directory,(8,),ranks=4,native_references={8:native_metadata})
    manifest=json.loads((directory.parent/'root-references.json').read_text())
    roots=OWN['review_roots'](rows,directory,manifest,directory.parent/'root-reference-files')
    prefixes=OWN['PILOT']['prefixes'](65536-7381)
    files=[]
    for rank in range(4):
        names=[f's8-rank-{rank}-{suffix}.bin'for suffix in ('support','identity','root')]
        names.extend(f's8-rank-{rank}-newton-{prefix}.bin'for prefix in prefixes)
        for name in names:
            assert (directory/name).read_bytes()==(baseline_directory/name).read_bytes(),name
            files.append(name)
    native=directory.parent/'reference-inputs'
    assert (directory/'s8-code.bin').read_bytes()==(native/'s8-reference-code.bin').read_bytes()==(baseline_directory/'s8-code.bin').read_bytes()
    assert (directory/'s8-generator.bin').read_bytes()==(native/'s8-rank-0-reference-g.bin').read_bytes()==(baseline_directory/'s8-generator.bin').read_bytes()
    support=[]
    for rank in range(4):
        data=(baseline_directory/f's8-rank-{rank}-support.bin').read_bytes()
        support.extend(int.from_bytes(data[i:i+8],'little')for i in range(0,len(data),8))
    expected=b''.join(v.to_bytes(8,'little')for v in sorted(support))
    assert (directory/'s8-archive.raw').read_bytes()==expected
    publication=None
    modes={r.get('publication_mode','direct')for r in values}
    if require_publication:assert modes=={'commit'}
    if modes=={'commit'}:publication=publication_review(rows,directory)
    return dict(publication=publication,measurements=values,proof_recording=proof,discovery=discovery,hierarchy=roots,matched_checkpoint_files=files,canonical_code_sha256=hashlib.sha256((directory/'s8-code.bin').read_bytes()).hexdigest(),scope='Reference-independent runtime;test/controller owns bounded diagnostic persistence,not production restart/publication')


def suite(c,extra_arms=None):
    c['run']('workflow-review-tests',['python3','-B','experiments/distributed-workflow/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_RESTART_SUPPORT='coordinator',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    c['run']('workflow-admission-tests',[c['go'],'test','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','./encoding/liftrs-shared','-run','^TestRankWorkflowAdmission','-count=1','-timeout=20s'],60)
    def assembled(rows,directory):
        reference=VECTORS['BASE']['parse_rank_logs'](c['out']/'cold.log',ranks=4)
        return assembled_review(rows,directory,reference,c['cache']/'cold')
    extra=[dict(name='assembled',test='TestAssembledRankWorkflow',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded'),review=assembled)]
    extra.extend(extra_arms or [])
    result=SCALE['suite'](c,review_fn=BIND['review'],extra_arms=extra)
    modes=('cold','resume','assembled',*(arm['name']for arm in extra_arms or []))
    result['diagnostic_writes']={mode:DIAG['diagnostic_review'](c['out'],mode,c['cache']/mode)for mode in modes}
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4))for mode in modes}
    c['resources'].update(workflow='Reference-independent runtime cold workflow with external output/checkpoint policy;baseline pilot/native checks retained',scope='Genuine s8 reviewed pilot cold/restart plus assembled runtime cold;not fullsize/public production entry/restart')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2nodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;pilot cold/restart+assembled runtime;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
