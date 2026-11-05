#!/usr/bin/env python3
"""Verify clock ABI safety and unchanged native/Go encoding paths."""
import argparse
import hashlib
import json
import os
import subprocess
import tempfile
import time
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--go',default='go')
parser.add_argument('--run-name',default='validation')
parser.add_argument('--native-only',action='store_true')
args=parser.parse_args()
out=HERE/args.run_name
out.mkdir(exist_ok=False)
sources=list((ROOT/'encoding').rglob('*.go'))+list((ROOT/'encoding/liftrs-cpp').glob('*.*'))+list(HERE.glob('*.py'))+[HERE/'main.go']
manifest={str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sources}
for name,digest in manifest.items():
    original=subprocess.run(['git','show','HEAD:'+name],cwd=ROOT,capture_output=True)
    if original.returncode or hashlib.sha256(original.stdout).hexdigest()!=digest:
        (ROOT/'experiments/source-snapshots'/(digest+'.txt')).write_bytes((ROOT/name).read_bytes())
record={'revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'source_sha256':manifest,'status':'running','checks':[]}
with (out/'checks.log').open('w') as log:
    def run(command,env=None):
        log.write('$ '+' '.join(command)+'\n');log.flush()
        start=time.monotonic()
        result=subprocess.run(command,cwd=ROOT,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=600)
        record['checks'].append({'command':command,'exit_code':result.returncode,'wall_s':time.monotonic()-start})
        (out/'verification.json').write_text(json.dumps(record,indent=2)+'\n')
        if result.returncode:raise SystemExit('validation failed; evidence retained')
    if not args.native_only:
        run([args.go,'test','-race','-tags','liftrs_cpp','./encoding/liftrs-shared','-run','TestCPPOnline|TestCPPMatrix','-count=1'])
        run([args.go,'test','-tags','liftrs_cpp','./encoding/liftrs-cpp','./experiments/encoding-timing'])
        run([args.go,'vet','-tags','liftrs_cpp','./encoding/liftrs-cpp','./encoding/liftrs-shared','./experiments/encoding-timing'])
        run([args.go,'test','./encoding/liftrs-shared','./encoding/liftrs-go'],env=dict(os.environ,CGO_ENABLED='0'))
    with tempfile.TemporaryDirectory(prefix='depir-clock-asan-') as temporary:
        source=Path(temporary)/'clocks.cpp';binary=Path(temporary)/'clocks'
        source.write_text('''#include "c_api.h"
#include <cstdint>
#include <thread>
#include <chrono>
int main() {
 uint64_t a[4]={}, b[4]={}; char error[256]={};
 if (!liftrs_read_clocks(nullptr,error,sizeof(error))) return 1;
 if (liftrs_read_clocks(a,error,sizeof(error))) return 2;
 std::this_thread::sleep_for(std::chrono::milliseconds(5));
 if (liftrs_read_clocks(b,error,sizeof(error))) return 3;
 for (unsigned i=0;i<4;i++) if (!a[i] || !b[i]) return 4;
 // Wall time may legitimately step; only elapsed/CPU clocks must not go back.
 for (unsigned i : {0u,2u,3u}) if (b[i]<a[i]) return 5;
 return 0;
}
''')
        run(['clang++','-std=c++17','-O1','-g','-fsanitize=address,undefined','-fno-omit-frame-pointer','-pthread','-Iencoding/liftrs-cpp',str(source),'encoding/liftrs-cpp/c_api.cpp','-o',str(binary)])
        run([str(binary)])
record['status']='passed'
(out/'verification.json').write_text(json.dumps(record,indent=2)+'\n')
print('clock and encoding validation PASS')
