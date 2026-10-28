#!/usr/bin/env python3
"""Benchmark s=13 under the user's two-hour allowance, preserving partial results."""
import argparse,hashlib,json,os,platform,statistics,subprocess,tempfile,time
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--go',default='go');args=parser.parse_args()
    policy=json.loads((HERE/'policy.json').read_text())
    sources=sorted((ROOT/'encoding/liftrs').glob('*.go'))+[ROOT/'experiments/cyclic-scaling/main.go',HERE/'run.py',HERE/'policy.json',ROOT/'go.mod',ROOT/'go.sum']
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
        'source_sha256':{str(p.relative_to(ROOT)):sha(p) for p in sources},'policy':policy,'environment':{'platform':platform.platform(),'python':platform.python_version(),'cpu_count':os.cpu_count(),'go':subprocess.check_output([args.go,'version'],text=True).strip()},'status':'running'}
    def save():(HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    save()
    with tempfile.TemporaryDirectory(prefix='depir-s13-') as directory:
        binary=Path(directory)/'bench';subprocess.run([args.go,'build','-o',str(binary),'./experiments/cyclic-scaling'],cwd=ROOT,check=True)
        command=[str(binary),'-s','13','-t','5','-repeats',str(policy['encoding_repeats']),'-memory-bytes',str(int(policy['memory_gib']*2**30))]
        result['command']=['cyclic-scaling']+command[1:];save()
        events_path=HERE/'events.jsonl';error_path=HERE/'stderr.txt';progress_path=HERE/'progress.jsonl'
        with events_path.open('w') as output,error_path.open('w') as stderr,progress_path.open('w') as progress:
            process=subprocess.Popen(command,cwd=ROOT,stdout=output,stderr=stderr)
            started=time.monotonic();reported=0;last_heartbeat=started-60;phase='starting';phase_start=started;timed_out=False
            print('s=13 started; two-hour case timeout, 12 GiB Go-memory guard',flush=True)
            while True:
                pid,status,usage=os.wait4(process.pid,os.WNOHANG)
                # The child flushes whole JSON lines; ignore an incomplete final
                # line while reading concurrently and consume it on the next poll.
                lines=events_path.read_text().splitlines(keepends=True)
                events=[json.loads(line) for line in lines if line.endswith('\n')]
                for event in events[reported:]:
                    if event['event']=='started':phase=event['phase'];phase_start=time.monotonic();print('starting '+phase,flush=True)
                    elif event['event']=='phase':print(f'{event["phase"]}: {event["elapsed_ns"]/1e9:.3f}s',flush=True)
                    elif event['event']=='code':print(f'exact K={event["k"]}, samples={event["samples"]}',flush=True)
                reported=len(events)
                if pid:process.returncode=os.waitstatus_to_exitcode(status);break
                now=time.monotonic()
                if now-last_heartbeat>=60:
                    rss=subprocess.run(['ps','-o','rss=','-p',str(process.pid)],capture_output=True,text=True)
                    sample={'elapsed_seconds':now-started,'phase':phase,'phase_elapsed_seconds':now-phase_start,'rss_bytes':int(rss.stdout.strip())*1024 if rss.returncode==0 and rss.stdout.strip() else None}
                    progress.write(json.dumps(sample)+'\n');progress.flush();last_heartbeat=now
                    print('progress '+json.dumps(sample),flush=True)
                if now-started>policy['case_timeout_seconds']:
                    process.kill();_,status,usage=os.wait4(process.pid,0);process.returncode=os.waitstatus_to_exitcode(status);timed_out=True;break
                time.sleep(.5)
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
            result['measurement']={'discovery_s':phases['discovery']['elapsed_ns']/1e9,'preparation_s':phases['preparation']['elapsed_ns']/1e9,'encoding_s':encoding}
            result['measurement']['first_use_s']=sum(result['measurement'].values())
        save();print(json.dumps({'status':result['status'],'wall_seconds':result['wall_seconds'],'peak_gib':result['process_peak_rss_bytes']/2**30,'measurement':result.get('measurement')}),flush=True)
        if process.returncode:raise SystemExit('benchmark did not complete; partial evidence preserved')
if __name__=='__main__':main()
