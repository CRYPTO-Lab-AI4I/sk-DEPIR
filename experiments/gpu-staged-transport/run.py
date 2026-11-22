#!/usr/bin/env python3
"""Bounded two-node GPU runner, reused by explicit component suites.

The default suite is host-staged field arithmetic; extra_suite replaces only
its workload. Source freezing, limits, deadlines, shutdown and seals stay shared.
"""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import resource
import runpy
import shlex
import shutil
import subprocess
import time

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]
SIZES=(1<<20,8<<20,32<<20)


def verify(rows,small=False):
    hosts={r['rank']:r['host'] for r in rows if r.get('kind')=='host'}
    assert set(hosts)=={0,1} and len(set(hosts.values()))==2
    fields=('host_seconds','producer_seconds','d2h_seconds','exchange_seconds','h2d_seconds','consumer_seconds','pipeline_seconds')
    for rank in (0,1):
        assert sum(r.get('rank')==rank and r.get('phase')=='workflow' and r.get('event')=='completed' for r in rows)==1
        for size in ((32<<10,) if small else SIZES):
            samples=[r for r in rows if r.get('kind')=='measurement' and r['rank']==rank and r['bytes']==size]
            assert {r['sample'] for r in samples}==({-1,0} if small else {-1,0,1,2})
            assert len(samples)==(2 if small else 4)
            for row in samples:
                assert row['matches'] and all(row[name]>0 for name in fields)
                assert row['pipeline_seconds']>=max(row[name]for name in fields if name not in ('host_seconds','pipeline_seconds'))
                assert row['owned_device_bytes']<=128<<20 and row['pinned_host_bytes']==2*size
                assert row['peak_rss_bytes']<=2<<30
    for row in rows:
        if row.get('kind')=='measurement' and row['rank']==0:
            pair=[r for r in rows if r.get('kind')=='measurement' and r['rank']==1 and r['bytes']==row['bytes'] and r['sample']==row['sample']]
            assert len(pair)==1 and all(row[name]==pair[0][name] for name in fields)


def parse(path):
    rows=[]
    for line in path.read_text().splitlines():
        start=line.find('{')
        if start<0:continue
        payload=line[start:]
        # Go timeout stacks contain struct literals; these are not JSON events.
        if not payload[1:].lstrip().startswith('"'):continue
        rows.append(json.loads(payload))  # Malformed event JSON must still fail.
    return rows


def parse_rank_logs(path,ranks=4):
    """Strict rank streams when enabled; otherwise historical launcher output."""
    sources=[path.with_name(path.stem+f'-{rank}.{stream}.log')for rank in range(ranks)for stream in ('stdout','stderr')]
    if not any(p.exists()for p in sources):return parse(path)
    assert all(p.is_file()for p in sources),'incomplete rank log set'
    return [row for source in sources for row in parse(source)]


def rank_output_text(path,ranks=4):
    sources=[path.with_name(path.stem+f'-{rank}.{stream}.log')for rank in range(ranks)for stream in ('stdout','stderr')]
    if not any(p.exists()for p in sources):return path.read_text()
    assert all(p.is_file()for p in sources),'incomplete rank log set'
    return '\n'.join(p.read_text()for p in sources)


def build_input(name):
    path=Path(name)
    if any(part in ('runs','artifacts','message-board','docs') for part in path.parts):return False
    code={'.go','.c','.cc','.cpp','.cxx','.h','.hpp','.cu','.cuh','.s','.S','.py','.sh','.sbatch','.inc','.syso'}
    # This reviewed input manifest is required at runtime; other experiment JSON
    # remains excluded so historical results cannot inflate the frozen source.
    return name=='experiments/distributed-coordinator-stream/reference-inputs.json' or path.suffix in code or path.name in ('go.mod','go.sum') or (path.suffix=='.json' and not name.startswith('experiments/'))


def snapshot(revision,destination):
    # Select build inputs, not historical datasets. Preflight Git object sizes
    # before reading any contents or encountering filesystem/resource limits.
    tree=subprocess.check_output(['git','ls-tree','-rl',revision],cwd=ROOT,text=True).splitlines()
    selected={}
    for line in tree:
        metadata,separator,name=line.partition('\t')
        if not build_input(name):continue
        fields=metadata.split();assert separator and fields[1]=='blob',name
        count=int(fields[3])
        if count>32<<20:raise ValueError('build input exceeds32MiB file cap: '+name)
        selected[name]=count
    if sum(selected.values())>128<<20:raise ValueError('build snapshot exceeds128MiB cap')
    requests=''.join(revision+':'+name+'\n' for name in selected)
    data=subprocess.check_output(['git','cat-file','--batch'],cwd=ROOT,input=requests.encode())
    position=0;digests={}
    for name,expected_size in selected.items():
        end=data.index(b'\n',position);header=data[position:end].split();assert header[1]==b'blob',name
        count=int(header[2]);assert count==expected_size,name
        position=end+1;blob=data[position:position+count];position+=count+1
        path=Path(name);assert not path.is_absolute() and '..'not in path.parts
        target=destination/path;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(blob)
        digests[name]=hashlib.sha256(blob).hexdigest()
    return digests


def main(here=HERE,extra_suite=None,require_gpu=True):
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('name',nargs='?');parser.add_argument('--check',action='store_true');args=parser.parse_args()
    if args.check:
        assert (here/'run.py').is_file() and (HERE/'rank.sh').is_file()
        print('2 nodes/2 H200/4 CPU/8GiB host total;128MiB owned/card,64MiB pinned/rank,2GiB RSS/rank; no submission')
        return
    if not args.name or not args.name.replace('-','').isalnum():parser.error('fresh name required')
    if not require_gpu and extra_suite is None:raise ValueError('host-only mode requires an explicit workload')
    if os.environ.get('SLURM_JOB_NUM_NODES')!='2':raise RuntimeError('two nodes required')
    resource.setrlimit(resource.RLIMIT_FSIZE,(32<<20,32<<20))
    out=here/'runs'/args.name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs'/here.name/args.name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('depir-'+here.name+'-'+os.environ['SLURM_JOB_ID']);scratch.mkdir(exist_ok=False)
    source=scratch/'source';source.mkdir()
    env=dict(os.environ);env.update(GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),
        GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),CGO_ENABLED='0',GOMAXPROCS='2',GOFLAGS='-p=2',GOMEMLIMIT=str(1<<30))
    start=time.monotonic();record=dict(job_id=os.environ['SLURM_JOB_ID'],status='running',started_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
        revision=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),commands=[],cache=str(cache),
        resources=dict(nodes=2,ranks=2,gpus=2 if require_gpu else 0,gpu_model='H200'if require_gpu else None,cpus_per_node=2,host_request_GiB_per_node=4,
          owned_cap_MiB_per_card=128,pinned_cap_MiB_per_rank=64,rss_cap_GiB_per_rank=2,
          retained_cap_MiB=128,scratch_cap_GiB=2,controller_seconds=720,scheduler_seconds=900))
    def save():(out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
    def event(label,kind):
        row=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),phase=label,event=kind,completed=len(record['commands']),unit='commands',elapsed_seconds=time.monotonic()-start)
        with (out/'progress.jsonl').open('a')as f:f.write(json.dumps(row)+'\n')
        print(json.dumps(row),flush=True)
    def tree_bytes(path):return sum(p.stat().st_size for p in path.rglob('*')if p.is_file())
    def run(label,argv,timeout=30):
        if time.monotonic()-start>720:raise RuntimeError('controller deadline')
        event(label,'started');began=time.monotonic()
        process=None
        try:
            with (out/(label+'.log')).open('w')as log:
                process=subprocess.Popen(list(map(str,argv)),cwd=source,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
                code=process.wait(timeout=min(timeout,max(1,720-(time.monotonic()-start))))
            record['commands'].append(dict(label=label,argv=list(map(str,argv)),seconds=time.monotonic()-began,exit_code=code));save()
            if code:raise RuntimeError(label+' failed')
            if tree_bytes(cache)>128<<20 or tree_bytes(scratch)>2<<30:raise RuntimeError('artifact/scratch cap')
            event(label,'completed')
        except BaseException:
            if process is not None and process.poll()is None:
                record['shutdown']=stop(process);save()
            event(label,'failed');raise
    save()
    try:
        record['source_sha256']=snapshot(record['revision'],source);save()
        stop=runpy.run_path(str(source/'experiments/gpu-preprocessing-workflow/run.py'))['stop_process']
        run('hosts',['scontrol','show','hostnames',os.environ['SLURM_JOB_NODELIST']])
        run('mpi-plugins',['srun','--mpi=list'])
        run('mpi',['ompi_info','--parsable','--all'])
        if require_gpu:run('nvcc',['nvcc','--version'])
        run('compiler',['mpicxx','--version'])
        go=ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go'
        run('go',[go,'version'])
        if extra_suite is not None:
            record.update(extra_suite(dict(run=run,source=source,cache=cache,out=out,go=go,env=env,resources=record['resources'],save=save)))
            record['status']='passed'
        else:
            fixture=cache/'fixture.test';field=cache/'field.bin'
            run('fixture-build',[go,'test','-c','-o',fixture,'./encoding/liftrs-shared'],60)
            env['PIR_GPU_TRANSPORT_FIELD']=str(field)
            run('field-fixture',[fixture,'-test.run=^TestGPUTransportFieldFixture$','-test.v','-test.cpuprofile='+str(out/'fixture.pprof')])
            env.pop('PIR_GPU_TRANSPORT_FIELD')
            includes=shlex.split(subprocess.check_output(['mpicxx','--showme:compile'],env=env,text=True))
            obj=cache/'probe.o';binary=cache/'probe'
            run('cuda-build',['nvcc','-std=c++17','-O3','-lineinfo','-pg','-arch=sm_90',*includes,'-c','experiments/gpu-staged-transport/probe.cu','-o',obj],60)
            cuda_lib=Path(shutil.which('nvcc')).resolve().parents[1]/'lib64'
            run('link',['mpicxx','-pg',obj,'-L'+str(cuda_lib),'-Wl,-rpath,'+str(cuda_lib),'-lcudart','-o',binary],60)
            rank_wrapper=cache/'rank.sh';shutil.copyfile(source/'experiments/gpu-staged-transport/rank.sh',rank_wrapper)
            launcher=['timeout','--signal=TERM','--kill-after=10s','280s','srun','--mpi=pmix_v3',
                      '--nodes=2','--ntasks=2','--ntasks-per-node=1','--cpus-per-task=2',
                      '--gpus-per-task=h200:1','--gpu-bind=single:1','--cpu-bind=cores','--label','--kill-on-bad-exit=1',
                      'bash',rank_wrapper]
            run('normal',[*launcher,cache/'native-normal',binary,field],300)
            rows=parse(out/'normal.log');verify(rows)
            run('memcheck',[*launcher,cache/'native-memcheck','compute-sanitizer','--tool','memcheck','--leak-check','full','--error-exitcode','99',binary,field,'--small'],300)
            sanitized=parse(out/'memcheck.log');verify(sanitized,small=True)
            text=(out/'memcheck.log').read_text()
            assert text.count('ERROR SUMMARY: 0 errors')==2 and text.count('LEAK SUMMARY: 0 bytes')==2
            for rank in (0,1):
                profiles=sorted(cache.glob('native-normal-'+str(rank)+'.*'));assert len(profiles)==1
                run('native-profile-'+str(rank),['gprof','-b',binary,profiles[0]])
            record.update(status='passed',measurements=rows,sanitized=sanitized)
    except BaseException as error:
        record.update(status='failed',error=str(error));raise
    finally:
        record.update(elapsed_seconds=time.monotonic()-start,
            evidence_sha256={str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest()for p in out.rglob('*')if p.is_file()and p.name!='results.json'},
            cache_sha256={str(p.relative_to(cache)):hashlib.sha256(p.read_bytes()).hexdigest()for p in cache.rglob('*')if p.is_file()})
        save()
        # Scratch is bounded and left intact on failure for forensics; it is
        # never used as an exchange array or as retained preprocessing storage.


if __name__=='__main__':main()
