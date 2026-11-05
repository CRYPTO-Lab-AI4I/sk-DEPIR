#!/usr/bin/env python3
"""Run the complete-online C ABI fixture with address and UB sanitizers."""
import argparse
import hashlib
import json
import os
import subprocess
import tempfile
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--cxx', default=os.environ.get('CXX', 'clang++'))
parser.add_argument('--run-name', default='native-validation')
args = parser.parse_args()
out = HERE / args.run_name
out.mkdir(exist_ok=False)
sources = [ROOT / 'encoding/liftrs-cpp' / name for name in
           ['online_test.cpp', 'c_api.cpp', 'c_api.h', 'transform.hpp']] + [Path(__file__)]
record = {'revision': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
          'source_sha256': {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sources},
          'checks': [], 'status': 'running'}
with tempfile.TemporaryDirectory(prefix='depir-online-asan-') as temporary, (out / 'checks.log').open('w') as log:
    binary = str(Path(temporary) / 'online')
    commands = [[args.cxx, '-std=c++17', '-O1', '-g', '-fsanitize=address,undefined',
                 '-fno-omit-frame-pointer', '-pthread', 'encoding/liftrs-cpp/online_test.cpp',
                 'encoding/liftrs-cpp/c_api.cpp', '-o', binary], [binary]]
    for command in commands:
        log.write('$ ' + ' '.join(command) + '\n')
        log.flush()
        start = time.monotonic()
        result = subprocess.run(command, cwd=ROOT, stdout=log, stderr=subprocess.STDOUT, timeout=120)
        record['checks'].append({'command': command, 'exit_code': result.returncode,
                                 'wall_s': time.monotonic() - start})
        (out / 'verification.json').write_text(json.dumps(record, indent=2) + '\n')
        if result.returncode:
            raise SystemExit('native online sanitizer check failed')
record['status'] = 'passed'
(out / 'verification.json').write_text(json.dumps(record, indent=2) + '\n')
print('native online sanitizer PASS')
