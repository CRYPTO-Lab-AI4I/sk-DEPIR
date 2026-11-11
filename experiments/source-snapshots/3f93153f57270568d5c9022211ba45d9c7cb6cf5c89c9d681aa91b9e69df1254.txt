#!/usr/bin/env python3
"""Guarded primary CUDA s16 encoding against retained CPU output hashes."""
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
    cache=ROOT/'.cache/liftrs/cuda-s16-encoding'/name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-cuda-s16-'+job);scratch.mkdir(exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='16',GOFLAGS='-p=4')
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    policy={'cpus':16,'rss_guard_bytes':96<<30,'go_soft_bytes':64<<30,'device_cap_bytes':96<<30,'controller_seconds':6600,'worker_seconds':5400,'cache_guard_bytes':16<<30,'scratch_guard_bytes':16<<30,'minimum_available_node_bytes':32<<30,'repeats':2,'gpu':True}
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
        for label,argv in [('hostname',['hostname']),('cpu',['lscpu']),('memory',['free','-b']),('compiler',['g++','--version']),('go',[go,'version']),('job',['scontrol','show','job',job]),('os',['cat','/etc/os-release']),('gpu',['nvidia-smi','--query-gpu=name,uuid,driver_version,memory.total','--format=csv']),('nvcc',['nvcc','--version'])]:run(label,argv)
        record['cpu_affinity']=sorted(os.sched_getaffinity(0))
        visible=env['CUDA_VISIBLE_DEVICES'];assert ',' not in visible
        free=int(subprocess.check_output(['nvidia-smi','-i',visible,'--query-gpu=memory.free','--format=csv,noheader,nounits'],text=True).strip())*2**20
        record['initial_free_device_bytes']=free
        if free<98<<30:raise RuntimeError('insufficient free allocated GPU memory')
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum','Makefile'))}
        for name,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists():dest.write_bytes((source/name).read_bytes())
        admission=json.loads((ROOT/'experiments/cuda-encoding/runs/validation-01/results.json').read_text());assert admission['status']=='passed'
        for name,digest in admission['source_sha256'].items():
            if name.startswith('encoding/') and not name.endswith('_test.go'):assert record['source_sha256'].get(name)==digest,('encoding source changed',name)
        preparation=json.loads((ROOT/'experiments/gpu-preparation/runs/actual-01/results.json').read_text());assert preparation['status']=='passed' and preparation['cache_sha256']['online.bin']==EXPECTED
        reference=Path(preparation['cache']);meta=json.loads((reference/'online-metadata.json').read_text());validate_metadata(meta,(reference/'online.bin').stat().st_size)
        record['input']={'reference':str(reference),'sha256':EXPECTED,'bytes':meta['bytes'],'preparation_job':preparation['job_id']}
        cpu=json.loads((ROOT/'experiments/cached-s16-encoding/runs/encoding-01/results.json').read_text());assert cpu['status']=='passed' and cpu['input']['sha256']==EXPECTED
        record['cpu_reference']={'path':'experiments/cached-s16-encoding/runs/encoding-01/results.json','job_id':cpu['job_id'],'encoding':cpu['encoding']}
        library=cache/'libliftrs-cuda.so';env['PIR_GPU_LIBRARY']=str(library)
        run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native','-shared','-Xcompiler=-fPIC','encoding/liftrs-cuda/matrix.cu','-o',str(library)],source)
        binary=cache/'encoding.test'
        run('build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary);record['binary_sha256']=sha(binary)
        run('backend-checks',[str(binary),'-test.run=TestCUDAEncoderLifecycle|TestCUDAAdmission|TestCUDAMatrixDescriptions','-test.v','-test.timeout=5m'],source)
        smoke_reference=ROOT/'.cache/liftrs/parallel-s15/s6-m2-t4';smoke=cache/'smoke';link_inputs(smoke,smoke_reference)
        for mode,inputs in [('smoke',smoke),('encoding',cache/'encoding')]:
            if mode=='encoding':link_inputs(inputs,reference)
            phase_out=out/mode;phase_out.mkdir()
            env.update(PIR_CUDA_CACHED_OUT=str(phase_out),PIR_CUDA_CACHED_CACHE=str(inputs))
            if mode=='smoke':env['PIR_CUDA_CACHED_SMOKE']='1'
            else:env.pop('PIR_CUDA_CACHED_SMOKE',None)
            run(mode,[str(binary),'-test.run=^TestCUDACachedEncoding$','-test.v','-test.timeout='+('5m' if mode=='smoke' else '90m')],source)
            record[mode]=verify_phases(phase_out,6 if mode=='smoke' else 16)
            expected=json.loads((smoke_reference/'word-metadata.json').read_text()) if mode=='smoke' else cpu['encoding']
            packed=expected['sha256'] if mode=='smoke' else expected['packed_sha256']
            assert record[mode]['canonical_sha256']==expected['canonical_sha256'] and record[mode]['packed_sha256']==packed,'CPU/CUDA digest mismatch'
        for profile in sorted((out/'encoding').glob('*.pprof')):run('pprof-'+profile.stem,[go,'tool','pprof','-top','-nodecount=25',str(binary),str(profile)])
        record['status']='passed'
    except BaseException as error:
        record['status']='stopped' if cancelled else 'failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file() and not p.is_symlink() and not p.name.endswith('.tmp')}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','results.json.tmp','analysis.json')};save()

def verify_phases(out,s):
    phases=json.loads((out/'phases.json').read_text());p={row['phase']:row for row in phases};n,k=p['load-online']['n'],p['load-online']['k']
    assert n==1<<(2*s) and 0<k<=n//2
    canonical=None;encodings=[]
    for repeat in (0,1):
        check=p[f'verify-{repeat}'];assert check['systematic_checked']==k and check['curves_checked']==16 and check['repeat_match']
        if canonical is None:canonical=check['canonical_sha256']
        assert check['canonical_sha256']==canonical
        enc=p[f'encode-{repeat}'];stats=enc['cuda_stages'];assert stats['native_total_ns']<=stats['elapsed_ns']+1_000_000
        assert stats['owned_device_bytes']<=(512<<20 if s==6 else 96<<30)
        encodings.append({'seconds':enc['elapsed_ns']/1e9,'cuda_stages':stats})
    assert p['pack-output']['canonical_sha256']==canonical and p['pack-output']['packed_bytes']==(n*s+7)//8
    assert p['verify-packed']['positions_checked']==1027 and p['verify-packed']['readback_match']
    return {'s':s,'n':n,'k':k,'canonical_sha256':canonical,'packed_sha256':p['pack-output']['sha256'],'encodings':encodings,'phase_seconds':{row['phase']:row['elapsed_ns']/1e9 for row in phases}}

def validate_metadata(meta,size):
    assert meta['s']==16 and meta['t']==5 and meta['k']==918672421 and meta['n']==1<<32
    assert meta['identity_verified'] and meta['sha256']==EXPECTED and meta['bytes']==size==20854558905

if __name__=='__main__':main()
