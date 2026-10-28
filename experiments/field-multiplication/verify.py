#!/usr/bin/env python3
"""Run arithmetic/encoder regression checks without changing legacy CSV fixtures."""
import argparse,json,os,shutil,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');args=p.parse_args()
checks=[]
with (HERE/'verification.log').open('w') as log:
    def run(command,cwd=ROOT):
        print(' '.join(command),flush=True);log.write('$ '+' '.join(command)+'\n');log.flush();start=time.monotonic()
        result=subprocess.run(command,cwd=cwd,stdout=log,stderr=subprocess.STDOUT)
        checks.append({'command':command,'isolated_copy':cwd!=ROOT,'exit_code':result.returncode,'seconds':time.monotonic()-start})
        (HERE/'verification.json').write_text(json.dumps(checks,indent=2)+'\n')
        if result.returncode:raise SystemExit('verification failed; see verification.log')
    run([args.go,'test','./encoding/liftrs-shared','./pirconfig','./experiments/field-multiplication','-count=1'])
    run([args.go,'test','./encoding/liftrs-shared','-run','TestConstantTableSizes','-v','-count=1'])
    run([args.go,'test','-race','./encoding/liftrs-shared','./pirconfig','./experiments/field-multiplication'])
    run([args.go,'vet','./encoding/liftrs-shared','./pirconfig','./experiments/field-multiplication'])
    with tempfile.TemporaryDirectory(prefix='depir-mul-tests-') as directory:
        copy=Path(directory)/'repo'
        shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','.cache','__pycache__','.DS_Store'))
        run([args.go,'test','./...'],cwd=copy)
