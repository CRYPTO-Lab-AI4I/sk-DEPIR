#!/usr/bin/env python3
"""Measure table-free s=8/9 CPU encoders with isolated preprocessing processes.

No log/exp arithmetic tables are used. Discovery, each method's preparation and
encoding run in independent processes; complete public artifacts are reused.
"""
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

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go',default=os.environ.get('GO','go'))
    args=parser.parse_args()
    artifacts=HERE/'artifacts';artifacts.mkdir(exist_ok=True)
    def invoke(executable,mode,s,t):
        command=[executable,'-mode',mode,'-s',str(s),'-t',str(t),'-artifacts',str(artifacts),'-repeats','3']
        with tempfile.TemporaryFile() as stdout,tempfile.TemporaryFile() as stderr:
            p=subprocess.Popen(command,cwd=ROOT,stdout=stdout,stderr=stderr)
            _,status,usage=os.wait4(p.pid,0);p.returncode=os.waitstatus_to_exitcode(status)
            stdout.seek(0);stderr.seek(0)
            if p.returncode:raise RuntimeError(stderr.read().decode())
            value=json.load(stdout)
        value['process_peak_rss_bytes']=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024)
        return value
    cases=[]
    with tempfile.TemporaryDirectory(prefix='sk-depir-tablefree-build-') as temp:
        executable=str(Path(temp)/'compare')
        subprocess.run([args.go,'build','-o',executable,'./experiments/tablefree-encoding'],cwd=ROOT,check=True)
        for s,t in [(8,8),(9,7)]:
            discovery=invoke(executable,'discover',s,t)
            print(f's={s}: discovered exact K={discovery["k"]}',flush=True)
            methods=[]
            for name in ['paper','cyclic-fft']:
                preparation=invoke(executable,'prepare-'+name,s,t)
                print(f's={s}: {name} prepared in {preparation["prepare_ns"]/1e9:.3f}s',flush=True)
                encoding=invoke(executable,name,s,t)
                methods.append({'preparation':preparation,'encoding':encoding})
                print(f's={s}: {name} median encode {statistics.median(encoding["encode_ns"])/1e9:.3f}s',flush=True)
            assert len({m['encoding']['output_sha256'] for m in methods})==1
            assert len({m['encoding']['packed_sha256'] for m in methods})==1
            assert all(m['encoding']['output_bytes']==(discovery['n']*s+7)//8 for m in methods)
            cases.append({'discovery':discovery,'methods':methods})
    previous=ROOT/'experiments/encoder-comparison/results.json'
    old=json.loads(previous.read_text())['scenarios'][-1]
    assert all(m['encoding']['output_sha256']==old['methods'][0]['output_sha256'] for m in cases[0]['methods'])
    def inspect(command):
        p=subprocess.run(command,capture_output=True,text=True)
        return {'command':command,'exit_code':p.returncode,'stdout':p.stdout.strip(),'stderr':p.stderr.strip()}
    environment={'platform':platform.platform(),'python_version':platform.python_version(),'cpu_count':os.cpu_count(),
                 'processor':platform.processor(),'gomaxprocs':os.environ.get('GOMAXPROCS','default'),
                 'output_directory':tempfile.gettempdir(),'cache_policy':'No cache flush; final output sync included; readback is cached.'}
    if platform.system()=='Darwin':
        environment['hardware']=inspect(['/usr/sbin/system_profiler','SPHardwareDataType','-detailLevel','mini'])
        # Only retain hardware model/CPU/memory; do not publish serial numbers or UUIDs.
        environment['hardware']['stdout']='\n'.join(line for line in environment['hardware']['stdout'].splitlines()
            if any(label in line for label in ['Model Name:','Model Identifier:','Chip:','Processor Name:','Total Number of Cores:','Memory:']))
        environment['filesystem']=inspect(['/bin/df','-h',tempfile.gettempdir()])
        environment['storage']=inspect(['/usr/sbin/diskutil','info',tempfile.gettempdir()])
        # Device/volume UUIDs and mount-specific identifiers are not needed.
        environment['storage']['stdout']='\n'.join(line for line in environment['storage']['stdout'].splitlines()
            if any(label in line for label in ['File System Personality:','Type (Bundle):','Protocol:','Device / Media Name:','Solid State:','Disk Size:','Volume Total Space:']))
    else:
        environment['hardware']=inspect(['uname','-a'])
        environment['filesystem']=inspect(['df','-T',tempfile.gettempdir()])
    sources=list((ROOT/'encoding/liftrs-go').glob('*.go'))+[HERE/'main.go',HERE/'main_test.go',HERE/'run.py',ROOT/'go.mod',previous]
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),
            'base_revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
            'working_tree_dirty':bool(subprocess.check_output(['git','status','--porcelain'],cwd=ROOT,text=True).strip()),
            'environment':environment,'source_sha256':{str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sources},
            'artifact_sha256':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(artifacts.glob('*.json'))},'scenarios':cases}
    (HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    lines=['# Table-free encoding checkpoint','','Times measured on CPU; median of three encoding calls. All preparation is code-only and excluded from encoding.', '',
           '| s / t | N / exact K | Method | Discovery ms | Preparation ms | Artifact KiB | Load ms | Encode ms | Preparation / encoding-process RSS MiB |',
           '|---|---:|---|---:|---:|---:|---:|---:|---:|']
    for case in cases:
        d=case['discovery']
        for m in case['methods']:
            p,e=m['preparation'],m['encoding']
            lines.append(f"| {d['s']} / {d['t']} | {d['n']:,} / {d['k']:,} | {p['method']} | {d['discovery_ns']/1e6:.2f} | {p['prepare_ns']/1e6:.2f} | {p['artifact_bytes']/1024:.1f} | {e['artifact_load_ns']/1e6:.2f} | {statistics.median(e['encode_ns'])/1e6:.2f} | {p['process_peak_rss_bytes']/2**20:.2f} / {e['process_peak_rss_bytes']/2**20:.2f} |")
    (HERE/'table.md').write_text('\n'.join(lines)+'\n')

if __name__=='__main__':main()
