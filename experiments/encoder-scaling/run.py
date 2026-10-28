#!/usr/bin/env python3
"""Measure s=10 preprocessing, encoding, isolated heap samples and packed I/O."""
import argparse
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import platform
import statistics
import subprocess
import tempfile

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent

def digest(path):return hashlib.sha256(path.read_bytes()).hexdigest()

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go',default='go')
    args=parser.parse_args()
    artifacts=HERE/'artifacts';artifacts.mkdir(exist_ok=True)
    def inspect(command):
        p=subprocess.run(command,capture_output=True,text=True)
        return {'command':command,'exit_code':p.returncode,'stdout':p.stdout.strip(),'stderr':p.stderr.strip()}
    environment={'platform':platform.platform(),'python_version':platform.python_version(),
                 'cpu_count':os.cpu_count(),'go':subprocess.check_output([args.go,'version'],text=True).strip(),
                 'gomaxprocs':os.environ.get('GOMAXPROCS','default'),'filesystem':inspect(['df','-h',tempfile.gettempdir()]),
                 'output_directory':tempfile.gettempdir(),'cache_policy':'No flush; packed output sync included and readback is cached.'}
    if platform.system()=='Darwin':
        environment['hardware']=inspect(['/usr/sbin/system_profiler','SPHardwareDataType','-detailLevel','mini'])
        environment['hardware']['stdout']='\n'.join(s for s in environment['hardware']['stdout'].splitlines() if any(k in s for k in ['Model Name:','Model Identifier:','Chip:','Total Number of Cores:','Memory:']))
        environment['storage']=inspect(['/usr/sbin/diskutil','info',tempfile.gettempdir()])
        environment['storage']['stdout']='\n'.join(s for s in environment['storage']['stdout'].splitlines() if any(k in s for k in ['File System Personality:','Type (Bundle):','Protocol:','Device / Media Name:','Solid State:','Disk Size:','Volume Total Space:']))
    def invoke(command):
        with tempfile.TemporaryFile() as stdout,tempfile.TemporaryFile() as stderr:
            p=subprocess.Popen([str(x) for x in command],cwd=ROOT,stdout=stdout,stderr=stderr)
            _,status,usage=os.wait4(p.pid,0);p.returncode=os.waitstatus_to_exitcode(status)
            stdout.seek(0);stderr.seek(0)
            if p.returncode:raise RuntimeError(stderr.read().decode())
            value=json.load(stdout)
        value['process_peak_rss_bytes']=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024)
        return value
    with tempfile.TemporaryDirectory(prefix='depir-scaling-build-') as directory:
        binary,audit=Path(directory)/'stages',Path(directory)/'audit'
        subprocess.run([args.go,'build','-o',str(binary),'./experiments/tablefree-encoding'],cwd=ROOT,check=True)
        subprocess.run([args.go,'build','-o',str(audit),'./experiments/encoder-scaling'],cwd=ROOT,check=True)
        def stage(mode):return invoke([binary,'-mode',mode,'-s','10','-t','6','-artifacts',artifacts,'-repeats','3'])
        discovery=stage('discover');print(f'Discovery: K={discovery["k"]}, {discovery["discovery_ns"]/1e9:.3f}s',flush=True)
        methods=[]
        for method in ['cyclic-fft','paper']:
            prep=stage('prepare-'+method);print(f'{method} preparation {prep["prepare_ns"]/1e9:.3f}s; artifact {prep["artifact_bytes"]/2**20:.2f}MiB',flush=True)
            encoding=stage(method);print(f'{method} median encode {statistics.median(encoding["encode_ns"])/1e9:.3f}s',flush=True)
            memory=invoke([audit,'-mode','memory','-method',method,'-artifacts',artifacts])
            methods.append({'method':method,'preparation':prep,'encoding':encoding,'memory':memory})
        verification=invoke([audit,'-mode','verify','-artifacts',artifacts])
        assert verification['k']==discovery['k']
        assert len({x['encoding']['output_sha256'] for x in methods})==1
        assert len({x['encoding']['packed_sha256'] for x in methods})==1
        assert len({x['memory']['sha256'] for x in methods} | {verification['sha256']})==1
        assert all(x['encoding']['output_bytes']==1048576*10//8 for x in methods)
    sources=sorted((ROOT/'encoding/liftrs-go').glob('*.go'))+[HERE/'main.go',HERE/'main_test.go',HERE/'run.py',ROOT/'experiments/tablefree-encoding/main.go',ROOT/'go.mod',ROOT/'go.sum']
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'working_tree_dirty':bool(subprocess.check_output(['git','status','--porcelain'],cwd=ROOT,text=True).strip()),
            'environment':environment,'source_sha256':{str(p.relative_to(ROOT)):digest(p) for p in sources},
            'artifact_sha256':{p.name:digest(p) for p in sorted(artifacts.glob('*.json'))},
            'discovery':discovery,'methods':methods,'verification':verification}
    (HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    lines=['# CPU encoding at s=10,m=2,t=6','','Exact N=1,048,576; K='+str(discovery['k'])+'. Times are CPU-host wall times; encode is the median of three calls.','','| Method | Preparation s | Artifact MiB | Load ms | Encode s | Allocated MiB/call | Preparation / encoding process RSS MiB | Sampled encoding heap increase MiB |','|---|---:|---:|---:|---:|---:|---:|---:|']
    for m in methods:
        p,e,h=m['preparation'],m['encoding'],m['memory']
        lines.append(f'| {m["method"]} | {p["prepare_ns"]/1e9:.3f} | {p["artifact_bytes"]/2**20:.2f} | {e["artifact_load_ns"]/1e6:.2f} | {statistics.median(e["encode_ns"])/1e9:.3f} | {statistics.median(e["encode_allocated_bytes"])/2**20:.2f} | {p["process_peak_rss_bytes"]/2**20:.2f} / {e["process_peak_rss_bytes"]/2**20:.2f} | {h["sampled_incremental_heap_alloc_bytes"]/2**20:.2f} |')
    lines.extend(['','Heap sampling is a separate instrumented call, includes its 4 MiB output, and can miss peaks. RSS includes the whole stage process.'])
    (HERE/'table.md').write_text('\n'.join(lines)+'\n')

if __name__=='__main__':main()
