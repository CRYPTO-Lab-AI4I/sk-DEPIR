#!/usr/bin/env python3
"""Larger genuine cold/restart using the maintained owned pipeline and reviewers."""
from pathlib import Path
import hashlib
import json
import runpy
import sys
HERE=Path(__file__).resolve().parent
COMPACT=runpy.run_path(str(HERE.parent/'distributed-prefix-checkpoints/run.py'))
OWN=COMPACT['OWN'];VECTORS=OWN['VECTORS'];RESTART=OWN['FOUR']['RESTART']
CASES=(8,)


def reference_manifest(directory,root):
    """Bind exact native root files to reviewed tree geometry before GPU launch."""
    cases=[json.loads((directory/f's{s}-roots.json').read_text())for s in CASES]
    files={}
    for case in cases:
        s,label=case['s'],case['label'];assert label==f's{s}'and s in CASES
        counts=case['counts'];assert len(counts)==4 and sum(counts)==case['count']+4
        names=[f'{label}-rank-{r}-{stage}.bin'for r in range(4)for stage in ('leaf','final')]
        names.extend(f'{label}-level-{st["Level"]}-group-{st["First"]}-rank-{r}-output.bin'for st in OWN['TREE']['stages'](counts)for r in range(st['Ranks']))
        for name in names:
            path=directory/name;data=path.read_bytes()
            files[name]=dict(source=str(path.relative_to(root)),bytes=len(data),sha256=hashlib.sha256(data).hexdigest())
    result=dict(cases=cases,files=files,provenance='Independent native export from fixed seed109;not distributed arithmetic inputs')
    OWN['reference_map'](result)
    return result


def canonical(directory):
    return{s:json.loads((directory/f's{s}-reference.json').read_text())['sha256'][f's{s}-reference-code.bin']for s in CASES}


def review(rows,directory,resume=False,source=None,cold=None):
    result=OWN['review'](rows,directory,resume,source,cold,cases=CASES,canonical=canonical(directory.parent/'reference-inputs'),curve=5)
    assert all(r['checkpoint_layout']=='prefix'for r in result['measurements'])
    return result


def suite(c,review_fn=review,extra_arms=None):
    c['run']('scaling-review-tests',['python3','-B','experiments/distributed-pipeline-scaling/test_review.py'],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',PIR_PIPELINE_CASES='8',PIR_PIPELINE_CURVE='5',PIR_PIPELINE_SMALL='0',PIR_PIPELINE_ROOT_MODE='owned',PIR_PIPELINE_CHECKPOINT_LAYOUT='prefix')
    target=c['cache']/'reference-inputs';target.mkdir()
    binary=c['cache']/'reference.test'
    c['run']('native-reference-build',[c['go'],'test','-c','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','-o',binary,'./encoding/liftrs-shared'],120)
    c['env'].update(PIR_PIPELINE_REFERENCE_EXPORT=str(target),PIR_PIPELINE_ROOT_EXPORT='1')
    c['run']('native-reference-export',[binary,'-test.run=^TestExportPilotReferences$','-test.v','-test.timeout=25s','-test.cpuprofile='+str(c['out']/'native-reference.pprof')],30)
    c['env'].pop('PIR_PIPELINE_REFERENCE_EXPORT');c['env'].pop('PIR_PIPELINE_ROOT_EXPORT')
    manifest=reference_manifest(target,VECTORS['BASE']['ROOT'])
    (c['cache']/'root-references.json').write_text(json.dumps(manifest,indent=2)+'\n')
    total=OWN['copy_references'](VECTORS['BASE']['ROOT'],manifest,c['cache']/'root-reference-files')
    c['env']['PIR_PIPELINE_REFERENCE_FROM']=str(target)
    c['run']('native-input-admission',[binary,'-test.run=^TestPilotHierarchyReferences$','-test.v','-test.timeout=20s'],25)
    # Verify shared reference paths/cwd on every worker, before costly GPU phases.
    c['run']('reference-preflight',['srun','--nodes=2','--ntasks=4','--ntasks-per-node=2','--cpus-per-task=1','--export=ALL,TMPDIR=/tmp','--chdir='+str(c['cache']),
        'bash','-c','set -eu; for p in "$1"/* "$2"/*; do test -r "$p"; done','reference-preflight',target,c['cache']/'root-reference-files'],30)
    c['resources'].update(reference_cases=list(CASES),root_reference_bytes=total,reference_strategy='Independent native CPU export,retained exact leaf/stage/final/canonical data;not GPU arithmetic inputs')
    c['save']()
    result=RESTART['suite'](c,reference_from=target,review_fn=review_fn,ranks=4,extra_arms=extra_arms)
    result['checkpoint_storage']={mode:COMPACT['storage'](c['cache']/mode,cases=CASES)for mode in ('cold','resume')}
    c['resources'].update(scope='Genuine complete s8 discovery/owned roots/Newton/identity/spectrum/canonical publication plus fresh prefix8 restart;not fulls18',checkpoint_layout='explicit prefix;legacy padded default retained')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min,0CUDApinned;genuine s8 cold/fresh-prefix8 restart;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
