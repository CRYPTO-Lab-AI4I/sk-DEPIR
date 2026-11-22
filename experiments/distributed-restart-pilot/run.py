#!/usr/bin/env python3
"""Cold execution and fresh-process prefix resume with bounded publication."""
from pathlib import Path
import json
import runpy
import sys
HERE=Path(__file__).resolve().parent
DISCOVERY=runpy.run_path(str(HERE.parent/'distributed-discovery-pilot/run.py'))
PILOT=DISCOVERY['PILOT'];VECTORS=PILOT['VECTORS']


def review(rows,directory,resume=False,source=None,cold=None):
    if resume:
        assert not any(r.get('discovery')is True for r in rows)
        result=PILOT['review'](rows,directory,initial=8,streamed=True)
    else:result=DISCOVERY['review'](rows,directory,streamed=True)
    processes={}
    hosts={r['rank']:r['host']for r in rows if r.get('kind')=='host'}
    for r in result['measurements']:
        assert r['initial_prefix']==(8 if resume else 1)and r['publication']=='stream'and r['discovery_performed']is(not resume)and r['root_performed']is(not resume)
        assert r['pid']>0
        processes[str(r['rank'])]=dict(pid=r['pid'],host=hosts[r['rank']])
        s=r['s'];label=f's{s}'
        assert result['artifacts'][label]['code_sha256']==DISCOVERY['CANONICAL'][s]
        assert (directory/f'{label}-generator.bin').read_bytes()==(directory/f'{label}-rank-0-reference-g.bin').read_bytes()
    if resume:
        assert source is not None and cold is not None
        events=[r for r in rows if r.get('kind')=='resume'];assert len(events)==4
        for r in events:assert Path(r['source'])==source and r['prefix']==8 and r['verified']is True
        for rank,process in processes.items():assert process!=cold['processes'][rank]
        for s in (4,6):
            for rank in (0,1):
                for stage in ('root','newton-8'):
                    name=f's{s}-rank-{rank}-{stage}.bin';assert (directory/name).read_bytes()==(source/name).read_bytes()
            assert result['artifacts'][f's{s}']==dict(cold['artifacts'][f's{s}'],checkpoint_prefixes=[x for x in cold['artifacts'][f's{s}']['checkpoint_prefixes']if x>=8])
    result['processes']=processes;return result


def suite(c):
    c['run']('restart-review-tests',['python3','-B','experiments/distributed-restart-pilot/test_review.py'],30)
    c['env'].update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx')
    for name,race in (('restart-publication-cpu',[]),('restart-publication-race',['-race'])):
        c['run'](name,[c['go'],'test',*race,'./encoding/liftrs-shared','-run','^TestRankPrefix|^TestRankStream','-count=1','-timeout=20s'],60)
    source=c['cache']/'cold'
    def resumed(rows,directory):
        cold=json.loads((c['out']/'composition.json').read_text())['arms']['cold']
        return review(rows,directory,True,source,cold)
    arms=[dict(name='cold',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,
        env=dict(PIR_PIPELINE_SMALL='0',PIR_PIPELINE_DISCOVERY='distributed',PIR_PIPELINE_PUBLICATION='stream',PIR_PIPELINE_RESUME_FROM=''),review=review),
        dict(name='resume',test='TestDistributedPreprocessingPilot',memcheck=False,profile=True,
        env=dict(PIR_PIPELINE_SMALL='0',PIR_PIPELINE_DISCOVERY='distributed',PIR_PIPELINE_PUBLICATION='stream',PIR_PIPELINE_RESUME_FROM=str(source)),review=resumed)]
    # No CUDA arithmetic changed. The preceding integrated discovery allocation
    # already sanitized the maintained kernels; validate new host boundaries here.
    return VECTORS['suite'](c,full_fft=False,arms=arms,scope='genuine s4/s6 cold plus fresh MPI-process resume from prefix8,streamed canonical/generator artifacts;trusted manifests,same rank count;no node-failure or fulls18 claim')


if __name__=='__main__':
    if '--check'in sys.argv:print('same2 HPC nodes/2H200/4CPU/8GiB RAM;128MiB owned/card,1GiB RSS/rank,128MiB cache/2GiB scratch/32MiB file;720s/15min;2arms cold/fresh-resume;no CUDA-kernel change or repeated sanitizer;no admission/default change;no submission')
    else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
