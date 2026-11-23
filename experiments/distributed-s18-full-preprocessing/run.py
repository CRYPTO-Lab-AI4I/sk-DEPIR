#!/usr/bin/env python3
"""Explicitly admitted full s18 trial; shared numerical command, sealed source."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import runpy
import signal
import shutil
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
BASE = runpy.run_path(str(ROOT/'experiments/gpu-staged-transport/run.py'))
STOP = runpy.run_path(str(ROOT/'experiments/gpu-preprocessing-workflow/run.py'))['stop_process']
GIB = 1 << 30
K = 16300262095


def config(s, directory, library, budget):
    words = (1 << (2*s))//budget['ranks']
    return dict(spec=dict(FieldDegree=s, Dimension=2, CurveDegree=5), ranks=budget['ranks'],
        directory=str(directory), library=str(library), device=0,
        device_bytes=(budget['owned_gpu_GiB_per_rank']*GIB if s==18 else 32<<20),
        max_local_words=words, max_good_words=words, max_root_words=words-1,
        max_archive_words=0, rss_bytes=(budget['rss_GiB_per_rank']*GIB if s==18 else GIB),
        host_reserve_bytes=(budget['host_reserve_GiB_per_rank']*GIB if s==18 else 768<<20),
        storage_bytes=(budget['output_GiB']*GIB if s==18 else 32<<20),
        file_bytes=(budget['file_GiB']*GIB if s==18 else 32<<20),
        deadline_seconds=(budget['worker_hours']*3600 if s==18 else 300),
        newton_input='prefix', spectrum='resident', discovery_result='stream', proof_log='summary')


def validate(b):
    assert b['nodes']==8 and b['ranks']==16 and b['gpus_per_node']==2 and b['cpus_per_rank']==4
    assert 2*b['rss_GiB_per_rank']+64 <= b['host_request_GiB_per_node']
    # Worst admitted lengths, including 2x append capacity, not uniform K/R.
    words = (1<<36)//16
    named = 44*words + (words+7)//8 + (16<<20)
    assert named+b['host_reserve_GiB_per_rank']*GIB < b['rss_GiB_per_rank']*GIB
    assert 3*words*8 + (1<<30) < b['owned_gpu_GiB_per_rank']*GIB
    assert b['worker_hours'] < b['controller_hours'] < b['scheduler_hours'] == 100
    assert b['output_GiB']+128 <= b['retained_GiB'] < b['minimum_initial_free_GiB']


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name', nargs='?')
    parser.add_argument('--check', action='store_true')
    args=parser.parse_args()
    budget=json.loads((HERE/'budget.json').read_text()); validate(budget)
    if args.check:
        print(json.dumps(dict(status='resource/configuration plan valid;no GPU initialized or job submitted',budget=budget)))
        return
    if not budget['admitted']:
        raise RuntimeError('fullsize resource agreement has not been recorded')
    if not args.name or any(c not in 'abcdefghijklmnopqrstuvwxyz0123456789-' for c in args.name):
        raise ValueError('fresh lowercase run name required')
    if os.environ.get('SLURM_JOB_NUM_NODES')!='8' or os.environ.get('SLURM_NTASKS')!='16':
        raise RuntimeError('requires approved8-node16-task allocation')
    if shutil.disk_usage(ROOT).free < budget['minimum_initial_free_GiB']*GIB:
        raise RuntimeError('insufficient retained-storage headroom')
    output=HERE/'runs'/args.name; output.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/distributed-s18-full-preprocessing'/args.name; cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('depir-s18-'+os.environ['SLURM_JOB_ID']); scratch.mkdir(exist_ok=False)
    source=scratch/'source'; source.mkdir()
    began=time.monotonic()
    env=dict(os.environ, GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',
        GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),
        TMPDIR=str(scratch),CGO_ENABLED='1',CC='mpicc',CXX='mpicxx',GOMAXPROCS='4',
        GOFLAGS='-p=4',LIFTRS_CPU_PROGRESS='1',GOMEMLIMIT=str(budget['go_soft_GiB_per_rank']*GIB),LIFTRS_GPU_TAYLOR_PROGRESS='bounded')
    record=dict(status='running',job_id=os.environ['SLURM_JOB_ID'],budget=budget,
        revision=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),commands=[],cache=str(cache))
    def save():
        temporary=output/'results.json.tmp'; temporary.write_text(json.dumps(record,indent=2)+'\n'); temporary.replace(output/'results.json')
    def event(phase,kind,**values):
        row=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),phase=phase,event=kind,
                 elapsed_seconds=time.monotonic()-began,**values)
        with (output/'progress.jsonl').open('a') as stream: stream.write(json.dumps(row)+'\n'); stream.flush()
        print(json.dumps(row),flush=True)
    def size(path):
        total=0
        for file in path.rglob('*'):
            try:
                if file.is_file(): total+=file.stat().st_size
            except FileNotFoundError: pass # atomic temporary-file publication
        return total
    def run(label,command,seconds=180):
        event(label,'started'); start=last=time.monotonic()
        code=None
        with (output/(label+'.log')).open('w') as log:
            child=subprocess.Popen(list(map(str,command)),cwd=source,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while code is None:
                    now=time.monotonic()
                    if now-start>seconds or now-began>budget['controller_hours']*3600:
                        raise TimeoutError(label+' deadline')
                    if now-last>=30:
                        retained=size(cache)+size(output); temporary=size(scratch)
                        if retained>budget['retained_GiB']*GIB or temporary>budget['scratch_GiB']*GIB or shutil.disk_usage(cache).free<budget['minimum_free_GiB']*GIB:
                            raise RuntimeError('retained/scratch/free-space guard')
                        event(label,'resource-observation',retained_bytes=retained,scratch_bytes=temporary,
                              scope='not forward progress;see rank logs for completed tiles/passes')
                        last=now
                    try: code=child.wait(timeout=1)
                    except subprocess.TimeoutExpired: pass
            finally:
                if child.poll() is None: STOP(child)
                record['commands'].append(dict(label=label,command=list(map(str,command)),seconds=time.monotonic()-start,exit_code=child.returncode)); save()
        if code: raise subprocess.CalledProcessError(code,command)
        event(label,'completed')
    def terminate(*_): raise InterruptedError('termination signal;completed checkpoints retained')
    signal.signal(signal.SIGTERM,terminate)
    save()
    try:
        record['source_sha256']=BASE['snapshot'](record['revision'],source); save()
        go=ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go'
        library=cache/'libliftrs-cuda.so'; binary=cache/'pir-preprocess-mpi'
        run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=sm_90','-shared','-Xcompiler=-fPIC,-pthread',
            'encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/preparation.cu','-o',library])
        run('command-build',[go,'build','-tags','liftrs_cpp,liftrs_cuda,liftrs_mpi','-o',binary,'./cmd/pir-preprocess-mpi'])
        wrapper=cache/'rank.sh'; shutil.copyfile(source/'cmd/pir-preprocess-mpi/rank.sh',wrapper)
        node=cache/'node.sh'; shutil.copyfile(source/'experiments/distributed-s18-full-preprocessing/node.sh',node)
        launcher=['srun','--mpi=pmix_v3','--nodes=8','--ntasks=16','--ntasks-per-node=2','--cpus-per-task=4',
                  '--gpus-per-task=h200:1','--gpu-bind=single:1','--cpu-bind=cores','--kill-on-bad-exit=1']
        for s in (8,18):
            directory=cache/f's{s}'; directory.mkdir()
            path=cache/f's{s}.json'; path.write_text(json.dumps(config(s,directory,library,budget))+'\n')
            run(f'config-s{s}',[binary,'--config',path,'--check-config'])
            if s==8:
                run('node-preflight',['srun','--nodes=8','--ntasks=8','--ntasks-per-node=1','--cpus-per-task=1',
                    '--kill-on-bad-exit=1','bash',node,binary,library,path,output])
            run(f's{s}',[*launcher,'bash',wrapper,output/f's{s}',binary,'--config',path],
                360 if s==8 else budget['worker_hours']*3600+60)
            assert all((output/f's{s}-{r}.pprof').stat().st_size>0 for r in range(16))
            publication=json.loads((directory/f's{s}-publication.json').read_text())
            assert publication['identity_verified'] is True and publication['header']['K']==(7381 if s==8 else K)
            # Reconcile small canonical files independently before large allocations.
            # For fullsize, streaming SHA verification is an explicit separate phase.
            expected=None
            if s==8:
                expected=runpy.run_path(str(source/'experiments/distributed-host-admission/reference.py'))['REFERENCE']['files']
            event(f'verify-s{s}','started',total=sum(f['bytes'] for f in publication['files']),unit='bytes')
            for item in publication['files']:
                target=directory/item['name']; assert target.stat().st_size==item['bytes']
                if expected is not None: assert item==expected[item['name']]
                digest=hashlib.sha256();done=0;last=time.monotonic()
                with target.open('rb') as stream:
                    while block:=stream.read(8<<20):
                        digest.update(block);done+=len(block)
                        now=time.monotonic()
                        if now-began>budget['controller_hours']*3600: raise TimeoutError('verification deadline')
                        if now-last>=30:
                            event(f'verify-s{s}','progress',file=item['name'],completed=done,total=item['bytes'],unit='bytes');last=now
                assert digest.hexdigest()==item['sha256']
            record[f's{s}_publication']=publication; save(); event(f'verify-s{s}','completed')
        record['status']='passed'
    except BaseException as error:
        record.update(status='failed',error=repr(error)); raise
    finally:
        record['elapsed_seconds']=time.monotonic()-began; save()
        # Keep all successful checkpoints, payloads, profiles and diagnostic files.


if __name__=='__main__': main()
