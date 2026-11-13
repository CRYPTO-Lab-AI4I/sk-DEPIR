#!/usr/bin/env python3
"""Complete CPU disk-backed encoding with explicit host output and file I/O."""
import datetime,hashlib,json,os,runpy,shutil,signal,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
sha=HELP['sha']

def verify_small(rows):
    assert len(rows)==9 and [r['case'] for r in rows]==list(range(9))
    for row in rows:
        st=row['stats'];n=row['n'];assert n==1<<(2*row['s']) and row['matches']==3 and row['errors_recovery_owning'] and st['calls']==5 and st['boundary_calls']==1 and st['BufferBytes']<=16*min(n,1536)+2*65536 and st['BackingBytes']==8*n and st['OutputBytes']==4*n and st['ReadBytes']>0 and st['WriteBytes']>0
    return rows

def verify_workflow(out,cache):
    result={};canonical=None;packed=None
    for mode in ('file','disk'):
        by={p['phase']:p for p in json.loads((out/'workflow'/mode/'phases.json').read_text())}
        assert by['load-online']['plan_bytes']==0 and by['online-setup']['mode']==mode
        for i in (0,1):
            check=by['verify-%d'%i];assert check['systematic_checked']==146641 and check['curves_checked']==16 and check['repeat_match']
            if canonical is None:canonical=check['canonical_sha256']
            assert check['canonical_sha256']==canonical
        assert by['verify-packed']['positions_checked']==1027 and by['verify-packed']['readback_match']
        meta=json.loads((cache/'workflow'/mode/'word-metadata.json').read_text())
        if packed is None:packed=meta['sha256']
        assert meta['sha256']==packed and meta['canonical_sha256']==canonical
        enc=[by['encode-%d'%i] for i in (0,1)]
        for row in enc:
            st=row['native_stages'];assert st['calls']==5 and st['boundary_calls']==1
            if mode=='disk':
                st=row['disk_stats'];assert st['BufferBytes']<=16*65536+2*65536 and st['BackingBytes']==8<<20 and st['OutputBytes']==4<<20 and st['ReadBytes']>0 and st['WriteBytes']>0
        result[mode]={'setup_seconds':by['online-setup']['elapsed_ns']/1e9,'load_seconds':by['load-online']['elapsed_ns']/1e9,'encode_seconds':[p['elapsed_ns']/1e9 for p in enc],'encoding':enc,'canonical_sha256':canonical,'packed_sha256':packed}
    return result

def main():
    import argparse
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');parser.add_argument('--memory-mode',required=True,choices=['disk']);args=parser.parse_args()
    if not args.name.replace('-','').isalnum():raise SystemExit('fresh alphanumeric/hyphen run name required')
    job=os.environ['SLURM_JOB_ID'];out=HERE/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/cpu-disk-encoding'/args.name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-cpu-disk-encoding-'+job);scratch.mkdir(exist_ok=False)
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
        checkpoint=ROOT/'experiments/cpu-disk-transform/runs/disk-01/results.json';old=json.loads(checkpoint.read_text());assert old['status']=='passed'
        record['component_checkpoint']={'job_id':old['job_id'],'path':str(checkpoint),'sha256':sha(checkpoint)}
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum'))}
        changed={'encoding/liftrs-cpp/c_api.cpp','encoding/liftrs-cpp/c_api.h','encoding/liftrs-cpp/plan.go','encoding/liftrs-cpp/disk.hpp','encoding/liftrs-shared/cpu_disk.go'}
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
        previous=json.loads((ROOT/'experiments/cpu-packed-output/runs/packed-01/results.json').read_text());assert previous['status']=='passed';reference=Path(previous['cache'])/'small'
        record['input_fixtures']={}
        for i in range(9):
            path=reference/('case-%d.bin'%i);digest=sha(path);assert digest==previous['cache_sha256']['small/'+path.name];record['input_fixtures'][str(path)]=digest
        env['PIR_CPU_DISK_ONLINE_FIXTURES']=str(reference)
        for label in ('online','asan'):
            selected=binary
            if label=='asan':
                selected=cache/'disk-asan.test';run('asan-build',[go,'test','-asan','-tags','liftrs_cpp','-c','-o',str(selected),'./encoding/liftrs-shared'],source)
            dest=out/label;dest.mkdir();saved=cache/label;saved.mkdir();env.update(PIR_CPU_DISK_ONLINE_OUT=str(dest),PIR_CPU_DISK_ONLINE_CACHE=str(saved))
            run(label,[str(selected),'-test.run=^TestCPPDiskOnline$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(dest/'cpu.pprof')],source)
            record[label]=verify_small(json.loads((dest/'online-checks.json').read_text()))
        env.pop('PIR_CPU_DISK_ONLINE_OUT')
        reference=ROOT/'.cache/liftrs/gpu-products/products-02/cpu-0';metadata=json.loads((reference/'online-metadata.json').read_text());assert sha(reference/'online.bin')==metadata['sha256']
        record['workflow_reference']={'path':str(reference),'online_sha256':metadata['sha256']}
        env.update(PIR_CPU_DISK_WORKFLOW_REFERENCE=str(reference),PIR_CPU_DISK_WORKFLOW_OUT=str(out/'workflow'),PIR_CPU_DISK_WORKFLOW_CACHE=str(cache/'workflow'))
        run('workflow',[str(binary),'-test.run=^TestCPPDiskOnlineWorkflow$','-test.v','-test.timeout=10m'],source);record['workflow']=verify_workflow(out,cache)
        for profile in sorted(out.rglob('*.pprof')):run('pprof-'+str(profile.relative_to(out)).replace('/','-'),[go,'tool','pprof','-top','-nodecount=25',str(binary),str(profile)])
        memory=runpy.run_path(str(ROOT/'experiments/cpu-streamed-constants/cache_memory.py'));record['job_cgroup']=memory['job_cgroup'](job);observer=memory['FileCache']()
        record['retained_file_cache_bytes']=observer.sample([p for p in cache.rglob('*.bin')]);observer.close()
        record['status']='passed'
    except BaseException as error:record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','analysis.json')};save()

if __name__=='__main__':main()
