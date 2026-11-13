#!/usr/bin/env python3
"""Bounded CPU-only scalar checks; no domain allocations or existing-job inspection."""
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
    args = parser.parse_args()
    if args.check:
        ast.parse(Path(__file__).read_text())
        assert 600 < 900 < 1200
        print('syntax and worker/controller/Slurm deadline order checked')
        return
    if not args.name or not args.name.replace('-', '').isalnum(): parser.error('fresh run name required')
    out = HERE / 'runs' / args.name
    out.mkdir(parents=True, exist_ok=False)
    cache = ROOT / '.cache/liftrs/s17-arithmetic' / args.name
    cache.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp') / ('sk-depir-wide-' + os.environ['SLURM_JOB_ID'])
    scratch.mkdir(exist_ok=False)
    env = {k: v for k, v in os.environ.items() if not k.startswith('PIR_')}
    env.update(CGO_ENABLED='0', GOMAXPROCS='4', GOFLAGS='-p=2', GOTOOLCHAIN='local',
               GOPROXY='off', GOSUMDB='off', GOMEMLIMIT=str(4<<30),
               GOPATH=str(ROOT / '.cache/liftrs/toolchains/gopath'), GOCACHE=str(scratch / 'go-cache'), TMPDIR=str(scratch))
    go = str(ROOT / '.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')
    binary = cache / 'wide.test'
    start = time.monotonic()
    cancelled = []
    signal.signal(signal.SIGTERM, lambda *_: cancelled.append(True))
    record = {'status':'running', 'job_id':os.environ['SLURM_JOB_ID'],
              'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
              'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(), 'cache':str(cache),
              'policy':{'cpus':4,'request_bytes':8<<30,'rss_guard_bytes':6<<30,'cache_guard_bytes':2<<30,
                        'scratch_guard_bytes':4<<30,'worker_seconds':600,'controller_seconds':900,'slurm_seconds':1200},
              'commands':[]}
    def save():
        (out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
    def run(label,argv,cwd=ROOT):
        began=time.monotonic();peak=0;reason=None
        with (out/(label+'.log')).open('w') as log:
            proc=subprocess.Popen(argv,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while proc.poll() is None:
                    peak=max(peak,HELP['group_rss'](proc.pid))
                    if peak>6<<30:reason='RSS guard'
                    if GUARD['tree_bytes'](scratch)>4<<30 or GUARD['tree_bytes'](cache)>2<<30:reason='storage guard'
                    if time.monotonic()-began>600 or time.monotonic()-start>900:reason='deadline'
                    if cancelled:reason='termination signal'
                    if reason:GUARD['stop_group'](proc);break
                    time.sleep(2)
            finally:
                if proc.poll() is None:GUARD['stop_group'](proc)
        record['commands'].append({'label':label,'argv':argv,'seconds':time.monotonic()-began,
                                   'peak_rss_bytes':peak,'exit_code':proc.returncode,'guard':reason});save()
        if reason or proc.returncode:raise RuntimeError(label+': '+str(reason or proc.returncode))
    save()
    try:
        source=scratch/'repo'
        shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):HELP['sha'](p) for p in source.rglob('*.go')}
        for label,argv in [('host',['hostname']),('cpu',['lscpu']),('go',[go,'version']),('job',['scontrol','show','job',record['job_id']])]:run(label,argv)
        run('build',[go,'test','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        run('arithmetic',[str(binary),'-test.run=^(TestWideOrbitArithmetic|TestTableFreeFields)$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'cpu.pprof')],source)
        text=(out/'arithmetic.log').read_text()
        assert '--- PASS: TestWideOrbitArithmetic' in text and '--- PASS: TestTableFreeFields' in text
        run('profile',[go,'tool','pprof','-top','-nodecount=25',str(binary),str(out/'cpu.pprof')])
        record['status']='passed'
    except BaseException as error:
        record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):HELP['sha'](p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):HELP['sha'](p) for p in out.rglob('*') if p.is_file() and p.name!='results.json'}
        save()


if __name__=='__main__':main()
