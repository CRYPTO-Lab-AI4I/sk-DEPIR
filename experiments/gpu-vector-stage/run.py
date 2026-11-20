#!/usr/bin/env python3
"""Bounded persistent GPU vector stages; no distributed FFT or full grid."""
from pathlib import Path
import json
import runpy
import shutil
HERE=Path(__file__).resolve().parent
BASE=runpy.run_path(str(HERE.parent/'gpu-staged-transport/run.py'))


def review(rows,directory,wide_axis=None):
    values=[r for r in rows if r.get('kind')=='measurement']
    cases=24 if wide_axis is None else 21
    assert wide_axis in (None,0,1)
    assert len(values)==cases and len({r['case']for r in values})==cases
    expected=set()
    for s,dims in ((4,(4,5,8)),(17,(7,12,34))):
        for dim in dims:
            n=1<<dim;longest=1<<((dim+1)//2)
            for columns in (False,True):
                if dim==34 and wide_axis is not None and int(columns)!=wide_axis:continue
                length=1<<(dim-dim//2 if columns else dim//2);jobs=n//length;count=1 if dim==34 else min(3,jobs)
                for first in (0,jobs-count):
                    if dim==34 and wide_axis is not None and first==0:continue
                    steps=2 if dim==34 and wide_axis is not None else 4
                    label=f's{s}-d{dim}-'+('columns-'if columns else'rows-')+str(first);expected.add(label)
                    selected=[r for r in values if r['case']==label];assert len(selected)==1
                    r=selected[0]
                    assert (r['s'],r['dimension'],r['columns'],r['length'],r['count'],r['first'],r['word_bytes'],r['capacity_words'])==(s,dim,columns,length,count,first,4 if s==4 else 8,min(n,3*longest))
                    assert r['matches']is True and 0<r['owned_device_bytes']<=128<<20 and 0<r['peak_rss_bytes']<1<<30
                    assert all(r[k]>0 for k in ('setup_seconds','forward_kernel_wall_seconds','inverse_kernel_wall_seconds'))
                    for a,b in (('expected','forward'),('input','inverse')):
                        left=directory/(label+'-'+a+'.bin');right=directory/(label+'-'+b+'.bin')
                        assert left.stat().st_size==right.stat().st_size==length*count*r['word_bytes']
                        assert left.read_bytes()==right.read_bytes(),label
                    assert any(x.get('phase')==label and x.get('event')=='completed'and x['completed']==x['total']==steps for x in rows)
    assert {r['case']for r in values}==expected
    assert sum(r.get('kind')=='completed'and r['cases']==cases for r in rows)==1
    return values


def sanitizer_summary(stdout,stderr):
    """CUDA may put summaries on either stream; require one clean pair total."""
    lines=(stdout+'\n'+stderr).splitlines()
    errors=[line.split('ERROR SUMMARY:',1)[1].strip()for line in lines if 'ERROR SUMMARY:'in line]
    leaks=[line.split('LEAK SUMMARY:',1)[1].strip()for line in lines if 'LEAK SUMMARY:'in line]
    assert errors==['0 errors'],'missing, duplicate or nonzero CUDA error summary'
    assert leaks==['0 bytes leaked in 0 allocations'],'missing, duplicate or nonzero CUDA leak summary'
    return dict(errors=0,leaked_bytes=0,leaked_allocations=0)


def suite(c):
    run=c['run'];env=c['env'];cache=c['cache'];out=c['out'];go=c['go'];source=c['source']
    c['resources'].update(rss_cap_GiB_per_rank=1,pinned_cap_MiB_per_rank=0,vector_memcheck_split=True,scope='independent GPU stage validation on each node, no MPI payloads')
    c['save']()
    fixture=cache/'fields.test';fields=cache/'fields.bin';binary=cache/'check'
    run('fixture-build',[go,'test','-c','-o',fixture,'./encoding/liftrs-shared'],120)
    env['PIR_GPU_VECTOR_FIELDS']=str(fields)
    run('fixtures',[fixture,'-test.run=^TestGPUVectorStageFixtures$','-test.v','-test.timeout=60s','-test.cpuprofile='+str(out/'fixtures.pprof')],90)
    env.pop('PIR_GPU_VECTOR_FIELDS')
    run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-pg','-arch=sm_90','-Xcompiler=-pthread',
        'experiments/gpu-vector-stage/check.cu','-o',binary],120)
    run('review-tests',['python3','-B','experiments/gpu-vector-stage/test_review.py'],30)
    # Frozen source lives only on the controller node; rank entry must be shared.
    wrapper=cache/'rank.sh';shutil.copyfile(source/'experiments/gpu-vector-stage/rank.sh',wrapper)
    summaries={}
    for mode in ('normal','memcheck'):
        dest=cache/mode;dest.mkdir()
        limit=540 if mode=='memcheck'else 150
        launcher=['timeout','--signal=TERM','--kill-after=10s',str(limit)+'s','srun','--nodes=2','--ntasks=2','--ntasks-per-node=1','--cpus-per-task=2',
            '--gpus-per-task=h200:1','--gpu-bind=single:1','--cpu-bind=cores','--kill-on-bad-exit=1','--chdir='+str(cache),
            '--output='+str(out/(mode+'-%t.stdout.log')),'--error='+str(out/(mode+'-%t.stderr.log')),
            'bash',wrapper,dest]
        command=[binary,fields]
        if mode=='memcheck':command=['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',*command]
        env['LIFTRS_CPU_PROGRESS']='1'
        env['PIR_VECTOR_MEMCHECK']='1'if mode=='memcheck'else'0'
        run(mode,[*launcher,*command],limit+15)
        ranks=[]
        for rank in (0,1):
            rows=BASE['parse'](out/f'{mode}-{rank}.stdout.log')
            ranks.append(review(rows,dest/f'rank-{rank}',rank if mode=='memcheck'else None))
            if mode=='memcheck':
                sanitizer_summary((out/f'{mode}-{rank}.stdout.log').read_text(),
                    (out/f'{mode}-{rank}.stderr.log').read_text())
        summaries[mode]=ranks
        if mode=='normal':
            # Persist usable profile reports before a later sanitizer failure.
            profiles=sorted((cache/'normal').glob('profile-*.*'));assert len(profiles)==2
            for i,p in enumerate(profiles):run('profile-'+str(i),['gprof','-b',binary,p],30)

    return dict(vector_stages=summaries,vector_memcheck_split=True)


if __name__=='__main__':
    import sys
    if '--check'in sys.argv:print('2 HPC nodes / 2 H200 / 4 CPU / 8 GiB host;128 MiB owned/card,1 GiB RSS/process;128 MiB retained/2 GiB scratch;720s controller/15min scheduler; no MPI payloads or full-grid allocation; no submission')
    else:BASE['main'](here=HERE,extra_suite=suite)
