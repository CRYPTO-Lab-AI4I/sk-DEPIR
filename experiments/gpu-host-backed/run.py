#!/usr/bin/env python3
"""Bounded-VRAM host-backed transform correctness/traffic study, not online encoding."""
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
        n,tile=row['n'],row['tile_elements']
        assert n==1<<row['dimension'] and row['matches']==3 and row['closed'] and row['recovery']
        assert row['resident_owned_bytes']-row['backed_owned_bytes']==8*(n-min(n,tile))
        assert row['backed_owned_bytes']<=512<<20
        for key in ('forward_stages','inverse_stages'):
            st=row[key];assert len(st)==12 and st[8]>=8*n and st[9]>=8*n and 0<=st[10]<=4*tile and st[11]>0 and all(x>=0 for x in st)
        assert row['forward_stages'][8:10]==row['inverse_stages'][8:10]
    return rows

def main():
    import argparse
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name');args=parser.parse_args()
    if not args.name.replace('-','').isalnum():raise SystemExit('fresh alphanumeric/hyphen run name required')
    job=os.environ['SLURM_JOB_ID'];out=HERE/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/gpu-host-backed'/args.name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-backed-'+job);scratch.mkdir(exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='16',GOFLAGS='-p=4',GOMEMLIMIT=str(8<<30))
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    record={'status':'running','job_id':job,'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'cache':str(cache),'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'commands':[],'policy':{'cpus':16,'host_request_bytes':16<<30,'rss_guard_bytes':12<<30,'go_soft_bytes':8<<30,'device_cap_bytes':512<<20,'controller_seconds':1200,'slurm_seconds':1800,'new_cache_guard_bytes':16<<30,'scratch_guard_bytes':16<<30}}
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
        checkpoint=ROOT/'experiments/gpu-packed-output/runs/packed-02/results.json';old=json.loads(checkpoint.read_text());assert old['status']=='passed'
        record['packed_checkpoint']={'job_id':old['job_id'],'path':str(checkpoint),'sha256':sha(checkpoint)}
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum'))}
        changed={'encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/backed.hpp','encoding/liftrs-cuda/gpu_api.h','encoding/liftrs-cuda/bridge.go'}
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
        run('vet',[go,'vet','-tags','liftrs_cpp,liftrs_cuda','./encoding/...'],source)
        run('race',[go,'test','-race','-tags','liftrs_cpp,liftrs_cuda','./encoding/liftrs-shared','-run','TestCUDAEncoderLifecycle|TestCUDAAdmission|TestCUDAMatrixDescriptions','-count=1','-timeout=2m'],source)
        matrix=json.loads((ROOT/'experiments/gpu-matrix/runs/matrix-01/results.json').read_text());assert matrix['status']=='passed'
        fixture=Path(matrix['cache'])/'matrix.bin';assert sha(fixture)==matrix['cache_sha256']['matrix.bin']
        record['matrix_reference']={'path':str(fixture),'sha256':sha(fixture)}
        run('matrix-driver-build',['g++','-std=c++17','-O3','-pthread','experiments/gpu-matrix/bench.cpp','encoding/liftrs-cpp/c_api.cpp','-L'+str(cache),'-l:libliftrs-cuda.so','-Wl,-rpath,'+str(cache),'-o',str(cache/'matrix-bench')],source)
        run('matrix-regression',[str(cache/'matrix-bench'),str(fixture)],source)
        matrix_rows=[json.loads(line) for line in (out/'matrix-regression.log').read_text().splitlines()];assert len(matrix_rows)==80 and all(row.get('match',True) for row in matrix_rows)
        for label,small in [('backed',False),('memcheck',True)]:
            dest=out/label;dest.mkdir();fixtures=cache/label;fixtures.mkdir();env.update(PIR_GPU_BACKED_OUT=str(dest),PIR_GPU_BACKED_CACHE=str(fixtures),PIR_GPU_BACKED_SMALL='1' if small else '0')
            argv=[str(binary),'-test.run=^TestCUDAHostBacked$','-test.v','-test.timeout=15m','-test.cpuprofile='+str(dest/'cpu.pprof')]
            if small:argv=['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99']+argv
            run(label,argv,source);record[label]=verify(json.loads((dest/'backed-checks.json').read_text()),small)
        assert 'ERROR SUMMARY: 0 errors' in (out/'memcheck.log').read_text()
        for label in ('backed','memcheck'):run('pprof-'+label,[go,'tool','pprof','-top','-nodecount=25',str(binary),str(out/label/'cpu.pprof')])
        record['status']='passed'
    except BaseException as error:record['status']='failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file()}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','analysis.json')};save()

if __name__=='__main__':main()
