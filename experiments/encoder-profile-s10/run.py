#!/usr/bin/env python3
"""Profile s=10 cyclic preparation and database encoding in separate processes."""
import argparse
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import platform
import subprocess
import tempfile

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
INPUT=ROOT/'experiments/encoder-scaling'

def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go',default='go')
    parser.add_argument('--seconds',type=float,default=10)
    args=parser.parse_args()
    if args.seconds<=0:parser.error('--seconds must be positive')
    def command(args):return subprocess.check_output([str(x) for x in args],cwd=ROOT,text=True)
    def inspect(args):
        p=subprocess.run(args,capture_output=True,text=True)
        return {'command':args,'exit_code':p.returncode,'stdout':p.stdout.strip(),'stderr':p.stderr.strip()}
    environment={'platform':platform.platform(),'python':platform.python_version(),'cpu_count':os.cpu_count(),
                 'go':command([args.go,'version']).strip(),'gomaxprocs_environment':os.environ.get('GOMAXPROCS','default')}
    if platform.system()=='Darwin':
        hardware=inspect(['/usr/sbin/system_profiler','SPHardwareDataType','-detailLevel','mini'])
        hardware['stdout']='\n'.join(s for s in hardware['stdout'].splitlines() if any(k in s for k in ['Model Name:','Model Identifier:','Chip:','Total Number of Cores:','Memory:']))
        environment['hardware']=hardware
    paths=sorted((ROOT/'encoding/liftrs').glob('*.go'))+[ROOT/'experiments/encoder-profile/main.go',HERE/'run.py',ROOT/'go.mod',ROOT/'go.sum']
    inputs=[INPUT/'results.json',INPUT/'artifacts/s10-m2-t6-basis.json',INPUT/'artifacts/s10-m2-t6-cyclic-fft.json']
    result={'timestamp_utc':datetime.now(timezone.utc).isoformat(),'revision':command(['git','rev-parse','HEAD']).strip(),
            'working_tree_dirty':bool(command(['git','status','--porcelain']).strip()),'environment':environment,
            'minimum_seconds':args.seconds,'source_sha256':{str(p.relative_to(ROOT)):digest(p) for p in paths},
            'input_sha256':{str(p.relative_to(ROOT)):digest(p) for p in inputs},'profiles':[]}
    checkpoint=json.loads((INPUT/'results.json').read_text())
    expected=next(m['encoding']['output_sha256'] for m in checkpoint['methods'] if m['method']=='cyclic-fft')
    with tempfile.TemporaryDirectory(prefix='depir-s10-profile-') as directory:
        binary=Path(directory)/'profile'
        command([args.go,'build','-o',binary,'./experiments/encoder-profile'])
        for stage in ['prepare','encode']:
            name=stage+'-cyclic-fft';profile=HERE/(name+'.pprof')
            measured=json.loads(command([binary,'-s','10','-t','6','-stage',stage,'-method','cyclic-fft',
                                         '-artifacts',INPUT/'artifacts','-duration',str(args.seconds)+'s','-profile',profile]))
            if measured['s']!=10 or measured['t']!=6 or measured['k']!=103556 or measured['n']!=1048576:
                raise RuntimeError('configuration differs from checkpoint')
            if stage=='encode' and measured['sha256']!=expected:raise RuntimeError('codeword differs from checkpoint')
            views={}
            for order in ['flat','cum']:
                view=HERE/(name+'-'+order+'.txt')
                view.write_text(command([args.go,'tool','pprof','-top','-nodecount=40','-'+order,binary,profile]))
                views[view.name]=digest(view)
            # Source line attribution separates FFTs in the prefix helper from
            # the two transforms directly called by the fixed-generator product.
            source_pattern='PrepareCyclic' if stage=='prepare' else 'CyclicEncoder.*Encode'
            for label,flags in [('tree',['-tree','-nodecount=50']),('source',['-list',source_pattern])]:
                view=HERE/(name+'-'+label+'.txt')
                view.write_text(command([args.go,'tool','pprof',*flags,binary,profile]))
                views[view.name]=digest(view)
            measured['profile']=profile.name;measured['profile_sha256']=digest(profile);measured['views_sha256']=views
            result['profiles'].append(measured)
            print(f'{stage}: {measured["calls"]} calls, {measured["elapsed_ns"]/1e9:.3f}s sampled interval',flush=True)
    (HERE/'results.json').write_text(json.dumps(result,indent=2)+'\n')

if __name__=='__main__':main()
