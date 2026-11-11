#!/usr/bin/env python3
"""Validate within-block Taylor gather and compare exact old/new cached encoding."""
import datetime,hashlib,json,os,runpy,shutil,signal,subprocess,sys,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
OLD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
STOP=OLD['stop_group']
sha=HELP['sha']

def main():
    if len(sys.argv)!=2 or not sys.argv[1] or not sys.argv[1].replace('-','').isalnum():raise SystemExit('Usage: run.py FRESH_RUN_NAME')
    name=sys.argv[1];job=os.environ['SLURM_JOB_ID']
    out=HERE/'runs'/name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/cpu-taylor'/name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-cpu-taylor-'+job);scratch.mkdir(exist_ok=False)
    policy={'workers':[16,32],'repeats':2,'rss_guard_bytes':12<<30,'deadline_seconds':1500,'new_cache_guard_bytes':8<<30,'backend':'native','s':14}
    r={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'policy':policy,'cache':str(cache),'commands':[]}
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='32',GOFLAGS='-p=4')
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');began=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    def save():
        p=out/'results.json.tmp';p.write_text(json.dumps(r,indent=2)+'\n');p.replace(out/'results.json')
    def run(label,argv,cwd=ROOT,required=True):
        start=time.monotonic();peak=0;reason=None
        with (out/(label+'.log')).open('w') as log:
            proc=subprocess.Popen(argv,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while proc.poll() is None:
                    peak=max(peak,HELP['group_rss'](proc.pid))
                    if peak>policy['rss_guard_bytes']:reason='RSS guard'
                    if time.monotonic()-began>1500:reason='controller deadline'
                    if cancelled:reason='Slurm/user termination'
                    if OLD['tree_bytes'](cache)>8<<30:reason='cache guard'
                    if reason:STOP(proc);break
                    time.sleep(2)
            finally:
                if proc.poll() is None:STOP(proc)
        r['commands'].append({'label':label,'argv':list(map(str,argv)),'seconds':time.monotonic()-start,'peak_rss_bytes':peak,'exit_code':proc.returncode,'guard':reason});save()
        if reason or (required and proc.returncode):raise RuntimeError(label+': '+str(reason or proc.returncode))
        return proc.returncode
    save()
    try:
        for label,argv in [('hostname',['hostname']),('cpu',['lscpu']),('memory',['free','-b']),('compiler',['g++','--version']),('go',[go,'version']),('job',['scontrol','show','job',job]),('os',['cat','/etc/os-release'])]:run(label,argv)
        r['perf_event_paranoid']=Path('/proc/sys/kernel/perf_event_paranoid').read_text().strip()
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        r['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.py','.sbatch') or p.name in ('go.mod','go.sum','Makefile'))}
        for name,digest in r['source_sha256'].items():
            blob=subprocess.run(['git','show',r['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists():dest.write_bytes((source/name).read_bytes())
        original=ROOT/'.cache/liftrs/parallel-s14/s14-m2-t5/online.bin';digest='88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8'
        start=time.monotonic();assert sha(original)==digest
        staged=cache/'online.bin';shutil.copy2(original,staged);assert sha(staged)==digest
        r['input']={'reference':str(original),'sha256':digest,'bytes':staged.stat().st_size,'verify_stage_seconds':time.monotonic()-start};save()
        r['cpu_affinity']=sorted(os.sched_getaffinity(0))
        r['baseline_revision']='e29258d'
        baseline=scratch/'baseline';shutil.copytree(source,baseline)
        original_cpp=subprocess.check_output(['git','show',r['baseline_revision']+':encoding/liftrs-cpp/c_api.cpp'],cwd=ROOT)
        (baseline/'encoding/liftrs-cpp/c_api.cpp').write_bytes(original_cpp)
        r['baseline_cpp_sha256']=hashlib.sha256(original_cpp).hexdigest()
        run('legacy-build',['make','-C','encoding/rme_legacy','rme'],source)
        env['CGO_ENABLED']='0';run('suite-go',[go,'test','./...','-timeout=8m'],source);env['CGO_ENABLED']='1'
        run('suite-native',[go,'test','-tags','liftrs_cpp','./...','-timeout=8m'],source)
        run('race',[go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared','-run','TestCPPTaylorParallelMatrix|TestCPPMatrixTransform','-count=1'],source)
        run('vet',[go,'vet','-tags','liftrs_cpp','./encoding/...'],source)
        unit=cache/'taylor-asan'
        run('asan-build',['g++','-std=c++17','-O1','-g','-pthread','-fsanitize=address,undefined','-fno-omit-frame-pointer','encoding/liftrs-cpp/taylor_parallel_test.cpp','-o',str(unit)],source)
        run('asan',[str(unit)],source)
        run('tsan-build',['g++','-std=c++17','-O1','-g','-pthread','-fsanitize=thread','encoding/liftrs-cpp/taylor_parallel_test.cpp','-o',str(cache/'taylor-tsan')],source)
        tsan=run('tsan',[str(cache/'taylor-tsan')],source,required=False)
        if tsan:
            log=(out/'tsan.log').read_text()
            if 'unexpected memory mapping' not in log:raise RuntimeError('ThreadSanitizer failed: inspect tsan.log')
            r['tsan']='unavailable: runtime unexpected memory mapping; not a passing race check'
        else:r['tsan']='passed'
        binaries={}
        for variant,tree in [('baseline',baseline),('candidate',source)]:
            binary=cache/('probe-'+variant);binaries[variant]=binary
            run('build-'+variant,[go,'build','-tags','liftrs_cpp','-o',str(binary),'./experiments/encoding-timing'],tree)
            run('smoke-'+variant,[str(binary),'-out',str(out/('smoke-'+variant)),'-cache',str(cache/('smoke-'+variant)),'-s','8','-workers','16','-backend','native','-repeats','1'],tree)
        # Alternate old/new cases, with a fresh process per encoding. This avoids
        # treating a previous-node profiling sample as the matched baseline.
        for repeat in range(2):
            order=['baseline','candidate'] if repeat==0 else ['candidate','baseline']
            for workers in (16,32):
                for variant in order:
                    label=f'{variant}-w{workers}-r{repeat}'
                    run(label,[str(binaries[variant]),'-out',str(out/label),'-cache',str(staged),'-s','14','-workers',str(workers),'-backend','native','-repeats','1'],source)
                    verify_case(json.loads((out/label/'phases.json').read_text()))
        for folder in sorted(out.glob('*-w*-r*')):
            variant=folder.name.split('-')[0]
            for profile in folder.glob('encode*.pprof'):
                run('pprof-'+folder.name,[go,'tool','pprof','-top','-nodecount=25',str(binaries[variant]),str(profile)])
        r['status']='passed'
    except BaseException as error:
        r['status']='stopped' if cancelled else 'failed';r['error']=repr(error);raise
    finally:
        r['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()}
        r['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','results.json.tmp','analysis.json')};save()

def verify_case(phases):
    enc=[p for p in phases if p['phase'].startswith('encode-')];checks=[p for p in phases if p['phase'].startswith('verify-')]
    assert len(enc)==len(checks)==1
    assert checks[0]['full_hash_match'] and checks[0]['canonical_sha256']=='b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c'
    p=enc[0];s=p['native_stages'];assert s['calls']==5 and s['boundary_calls']==1
    assert sum(s[k] for k in ('buffer_ns','pointwise_ns','assembly_ns','taylor_ns','columns_transposes_ns','rows_ns'))<=p['native_after']['steady_ns']-p['native_before']['steady_ns']+1_000_000
    return p

if __name__=='__main__':main()
