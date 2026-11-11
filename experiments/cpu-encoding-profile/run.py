#!/usr/bin/env python3
"""Guarded cached s14 encoding stage/worker profile using the existing probe."""
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
    cache=ROOT/'.cache/liftrs/cpu-encoding-profile'/name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-cpu-encoding-profile-'+job);scratch.mkdir(exist_ok=False)
    policy={'workers':[8,16,32],'repeats':2,'rss_guard_bytes':12<<30,'deadline_seconds':1500,'new_cache_guard_bytes':8<<30,'backend':'native','s':14}
    r={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'policy':policy,'cache':str(cache),'commands':[]}
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='32',GOFLAGS='-p=4')
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');began=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    def save():
        p=out/'results.json.tmp';p.write_text(json.dumps(r,indent=2)+'\n');p.replace(out/'results.json')
    def run(label,argv,cwd=ROOT):
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
        if reason or proc.returncode:raise RuntimeError(label+': '+str(reason or proc.returncode))
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
        binary=cache/'probe'
        run('build',[go,'build','-tags','liftrs_cpp','-o',str(binary),'./experiments/encoding-timing'],source)
        r['binary']=str(binary);r['binary_sha256']=sha(binary)
        run('smoke',[str(binary),'-out',str(out/'smoke'),'-cache',str(cache/'smoke'),'-s','8','-workers','8','-backend','native','-repeats','1'],source)
        run('profile',[str(binary),'-out',str(out/'profile'),'-cache',str(staged),'-s','14','-workers','8,16,32','-backend','native','-repeats','2'],source)
        phases=json.loads((out/'profile/phases.json').read_text());r['encodings']=verify_phases(phases)
        for p in sorted((out/'profile').glob('encode*.pprof')):run('pprof-'+p.stem,[go,'tool','pprof','-top','-nodecount=25',str(binary),str(p)])
        r['status']='passed'
    except BaseException as error:
        r['status']='stopped' if cancelled else 'failed';r['error']=repr(error);raise
    finally:
        r['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()}
        r['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','results.json.tmp','analysis.json')};save()

def verify_phases(phases):
    enc=[p for p in phases if p['phase'].startswith('encode-')];checks=[p for p in phases if p['phase'].startswith('verify-')]
    assert len(enc)==len(checks)==6
    assert [p['workers'] for p in enc]==[8,8,16,16,32,32]
    assert all(p['full_hash_match'] and p['canonical_sha256']=='b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c' for p in checks)
    for p in enc:
        s=p['native_stages'];assert s['calls']==5 and s['boundary_calls']==1
        assert sum(s[k] for k in ('buffer_ns','pointwise_ns','assembly_ns','taylor_ns','columns_transposes_ns','rows_ns'))<=p['native_after']['steady_ns']-p['native_before']['steady_ns']+1_000_000
    return enc

if __name__=='__main__':main()
