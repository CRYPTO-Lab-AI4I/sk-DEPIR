#!/usr/bin/env python3
"""Fixed-load buffered publication gate; no GPU and no compute exchange files."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import resource
import stat
import runpy
import subprocess
import time

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]
SOURCE=runpy.run_path(str(HERE.parent/'gpu-staged-transport/run.py'))
PROCESS=runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))
RSS=runpy.run_path(str(HERE.parent/'hpc-cached-s15/run.py'))
MEM=runpy.run_path(str(HERE.parent/'cpu-streamed-constants/cache_memory.py'))
MIB=1<<20


def sha(path):
    h=hashlib.sha256()
    with path.open('rb')as f:
        for block in iter(lambda:f.read(1<<20),b''):h.update(block)
    return h.hexdigest()


def review(rows,cache,small=False):
    """Require complete paired modes, actual file-call aggregation and identities."""
    values=[r for r in rows if r.get('kind')=='measurement']
    modes=[(0,1),(4*MIB,1),(4*MIB,2),(0,2),(0,4),(4*MIB,4)]
    expected=modes+list(reversed(modes))
    assert len(values)==12 and [(r['buffer_bytes'],r['writers'])for r in values]==expected
    assert [r['trial']for r in values]==list(range(12))
    assert sum(r.get('kind')=='completed'and r['trials']==12 for r in rows)==1
    size=128<<10 if small else 24*MIB
    digests={}
    for trial,row in enumerate(values):
        assert row['repeat']==trial//6 and row['matches'] and row['file_bytes']==size and row['total_bytes']==4*size
        assert row['codec_chunk_bytes']==65536 and row['payloads_retained']==(trial==11)
        assert 0<row['peak_rss_bytes']<=2<<30
        assert row['publication_seconds']>0 and row['readback_seconds']>0
        assert row['publish_readback_seconds']>=row['publication_seconds']+row['readback_seconds']
        assert len(row['files'])==4 and {f['shard']for f in row['files']}==set(range(4))
        for entry in row['files']:
            shard=entry['shard'];stage=f'io-trial-{trial:02d}-file-{shard}'
            assert entry['stage']==stage and entry['writer']==shard%row['writers'] and entry['pid']>0
            record=entry['record'];stats=entry['stats'];digest=record['sha256']
            assert record['bytes']==size and len(digest)==64 and all(c in '0123456789abcdef'for c in digest)
            if shard in digests:assert digest==digests[shard]
            else:digests[shard]=digest
            chunk=min(size,row['buffer_bytes']or 65536)
            assert stats['bytes']==size and stats['calls']==(size+chunk-1)//chunk
            assert stats['sizes']=={str(chunk):size//chunk}
            assert all(stats[f]>=0 for f in ('file_write_seconds','produce_flush_seconds','flush_seconds','sync_seconds','close_seconds'))
            assert entry['publish_seconds']>=stats['produce_flush_seconds']+stats['sync_seconds']+stats['close_seconds']
            manifest=json.loads((cache/stage/'manifest.json').read_text())
            assert manifest['stage']==stage and manifest['files']['payload.bin']==record
            payload=cache/stage/'payload.bin'
            if trial==11:assert payload.stat().st_size==size and sha(payload)==digest
            else:assert not payload.exists()
    return values


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('name',nargs='?');p.add_argument('--check',action='store_true');a=p.parse_args()
    if a.check:
        assert 4*24*MIB+24*MIB<128*MIB
        print('1HPCnode/2CPU/4GiB host;0GPU;2GiB RSS,128MiB retained,2GiB scratch,32MiB/file;15min scheduler;96MiB/trial')
        return
    if not a.name or not a.name.replace('-','').isalnum():p.error('fresh alphanumeric/hyphen run name required')
    if os.environ.get('SLURM_JOB_NUM_NODES')!='1':raise RuntimeError('one-node allocation required')
    resource.setrlimit(resource.RLIMIT_FSIZE,(32*MIB,32*MIB))
    out=HERE/'runs'/a.name;out.mkdir(parents=True,exist_ok=False)
    cache=ROOT/'.cache/liftrs/preprocessing-persistence'/a.name;cache.mkdir(parents=True,exist_ok=False)
    scratch=Path('/tmp')/('depir-persistence-'+os.environ['SLURM_JOB_ID']);scratch.mkdir(exist_ok=False)
    source=scratch/'source';source.mkdir()
    began=time.monotonic();record=dict(status='running',job_id=os.environ['SLURM_JOB_ID'],
        revision=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
        started_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),cache=str(cache),commands=[],
        policy=dict(nodes=1,gpus=0,cpus=2,host_request_GiB=4,rss_guard_GiB=2,cgroup_guard_GiB=3.5,
            retained_cap_MiB=128,scratch_cap_GiB=2,file_cap_MiB=32,worker_seconds=540,controller_seconds=720,scheduler_seconds=900))
    env=dict(os.environ);env.update(GOTOOLCHAIN='local',GOPROXY='off',GOSUMDB='off',
        GOPATH=str(ROOT/'.cache/liftrs/toolchains/gopath'),GOCACHE=str(scratch/'go-cache'),TMPDIR=str(scratch),
        GOMAXPROCS='2',GOFLAGS='-p=2',GOMEMLIMIT=str(1<<30),CGO_ENABLED='1',CC='gcc',CXX='g++',LIFTRS_CPU_PROGRESS='1')
    def save():(out/'results.json').write_text(json.dumps(record,indent=2)+'\n')
    def event(phase,kind,**extra):
        row=dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),phase=phase,event=kind,
            elapsed_seconds=time.monotonic()-began,**extra)
        with (out/'progress.jsonl').open('a')as f:f.write(json.dumps(row)+'\n')
        print(json.dumps(row),flush=True)
    def tree_bytes(path):
        total=0
        for entry in path.rglob('*'):
            try:info=entry.stat()
            except FileNotFoundError:continue  # Just-verified synthetic files may be pruned.
            if stat.S_ISREG(info.st_mode):total+=info.st_size
        return total
    def run(label,argv,limit=90):
        event(label,'started');start=time.monotonic();last=start;peak=0;peak_cache=0;peak_combined=0;peak_cgroup=0;guard=None;shutdown=None
        with (out/(label+'.log')).open('w')as log:
            process=subprocess.Popen(list(map(str,argv)),cwd=source,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                while process.poll()is None:
                    rss=RSS['group_rss'](process.pid);peak=max(peak,rss)
                    # Observe only the final retained case. Transient-case maps
                    # could keep unlinked synthetic files alive beyond the cap.
                    observer=MEM['FileCache']()
                    try:resident=sum(observer.sample(list((cache/'probe').glob('io-trial-11-file-*/payload.bin'))).values())
                    finally:observer.close()
                    peak_cache=max(peak_cache,resident);peak_combined=max(peak_combined,rss+resident)
                    cg=MEM['job_cgroup'](record['job_id'])
                    if cg:peak_cgroup=max(peak_cgroup,cg['current_bytes'])
                    if rss>2<<30:guard='process-group RSS guard'
                    if cg and cg['current_bytes']>int(3.5*(1<<30)):guard='cgroup memory guard'
                    if tree_bytes(cache)>128*MIB or tree_bytes(scratch)>2<<30:guard='storage guard'
                    if time.monotonic()-start>limit or time.monotonic()-began>720:guard='deadline'
                    if time.monotonic()-last>=30:
                        event(label,'heartbeat',remaining_work='unknown',peak_rss_bytes=peak,resident_retained_payload_bytes=resident,cgroup=cg);last=time.monotonic()
                    with (out/'resources.jsonl').open('a')as f:f.write(json.dumps(dict(phase=label,elapsed_seconds=time.monotonic()-began,rss_bytes=rss,resident_retained_payload_bytes=resident,cgroup=cg))+'\n')
                    if guard:shutdown=PROCESS['stop_process'](process);break
                    time.sleep(.25)
            finally:
                if process.poll()is None:shutdown=PROCESS['stop_process'](process)
        # Final-case files remain stable after the worker exits; capture them
        # even when a fast final trial falls between periodic observations.
        observer=MEM['FileCache']()
        try:resident=sum(observer.sample(list((cache/'probe').glob('io-trial-11-file-*/payload.bin'))).values())
        finally:observer.close()
        peak_cache=max(peak_cache,resident)
        record['commands'].append(dict(label=label,argv=list(map(str,argv)),seconds=time.monotonic()-start,
            peak_rss_bytes=peak,peak_retained_payload_cache_bytes=peak_cache,peak_rss_plus_retained_payload_cache_bytes=peak_combined,
            peak_sampled_cgroup_bytes=peak_cgroup or None,guard=guard,shutdown=shutdown,exit_code=process.returncode));save()
        event(label,'failed'if guard or process.returncode else'completed',peak_rss_bytes=peak)
        if guard or process.returncode:raise RuntimeError(label+': '+str(guard or process.returncode))
    save()
    try:
        record['source_sha256']=SOURCE['snapshot'](record['revision'],source);save()
        go=ROOT/'.cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go'
        run('host',['hostname']);run('cpu',['lscpu']);run('filesystem',['stat','-f','-c','type=%T block=%S',cache])
        run('go',[go,'version']);run('compiler',['g++','--version'])
        binary=cache/'persistence.test'
        run('build',[go,'test','-c','-tags','liftrs_cpp','-o',binary,'./cmd/pir-preprocess'],180)
        tests='TestBufferedPayloadEquivalenceAndFailure|TestBufferedPublicationFailureAndReuse|TestPublicationFailureAndReuse|TestCLIRecoveryRestart|TestPersistentReferenceAndRestart'
        run('correctness',[binary,'-test.run',tests,'-test.v','-test.timeout=90s'],100)
        race=cache/'persistence-race.test'
        run('race-build',[go,'test','-race','-c','-o',race,'./cmd/pir-preprocess'],120)
        env.update(PIR_PERSISTENCE_OUTPUT=str(cache/'race-probe'),PIR_PERSISTENCE_SMALL='1')
        run('race',[race,'-test.run=TestBufferedPayloadEquivalenceAndFailure|TestBufferedPublicationFailureAndReuse|TestPersistenceProbe',
            '-test.v','-test.timeout=90s'],100)
        race_rows=SOURCE['parse'](out/'race.log');review(race_rows,cache/'race-probe',small=True)
        env.pop('PIR_PERSISTENCE_SMALL')
        env['PIR_PERSISTENCE_OUTPUT']=str(cache/'probe')
        run('persistence',[binary,'-test.run=^TestPersistenceProbe$','-test.v','-test.timeout=520s','-test.cpuprofile='+str(out/'persistence.pprof')],540)
        rows=SOURCE['parse'](out/'persistence.log');review(rows,cache/'probe')
        run('profile',[go,'tool','pprof','-top',binary,out/'persistence.pprof'])
        record.update(status='passed',measurements=rows,race_measurements=race_rows)
    except BaseException as error:
        record.update(status='failed',error=str(error));raise
    finally:
        record.update(elapsed_seconds=time.monotonic()-began,
            evidence_sha256={str(p.relative_to(out)):sha(p)for p in out.rglob('*')if p.is_file()and p.name!='results.json'},
            cache_sha256={str(p.relative_to(cache)):sha(p)for p in cache.rglob('*')if p.is_file()})
        save()


if __name__=='__main__':main()
