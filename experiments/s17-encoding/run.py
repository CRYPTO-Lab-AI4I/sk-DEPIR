#!/usr/bin/env python3
"""Bounded complete wide encoding integration with retained small code fixtures."""
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
    cache = ROOT / '.cache/liftrs/s17-encoding' / args.name
    cache.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp') / ('sk-depir-wide-encoding-' + os.environ['SLURM_JOB_ID'])
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
        record['source_sha256']={str(p.relative_to(source)):HELP['sha'](p) for p in source.rglob('*') if p.is_file() and p.suffix in ('.go','.cpp','.h','.hpp','.py','.sbatch')}
        for label,argv in [('host',['hostname']),('cpu',['lscpu']),('go',[go,'version']),('job',['scontrol','show','job',record['job_id']])]:run(label,argv)
        run('build',[go,'test','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        package=source/'encoding/liftrs-shared'
        for path in ('../../experiments/encoder-scaling/artifacts/s10-m2-t6-cyclic-fft.json',
                     '../../experiments/encoder-comparison/artifacts/s04-m2-t2-cyclic-fft.json'):
            assert (package/path).is_file(),path
        run('go-regressions',[str(binary),'-test.timeout=10m','-test.cpuprofile='+str(out/'go-regressions.pprof')],package)
        assert 'PASS' in (out/'go-regressions.log').read_text()
        env['CGO_ENABLED']='1';env['CXX']='g++'
        native=cache/'shared-cpp.test'
        run('cpp-build',[go,'test','-tags','liftrs_cpp','-c','-o',str(native),'./encoding/liftrs-shared'],source)
        fixtures=cache/'fixtures';fixtures.mkdir();env['PIR_WIDE_ENCODING_CACHE']=str(fixtures)
        run('encoding',[str(native),'-test.run=^(TestScalingWideCPP.*|ExampleScalingCodeData_WithWideCPP)$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'encoding.pprof')],package)
        text=(out/'encoding.log').read_text()
        for test in ('TestScalingWideCPPEncoding','TestScalingWideCPPLifetime','ExampleScalingCodeData_WithWideCPP'):
            assert '--- PASS: '+test+' ' in text,test
        env.pop('PIR_WIDE_ENCODING_CACHE')
        for meta in fixtures.glob('*/metadata.json'):
            m=json.loads(meta.read_text());assert m['identity_verified'] and HELP['sha'](meta.parent/'online.bin')==m['sha256'] and HELP['sha'](meta.parent/'g.bin')==m['generator_sha256']
        assert len(list(fixtures.glob('*/metadata.json')))==4
        run('profile',[go,'tool','pprof','-top',str(native),str(out/'encoding.pprof')])
        run('cpp-regressions',[str(native),'-test.timeout=10m','-test.cpuprofile='+str(out/'cpp-regressions.pprof')],package)
        bound=cache/'native-api.test';run('native-api-build',[go,'test','-tags','liftrs_cpp','-c','-o',str(bound),'./encoding/liftrs-cpp'],source)
        run('native-api',[str(bound),'-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'native-api.pprof')],source/'encoding/liftrs-cpp')
        run('lifetime-race',[go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared','-run','TestScalingWideCPPLifetime','-count=1','-timeout=10m'],source)
        asan=cache/'shared-asan.test';env['CGO_CXXFLAGS']='-fsanitize=undefined -fno-sanitize-recover=all';env['CGO_LDFLAGS']='-fsanitize=undefined'
        run('asan-build',[go,'test','-asan','-tags','liftrs_cpp','-c','-o',str(asan),'./encoding/liftrs-shared'],source)
        run('asan',[str(asan),'-test.run=^(TestScalingWideCPP.*|ExampleScalingCodeData_WithWideCPP)$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'asan.pprof')],package)
        env.pop('CGO_CXXFLAGS');env.pop('CGO_LDFLAGS')
        for label,flags in [('schedule',['-O3']),('schedule-sanitizer',['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-fno-omit-frame-pointer'])]:
            target=cache/label
            run(label+'-build',['g++','-std=c++17','-pthread',*flags,'-Iencoding/liftrs-cpp','experiments/s17-encoding/check.cpp','encoding/liftrs-cpp/wide_api.cpp','-o',str(target)],source)
            run(label,[str(target)],source)
            r=json.loads((out/(label+'.log')).read_text());assert r['cases']==12 and r['match'];record[label]=r
        record['status']='passed'
    except BaseException as error:
        record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):HELP['sha'](p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):HELP['sha'](p) for p in out.rglob('*') if p.is_file() and p.name!='results.json'}
        save()


if __name__=='__main__':main()
