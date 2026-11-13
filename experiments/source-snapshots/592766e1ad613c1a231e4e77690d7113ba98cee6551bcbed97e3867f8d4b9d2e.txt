#!/usr/bin/env python3
"""Bounded private-RAM disk transform component, explicit repeated file I/O."""
import datetime,hashlib,json,os,runpy,shutil,signal,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
sha=HELP['sha']

def verify(rows,small=False):
    assert len(rows)==(6 if small else 8)
    for row in rows:
        n,tile=row['n'],row['tile_elements'];assert n==1<<row['dimension'] and row['matches']==3 and row['errors_recovery_canaries'] and row['backing_bytes']==8*n+64
        for key in ('forward','inverse'):
            st=row[key];assert st['buffer_bytes']<=8*min(n,tile)+65536 and st['transposes']==2 and st['read_bytes']>=16*n and st['write_bytes']>=16*n and st['read_calls']>0 and st['write_calls']>0
        assert row['forward']['read_bytes']==row['inverse']['read_bytes'] and row['forward']['write_bytes']==row['inverse']['write_bytes']
    return rows

def main():
    import argparse
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');parser.add_argument('--memory-mode',required=True,choices=['disk']);args=parser.parse_args()
    if not args.name.replace('-','').isalnum():raise SystemExit('fresh alphanumeric/hyphen run name required')
    job=os.environ['SLURM_JOB_ID'];out=HERE/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/cpu-disk-transform'/args.name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-cpu-disk-'+job);scratch.mkdir(exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='16',GOFLAGS='-p=4',GOMEMLIMIT=str(8<<30))
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'cache':str(cache),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'commands':[],'policy':{'memory_mode':args.memory_mode,'cpus':16,'host_request_bytes':16<<30,'rss_guard_bytes':12<<30,'go_soft_bytes':8<<30,'controller_seconds':1200,'slurm_seconds':1800,'new_cache_guard_bytes':16<<30,'scratch_guard_bytes':16<<30}}
    def save():(out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
    def run(label,argv,cwd=ROOT):
        began=time.monotonic();peak=0;reason=None
        with (out/(label+'.log')).open('w') as log:
            proc=subprocess.Popen(argv,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while proc.poll() is None:
                    peak=max(peak,HELP['group_rss'](proc.pid))
                    if peak>12<<30:reason='RSS guard'
                    if GUARD['tree_bytes'](cache)>16<<30 or GUARD['tree_bytes'](scratch)>16<<30:reason='cache/scratch guard'
                    if GUARD['mem_available']()<16<<30:reason='node headroom'
                    if time.monotonic()-start>1200:reason='controller deadline'
                    if cancelled:reason='termination'
                    if reason:GUARD['stop_group'](proc);break
                    time.sleep(1)
            finally:
                if proc.poll() is None:GUARD['stop_group'](proc)
        record['commands'].append({'label':label,'argv':list(map(str,argv)),'seconds':time.monotonic()-began,'peak_rss_bytes':peak,'guard':reason,'exit_code':proc.returncode});save()
        if reason or proc.returncode:raise RuntimeError(label+': '+str(reason or proc.returncode))
    save()
    try:
        if GUARD['mem_available']()<32<<30 or shutil.disk_usage(cache).free<20<<30 or shutil.disk_usage(scratch).free<20<<30:raise RuntimeError('initial RAM/disk guard')
        for label,argv in [('hostname',['hostname']),('cpu',['lscpu']),('compiler',['g++','--version']),('go',[go,'version']),('job',['scontrol','show','job',job])]:run(label,argv)
        checkpoint=ROOT/'experiments/cpu-packed-output/runs/packed-01/results.json';old=json.loads(checkpoint.read_text());assert old['status']=='passed'
        record['component_checkpoint']={'job_id':old['job_id'],'path':str(checkpoint),'sha256':sha(checkpoint)}
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum'))}
        changed={'encoding/liftrs-cpp/c_api.cpp','encoding/liftrs-cpp/c_api.h','encoding/liftrs-cpp/plan.go','encoding/liftrs-cpp/disk.hpp'}
        for name,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists():dest.write_bytes((source/name).read_bytes())
            if name.startswith('encoding/') and name not in changed and not name.endswith('_test.go'):assert old['source_sha256'][name]==digest,name
        binary=cache/'disk.test';run('build',[go,'test','-tags','liftrs_cpp','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary)
        env['CGO_ENABLED']='0';run('go-only',[go,'test','./encoding/liftrs-shared','-timeout=5m'],source);env['CGO_ENABLED']='1'
        run('cpp-checks',[go,'test','-tags','liftrs_cpp','./encoding/...','-timeout=5m'],source)
        run('combined-build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','./encoding/liftrs-shared','-run','TestCUDAAdmission|TestCUDAEncoderLifecycle','-timeout=2m'],source)
        run('vet',[go,'vet','-tags','liftrs_cpp','./encoding/...'],source)
        run('filesystem',['df','-T',str(cache)])
        previous=json.loads((ROOT/'experiments/gpu-host-backed/runs/backed-01/results.json').read_text());assert previous['status']=='passed'
        reference=Path(previous['cache'])/'backed';record['reference']={'path':str(reference),'job_id':previous['job_id']}
        record['input_fixtures']={}
        for row in previous['backed']:
            for suffix in ('input','forward'):
                name='s%d-d%d-%s.bin'%(row['s'],row['dimension'],suffix);path=reference/name;digest=sha(path);assert digest==previous['cache_sha256']['backed/'+name];record['input_fixtures'][str(path)]=digest
        env['PIR_CPU_DISK_REFERENCE']=str(reference)
        for label,small in [('disk',False),('asan',True)]:
            selected=binary
            if small:
                selected=cache/'disk-asan.test';run('asan-build',[go,'test','-asan','-tags','liftrs_cpp','-c','-o',str(selected),'./encoding/liftrs-shared'],source)
            dest=out/label;dest.mkdir();saved=cache/label;saved.mkdir();env.update(PIR_CPU_DISK_OUT=str(dest),PIR_CPU_DISK_CACHE=str(saved),PIR_CPU_DISK_SMALL='1' if small else '0')
            run(label,[str(selected),'-test.run=^TestCPPDiskTransform$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(dest/'cpu.pprof')],source)
            record[label]=verify(json.loads((dest/'disk-checks.json').read_text()),small)
            for row in record[label]:
                path=reference/('s%d-d%d-forward.bin'%(row['s'],row['dimension']));assert sha(path)==row['forward_sha256']
        for profile in sorted(out.rglob('*.pprof')):run('pprof-'+str(profile.relative_to(out)).replace('/','-'),[go,'tool','pprof','-top','-nodecount=25',str(binary),str(profile)])
        memory=runpy.run_path(str(ROOT/'experiments/cpu-streamed-constants/cache_memory.py'));record['job_cgroup']=memory['job_cgroup'](job);observer=memory['FileCache']()
        record['retained_file_cache_bytes']=observer.sample([p for p in cache.rglob('*.bin')]);observer.close()
        record['status']='passed'
    except BaseException as error:record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','analysis.json')};save()

if __name__=='__main__':main()
