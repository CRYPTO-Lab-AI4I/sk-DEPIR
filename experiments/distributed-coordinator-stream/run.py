#!/usr/bin/env python3
"""Bound coordinator diagnostics in the assembled cold/restart GPU pilot."""
from pathlib import Path
import hashlib
import json
import runpy
import shutil
import sys
HERE=Path(__file__).resolve().parent
RESTART=runpy.run_path(str(HERE.parent/'distributed-restart-pilot/run.py'))
VECTORS=RESTART['VECTORS']


def digest(path):
    h=hashlib.sha256()
    with path.open('rb')as stream:
        for block in iter(lambda:stream.read(16<<10),b''):h.update(block)
    return h.hexdigest()


def copy_inputs(root,manifest,target):
    source=Path(manifest['source']);assert not source.is_absolute()and '..'not in source.parts
    target.mkdir();total=0
    assert set(manifest['files'])=={f's{s}-{suffix}'for s in (4,6)for suffix in ('reference.json','support.bin','reference-code.bin','rank-0-reference-p.bin','rank-0-reference-g.bin','rank-0-reference-spectrum.bin')}
    for name,record in manifest['files'].items():
        assert Path(name).name==name and 0<record['bytes']<32<<20
        src=root/source/name;assert src.is_file()and src.stat().st_size==record['bytes']and digest(src)==record['sha256']
        dst=target/name;shutil.copyfile(src,dst);assert digest(dst)==record['sha256'];dst.chmod(0o400);total+=record['bytes']
    return total


def review(rows,directory,resume=False,source=None,cold=None):
    result=RESTART['review'](rows,directory,resume,source,cold)
    for r in result['measurements']:assert r['reference_mode']=='retained'and r['support_streamed']is True
    if resume:assert not any(r.get('phase')=='local-roots'and r.get('event')=='started'for r in rows)
    return result


def suite(c):
    c['run']('coordinator-review-tests',['python3','-B','experiments/distributed-coordinator-stream/test_review.py'],30)
    manifest=json.loads((c['source']/'experiments/distributed-coordinator-stream/reference-inputs.json').read_text())
    target=c['cache']/'reference-inputs'
    total=copy_inputs(VECTORS['BASE']['ROOT'],manifest,target)
    c['resources'].update(reference_input_bytes=total,reference_strategy='retained native files,streamed support;no full coordinator reference/archive arrays')
    c['save']()
    c['run']('reference-preflight',['srun','--nodes=2','--ntasks=2','--ntasks-per-node=1','--cpus-per-task=2','--export=ALL,TMPDIR=/tmp','--chdir='+str(c['cache']),
        'bash','-c','set -eu; for p in "$1"/*; do test -r "$p"; done','reference-preflight',target],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for name,race in (('bounded-reference-cpu',[]),('bounded-reference-race',['-race'])):
        c['run'](name,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestRetainedPilotReference$|^TestOrderedSupportWriter$','-count=1','-timeout=20s'],60)
    return RESTART['suite'](c,reference_from=target,review_fn=review)


if __name__=='__main__':
    if '--check'in sys.argv:print('same2 HPC nodes/2H200/4CPU/8GiB RAM;128MiB owned/card,1GiB RSS/rank,128MiB cache/2GiB scratch/32MiB file;720s/15min;cold/fresh-resume retained-native references+streamed support;no CUDA change/repeated sanitizer;no fulls18/default/resource change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
