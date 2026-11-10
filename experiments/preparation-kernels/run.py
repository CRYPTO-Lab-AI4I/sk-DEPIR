#!/usr/bin/env python3
"""Shared bounded controller for the matched CPU and GPU component trials."""
import argparse,hashlib,json,os,runpy,shutil,signal,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
STOP=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))['stop_group']

def main():
    p=argparse.ArgumentParser();p.add_argument('mode',choices=['cpu','gpu']);p.add_argument('name');p.add_argument('--cpu-run',default='cpu-01');a=p.parse_args()
    for name in (a.name,a.cpu_run):
        if not name.replace('-','').isalnum():raise ValueError('invalid name')
    here=ROOT/'experiments'/('preparation-kernels' if a.mode=='cpu' else 'gpu-transform-tiles')
    out=here/'runs'/a.name;out.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-trials-'+os.environ['SLURM_JOB_ID']);scratch.mkdir(exist_ok=False)
    cache=ROOT/'.cache/liftrs'/here.name/a.name;cache.mkdir(parents=True,exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='4',GOFLAGS='-p=4',PIR_TRIAL_OUT=str(cache))
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    r={'status':'running','mode':a.mode,'job_id':os.environ['SLURM_JOB_ID'],'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'cache':str(cache),'commands':[],
       'policy':{'cpus':4,'rss_guard_gib':12,'deadline_seconds':1500,'gpu_count':int(a.mode=='gpu'),'device_allocation_cap_mib':512 if a.mode=='gpu' else 0}}
    start=time.monotonic();cancel=[];signal.signal(signal.SIGTERM,lambda *_:cancel.append(True))
    def save():
        tmp=out/'results.json.tmp';tmp.write_text(json.dumps(r,indent=2)+'\n');tmp.replace(out/'results.json')
    def run(label,argv,cwd=ROOT,extra=None):
        begin=time.monotonic();peak=0;reason=None
        with (out/(label+'.log')).open('w') as log:
            child=subprocess.Popen(argv,cwd=cwd,env=dict(env,**(extra or {})),stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while child.poll() is None:
                    peak=max(peak,HELP['group_rss'](child.pid))
                    if peak>12*2**30:reason='RSS guard'
                    if time.monotonic()-start>1500:reason='deadline'
                    if cancel:reason='Slurm/user termination'
                    if reason:STOP(child);break
                    time.sleep(1)
            finally:
                if child.poll() is None:STOP(child)
        r['commands'].append({'label':label,'argv':list(map(str,argv)),'seconds':time.monotonic()-begin,'peak_rss_bytes':peak,'guard':reason,'exit_code':child.returncode});save()
        if reason or child.returncode:raise RuntimeError(label+': '+str(reason or child.returncode))
    save()
    try:
        for label,cmd in [('cpu',['lscpu']),('compiler',['g++','--version']),('job',['scontrol','show','job',r['job_id']]),('go',[go,'version']),('os',['cat','/etc/os-release'])]:run(label,cmd)
        src=scratch/'repo';shutil.copytree(ROOT,src,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        r['source_sha256']={str(f.relative_to(src)):HELP['sha'](f) for f in src.rglob('*') if f.is_file() and (f.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or f.name in ('go.mod','go.sum','Makefile'))}
        for name,digest in r['source_sha256'].items():
            blob=subprocess.run(['git','show',r['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists():dest.write_bytes((src/name).read_bytes())
        if a.mode=='cpu':
            binary=cache/'cpu.test';run('go-build',[go,'test','-tags','liftrs_cpp','-c','-o',str(binary),'./encoding/liftrs-shared'],src)
            run('go-kernels',[str(binary),'-test.run=^TestPreparationCPUTrials$','-test.v','-test.timeout=12m'])
            shutil.copy2(cache/'cpu-go.json',out/'cpu-go.json');shutil.copy2(cache/'cpu-kernels.pprof',out/'cpu-kernels.pprof')
            run('pprof',[go,'tool','pprof','-top','-nodecount=20',str(binary),str(out/'cpu-kernels.pprof')])
            run('cpp-build',['g++','-std=c++17','-O3','-pthread','experiments/preparation-kernels/cpu.cpp','-o',str(cache/'cpu-native')],src)
            run('cpp-kernels',[str(cache/'cpu-native'),str(cache/'fixtures.bin')])
            for cutoff in (32768,1024,128):
                for repeat in range(2):
                    label=f'full-{cutoff}-{repeat}';dest=out/label;dest.mkdir()
                    run(label,[str(binary),'-test.run=^TestPreparationCrossoverFull$','-test.v','-test.timeout=5m'],extra={'PIR_TRIAL_FULL_OUT':str(dest),'PIR_TRIAL_FULL_CACHE':str(cache/label),'PIR_TRIAL_CUTOFF':str(cutoff)})
        else:
            reference=json.loads((ROOT/'experiments/preparation-kernels/runs'/a.cpu_run/'results.json').read_text());assert reference['status']=='passed'
            fixture=Path(reference['cache'])/'fixtures.bin';assert HELP['sha'](fixture)==reference['cache_sha256']['fixtures.bin'];r['fixture_sha256']=HELP['sha'](fixture);r['cpu_reference']=a.cpu_run
            run('gpu',['nvidia-smi','--query-gpu=name,uuid,driver_version,memory.total,compute_cap','--format=csv'])
            run('nvcc',['nvcc','--version'])
            run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native','-Xcompiler=-pthread','experiments/gpu-transform-tiles/tiles.cu','-o',str(cache/'gpu-tiles')],src)
            run('gpu-tiles',[str(cache/'gpu-tiles'),str(fixture)])
        r['cache_sha256']={str(f.relative_to(cache)):HELP['sha'](f) for f in cache.rglob('*') if f.is_file()}
        r['status']='passed'
    except Exception as error:
        r['status']='stopped' if cancel else 'failed';r['error']=repr(error);raise
    finally:
        r['evidence_sha256']={str(f.relative_to(out)):HELP['sha'](f) for f in out.rglob('*') if f.is_file() and f.name not in ('results.json','results.json.tmp','analysis.json')};save()

if __name__=='__main__':main()
