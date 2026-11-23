#!/usr/bin/env python3
"""Bounded GPU row/column composition through the existing rank transpose."""
from pathlib import Path
import hashlib
import json
import runpy
import shutil
import sys
HERE=Path(__file__).resolve().parent
CONTROL=runpy.run_path(str(HERE.parent/'ucx-registration-cache/run.py'))
BASE=CONTROL['HOST']['BASE']
STAGE=runpy.run_path(str(HERE.parent/'gpu-vector-stage/run.py'))
CASES=((4,4),(4,5),(4,8),(17,5),(17,9),(17,12))


def transpose_key(rank,label,pass_id,operation,offset,direction,peer,ranks):
    base=(rank,label,pass_id,operation,offset,direction)
    return base if ranks==2 else(*base,peer)


def transpose_packets(rank,label,pass_id,operation,n,tile,ranks):
    assert ranks in (2,4)and n%(ranks*ranks)==0
    words=n//(ranks*ranks)
    for peer in range(ranks):
        if peer==rank:continue
        for offset in range(0,words,tile):
            for direction in ('send','receive'):
                yield transpose_key(rank,label,pass_id,operation,offset,direction,peer,ranks),min(tile,words-offset)


def review_transpose_routes(rows,expected_packets,ranks=2):
    proofs=[r for r in rows if r.get('kind')=='tile-proof'];actual_packets={};packets={}
    for r in proofs:
        rank,peer=r['rank'],r['peer'];assert 0<=rank<ranks and 0<=peer<ranks and rank!=peer
        direction=r['direction'];assert direction in ('send','receive')
        assert r['round']==((peer-rank)if direction=='send'else(rank-peer))%ranks
        key=transpose_key(rank,r['case'],r['pass'],r['operation'],r['offset'],direction,peer,ranks);assert key not in actual_packets
        own,remote=bytes.fromhex(r['own_proof']),bytes.fromhex(r['remote_proof'])
        assert len(own)==len(remote)==64 and own[:32]==own[32:]==remote[:32]==remote[32:]
        actual_packets[key]=r['words'];packets[key]=r
    assert actual_packets==expected_packets
    for r in packets.values():
        key=transpose_key(r['peer'],r['case'],r['pass'],r['operation'],r['offset'],'receive'if r['direction']=='send'else'send',r['rank'],ranks)
        peer=packets[key]
        assert r['own_proof']==peer['remote_proof']and r['remote_proof']==peer['own_proof']
    return len(proofs)


def review(rows,directory,operations=None,full_fft=False,cases=None,ranks=2):
    cases=CASES if cases is None else cases
    hosts=[r for r in rows if r.get('kind')=='host']
    assert len(hosts)==ranks and {r['rank']for r in hosts}==set(range(ranks))and len({r['host']for r in hosts})==2
    assert all(r['visible_devices']and ','not in r['visible_devices']for r in hosts)
    values=[r for r in rows if r.get('kind')=='measurement'];assert len(values)==2*ranks*len(cases)
    proofs=[r for r in rows if r.get('kind')=='tile-proof'];expected_packets={};actual_packets={};packets={}
    for index,(s,dim)in enumerate(cases):
        label=f's{s}-d{dim}';nr,nc=1<<(dim-dim//2),1<<(dim//2);width=4 if s==4 else 8
        for pass_id in (0,1):
            selected=[r for r in values if r['case']==label and r['pass']==pass_id]
            assert len(selected)==ranks and {r['rank']for r in selected}==set(range(ranks))
            assert all(r['seconds']==selected[0]['seconds']>0 for r in selected)
            for r in selected:
                rank=r['rank'];assert r['s']==s and r['dimension']==dim and r['rank_first']==rank*(nr//ranks)and r['rank_rows']==nr//ranks
                assert r.get('full_fft',False)==full_fft
                assert r['word_bytes']==width and r['words']==nr*nc//ranks and r['matches']is True
                assert 0<r['owned_device_bytes']<=128<<20 and 0<r['peak_rss_bytes']<1<<30
                prefix=f'rank-{rank}-{label}-pass-{pass_id}';a=(directory/(prefix+'-output.bin')).read_bytes();b=(directory/(prefix+'-expected.bin')).read_bytes()
                assert a==b and len(a)==r['words']*width and hashlib.sha256(a).hexdigest()==r['output_sha256']
                if pass_id==1:
                    expected=b''.join(((((i+1)*0x9e3779b9)^(1<<(2*s-1)))&((1<<(2*s))-1)).to_bytes(width,'little')for i in range(rank*nr*nc//ranks,(rank+1)*nr*nc//ranks))
                    assert a==expected
                count=0
                for op in ((operations[(index,pass_id)],operations[(index,pass_id)]+1)if operations is not None else(4*index+2*pass_id+1,4*index+2*pass_id+2)):
                    for key,words in transpose_packets(rank,label,pass_id,op,nr*nc,127,ranks):
                        expected_packets[key]=words;count+=1
                assert r['proof_events']==count
    proof_count=review_transpose_routes(rows,expected_packets,ranks)
    assert all(sum(r.get('kind')=='completed'and r['rank']==rank and r['cases']==2*len(cases) for r in rows)==1 for rank in range(ranks))
    assert not list(directory.glob('capture-*'))
    return dict(measurements=values,proof_events=proof_count)


def suite(c,full_fft=False,review_fn=review,test_name="TestDistributedGPUVectors",scope=None,baseline=False,reference_review=None,reference_test="TestDistributedGPUVectors",arms=None,ranks=2,cpu_cores=4):
    assert ranks in (2,4,8)and(ranks<=4 or arms is not None)
    assert cpu_cores>=ranks and cpu_cores%ranks==0
    per_node=ranks//2;cpus=cpu_cores//ranks
    gpu_binding=['--gpus-per-task=h200:1','--gpu-bind=single:1']if ranks==2 else['--gpus-per-node=h200:1','--gpu-bind=single:'+str(per_node),'--gres-flags=allow-task-sharing']
    CONTROL['configure'](c)
    run=c['run'];go=c['go'];source=c['source'];cache=c['cache'];out=c['out'];env=c['env']
    c['resources'].update(ranks=ranks,cpus_per_node=cpu_cores//2,owned_cap_MiB_per_worker=128//per_node,rss_cap_GiB_per_rank=1,pinned_cap_MiB_per_rank=0,mpi_mode='staged',tile_words=127 if arms is None else 'per-arm',scope=scope or ('bounded distributed polynomial product'if test_name=='TestDistributedGPUProducts'else('bounded additive FFT with distributed Taylor'if full_fft else'separable GPU vectors with two rank transposes; Taylor absent')))
    c['save']();env.update(PIR_GPU_MATRIX_STAGE_TAYLOR='1'if full_fft else'0',CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',GOMAXPROCS='1',PIR_RANK_LOG_FILES='1',LIFTRS_CPU_PROGRESS='1')
    run('review-tests',['python3','-B','experiments/distributed-gpu-vectors/test_review.py'],30)
    run('cpu-orchestration',[go,'test','./encoding/liftrs-shared','-run','^TestRankTranspose|^TestRankMatrixVectors$|^TestRankVectorAdmission$|^TestRankTaylor|^TestRankAdditive|^TestRankProduct','-count=1','-timeout=45s'],90)
    run('cpu-race',[go,'test','-race','./encoding/liftrs-shared','-run','^TestRankTranspose|^TestRankMatrixVectors$|^TestRankVectorAdmission$|^TestRankTaylor|^TestRankAdditive|^TestRankProduct','-count=1','-timeout=45s'],90)
    library=cache/'libliftrs-cuda.so';binary=cache/'vectors.test'
    run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=sm_90','-shared','-Xcompiler=-fPIC,-pthread','encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/preparation.cu','-o',library],120)
    run('test-build',[go,'test','-c','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','-o',binary,'./encoding/liftrs-shared'],120)
    if full_fft:
        run('tiled-adapter-build',['nvcc','-std=c++17','-O2','-arch=sm_90','-Xcompiler=-pthread','experiments/preprocessing-tiles/check.cu','-o',cache/'host-tiled-reference-check'],120)
    wrapper=cache/'rank.sh';shutil.copyfile(source/'experiments/distributed-root/rank.sh',wrapper)
    env['PIR_GPU_PREPARATION_LIBRARY']=str(library)
    # Test every participating node before GPU work, using shared paths/cwd.
    run('rank-preflight',['srun','--nodes=2','--ntasks='+str(ranks),'--ntasks-per-node='+str(per_node),'--cpus-per-task='+str(cpus),'--export=ALL,TMPDIR=/tmp','--chdir='+str(cache),
        'bash','-c','set -eu; test -x "$1"; test -r "$2"; test -r "$3"; test -w "$4"',
        'rank-preflight',binary,wrapper,library,cache],30)
    state=dict(status='running',arms={})
    def save():(out/'composition.json').write_text(json.dumps(state,indent=2)+'\n')
    save()
    if arms is None:
        modes=('full-reference','normal','memcheck')if reference_review is not None else(('separable','normal','memcheck')if full_fft or baseline else('normal','memcheck'))
        arms=[dict(name=mode)for mode in modes]
    for arm_spec in arms:
        mode=arm_spec['name'];is_memcheck=arm_spec.get('memcheck',mode=='memcheck')
        env.update(arm_spec.get('env',{}))
        dest=cache/mode;dest.mkdir();env['PIR_GPU_MATRIX_STAGE_OUTPUT']=str(dest)
        env['PIR_GPU_MATRIX_STAGE_TAYLOR']='1'if full_fft and mode!='separable'else'0'
        rank_wrapper=arm_spec.get('wrapper',wrapper)
        if 'command'in arm_spec:
            run(mode+'-entry-preflight',['srun','--nodes=2','--ntasks='+str(ranks),'--ntasks-per-node='+str(per_node),'--cpus-per-task='+str(cpus),'--chdir='+str(cache),'bash','-c','set -eu; test -x "$1"; test -r "$2"','entry-preflight',arm_spec['command'][0],rank_wrapper],30)
        launcher=['timeout','--signal=TERM','--kill-after=10s','240s','srun','--mpi=pmix_v3','--nodes=2','--ntasks='+str(ranks),'--ntasks-per-node='+str(per_node),'--cpus-per-task='+str(cpus),
            '--export=ALL,TMPDIR=/tmp',*gpu_binding,'--cpu-bind=cores','--distribution=block:block','--chdir='+str(cache),'--kill-on-bad-exit=1','bash',rank_wrapper,out/mode]
        selected_test=arm_spec.get('test',reference_test if mode=='full-reference'else('TestDistributedGPUVectors'if mode=='separable'else test_name))
        command=arm_spec.get('command',[binary,'-test.run=^'+selected_test+'$','-test.v','-test.timeout=220s'])
        profile_binary=command[0]
        if is_memcheck:command=['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',*command]
        run(mode,[*launcher,*command],260)
        checker=arm_spec.get('review',reference_review if mode=='full-reference'else(review if mode=='separable'else review_fn))
        arm=checker(BASE['parse_rank_logs'](out/(mode+'.log'),ranks=ranks),dest)
        if is_memcheck:
            arm['sanitizer']=[STAGE['sanitizer_summary']((out/f'{mode}-{r}.stdout.log').read_text(),(out/f'{mode}-{r}.stderr.log').read_text())for r in range(ranks)]
        state['arms'][mode]=arm;save()
        if arm_spec.get('profile',mode=='normal'):
            for rank in range(ranks):run(('profile-'if mode=='normal'else mode+'-profile-')+str(rank),[go,'tool','pprof','-symbolize=none','-top',profile_binary,out/f'{mode}-{rank}.pprof'],45)
    state['status']='passed';save();return dict(composed_vectors=state)


if __name__=='__main__':
    if '--check'in sys.argv:print('2 HPC nodes/2 H200/4 CPU/8 GiB host;128 MiB owned/card/1 GiB RSS/rank;128 MiB cache/2 GiB scratch/720s controller/15min scheduler; shared-path preflight, staged UCX; full small sanitizer suite; no Taylor/full FFT claim; no submission')
    else:BASE['main'](here=HERE,extra_suite=suite)
