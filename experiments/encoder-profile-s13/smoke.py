#!/usr/bin/env python3
"""Check both profiling/cache paths on a disposable small correctness fixture."""
import argparse,json,subprocess,tempfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--go',default='go');args=parser.parse_args()
with tempfile.TemporaryDirectory(prefix='depir-profile-smoke-') as directory:
    base=Path(directory);binary=base/'driver';cache=base/'cache'
    subprocess.run([args.go,'build','-o',str(binary),'./experiments/encoder-profile-s13'],cwd=ROOT,check=True)
    checks=[]
    for reuse in [False,True]:
        out=base/('reuse' if reuse else 'setup');out.mkdir()
        command=[str(binary),'-s','4','-t','2','-profiles',str(out),'-cache',str(cache)]
        if reuse:command+=['-reuse-cache']
        result=subprocess.run(command,cwd=ROOT,capture_output=True,text=True,check=True)
        events=[json.loads(line) for line in result.stdout.splitlines()]
        checked=next(e for e in events if e['event']=='verified')
        checks.append({'reuse':reuse,'code':next(e for e in events if e['event']=='code'),'canonical':checked['canonical_sha256'],'profiles':sorted(p.name for p in out.glob('*.pprof'))})
    assert checks[0]['canonical']==checks[1]['canonical']
    (HERE/'smoke-results.json').write_text(json.dumps({'configuration':'s4-m2-t2','checks':checks},indent=2)+'\n')
print('Profile/cache save-load/reuse smoke checks passed.')
