#!/usr/bin/env python3
"""Validate renamed Go imports, independent native arithmetic, and native memory safety."""
import argparse,json,shutil,subprocess,tempfile,time,runpy
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');p.add_argument('--cxx',default='clang++');args=p.parse_args();checks=[]
with (HERE/'verification.log').open('w') as log:
 def run(command,cwd=ROOT):
  log.write('$ '+' '.join(command)+'\n');log.flush();start=time.monotonic();r=subprocess.run(command,cwd=cwd,stdout=log,stderr=subprocess.STDOUT)
  checks.append({'command':command,'isolated_copy':cwd!=ROOT,'exit_code':r.returncode,'seconds':time.monotonic()-start});(HERE/'verification.json').write_text(json.dumps(checks,indent=2)+'\n')
  if r.returncode:raise SystemExit('check failed; see verification.log')
 run([args.go,'test','./encoding/liftrs-shared','./pirconfig','./planner','-count=1'])
 run([args.go,'vet','./encoding/liftrs-shared','./pirconfig','./planner'])
 with tempfile.TemporaryDirectory(prefix='depir-language-checks-') as directory:
  target=Path(directory)/'native-tests'
  run([args.cxx,'-std=c++17','-O1','-g','-fsanitize=address,undefined','-fno-omit-frame-pointer',str(ROOT/'encoding/liftrs-cpp/test.cpp'),'-o',str(target)])
  run([str(target)])
  copy=Path(directory)/'repo';shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','.cache','__pycache__','.DS_Store'))
  run([args.go,'test','./...'],cwd=copy)
 print('Go/native sanitizer/full-suite checks passed.')
