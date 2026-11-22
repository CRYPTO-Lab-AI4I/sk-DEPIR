#!/usr/bin/env python3
"""Compact checkpoint cold/restart with the maintained owned-root and wide suite."""
from pathlib import Path
import json
import runpy
import sys
HERE=Path(__file__).resolve().parent
OWN=runpy.run_path(str(HERE.parent/'distributed-owned-roots/run.py'))
VECTORS=OWN['VECTORS']


def storage(directory):
    result={}
    for s in (4,6):
        label=f's{s}';n=1<<(2*s);payload=padded=empty=stages=0
        for path in sorted(directory.glob(f'{label}-*-commit.json')):
            record=json.loads(path.read_text());assert record['layout']=='prefix'and record['n']==n and record['ranks']==4
            stage=record['stage'];assert path.name==f'{label}-{stage}-commit.json'
            OWN['PILOT']['checkpoint_bytes'](directory,label,stage,record)
            sizes=[(directory/f'{label}-rank-{rank}-{stage}.bin').stat().st_size for rank in range(4)]
            assert sum(sizes)==4*record['prefix']
            payload+=sum(sizes);padded+=4*n;empty+=sum(size==0 for size in sizes);stages+=1
        assert stages>0 and payload<padded
        result[label]=dict(payload_bytes=payload,padded_equivalent_bytes=padded,saved_bytes=padded-payload,zero_byte_rank_files=empty,completed_stages=stages)
    return result


def suite(c):
    c['run']('compact-checkpoint-review-tests',['python3','-B','experiments/distributed-prefix-checkpoints/test_review.py'],30)
    c['env']['PIR_PIPELINE_CHECKPOINT_LAYOUT']='prefix'
    result=OWN['suite'](c)
    result['checkpoint_storage']={mode:storage(c['cache']/mode)for mode in ('cold','resume')}
    c['resources']['checkpoint_layout']='prefix;only verified zero padding omitted'
    c['resources']['scope']='Compact prefix owned-root genuine cold/restart plus existing bounded field18 regressions;not fulls18'
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBowned/worker=128MiB/card,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min;compact checkpoint cold/restart plus owned narrow/wide regression;legacy padded default retained;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
