#!/usr/bin/env python3
"""Root preparation/subgroup lifecycle without test helpers or oracle inputs."""
from pathlib import Path
import hashlib
import json
import subprocess
import runpy
import sys
HERE=Path(__file__).resolve().parent
RECIP=runpy.run_path(str(HERE.parent/'distributed-reciprocal/run.py'))
BIND=RECIP['BIND'];SCALE=RECIP['SCALE'];VECTORS=RECIP['VECTORS'];PROOF=RECIP['PROOF'];DIAG=RECIP['DIAG'];OWN=SCALE['OWN']


def producing_manifest(raw):
    # Build source lives in scratch, not retained cache. Bind review to the
    # producer's immutable Git object and its independently sealed source hash.
    name='experiments/distributed-owned-roots/root-references.json'
    data=subprocess.check_output(['git','show',raw['revision']+':'+name],cwd=VECTORS['BASE']['ROOT'])
    assert hashlib.sha256(data).hexdigest()==raw['source_sha256'][name]
    return json.loads(data)


def admission_review(rows,cases):
    expected={c['label']:c for c in cases};events=[r for r in rows if r.get('kind')=='root-admission']
    assert len(events)==4*len(cases)and {(r['case'],r['rank'])for r in events}=={(label,rank)for label in expected for rank in range(4)}
    for r in events:
        c=expected[r['case']];rank=r['rank'];stages=[s for s in OWN['TREE']['stages'](c['counts'])if s['First']<=rank<s['First']+s['Ranks']]
        largest=max(c['counts'][rank],*(s['Domain']//s['Ranks']for s in stages))
        assert r['leaf_words']==c['counts'][rank]and r['largest_shard_words']==largest
        assert r['max_local_words']==8<<20 and largest<=r['max_local_words']and r['base_word_bytes']==4
        assert r['max_domain']==1<<(2*c['s'])and r['stages']==len(stages)
    return events


def review(rows,directory,resume=False,source=None,cold=None):
    result=BIND['review'](rows,directory,resume,source,cold)
    manifest=json.loads((directory.parent/'root-references.json').read_text())
    if not resume:admission_review(rows,manifest['cases'])
    else:assert not any(r.get('kind')=='root-admission'for r in rows)
    return result


def wide_review(rows,directory,manifest,references):
    result=OWN['review_roots'](rows,directory,manifest,references,wide=True)
    admission_review(rows,[c for c in manifest['cases']if c['s']==18])
    PROOF['check_summaries'](rows,rows,'trace')
    return result


def suite(c):
    c['run']('root-runtime-review-tests',['python3','-B','experiments/distributed-root-runtime/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_RESTART_SUPPORT='coordinator',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('root-runtime-cpu',[]),('root-runtime-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestRankRoot|^TestRootShard','-count=1','-timeout=20s'],60)
    manifest=json.loads((c['source']/'experiments/distributed-owned-roots/root-references.json').read_text())
    references=c['cache']/'owned-wide-reference-files'
    total=OWN['copy_references'](VECTORS['BASE']['ROOT'],manifest,references)
    c['resources']['owned_wide_reference_bytes']=total;c['save']()
    c['run']('wide-reference-preflight',['srun','--nodes=2','--ntasks=4','--ntasks-per-node=2','--cpus-per-task=1','--export=ALL,TMPDIR=/tmp','--chdir='+str(c['cache']),'bash','-c','set -eu; for p in "$1"/*; do test -r "$p"; done','wide-reference-preflight',references],30)
    extra=[dict(name='owned-wide',test='TestOwnedWideRoots',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded'),review=lambda rows,directory:wide_review(rows,directory,manifest,references))]
    result=SCALE['suite'](c,review_fn=review,extra_arms=extra)
    modes=('cold','resume','owned-wide')
    result['diagnostic_writes']={mode:DIAG['diagnostic_review'](c['out'],mode,c['cache']/mode)for mode in modes}
    result['native_taylor_totals']={mode:PROOF['native_review'](c['out'],mode,VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4))for mode in modes}
    c['resources'].update(root_runtime='Borrowed local engine/shared owned MPI subgroup lifecycle;actual-count host-shard admission,not full resource feasibility',scope='Genuine s8 cold/coordinated restart and bounded field18 owned roots;not fulls18/public production entry')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;root runtime cold/restart/owned-wide;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
