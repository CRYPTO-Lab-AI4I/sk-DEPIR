#!/usr/bin/env python3
"""Two-node GPU normalized-root/complete-polynomial exchange and merge gate."""
from pathlib import Path
import hashlib
import json
import runpy
import shutil
import struct

HERE=Path(__file__).resolve().parent
TRANSPORT=runpy.run_path(str(HERE.parent/'gpu-staged-transport/run.py'))
FIELDS=('seconds','partition_seconds','root_seconds','exchange_seconds','merge_seconds','verify_seconds')


def review(rows,directory,small=False):
    """Reconcile both ranks, every pass and retained complete coefficient arrays."""
    cases=1 if small else 3
    hosts=[r for r in rows if r.get('kind')=='host']
    assert len(hosts)==2 and {r['rank']for r in hosts}=={0,1}
    assert len({r['host']for r in hosts})==2
    assert all(r['visible_devices'] and ','not in r['visible_devices']for r in hosts)
    values=[r for r in rows if r.get('kind')=='measurement']
    assert len(values)==cases*2*3
    def words(path):
        blob=path.read_bytes();assert len(blob)%4==0
        return blob,struct.unpack('<'+'I'*(len(blob)//4),blob)
    for rank in (0,1):
        assert sum(r.get('kind')=='completed' and r['rank']==rank and r['cases']==cases for r in rows)==1
        dest=directory/('rank-'+str(rank))
        fixtures=[r for r in rows if r.get('kind')=='fixture' and r['rank']==rank]
        assert len(fixtures)==cases and {r['case']for r in fixtures}==set(range(cases))
        for case in range(cases):
            fixture=next(r for r in fixtures if r['case']==case)
            assert fixture['s']==(8 if case==0 else 17) and fixture['genuine_support']==(case==0)
            data,reference=words(dest/f'case-{case}-reference.bin')
            digest=hashlib.sha256(data).hexdigest()
            assert fixture['sha256']==digest and len(reference)==fixture['support_count']+1 and reference[0]==1
            measurements=[r for r in values if r['rank']==rank and r['case']==case]
            assert len(measurements)==3 and {r['sample']for r in measurements}=={-1,0,1}
            for row in measurements:
                assert row['matches'] and row['sha256']==digest and row['s']==fixture['s'] and row['support_count']==fixture['support_count']
                assert row['local_count']+row['peer_count']==fixture['support_count']
                assert abs(row['local_count']-row['peer_count'])<=2
                assert row['remote_bytes']==4*(row['local_count']+1)
                assert 0<row['owned_device_bytes']<=128<<20 and 0<row['peak_rss_bytes']<=2<<30
                assert all(row[f]>0 for f in FIELDS) and row['seconds']>=max(row[f]for f in FIELDS if f!='seconds')
                pair=[r for r in values if r['rank']==1-rank and r['case']==case and r['sample']==row['sample']]
                assert len(pair)==1 and all(row[f]==pair[0][f]for f in FIELDS)
                assert row['local_count']==pair[0]['peer_count'] and row['sha256']==pair[0]['sha256']
            local_blob,local=words(dest/f'case-{case}-local.bin')
            peer_blob,peer=words(dest/f'case-{case}-peer.bin')
            assert len(local)==measurements[0]['local_count']+1 and len(peer)==measurements[0]['peer_count']+1
            assert local[0]==peer[0]==1
            assert all(v<1<<fixture['s']for v in (*local,*peer,*reference))
            assert peer_blob==(directory/f'rank-{1-rank}'/f'case-{case}-local.bin').read_bytes()
            assert (dest/f'case-{case}-merged.bin').read_bytes()==data
    return values


def suite(context):
    run=context['run'];source=context['source'];cache=context['cache'];out=context['out'];go=context['go'];env=context['env']
    # All programs are built from the runner's sealed tracked source snapshot.
    env.update(CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',LIFTRS_CPU_PROGRESS='1')
    library=cache/'libliftrs-cuda.so'
    run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=sm_90','-shared','-Xcompiler=-fPIC,-pthread',
        'encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/preparation.cu','-o',library],120)
    binary=cache/'root.test'
    run('root-build',[go,'test','-c','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','-o',binary,'./encoding/liftrs-shared'],120)
    run('fixture-regression',[binary,'-test.run=^TestRootFixtureReferenceDigests$','-test.v','-test.timeout=60s'],90)
    run('partition-tests',[go,'test','./encoding/liftrs-shared','-run','TestPreparationRootPartitions|TestRootPartitionAdmission',
        '-count=1','-timeout=60s'],90)
    run('partition-race',[go,'test','-race','./encoding/liftrs-shared','-run','TestPreparationRootPartitions|TestRootPartitionAdmission',
        '-count=1','-timeout=60s'],90)
    wrapper=cache/'rank.sh';shutil.copyfile(source/'experiments/distributed-root/rank.sh',wrapper)
    env['PIR_GPU_PREPARATION_LIBRARY']=str(library)
    launcher=['timeout','--signal=TERM','--kill-after=10s','240s','srun','--mpi=pmix_v3','--nodes=2','--ntasks=2','--ntasks-per-node=1',
        '--cpus-per-task=2','--gpus-per-task=h200:1','--gpu-bind=single:1','--cpu-bind=cores','--label','--kill-on-bad-exit=1','bash',wrapper]
    test_args=['-test.run=^TestDistributedRootProducts$','-test.v','-test.timeout=220s']
    env['PIR_DISTRIBUTED_ROOT_OUTPUT']=str(cache/'normal')
    run('normal',[*launcher,out/'normal',binary,*test_args],260)
    measured=TRANSPORT['parse'](out/'normal.log');review(measured,cache/'normal')
    env.update(PIR_DISTRIBUTED_ROOT_SMALL='1',PIR_DISTRIBUTED_ROOT_OUTPUT=str(cache/'sanitized'))
    run('memcheck',[*launcher,out/'memcheck','compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',binary,*test_args],260)
    sanitized=TRANSPORT['parse'](out/'memcheck.log');review(sanitized,cache/'sanitized',small=True)
    text=(out/'memcheck.log').read_text()
    assert text.count('ERROR SUMMARY: 0 errors')==2 and text.count('LEAK SUMMARY: 0 bytes')==2
    for rank in (0,1):
        profile=out/f'normal-{rank}.pprof';assert profile.stat().st_size>0
        run('profile-'+str(rank),[go,'tool','pprof','-top',binary,profile])
    return dict(measurements=measured,sanitized=sanitized)


if __name__=='__main__':TRANSPORT['main'](here=HERE,extra_suite=suite)
