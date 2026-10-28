#!/usr/bin/env python3
"""Profile s=13 and retain reloadable local preprocessing, with bounded resources."""
import argparse,hashlib,json,os,platform,shutil,statistics,subprocess,tempfile,time
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
def sha(p):
    h=hashlib.sha256()
    with p.open('rb') as file:
        for block in iter(lambda:file.read(1<<20),b''):h.update(block)
    return h.hexdigest()
def describe_code(event):
    # The profiling driver emits code metadata after cache loading, so discovery
    # sample counts are deliberately not required here.
    return f'exact K={event["k"]}, reused_cache={event.get("reused_cache", False)}'

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go',default='go')
    parser.add_argument('--reuse-cache',action='store_true')
    parser.add_argument('--run-name',default='setup')
    args=parser.parse_args()
    if not args.run_name or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.run_name):raise SystemExit('invalid run name')
    output_dir=HERE/args.run_name
    output_dir.mkdir(exist_ok=False)
    cache=ROOT/'.cache/liftrs/s13-m2-t5'
    cache.mkdir(parents=True,exist_ok=True)
    if not args.reuse_cache and (cache/'cyclic.json').exists():raise SystemExit('cache exists; reuse it or select a different experiment')
    policy=json.loads((HERE/'policy.json').read_text())
    sources=sorted((ROOT/'encoding/liftrs-go').glob('*.go'))+[HERE/'main.go',HERE/'run.py',HERE/'policy.json',ROOT/'go.mod',ROOT/'go.sum']
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
        'source_sha256':{str(p.relative_to(ROOT)):sha(p) for p in sources},'policy':policy,'environment':{'platform':platform.platform(),'python':platform.python_version(),'cpu_count':os.cpu_count(),'go':subprocess.check_output([args.go,'version'],text=True).strip()},'status':'running','reused_cache':args.reuse_cache,'cache_directory':str(cache.relative_to(ROOT))}
    def save():(output_dir/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    save()
    with tempfile.TemporaryDirectory(prefix='depir-profile-s13-') as directory:
        binary=Path(directory)/'bench';subprocess.run([args.go,'build','-o',str(binary),'./experiments/encoder-profile-s13'],cwd=ROOT,check=True)
        provenance={'revision':result['revision'],'source_sha256':result['source_sha256']}
        identifier=hashlib.sha256(json.dumps(provenance,sort_keys=True).encode()).hexdigest()
        retained_binary=cache/('profile-bin-'+identifier[:16])
        shutil.copy2(binary,retained_binary)
        result['profile_binary']={'path':str(retained_binary.relative_to(ROOT)),'sha256':sha(retained_binary)}
        if not args.reuse_cache:(cache/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
        result['cache_build_provenance']=json.loads((cache/'provenance.json').read_text())
        command=[str(binary),'-s','13','-t','5','-repeats',str(policy['encoding_repeats']),'-memory-bytes',str(int(policy['memory_gib']*2**30)), '-profiles',str(output_dir),'-cache',str(cache),'-source-manifest',identifier]
        if args.reuse_cache:command.append('-reuse-cache')
        result['command']=['encoder-profile-s13']+command[1:];save()
        events_path=output_dir/'events.jsonl';error_path=output_dir/'stderr.txt';progress_path=output_dir/'progress.jsonl'
        with events_path.open('w') as output,error_path.open('w') as stderr,progress_path.open('w') as progress:
            process=subprocess.Popen(command,cwd=ROOT,stdout=output,stderr=stderr)
            started=time.monotonic();reported=0;last_heartbeat=started-60;phase='starting';phase_start=started;timed_out=False
            print('s=13 profiling started; retained cache, two-hour timeout, 12 GiB Go-memory guard',flush=True)
            try:
                while True:
                    pid,status,usage=os.wait4(process.pid,os.WNOHANG)
                    # The child flushes whole JSON lines; ignore an incomplete final
                    # line while reading concurrently and consume it on the next poll.
                    lines=events_path.read_text().splitlines(keepends=True)
                    events=[json.loads(line) for line in lines if line.endswith('\n')]
                    for event in events[reported:]:
                        if event['event']=='started':phase=event['phase'];phase_start=time.monotonic();print('starting '+phase,flush=True)
                        elif event['event']=='phase':print(f'{event["phase"]}: {event["elapsed_ns"]/1e9:.3f}s',flush=True)
                        elif event['event']=='code':print(describe_code(event),flush=True)
                    reported=len(events)
                    if pid:process.returncode=os.waitstatus_to_exitcode(status);break
                    now=time.monotonic()
                    if now-last_heartbeat>=60:
                        sample={'elapsed_seconds':now-started,'phase':phase,'phase_elapsed_seconds':now-phase_start}
                        progress.write(json.dumps(sample)+'\n');progress.flush();last_heartbeat=now
                        print('progress '+json.dumps(sample),flush=True)
                    if now-started>policy['case_timeout_seconds']:
                        process.kill();_,status,usage=os.wait4(process.pid,0);process.returncode=os.waitstatus_to_exitcode(status);timed_out=True;break
                    time.sleep(.5)
            except BaseException as error:
                if process.returncode is None:
                    process.kill()
                    _,status,usage=os.wait4(process.pid,0)
                    process.returncode=os.waitstatus_to_exitcode(status)
                complete=[json.loads(line) for line in events_path.read_text().splitlines(keepends=True) if line.endswith('\n')]
                result.update(status='monitor_failed',monitor_error=repr(error),returncode=process.returncode,
                    wall_seconds=time.monotonic()-started,process_peak_rss_bytes=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024),
                    events='events.jsonl',events_sha256=sha(events_path),stderr='stderr.txt',stderr_sha256=sha(error_path),progress='progress.jsonl',progress_sha256=sha(progress_path),
                    phases={event['phase']:event for event in complete if event['event']=='phase'})
                save()
                raise
            finally:
                if process.returncode is None:
                    process.kill()
                    process.wait()
        events=[json.loads(line) for line in events_path.read_text().splitlines()]
        result.update(returncode=process.returncode,wall_seconds=time.monotonic()-started,process_peak_rss_bytes=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024),events='events.jsonl',events_sha256=sha(events_path),stderr='stderr.txt',stderr_sha256=sha(error_path),progress='progress.jsonl',progress_sha256=sha(progress_path))
        result['phases']={e['phase']:e for e in events if e['event']=='phase'}
        result['status']='timeout' if timed_out else 'failed' if process.returncode else 'completed'
        codes=[e for e in events if e['event']=='code']
        if codes:result['code']=codes[-1]
        if not process.returncode:
            verified=next(e for e in events if e['event']=='verified');result['verification']=verified
            assert verified['systematic_positions_checked']==result['code']['k']
            assert verified['curve_checks']==16 and verified['packed_bytes']==(result['code']['n']*13+7)//8
            phases=result['phases'];encoding=statistics.median(e['elapsed_ns'] for name,e in phases.items() if name.startswith('encode_'))/1e9
            result['measurement']={'discovery_s':phases.get('discovery',{}).get('elapsed_ns',0)/1e9,'preparation_s':phases.get('preparation',{}).get('elapsed_ns',0)/1e9,'encoding_s':encoding}
            result['measurement']['first_use_s']=sum(result['measurement'].values())
            result['measurement']['cache_save_s']=phases.get('cache_save',{}).get('elapsed_ns',0)/1e9
            result['measurement']['cache_load_s']=phases['cache_load']['elapsed_ns']/1e9
            result['measurement']['workflow_s']=result['measurement']['first_use_s']+result['measurement']['cache_save_s']+result['measurement']['cache_load_s']
            baseline=json.loads((ROOT/'experiments/cyclic-s13/results.json').read_text())
            for key in ['canonical_sha256','packed_sha256']:
                assert verified[key]==baseline['verification'][key],key+' differs from unprofiled baseline'
            result['baseline_match']=True
            metadata=json.loads((cache/'metadata.json').read_text())
            assert sha(cache/'cyclic.json')==metadata['sha256']
            (output_dir/'cache-metadata.json').write_text(json.dumps(metadata,indent=2)+'\n')
            result['cache_metadata']=metadata
            result['cache_build_provenance']=json.loads((cache/'provenance.json').read_text())
            result['profile_sha256']={}
            for profile in sorted(output_dir.glob('*.pprof')):
                result['profile_sha256'][profile.name]=sha(profile)
                for tag,options in [('flat',['-top']),('cum',['-top','-cum']),('tree',['-tree'])]:
                    path=output_dir/(profile.stem+'-'+tag+'.txt')
                    with path.open('w') as output:
                        subprocess.run([args.go,'tool','pprof',*options,str(binary),str(profile)],stdout=output,stderr=subprocess.STDOUT,check=True)
                if profile.stem in ['preparation','encode_0']:
                    with (output_dir/(profile.stem+'-source.txt')).open('w') as output:
                        subprocess.run([args.go,'tool','pprof','-list','recurse|polyMul|constantProduct',str(binary),str(profile)],stdout=output,stderr=subprocess.STDOUT,check=True)

        save();print(json.dumps({'status':result['status'],'wall_seconds':result['wall_seconds'],'peak_gib':result['process_peak_rss_bytes']/2**30,'measurement':result.get('measurement')}),flush=True)
        if process.returncode:raise SystemExit('benchmark did not complete; partial evidence preserved')
if __name__=='__main__':main()
