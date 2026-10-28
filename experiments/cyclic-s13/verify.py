#!/usr/bin/env python3
"""Validate the recorded catalogue and run Go checks after timing completes."""
import argparse,json,shutil,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--go',default='go');args=parser.parse_args()
assert json.loads((HERE/'results.json').read_text())['status']=='completed'
checks=[]
with (HERE/'verification.log').open('w') as log:
    def run(command,cwd=ROOT):
        log.write('$ '+' '.join(command)+'\n');log.flush();start=time.monotonic()
        result=subprocess.run(command,cwd=cwd,stdout=log,stderr=subprocess.STDOUT)
        checks.append({'command':command,'isolated_copy':cwd!=ROOT,'exit_code':result.returncode,'seconds':time.monotonic()-start})
        (HERE/'verification.json').write_text(json.dumps(checks,indent=2)+'\n')
        if result.returncode:raise SystemExit('verification failed; see verification.log')
    run([args.go,'test','./pirconfig','./planner','-count=1'])
    run([args.go,'vet','./pirconfig','./experiments/cyclic-scaling'])
    with tempfile.TemporaryDirectory(prefix='depir-s13-tests-') as directory:
        copy=Path(directory)/'repo';shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','__pycache__','.DS_Store'))
        run([args.go,'test','./...'],cwd=copy)
print('Catalogue/planner tests, vet and isolated full suite passed.')
