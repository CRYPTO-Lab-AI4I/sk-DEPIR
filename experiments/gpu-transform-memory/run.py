#!/usr/bin/env python3
"""Compare explicit GPU transform layouts for scratch memory and latency."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import runpy
import shutil
import signal
import subprocess
import sys
import time

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
SHARED=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
STOP=SHARED['stop_group']


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    parser.add_argument('--layout',choices=('fast','direct','compare'),default='fast',help='direct/compare explicitly opt into a potentially slower memory-saving experiment')
    args=parser.parse_args()
    name=args.name
    if not name or not name.replace('-','').isalnum():raise SystemExit('Use a fresh alphanumeric/hyphen run name')
    job=os.environ['SLURM_JOB_ID']
    out=HERE/'runs'/name;out.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-gpu-transform-memory-'+job);scratch.mkdir(exist_ok=False)
    cache=ROOT/'.cache/liftrs/gpu-transform-memory'/name;cache.mkdir(parents=True,exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',
             GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),
             CGO_ENABLED='1',CXX='g++',GOMAXPROCS='4',GOFLAGS='-p=4',PIR_CUDA_ONLINE_CACHE=str(cache/'code-fixtures'))
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'commands':[],
            'cache':str(cache),
            'policy':{'cpus':16,'rss_guard_gib':12,'deadline_seconds':1500,'device_cap_mib':512,'layout':args.layout,'dimensions':[8,9,16,17,20,21,24,25],'repeats':3,'cache_guard_bytes':2<<30,'scratch_guard_bytes':4<<30,'full_online':False}}
    cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    began=time.monotonic()

    def save():
        p=out/'results.json.tmp';p.write_text(json.dumps(record,indent=2)+'\n');p.replace(out/'results.json')

    def run(label,argv,cwd=ROOT):
        start=time.monotonic();peak=0;reason=None
        with (out/(label+'.log')).open('w') as log:
            p=subprocess.Popen(argv,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while p.poll() is None:
                    peak=max(peak,HELP['group_rss'](p.pid))
                    if peak>12*2**30: reason='RSS guard'
                    if time.monotonic()-began>1500: reason='experiment deadline'
                    if SHARED['tree_bytes'](cache)>2<<30:reason='cache guard'
                    if SHARED['tree_bytes'](scratch)>4<<30:reason='scratch guard'
                    if SHARED['mem_available']()<16<<30:reason='node headroom guard'
                    if cancelled: reason='user/Slurm termination'
                    if reason: STOP(p);break
                    time.sleep(1)
            finally:
                if p.poll() is None: STOP(p)
        record['commands'].append({'label':label,'argv':list(map(str,argv)),'seconds':time.monotonic()-start,
                                   'peak_rss_bytes':peak,'exit_code':p.returncode,'guard':reason});save()
        if reason or p.returncode: raise RuntimeError(label+': '+str(reason or p.returncode))

    save()
    try:
        if SHARED['mem_available']()<32<<30:raise RuntimeError('initial node memory guard')
        if min(shutil.disk_usage(cache).free,shutil.disk_usage(scratch).free)<6<<30:raise RuntimeError('free cache/scratch space guard')
        record['cpu_affinity']=sorted(os.sched_getaffinity(0))
        for label,argv in [('hostname',['hostname']),('cpu',['lscpu']),('compiler',['g++','--version']),('go',[go,'version']),
                           ('job',['scontrol','show','job',job]),('os',['cat','/etc/os-release']),('memory',['free','-b'])]:run(label,argv)
        source=scratch/'repo'
        shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):HELP['sha'](p) for p in source.rglob('*')
            if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum','Makefile'))}
        for path,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+path],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists(): dest.write_bytes((source/path).read_bytes())
        run('gpu',['nvidia-smi','--query-gpu=name,uuid,driver_version,memory.total,compute_cap','--format=csv'])
        run('nvcc',['nvcc','--version'])
        library=cache/'libliftrs-cuda.so';env['PIR_GPU_LIBRARY']=str(library)
        run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native','-shared','-Xcompiler=-fPIC','encoding/liftrs-cuda/matrix.cu','-o',str(library)],source)
        record['library']=str(library);record['library_sha256']=HELP['sha'](library)
        binary=cache/'transform.test'
        run('build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary);record['binary_sha256']=HELP['sha'](binary)
        run('cuda-build-checks',[go,'test','-tags','liftrs_cuda','./encoding/liftrs-shared','-run','TestCUDAEncoderLifecycle|TestCUDAAdmission|TestCUDAMatrixDescriptions','-count=1'],source)
        run('combined-checks',[str(binary),'-test.run=TestCUDAEncoderLifecycle|TestCUDAAdmission|TestCUDAMatrixDescriptions','-test.v'],source)
        old=json.loads((ROOT/'experiments/cuda-buffer-reuse/runs/reuse-01/results.json').read_text());assert old['status']=='passed'
        old_library=Path(old['cache'])/'libliftrs-cuda.so';assert HELP['sha'](old_library)==old['cache_sha256'][old_library.name]
        record['old_library']={'path':str(old_library),'sha256':HELP['sha'](old_library)}
        # Exercise old online modes without changing the running CPU experiment.
        for mode in ('default-online','reuse-online'):
            dest=out/mode;dest.mkdir();fixtures=cache/mode;fixtures.mkdir()
            env.update(PIR_CUDA_ONLINE_OUT=str(dest),PIR_CUDA_ONLINE_CACHE=str(fixtures),PIR_CUDA_REUSE='1' if mode=='reuse-online' else '0')
            run(mode,[str(binary),'-test.run=^TestCUDAOnlineEncoding$','-test.v','-test.timeout=5m'],source)
        env.pop('PIR_CUDA_ONLINE_OUT',None)
        for mode in ('memcheck','transforms'):
            dest=out/mode;dest.mkdir();fixtures=cache/mode;fixtures.mkdir()
            env.update(PIR_GPU_ROWS_OUT=str(dest),PIR_GPU_ROWS_CACHE=str(fixtures),PIR_GPU_ROWS_LAYOUT=args.layout,PIR_GPU_ROWS_OLD_LIBRARY=str(old_library))
            argv=[str(binary),'-test.run=^TestGPUDirectRows$','-test.v','-test.timeout=20m']
            if mode=='memcheck':
                env['PIR_GPU_ROWS_SMOKE']='1'
                argv=['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99']+argv
            else:env.pop('PIR_GPU_ROWS_SMOKE',None)
            run(mode,argv,source)
            record[mode]=verify_transforms(dest,args.layout,mode=='memcheck')
        assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
        for mode in ('memcheck','transforms'):run('pprof-'+mode,[go,'tool','pprof','-top','-nodecount=25',str(binary),str(out/mode/'cpu.pprof')])
        record['cache_sha256']={str(p.relative_to(cache)):HELP['sha'](p) for p in cache.rglob('*') if p.is_file()}
        record['status']='passed'
    except Exception as error:
        record['status']='stopped' if cancelled else 'failed';record['error']=repr(error)
        raise
    finally:
        record['evidence_sha256']={str(p.relative_to(out)):HELP['sha'](p) for p in out.rglob('*')
                                  if p.is_file() and p.name not in ('results.json','results.json.tmp','analysis.json')}
        save()


def verify_transforms(out,layout,small=False):
    rows=json.loads((out/'transforms.json').read_text())
    modes=('fast','direct') if layout=='compare' else (layout,)
    dims=(8,9) if small else (8,9,16,17,20,21,24,25)
    assert len(rows)==len(modes)*len(dims)
    by={(r['dimension'],r['layout']):r for r in rows};assert len(by)==len(rows)
    for dim in dims:
        for mode in modes:
            r=by[dim,mode];n=1<<dim;assert r['n']==n and r['matches'] and r['closed']
            assert r['owned_bytes']<=512<<20
            assert r['working_array_bytes']==n*4*(1 if mode=='direct' else 2)
            assert 0<r['owned_bytes']-r['table_bytes']-r['working_array_bytes']<1<<20
            for direction in ('forward','inverse'):
                assert len(r[direction+'_seconds'])==(1 if small else 3)
                assert len(r[direction+'_stages'])==(1 if small else 3)
                for elapsed,st in zip(r[direction+'_seconds'],r[direction+'_stages']):
                    assert len(st)==8 and 0<=st[7]<=elapsed+.001
                    if mode=='direct':assert st[4]==0
        if layout=='compare':
            a,b=by[dim,'fast'],by[dim,'direct']
            assert a['owned_bytes']-b['owned_bytes']==4*(1<<dim)
            assert a['input_sha256']==b['input_sha256'] and a['reference_sha256']==b['reference_sha256']
    return rows


if __name__=='__main__':main()
