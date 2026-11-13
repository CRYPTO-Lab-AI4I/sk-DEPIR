#!/usr/bin/env python3
"""Explicit CPU direct packed sink, same transform arrays and bounded writer."""
import datetime,hashlib,json,os,runpy,shutil,signal,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
sha=HELP['sha']

def verify_small(rows):
    assert len(rows)==18 and [(r['case'],r['workers']) for r in rows]==[(i,w) for i in range(9) for w in (1,16)]
    for r in rows:
        st=r['stats'];n=1<<(2*r['s']);assert r['matches']==3 and r['writer_errors_recovery_owning'] and st['WorkspaceBytes']==8*n and st['WrittenBytes']==(n*r['s']+7)//8 and st['BufferBytes']<=257 and len(st['SHA256'])==64
    return rows

def verify_matched(rows):
    assert [(r['mode'],r['repeat']) for r in rows]==[(m,i) for m in ('owning','packed') for i in (0,1)]
    digest=rows[0]['packed_sha256']
    for r in rows:
        assert r['packed_sha256']==digest and r['packed_bytes']==(1<<20)*10//8 and 0<=r['pipeline_seconds']<=r['sync_close_seconds']
        if r['mode']=='packed':assert r['packed_stats']['WorkspaceBytes']==8<<20 and r['packed_stats']['SHA256']==digest
    return rows

def main():
    import argparse
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');parser.add_argument('--output-mode',required=True,choices=['packed']);args=parser.parse_args()
    if not args.name.replace('-','').isalnum():raise SystemExit('fresh alphanumeric/hyphen run name required')
    job=os.environ['SLURM_JOB_ID'];out=HERE/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/cpu-packed-output'/args.name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-cpu-packed-'+job);scratch.mkdir(exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='16',GOFLAGS='-p=4',GOMEMLIMIT=str(8<<30))
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'cache':str(cache),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'commands':[],'policy':{'output_mode':args.output_mode,'cpus':16,'host_request_bytes':16<<30,'rss_guard_bytes':12<<30,'go_soft_bytes':8<<30,'controller_seconds':1200,'slurm_seconds':1800,'new_cache_guard_bytes':16<<30,'scratch_guard_bytes':16<<30}}
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
        checkpoint=ROOT/'experiments/gpu-host-encoding/runs/online-01/results.json';old=json.loads(checkpoint.read_text());assert old['status']=='passed'
        record['component_checkpoint']={'job_id':old['job_id'],'path':str(checkpoint),'sha256':sha(checkpoint)}
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum'))}
        changed={'encoding/liftrs-cpp/c_api.cpp','encoding/liftrs-cpp/c_api.h','encoding/liftrs-cpp/plan.go','encoding/liftrs-shared/cpu_file.go','encoding/liftrs-shared/cpu_file_packed.go'}
        for name,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists():dest.write_bytes((source/name).read_bytes())
            if name.startswith('encoding/') and name not in changed and not name.endswith('_test.go'):assert old['source_sha256'][name]==digest,name
        binary=cache/'packed.test';run('build',[go,'test','-tags','liftrs_cpp','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary)
        env['CGO_ENABLED']='0';run('go-only',[go,'test','./encoding/liftrs-shared','-timeout=5m'],source);env['CGO_ENABLED']='1'
        run('cpp-checks',[go,'test','-tags','liftrs_cpp','./encoding/...','-timeout=5m'],source)
        run('combined-build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','./encoding/liftrs-shared','-run','TestCUDAAdmission|TestCUDAEncoderLifecycle','-timeout=2m'],source)
        run('vet',[go,'vet','-tags','liftrs_cpp','./encoding/...'],source)
        packed=json.loads((ROOT/'experiments/gpu-packed-output/runs/packed-02/results.json').read_text());assert packed['status']=='passed'
        fixtures=Path(packed['cache'])/'packed-tests-fixtures'
        record['input_fixtures']={}
        for degree in (2,3,4,6,8):
            fixture_path=fixtures/('s%d-code.bin'%degree);digest=sha(fixture_path);assert digest==packed['cache_sha256'][str(fixture_path.relative_to(Path(packed['cache'])))];record['input_fixtures'][str(fixture_path)]=digest
        env['PIR_CPU_PACKED_FIXTURES']=str(fixtures)
        for label in ('small','asan'):
            selected=binary
            if label=='asan':
                selected=cache/'packed-asan.test';run('asan-build',[go,'test','-asan','-tags','liftrs_cpp','-c','-o',str(selected),'./encoding/liftrs-shared'],source)
            dest=out/label;dest.mkdir();saved=cache/label;saved.mkdir();env.update(PIR_CPU_PACKED_OUT=str(dest),PIR_CPU_PACKED_CACHE=str(saved))
            run(label,[str(selected),'-test.run=^TestCPPFilePacked$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(dest/'cpu.pprof')],source)
            record[label]=verify_small(json.loads((dest/'packed-checks.json').read_text()))
        env.pop('PIR_CPU_PACKED_OUT')
        reference=ROOT/'.cache/liftrs/gpu-products/products-02/cpu-0';metadata=json.loads((reference/'online-metadata.json').read_text());assert sha(reference/'online.bin')==metadata['sha256']
        record['workflow_reference']={'path':str(reference),'online_sha256':metadata['sha256']}
        dest=out/'matched';dest.mkdir();saved=cache/'matched';saved.mkdir();env.update(PIR_CPU_PACKED_REFERENCE=str(reference),PIR_CPU_PACKED_MATCHED_OUT=str(dest),PIR_CPU_PACKED_MATCHED_CACHE=str(saved))
        run('matched',[str(binary),'-test.run=^TestCPPFilePackedCached$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(dest/'cpu.pprof')],source)
        record['matched']=verify_matched(json.loads((dest/'matched.json').read_text()))
        for label in ('owning','packed'):
            for repeat in (0,1):assert sha(saved/('%s-%d.bin'%(label,repeat)))==record['matched'][0]['packed_sha256']
        for profile in sorted(out.rglob('*.pprof')):run('pprof-'+str(profile.relative_to(out)).replace('/','-'),[go,'tool','pprof','-top','-nodecount=25',str(binary),str(profile)])
        record['status']='passed'
    except BaseException as error:record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','analysis.json')};save()

if __name__=='__main__':main()
