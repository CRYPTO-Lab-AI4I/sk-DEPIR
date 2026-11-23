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


def workflow_claims(rows):
    values=[r for r in rows if r.get('kind')=='workflow-result']
    assert len(values)==4 and {r['rank']for r in values}==set(range(4))
    for r in values:
        assert (r['case'],r['n'],r['k'],r['samples'])==('s8',65536,7381,2)
        assert r['identity_verified']is True and r['reference_input']is False
        assert r['root_counts']==[3442,1507,1541,895]
        assert r['refinement']['Steps']==16 and r['refinement']['RootCloneWordsPerStep']==0
    assert sum(r['refinement']['UploadedWords']for r in values)==161392
    return values


def assembled_review(rows,directory,baseline_rows,baseline_directory):
    values=workflow_claims(rows)
    proof=PROOF['check_summaries'](rows,baseline_rows,'trace')
    discovery=OWN['DISCOVERY']['review_discovery']([r for r in rows if r.get('discovery')is True],directory,(8,),ranks=4)
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
    return dict(measurements=values,proof_recording=proof,discovery=discovery,hierarchy=roots,matched_checkpoint_files=files,canonical_code_sha256=hashlib.sha256((directory/'s8-code.bin').read_bytes()).hexdigest(),scope='Reference-independent runtime;test/controller owns bounded diagnostic persistence,not production restart/publication')


def suite(c):
    c['run']('workflow-review-tests',['python3','-B','experiments/distributed-workflow/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_RESTART_SUPPORT='coordinator',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    c['run']('workflow-admission-tests',[c['go'],'test','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','./encoding/liftrs-shared','-run','^TestRankWorkflowAdmission','-count=1','-timeout=20s'],60)
    def assembled(rows,directory):
        reference=VECTORS['BASE']['parse_rank_logs'](c['out']/'cold.log',ranks=4)
        return assembled_review(rows,directory,reference,c['cache']/'cold')
    extra=[dict(name='assembled',test='TestAssembledRankWorkflow',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded'),review=assembled)]
    result=SCALE['suite'](c,review_fn=BIND['review'],extra_arms=extra)
    modes=('cold','resume','assembled')
    result['diagnostic_writes']={mode:DIAG['diagnostic_review'](c['out'],mode,c['cache']/mode)for mode in modes}
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4))for mode in modes}
    c['resources'].update(workflow='Reference-independent runtime cold workflow with external output/checkpoint policy;baseline pilot/native checks retained',scope='Genuine s8 reviewed pilot cold/restart plus assembled runtime cold;not fullsize/public production entry/restart')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2nodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;pilot cold/restart+assembled runtime;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
