#!/usr/bin/env python3
"""Validate modular primary CUDA/CPU online encoding before full s16 admission."""
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
STOP=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))['stop_group']


def main():
    if len(sys.argv)!=2 or not sys.argv[1] or not sys.argv[1].replace('-','').isalnum():raise SystemExit('Usage: run.py FRESH_RUN_NAME')
    job=os.environ['SLURM_JOB_ID'];name=sys.argv[1]
    out=HERE/'runs'/name;out.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-cuda-encoding-'+job);scratch.mkdir(exist_ok=False)
    cache=ROOT/'.cache/liftrs/cuda-encoding'/name;cache.mkdir(parents=True,exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',
             GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),
             CGO_ENABLED='1',CXX='g++',GOMAXPROCS='4',GOFLAGS='-p=4',PIR_CUDA_ONLINE_CACHE=str(cache/'code-fixtures'))
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'commands':[],
            'cache':str(cache),
            'policy':{'cpus':16,'rss_guard_gib':12,'deadline_seconds':1500,'device_cap_mib':512,'s_values':[2,3,4,6,8,10],'full_s16':False}}
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
        for label,argv in [('cpu',['lscpu']),('compiler',['g++','--version']),('go',[go,'version']),
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
        run('legacy-build',['make','-C','encoding/rme_legacy','rme'],source)
        env['CGO_ENABLED']='0';run('suite-go',[go,'test','./...','-timeout=8m'],source)
        run('cuda-disabled-build',[go,'test','-tags','liftrs_cuda','./encoding/liftrs-shared','-run','TestCUDAUnavailable','-count=1'],source);env['CGO_ENABLED']='1'
        run('suite-cpu',[go,'test','-tags','liftrs_cpp','./...','-timeout=8m'],source)
        run('suite-cuda-only',[go,'test','-tags','liftrs_cuda','./...','-timeout=8m'],source)
        run('suite-combined',[go,'test','-tags','liftrs_cpp,liftrs_cuda','./...','-timeout=8m'],source)
        run('race',[go,'test','-race','-tags','liftrs_cpp,liftrs_cuda','./encoding/liftrs-shared','-run','TestCUDAEncoderLifecycle|TestCUDAAdmission|TestCPPMatrixTransform','-count=1'],source)
        run('vet',[go,'vet','-tags','liftrs_cpp,liftrs_cuda','./encoding/...'],source)
        binary=cache/'online.test';run('build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary);record['binary_sha256']=HELP['sha'](binary)
        old=json.loads((ROOT/'experiments/gpu-scaling/runs/scaling-02/results.json').read_text());assert old['status']=='passed'
        old_library=Path(old['cache'])/'libgpu-matrix.so';assert HELP['sha'](old_library)==old['cache_sha256'][old_library.name]
        env['PIR_OLD_ONLINE_LIBRARY']=str(old_library);record['old_online_library']={'path':str(old_library),'sha256':HELP['sha'](old_library)}
        fixture_record=json.loads((ROOT/'experiments/gpu-matrix/runs/matrix-01/results.json').read_text());assert fixture_record['status']=='passed'
        fixture=Path(fixture_record['cache'])/'matrix.bin';assert HELP['sha'](fixture)==fixture_record['cache_sha256']['matrix.bin']
        record['matrix_reference']={'path':str(fixture),'sha256':HELP['sha'](fixture)}
        run('matrix-driver-build',['g++','-std=c++17','-O3','-pthread','experiments/gpu-matrix/bench.cpp','encoding/liftrs-cpp/c_api.cpp','-L'+str(cache),'-l:libliftrs-cuda.so','-Wl,-rpath,'+str(cache),'-o',str(cache/'matrix-bench')],source)
        run('matrix-regression',[str(cache/'matrix-bench'),str(fixture)],source)
        (cache/'code-fixtures').mkdir();dest=out/'online-checks';dest.mkdir();env['PIR_CUDA_ONLINE_OUT']=str(dest)
        run('online-checks',[str(binary),'-test.run=^TestCUDAOnlineEncoding$','-test.v','-test.timeout=10m'],source)
        if not shutil.which('compute-sanitizer'):raise RuntimeError('CUDA memcheck required')
        dest=out/'memcheck';dest.mkdir();fixtures=cache/'memcheck-fixtures';fixtures.mkdir();env.update(PIR_CUDA_ONLINE_OUT=str(dest),PIR_CUDA_ONLINE_CACHE=str(fixtures))
        run('memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',str(binary),'-test.run=^TestCUDAOnlineEncoding$','-test.v','-test.timeout=10m'],source)
        env.pop('PIR_CUDA_ONLINE_OUT',None)
        reference=ROOT/'.cache/liftrs/gpu-products/products-02/cpu-0';metadata=json.loads((reference/'online-metadata.json').read_text())
        assert HELP['sha'](reference/'online.bin')==metadata['sha256']
        record['workflow_reference']={'path':str(reference),'online_sha256':metadata['sha256']}
        env.update(PIR_CUDA_WORKFLOW_OUT=str(out/'workflow'),PIR_CUDA_WORKFLOW_CACHE=str(cache/'workflow'),PIR_CUDA_WORKFLOW_REFERENCE=str(reference))
        run('workflow',[str(binary),'-test.run=^TestCUDAEncodingWorkflow$','-test.v','-test.timeout=10m'],source)
        for profile in sorted((out/'workflow').rglob('*.pprof')):
            run('pprof-'+str(profile.relative_to(out/'workflow')).replace('/','-'),[go,'tool','pprof','-top','-nodecount=20',str(binary),str(profile)])
        record['cache_sha256']={str(p.relative_to(cache)):HELP['sha'](p) for p in cache.rglob('*') if p.is_file()}
        record['status']='passed'
    except Exception as error:
        record['status']='stopped' if cancelled else 'failed';record['error']=repr(error)
        raise
    finally:
        record['evidence_sha256']={str(p.relative_to(out)):HELP['sha'](p) for p in out.rglob('*')
                                  if p.is_file() and p.name not in ('results.json','results.json.tmp','analysis.json')}
        save()


if __name__=='__main__':main()
