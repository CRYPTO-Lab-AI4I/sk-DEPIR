#!/usr/bin/env python3
"""Run new-loader tests, profiling smoke checks, vet and an isolated full suite."""
import argparse,json,shutil,subprocess,sys,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');args=p.parse_args();checks=[]
with (HERE/'verification.log').open('w') as log:
    def run(command,cwd=ROOT):
        log.write('$ '+' '.join(command)+'\n');log.flush();start=time.monotonic()
        r=subprocess.run(command,cwd=cwd,stdout=log,stderr=subprocess.STDOUT)
        checks.append({'command':command,'isolated_copy':cwd!=ROOT,'exit_code':r.returncode,'seconds':time.monotonic()-start})
        (HERE/'verification.json').write_text(json.dumps(checks,indent=2)+'\n')
        if r.returncode:raise SystemExit('check failed; see verification.log')
    run([args.go,'test','./encoding/liftrs','./experiments/encoder-profile-s13','-count=1'])
    run([sys.executable,str(HERE/'smoke.py'),'--go',args.go])
    run([args.go,'vet','./encoding/liftrs','./experiments/encoder-profile-s13'])
    with tempfile.TemporaryDirectory(prefix='depir-profile-tests-') as directory:
        copy=Path(directory)/'repo'
        shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','.cache','__pycache__','.DS_Store'))
        run([args.go,'test','./...'],cwd=copy)
print('Loader tests, smoke checks, vet and isolated full suite passed.')
