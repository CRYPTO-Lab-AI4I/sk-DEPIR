#!/usr/bin/env python3
"""Guarded CPU scratch/output ownership, matched s10 and full s16 memory validation."""
import datetime,hashlib,json,os,runpy,shutil,signal,subprocess,sys,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
OLD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
sha=HELP['sha']
EXPECTED='f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874'

def link_inputs(destination,reference):
    destination.mkdir(parents=True,exist_ok=False)
    for name in ('online.bin','online-metadata.json'):(destination/name).symlink_to(reference/name)

def main():
    if len(sys.argv)!=2 or not sys.argv[1] or not sys.argv[1].replace('-','').isalnum():raise SystemExit('Usage: run.py FRESH_RUN_NAME')
    name=sys.argv[1];job=os.environ['SLURM_JOB_ID'];out=HERE/'runs'/name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/cpu-buffer-reuse'/name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-cpu-reuse-'+job);scratch.mkdir(exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='32',GOFLAGS='-p=4')
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    policy={'cpus':32,'rss_guard_bytes':68<<30,'go_soft_bytes':64<<30,'controller_seconds':6600,'worker_seconds':5400,'cache_guard_bytes':16<<30,'scratch_guard_bytes':16<<30,'minimum_available_node_bytes':32<<30,'repeats':2,'gpu':False}
    record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'policy':policy,'cache':str(cache),'commands':[]}
    def save():
        p=out/'results.json.tmp';p.write_text(json.dumps(record,indent=2)+'\n');p.replace(out/'results.json')
    def run(label,argv,cwd=ROOT):
        began=time.monotonic();peak=0;reason=None;last_disk=0;sizes=(0,0)
        with (out/(label+'.log')).open('w') as log,(out/'progress.jsonl').open('a') as progress:
            proc=subprocess.Popen(argv,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while proc.poll() is None:
                    now=time.monotonic();peak=max(peak,HELP['group_rss'](proc.pid))
                    if now-last_disk>30:sizes=OLD['tree_bytes'](cache),OLD['tree_bytes'](scratch);last_disk=now
                    if peak>policy['rss_guard_bytes']:reason='RSS guard'
                    if max(sizes)>16<<30:reason='new cache/scratch guard'
                    if OLD['mem_available']()<policy['minimum_available_node_bytes']:reason='node available-memory guard'
                    if now-start>6600:reason='controller deadline'
                    if cancelled:reason='Slurm/user termination'
                    state={}
                    try:state=json.loads((out/label/'state.json').read_text())
                    except (FileNotFoundError,json.JSONDecodeError):pass
                    progress.write(json.dumps({'utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'command':label,'rss_bytes':peak,'cache_bytes':sizes[0],'scratch_bytes':sizes[1],'worker':state,'guard':reason})+'\n');progress.flush()
                    if reason:OLD['stop_group'](proc);break
                    time.sleep(2)
            finally:
                if proc.poll() is None:OLD['stop_group'](proc)
        record['commands'].append({'label':label,'argv':list(map(str,argv)),'seconds':time.monotonic()-began,'peak_rss_bytes':peak,'exit_code':proc.returncode,'guard':reason});save()
        if reason or proc.returncode:raise RuntimeError(label+': '+str(reason or proc.returncode))
    save()
    try:
        if OLD['mem_available']()<128<<30:raise RuntimeError('insufficient initially available node memory')
        if shutil.disk_usage(cache).free<20<<30 or shutil.disk_usage(scratch).free<20<<30:raise RuntimeError('insufficient cache/scratch space')
        for label,argv in [('hostname',['hostname']),('cpu',['lscpu']),('memory',['free','-b']),('compiler',['g++','--version']),('go',[go,'version']),('job',['scontrol','show','job',job]),('os',['cat','/etc/os-release'])]:run(label,argv)
        record['cpu_affinity']=sorted(os.sched_getaffinity(0))
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum','Makefile'))}
        for name,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists():dest.write_bytes((source/name).read_bytes())
        admission=json.loads((ROOT/'experiments/cuda-buffer-reuse/runs/reuse-01/results.json').read_text());assert admission['status']=='passed'
        changed={'encoding/liftrs-cpp/plan.go','encoding/liftrs-cpp/c_api.cpp','encoding/liftrs-cpp/c_api.h','encoding/liftrs-cpp/online_test.cpp','encoding/liftrs-shared/scaling_matrix_cpp.go'}
        record['changed_source_paths']=sorted(changed)
        for name,digest in admission['source_sha256'].items():
            if name.startswith('encoding/') and name not in changed and not name.endswith('_test.go'):assert record['source_sha256'].get(name)==digest,('encoding source changed',name)
        preparation=json.loads((ROOT/'experiments/gpu-preparation/runs/actual-01/results.json').read_text());assert preparation['status']=='passed' and preparation['cache_sha256']['online.bin']==EXPECTED
        reference=Path(preparation['cache']);meta=json.loads((reference/'online-metadata.json').read_text());validate_metadata(meta,(reference/'online.bin').stat().st_size)
        record['input']={'reference':str(reference),'sha256':EXPECTED,'bytes':meta['bytes'],'preparation_job':preparation['job_id']}
        binary=cache/'encoding.test'
        run('build',[go,'test','-tags','liftrs_cpp','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary);record['binary_sha256']=sha(binary)
        run('native-checks',[str(binary),'-test.run=TestCPPOutputOwnership|TestCPPOnlineOverlap|TestCPPOnlineNative|TestCPPMatrixEncoding|TestCPPMatrixConcurrent|TestCPPTaylorParallelMatrix','-test.v','-test.timeout=5m'],source)
        run('native-suite',[go,'test','-tags','liftrs_cpp','./encoding/...','-timeout=5m'],source)
        env['CGO_ENABLED']='0';run('go-only-checks',[go,'test','./encoding/liftrs-shared','-timeout=5m'],source);env['CGO_ENABLED']='1'
        run('race',[go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared','-run','TestCPPOutputOwnership|TestCPPMatrixConcurrent|TestCPPOnlineOverlap','-count=1'],source)
        run('vet',[go,'vet','-tags','liftrs_cpp','./encoding/...'],source)
        sanitizer=cache/'online-asan'
        run('sanitizer-build',['g++','-std=c++17','-O1','-g','-pthread','-fno-omit-frame-pointer','-fsanitize=address,undefined','encoding/liftrs-cpp/online_test.cpp','encoding/liftrs-cpp/c_api.cpp','-o',str(sanitizer)],source)
        env['ASAN_OPTIONS']='detect_leaks=1:halt_on_error=1';env['UBSAN_OPTIONS']='halt_on_error=1'
        run('sanitizer',[str(sanitizer)],source)
        reference10=ROOT/'.cache/liftrs/gpu-products/products-02/cpu-0'
        env['PIR_CPU_MEMORY_SMALL']='1'
        for mode in ('matched-baseline','matched-reuse'):
            inputs=cache/mode;link_inputs(inputs,reference10);dest=out/mode;dest.mkdir()
            env.update(PIR_CPU_MEMORY_BASELINE='1' if mode=='matched-baseline' else '0',PIR_CPU_MEMORY_OUT=str(dest),PIR_CPU_MEMORY_CACHE=str(inputs))
            run(mode,[str(binary),'-test.run=^TestCPUMemoryCached$','-test.v','-test.timeout=5m'],source)
            record[mode]=verify_phases(dest,10,mode=='matched-reuse')
        assert record['matched-baseline']['canonical_sha256']==record['matched-reuse']['canonical_sha256']
        assert record['matched-baseline']['packed_sha256']==record['matched-reuse']['packed_sha256']
        env.pop('PIR_CPU_MEMORY_SMALL',None);env['PIR_CPU_MEMORY_BASELINE']='0'
        previous=json.loads((ROOT/'experiments/cached-s16-encoding/runs/encoding-01/results.json').read_text());assert previous['status']=='passed'
        record['cpu_reference']={'path':'experiments/cached-s16-encoding/runs/encoding-01/results.json','encoding':previous['encoding'],'job_id':previous['job_id']}
        smoke_reference=ROOT/'.cache/liftrs/parallel-s15/s6-m2-t4';smoke=cache/'smoke';link_inputs(smoke,smoke_reference)
        for mode,inputs in [('smoke',smoke),('encoding',cache/'encoding')]:
            if mode=='encoding':link_inputs(inputs,reference)
            phase_out=out/mode;phase_out.mkdir()
            env.update(PIR_CPU_MEMORY_OUT=str(phase_out),PIR_CPU_MEMORY_CACHE=str(inputs))
            if mode=='smoke':env['PIR_CPU_MEMORY_SMOKE']='1'
            else:env.pop('PIR_CPU_MEMORY_SMOKE',None)
            run(mode,[str(binary),'-test.run=^TestCPUMemoryCached$','-test.v','-test.timeout='+('5m' if mode=='smoke' else '90m')],source)
            record[mode]=verify_phases(phase_out,6 if mode=='smoke' else 16,True)
            if mode=='smoke':
                expected=json.loads((smoke_reference/'word-metadata.json').read_text());assert record[mode]['canonical_sha256']==expected['canonical_sha256'] and record[mode]['packed_sha256']==expected['sha256']
        for mode in ('matched-baseline','matched-reuse','encoding'):
            for profile in sorted((out/mode).glob('*.pprof')):run('pprof-'+mode+'-'+profile.stem,[go,'tool','pprof','-top','-nodecount=25',str(binary),str(profile)])
        for key in ('canonical_sha256','packed_sha256'):assert record['encoding'][key]==previous['encoding'][key]
        record['principal_memory_bytes']=memory_model(918672421,1<<32)
        record['status']='passed'
    except BaseException as error:
        record['status']='stopped' if cancelled else 'failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file() and not p.is_symlink() and not p.name.endswith('.tmp')}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','results.json.tmp','analysis.json')};save()

def validate_metadata(meta,size):
    assert meta['s']==16 and meta['t']==5 and meta['k']==918672421 and meta['n']==1<<32
    assert meta['identity_verified'] and meta['sha256']==EXPECTED and meta['bytes']==size==20854558905


def memory_model(k,n):
    old=16*n+12*k+4
    return {'old_principal_bytes':old,'reuse_principal_bytes':old-4*n,'saved_bytes':4*n}

def verify_phases(out,s,reuse):
    phases=json.loads((out/'phases.json').read_text());p={row['phase']:row for row in phases};n,k=p['load-online']['n'],p['load-online']['k']
    assert n==1<<(2*s) and 0<k<=n//2
    canonical=None;encodings=[]
    for repeat in (0,1):
        check=p[f'verify-{repeat}'];assert check['systematic_checked']==k and check['curves_checked']==16 and check['repeat_match']
        if canonical is None:canonical=check['canonical_sha256']
        assert canonical==check['canonical_sha256']
        enc=p[f'encode-{repeat}'];st=enc['native_stages'];assert st['calls']==5 and st['boundary_calls']==1
        assert enc['scratch_transferred']==reuse
        if reuse:
            assert enc['workspace_after_bytes']==(max(2*(1<<(2*k-2).bit_length()),n)+k)*4
            if repeat==0:assert enc['workspace_before_bytes']-enc['workspace_after_bytes']==4*n
            else:assert enc['workspace_before_bytes']==enc['workspace_after_bytes']
        else:assert enc['workspace_before_bytes']==enc['workspace_after_bytes']
        total=sum(st[key] for key in ('taylor_ns','columns_transposes_ns','rows_ns','buffer_ns','pointwise_ns','assembly_ns'))
        assert total<=enc['native_after']['steady_ns']-enc['native_before']['steady_ns']+1_000_000
        encodings.append({'seconds':enc['elapsed_ns']/1e9,'workspace_after_bytes':enc['workspace_after_bytes'],'scratch_transferred':enc['scratch_transferred']})
    assert p['pack-output']['canonical_sha256']==canonical and p['pack-output']['packed_bytes']==(n*s+7)//8
    assert p['verify-packed']['positions_checked']==1027 and p['verify-packed']['readback_match']
    return {'s':s,'n':n,'k':k,'canonical_sha256':canonical,'packed_sha256':p['pack-output']['sha256'],'encodings':encodings,'phase_seconds':{row['phase']:row['elapsed_ns']/1e9 for row in phases}}

if __name__=='__main__':main()
