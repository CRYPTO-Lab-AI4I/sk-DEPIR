#!/usr/bin/env python3
"""Guarded Slurm first use; completed code data goes directly to shared storage."""
import argparse
import datetime
import fcntl
import hashlib
import json
import os
from pathlib import Path
import runpy
import shutil
import signal
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
GIB = 1 << 30
helpers = runpy.run_path(str(ROOT/'experiments/hpc-cached-s15/run.py'))
sha, group_rss = helpers['sha'], helpers['group_rss']


def mem_available():
    return int(next(x.split()[1] for x in Path('/proc/meminfo').read_text().splitlines()
                    if x.startswith('MemAvailable:'))) * 1024


def tree_bytes(path):
    total=0
    for p in path.rglob('*'):
        try:
            if p.is_file() and not p.is_symlink(): total+=p.stat().st_size
        except FileNotFoundError: pass # A checkpoint may atomically rename a temp file.
    return total


def stop_group(process):
    try: os.killpg(process.pid,signal.SIGTERM)
    except ProcessLookupError: pass
    try: process.wait(timeout=60)
    except subprocess.TimeoutExpired:
        try: os.killpg(process.pid,signal.SIGKILL)
        except ProcessLookupError: pass
        process.wait()


def guard_reason(rss, available, cache, scratch, elapsed, timeout, policy):
    if rss > policy['rss_guard_gib']*GIB: return 'RSS guard'
    if available < policy['minimum_available_memory_gib']*GIB: return 'node available-memory guard'
    if cache > policy['cache_guard_gib']*GIB: return 'cache guard'
    if scratch > policy['scratch_guard_gib']*GIB: return 'scratch guard'
    if elapsed > timeout: return 'command deadline'
    return None


def verify_phases(out, s):
    phases = json.loads((out/'phases.json').read_text())
    p = {x['phase']: x for x in phases}
    n, k = p['load-online']['n'], p['load-online']['k']
    assert n == 1 << (2*s) and 1 < k <= n//2
    assert p['verify-0']['systematic_checked'] == k
    assert p['verify-0']['curves_checked'] == 16
    assert p['verify-packed']['positions_checked'] == 1027 and p['verify-packed']['readback_match']
    assert p['pack-output']['packed_bytes'] == (n*s+7)//8
    assert p['verify-0']['canonical_sha256'] == p['pack-output']['canonical_sha256']
    enc = p['encode-0']
    total = sum(enc['native_stages'][x] for x in
                ('taylor_ns','columns_transposes_ns','rows_ns','buffer_ns','pointwise_ns','assembly_ns'))
    assert total <= enc['native_after']['steady_ns']-enc['native_before']['steady_ns']+1000000
    return {'s':s, 'k':k, 'n':n, 'canonical_sha256':p['verify-0']['canonical_sha256'],
            'packed_sha256':p['pack-output']['sha256'],
            'phase_seconds':{x['phase']:x['elapsed_ns']/1e9 for x in phases}}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('name')
    parser.add_argument('--validate', action='store_true')
    parser.add_argument('--admission', default='validation-01')
    args = parser.parse_args()
    if not os.environ.get('SLURM_JOB_ID'): raise SystemExit('Use Slurm')
    if any(not name or any(c not in 'abcdefghijklmnopqrstuvwxyz0123456789-_' for c in name)
           for name in (args.name,args.admission)):
        raise SystemExit('Use a fresh lowercase/digit/-/_ run name')
    job = os.environ['SLURM_JOB_ID']
    policy = json.loads((HERE/'policy.json').read_text())
    if args.validate:
        policy.update(workers=2, rss_guard_gib=8, go_memory_gib=1,
                      minimum_initial_available_memory_gib=12, minimum_available_memory_gib=4,
                      minimum_free_scratch_gib=8, minimum_free_cache_gib=2, runner_wall_seconds=800)
    elif len(os.sched_getaffinity(0)) < policy['workers']:
        raise SystemExit('Allocation provides fewer CPUs than the policy')
    out = HERE/'runs'/args.name
    out.mkdir(parents=True, exist_ok=False)
    scratch = Path('/tmp')/('sk-depir-s16-'+job)
    scratch.mkdir(exist_ok=False)
    cache_root = ROOT/'.cache/liftrs/hpc-s16'
    cache_root.mkdir(parents=True, exist_ok=True)
    cache = cache_root/('smoke-'+args.name if args.validate else 's16-m2-t5')
    cache.mkdir(exist_ok=True)
    cancelled = []
    cache_lock = None
    signal.signal(signal.SIGTERM, lambda *_: cancelled.append('Slurm/user termination'))
    started = time.monotonic()
    record = {'status':'running','job_id':job,'policy':policy,'cache':str(cache),
              'scratch':str(scratch),'validation':args.validate,'commands':[],
              'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip()}
    env = dict(os.environ, GOTOOLCHAIN='local', GOPROXY='off', GOSUMDB='off',
               GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'), GOCACHE=str(scratch/'go-cache'),
               TMPDIR=str(scratch), CGO_ENABLED='1', CXX='g++',
               GOMAXPROCS=str(4 if args.validate else policy['workers']), GOFLAGS='-p=4')
    go = str(ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go')

    def save():
        p = out/'results.json.tmp'
        p.write_text(json.dumps(record,indent=2)+'\n'); p.replace(out/'results.json')

    def run(label, argv, cwd=ROOT, timeout=600, required=True):
        if time.monotonic()-started >= policy['runner_wall_seconds']: raise RuntimeError('runner deadline')
        begin = time.monotonic()
        peak, reason = 0, None
        with (out/(label+'.log')).open('w') as log, (out/'progress.jsonl').open('a') as progress:
            process = subprocess.Popen(argv,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            last_disk, sizes = 0, (0,0)
            try:
                while process.poll() is None:
                    now = time.monotonic()
                    peak = max(peak,group_rss(process.pid))
                    if now-last_disk > 30:
                        sizes = tree_bytes(cache_root), tree_bytes(scratch); last_disk=now
                    reason = guard_reason(peak,mem_available(),*sizes,now-begin,timeout,policy)
                    if now-started > policy['runner_wall_seconds']: reason='runner deadline'
                    if cancelled: reason=cancelled[0]
                    worker_state={}
                    try: worker_state=json.loads((out/label/'state.json').read_text())
                    except (FileNotFoundError,json.JSONDecodeError): pass
                    progress.write(json.dumps({'timestamp_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),
                                              'command':label,'rss_bytes':peak,'cache_bytes':sizes[0],
                                          'scratch_bytes':sizes[1],'elapsed_seconds':now-begin,
                                          'worker':worker_state,'guard':reason})+'\n')
                    progress.flush()
                    if reason:
                        stop_group(process)
                        break
                    time.sleep(policy['monitor_seconds'])
            finally:
                if process.poll() is None: stop_group(process)
        record['commands'].append({'label':label,'argv':list(map(str,argv)), 'exit_code':process.returncode,
                                   'elapsed_seconds':time.monotonic()-begin,'sampled_group_peak_rss_bytes':peak,'guard':reason})
        save()
        if reason or (required and process.returncode): raise RuntimeError(label+': '+str(reason or process.returncode))

    save()
    try:
        if mem_available() < policy['minimum_initial_available_memory_gib']*GIB: raise RuntimeError('insufficient node memory for admission')
        if shutil.disk_usage(scratch).free < policy['minimum_free_scratch_gib']*GIB: raise RuntimeError('insufficient scratch space')
        if shutil.disk_usage(cache_root).free < policy['minimum_free_cache_gib']*GIB: raise RuntimeError('insufficient cache space')
        for label, argv in [('hostname',['hostname']),('cpu',['lscpu']),('memory',['free','-b']),
                            ('filesystem',['df','-T','-B1',str(ROOT),str(scratch)]),
                            ('compiler',['g++','--version']),('go',[go,'version']),
                            ('job',['scontrol','show','job',job]),('os',['cat','/etc/os-release'])]: run(label,argv)
        record['proc_status']=Path('/proc/self/status').read_text()
        record['cgroup']=Path('/proc/self/cgroup').read_text()
        record['cgroup_limits']={}
        for line in record['cgroup'].splitlines():
            if line.startswith('0::'):
                base=Path('/sys/fs/cgroup')/line[3:].lstrip('/')
                for directory in [base,*base.parents]:
                    if not str(directory).startswith('/sys/fs/cgroup'): continue
                    for field in ('memory.max','memory.high','cpu.max','cpuset.cpus.effective'):
                        p=directory/field
                        if p.is_file():record['cgroup_limits'][str(p)]=p.read_text().strip()
        for label,argv in [('perf',['perf','stat','-e','task-clock','true'])]:
            try:run(label,argv,required=False)
            except FileNotFoundError: record[label]='unavailable'
        source=scratch/'repo'
        shutil.copytree(ROOT,source,ignore=shutil.ignore_patterns('.git','.cache','runs','__pycache__'))
        record['source_sha256']={str(p.relative_to(source)):sha(p) for p in source.rglob('*')
                                if p.is_file() and (p.suffix in ('.go','.cpp','.h','.hpp','.py','.sbatch')
                                                   or p.name in ('go.mod','go.sum','Makefile','policy.json'))}
        for name,digest in record['source_sha256'].items():
            blob=subprocess.run(['git','show',record['revision']+':'+name],cwd=ROOT,capture_output=True)
            if blob.returncode or hashlib.sha256(blob.stdout).hexdigest()!=digest:
                target=ROOT/'experiments/source-snapshots'/(digest+'.txt')
                if not target.exists():target.write_bytes((source/name).read_bytes())
        save()
        if args.validate:
            run('race',[go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared',
                        '-run','TestScalingPreparation|TestScalingCPPLargeFields|TestCPPCountBoundary','-count=1'],source)
            run('vet',[go,'vet','-tags','liftrs_cpp','./encoding/...','./pirconfig','./planner'],source)
            run('legacy-build',['make','-C','encoding/rme_legacy','rme'],source)
            env['CGO_ENABLED']='0'; run('suite-go',[go,'test','./...','-timeout=8m'],source)
            env['CGO_ENABLED']='1'; run('suite-native',[go,'test','-tags','liftrs_cpp','./...','-timeout=8m'],source)
        else:
            admission=json.loads((HERE/'runs'/args.admission/'results.json').read_text())
            assert admission['status']=='passed' and admission['validation']
            # Only unchanged runtime sources may use the preceding admission checks.
            for name,digest in admission['source_sha256'].items():
                assert record['source_sha256'].get(name)==digest, ('admission source changed',name)
        if not args.validate:
            cache_lock=(cache_root/'s16.lock').open('a+')
            fcntl.flock(cache_lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
            runtime_sources={name:digest for name,digest in record['source_sha256'].items() if name.startswith('encoding/')}
            provenance=cache/'provenance.json'
            if provenance.exists():
                old=json.loads(provenance.read_text())
                assert old['runtime_sources']==runtime_sources, 'cache source changed: review retained provenance before reuse'
            else:
                assert not any((cache/name).exists() for name in ('support.bin','online.bin','g.bin')), 'cache lacks provenance'
                temporary=cache/'provenance.json.tmp'
                temporary.write_text(json.dumps({'revision':record['revision'],'runtime_sources':runtime_sources,
                                                 'spec':{'s':16,'m':2,'t':5}},indent=2)+'\n')
                temporary.replace(provenance)
        binary=cache_root/('probe-'+args.name)
        run('build',[go,'test','-c','-tags','liftrs_cpp','-o',str(binary),'./encoding/liftrs-shared'],source)
        record['binary']=str(binary); record['binary_sha256']=sha(binary); save()
        if args.validate:env['PIR_S16_SMOKE']='1'
        env['PIR_S16_CACHE']=str(cache)
        summaries=[]
        for mode in ('initial','cached'):
            phase_out=out/mode; phase_out.mkdir()
            env['PIR_S16_OUT']=str(phase_out)
            env['PIR_S16_ONLINE_ONLY']='1' if mode=='cached' else '0'
            if mode=='cached':
                cached=cache_root/('cached-'+args.name); cached.mkdir(exist_ok=False)
                for name in ('online.bin','online-metadata.json'):(cached/name).symlink_to(cache/name)
                env['PIR_S16_CACHE']=str(cached)
            seconds=600 if args.validate else policy['initial_timeout_seconds'] if mode=='initial' else policy['cached_timeout_seconds']
            run(mode,['/usr/bin/time','-v',str(binary),'-test.run=^TestS16FirstUse$',
                      '-test.v','-test.timeout='+str(seconds+120)+'s'],source,seconds)
            summary=verify_phases(phase_out,6 if args.validate else 16)
            summaries.append(summary);record[mode]=summary;save()
            if args.validate:
                old=json.loads((ROOT/'.cache/liftrs/parallel-s15/s6-m2-t4/word-metadata.json').read_text())
                assert summary['canonical_sha256']==old['canonical_sha256'] and summary['packed_sha256']==old['sha256']
                if mode=='initial':
                    for name in ('support.bin','online.bin','g.bin'):
                        assert sha(cache/name)==sha(ROOT/'.cache/liftrs/parallel-s15/s6-m2-t4'/name), ('smoke constants differ',name)
        assert summaries[0]['canonical_sha256']==summaries[1]['canonical_sha256']
        assert summaries[0]['packed_sha256']==summaries[1]['packed_sha256']
        record['status']='finalizing';save()
        record['cache_sha256']={p.name:sha(p) for p in cache.iterdir() if p.is_file() and not p.name.endswith('.tmp')}
        record['evidence_sha256']={str(p.relative_to(out)):sha(p) for p in out.rglob('*')
                                  if p.is_file() and p.name not in ('results.json','results.json.tmp')}
        record['status']='passed';save()
    except BaseException as error:
        record['status']='failed';record['error']=repr(error);save();raise
    finally:
        if cache_lock is not None:cache_lock.close()
        summary=['# s16 experiment status','', 'Status: **'+record['status']+'**.','',
                 'Exact evidence: [results.json](results.json). Completed cache stages remain at `'+str(cache)+'`.','']
        if 'initial' in record:summary += ['Initial run: `'+json.dumps(record['initial'])+'`.','']
        if 'error' in record:summary += ['Stop reason: `'+record['error']+'`.','']
        (out/'summary.md').write_text('\n'.join(summary))
        if not args.validate:
            now=datetime.datetime.now(datetime.timezone.utc)
            filename=now.strftime('%Y-%m-%d-%H%M%S')+'-codex-scale-012-results.md'
            message='# Guarded s16 run '+record['status']+'\n\nTimestamp: '+now.isoformat()+' UTC. Agent: Codex supervisor. Work: SCALE-012.\n\n'
            message+='See [results](../experiments/hpc-s16/runs/'+args.name+'/results.json) and [summary](../experiments/hpc-s16/runs/'+args.name+'/summary.md). '
            message+='Status is '+record['status']+'. Cache stages and profiles are retained; no catalogue or TRL claim is changed automatically. '
            message+='README/report links were checked before launch; review final evidence before updating capability claims.\n'
            (ROOT/'message-board'/filename).write_text(message)
            with (ROOT/'message-board/README.md').open('a') as board:
                board.write('| '+now.strftime('%Y-%m-%d %H:%M')+' | SCALE-012 | Results | [Guarded s16 '+record['status']+']('+filename+') |\n')


if __name__=='__main__':main()
