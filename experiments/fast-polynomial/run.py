#!/usr/bin/env python3
"""Compare hybrid polynomial products with the committed Karatsuba checkpoint."""
import argparse
from datetime import datetime, timezone
import hashlib
import io
import json
import os
from pathlib import Path
import platform
import shutil
import statistics
import subprocess
import tarfile
import tempfile

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent

def digest(p): return hashlib.sha256(p.read_bytes()).hexdigest()

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go',default='go')
    parser.add_argument('--baseline',default='7e5a8ac')
    args=parser.parse_args()
    def command(cmd,cwd=ROOT): return subprocess.check_output([str(x) for x in cmd],cwd=cwd,text=True)
    def inspect(cmd):
        p=subprocess.run(cmd,capture_output=True,text=True)
        return {'command':cmd,'exit_code':p.returncode,'stdout':p.stdout.strip(),'stderr':p.stderr.strip()}
    baseline=command(['git','rev-parse',args.baseline]).strip()
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),
            'revision':command(['git','rev-parse','HEAD']).strip(),
            'working_tree_dirty':bool(command(['git','status','--porcelain']).strip()),
            'baseline_revision':baseline,
            'environment':{'platform':platform.platform(),'python':platform.python_version(),
                           'cpu_count':os.cpu_count(),'go':command([args.go,'version']).strip(),
                           'gomaxprocs':os.environ.get('GOMAXPROCS','default')},'scenarios':[]}
    if platform.system()=='Darwin':
        hardware=inspect(['/usr/sbin/system_profiler','SPHardwareDataType','-detailLevel','mini'])
        hardware['stdout']='\n'.join(s for s in hardware['stdout'].splitlines() if any(k in s for k in ['Model Name:','Model Identifier:','Chip:','Total Number of Cores:','Memory:']))
        result['environment']['hardware']=hardware
    result['environment']['filesystem']=inspect(['df','-h',tempfile.gettempdir()])
    checkpoints=ROOT/'experiments/tablefree-encoding/artifacts'
    inputs=sorted(checkpoints.glob('*.json'))
    result['input_sha256']={str(p.relative_to(ROOT)):digest(p) for p in inputs}
    with tempfile.TemporaryDirectory(prefix='depir-polynomial-') as directory:
        temp=Path(directory);old=temp/'old';old.mkdir()
        archive=subprocess.check_output(['git','archive',baseline,'go.mod','go.sum','encoding/liftrs-go','experiments/tablefree-encoding'],cwd=ROOT)
        # Archive contains our own committed repository paths, not external input.
        with tarfile.open(fileobj=io.BytesIO(archive)) as tar:tar.extractall(old)
        binaries={}
        for backend,source in [('old',old),('new',ROOT)]:
            binary=temp/backend/'compare' if backend=='old' else temp/'compare-new'
            command([args.go,'build','-o',binary,'./experiments/tablefree-encoding'],cwd=source)
            binaries[backend]=binary
        def invoke(binary,mode,s,t,artifacts):
            with tempfile.TemporaryFile() as stdout,tempfile.TemporaryFile() as stderr:
                p=subprocess.Popen([str(binary),'-mode',mode,'-s',str(s),'-t',str(t),'-artifacts',str(artifacts),'-repeats','3'],cwd=ROOT,stdout=stdout,stderr=stderr)
                _,status,usage=os.wait4(p.pid,0);p.returncode=os.waitstatus_to_exitcode(status)
                stdout.seek(0);stderr.seek(0)
                if p.returncode:raise RuntimeError(stderr.read().decode())
                measured=json.load(stdout)
            measured['process_peak_rss_bytes']=usage.ru_maxrss*(1 if platform.system()=='Darwin' else 1024)
            return measured
        for s,t in [(8,8),(9,7)]:
            for method in ['paper','cyclic-fft']:
                case={'s':s,'t':t,'m':2,'method':method,'backends':{}}
                for backend in (['old','new'] if s==8 else ['new','old']):
                    artifacts=temp/(backend+'-artifacts');artifacts.mkdir(exist_ok=True)
                    stem=f's{s:02d}-m2-t{t}'
                    shutil.copyfile(checkpoints/(stem+'-basis.json'),artifacts/(stem+'-basis.json'))
                    prep=invoke(binaries[backend],'prepare-'+method,s,t,artifacts)
                    artifact=artifacts/(stem+'-'+method+'.json')
                    assert digest(artifact)==digest(checkpoints/artifact.name),'artifact changed'
                    enc=invoke(binaries[backend],method,s,t,artifacts)
                    case['backends'][backend]={'preparation':prep,'encoding':enc,'artifact_sha256':digest(artifact)}
                    print(f's={s} {method} {backend}: prep {prep["prepare_ns"]/1e6:.1f}ms, encode {statistics.median(enc["encode_ns"])/1e6:.1f}ms',flush=True)
                for key in ['output_sha256','packed_sha256','output_bytes']:
                    assert case['backends']['old']['encoding'][key]==case['backends']['new']['encoding'][key]
                result['scenarios'].append(case)
    # Benchmark inputs and immutable transform construction are outside each timed
    # kernel; old is the retained full Karatsuba-and-slice path in the test driver.
    kernels=command([args.go,'test','./encoding/liftrs-go','-run','^$','-bench','^BenchmarkPolynomialKernels$','-benchtime=1x','-count=3'])
    (HERE/'kernels.txt').write_text(kernels)
    result['kernels_sha256']=digest(HERE/'kernels.txt')
    sources=sorted((ROOT/'encoding/liftrs-go').glob('*.go'))+[HERE/'run.py',ROOT/'experiments/tablefree-encoding/main.go',ROOT/'go.mod',ROOT/'go.sum']
    result['source_sha256']={str(p.relative_to(ROOT)):digest(p) for p in sources}
    (HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')
    lines=['# Fast polynomial multiplication: encoder comparison','','Same-machine runs; median of three per-database encodings. Reusable preparation is separate.','','| s / t | Method | Preparation old → new (ms) | Encode old → new (ms) | Encode speedup | Allocated per encode old → new (MiB) | Encode-process RSS old → new (MiB) |','|---|---|---:|---:|---:|---:|---:|']
    for c in result['scenarios']:
        old,new=c['backends']['old'],c['backends']['new'];oe,ne=old['encoding'],new['encoding']
        a,b=statistics.median(oe['encode_ns']),statistics.median(ne['encode_ns'])
        lines.append(f'| {c["s"]} / {c["t"]} | {c["method"]} | {old["preparation"]["prepare_ns"]/1e6:.2f} → {new["preparation"]["prepare_ns"]/1e6:.2f} | {a/1e6:.2f} → {b/1e6:.2f} | {a/b:.2f}× | {statistics.median(oe["encode_allocated_bytes"])/2**20:.2f} → {statistics.median(ne["encode_allocated_bytes"])/2**20:.2f} | {oe["process_peak_rss_bytes"]/2**20:.2f} → {ne["process_peak_rss_bytes"]/2**20:.2f} |')
    (HERE/'table.md').write_text('\n'.join(lines)+'\n')

if __name__=='__main__':main()
