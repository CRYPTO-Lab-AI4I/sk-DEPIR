#!/usr/bin/env python3
"""Bounded CPU modularity regression, sanitizers and matched old/new timings."""
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
    cache = ROOT / '.cache/liftrs/cpu-encoding-modules' / args.name
    cache.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp') / ('sk-depir-cpu-modules-' + os.environ['SLURM_JOB_ID'])
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
    record = {'baseline_revision':'671bf5e','status':'running', 'job_id':os.environ['SLURM_JOB_ID'],
              'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
              'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(), 'cache':str(cache),
              'policy':{'cpus':4,'request_bytes':8<<30,'rss_guard_bytes':6<<30,'cache_guard_bytes':2<<30,
                        'scratch_guard_bytes':4<<30,'worker_seconds':600,'controller_seconds':900,'slurm_seconds':1200},
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
        record['source_sha256']={str(p.relative_to(source)):HELP['sha'](p) for p in source.rglob('*') if p.is_file() and p.suffix in ('.go','.cpp','.h','.hpp','.py','.sbatch')}
        for label,argv in [('host',['hostname']),('cpu',['lscpu']),('go',[go,'version']),('job',['scontrol','show','job',record['job_id']])]:run(label,argv)
        run('legacy-build',['make','-C','encoding/rme_legacy','rme'],source)
        run('go-build',[go,'test','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        package=source/'encoding/liftrs-shared'
        run('go-regressions',[str(binary),'-test.timeout=10m','-test.cpuprofile='+str(out/'go-regressions.pprof')],package)
        run('go-builds',[go,'test','./encoding/...','./pir/...','./pirconfig/...','-timeout=10m'],source)
        env['CGO_ENABLED']='1';env['CXX']='g++'
        native=cache/'shared-cpp.test'
        run('cpp-build',[go,'test','-tags','liftrs_cpp','-c','-o',str(native),'./encoding/liftrs-shared'],source)
        env['LIFTRS_CPU_PROGRESS']='1'
        fixtures=cache/'fixtures';fixtures.mkdir();env['PIR_WIDE_ENCODING_CACHE']=str(fixtures)
        targeted='^(TestCPUEncodingModules.*|TestCPUWorkspaceLifetime|TestScalingWideCPP.*|TestCPP.*(Ownership|Overlap|Shape)|ExampleScalingCodeData_WithWideCPP)$'
        run('modules',[str(native),'-test.run='+targeted,'-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'modules.pprof')],package)
        env.pop('LIFTRS_CPU_PROGRESS');env.pop('PIR_WIDE_ENCODING_CACHE')
        progress=[]
        for line in (out/'modules.log').read_text().splitlines():
            if line.startswith('{'):
                row=json.loads(line)
                if row.get('work')=='encoding-operations':progress.append(row)
        assert any(p['event']=='completed' and p['completed']==p['total']==10 for p in progress)
        assert any(p['event']=='failed' for p in progress)
        record['native_progress_records']=len(progress)
        run('cpp-regressions',[str(native),'-test.timeout=10m','-test.cpuprofile='+str(out/'cpp-regressions.pprof')],package)
        run('cpp-builds',[go,'test','-tags','liftrs_cpp','./encoding/...','./pir/...','./pirconfig/...','-timeout=10m'],source)
        run('vet',[go,'vet','-tags','liftrs_cpp','./encoding/liftrs-shared','./encoding/liftrs-cpp'],source)
        run('race',[go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared','./encoding/liftrs-cpp','-run','TestCPUWorkspaceLifetime|TestCPPMatrixConcurrent|TestScalingWideCPPLifetime|TestPlanLocks','-count=1','-timeout=10m'],source)
        asan=cache/'shared-asan.test';env['CGO_CXXFLAGS']='-fsanitize=undefined -fno-sanitize-recover=all';env['CGO_LDFLAGS']='-fsanitize=undefined'
        run('asan-build',[go,'test','-asan','-tags','liftrs_cpp','-c','-o',str(asan),'./encoding/liftrs-shared'],source)
        run('asan',[str(asan),'-test.run='+targeted,'-test.v','-test.timeout=10m','-test.cpuprofile='+str(out/'asan.pprof')],package)
        env.pop('CGO_CXXFLAGS');env.pop('CGO_LDFLAGS')
        drivers=[('helpers','experiments/cpu-encoding-modules/check.cpp',[]),
                 ('online','encoding/liftrs-cpp/online_test.cpp',['encoding/liftrs-cpp/c_api.cpp']),
                 ('matrix','encoding/liftrs-cpp/matrix_test.cpp',['encoding/liftrs-cpp/c_api.cpp']),
                 ('taylor','encoding/liftrs-cpp/taylor_parallel_test.cpp',[]),
                 ('wide-schedule','experiments/s17-encoding/check.cpp',['encoding/liftrs-cpp/wide_api.cpp']),
                 ('wide-matrix','experiments/s17-matrix/check.cpp',['encoding/liftrs-cpp/wide_api.cpp'])]
        for mode,flags in [('optimized',['-O3']),('sanitized',['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-fno-omit-frame-pointer'])]:
            for label,driver,extra in drivers:
                target=cache/(label+'-'+mode)
                run(label+'-'+mode+'-build',['g++','-std=c++17','-pthread',*flags,'-Iencoding/liftrs-cpp',driver,*extra,'-o',str(target)],source)
                run(label+'-'+mode,[str(target)],source)
        # Retain an exact old CPU tree, replacing only the benchmark test file.
        baseline=scratch/'baseline'
        shutil.copytree(source,baseline)
        for directory in ('encoding/liftrs-cpp','encoding/liftrs-shared'):
            shutil.rmtree(baseline/directory)
        archive=scratch/'baseline.tar'
        with archive.open('wb') as stream:
            subprocess.run(['git','archive','671bf5e','encoding/liftrs-cpp','encoding/liftrs-shared'],cwd=ROOT,stdout=stream,check=True)
        run('baseline-source',['tar','-xf',str(archive),'-C',str(baseline)])
        bench='encoding/liftrs-shared/cpu_modules_test.go'
        shutil.copyfile(source/bench,baseline/bench)
        record['baseline_sha256']={str(p.relative_to(baseline)):HELP['sha'](p) for d in ('encoding/liftrs-cpp','encoding/liftrs-shared') for p in (baseline/d).rglob('*') if p.is_file()}
        old=cache/'baseline.test';run('baseline-build',[go,'test','-tags','liftrs_cpp','-c','-o',str(old),'./encoding/liftrs-shared'],baseline)
        code=ROOT/'.cache/liftrs/gpu-products/products-02/cpu-0/online.bin'
        expected='3ab552f2ea40fe7b8ee39e2aabde2852f8fc536e9945fb7237fee188f357877e'
        assert HELP['sha'](code)==expected
        env['PIR_CPU_MODULES_CODE']=str(code);env['PIR_CPU_MODULES_SHA256']=expected
        record['benchmark_input']={'path':str(code),'sha256':expected,'s':10,'workers':4}
        # Alternate whole-binary order; exact iteration counts, no profiling.
        for repeat,order in enumerate([('old','new'),('new','old'),('old','new')]):
            for version in order:
                exe=old if version=='old' else native
                cwd=baseline/'encoding/liftrs-shared' if version=='old' else package
                run('timing-'+str(repeat)+'-'+version,[str(exe),'-test.run=^$','-test.bench=^BenchmarkCPUEncodingModules$','-test.benchtime=3x','-test.count=1','-test.timeout=10m'],cwd)
        run('profile',[str(native),'-test.run=^$','-test.bench=^BenchmarkCPUEncodingModules$','-test.benchtime=1x','-test.count=1','-test.cpuprofile='+str(out/'encoding.pprof'),'-test.timeout=10m'],package)
        run('profile-summary',[go,'tool','pprof','-top',str(native),str(out/'encoding.pprof')])
        record['status']='passed'
    except BaseException as error:
        record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):HELP['sha'](p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):HELP['sha'](p) for p in out.rglob('*') if p.is_file() and p.name!='results.json'}
        save()


if __name__=='__main__':main()
