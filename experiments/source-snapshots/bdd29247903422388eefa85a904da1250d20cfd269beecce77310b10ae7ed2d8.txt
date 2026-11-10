#!/usr/bin/env python3
"""Read-only s16 support sampling and bounded preparation stage profiling."""
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
    if len(sys.argv)!=2 or not sys.argv[1].replace('-','').isalnum(): raise SystemExit('Use a fresh alphanumeric/hyphen run name')
    job=os.environ['SLURM_JOB_ID']; name=sys.argv[1]
    out=HERE/'runs'/name;out.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-target-'+job);scratch.mkdir(exist_ok=False)
    cache=ROOT/'.cache/liftrs/preparation-profile'/name;cache.mkdir(parents=True,exist_ok=False)
    support=ROOT/'.cache/liftrs/hpc-s16/s16-m2-t5'
    reference=json.loads((ROOT/'experiments/hpc-s16/runs/first-01/interruption.json').read_text())['basis']
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',
             GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),
             CGO_ENABLED='1',CXX='g++',GOMAXPROCS='4',GOFLAGS='-p=4',
             PIR_TARGET_OUT=str(out),PIR_TARGET_SUPPORT=str(support),PIR_TARGET_SHA256=reference['sha256'])
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'commands':[],
            'support_reference':reference,'support_directory':str(support),'cache':str(cache),
            'policy':{'cpus':4,'rss_guard_gib':12,'deadline_seconds':1500,'sampled_exponents':65536,'maximum_transform_points':1<<22}}
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
            if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.py','.sbatch') or p.name in ('go.mod','go.sum','Makefile'))}
        for path,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+path],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists(): dest.write_bytes((source/path).read_bytes())
        binary=cache/'profile.test'
        run('build',[go,'test','-tags','liftrs_cpp','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary);record['binary_sha256']=HELP['sha'](binary)
        run('profile',[str(binary),'-test.run=^TestPreparationTargetProfile$','-test.v','-test.timeout=20m'])
        for p in sorted(out.glob('*.pprof')):
            run('pprof-'+p.stem,[go,'tool','pprof','-top','-nodecount=20',str(binary),str(p)])
        record['status']='passed'
    except Exception as error:
        record['status']='stopped' if cancelled else 'failed';record['error']=repr(error)
        raise
    finally:
        record['evidence_sha256']={str(p.relative_to(out)):HELP['sha'](p) for p in out.rglob('*')
                                  if p.is_file() and p.name not in ('results.json','results.json.tmp','analysis.json')}
        save()


if __name__=='__main__':main()
