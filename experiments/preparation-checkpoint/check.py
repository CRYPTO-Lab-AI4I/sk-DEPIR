#!/usr/bin/env python3
"""Verify bounded preparation and both isolated repository suites."""
import argparse
import hashlib
import json
import os
import shutil
import subprocess
import tempfile
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--go', default='go')
parser.add_argument('--run-name', default='validation')
args = parser.parse_args()
out = HERE / args.run_name
out.mkdir(exist_ok=False)
sources = list((ROOT / 'encoding').rglob('*.go')) + list((ROOT / 'encoding/liftrs-cpp').glob('*.*')) + [Path(__file__)]
manifest = {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sources}
# Preserve final dirty-source verification independently of benchmark manifests.
for name, digest in manifest.items():
    historical = subprocess.run(['git', 'show', 'HEAD:' + name], cwd=ROOT, capture_output=True)
    if historical.returncode or hashlib.sha256(historical.stdout).hexdigest() != digest:
        (ROOT / 'experiments/source-snapshots' / (digest + '.txt')).write_bytes((ROOT / name).read_bytes())
record = {'status': 'running', 'checks': [], 'source_sha256': manifest,
          'revision': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip()}


def save():
    (out / 'verification.json').write_text(json.dumps(record, indent=2) + '\n')


with (out / 'checks.log').open('w') as log:
    def run(command, cwd=ROOT, env=None):
        log.write('$ ' + ' '.join(command) + '\n')
        log.flush()
        start = time.monotonic()
        result = subprocess.run(command, cwd=cwd, env=env, stdout=log, stderr=subprocess.STDOUT, timeout=600)
        record['checks'].append({'command': command, 'isolated': cwd != ROOT,
                                 'exit_code': result.returncode, 'wall_s': time.monotonic() - start})
        save()
        if result.returncode:
            raise SystemExit('verification failed; evidence retained')

    run([args.go, 'test', '-race', '-tags', 'liftrs_cpp', './encoding/liftrs-shared', '-run',
         'TestBatchedFactorProducts|TestCompactSupportValidation|TestPreparationProductWorkspace|TestObservedPreparationWorkspace|TestScalingPreparationCPP|TestCPPOnline|TestCPPMatrix', '-count=1'])
    run([args.go, 'vet', '-tags', 'liftrs_cpp', './encoding/...', './pirconfig', './planner'])
    with tempfile.TemporaryDirectory(prefix='depir-online-full-') as temporary:
        copy = Path(temporary) / 'repo'
        shutil.copytree(ROOT, copy, ignore=shutil.ignore_patterns('.git', '.cache', '__pycache__', '.DS_Store'))
        run(['make', '-C', 'encoding/rme_legacy', 'rme'], cwd=copy)
        run([args.go, 'test', './...'], cwd=copy, env=dict(os.environ, CGO_ENABLED='0'))
        run([args.go, 'test', '-tags', 'liftrs_cpp', './...'], cwd=copy)
record['status'] = 'passed'
save()
print('preparation-checkpoint validation PASS')
