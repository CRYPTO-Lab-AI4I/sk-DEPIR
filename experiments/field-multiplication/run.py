#!/usr/bin/env python3
"""Reproduce portable kernels and cyclic scaling; preserve the historical baseline."""
import argparse,hashlib,json,os,platform,statistics,subprocess,tempfile,time
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent

def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go',default='go')
    parser.add_argument('--stage',choices=['kernels','scaling'],required=True)
    args=parser.parse_args()
    sources=(sorted((ROOT/'encoding/liftrs-shared').rglob('*.go'))+list((ROOT/'encoding/liftrs-go').glob('*.go')))+[HERE/'main.go',HERE/'run.py',ROOT/'experiments/cyclic-scaling/main.go',ROOT/'go.mod',ROOT/'go.sum']
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
        'source_sha256':{str(p.relative_to(ROOT)):digest(p) for p in sources},'environment':{'platform':platform.platform(),'cpu_count':os.cpu_count(),'python':platform.python_version(),'go':subprocess.check_output([args.go,'version'],text=True).strip()},'stage':args.stage}
    with tempfile.TemporaryDirectory(prefix='depir-mul-') as temp:
        binary=Path(temp)/'bench'
        if args.stage=='kernels':
            subprocess.run([args.go,'build','-o',str(binary),'./experiments/field-multiplication'],cwd=ROOT,check=True)
            path=HERE/'kernels-after.json'
            with path.open('w') as output:subprocess.run([str(binary)],cwd=ROOT,stdout=output,check=True)
            result.update(results=path.name,results_sha256=digest(path))
        else:
            subprocess.run([args.go,'build','-o',str(binary),'./experiments/cyclic-scaling'],cwd=ROOT,check=True)
            result['baseline_sha256']=digest(ROOT/'experiments/cyclic-scaling/results.json')
            result['cases']=[]
            for s,t in [(9,7),(10,6),(11,6),(12,6)]:
                path=HERE/f's{s}-events.jsonl';error=HERE/f's{s}-stderr.txt'
                print(f's={s} started',flush=True)
                with path.open('w') as output,error.open('w') as stderr:
                    proc=subprocess.Popen([str(binary),'-s',str(s),'-t',str(t),'-repeats','3','-memory-bytes',str(12*2**30)],cwd=ROOT,stdout=output,stderr=stderr)
                    started=time.monotonic();reported=0
                    while True:
                        pid,status,usage=os.wait4(proc.pid,os.WNOHANG)
                        events=[json.loads(line) for line in path.read_text().splitlines()]
                        for event in events[reported:]:
                            if event['event']=='phase':print(f's={s} {event["phase"]}: {event["elapsed_ns"]/1e9:.3f}s',flush=True)
                        reported=len(events)
                        if pid:proc.returncode=os.waitstatus_to_exitcode(status);break
                        if time.monotonic()-started>3600:
                            proc.kill();os.wait4(proc.pid,0);raise RuntimeError('one-hour whole-case timeout')
                        time.sleep(.5)
                if proc.returncode:raise RuntimeError(f's={s}: exit {proc.returncode}')
                code=next(e for e in events if e['event']=='code');verified=next(e for e in events if e['event']=='verified')
                phases={e['phase']:e for e in events if e['event']=='phase'}
                case={'s':s,'t':t,'k':code['k'],'n':code['n'],'events':path.name,'events_sha256':digest(path),'stderr':error.name,'wall_seconds':time.monotonic()-started,
                    'peak_bytes':usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024),'discovery_s':phases['discovery']['elapsed_ns']/1e9,
                    'preparation_s':phases['preparation']['elapsed_ns']/1e9,'encoding_s':statistics.median(e['elapsed_ns'] for name,e in phases.items() if name.startswith('encode_'))/1e9,'verification':verified}
                assert verified['systematic_positions_checked']==case['k'] and verified['curve_checks']==16
                result['cases'].append(case)
                (HERE/'scaling-results.json').write_text(json.dumps(result,indent=2)+'\n')
        (HERE/(args.stage+'-provenance.json')).write_text(json.dumps(result,indent=2)+'\n')
if __name__=='__main__':main()
