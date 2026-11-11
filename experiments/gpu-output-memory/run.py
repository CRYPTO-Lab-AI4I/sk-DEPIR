#!/usr/bin/env python3
"""Validate explicit compact GPU output: bounded compaction, assembly and owning host output."""
import argparse
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
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('name')
    parser.add_argument('--memory-mode',choices=('resident','compact'),default='resident',help='compact explicitly opts into extra compaction/expansion work')
    args=parser.parse_args();name=args.name
    if not name or not name.replace('-','').isalnum():raise SystemExit('Use a fresh alphanumeric/hyphen name')
    job=os.environ['SLURM_JOB_ID'];out=HERE/'runs'/name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/gpu-output-memory'/name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('sk-depir-gpu-output-'+job);scratch.mkdir(exist_ok=False)
    env=dict(os.environ,GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='1',CXX='g++',GOMAXPROCS='16',GOFLAGS='-p=4')
    go=str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go');start=time.monotonic();cancelled=[];signal.signal(signal.SIGTERM,lambda *_:cancelled.append(True))
    policy={'cpus':16,'rss_guard_bytes':96<<30,'go_soft_bytes':64<<30,'device_cap_bytes':(40 if args.memory_mode=='compact' else 56)<<30,'memory_mode':args.memory_mode,'controller_seconds':6600,'worker_seconds':5400,'cache_guard_bytes':16<<30,'scratch_guard_bytes':16<<30,'minimum_available_node_bytes':32<<30,'repeats':2,'gpu':True}
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
        if free<policy['device_cap_bytes']+(2<<30):raise RuntimeError('insufficient free allocated GPU memory')
        source=scratch/'repo';shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__','slurm-*.log'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*') if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.cu','.py','.sbatch') or p.name in ('go.mod','go.sum','Makefile'))}
        for name,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                dest=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not dest.exists():dest.write_bytes((source/name).read_bytes())
        admission=json.loads((ROOT/'experiments/gpu-transform-memory/runs/rows-01/results.json').read_text());assert admission['status']=='passed'
        changed_core={'encoding/liftrs-cuda/matrix.cu','encoding/liftrs-cuda/gpu_api.h','encoding/liftrs-cuda/bridge.go','encoding/liftrs-shared/cuda_available.go'}
        record['changed_core_paths']=sorted(changed_core)
        for name,digest in admission['source_sha256'].items():
            if name.startswith('encoding/') and name not in changed_core and not name.endswith('_test.go'):assert record['source_sha256'].get(name)==digest,('encoding source changed',name)
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
        run('cpu-build-checks',[go,'test','./encoding/liftrs-shared','-timeout=5m'],source)
        run('cuda-build-checks',[go,'test','-tags','liftrs_cuda','./encoding/liftrs-shared','-timeout=5m'],source)
        run('combined-checks',[go,'test','-tags','liftrs_cpp,liftrs_cuda','./encoding/...','-timeout=5m'],source)
        run('vet',[go,'vet','-tags','liftrs_cpp,liftrs_cuda','./encoding/...'],source)
        env['CGO_ENABLED']='0'
        run('go-only-checks',[go,'test','./encoding/liftrs-shared','-timeout=5m'],source)
        env['CGO_ENABLED']='1'
        matrix=json.loads((ROOT/'experiments/gpu-matrix/runs/matrix-01/results.json').read_text());assert matrix['status']=='passed'
        fixture=Path(matrix['cache'])/'matrix.bin';assert sha(fixture)==matrix['cache_sha256']['matrix.bin']
        record['matrix_reference']={'path':str(fixture),'sha256':sha(fixture)}
        run('matrix-driver-build',['g++','-std=c++17','-O3','-pthread','experiments/gpu-matrix/bench.cpp','encoding/liftrs-cpp/c_api.cpp','-L'+str(cache),'-l:libliftrs-cuda.so','-Wl,-rpath,'+str(cache),'-o',str(cache/'matrix-bench')],source)
        run('matrix-regression',[str(cache/'matrix-bench'),str(fixture)],source)
        old_manifest=json.loads((ROOT/'experiments/cuda-buffer-reuse/runs/reuse-01/results.json').read_text());assert old_manifest['status']=='passed'
        old_library=Path(old_manifest['cache'])/'libliftrs-cuda.so'
        assert sha(old_library)==old_manifest['cache_sha256']['libliftrs-cuda.so']
        record['old_reuse_library']={'path':str(old_library),'sha256':sha(old_library)}
        env['PIR_CUDA_REUSE']='1'
        for mode in ('baseline','candidate','memcheck'):
            dest=out/mode;dest.mkdir();fixtures=cache/(mode+'-fixtures');fixtures.mkdir()
            env.update(PIR_CUDA_REUSE='1',PIR_CUDA_COMPACT='1' if args.memory_mode=='compact' and mode!='baseline' else '0',PIR_CUDA_ONLINE_OUT=str(dest),PIR_CUDA_ONLINE_CACHE=str(fixtures))
            if mode!='baseline' and args.memory_mode=='compact':env['PIR_OLD_ONLINE_LIBRARY']=str(old_library)
            argv=[str(binary),'-test.run=^TestCUDAOnlineEncoding$','-test.v','-test.timeout=10m']
            if mode=='memcheck':argv=['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99']+argv
            run(mode,argv,source)
        if 'ERROR SUMMARY: 0 errors' not in (out/'memcheck.log').read_text():raise RuntimeError('memcheck failed')
        env.pop('PIR_CUDA_ONLINE_OUT',None);env.pop('PIR_OLD_ONLINE_LIBRARY',None)
        record['small_checks']=verify_small(out,args.memory_mode=='compact')
        matrix_rows=[json.loads(line) for line in (out/'matrix-regression.log').read_text().splitlines()]
        assert len(matrix_rows)==80 and all(row.get('match',True) for row in matrix_rows)
        if args.memory_mode=='compact':
            env['PIR_CUDA_COMPACT']='1'
            for mode in ('boundaries','boundary-memcheck'):
                dest=out/mode;dest.mkdir();fixtures=cache/(mode+'-fixtures');fixtures.mkdir();env.update(PIR_CUDA_COMPACT_BOUNDARY_OUT=str(dest),PIR_CUDA_COMPACT_BOUNDARY_CACHE=str(fixtures))
                argv=[str(binary),'-test.run=^TestCUDACompactBoundaries$','-test.v','-test.timeout=10m','-test.cpuprofile='+str(dest/'cpu.pprof')]
                if mode=='boundary-memcheck':argv=['compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99']+argv
                run(mode,argv,source)
                records=json.loads((dest/'boundaries.json').read_text());assert [r['s'] for r in records]==[6,10,11] and all(r['matches']==3 for r in records) and records[-1]['invalid_truncation_rejected']
            assert 'ERROR SUMMARY: 0 errors' in (out/'boundary-memcheck.log').read_text()
            env.pop('PIR_CUDA_COMPACT_BOUNDARY_OUT',None)
        reference10=ROOT/'.cache/liftrs/gpu-products/products-02/cpu-0'
        env['PIR_CUDA_CACHED_S10']='1'
        for mode in ('matched-baseline','matched-candidate'):
            inputs=cache/mode;link_inputs(inputs,reference10);dest=out/mode;dest.mkdir()
            env.update(PIR_CUDA_REUSE='1',PIR_CUDA_COMPACT='1' if args.memory_mode=='compact' and mode=='matched-candidate' else '0',PIR_CUDA_CACHED_OUT=str(dest),PIR_CUDA_CACHED_CACHE=str(inputs))
            run(mode,[str(binary),'-test.run=^TestCUDACachedEncoding$','-test.v','-test.timeout=5m'],source)
            record[mode]=verify_phases(dest,10,args.memory_mode=='compact' and mode=='matched-candidate')
        assert record['matched-baseline']['canonical_sha256']==record['matched-candidate']['canonical_sha256']
        assert record['matched-baseline']['packed_sha256']==record['matched-candidate']['packed_sha256']
        env.pop('PIR_CUDA_CACHED_S10',None);env['PIR_CUDA_REUSE']='1';env['PIR_CUDA_COMPACT']='1' if args.memory_mode=='compact' else '0'
        smoke_reference=ROOT/'.cache/liftrs/parallel-s15/s6-m2-t4';smoke=cache/'smoke';link_inputs(smoke,smoke_reference)
        for mode,inputs in [('smoke',smoke),('encoding',cache/'encoding')]:
            if mode=='encoding':link_inputs(inputs,reference)
            phase_out=out/mode;phase_out.mkdir()
            env.update(PIR_CUDA_CACHED_OUT=str(phase_out),PIR_CUDA_CACHED_CACHE=str(inputs))
            if mode=='smoke':env['PIR_CUDA_CACHED_SMOKE']='1'
            else:env.pop('PIR_CUDA_CACHED_SMOKE',None)
            run(mode,[str(binary),'-test.run=^TestCUDACachedEncoding$','-test.v','-test.timeout='+('5m' if mode=='smoke' else '90m')],source)
            record[mode]=verify_phases(phase_out,6 if mode=='smoke' else 16,args.memory_mode=='compact')
            expected=json.loads((smoke_reference/'word-metadata.json').read_text()) if mode=='smoke' else cpu['encoding']
            packed=expected['sha256'] if mode=='smoke' else expected['packed_sha256']
            assert record[mode]['canonical_sha256']==expected['canonical_sha256'] and record[mode]['packed_sha256']==packed,'CPU/CUDA digest mismatch'
        for mode in ('matched-baseline','matched-candidate','encoding'):
            for profile in sorted((out/mode).glob('*.pprof')):run('pprof-'+mode+'-'+profile.stem,[go,'tool','pprof','-top','-nodecount=25',str(binary),str(profile)])
        record['memory_comparison']=verify_memory(record,old_manifest,args.memory_mode=='compact')
        record['status']='passed'
    except BaseException as error:
        record['status']='stopped' if cancelled else 'failed';record['error']=repr(error);raise
    finally:
        record['cache_sha256']={str(p.relative_to(cache)):sha(p) for p in cache.rglob('*') if p.is_file() and not p.is_symlink() and not p.name.endswith('.tmp')}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*') if p.is_file() and p.name not in ('results.json','results.json.tmp','analysis.json')};save()

def verify_phases(out,s,compact=False):
    phases=json.loads((out/'phases.json').read_text());p={row['phase']:row for row in phases};n,k=p['load-online']['n'],p['load-online']['k']
    assert n==1<<(2*s) and 0<k<=n//2
    canonical=None;encodings=[]
    for repeat in (0,1):
        check=p[f'verify-{repeat}'];assert check['systematic_checked']==k and check['curves_checked']==16 and check['repeat_match']
        if canonical is None:canonical=check['canonical_sha256']
        assert check['canonical_sha256']==canonical
        enc=p[f'encode-{repeat}'];stats=enc['cuda_stages'];assert stats['native_total_ns']<=stats['elapsed_ns']+1_000_000
        assert stats['owned_device_bytes']<=(512<<20 if s<=10 else (40 if compact else 56)<<30)
        item={'seconds':enc['elapsed_ns']/1e9,'cuda_stages':stats}
        if compact:
            assert enc['bulk_download_bytes']==2*n and enc['compact_gpu_seconds']>=0 and enc['expand_host_seconds']>=0
            item.update({key:enc[key] for key in ('bulk_download_bytes','compact_gpu_seconds','expand_host_seconds')})
        encodings.append(item)
    assert p['pack-output']['canonical_sha256']==canonical and p['pack-output']['packed_bytes']==(n*s+7)//8
    assert p['verify-packed']['positions_checked']==1027 and p['verify-packed']['readback_match']
    return {'s':s,'n':n,'k':k,'canonical_sha256':canonical,'packed_sha256':p['pack-output']['sha256'],'encodings':encodings,'phase_seconds':{row['phase']:row['elapsed_ns']/1e9 for row in phases}}

def validate_metadata(meta,size):
    assert meta['s']==16 and meta['t']==5 and meta['k']==918672421 and meta['n']==1<<32
    assert meta['identity_verified'] and meta['sha256']==EXPECTED and meta['bytes']==size==20854558905


def verify_small(out,compact):
    baseline=json.loads((out/'baseline/online-checks.json').read_text())
    candidate=json.loads((out/'candidate/online-checks.json').read_text())
    by={r['s']:r for r in candidate}
    assert sorted(by)==[2,3,4,6,8]
    for row in baseline:
        other=by[row['s']]
        if row['s'] in (2,3,4):assert row['unsupported_prefix'] and other['unsupported_prefix'];continue
        assert row['matches']==other['matches']==4 and row['closed'] and other['closed'] and row['invalid_messages'] and other['invalid_messages']
        expected=4*row['n']-2*min(row['n'],1<<20) if compact else 0
        assert row['owned_bytes']-other['owned_bytes']==expected
    memcheck=json.loads((out/'memcheck/online-checks.json').read_text())
    assert [(r['s'],r.get('matches'),r.get('unsupported_prefix')) for r in candidate]==[(r['s'],r.get('matches'),r.get('unsupported_prefix')) for r in memcheck]
    return {'word_cases':2,'prefix_rejections':3,'memcheck':True,'compact':compact}

def verify_memory(record,baseline,compact):
    old=baseline['encoding']['encodings'][0]['cuda_stages']['owned_device_bytes']
    n=record['encoding']['n'];tile_bytes=2*min(n,1<<20) if compact else 0;expected=old-4*n+tile_bytes if compact else old
    for enc in record['encoding']['encodings']:assert enc['cuda_stages']['owned_device_bytes']==expected
    return {'resident_reuse_owned_bytes':old,'candidate_owned_bytes':expected,'saved_device_bytes':old-expected,'compaction_tile_bytes':tile_bytes,'compact':compact,'comparison':'same tables/plans; historical full resident reuse timing, matched s10 resident/candidate'}

if __name__=='__main__':main()
