#!/usr/bin/env python3
"""Validate optional Go/C++ builds, exact outputs, lifecycle and native boundary safety."""
import argparse,json,os,shutil,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--go',default='go');p.add_argument('--cxx',default='clang++');args=p.parse_args();checks=[]
with (HERE/'verification.log').open('w') as log:
 def run(command,cwd=ROOT,env=None):
  log.write('$ '+' '.join(command)+'\n');log.flush();start=time.monotonic();r=subprocess.run(command,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT)
  checks.append({'command':command,'isolated_copy':cwd!=ROOT,'env_overrides':{'CGO_ENABLED':env['CGO_ENABLED']} if env and 'CGO_ENABLED' in env else {},'exit_code':r.returncode,'seconds':time.monotonic()-start});(HERE/'verification.json').write_text(json.dumps(checks,indent=2)+'\n')
  if r.returncode:raise SystemExit('check failed; see verification.log')
 run([args.go,'test','-tags','liftrs_cpp','./encoding/liftrs-shared','./encoding/liftrs-cpp','-count=1'])
 run([args.go,'test','./encoding/liftrs-shared','-run','TestCPPOptionalBuild','-count=1'],env=dict(os.environ,CGO_ENABLED='0'))
 run([args.go,'vet','-tags','liftrs_cpp','./encoding/liftrs-shared','./encoding/liftrs-cpp'])
 with tempfile.TemporaryDirectory(prefix='depir-native-checks-') as directory:
  binary=Path(directory)/'c-api-check'
  run([args.cxx,'-std=c++17','-O1','-g','-fsanitize=address,undefined','-fno-omit-frame-pointer',str(ROOT/'encoding/liftrs-cpp/c_api.cpp'),str(ROOT/'encoding/liftrs-cpp/c_api_test.cpp'),'-o',str(binary)])
  run([str(binary),str(ROOT/'.cache/liftrs/language-s13/transform.bin')])
  copy=Path(directory)/'repo';shutil.copytree(ROOT,copy,ignore=shutil.ignore_patterns('.git','.cache','__pycache__','.DS_Store'))
  run([args.go,'test','./...'],cwd=copy)
  run([args.go,'test','-tags','liftrs_cpp','./...'],cwd=copy)
print('Native/default/full-suite and C ABI sanitizer checks passed.')
