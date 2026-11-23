#!/usr/bin/env python3
"""Bounded diagnostic codec versus identical bulk files, using shared executors."""
from pathlib import Path
import json
import runpy
import sys
HERE=Path(__file__).resolve().parent
PROOF=runpy.run_path(str(HERE.parent/'distributed-proof-recording/run.py'))
PREFIX=PROOF['PREFIX'];SCALE=PROOF['SCALE'];VECTORS=PROOF['VECTORS']


def diagnostic_review(out,mode,directory,bounded=True):
    files=0;total_bytes=0
    for rank in range(4):
        last={}
        for line in (out/f'{mode}-{rank}.stderr.log').read_text().splitlines():
            try:row=json.loads(line)
            except json.JSONDecodeError:continue
            if row.get('work')!='diagnostic-write':continue
            assert bounded and row.get('event')!='failed'and row['mode']=='bounded'and row['unit']=='bytes'
            name=row['phase'];assert Path(name).name==name
            done,total=row['completed'],row['total'];assert 0<=done<=total
            if name in last:assert last[name][1]==total and last[name][0]<=done
            else:assert done==0
            last[name]=done,total
        if bounded:
            assert last,f"missing bounded diagnostic progress: {mode}, rank {rank}"
            for name,(done,total)in last.items():
                assert done==total==(directory/name).stat().st_size
                files+=1;total_bytes+=total
        else:assert not last
    return dict(files=files,bytes=total_bytes,serialization_buffer_limit_bytes=65536 if bounded else None)


def review(rows,directory,resume=False,source=None,cold=None):
    return PROOF['review'](rows,directory,resume,source,cold)


def comparison_arms():
    # The shared runner carries overrides forward: every codec transition must be explicit.
    return [dict(name='bulk',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_DIAGNOSTIC_WRITE='full'),review=lambda rows,directory:review(rows,directory)),dict(name='wide',test='TestWidePipelineTail',memcheck=False,profile=True,env=dict(PIR_PIPELINE_RESUME_FROM='',PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded'),review=PROOF['trace_wide_review'])]


def suite(c):
    c['run']('diagnostic-review-tests',['python3','-B','experiments/distributed-diagnostic-writes/test_review.py'],30)
    c['env'].update(PIR_PIPELINE_DIAGNOSTIC_WRITE='bounded',PIR_PIPELINE_PROOF_LOG='trace',LIFTRS_GPU_TAYLOR_PROGRESS='bounded',PIR_PIPELINE_NEWTON_INPUT='prefix',PIR_PIPELINE_DISCOVERY_RESULT='stream',PIR_PIPELINE_SPECTRUM='stream',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for label,race in (('diagnostic-cpu',[]),('diagnostic-race',['-race'])):
        c['run'](label,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestDiagnosticWrite|^TestProofLog','-count=1','-timeout=20s'],60)
    extra=comparison_arms()
    result=SCALE['suite'](c,review_fn=review,extra_arms=extra)
    arms=result['composed_vectors']['arms'];assert arms['cold']['artifacts']==arms['bulk']['artifacts']
    counts={};native={}
    for mode in ('cold','resume','bulk','wide'):
        counts[mode]=diagnostic_review(c['out'],mode,c['cache']/mode,mode!='bulk')
        rows=VECTORS['BASE']['parse_rank_logs'](c['out']/(mode+'.log'),ranks=4)
        native[mode]=PROOF['native_review'](c['out'],mode,rows)
    result.update(diagnostic_writes=counts,native_taylor_totals=native)
    c['resources'].update(diagnostic_writes='Explicit bounded 64KiB codec;identical diagnostic bytes;bulk default retained',scope='Genuine s8 cold/restart, matched bulk write control and bounded field18;not fulls18')
    return result


if __name__=='__main__':
    if '--check'in sys.argv:print('same2HPCnodes/2H200/4CPU/8GiBhost;4workers,64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;4 bounded/bulk arms;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
