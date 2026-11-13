#!/usr/bin/env python3
"""Complete host-backed GPU encoding, explicit low-device-memory mode."""
import datetime,hashlib,json,os,runpy,shutil,signal,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
HELP=runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
GUARD=runpy.run_path(str(ROOT/'experiments/hpc-s16/run.py'))
sha=HELP['sha']

def verify_host(rows):
    rejected=[r for r in rows if r.get('rejected_prefix')];success=[r for r in rows if not r.get('rejected_prefix')]
    assert len(rows)==13 and [r['s'] for r in rejected]==[2,3,4]
    assert [(r['s'],r['tile']) for r in success]==[(s,t) for s in (2,3,4,6,8) for t in (256,1536)]
    for row in success:
        st=row['stats'];assert row['matches']==3 and row['errors_recovery_close'] and st['transforms']==5 and st['owned_device_bytes']<512<<20 and st['host_arena_bytes']==4*(1<<(2*row['s'])) and st['host_stage_bytes']<=4*row['tile'] and st['host_to_device_bytes']>0 and st['device_to_host_bytes']>0
    return rows

def verify_workflow(out,cache):
    result={};canonical=None;packed=None
    for backend in ('cpu','cuda'):
        by={p['phase']:p for p in json.loads((out/'workflow'/backend/'phases.json').read_text())}
        for i in (0,1):
            check=by['verify-%d'%i];assert check['systematic_checked']>0 and check['curves_checked']==16 and check['repeat_match']
            if canonical is None:canonical=check['canonical_sha256']
            assert check['canonical_sha256']==canonical
        assert by['verify-packed']['positions_checked']==1027 and by['verify-packed']['readback_match']
        meta=json.loads((cache/'workflow'/backend/'word-metadata.json').read_text())
        if packed is None:packed=meta['sha256']
        assert meta['sha256']==packed and meta['canonical_sha256']==canonical
        enc=[by['encode-%d'%i] for i in (0,1)]
        if backend=='cuda':
            assert by['online-setup']['backend']=='cuda-host-backed'
            for row in enc:
                st=row['cuda_stages'];assert st['transforms']==5 and st['host_arena_bytes']==4<<20 and st['host_stage_bytes']<=4*65536 and st['owned_device_bytes']<512<<20 and st['host_to_device_bytes']>0 and st['device_to_host_bytes']>0
        result[backend]={'setup_seconds':by['online-setup']['elapsed_ns']/1e9,'encode_seconds':[p['elapsed_ns']/1e9 for p in enc],'encoding':enc,'canonical_sha256':canonical,'packed_sha256':packed}
    return result

def main():
    import argparse
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');parser.add_argument('--memory-mode',required=True,choices=['host-backed']);args=parser.parse_args()
    if not args.name.replace('-','').isalnum():raise SystemExit('fresh alphanumeric/hyphen run name required')
    job=os.environ['SLURM_JOB_ID'];out=HERE/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/gpu-host-encoding'/args.name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-host-encoding-'+job);scratch.mkdir(exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='16',GOFLAGS='-p=4',GOMEMLIMIT=str(8<<30))
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'cache':str(cache),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'commands':[],'policy':{'memory_mode':args.memory_mode,'cpus':16,'host_request_bytes':16<<30,'rss_guard_bytes':12<<30,'go_soft_bytes':8<<30,'device_cap_bytes':512<<20,'controller_seconds':1200,'slurm_seconds':1800,'new_cache_guard_bytes':16<<30,'scratch_guard_bytes':16<<30}}
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
        for label,argv in [('hostname',['hostname']),('cpu',['lscpu']),('compiler',['g++','--version']),('go',[go,'version']),('gpu',['nvidia-smi','--query-gpu=name,uuid,driver_version,memory.total','--format=csv']),('nvcc',['nvcc','--version']),('job',['scontrol','show','job',job])]:run(label,argv)
        visible=env['CUDA_VISIBLE_DEVICES'];assert ',' not in visible
        free=int(subprocess.check_output(['nvidia-smi','-i',visible,'--query-gpu=memory.free','--format=csv,noheader,nounits'],text=True).strip())<<20
        assert free>=1<<30;record['initial_free_device_bytes']=free
        checkpoint=ROOT/'experiments/gpu-backed-parallel/runs/parallel-01/results.json';old=json.loads(checkpoint.read_text());assert old['status']=='passed'
        record['component_checkpoint']={'job_id':old['job_id'],'path':str(checkpoint),'sha256':sha(checkpoint)}
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum'))}
        changed={'encoding/liftrs-cuda/backed.hpp','encoding/liftrs-cuda/bridge.go','encoding/liftrs-cuda/gpu_api.h','encoding/liftrs-shared/cuda_encoder.go','encoding/liftrs-shared/cuda_available.go','encoding/liftrs-shared/cuda_host.go','encoding/liftrs-shared/cuda_host_available.go','encoding/liftrs-shared/cuda_host_unavailable.go'}
        for name,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists():dest.write_bytes((source/name).read_bytes())
            if name.startswith('encoding/') and name not in changed and not name.endswith('_test.go'):assert old['source_sha256'][name]==digest,name
        library=cache/'libliftrs-cuda.so';env['PIR_GPU_LIBRARY']=str(library)
        run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-arch=native','-shared','-Xcompiler=-fPIC','encoding/liftrs-cuda/matrix.cu','-o',str(library)],source)
        binary=cache/'backed.test';run('build',[go,'test','-tags','liftrs_cpp,liftrs_cuda','-c','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary)
        run('static-checks',[str(binary),'-test.run=TestCUDAAdmission|TestCUDAEncoderLifecycle|TestCUDAMatrixDescriptions|TestPackedFileAccessor','-test.timeout=2m'],source)
        run('build-compatibility',[go,'test','-tags','liftrs_cuda','./encoding/liftrs-shared','-run','TestCUDAAdmission|TestCUDAEncoderLifecycle','-timeout=2m'],source)
        env['CGO_ENABLED']='0';run('go-only',[go,'test','./encoding/liftrs-shared','-timeout=5m'],source);env['CGO_ENABLED']='1'
        run('combined-checks',[go,'test','-tags','liftrs_cpp,liftrs_cuda','./encoding/...','-timeout=5m'],source)
        run('vet',[go,'vet','-tags','liftrs_cpp,liftrs_cuda','./encoding/...'],source)
        run('race',[go,'test','-race','-tags','liftrs_cpp,liftrs_cuda','./encoding/liftrs-shared','-run','TestCUDAEncoderLifecycle|TestCUDAAdmission|TestCUDAMatrixDescriptions','-count=1','-timeout=2m'],source)
        matrix=json.loads((ROOT/'experiments/gpu-matrix/runs/matrix-01/results.json').read_text());assert matrix['status']=='passed'
        fixture=Path(matrix['cache'])/'matrix.bin';assert sha(fixture)==matrix['cache_sha256']['matrix.bin']
        record['matrix_reference']={'path':str(fixture),'sha256':sha(fixture)}
        run('matrix-driver-build',['g++','-std=c++17','-O3','-pthread','experiments/gpu-matrix/bench.cpp','encoding/liftrs-cpp/c_api.cpp','-L'+str(cache),'-l:libliftrs-cuda.so','-Wl,-rpath,'+str(cache),'-o',str(cache/'matrix-bench')],source)
        run('matrix-regression',[str(cache/'matrix-bench'),str(fixture)],source)
        matrix_rows=[json.loads(line) for line in (out/'matrix-regression.log').read_text().splitlines()];assert len(matrix_rows)==80 and all(row.get('match',True) for row in matrix_rows)
        record['policy']['memory_mode']=args.memory_mode
        packed=json.loads((ROOT/'experiments/gpu-packed-output/runs/packed-02/results.json').read_text());assert packed['status']=='passed'
        fixtures=Path(packed['cache'])/'packed-tests-fixtures'
        record['input_fixtures']={}
        for degree in (2,3,4,6,8):
            fixture_path=fixtures/('s%d-code.bin'%degree);digest=sha(fixture_path);assert digest==packed['cache_sha256'][str(fixture_path.relative_to(Path(packed['cache'])))];record['input_fixtures'][str(fixture_path)]=digest
        old_library=Path(old['cache'])/'libliftrs-cuda.so';assert sha(old_library)==old['cache_sha256']['libliftrs-cuda.so']
        record['old_library']={'path':str(old_library),'sha256':sha(old_library)}
        env.update(PIR_CUDA_HOST_FIXTURES=str(fixtures),PIR_CUDA_HOST_OLD_LIBRARY=str(old_library))
        for label in ('host-checks','memcheck'):
            dest=out/label;dest.mkdir();env['PIR_CUDA_HOST_OUT']=str(dest)
            argv=[str(binary),'-test.run=^TestCUDAHostOnline$','-test.v','-test.timeout=15m','-test.cpuprofile='+str(dest/'cpu.pprof')]
            if label=='memcheck':argv=['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99']+argv
            run(label,argv,source);record[label]=verify_host(json.loads((dest/'host-checks.json').read_text()))
        assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text();env.pop('PIR_CUDA_HOST_OUT')
        reference=ROOT/'.cache/liftrs/gpu-products/products-02/cpu-0';metadata=json.loads((reference/'online-metadata.json').read_text());assert sha(reference/'online.bin')==metadata['sha256']
        record['workflow_reference']={'path':str(reference),'online_sha256':metadata['sha256']}
        env.update(PIR_CUDA_HOST_WORKFLOW='1',PIR_CUDA_WORKFLOW_REFERENCE=str(reference),PIR_CUDA_WORKFLOW_OUT=str(out/'workflow'),PIR_CUDA_WORKFLOW_CACHE=str(cache/'workflow'))
        run('workflow',[str(binary),'-test.run=^TestCUDAEncodingWorkflow$','-test.v','-test.timeout=15m'],source)
        record['workflow']=verify_workflow(out,cache)
        for profile in sorted(out.rglob('*.pprof')):run('pprof-'+str(profile.relative_to(out)).replace('/','-'),[go,'tool','pprof','-top','-nodecount=25',str(binary),str(profile)])
        record['status']='passed'
    except BaseException as error:record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','analysis.json')};save()

if __name__=='__main__':main()
