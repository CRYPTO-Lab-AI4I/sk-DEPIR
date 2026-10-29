#!/usr/bin/env python3
"""Run the standalone native matrix/Horner check with address/UB sanitizers."""
import argparse,hashlib,json,os,subprocess,tempfile,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--cxx',default=os.environ.get('CXX','clang++'));a=p.parse_args();out=HERE/'native-validation';out.mkdir(exist_ok=False)
sources=[ROOT/'encoding/liftrs-cpp'/n for n in ['matrix_test.cpp','c_api.cpp','c_api.h','transform.hpp']]+[Path(__file__)]
r={'source_sha256':{str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sources},'checks':[],'status':'running'}
with tempfile.TemporaryDirectory(prefix='depir-matrix-asan-') as tmp,(out/'checks.log').open('w') as log:
 binary=str(Path(tmp)/'matrix');commands=[[a.cxx,'-std=c++17','-O1','-g','-fsanitize=address,undefined','-fno-omit-frame-pointer','-pthread','encoding/liftrs-cpp/matrix_test.cpp','encoding/liftrs-cpp/c_api.cpp','-o',binary],[binary]]
 for cmd in commands:
  log.write('$ '+' '.join(cmd)+'\n');log.flush();start=time.monotonic();x=subprocess.run(cmd,cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,timeout=120)
  r['checks'].append({'command':cmd,'exit_code':x.returncode,'wall_s':time.monotonic()-start});(out/'verification.json').write_text(json.dumps(r,indent=2)+'\n')
  if x.returncode:raise SystemExit('native matrix sanitizer check failed')
r['status']='passed';(out/'verification.json').write_text(json.dumps(r,indent=2)+'\n');print('native matrix sanitizer PASS')
