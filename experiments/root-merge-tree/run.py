#!/usr/bin/env python3
"""Four logical ranks on two GPUs: owned support and a two-level root tree."""
from pathlib import Path
from collections import Counter
import hashlib
import json
import runpy
import shutil
import struct
import sys
HERE=Path(__file__).resolve().parent
BASE=runpy.run_path(str(HERE.parent/'gpu-staged-transport/run.py'))
FIELDS=('seconds','route_seconds','root_seconds','tree_seconds','verify_seconds')


def review(rows,directory,small=False,repeats=2):
    """Verify rank topology, exact source coverage and all reduction handoffs."""
    assert 2<=repeats<=12
    cases=1 if small else 3
    sample_sets=[{-1,2}if small else {-1,0,1,2}for _ in range(cases)]
    if not small:sample_sets[2]={-1,0,1,2,*range(3,repeats+1)}
    hosts=[r for r in rows if r.get('kind')=='host']
    assert len(hosts)==4 and {r['rank']for r in hosts}==set(range(4))
    assert all(r['chunk_words']==(1024 if small else 65536)for r in hosts)
    nodes={r['host']for r in hosts};assert len(nodes)==2
    placement={r['rank']:r['host']for r in hosts}
    assert placement[0]==placement[1] and placement[2]==placement[3] and placement[0]!=placement[2]
    for host in nodes:
        group=[r for r in hosts if r['host']==host]
        assert len(group)==2 and {r['local_rank']for r in group}=={'0','1'}
        assert len({r['visible_devices']for r in group})==1 and group[0]['visible_devices'] and ','not in group[0]['visible_devices']
    values=[r for r in rows if r.get('kind')=='measurement']
    assert len(values)==4*sum(map(len,sample_sets))
    assert sum(r.get('kind')=='transfer'for r in rows)==6*sum(map(len,sample_sets))
    assert sum(r.get('kind')=='tree'for r in rows)==12*sum(map(len,sample_sets))
    def read(rank,name,width):
        data=(directory/f'rank-{rank}'/name).read_bytes();assert len(data)%width==0
        return data,struct.unpack('<'+('I'if width==4 else'Q')*(len(data)//width),data)
    def digest(data):return hashlib.sha256(data).hexdigest()
    schedule={0:[(1,1,False),(2,2,False)],1:[(1,0,True)],2:[(1,3,False),(2,0,True)],3:[(1,2,True)]}
    for rank in range(4):
        assert sum(r.get('kind')=='completed' and r['rank']==rank and r['cases']==cases for r in rows)==1
    expected_chunks=Counter()
    chunk=1024 if small else 65536
    for case in range(cases):
        samples=sample_sets[case]
        fixtures=[r for r in rows if r.get('kind')=='fixture' and r['case']==case]
        assert len(fixtures)==4 and {r['rank']for r in fixtures}==set(range(4))
        count=fixtures[0]['support_count'];s=8 if case==0 else 17;q=1<<s;n=1<<(2*s)
        assert all(r['support_count']==count and r['s']==s for r in fixtures)
        if case>0:assert count==(1<<16 if case==1 else 1<<18)
        reference_data,reference=read(0,f'case-{case}-reference.bin',4)
        ref_digest=digest(reference_data)
        assert len(reference)==count+1 and reference[0]==1 and all(v<q for v in reference)
        shards=[]
        for rank in range(4):
            fixture=next(r for r in fixtures if r['rank']==rank)
            assert fixture['coordinator_support_count']==(count if rank==0 else 0)
            assert fixture['reference_count']==(count+1 if rank==0 else 0)
            subtree=(4,1,2,1)[rank];bound=min(count+1,(count//4+2)*subtree+1);maximum=4
            while maximum<bound:maximum*=2
            assert fixture['subtree_ranks']==subtree and fixture['max_elements']==maximum
            assert fixture['reference_sha256']==(ref_digest if rank==0 else digest(b''))
            _,support=read(rank,f'case-{case}-support.bin',8)
            assert list(support)==sorted(set(support)) and all(v<n-1 for v in support)
            own=set(support);assert all(x*q%(n-1)in own for x in support)
            if s==17 and support:assert max(support)>1<<32
            shards.append(own)
            _,local=read(rank,f'case-{case}-local.bin',4)
            assert len(local)==len(support)+1 and local[0]==1 and all(v<q for v in local)
            if rank:assert not (directory/f'rank-{rank}'/f'case-{case}-reference.bin').exists()
            for sample in samples:
                selected=[r for r in values if r['case']==case and r['rank']==rank and r['sample']==sample]
                assert len(selected)==1;row=selected[0]
                assert row['matches'] and row['shard_count']==len(support) and row['support_count']==count and row['ranks']==4
                assert row['owned_result_words']==(count+1 if rank==0 else 0) and row['validation_io']==(sample==2)
                assert 0<row['owned_device_bytes']<=64<<20 and 0<row['peak_rss_bytes']<=1<<30
                assert all(row[f]>0 for f in FIELDS) and row['seconds']>=max(row[f]for f in FIELDS if f!='seconds')
                peer=[r for r in values if r['case']==case and r['rank']==0 and r['sample']==sample]
                assert len(peer)==1 and all(row[f]==peer[0][f]for f in FIELDS)
                steps=[r for r in rows if r.get('kind')=='tree' and r['case']==case and r['sample']==sample and r['rank']==rank]
                assert [(r['stride'],r['peer'],r['sending'],r['event'])for r in steps]==[(stride,peer,sending,event)for stride,peer,sending in schedule[rank]for event in ('started','completed')]
        assert sum(map(len,shards))==count and len(set.union(*shards))==count
        assert max(map(len,shards))-min(map(len,shards))<=2
        if chunk<16384:
            # The small case fits one source packet per owner; normal packets
            # are at most16384 words, below its65536-word chunk threshold.
            assert max(map(len,shards))<=16384
            for peer in (1,2,3):
                if len(shards[peer])>chunk:
                    expected_chunks[('support-transfer',0,peer,True,len(shards[peer]))]+=len(samples)
                    expected_chunks[('support-transfer',peer,0,False,len(shards[peer]))]+=len(samples)

        for sample in samples:
            for sender,receiver,stride in ((1,0,1),(3,2,1),(2,0,2)):
                transfers=[r for r in rows if r.get('kind')=='transfer' and r['case']==case and r['sample']==sample and r['stride']==stride and
                    ((r['rank']==sender and r['peer']==receiver and r['direction']=='send')or(r['rank']==receiver and r['peer']==sender and r['direction']=='receive'))]
                assert len(transfers)==2 and {r['rank']for r in transfers}=={sender,receiver}
                assert len({r['sha256']for r in transfers})==len({r['words']for r in transfers})==1
                subtree=shards[sender]if stride==1 else shards[2]|shards[3]
                assert transfers[0]['words']==len(subtree)+1
                if sample==2:
                    arrays=[]
                    for r in transfers:
                        data,p=read(r['rank'],r['file'],4);assert len(p)==r['words'] and p[0]==1 and all(v<q for v in p)
                        assert digest(data)==r['sha256'];arrays.append(data)
                    assert arrays[0]==arrays[1]
        # Every retained state must continue the preceding owned state, not
        # merely have the right final digest or the right transfer dimensions.
        for rank in range(4):
            current,_=read(rank,f'case-{case}-local.bin',4)
            for stride,peer,sending in schedule[rank]:
                trace=[r for r in rows if r.get('kind')=='tree' and r['rank']==rank and r['case']==case and r['sample']==2 and r['stride']==stride]
                assert trace[0]['sha256']==digest(current)
                if sending:
                    sent,_=read(rank,f'case-{case}-stride-{stride}-send-{peer}.bin',4);assert sent==current
                else:
                    merged,coefficients=read(rank,f'case-{case}-merged-{stride}.bin',4)
                    remote,_=read(rank,f'case-{case}-stride-{stride}-receive-{peer}.bin',4)
                    assert len(merged)==len(current)+len(remote)-4 and coefficients[0]==1 and all(v<q for v in coefficients)
                    current=merged
                assert trace[1]['sha256']==digest(current)
        # Reconstruct the ownership chain from retained bytes: the first sends
        # are leaves; the second-level send is the owner's previous merged state.
        for sender,receiver in ((1,0),(3,2)):
            leaf,_=read(sender,f'case-{case}-local.bin',4)
            sent,_=read(sender,f'case-{case}-stride-1-send-{receiver}.bin',4);assert leaf==sent
        middle,_=read(2,f'case-{case}-merged-1.bin',4)
        sent,_=read(2,f'case-{case}-stride-2-send-0.bin',4);assert middle==sent
        final,_=read(0,f'case-{case}-final.bin',4);merged,_=read(0,f'case-{case}-merged-2.bin',4)
        assert final==merged==reference_data
    for r in rows:
        if r.get('kind')=='transfer'and r['words']>chunk:
            expected_chunks[('root-transfer',r['rank'],r['peer'],r['direction']=='send',r['words'])]+=1
    completed=[r for r in rows if r.get('kind')=='mpi-chunk'and r.get('event')=='completed']
    actual_chunks=Counter()
    for r in completed:
        assert r['total']>chunk and r['completed']==r['total'] and r['unit']=='words'
        assert r['rank']in range(4) and r['peer']in range(4) and r['peer']!=r['rank']
        actual_chunks[(r['phase'],r['rank'],r['peer'],r['sending'],r['total'])]+=1
    assert actual_chunks==expected_chunks
    return values


def suite(context):
    run=context['run'];cache=context['cache'];out=context['out'];source=context['source'];env=context['env'];go=context['go']
    context['resources'].update(ranks=4,ranks_per_node=2,cpus_per_rank=1,owned_cap_MiB_per_rank=64,rss_cap_GiB_per_rank=1,pinned_cap_MiB_per_rank=0,
        physical_gpus=2,logical_ranks_share_card=True)
    context['save']()
    env.update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',LIFTRS_CPU_PROGRESS='1')
    library=cache/'libliftrs-cuda.so'
    run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=sm_90','-shared','-Xcompiler=-fPIC,-pthread',
        'encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/preparation.cu','-o',library],120)
    binary=cache/'tree.test'
    run('tree-build',[go,'test','-c','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','-o',binary,'./encoding/liftrs-shared'],120)
    run('fixture-regression',[binary,'-test.run=TestRootFixtureReferenceDigests|TestRootTransferProofAgreement|TestIndexTransferFingerprintHighBits','-test.v','-test.timeout=60s'],90)
    tests='TestRootTree|TestRootRouting|TestPreparationRootPartitions|TestRootPartitionAdmission'
    run('shared-tests',[go,'test','./encoding/liftrs-shared','-run',tests,'-count=1','-timeout=60s'],90)
    run('shared-race',[go,'test','-race','./encoding/liftrs-shared','-run',tests,'-count=1','-timeout=60s'],90)
    wrapper=cache/'rank.sh';shutil.copyfile(source/'experiments/distributed-root/rank.sh',wrapper)
    env.update(PIR_GPU_PREPARATION_LIBRARY=str(library),LIFTRS_GPU_PREPARATION_VECTOR_LAYOUT='local',PIR_ROOT_TREE_CHUNK_WORDS='65536',GOMAXPROCS='1',GOMEMLIMIT=str(512<<20))
    launcher=['timeout','--signal=TERM','--kill-after=10s','240s','srun','--mpi=pmix_v3','--nodes=2','--ntasks=4','--ntasks-per-node=2',
        '--cpus-per-task=1','--gpus-per-node=h200:1','--gpu-bind=single:2','--gres-flags=allow-task-sharing',
        '--distribution=block:block','--cpu-bind=cores','--label','--kill-on-bad-exit=1','bash',wrapper]
    test_args=['-test.run=^TestRankOwnedRootTree$','-test.v','-test.timeout=220s','-root-mpi-buffer='+env.get('PIR_ROOT_TREE_MPI_BUFFER','direct')]
    env['PIR_ROOT_TREE_OUTPUT']=str(cache/'normal')
    run('normal',[*launcher,out/'normal',binary,*test_args],260)
    normal=BASE['parse'](out/'normal.log');review(normal,cache/'normal',repeats=int(env.get('PIR_ROOT_TREE_REPEATS','2')))
    env.update(PIR_ROOT_TREE_SMALL='1',PIR_ROOT_TREE_CHUNK_WORDS='1024',PIR_ROOT_TREE_OUTPUT=str(cache/'sanitized'))
    run('memcheck',[*launcher,out/'memcheck','compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',binary,*test_args],260)
    sanitized=BASE['parse'](out/'memcheck.log');review(sanitized,cache/'sanitized',small=True)
    text=(out/'memcheck.log').read_text()
    assert text.count('ERROR SUMMARY: 0 errors')==4 and text.count('LEAK SUMMARY: 0 bytes')==4
    for rank in range(4):
        assert (out/f'normal-{rank}.pprof').stat().st_size>0
    # Keep post-run summaries offline; retain full profiles for later native symbolization.
    run('profile',[go,'tool','pprof','-symbolize=none','-top',binary,out/'normal-0.pprof'])
    return dict(measurements=normal,sanitized=sanitized)


if __name__=='__main__':
    if '--check'in sys.argv:
        print('2HPCnodes/2H200/4logical ranks/4CPU/8GiB host;2ranks/card,64MiB owned/rank≤128MiB/card,1GiB RSS/rank;15min')
    else:BASE['main'](here=HERE,extra_suite=suite)
