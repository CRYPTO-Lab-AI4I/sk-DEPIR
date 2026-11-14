#!/usr/bin/env python3
"""Bounded resident GPU preprocessing: shared widths, peer shards and legacy regressions."""
import ast
import datetime
import json
import os
from pathlib import Path
import runpy
import shutil
import signal
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
HELP = runpy.run_path(str(ROOT / 'experiments/hpc-cached-s15/run.py'))
GUARD = runpy.run_path(str(ROOT / 'experiments/hpc-s16/run.py'))


def main():
    import argparse
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name', nargs='?')
    parser.add_argument('--check', action='store_true')
    parser.add_argument('--devices', type=int, choices=(1,2,4), default=1)
    args = parser.parse_args()
    if args.check:
        ast.parse(Path(__file__).read_text())
        assert 600 < 900 < 1200
        assert ((3**17)+(1<<17)*4+17*8+8) <= (512<<20)//4, "wide-window driver exceeds smallest per-card cap"
        print('syntax and worker/controller/Slurm deadline order checked')
        return
    if not args.name or not args.name.replace('-', '').isalnum(): parser.error('fresh run name required')
    out = HERE / 'runs' / args.name
    out.mkdir(parents=True, exist_ok=False)
    cache = ROOT / '.cache/liftrs/gpu-s17-preprocessing' / args.name
    cache.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp') / ('sk-depir-gpu-preparation-' + os.environ['SLURM_JOB_ID'])
    scratch.mkdir(exist_ok=False)
    env = {k: v for k, v in os.environ.items() if not k.startswith('PIR_')}
    env.update(CGO_ENABLED='0', GOMAXPROCS='4', GOFLAGS='-p=2', GOTOOLCHAIN='local',
               GOPROXY='off', GOSUMDB='off', GOMEMLIMIT=str(4<<30),
               GOPATH=str(ROOT / '.cache/liftrs/toolchains/gopath'), GOCACHE=str(scratch / 'go-cache'), TMPDIR=str(scratch))
    go = str(ROOT / '.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    binary = cache / 'shared-go.test'
    start = time.monotonic()
    cancelled = []
    signal.signal(signal.SIGTERM, lambda *_: cancelled.append(True))
    record = {'status':'running', 'job_id':os.environ['SLURM_JOB_ID'],
              'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
              'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(), 'cache':str(cache),
              'policy':{'cpus':4,'request_bytes':8<<30,'rss_guard_bytes':6<<30,'cache_guard_bytes':2<<30,
                        'scratch_guard_bytes':4<<30,'worker_seconds':600,'controller_seconds':900,'slurm_seconds':1200,'gpus':args.devices,'aggregate_device_cap_bytes':512<<20,'per_device_cap_bytes':(512<<20)//args.devices},
              'cuda_visible_devices':env.get('CUDA_VISIBLE_DEVICES'),
              'slurm_job_gpus':env.get('SLURM_JOB_GPUS'),'visible_ordinals':list(range(args.devices)),
              'commands':[]}
    def save():
        (out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
    def event(label, kind, **values):
        entry = dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
                     phase=label, event=kind, completed_commands=len(record['commands']),
                     elapsed_seconds=time.monotonic()-start, **values)
        with (out/'progress.jsonl').open('a') as stream:
            stream.write(json.dumps(entry)+'\n'); stream.flush()
        print(json.dumps(entry),flush=True)
    def run(label,argv,cwd=ROOT):
        began=time.monotonic();peak=0;reason=None;last=began
        event(label,'started')
        with (out/(label+'.log')).open('w') as log:
            proc=subprocess.Popen(argv,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while proc.poll() is None:
                    peak=max(peak,HELP['group_rss'](proc.pid))
                    if peak>6<<30:reason='RSS guard'
                    if GUARD['tree_bytes'](scratch)>4<<30 or GUARD['tree_bytes'](cache)>2<<30:reason='storage guard'
                    if time.monotonic()-began>600 or time.monotonic()-start>900:reason='deadline'
                    if cancelled:reason='termination signal'
                    if time.monotonic()-last>=30:
                        event(label,'heartbeat',command_seconds=time.monotonic()-began,
                              peak_rss_bytes=peak,remaining_work='unknown')
                        last=time.monotonic()
                    if reason:GUARD['stop_group'](proc);break
                    time.sleep(2)
            finally:
                if proc.poll() is None:GUARD['stop_group'](proc)
        record['commands'].append({'label':label,'argv':argv,'seconds':time.monotonic()-began,
                                   'peak_rss_bytes':peak,'exit_code':proc.returncode,'guard':reason});save()
        event(label,'failed' if reason or proc.returncode else 'completed',peak_rss_bytes=peak,guard=reason)
        if reason or proc.returncode:raise RuntimeError(label+': '+str(reason or proc.returncode))
    save()
    try:
        source=scratch/'repo'
        shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):HELP['sha'](p) for p in source.rglob('*') if p.is_file() and p.suffix in ('.go','.cpp','.h','.hpp','.cu','.cuh','.py','.sbatch')}
        for label,argv in [('host',['hostname']),('cpu',['lscpu']),('go',[go,'version']),('compiler',['g++','--version']),('job',['scontrol','show','job',record['job_id']])]:run(label,argv)
        run('legacy-build',['make','-C','encoding/rme_legacy','rme'],source)
        run('go-build',[go,'test','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        package=source/'encoding/liftrs-shared'
        run('go-regressions',[str(binary),'-test.timeout=10m','-test.cpuprofile='+str(out/'go-regressions.pprof')],package)
        run('go-builds',[go,'test','./encoding/...','./pir/...','./pirconfig/...','-timeout=10m'],source)
        env['CGO_ENABLED']='1';env['CXX']='g++'
        native=cache/'shared-cpp.test'
        run('cpp-build',[go,'test','-tags','liftrs_cpp','-c','-o',str(native),'./encoding/liftrs-shared'],source)
        library=cache/'libliftrs-cuda.so'
        run('gpu',['nvidia-smi','--query-gpu=name,uuid,driver_version,memory.total,compute_cap','--format=csv'])
        run('topology',['nvidia-smi','topo','-m'])
        run('nvcc',['nvcc','--version'])
        run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native','-shared','-Xcompiler=-fPIC','encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/preparation.cu','-o',str(library)],source)
        window=cache/'monomial-windows'
        run('window-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native','-Iencoding/liftrs-cuda','experiments/gpu-s17-preprocessing/check.cu','-o',str(window)],source)
        run('wide-monomial-windows',[str(window)],source)
        gpu=cache/'shared-cuda.test'
        run('cuda-go-build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','-c','-o',str(gpu),'./encoding/liftrs-shared'],source)
        fixtures=cache/'fixtures';fixtures.mkdir()
        env.update(PIR_GPU_LIBRARY=str(library),PIR_GPU_PREPARATION_LIBRARY=str(library),
                   PIR_GPU_PREPARATION_DEVICES=','.join(str(d) for d in range(args.devices)),
                   PIR_GPU_PREPARATION_CAP=str((512<<20)//args.devices),
                   PIR_GPU_PREPARATION_CACHE=str(fixtures),PIR_GPU_PREPARATION_OUT=str(out))
        target='^TestGPUPreparation(Small|Wide|Tile)$'
        env['LIFTRS_CPU_PROGRESS']='1'
        run('gpu-preprocessing',[str(gpu),'-test.run='+target,'-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'gpu-preprocessing.pprof')],package)
        env.pop('LIFTRS_CPU_PROGRESS')
        env.pop('PIR_GPU_PREPARATION_CACHE');env.pop('PIR_GPU_PREPARATION_OUT')
        if not shutil.which('compute-sanitizer'):raise RuntimeError('compute-sanitizer required for GPU validation')
        run('memcheck',['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',str(gpu),'-test.run=^TestGPUPreparation(Small|Wide)$','-test.v','-test.timeout=10m'],package)
        env['PIR_CUDA_ONLINE_OUT']=str(out);env['PIR_CUDA_ONLINE_CACHE']=str(fixtures)
        run('legacy-online',[str(gpu),'-test.run=^TestCUDAOnlineEncoding$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'legacy-online.pprof')],package)
        env.pop('PIR_CUDA_ONLINE_OUT');env.pop('PIR_CUDA_ONLINE_CACHE')
        run('cpp-regressions',[str(native),'-test.timeout=10m','-test.cpuprofile='+str(out/'cpp-regressions.pprof')],package)
        run('cuda-regressions',[str(gpu),'-test.run=TestCUDAEncoderLifecycle|TestCUDAAdmission|TestGPUPreparation|TestPreprocessing','-test.timeout=10m'],package)
        run('vet',[go,'vet','-tags','liftrs_cpp,liftrs_cuda','./encoding/liftrs-shared','./encoding/liftrs-cuda'],source)
        for mode,flags in [('optimized',['-O3']),('sanitized',['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-fno-omit-frame-pointer'])]:
            target=cache/('shared-recurrences-'+mode)
            run(mode+'-build',['g++','-std=c++17','-pthread',*flags,'-Iencoding/liftrs-cuda','experiments/gpu-s17-preprocessing/check.cpp','-o',str(target)],source)
            run(mode,[str(target)],source)
        run('profile-summary',[go,'tool','pprof','-top',str(gpu),str(out/'gpu-preprocessing.pprof')])
        record['status']='passed'
    except BaseException as error:
        record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):HELP['sha'](p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):HELP['sha'](p) for p in out.rglob('*') if p.is_file() and p.name!='results.json'}
        save()


if __name__=='__main__':main()
