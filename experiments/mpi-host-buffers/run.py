#!/usr/bin/env python3
"""Host-only replay of captured MPI payloads; no CUDA toolchain or GPU allocation."""
from pathlib import Path
import hashlib
import json
import runpy
import shutil
import struct
import sys
HERE=Path(__file__).resolve().parent
BASE=runpy.run_path(str(HERE.parent/'gpu-staged-transport/run.py'))
ARMS=(('staged-normal','staged',False),('staged-reclaim','staged',True),('direct-normal','direct',False),('direct-reclaim','direct',True))
PAYLOAD=Path('.cache/liftrs/mpi-buffer-comparison/compare-01/direct-a-normal/rank-1/case-2-sample-5-root-1-failed-send-0.bin')
PAYLOAD_SHA='4b73c559c2d99fa3e27ed1c60b584a4184265204d9bb55bcb3b9dcd1e7218d1d'


def review(rows,payload,mode,reclaim,transport=None):
    hosts=[r for r in rows if r.get('kind')=='host']
    assert len(hosts)==4 and {r['rank']for r in hosts}==set(range(4))
    if transport is not None:assert all(r['mpi_transport']==transport for r in hosts)
    placement={r['rank']:r['host']for r in hosts}
    assert placement[0]==placement[1]and placement[2]==placement[3]and placement[0]!=placement[2]
    expected={4:hashlib.sha256(payload).hexdigest(),8:hashlib.sha256(b''.join(struct.pack('<Q',(1<<34)|x)for x in struct.unpack('<16384I',payload[:65536]))).hexdigest()}
    assert expected[4]==PAYLOAD_SHA
    assert all(r['mode']==mode and r['reclaim']==reclaim and r['staging_bytes']==(524288 if mode=='staged'else 0)and r['chunk_words']==65536 and r['payload_sha256']==expected[4]for r in hosts)
    proofs=[r for r in rows if r.get('kind')=='proof']
    routes=((1,0,4),(2,0,4),(0,1,8),(0,2,8))
    assert len(proofs)==64*4*2
    for cycle in range(64):
        for case,(sender,receiver,width)in enumerate(routes):
            group=[r for r in proofs if r['cycle']==cycle and r['case']==case]
            assert len(group)==2
            assert {(r['rank'],r['peer'],r['direction'])for r in group}=={(sender,receiver,'send'),(receiver,sender,'receive')}
            assert all(r['matches']and r['word_bytes']==width and r['own_proof']==r['remote_proof']==expected[width]*2 and not r['file']for r in group)
    for rank in range(4):
        progress=[r for r in rows if r.get('kind')=='progress'and r['rank']==rank]
        assert [r['completed']for r in progress]==list(range(1,257))and all(r['total']==256 for r in progress)
        assert sum(r.get('kind')=='completed'and r['rank']==rank and r['handoffs']==256 for r in rows)==1
    assert all(r['peak_rss_bytes']<=1<<30 for r in rows if 'peak_rss_bytes'in r)
    return dict(mode=mode,reclaim=reclaim,proof_events=len(proofs),completed_handoffs=256,peak_rss_bytes_per_rank=max(r.get('peak_rss_bytes',0)for r in rows),hosts=placement)


def suite(context,transport=None):
    run=context['run'];cache=context['cache'];out=context['out'];env=context['env'];go=context['go'];source=context['source']
    if transport is not None:
        env.update(transport)
        context['resources']['diagnostic_transport']=transport
    context['resources'].update(ranks=4,ranks_per_node=2,cpus_per_rank=1,gpus=0,gpu_model=None,owned_cap_MiB_per_card=0,pinned_cap_MiB_per_rank=0,rss_cap_GiB_per_rank=1,churn_MiB_per_rank=16,max_staging_KiB_per_rank=512)
    context['save']()
    data=(BASE['ROOT']/PAYLOAD).read_bytes();assert len(data)==65537*4 and hashlib.sha256(data).hexdigest()==PAYLOAD_SHA
    payload=cache/'replay.bin';payload.write_bytes(data)
    (out/'fixture.json').write_text(json.dumps(dict(source=str(PAYLOAD),sha256=PAYLOAD_SHA,bytes=len(data),job='77868',scope='public synthetic preprocessing polynomial, not client data'),indent=2)+'\n')
    env.update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',GOMAXPROCS='1',GOMEMLIMIT=str(512<<20),PIR_HOST_BUFFER_PAYLOAD=str(payload))
    binary=cache/'host.test'
    run('build',[go,'test','-c','-tags','liftrs_mpi','-o',binary,'./experiments/mpi-host-buffers'],120)
    run('unit',[binary,'-test.run=^TestFingerprintHighBitsAndTail$','-test.v','-test.timeout=20s'])
    wrapper=cache/'rank.sh';shutil.copyfile(source/'experiments/distributed-root/rank.sh',wrapper)
    state=dict(status='running',arms=[])
    def save():(out/'host-progress.json').write_text(json.dumps(state,indent=2)+'\n')
    save()
    for name,mode,reclaim in ARMS:
        state['current_arm']=name;save();env['PIR_HOST_BUFFER_OUTPUT']=str(cache/name)
        try:
            run(name,['timeout','--signal=TERM','--kill-after=10s','100s','srun','--mpi=pmix_v3','--nodes=2','--ntasks=4','--ntasks-per-node=2','--cpus-per-task=1','--distribution=block:block','--cpu-bind=cores','--chdir='+str(cache),'--label','--kill-on-bad-exit=1','bash',wrapper,out/name,binary,'-test.run=^TestHostBuffers$','-test.v','-test.timeout=90s','-buffer-mode='+mode,'-reclaim='+str(reclaim).lower()],120)
            state['arms'].append(review(BASE['parse'](out/(name+'.log')),data,mode,reclaim,transport));save()
        except BaseException as error:state.update(status='failed',error=str(error));save();raise
    run('profile',[go,'tool','pprof','-symbolize=none','-top',binary,out/'staged-normal-0.pprof'],90)
    state['status']='passed';save();return dict(host_buffers=state)


if __name__=='__main__':
    if '--check'in sys.argv:print('2 HPC nodes / 4 ranks / 4 CPU / 8 GiB requested host / zero GPUs; captured payload; staged/direct and reclamation controls; unchanged 15 min scheduler, stop on failure')
    else:BASE['main'](here=HERE,extra_suite=suite,require_gpu=False)
