#!/usr/bin/env python3
"""Progressively benchmark cyclic encoders admitted by runtime/RAM forecasts."""
import argparse
from datetime import datetime,timezone
import hashlib,json,math,os,platform,statistics,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent

def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def model(case,anchor):
    s,n,k=case['s'],case['n'],case['k']
    a=anchor
    transform=n/a['n']*(s/a['s'])**3
    tree=k/a['k']*(math.log2(k)/math.log2(a['k']))**3*(s/a['s'])
    return {'discovery_s':a['discovery_s']*transform,'preparation_s':a['preparation_s']*max(transform,tree),
            'encoding_s':a['encoding_s']*transform,'peak_bytes':a['peak_bytes']*max(n/a['n'],k/a['k'])}

def main():
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--go',default='go');args=parser.parse_args()
    checkpoint=json.loads((ROOT/'experiments/encoder-scaling/results.json').read_text());method=next(x for x in checkpoint['methods'] if x['method']=='cyclic-fft')
    anchor={'s':10,'n':1048576,'k':103556,'discovery_s':checkpoint['discovery']['discovery_ns']/1e9,'preparation_s':method['preparation']['prepare_ns']/1e9,
            'encoding_s':statistics.median(method['encoding']['encode_ns'])/1e9,'peak_bytes':max(checkpoint['discovery']['process_peak_rss_bytes'],method['preparation']['process_peak_rss_bytes'],method['encoding']['process_peak_rss_bytes'])}
    catalogue=json.loads((ROOT/'pirconfig/configs_catalogue.json').read_text())
    sources=(sorted((ROOT/'encoding/liftrs-shared').rglob('*.go'))+list((ROOT/'encoding/liftrs-go').glob('*.go')))+[HERE/'main.go',HERE/'run.py',ROOT/'go.mod',ROOT/'go.sum']
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'source_sha256':{str(p.relative_to(ROOT)):digest(p) for p in sources},'seed_checkpoint_sha256':digest(ROOT/'experiments/encoder-scaling/results.json'),
            'environment':{'platform':platform.platform(),'python':platform.python_version(),'cpu_count':os.cpu_count(),'go':subprocess.check_output([args.go,'version'],text=True).strip()},'initial_anchor':anchor,'cases':[]}
    def save():
        (HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
        lines=['# Progressive cyclic scaling','','Forecasts include the configured safety factor; measured times do not. All memory figures are stage-process RSS.','','| s / t | Status | K | Discovery | Preparation | Encoding | Peak GiB |','|---|---|---:|---:|---:|---:|---:|']
        for case in result['cases']:
            if case['status']=='measured':
                x=case['measurement'];d,p,e=x['discovery_s'],x['preparation_s'],x['encoding_s'];mem=x['peak_bytes']/2**30
            else:
                x=case['forecast'];d,p,e=x['discovery_s'],x['preparation_s'],x['encoding_s'];mem=x['peak_bytes']/2**30
            lines.append(f'| {case["s"]} / {case["t"]} | {case["status"]} | {case["k"]:,} | {d:.2f} s | {p:.2f} s | {e:.2f} s | {mem:.2f} |')
        (HERE/'table.md').write_text('\n'.join(lines)+'\n')
    with tempfile.TemporaryDirectory(prefix='depir-cyclic-scaling-') as directory:
        binary=Path(directory)/'bench';subprocess.run([args.go,'build','-o',str(binary),'./experiments/cyclic-scaling'],cwd=ROOT,check=True)
        for entry in catalogue['configurations']:
            s=entry['code']['s']
            if s<=10:continue
            case={'s':s,'t':entry['code']['t'],'n':entry['capacity']['n'],'k':entry['capacity']['k'],'capacity_kind':'catalogue_estimate','anchor':dict(anchor)}
            policy=json.loads((HERE/'policy.json').read_text());case['policy']=policy
            forecast=model(case,anchor);forecast={key:value*policy['safety_factor'] for key,value in forecast.items()};case['forecast']=forecast
            runtime=forecast['encoding_s'] if policy['runtime_metric']=='encoding' else sum(forecast[key] for key in ['discovery_s','preparation_s','encoding_s'])
            reasons=[]
            if runtime>=policy['seconds']:reasons.append('runtime_forecast')
            if forecast['peak_bytes']>policy['memory_gib']*2**30:reasons.append('memory_forecast')
            if s>13:reasons.append('packed_backend_bound')
            if reasons:
                case.update(status='estimated_not_run',reasons=reasons);result['cases'].append(case);save();continue
            print(f's={s} admitted: forecast first-use {sum(forecast[k] for k in ["discovery_s","preparation_s","encoding_s"])/60:.1f}min, RSS {forecast["peak_bytes"]/2**30:.2f}GiB',flush=True)
            path=HERE/f's{s:02d}-events.jsonl';error=HERE/f's{s:02d}-stderr.txt'
            with path.open('w+') as output,error.open('w+') as stderr:
                process=subprocess.Popen([str(binary),'-s',str(s),'-t',str(case['t']),'-repeats',str(policy['encoding_repeats']),'-memory-bytes',str(int(policy['memory_gib']*2**30))],cwd=ROOT,stdout=output,stderr=stderr)
                started=time.monotonic();reported=0;timed_out=False
                while True:
                    pid,status,usage=os.wait4(process.pid,os.WNOHANG)
                    rows=path.read_text().splitlines()
                    for row in rows[reported:]:
                        event=json.loads(row)
                        if event['event']=='phase':print(f's={s} {event["phase"]}: {event["elapsed_ns"]/1e9:.2f}s',flush=True)
                        elif event['event']=='started':print(f's={s} starting {event["phase"]}',flush=True)
                    reported=len(rows)
                    if pid:process.returncode=os.waitstatus_to_exitcode(status);break
                    # A hard one-hour case timeout avoids surprise multi-hour runs;
                    # the policy's forecast metric and this operational cap differ.
                    if time.monotonic()-started>policy['seconds']:
                        process.kill();_,status,usage=os.wait4(process.pid,0);process.returncode=os.waitstatus_to_exitcode(status);timed_out=True;break
                    time.sleep(0.5)
            events=[json.loads(row) for row in path.read_text().splitlines()]
            case.update(events=path.name,events_sha256=digest(path),stderr=error.name,returncode=process.returncode,wall_seconds=time.monotonic()-started,
                        process_peak_rss_bytes=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024))
            if process.returncode:
                case.update(status='timeout' if timed_out else 'failed');result['cases'].append(case);save();break
            code=next(x for x in events if x['event']=='code');verified=next(x for x in events if x['event']=='verified');phases={x['phase']:x for x in events if x['event']=='phase'}
            assert verified['k']==code['k'] and verified['systematic_positions_checked']==code['k']
            assert verified['packed_bytes']==(code['n']*s+7)//8 and verified['curve_checks']==16
            measurement={'s':s,'n':code['n'],'k':code['k'],'discovery_s':phases['discovery']['elapsed_ns']/1e9,'preparation_s':phases['preparation']['elapsed_ns']/1e9,
                         'encoding_s':statistics.median(x['elapsed_ns'] for name,x in phases.items() if name.startswith('encode_'))/1e9,'peak_bytes':case['process_peak_rss_bytes']}
            case.update(status='measured',k=code['k'],capacity_kind='exact_enumeration',measurement=measurement,verification=verified)
            result['cases'].append(case);save();anchor=measurement
            print(f's={s} complete: exact K={code["k"]}, RSS={measurement["peak_bytes"]/2**30:.2f}GiB',flush=True)
    save()

if __name__=='__main__':main()
